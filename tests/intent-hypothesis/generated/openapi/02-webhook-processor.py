import time
import uuid
from datetime import datetime, timezone
from typing import Any, Optional

from fastapi import Depends, FastAPI, HTTPException, Security
from fastapi.responses import JSONResponse
from fastapi.security import APIKeyHeader
from pydantic import BaseModel

app = FastAPI(
    title="Webhook Processor",
    description=(
        "Webhook processing service with API key authentication, rate limiting, "
        "and event-based routing for Stripe and GitHub webhooks."
    ),
    version="1.0.0",
)

# ---------------------------------------------------------------------------
# Pydantic models
# ---------------------------------------------------------------------------


class WebhookPayload(BaseModel):
    event: str
    data: Any
    timestamp: int
    signature: str


class WebhookResult(BaseModel):
    processed: bool
    event_id: Optional[str] = None


# ---------------------------------------------------------------------------
# In-memory stores (simulating database tables)
# ---------------------------------------------------------------------------

_webhook_logs: dict[str, dict] = {}
_payments: dict[str, dict] = {}
_customers: dict[str, dict] = {}

# ---------------------------------------------------------------------------
# Rate limiter state
# ---------------------------------------------------------------------------

RATE_LIMIT_MAX = 1000
RATE_LIMIT_WINDOW = 60  # seconds

_rate_limit_store: dict[str, list[float]] = {}

# ---------------------------------------------------------------------------
# Security / dependencies
# ---------------------------------------------------------------------------

api_key_header = APIKeyHeader(name="X-API-Key")


async def verify_api_key(api_key: str = Security(api_key_header)) -> str:
    """Validate that an API key is present."""
    if not api_key:
        raise HTTPException(status_code=401, detail="Missing or invalid API key")
    return api_key


def _enforce_rate_limit(api_key: str) -> None:
    """Sliding-window rate limiter: max 1000 requests per 60-second window."""
    now = time.time()
    timestamps = _rate_limit_store.setdefault(api_key, [])
    # Prune entries outside the current window
    _rate_limit_store[api_key] = [t for t in timestamps if now - t < RATE_LIMIT_WINDOW]
    if len(_rate_limit_store[api_key]) >= RATE_LIMIT_MAX:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")
    _rate_limit_store[api_key].append(now)


def _generate_event_id() -> str:
    return str(uuid.uuid4())


# ---------------------------------------------------------------------------
# Event handlers for Stripe webhook routing
# ---------------------------------------------------------------------------


def _handle_payment_succeeded(data: Any) -> None:
    payment_id = data.get("id") if isinstance(data, dict) else None
    if payment_id:
        _payments.setdefault(payment_id, {"id": payment_id})
        _payments[payment_id]["status"] = "completed"


def _handle_payment_failed(data: Any) -> None:
    payment_id = data.get("id") if isinstance(data, dict) else None
    if payment_id:
        _payments.setdefault(payment_id, {"id": payment_id})
        _payments[payment_id]["status"] = "failed"


def _handle_customer_created(data: Any) -> None:
    if isinstance(data, dict):
        customer_id = data.get("id", str(uuid.uuid4()))
        _customers[customer_id] = {**data, "id": customer_id}


_STRIPE_EVENT_HANDLERS = {
    "payment.succeeded": _handle_payment_succeeded,
    "payment.failed": _handle_payment_failed,
    "customer.created": _handle_customer_created,
}

# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------


@app.post("/webhooks/stripe", response_model=WebhookResult, responses={
    401: {"description": "Missing or invalid API key"},
    429: {"description": "Rate limit exceeded"},
})
async def stripe_webhook(
    payload: WebhookPayload,
    api_key: str = Depends(verify_api_key),
) -> WebhookResult:
    """Process incoming Stripe webhooks.

    Rate limited to 1000 requests per minute. Routes events by type:
    - payment.succeeded -> update payment status to completed
    - payment.failed -> update payment status to failed
    - customer.created -> create a new customer record
    """
    _enforce_rate_limit(api_key)

    event_id = _generate_event_id()

    # Log webhook
    _webhook_logs[event_id] = {
        "id": event_id,
        "source": "stripe",
        "event": payload.event,
        "data": payload.data,
        "timestamp": payload.timestamp,
        "received_at": datetime.now(timezone.utc).isoformat(),
    }

    # Route by event type
    handler = _STRIPE_EVENT_HANDLERS.get(payload.event)
    if handler:
        handler(payload.data)

    return WebhookResult(processed=True, event_id=event_id)


@app.post("/webhooks/github", response_model=WebhookResult, responses={
    401: {"description": "Missing or invalid API key"},
})
async def github_webhook(
    payload: WebhookPayload,
    api_key: str = Depends(verify_api_key),
) -> WebhookResult:
    """Process incoming GitHub webhooks.

    Logs the webhook with source 'github' and returns confirmation.
    """
    event_id = _generate_event_id()

    # Log webhook
    _webhook_logs[event_id] = {
        "id": event_id,
        "source": "github",
        "event": payload.event,
        "data": payload.data,
        "timestamp": payload.timestamp,
        "received_at": datetime.now(timezone.utc).isoformat(),
    }

    return WebhookResult(processed=True, event_id=event_id)


# ---------------------------------------------------------------------------
# Custom exception handler
# ---------------------------------------------------------------------------


@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc: HTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": exc.detail},
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
