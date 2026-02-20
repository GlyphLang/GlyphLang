import time
import uuid
from typing import Optional

from fastapi import Depends, FastAPI, HTTPException, WebSocket, WebSocketDisconnect, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from jose import JWTError, jwt
from pydantic import BaseModel


SECRET_KEY = "secret"
ALGORITHM = "HS256"

app = FastAPI()
security = HTTPBearer()


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


# --- Database ---

class Database:
    def __init__(self) -> None:
        self._rooms: dict[str, Room] = {}
        self._messages: dict[str, list[Message]] = {}

    def create_room(self, name: str, created_by: str) -> Room:
        room = Room(
            id=str(uuid.uuid4()),
            name=name,
            created_by=created_by,
            created_at=int(time.time()),
        )
        self._rooms[room.id] = room
        self._messages[room.id] = []
        return room

    def list_rooms(self) -> list[Room]:
        return list(self._rooms.values())

    def get_messages(self, room_id: str) -> Optional[list[Message]]:
        if room_id not in self._rooms:
            return None
        return self._messages.get(room_id, [])

    def add_message(self, room: str, sender: str, content: str) -> Message:
        message = Message(
            id=str(uuid.uuid4()),
            room=room,
            sender=sender,
            content=content,
            timestamp=int(time.time()),
        )
        if room not in self._messages:
            self._messages[room] = []
        self._messages[room].append(message)
        return message


db = Database()


def get_db() -> Database:
    return db


# --- Auth ---

def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
) -> str:
    try:
        payload = jwt.decode(credentials.credentials, SECRET_KEY, algorithms=[ALGORITHM])
        user: Optional[str] = payload.get("sub")
        if user is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token"
            )
        return user
    except JWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token"
        )


# --- Connection Manager ---

class ConnectionManager:
    def __init__(self) -> None:
        self._connections: list[WebSocket] = []

    async def connect(self, websocket: WebSocket) -> None:
        await websocket.accept()
        self._connections.append(websocket)

    def disconnect(self, websocket: WebSocket) -> None:
        self._connections.remove(websocket)

    async def broadcast(self, message: dict) -> None:
        for connection in self._connections:
            await connection.send_json(message)


manager = ConnectionManager()


# --- REST Routes ---

@app.post("/api/rooms", status_code=201)
async def create_room(
    data: CreateRoomInput,
    user: str = Depends(get_current_user),
    database: Database = Depends(get_db),
) -> Room:
    return database.create_room(name=data.name, created_by=user)


@app.get("/api/rooms")
async def list_rooms(
    user: str = Depends(get_current_user),
    database: Database = Depends(get_db),
) -> list[Room]:
    return database.list_rooms()


@app.get("/api/rooms/{room_id}/messages")
async def get_messages(
    room_id: str,
    user: str = Depends(get_current_user),
    database: Database = Depends(get_db),
) -> list[Message]:
    messages = database.get_messages(room_id)
    if messages is None:
        raise HTTPException(status_code=404, detail="Room not found")
    return messages


# --- WebSocket Route ---

@app.websocket("/chat")
async def chat(websocket: WebSocket) -> None:
    await manager.connect(websocket)
    sender = "anonymous"

    await manager.broadcast({
        "type": "system",
        "content": "User joined",
        "timestamp": int(time.time()),
    })

    try:
        while True:
            data = await websocket.receive_json()
            room = data.get("room", "")
            content = data.get("content", "")
            sender = data.get("sender", sender)

            message = db.add_message(room=room, sender=sender, content=content)

            await websocket.send_json({
                "type": "ack",
                "id": message.id,
                "timestamp": message.timestamp,
            })

            await manager.broadcast({
                "type": "message",
                "sender": message.sender,
                "content": message.content,
                "timestamp": message.timestamp,
            })
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast({
            "type": "system",
            "content": "User left",
            "timestamp": int(time.time()),
        })
