import time
import uuid
from typing import Any, Optional

from fastapi import Depends, FastAPI, HTTPException, Request
from pydantic import BaseModel

app = FastAPI()

# --- In-memory stores ---

webhook_logs: list[dict] = []
payments: dict[str, dict] = {}
customers: dict[str, dict] = {}

# --- Rate limiter ---

rate_limit_store: dict[str, list[float]] = {}
RATE_LIMIT = 1000
RATE_WINDOW = 60


def check_rate_limit(api_key: str) -> None:
    now = time.time()
    if api_key not in rate_limit_store:
        rate_limit_store[api_key] = []
    timestamps = rate_limit_store[api_key]
    rate_limit_store[api_key] = [t for t in timestamps if now - t < RATE_WINDOW]
    if len(rate_limit_store[api_key]) >= RATE_LIMIT:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")
    rate_limit_store[api_key].append(now)


# --- Auth dependency ---

API_KEYS = {"test-key-1", "test-key-2"}


async def verify_api_key(request: Request) -> str:
    api_key = request.headers.get("X-API-Key")
    if not api_key or api_key not in API_KEYS:
        raise HTTPException(status_code=401, detail="Invalid API key")
    return api_key


# --- Models ---


class WebhookPayload(BaseModel):
    event: str
    data: Any
    timestamp: int
    signature: str


class WebhookResult(BaseModel):
    processed: bool
    event_id: Optional[str] = None


# --- Event handlers ---


def handle_payment_succeeded(data: Any) -> None:
    payment_id = data.get("payment_id") if isinstance(data, dict) else None
    if payment_id:
        payments[payment_id] = {**payments.get(payment_id, {}), "status": "completed"}


def handle_payment_failed(data: Any) -> None:
    payment_id = data.get("payment_id") if isinstance(data, dict) else None
    if payment_id:
        payments[payment_id] = {**payments.get(payment_id, {}), "status": "failed"}


def handle_customer_created(data: Any) -> None:
    customer_id = data.get("customer_id") if isinstance(data, dict) else None
    if customer_id:
        customers[customer_id] = {**data}


EVENT_HANDLERS = {
    "payment.succeeded": handle_payment_succeeded,
    "payment.failed": handle_payment_failed,
    "customer.created": handle_customer_created,
}


# --- Endpoints ---


@app.post("/webhooks/stripe", response_model=WebhookResult)
async def stripe_webhook(
    payload: WebhookPayload, api_key: str = Depends(verify_api_key)
) -> WebhookResult:
    check_rate_limit(api_key)

    event_id = str(uuid.uuid4())
    webhook_logs.append(
        {
            "id": event_id,
            "source": "stripe",
            "event": payload.event,
            "data": payload.data,
            "timestamp": payload.timestamp,
            "signature": payload.signature,
            "logged_at": int(time.time()),
        }
    )

    handler = EVENT_HANDLERS.get(payload.event)
    if handler:
        handler(payload.data)

    return WebhookResult(processed=True, event_id=event_id)


@app.post("/webhooks/github", response_model=WebhookResult)
async def github_webhook(
    payload: WebhookPayload, api_key: str = Depends(verify_api_key)
) -> WebhookResult:
    event_id = str(uuid.uuid4())
    webhook_logs.append(
        {
            "id": event_id,
            "source": "github",
            "event": payload.event,
            "data": payload.data,
            "timestamp": payload.timestamp,
            "signature": payload.signature,
            "logged_at": int(time.time()),
        }
    )

    return WebhookResult(processed=True, event_id=event_id)
