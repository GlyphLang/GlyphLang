import time
import uuid
from datetime import datetime, timezone
from typing import Optional

from fastapi import Depends, FastAPI, HTTPException, Request, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from jose import JWTError, jwt
from passlib.hash import bcrypt
from pydantic import BaseModel


SECRET_KEY = "secret"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE = 3600
REFRESH_TOKEN_EXPIRE = 86400 * 7

app = FastAPI()
security = HTTPBearer()


# --- Models ---


class User(BaseModel):
    id: str
    email: str
    name: str
    password_hash: str
    role: str
    created_at: str
    last_login: Optional[str] = None


class UserOut(BaseModel):
    id: str
    email: str
    name: str
    role: str
    created_at: str
    last_login: Optional[str] = None


class RegisterInput(BaseModel):
    email: str
    name: str
    password: str


class LoginInput(BaseModel):
    email: str
    password: str


class TokenRefreshInput(BaseModel):
    refresh_token: str


class AuthResponse(BaseModel):
    token: str
    refresh_token: str
    expires_in: int
    user: UserOut


class TokenResponse(BaseModel):
    token: str
    refresh_token: str
    expires_in: int


# --- Database ---


class Database:
    def __init__(self) -> None:
        self._users: dict[str, User] = {}
        self._sessions: dict[str, dict] = {}

    def find_user_by_email(self, email: str) -> Optional[User]:
        for user in self._users.values():
            if user.email == email:
                return user
        return None

    def get_user(self, user_id: str) -> Optional[User]:
        return self._users.get(user_id)

    def create_user(self, email: str, name: str, password_hash: str) -> User:
        user = User(
            id=str(uuid.uuid4()),
            email=email,
            name=name,
            password_hash=password_hash,
            role="user",
            created_at=datetime.now(timezone.utc).isoformat(),
        )
        self._users[user.id] = user
        return user

    def update_last_login(self, user_id: str) -> None:
        user = self._users.get(user_id)
        if user:
            self._users[user_id] = user.model_copy(
                update={"last_login": datetime.now(timezone.utc).isoformat()}
            )

    def list_users(self) -> list[User]:
        return list(self._users.values())

    def create_session(self, user_id: str, refresh_token: str) -> None:
        self._sessions[refresh_token] = {
            "user_id": user_id,
            "created_at": time.time(),
        }

    def get_session(self, refresh_token: str) -> Optional[dict]:
        return self._sessions.get(refresh_token)

    def delete_session(self, refresh_token: str) -> bool:
        if refresh_token in self._sessions:
            del self._sessions[refresh_token]
            return True
        return False

    def delete_sessions_for_user(self, user_id: str) -> None:
        to_delete = [
            k for k, v in self._sessions.items() if v["user_id"] == user_id
        ]
        for k in to_delete:
            del self._sessions[k]

    def query_expired_sessions(self) -> list[str]:
        now = time.time()
        expired = [
            k
            for k, v in self._sessions.items()
            if now - v["created_at"] > REFRESH_TOKEN_EXPIRE
        ]
        for k in expired:
            del self._sessions[k]
        return expired


db = Database()


def get_db() -> Database:
    return db


# --- Rate Limiting ---


rate_limit_store: dict[str, list[float]] = {}


def rate_limit(key: str, max_requests: int, window: int = 60) -> None:
    now = time.time()
    if key not in rate_limit_store:
        rate_limit_store[key] = []
    rate_limit_store[key] = [t for t in rate_limit_store[key] if now - t < window]
    if len(rate_limit_store[key]) >= max_requests:
        raise HTTPException(status_code=429, detail="Rate limit exceeded")
    rate_limit_store[key].append(now)


# --- Token Helpers ---


def sign_token(user: User) -> tuple[str, str, int]:
    now = int(time.time())
    access_payload = {
        "sub": user.id,
        "email": user.email,
        "role": user.role,
        "exp": now + ACCESS_TOKEN_EXPIRE,
    }
    access_token = jwt.encode(access_payload, SECRET_KEY, algorithm=ALGORITHM)
    refresh_token = str(uuid.uuid4())
    db.create_session(user.id, refresh_token)
    return access_token, refresh_token, ACCESS_TOKEN_EXPIRE


