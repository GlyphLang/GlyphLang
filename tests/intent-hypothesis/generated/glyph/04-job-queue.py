import asyncio
import time
import uuid
from datetime import datetime, timezone
from typing import Any, Callable, Optional

from fastapi import Depends, FastAPI, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel

import jwt as pyjwt


# --- Models ---

class EmailJob(BaseModel):
    to: str
    subject: str
    template: str
    data: Any = None


class ReportConfig(BaseModel):
    type: str
    date_range: str
    format: str


class AuthUser(BaseModel):
    id: str


# --- Database ---

class JobCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Get(self, id: str) -> Optional[dict]:
        return self._store.get(id)

    def Create(self, record: dict) -> dict:
        self._store[record["id"]] = record
        return record


class SessionCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Where(self, filters: dict) -> list[dict]:
        results = list(self._store.values())
        for key, value in filters.items():
            results = [r for r in results if r.get(key) == value]
        return results


class UserCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Where(self, filters: dict) -> list[dict]:
        results = list(self._store.values())
        for key, value in filters.items():
            results = [r for r in results if r.get(key) == value]
        return results


class EmailQueueCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Create(self, record: dict) -> dict:
        self._store[record["id"]] = record
        return record


class Database:
    def __init__(self) -> None:
        self.jobs = JobCollection()
        self.sessions = SessionCollection()
        self.users = UserCollection()
        self.email_queue = EmailQueueCollection()


# --- Auth ---

JWT_SECRET = "secret"
security = HTTPBearer()


def auth(credentials: HTTPAuthorizationCredentials = Depends(security)) -> AuthUser:
    try:
        payload = pyjwt.decode(credentials.credentials, JWT_SECRET, algorithms=["HS256"])
        return AuthUser(id=payload["sub"])
    except pyjwt.PyJWTError:
        raise HTTPException(status_code=401, detail="Invalid token")


# --- Helpers ---

def generate_id() -> str:
    return str(uuid.uuid4())


def now() -> str:
    return datetime.now(timezone.utc).isoformat()


# --- Queue Workers ---

class QueueWorker:
    def __init__(self) -> None:
        self._handlers: dict[str, Callable] = {}
        self._queues: dict[str, asyncio.Queue] = {}

    def register(self, queue_name: str, handler: Callable) -> None:
        self._handlers[queue_name] = handler
        self._queues[queue_name] = asyncio.Queue()

    async def enqueue(self, queue_name: str, message: dict) -> None:
        if queue_name in self._queues:
            await self._queues[queue_name].put(message)

    async def process(self, queue_name: str) -> Optional[dict]:
        if queue_name not in self._handlers:
            return None
        handler = self._handlers[queue_name]
        queue = self._queues[queue_name]
        if not queue.empty():
            message = await queue.get()
            return await handler(message)
        return None


queue_worker = QueueWorker()


async def handle_email_send(message: dict) -> dict:
    to = message["to"]
    subject = message["subject"]
    return {"success": True, "to": to, "subject": subject}


async def handle_report_generate(message: dict) -> dict:
    report_type = message["type"]
    format_ = message["format"]
    return {"success": True, "type": report_type, "format": format_}


queue_worker.register("email.send", handle_email_send)
queue_worker.register("report.generate", handle_report_generate)


# --- Cron Jobs ---

class CronScheduler:
    def __init__(self) -> None:
        self._jobs: list[dict] = []

    def register(self, schedule: str, name: str, handler: Callable) -> None:
        self._jobs.append({"schedule": schedule, "name": name, "handler": handler})

    async def run_job(self, name: str) -> Optional[dict]:
        for job in self._jobs:
            if job["name"] == name:
                return await job["handler"]()
        return None


cron_scheduler = CronScheduler()


async def cleanup_expired() -> dict:
    db = _database
    expired = db.sessions.Where({"expired": True})
    return {"task": "cleanup", "timestamp": now()}


async def weekly_digest() -> dict:
    db = _database
    users = db.users.Where({"digest_enabled": True})
    return {"task": "weekly_digest", "timestamp": now()}


cron_scheduler.register("0 2 * * *", "cleanup_expired", cleanup_expired)
cron_scheduler.register("0 8 * * 1", "weekly_digest", weekly_digest)


# --- Event Handlers ---

class EventBus:
    def __init__(self) -> None:
        self._handlers: dict[str, list[Callable]] = {}

    def on(self, event_name: str, handler: Callable) -> None:
        if event_name not in self._handlers:
            self._handlers[event_name] = []
        self._handlers[event_name].append(handler)

    async def emit(self, event_name: str, event: dict) -> list[dict]:
        results = []
        for handler in self._handlers.get(event_name, []):
            result = await handler(event)
            results.append(result)
        return results


event_bus = EventBus()


async def handle_user_created(event: dict) -> dict:
    user_email = event["email"]
    user_name = event["name"]
    return {"queued": True, "email": user_email, "name": user_name}


event_bus.on("user.created", handle_user_created)


# --- App ---

app = FastAPI()

_database = Database()


def get_db() -> Database:
    return _database


# --- Routes ---

@app.get("/api/jobs/{id}")
async def get_job(
    id: str,
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> dict:
    job = db.jobs.Get(id)
    if job is None:
        raise HTTPException(status_code=404, detail="Job not found")
    return job


@app.post("/api/jobs/email", status_code=201)
async def create_email_job(
    input: EmailJob,
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> dict:
    job_id = generate_id()
    created = db.email_queue.Create({
        "id": job_id,
        "to": input.to,
        "subject": input.subject,
        "template": input.template,
        "data": input.data,
        "status": "pending",
        "created_at": now(),
    })
    return {"job_id": job_id, "status": "pending"}
