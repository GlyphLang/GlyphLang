import time
import uuid
from typing import Any, Optional

from fastapi import Depends, FastAPI, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel

import jwt as pyjwt


# --- Models ---

class EmailJob(BaseModel):
    to: str
    subject: str
    template: str
    data: Optional[Any] = None


class ReportConfig(BaseModel):
    type: str
    date_range: str
    format: str


class JobRecord(BaseModel):
    id: str
    type: str
    status: str
    payload: dict
    created_at: int


class JobStatusResponse(BaseModel):
    id: str
    status: str


class JobSubmitResponse(BaseModel):
    id: str
    status: str


class AuthUser(BaseModel):
    id: str


# --- Database ---

class JobQueue:
    def __init__(self) -> None:
        self._jobs: dict[str, JobRecord] = {}

    def create(self, job_type: str, payload: dict) -> JobRecord:
        job_id = str(uuid.uuid4())
        record = JobRecord(
            id=job_id,
            type=job_type,
            status="pending",
            payload=payload,
            created_at=int(time.time()),
        )
        self._jobs[job_id] = record
        return record

    def get(self, job_id: str) -> Optional[JobRecord]:
        return self._jobs.get(job_id)


class Database:
    def __init__(self) -> None:
        self.jobs = JobQueue()

    def query(self, collection: str, filters: dict) -> list[dict]:
        return []


# --- Auth ---

JWT_SECRET = "secret"
security = HTTPBearer()


def auth(credentials: HTTPAuthorizationCredentials = Depends(security)) -> AuthUser:
    try:
        payload = pyjwt.decode(credentials.credentials, JWT_SECRET, algorithms=["HS256"])
        return AuthUser(id=payload["sub"])
    except pyjwt.PyJWTError:
        raise HTTPException(status_code=401, detail="Invalid token")


# --- Queue Workers ---

def process_email_send(message: dict) -> dict:
    recipient = message.get("to")
    subject = message.get("subject")
    return {"success": True, "recipient": recipient, "subject": subject}


def process_report_generate(message: dict) -> dict:
    report_type = message.get("type")
    report_format = message.get("format")
    return {"success": True, "type": report_type, "format": report_format}


QUEUE_WORKERS = {
    "email.send": process_email_send,
    "report.generate": process_report_generate,
}


# --- Cron Jobs ---

def cleanup_expired(db: Database) -> dict:
    expired = db.query("sessions", {"expired": True})
    return {"task": "cleanup_expired", "schedule": "0 2 * * *", "found": len(expired)}


def weekly_digest(db: Database) -> dict:
    users = db.query("users", {"digest_enabled": True})
    return {"task": "weekly_digest", "schedule": "0 8 * * 1", "found": len(users)}


CRON_JOBS = {
    "cleanup_expired": {"schedule": "0 2 * * *", "handler": cleanup_expired},
    "weekly_digest": {"schedule": "0 8 * * 1", "handler": weekly_digest},
}


# --- Event Handlers ---

def handle_user_created(event_data: dict) -> dict:
    email = event_data.get("email")
    name = event_data.get("name")
    return {"notification": "welcome", "email": email, "name": name}


EVENT_HANDLERS = {
    "user.created": handle_user_created,
}


# --- App ---

app = FastAPI()

_database = Database()


def get_db() -> Database:
    return _database


# --- API Endpoints ---

@app.get("/api/jobs/{job_id}")
async def get_job_status(
    job_id: str,
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> JobStatusResponse:
    job = db.jobs.get(job_id)
    if job is None:
        raise HTTPException(status_code=404, detail="Job not found")
    return JobStatusResponse(id=job.id, status=job.status)


@app.post("/api/jobs/email", status_code=201)
async def submit_email_job(
    email_job: EmailJob,
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> JobSubmitResponse:
    record = db.jobs.create("email.send", email_job.model_dump())
    return JobSubmitResponse(id=record.id, status=record.status)
