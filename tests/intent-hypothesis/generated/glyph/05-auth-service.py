import time
import uuid
from typing import Optional

from fastapi import Depends, FastAPI, HTTPException
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from passlib.context import CryptContext
from pydantic import BaseModel

import jwt as pyjwt


# --- Models ---

class User(BaseModel):
    id: str
    email: str
    name: str
    password_hash: str
    role: str
    created_at: int
    last_login: Optional[int] = None


class RegisterInput(BaseModel):
    email: str
    name: str
    password: str


class LoginInput(BaseModel):
    email: str
    password: str


class AuthResponse(BaseModel):
    token: str
    refresh_token: str
    expires_in: int
    user: User


class TokenRefreshInput(BaseModel):
    refresh_token: str


# --- Crypto ---

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
JWT_SECRET = "secret"
JWT_ALGORITHM = "HS256"
JWT_EXPIRY_SECONDS = 7 * 24 * 3600


def crypto_hash(password: str) -> str:
    return pwd_context.hash(password)


def crypto_verify(password: str, hashed: str) -> bool:
    return pwd_context.verify(password, hashed)


def jwt_sign(payload: dict, expiry: str) -> str:
    data = payload.copy()
    data["exp"] = int(time.time()) + JWT_EXPIRY_SECONDS
    return pyjwt.encode(data, JWT_SECRET, algorithm=JWT_ALGORITHM)


# --- Helpers ---

def generate_id() -> str:
    return str(uuid.uuid4())


def now() -> int:
    return int(time.time())


# --- Database ---

class UserCollection:
    def __init__(self) -> None:
        self._store: dict[str, User] = {}

    def Find(self) -> list[User]:
        return list(self._store.values())

    def Get(self, id: str) -> Optional[User]:
        return self._store.get(id)

    def Where(self, filters: dict) -> list[User]:
        results = list(self._store.values())
        for key, value in filters.items():
            results = [u for u in results if getattr(u, key) == value]
        return results

    def Create(self, data: dict) -> User:
        user = User(**data)
        self._store[user.id] = user
        return user

    def Update(self, id: str, fields: dict) -> User:
        existing = self._store[id]
        updated = existing.model_copy(update=fields)
        self._store[id] = updated
        return updated


class SessionCollection:
    def __init__(self) -> None:
        self._store: dict[str, dict] = {}

    def Where(self, filters: dict) -> list[dict]:
        results = list(self._store.values())
        for key, value in filters.items():
            results = [s for s in results if s.get(key) == value]
        return results

    def Create(self, data: dict) -> dict:
        self._store[data.get("id", generate_id())] = data
        return data

    def Delete(self, user_id: str) -> None:
        self._store = {
            k: v for k, v in self._store.items() if v.get("user_id") != user_id
        }


class Database:
    def __init__(self) -> None:
        self.users = UserCollection()
        self.sessions = SessionCollection()


# --- Auth Dependency ---

security = HTTPBearer()


class AuthUser(BaseModel):
    id: str
    email: str
    role: str


def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
) -> AuthUser:
    try:
        payload = pyjwt.decode(
            credentials.credentials, JWT_SECRET, algorithms=[JWT_ALGORITHM]
        )
        return AuthUser(
            id=payload["user_id"],
            email=payload.get("email", ""),
            role=payload.get("role", "user"),
        )
    except pyjwt.PyJWTError:
        raise HTTPException(status_code=401, detail="Invalid token")


def require_admin(user: AuthUser = Depends(get_current_user)) -> AuthUser:
    if user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin access required")
    return user


# --- Rate Limiting ---

RATE_LIMIT_WINDOW = 60
_rate_limit_store: dict[str, list[float]] = {}


def _check_rate_limit(key: str, max_requests: int) -> None:
    current = time.time()
    if key not in _rate_limit_store:
        _rate_limit_store[key] = []
    _rate_limit_store[key] = [
        t for t in _rate_limit_store[key] if current - t < RATE_LIMIT_WINDOW
    ]
    if len(_rate_limit_store[key]) >= max_requests:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")
    _rate_limit_store[key].append(current)


# --- App ---

app = FastAPI()

_database = Database()


def get_db() -> Database:
    return _database


# --- Routes ---

@app.post("/auth/register")
async def register(input: RegisterInput, db: Database = Depends(get_db)) -> AuthResponse:
    _check_rate_limit(f"register:{input.email}", 10)

    existing = db.users.Where({"email": input.email})
    if len(existing) > 0:
        raise HTTPException(status_code=409, detail="Email already registered")

    hashed_password = crypto_hash(input.password)

    user = db.users.Create({
        "id": generate_id(),
        "email": input.email,
        "name": input.name,
        "password_hash": hashed_password,
        "role": "user",
        "created_at": now(),
    })

    token = jwt_sign({
        "user_id": user.id,
        "email": user.email,
        "role": user.role,
    }, "7d")

    return AuthResponse(
        token=token,
        refresh_token=generate_id(),
        expires_in=3600,
        user=user,
    )


@app.post("/auth/login")
async def login(input: LoginInput, db: Database = Depends(get_db)) -> AuthResponse:
    _check_rate_limit(f"login:{input.email}", 20)

    users = db.users.Where({"email": input.email})
    if len(users) == 0:
        raise HTTPException(status_code=401, detail="Invalid credentials")

    user = users[0]

    password_valid = crypto_verify(input.password, user.password_hash)
    if not password_valid:
        raise HTTPException(status_code=401, detail="Invalid credentials")

    db.users.Update(user.id, {"last_login": now()})

    token = jwt_sign({
        "user_id": user.id,
        "email": user.email,
        "role": user.role,
    }, "7d")

    return AuthResponse(
        token=token,
        refresh_token=generate_id(),
        expires_in=3600,
        user=user,
    )


@app.post("/auth/refresh")
async def refresh_token(
    input: TokenRefreshInput,
    auth_user: AuthUser = Depends(get_current_user),
    db: Database = Depends(get_db),
) -> AuthResponse:
    sessions = db.sessions.Where({"refresh_token": input.refresh_token})
    if len(sessions) == 0:
        raise HTTPException(status_code=401, detail="Invalid refresh token")

    first_session = sessions[0]

    token = jwt_sign({
        "user_id": auth_user.id,
        "role": auth_user.role,
    }, "7d")

    session_user = db.users.Get(first_session.get("user_id", auth_user.id))
    if session_user is None:
        raise HTTPException(status_code=404, detail="User not found")

    return AuthResponse(
        token=token,
        refresh_token=generate_id(),
        expires_in=3600,
        user=session_user,
    )


@app.get("/auth/me")
async def get_me(
    auth_user: AuthUser = Depends(get_current_user),
    db: Database = Depends(get_db),
) -> User:
    user = db.users.Get(auth_user.id)
    if user is None:
        raise HTTPException(status_code=404, detail="User not found")
    return user


@app.post("/auth/logout")
async def logout(
    auth_user: AuthUser = Depends(get_current_user),
    db: Database = Depends(get_db),
) -> dict:
    db.sessions.Delete(auth_user.id)
    return {"logged_out": True}


@app.get("/auth/admin/users")
async def list_users(
    admin: AuthUser = Depends(require_admin),
    db: Database = Depends(get_db),
) -> list[User]:
    _check_rate_limit(f"admin_users:{admin.id}", 50)
    users = db.users.Find()
    return users


# --- Scheduled Task ---

async def session_cleanup(db: Database) -> dict:
    expired = db.sessions.Where({"expired": True})
    for session in expired:
        db.sessions.Delete(session.get("user_id", ""))
    return {"task": "session_cleanup", "timestamp": now()}
