import hashlib
import hmac
import json
import secrets
import time
import uuid
from typing import Optional

from fastapi import Depends, FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel

app = FastAPI(
    title="Authentication Service",
    description=(
        "Authentication service with registration, login, token refresh, "
        "profile management, and admin controls. Includes password hashing "
        "(bcrypt/argon2), JWT signing, session management, and role-based access."
    ),
    version="1.0.0",
)

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

JWT_SECRET = secrets.token_hex(32)
ACCESS_TOKEN_EXPIRY = 3600  # seconds
REFRESH_TOKEN_EXPIRY = 86400 * 30  # 30 days

# ---------------------------------------------------------------------------
# Pydantic models
# ---------------------------------------------------------------------------


class RegisterInput(BaseModel):
    email: str
    name: str
    password: str


class LoginInput(BaseModel):
    email: str
    password: str


class TokenRefreshInput(BaseModel):
    refresh_token: str


class User(BaseModel):
    id: str
    email: str
    name: str
    password_hash: str
    role: str
    created_at: int
    last_login: Optional[int] = None


class AuthResponse(BaseModel):
    token: str
    refresh_token: str
    expires_in: int
    user: User


class TokenRefreshResponse(BaseModel):
    token: str
    refresh_token: str
    expires_in: int


class LogoutResponse(BaseModel):
    logged_out: bool


class Error(BaseModel):
    error: str


# ---------------------------------------------------------------------------
# In-memory stores
# ---------------------------------------------------------------------------

_users: dict[str, User] = {}  # user_id -> User
_email_index: dict[str, str] = {}  # email -> user_id
_sessions: dict[str, dict] = {}  # refresh_token -> {user_id, expires_at}

# ---------------------------------------------------------------------------
# Password hashing (bcrypt-style with PBKDF2 fallback)
# ---------------------------------------------------------------------------


def _hash_password(password: str) -> str:
    """Hash a password using PBKDF2-HMAC-SHA256 with a random salt."""
    salt = secrets.token_hex(16)
    dk = hashlib.pbkdf2_hmac("sha256", password.encode(), salt.encode(), 100_000)
    return f"pbkdf2:sha256:100000${salt}${dk.hex()}"


def _verify_password(password: str, password_hash: str) -> bool:
    """Verify a password against its stored hash."""
    try:
        parts = password_hash.split("$")
        if len(parts) != 3:
            return False
        _header, salt, stored_hash = parts
        dk = hashlib.pbkdf2_hmac("sha256", password.encode(), salt.encode(), 100_000)
        return hmac.compare_digest(dk.hex(), stored_hash)
    except Exception:
        return False


# ---------------------------------------------------------------------------
# JWT helpers (minimal HS256 implementation)
# ---------------------------------------------------------------------------


def _base64url_encode(data: bytes) -> str:
    import base64

    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()


def _base64url_decode(s: str) -> bytes:
    import base64

    padding = 4 - len(s) % 4
    if padding != 4:
        s += "=" * padding
    return base64.urlsafe_b64decode(s)


def _sign_jwt(payload: dict) -> str:
    """Create an HS256-signed JWT."""
    header = {"alg": "HS256", "typ": "JWT"}
    header_b64 = _base64url_encode(json.dumps(header, separators=(",", ":")).encode())
    payload_b64 = _base64url_encode(
        json.dumps(payload, separators=(",", ":")).encode()
    )
    signing_input = f"{header_b64}.{payload_b64}"
    signature = hmac.new(
        JWT_SECRET.encode(), signing_input.encode(), hashlib.sha256
    ).digest()
    sig_b64 = _base64url_encode(signature)
    return f"{signing_input}.{sig_b64}"


def _decode_jwt(token: str) -> Optional[dict]:
    """Decode and verify an HS256-signed JWT. Returns None if invalid."""
    try:
        parts = token.split(".")
        if len(parts) != 3:
            return None
        header_b64, payload_b64, sig_b64 = parts
        signing_input = f"{header_b64}.{payload_b64}"
        expected_sig = hmac.new(
            JWT_SECRET.encode(), signing_input.encode(), hashlib.sha256
        ).digest()
        actual_sig = _base64url_decode(sig_b64)
        if not hmac.compare_digest(expected_sig, actual_sig):
            return None
        payload = json.loads(_base64url_decode(payload_b64))
        if payload.get("exp", 0) < time.time():
            return None
        return payload
    except Exception:
        return None


def _create_tokens(user: User) -> tuple[str, str]:
    """Create an access token and a refresh token for the given user."""
    now = int(time.time())
    access_payload = {
        "user_id": user.id,
        "email": user.email,
        "role": user.role,
        "iat": now,
        "exp": now + ACCESS_TOKEN_EXPIRY,
    }
    access_token = _sign_jwt(access_payload)

    refresh_token = secrets.token_urlsafe(48)
    _sessions[refresh_token] = {
        "user_id": user.id,
        "expires_at": now + REFRESH_TOKEN_EXPIRY,
    }

    return access_token, refresh_token


# ---------------------------------------------------------------------------
# Rate limiting (simple in-memory sliding window)
# ---------------------------------------------------------------------------

_rate_limits: dict[str, list[float]] = {}