def user_to_out(user: User) -> UserOut:
    return UserOut(
        id=user.id,
        email=user.email,
        name=user.name,
        role=user.role,
        created_at=user.created_at,
        last_login=user.last_login,
    )


# --- Auth Dependencies ---


def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    database: Database = Depends(get_db),
) -> User:
    try:
        payload = jwt.decode(credentials.credentials, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: Optional[str] = payload.get("sub")
        if user_id is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token"
            )
    except JWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token"
        )
    user = database.get_user(user_id)
    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="User not found"
        )
    return user


def require_admin(user: User = Depends(get_current_user)) -> User:
    if user.role != "admin":
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Admin access required"
        )
    return user


# --- Public Endpoints ---


@app.post("/auth/register", response_model=AuthResponse, status_code=201)
async def register(
    data: RegisterInput,
    request: Request,
    database: Database = Depends(get_db),
) -> AuthResponse:
    rate_limit(f"register:{request.client.host}", max_requests=10)
    existing = database.find_user_by_email(data.email)
    if existing:
        raise HTTPException(status_code=409, detail="Email already registered")
    password_hash = bcrypt.hash(data.password)
    user = database.create_user(
        email=data.email, name=data.name, password_hash=password_hash
    )
    access_token, refresh_token, expires_in = sign_token(user)
    return AuthResponse(
        token=access_token,
        refresh_token=refresh_token,
        expires_in=expires_in,
        user=user_to_out(user),
    )


@app.post("/auth/login", response_model=AuthResponse)
async def login(
    data: LoginInput,
    request: Request,
    database: Database = Depends(get_db),
) -> AuthResponse:
    rate_limit(f"login:{request.client.host}", max_requests=20)
    user = database.find_user_by_email(data.email)
    if user is None or not bcrypt.verify(data.password, user.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")
    database.update_last_login(user.id)
    user = database.get_user(user.id)
    access_token, refresh_token, expires_in = sign_token(user)
    return AuthResponse(
        token=access_token,
        refresh_token=refresh_token,
        expires_in=expires_in,
        user=user_to_out(user),
    )


# --- Authenticated Endpoints ---


@app.post("/auth/refresh", response_model=TokenResponse)
async def refresh_token(
    data: TokenRefreshInput,
    _user: User = Depends(get_current_user),
    database: Database = Depends(get_db),
) -> TokenResponse:
    session = database.get_session(data.refresh_token)
    if session is None:
        raise HTTPException(status_code=401, detail="Invalid refresh token")
    user = database.get_user(session["user_id"])
    if user is None:
        raise HTTPException(status_code=401, detail="User not found")
    database.delete_session(data.refresh_token)
    access_token, new_refresh_token, expires_in = sign_token(user)
    return TokenResponse(
        token=access_token,
        refresh_token=new_refresh_token,
        expires_in=expires_in,
    )


@app.get("/auth/me", response_model=UserOut)
async def get_me(user: User = Depends(get_current_user)) -> UserOut:
    return user_to_out(user)


@app.post("/auth/logout")
async def logout(
    user: User = Depends(get_current_user),
    database: Database = Depends(get_db),
) -> dict:
    database.delete_sessions_for_user(user.id)
    return {"logged_out": True}


# --- Admin Endpoints ---


@app.get("/auth/admin/users", response_model=list[UserOut])
async def list_users(
    request: Request,
    user: User = Depends(require_admin),
    database: Database = Depends(get_db),
) -> list[UserOut]:
    rate_limit(f"admin_users:{request.client.host}", max_requests=50)
    return [user_to_out(u) for u in database.list_users()]


# --- Background Tasks ---


async def cleanup_expired_sessions() -> list[str]:
    return db.query_expired_sessions()
