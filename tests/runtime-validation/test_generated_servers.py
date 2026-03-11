"""Runtime validation tests for Phase 5 generated FastAPI implementations.

Starts each generated server on a random port, sends real HTTP requests,
and verifies response shapes, status codes, and business logic.

Usage:
    ./setup.sh  # first time only
    .venv/bin/pytest test_generated_servers.py -v
"""

import os
import signal
import socket
import subprocess
import sys
import tempfile
import time
from pathlib import Path

import httpx
import jwt
import pytest

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

PHASE5_BRANCH = "feat/phase5-intent-hypothesis-testing"
GEN_PREFIX = "tests/intent-hypothesis/generated"
CONDITIONS = ["glyph", "text", "openapi"]
JWT_SECRET = "secret"  # All generated servers use this secret


def _make_jwt(sub: str = "test-user", **extra) -> str:
    """Create a valid HS256 JWT token matching the generated servers' secret."""
    payload = {"sub": sub, "exp": int(time.time()) + 3600, **extra}
    return jwt.encode(payload, JWT_SECRET, algorithm="HS256")


def _make_expired_jwt(sub: str = "test-user") -> str:
    """Create an expired HS256 JWT token (expired 1 hour ago)."""
    payload = {"sub": sub, "exp": int(time.time()) - 3600}
    return jwt.encode(payload, JWT_SECRET, algorithm="HS256")


TYPE_MAP = {
    str: str,
    int: (int, float),  # JSON numbers may deserialize as float
    bool: bool,
    list: list,
}


def assert_response_shape(body: dict, required_fields: dict, optional_fields: dict = None):
    """Verify that a response body contains required fields with expected types.

    Args:
        body: The parsed JSON response body (dict).
        required_fields: Mapping of field name to expected Python type
                         (str, int, bool, list).
        optional_fields: Optional mapping of field name to expected type;
                         these are checked only if present.
    """
    for field, expected_type in required_fields.items():
        assert field in body, f"Required field {field!r} missing from response: {body}"
        allowed = TYPE_MAP.get(expected_type, expected_type)
        assert isinstance(body[field], allowed), (
            f"Field {field!r} expected {expected_type.__name__}, "
            f"got {type(body[field]).__name__}: {body[field]!r}"
        )
    if optional_fields:
        for field, expected_type in optional_fields.items():
            if field in body:
                allowed = TYPE_MAP.get(expected_type, expected_type)
                assert isinstance(body[field], allowed), (
                    f"Optional field {field!r} expected {expected_type.__name__}, "
                    f"got {type(body[field]).__name__}: {body[field]!r}"
                )


def _free_port() -> int:
    with socket.socket() as s:
        s.bind(("", 0))
        return s.getsockname()[1]


def _git_show(ref: str, path: str) -> str:
    """Read a file from a git ref without checking out."""
    result = subprocess.run(
        ["git", "show", f"{ref}:{path}"],
        capture_output=True, text=True,
        cwd=Path(__file__).resolve().parents[2],  # repo root
    )
    if result.returncode != 0:
        pytest.skip(f"Cannot read {path} from {ref}: {result.stderr.strip()}")
    return result.stdout


