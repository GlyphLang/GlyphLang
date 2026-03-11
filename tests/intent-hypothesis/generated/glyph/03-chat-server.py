import time
import uuid

from fastapi import Depends, FastAPI, HTTPException, WebSocket, WebSocketDisconnect
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel

import jwt as pyjwt


# --- Models ---

class Message(BaseModel):
    id: str
    room: str
    sender: str
    content: str
    timestamp: int


class Room(BaseModel):
    id: str
    name: str
    created_by: str
    created_at: int


class CreateRoomInput(BaseModel):
    name: str


class AuthUser(BaseModel):
    id: str


# --- Database ---

class RoomCollection:
    def __init__(self) -> None:
        self._store: dict[str, Room] = {}

    def Create(self, data: dict) -> Room:
        room = Room(**data)
        self._store[room.id] = room
        return room

    def Find(self) -> list[Room]:
        return list(self._store.values())


class MessageCollection:
    def __init__(self) -> None:
        self._store: list[Message] = []

    def Create(self, data: dict) -> Message:
        message = Message(**data)
        self._store.append(message)
        return message

    def Where(self, filters: dict) -> list[Message]:
        results = self._store
        for key, value in filters.items():
            results = [m for m in results if getattr(m, key) == value]
        return results


class Database:
    def __init__(self) -> None:
        self.rooms = RoomCollection()
        self.messages = MessageCollection()


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


def now() -> int:
    return int(time.time())


# --- App ---

app = FastAPI()

_database = Database()


def get_db() -> Database:
    return _database


# --- WebSocket Manager ---

class ConnectionManager:
    def __init__(self) -> None:
        self._connections: list[WebSocket] = []

    async def connect(self, websocket: WebSocket) -> None:
        await websocket.accept()
        self._connections.append(websocket)

    def disconnect(self, websocket: WebSocket) -> None:
        self._connections.remove(websocket)

    async def broadcast(self, message: str) -> None:
        for connection in self._connections:
            await connection.send_text(message)


manager = ConnectionManager()


# --- REST Routes ---

@app.post("/api/rooms", status_code=201)
async def create_room(
    input: CreateRoomInput,
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> Room:
    room = db.rooms.Create({
        "id": generate_id(),
        "name": input.name,
        "created_by": user.id,
        "created_at": now(),
    })
    return room


@app.get("/api/rooms")
async def list_rooms(
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> list[Room]:
    rooms = db.rooms.Find()
    return rooms


@app.get("/api/rooms/{room_id}/messages")
async def list_messages(
    room_id: str,
    user: AuthUser = Depends(auth),
    db: Database = Depends(get_db),
) -> list[Message]:
    messages = db.messages.Where({"room": room_id})
    return messages


# --- WebSocket Route ---

@app.websocket("/chat")
async def chat(websocket: WebSocket) -> None:
    await manager.connect(websocket)
    await manager.broadcast("User joined the chat")
    try:
        while True:
            data = await websocket.receive_text()
            await manager.broadcast(data)
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast("User left the chat")
