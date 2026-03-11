import uuid
from datetime import datetime, timezone
from typing import Any, Optional

from fastapi import Depends, FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel

app = FastAPI(
    title="Background Job Processing System",
    description=(
        "Job processing system with queue workers, scheduled cron tasks, "
        "event handlers, and API endpoints. All REST endpoints require JWT auth."
    ),
    version="1.0.0",
)

# ---------------------------------------------------------------------------
# Security
# ---------------------------------------------------------------------------

bearer_scheme = HTTPBearer()


async def require_auth(
    credentials: HTTPAuthorizationCredentials = Depends(bearer_scheme),
) -> str:
    """Validate the Bearer token and return the token string."""
    if not credentials or not credentials.credentials:
        raise HTTPException(status_code=401, detail="Missing authentication token")
    return credentials.credentials


# ---------------------------------------------------------------------------
# Pydantic models
# ---------------------------------------------------------------------------


class EmailJob(BaseModel):
    to: str
    subject: str
    template: str
    data: Optional[Any] = None


class ReportConfig(BaseModel):
    type: str
    date_range: str
    format: str


class Error(BaseModel):
    error: str


class JobSubmittedResponse(BaseModel):
    job_id: str
    status: str


# ---------------------------------------------------------------------------
# In-memory stores
# ---------------------------------------------------------------------------

_jobs: dict[str, dict] = {}
_email_queue: list[dict] = []
_event_log: list[dict] = []


def _generate_job_id() -> str:
    return str(uuid.uuid4())


# ---------------------------------------------------------------------------
# REST API Routes
# ---------------------------------------------------------------------------


@app.get(
    "/api/jobs/{id}",
    response_model=dict,
    responses={404: {"model": Error}},
)
async def get_job_status(id: str, _token: str = Depends(require_auth)):
    """Check job status by ID."""
    job = _jobs.get(id)
    if job is None:
        raise HTTPException(status_code=404, detail="Job not found")
    return job


@app.post("/api/jobs/email", response_model=JobSubmittedResponse, status_code=201)
async def submit_email_job(body: EmailJob, _token: str = Depends(require_auth)):
    """Submit a new email job to the queue.

    Creates a record in the email queue with the job details,
    a generated job ID, status "pending", and a created_at timestamp.
    Returns the job ID and pending status.
    """
    job_id = _generate_job_id()
    record = {
        "job_id": job_id,
        "status": "pending",
        "to": body.to,
        "subject": body.subject,
        "template": body.template,
        "data": body.data,
        "created_at": datetime.now(timezone.utc).isoformat(),
    }
    _jobs[job_id] = record
    _email_queue.append(record)
    return {"job_id": job_id, "status": "pending"}


# ---------------------------------------------------------------------------
# Queue Workers (x-queue-workers)
# ---------------------------------------------------------------------------


async def worker_email_send(message: dict) -> dict:
    """Process email sending jobs.

    Extract recipient (to) and subject from the message.
    Return {success: true, to, subject}.
    """
    to = message.get("to", "")
    subject = message.get("subject", "")
    job_id = message.get("job_id")
    if job_id and job_id in _jobs:
        _jobs[job_id]["status"] = "completed"
    return {"success": True, "to": to, "subject": subject}


async def worker_report_generate(message: dict) -> dict:
    """Generate reports.

    Extract type and format from the message.
    Return {success: true, type, format}.
    """
    report_type = message.get("type", "")
    report_format = message.get("format", "")
    return {"success": True, "type": report_type, "format": report_format}


# ---------------------------------------------------------------------------
# Cron Jobs (x-cron-jobs)
# ---------------------------------------------------------------------------

_mock_db: dict[str, list[dict]] = {
    "sessions": [],
    "users": [],
}


async def cron_cleanup_expired():
    """Daily at 2 AM (0 2 * * *).

    Query expired sessions from the database and remove them.
    """
    now = datetime.now(timezone.utc)
    expired = [
        s for s in _mock_db["sessions"]
        if s.get("expires_at", "") < now.isoformat()
    ]
    for session in expired:
        _mock_db["sessions"].remove(session)
    return {"cleaned": len(expired)}


async def cron_weekly_digest():
    """Weekly on Monday at 8 AM (0 8 * * 1).

    Query users with digest_enabled=true from the database.
    """
    eligible_users = [
        u for u in _mock_db["users"]
        if u.get("digest_enabled") is True
    ]
    return {"users": eligible_users, "count": len(eligible_users)}


# ---------------------------------------------------------------------------
# Event Handlers (x-event-handlers)
# ---------------------------------------------------------------------------


async def handle_user_created(event: dict) -> dict:
    """When a new user signs up, capture their email and name
    for a welcome notification. Return {queued: true, email, name}.
    """
    email = event.get("email", "")
    name = event.get("name", "")
    _event_log.append({
        "event": "user.created",
        "email": email,
        "name": name,
        "processed_at": datetime.now(timezone.utc).isoformat(),
    })
    return {"queued": True, "email": email, "name": name}


# ---------------------------------------------------------------------------
# Custom exception handler to match the Error schema
# ---------------------------------------------------------------------------


@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": exc.detail},
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
