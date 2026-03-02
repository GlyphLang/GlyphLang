import time
import uuid
from datetime import datetime, timezone
from typing import Any, Optional

from fastapi import Depends, FastAPI, HTTPException, Security
from fastapi.security import APIKeyHeader
from pydantic import BaseModel


# --- Models ---

class WebhookPayload(BaseModel):
    event: str
    data: Any
    timestamp: int
    signature: str


class WebhookResult(BaseModel):
    processed: bool
    event_id: Optional[str] = None


# --- Database ---

class WebhookLogCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Create(self, record: dict) -> dict:
        self._store[record["id"]] = record
        return record


class PaymentCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Update(self, id: str, fields: dict) -> dict:
        if id not in self._store:
            self._store[id] = {"id": id}
        self._store[id].update(fields)
        return self._store[id]


class CustomerCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Create(self, data: Any) -> dict:
        record = data if isinstance(data, dict) else {"id": str(uuid.uuid4())}
        self._store[record.get("id", str(uuid.uuid4()))] = record
        return record


class Database:
    def __init__(self) -> None:
        self.webhook_logs = WebhookLogCollection()
        self.payments = PaymentCollection()
        self.customers = CustomerCollection()


# --- Dependencies ---

app = FastAPI()

_database = Database()

api_key_header = APIKeyHeader(name="X-API-Key")

RATE_LIMIT_WINDOW = 60
RATE_LIMIT_MAX_STRIPE = 1000
_rate_limit_store: dict[str, list[float]] = {}


def get_db() -> Database:
    return _database


async def verify_api_key(api_key: str = Security(api_key_header)) -> str:
    if not api_key:
        raise HTTPException(status_code=401, detail="Missing API key")
    return api_key


async def check_rate_limit_stripe(api_key: str = Security(api_key_header)) -> str:
    now = time.time()
    if api_key not in _rate_limit_store:
        _rate_limit_store[api_key] = []
    _rate_limit_store[api_key] = [
        t for t in _rate_limit_store[api_key] if now - t < RATE_LIMIT_WINDOW
    ]
    if len(_rate_limit_store[api_key]) >= RATE_LIMIT_MAX_STRIPE:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")
    _rate_limit_store[api_key].append(now)
    return api_key


def generate_id() -> str:
    return str(uuid.uuid4())


# --- Routes ---

@app.post("/webhooks/stripe")
async def stripe_webhook(
    input: WebhookPayload,
    api_key: str = Depends(verify_api_key),
    _rate: str = Depends(check_rate_limit_stripe),
    db: Database = Depends(get_db),
) -> WebhookResult:
    event_id = generate_id()

    db.webhook_logs.Create({
        "id": event_id,
        "event": input.event,
        "received_at": datetime.now(timezone.utc).isoformat(),
    })

    if input.event == "payment.succeeded":
        db.payments.Update(input.data["id"], {"status": "completed"})

    if input.event == "payment.failed":
        db.payments.Update(input.data["id"], {"status": "failed"})

    if input.event == "customer.created":
        db.customers.Create(input.data)

    return WebhookResult(processed=True, event_id=event_id)


@app.post("/webhooks/github")
async def github_webhook(
    input: WebhookPayload,
    api_key: str = Depends(verify_api_key),
    db: Database = Depends(get_db),
) -> WebhookResult:
    event_id = generate_id()

    db.webhook_logs.Create({
        "id": event_id,
        "source": "github",
        "event": input.event,
        "received_at": datetime.now(timezone.utc).isoformat(),
    })

    return WebhookResult(processed=True, event_id=event_id)