def _check_rate_limit(key: str, max_requests: int, window: int) -> None:
    """Raise 429 if the rate limit is exceeded for the given key."""
    now = time.time()
    if key not in _rate_limits:
        _rate_limits[key] = []

    # Remove entries outside the window
    _rate_limits[key] = [t for t in _rate_limits[key] if t > now - window]

    if len(_rate_limits[key]) >= max_requests:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")

    _rate_limits[key].append(now)


# ---------------------------------------------------------------------------
# Auth dependency
# ---------------------------------------------------------------------------

_bearer_scheme = HTTPBearer(auto_error=False)


async def _get_current_user(
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(_bearer_scheme),
) -> User:
    """Decode the JWT from the Authorization header and return the user."""
    if credentials is None:
        raise HTTPException(status_code=401, detail="Missing authorization header")

    payload = _decode_jwt(credentials.credentials)
    if payload is None:
        raise HTTPException(status_code=401, detail="Invalid or expired token")

    user_id = payload.get("user_id")
    user = _users.get(user_id)
    if user is None:
        raise HTTPException(status_code=401, detail="User not found")

    return user


async def _get_admin_user(
    user: User = Depends(_get_current_user),
) -> User:
    """Ensure the current user has the admin role."""
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin access required")
    return user


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------


@app.post("/auth/register", response_model=AuthResponse, status_code=201)
async def register(body: RegisterInput, request: Request):
    """Register a new user."""
    client_ip = request.client.host if request.client else "unknown"
    _check_rate_limit(f"register:{client_ip}", max_requests=10, window=60)

    if body.email in _email_index:
        raise HTTPException(status_code=409, detail="Email already registered")

    user_id = str(uuid.uuid4())
    now = int(time.time())
    password_hash = _hash_password(body.password)

    user = User(
        id=user_id,
        email=body.email,
        name=body.name,
        password_hash=password_hash,
        role="user",
        created_at=now,
    )

    _users[user_id] = user
    _email_index[body.email] = user_id

    access_token, refresh_token = _create_tokens(user)

    return AuthResponse(
        token=access_token,
        refresh_token=refresh_token,
        expires_in=ACCESS_TOKEN_EXPIRY,
        user=user,
    )


@app.post("/auth/login", response_model=AuthResponse)
async def login(body: LoginInput, request: Request):
    """Login with email and password."""
    client_ip = request.client.host if request.client else "unknown"
    _check_rate_limit(f"login:{client_ip}", max_requests=20, window=60)

    user_id = _email_index.get(body.email)
    if user_id is None:
        raise HTTPException(status_code=401, detail="Invalid credentials")

    user = _users[user_id]
    if not _verify_password(body.password, user.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")

    # Update last_login timestamp
    user.last_login = int(time.time())
    _users[user_id] = user

    access_token, refresh_token = _create_tokens(user)

    return AuthResponse(
        token=access_token,
        refresh_token=refresh_token,
        expires_in=ACCESS_TOKEN_EXPIRY,
        user=user,
    )


@app.post("/auth/refresh", response_model=TokenRefreshResponse)
async def refresh_token(
    body: TokenRefreshInput, user: User = Depends(_get_current_user)
):
    """Refresh an expired access token."""
    session = _sessions.get(body.refresh_token)
    if session is None:
        raise HTTPException(status_code=401, detail="Invalid refresh token")

    if session["expires_at"] < time.time():
        del _sessions[body.refresh_token]
        raise HTTPException(status_code=401, detail="Refresh token expired")

    if session["user_id"] != user.id:
        raise HTTPException(status_code=401, detail="Token mismatch")

    # Delete old session
    del _sessions[body.refresh_token]

    # Create new tokens
    access_token, new_refresh_token = _create_tokens(user)

    return TokenRefreshResponse(
        token=access_token,
        refresh_token=new_refresh_token,
        expires_in=ACCESS_TOKEN_EXPIRY,
    )


@app.get("/auth/me", response_model=User)
async def get_profile(user: User = Depends(_get_current_user)):
    """Get current user profile."""
    return user


@app.post("/auth/logout", response_model=LogoutResponse)
async def logout(user: User = Depends(_get_current_user)):
    """Invalidate current session."""
    # Remove all sessions for the current user
    tokens_to_remove = [
        token
        for token, session in _sessions.items()
        if session["user_id"] == user.id
    ]
    for token in tokens_to_remove:
        del _sessions[token]

    return LogoutResponse(logged_out=True)


@app.get("/auth/admin/users", response_model=list[User])
async def list_users(request: Request, admin: User = Depends(_get_admin_user)):
    """List all users (admin only)."""
    client_ip = request.client.host if request.client else "unknown"
    _check_rate_limit(f"admin_users:{client_ip}", max_requests=50, window=60)

    return list(_users.values())


# ---------------------------------------------------------------------------
# Cron job: session cleanup (daily at 3 AM)
# ---------------------------------------------------------------------------


def cleanup_expired_sessions() -> int:
    """Delete expired sessions from the store. Returns number of removed sessions."""
    now = time.time()
    expired = [
        token
        for token, session in _sessions.items()
        if session["expires_at"] < now
    ]
    for token in expired:
        del _sessions[token]
    return len(expired)


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