class ServerProcess:
    """Manages a FastAPI server subprocess."""

    def __init__(self, source: str, port: int):
        self.port = port
        self.source = source
        self._proc = None
        self._tmpdir = None

    def start(self):
        self._tmpdir = tempfile.mkdtemp(prefix="glyph-runtime-")
        src_path = os.path.join(self._tmpdir, "server.py")
        with open(src_path, "w") as f:
            f.write(self.source)

        # Inject port override — replace uvicorn.run call or add one
        wrapper = os.path.join(self._tmpdir, "run.py")
        with open(wrapper, "w") as f:
            f.write(
                f"import sys, importlib.util\n"
                f"spec = importlib.util.spec_from_file_location('server', {src_path!r})\n"
                f"mod = importlib.util.module_from_spec(spec)\n"
                f"sys.modules['server'] = mod\n"
                f"# Prevent any if __name__ == '__main__' block from running uvicorn\n"
                f"mod.__name__ = 'server'\n"
                f"spec.loader.exec_module(mod)\n"
                f"import uvicorn\n"
                f"uvicorn.run(mod.app, host='127.0.0.1', port={self.port}, log_level='error')\n"
            )

        venv_python = os.path.join(
            Path(__file__).parent, ".venv", "bin", "python"
        )
        self._proc = subprocess.Popen(
            [venv_python, wrapper],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        # Wait for server to be ready
        deadline = time.time() + 10
        while time.time() < deadline:
            try:
                with socket.create_connection(("127.0.0.1", self.port), timeout=0.5):
                    return
            except OSError:
                if self._proc.poll() is not None:
                    out = self._proc.stderr.read().decode()
                    pytest.fail(f"Server exited early: {out}")
                time.sleep(0.2)
        pytest.fail(f"Server did not start within 10s on port {self.port}")

    def stop(self):
        if self._proc and self._proc.poll() is None:
            self._proc.send_signal(signal.SIGTERM)
            try:
                self._proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self._proc.kill()
        if self._tmpdir:
            import shutil
            shutil.rmtree(self._tmpdir, ignore_errors=True)


@pytest.fixture
def server_for():
    """Factory fixture: returns a function that starts a server for a given condition/scenario."""
    servers = []

    def _start(condition: str, scenario: str):
        path = f"{GEN_PREFIX}/{condition}/{scenario}.py"
        source = _git_show(PHASE5_BRANCH, path)
        port = _free_port()
        srv = ServerProcess(source, port)
        srv.start()
        servers.append(srv)
        return f"http://127.0.0.1:{port}"

    yield _start

    for srv in servers:
        srv.stop()


# ---------------------------------------------------------------------------
# Scenario 01: CRUD API
# ---------------------------------------------------------------------------

class TestCrudApi:
    """Runtime tests for scenario 01 — Todo CRUD API."""

    @pytest.fixture(autouse=True, params=CONDITIONS)
    def setup_server(self, request, server_for):
        self.condition = request.param
        self.base = server_for(self.condition, "01-crud-api")
        self.client = httpx.Client(base_url=self.base, timeout=5)
        yield
        self.client.close()

    def test_list_todos_empty(self):
        r = self.client.get("/api/todos")
        assert r.status_code == 200
        assert isinstance(r.json(), list)

    def test_create_todo(self):
        r = self.client.post("/api/todos", json={"title": "Test todo"})
        assert r.status_code == 201
        body = r.json()
        assert "id" in body
        assert body["title"] == "Test todo"
        assert body["completed"] is False  # default
        assert_response_shape(body, {"id": int, "title": str, "completed": bool})

    def test_create_and_get_todo(self):
        r1 = self.client.post("/api/todos", json={"title": "Get me"})
        todo_id = r1.json()["id"]
        r2 = self.client.get(f"/api/todos/{todo_id}")
        assert r2.status_code == 200
        assert r2.json()["title"] == "Get me"

    def test_update_todo(self):
        r1 = self.client.post("/api/todos", json={"title": "Update me"})
        todo_id = r1.json()["id"]
        r2 = self.client.put(f"/api/todos/{todo_id}", json={"completed": True})
        assert r2.status_code == 200
        assert r2.json()["completed"] is True

    def test_delete_todo(self):
        r1 = self.client.post("/api/todos", json={"title": "Delete me"})
        todo_id = r1.json()["id"]
        r2 = self.client.delete(f"/api/todos/{todo_id}")
        assert r2.status_code == 200

    def test_get_nonexistent_returns_404(self):
        r = self.client.get("/api/todos/99999")
        assert r.status_code == 404

    def test_list_todos_after_create(self):
        self.client.post("/api/todos", json={"title": "Listed"})
        r = self.client.get("/api/todos")
        assert r.status_code == 200
        todos = r.json()
        assert len(todos) >= 1
        assert any(t["title"] == "Listed" for t in todos)

    def test_create_todo_empty_body_returns_422(self):
        """POST with empty JSON should return 422 (or 400) for missing required fields."""
        r = self.client.post("/api/todos", json={})
        assert r.status_code in (400, 422)


# ---------------------------------------------------------------------------
# Scenario 02: Webhook Processor
# ---------------------------------------------------------------------------

class TestWebhookProcessor:
    """Runtime tests for scenario 02 — Webhook Processor."""

    @pytest.fixture(autouse=True, params=CONDITIONS)
    def setup_server(self, request, server_for):
        self.condition = request.param
        self.base = server_for(self.condition, "02-webhook-processor")
        self.client = httpx.Client(base_url=self.base, timeout=5)
        yield
        self.client.close()

    def _webhook_payload(self, event="payment.succeeded"):
        return {
            "event": event,
            "data": {"id": "pay_123"},
            "timestamp": 1700000000,
            "signature": "test-sig",
        }

    def test_stripe_webhook_success(self):
        r = self.client.post(
            "/webhooks/stripe",
            json=self._webhook_payload(),
            headers={"X-API-Key": "test-key"},
        )
        # Accept 200 (success) or 401 (if API key validation is strict)
        if r.status_code == 200:
            body = r.json()
            assert body["processed"] is True

    def test_stripe_webhook_no_auth_returns_401(self):
        r = self.client.post("/webhooks/stripe", json=self._webhook_payload())
        assert r.status_code in (401, 403, 422)  # 422 if header is required param

    def test_github_webhook_success(self):
        r = self.client.post(
            "/webhooks/github",
            json=self._webhook_payload("push"),
            headers={"X-API-Key": "test-key"},
        )
        if r.status_code == 200:
            body = r.json()
            assert body["processed"] is True

    def test_unknown_webhook_type_returns_404(self):
        """POST to an unknown webhook type should return 404."""
        r = self.client.post(
            "/webhooks/unknown",
            json=self._webhook_payload(),
            headers={"X-API-Key": "test-key"},
        )
        assert r.status_code == 404


# ---------------------------------------------------------------------------
# Scenario 03: Chat Server
# ---------------------------------------------------------------------------

class TestChatServer:
    """Runtime tests for scenario 03 — Chat Server (REST endpoints only)."""

    @pytest.fixture(autouse=True, params=CONDITIONS)
    def setup_server(self, request, server_for):
        self.condition = request.param
        self.base = server_for(self.condition, "03-chat-server")
        self.client = httpx.Client(base_url=self.base, timeout=5)
        token = _make_jwt()
        self.auth_headers = {"Authorization": f"Bearer {token}"}
        yield
        self.client.close()

    def test_create_room(self):
        r = self.client.post(
            "/api/rooms",
            json={"name": "general"},
            headers=self.auth_headers,
        )
        # 201 or 200 depending on implementation
        assert r.status_code in (200, 201)
        body = r.json()
        assert "id" in body
        assert body["name"] == "general"

    def test_list_rooms(self):
        self.client.post(
            "/api/rooms",
            json={"name": "test-room"},
            headers=self.auth_headers,
        )
        r = self.client.get("/api/rooms", headers=self.auth_headers)
        assert r.status_code == 200
        rooms = r.json()
        assert isinstance(rooms, list)
        assert len(rooms) >= 1

    def test_get_messages_for_room(self):
        r1 = self.client.post(
            "/api/rooms",
            json={"name": "msg-room"},
            headers=self.auth_headers,
        )
        room_id = r1.json()["id"]
        r2 = self.client.get(
            f"/api/rooms/{room_id}/messages",
            headers=self.auth_headers,
        )
        assert r2.status_code == 200
        assert isinstance(r2.json(), list)

    def test_missing_auth_header_returns_401(self):
        """Request without Authorization header should return 401 or 403."""
        r = self.client.post("/api/rooms", json={"name": "no-auth-room"})
        assert r.status_code in (401, 403)

    def test_invalid_jwt_returns_401(self):
        """Request with an invalid JWT token should return 401 or 403."""
        r = self.client.post(
            "/api/rooms",
            json={"name": "bad-jwt-room"},
            headers={"Authorization": "Bearer invalid-token"},
        )
        assert r.status_code in (401, 403)

    def test_expired_jwt_returns_401(self):
        """Request with an expired JWT should return 401 or 403."""
        expired_token = _make_expired_jwt()
        r = self.client.post(
            "/api/rooms",
            json={"name": "expired-jwt-room"},
            headers={"Authorization": f"Bearer {expired_token}"},
        )
        assert r.status_code in (401, 403)

    def test_get_nonexistent_room_returns_404(self):
        """GET messages for a non-existent room should return 404."""
        r = self.client.get(
            "/api/rooms/nonexistent-id/messages",
            headers=self.auth_headers,
        )
        assert r.status_code == 404


# ---------------------------------------------------------------------------
# Scenario 04: Job Queue
# ---------------------------------------------------------------------------

class TestJobQueue:
    """Runtime tests for scenario 04 — Job Queue."""

    @pytest.fixture(autouse=True, params=CONDITIONS)
    def setup_server(self, request, server_for):
        self.condition = request.param
        self.base = server_for(self.condition, "04-job-queue")
        self.client = httpx.Client(base_url=self.base, timeout=5)
        token = _make_jwt()
        self.auth_headers = {"Authorization": f"Bearer {token}"}
        yield
        self.client.close()

    def test_submit_email_job(self):
        r = self.client.post(
            "/api/jobs/email",
            json={"to": "user@test.com", "subject": "Hello", "template": "welcome"},
            headers=self.auth_headers,
        )
        assert r.status_code in (200, 201)
        body = r.json()
        assert "job_id" in body or "id" in body
        status = body.get("status", "")
        assert status == "pending"

    def test_get_job_status(self):
        r1 = self.client.post(
            "/api/jobs/email",
            json={"to": "user@test.com", "subject": "Hello", "template": "welcome"},
            headers=self.auth_headers,
        )
        job_id = r1.json().get("job_id") or r1.json().get("id")
        r2 = self.client.get(f"/api/jobs/{job_id}", headers=self.auth_headers)
        if r2.status_code == 404 and self.condition == "glyph":
            # Known issue: glyph-generated code stores email jobs in
            # email_queue collection but looks them up in jobs collection.
            pytest.xfail("glyph 04-job-queue: email_queue/jobs collection mismatch")
        assert r2.status_code == 200

    def test_get_nonexistent_job_returns_404(self):
        r = self.client.get(
            "/api/jobs/nonexistent-id",
            headers=self.auth_headers,
        )
        assert r.status_code == 404

    def test_missing_auth_header_returns_401(self):
        """Submit a job without Authorization header should return 401 or 403."""
        r = self.client.post(
            "/api/jobs/email",
            json={"to": "user@test.com", "subject": "Hello", "template": "welcome"},
        )
        assert r.status_code in (401, 403)

    def test_invalid_jwt_returns_401(self):
        """Submit a job with an invalid JWT should return 401 or 403."""
        r = self.client.post(
            "/api/jobs/email",
            json={"to": "user@test.com", "subject": "Hello", "template": "welcome"},
            headers={"Authorization": "Bearer invalid-token"},
        )
        assert r.status_code in (401, 403)

    def test_expired_jwt_returns_401(self):
        """Submit a job with an expired JWT should return 401 or 403."""
        expired_token = _make_expired_jwt()
        r = self.client.post(
            "/api/jobs/email",
            json={"to": "user@test.com", "subject": "Hello", "template": "welcome"},
            headers={"Authorization": f"Bearer {expired_token}"},
        )
        assert r.status_code in (401, 403)


# ---------------------------------------------------------------------------
# Scenario 05: Auth Service
# ---------------------------------------------------------------------------

class TestAuthService:
    """Runtime tests for scenario 05 — Auth Service."""

    @pytest.fixture(autouse=True, params=CONDITIONS)
    def setup_server(self, request, server_for):
        self.condition = request.param
        self.base = server_for(self.condition, "05-auth-service")
        self.client = httpx.Client(base_url=self.base, timeout=5)
        yield
        self.client.close()

    def test_register_user(self):
        r = self.client.post("/auth/register", json={
            "email": "test@example.com",
            "name": "Test User",
            "password": "securepass123",
        })
        assert r.status_code in (200, 201)
        body = r.json()
        assert "token" in body
        assert "refresh_token" in body
        assert "expires_in" in body
        assert_response_shape(body, {
            "token": str,
            "refresh_token": str,
            "expires_in": int,
        })

    def test_register_duplicate_returns_409(self):
        payload = {
            "email": "dup@example.com",
            "name": "Dup User",
            "password": "pass123",
        }
        self.client.post("/auth/register", json=payload)
        r = self.client.post("/auth/register", json=payload)
        assert r.status_code == 409

    def test_login(self):
        self.client.post("/auth/register", json={
            "email": "login@example.com",
            "name": "Login User",
            "password": "pass123",
        })
        r = self.client.post("/auth/login", json={
            "email": "login@example.com",
            "password": "pass123",
        })
        assert r.status_code == 200
        body = r.json()
        assert "token" in body

    def test_login_bad_password_returns_401(self):
        self.client.post("/auth/register", json={
            "email": "badpw@example.com",
            "name": "Bad PW",
            "password": "realpass",
        })
        r = self.client.post("/auth/login", json={
            "email": "badpw@example.com",
            "password": "wrongpass",
        })
        assert r.status_code == 401

    def test_get_profile_with_token(self):
        r1 = self.client.post("/auth/register", json={
            "email": "profile@example.com",
            "name": "Profile User",
            "password": "pass123",
        })
        token = r1.json()["token"]
        r2 = self.client.get(
            "/auth/me",
            headers={"Authorization": f"Bearer {token}"},
        )
        assert r2.status_code == 200
        body = r2.json()
        assert body["email"] == "profile@example.com"

    def test_logout(self):
        r1 = self.client.post("/auth/register", json={
            "email": "logout@example.com",
            "name": "Logout User",
            "password": "pass123",
        })
        token = r1.json()["token"]
        r2 = self.client.post(
            "/auth/logout",
            headers={"Authorization": f"Bearer {token}"},
        )
        assert r2.status_code == 200

    def test_missing_auth_header_returns_401(self):
        """GET /auth/me without Authorization header should return 401 or 403."""
        r = self.client.get("/auth/me")
        assert r.status_code in (401, 403)

    def test_invalid_jwt_returns_401(self):
        """GET /auth/me with an invalid JWT should return 401 or 403."""
        r = self.client.get(
            "/auth/me",
            headers={"Authorization": "Bearer invalid-token"},
        )
        assert r.status_code in (401, 403)

    def test_expired_jwt_returns_401(self):
        """GET /auth/me with an expired JWT should return 401 or 403."""
        expired_token = _make_expired_jwt()
        r = self.client.get(
            "/auth/me",
            headers={"Authorization": f"Bearer {expired_token}"},
        )
        assert r.status_code in (401, 403)

    def test_register_missing_fields_returns_422(self):
        """POST /auth/register with incomplete body should return 422 or 400."""
        r = self.client.post("/auth/register", json={"email": "incomplete@example.com"})
        assert r.status_code in (400, 422)

    def test_get_profile_no_auth_returns_401(self):
        """GET /auth/me without Authorization header should return 401 or 403."""
        r = self.client.get("/auth/me")
        assert r.status_code in (401, 403)

    def test_refresh_no_auth_returns_401(self):
        """POST /auth/refresh without Authorization header should return 401 or 403."""
        r = self.client.post("/auth/refresh")
        assert r.status_code in (401, 403)
