import time
import uuid
from typing import List

from fastapi import (
    Depends,
    FastAPI,
    HTTPException,
    WebSocket,
    WebSocketDisconnect,
    status,
)
from fastapi.responses import JSONResponse
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel

app = FastAPI(
    title="Real-Time Chat Server",
    description=(
        "Chat server with REST endpoints for room management and WebSocket "
        "for real-time messaging. All REST endpoints require JWT authentication."
    ),
    version="1.0.0",
)

# ---------------------------------------------------------------------------
# Security
# ---------------------------------------------------------------------------

bearer_scheme = HTTPBearer()


def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(bearer_scheme),
) -> str:
    """Extract and validate the JWT bearer token.

    In a production system this would decode and verify the JWT.  Here we
    accept any non-empty token and treat the token value itself as the
    username for simplicity.
    """
    token = credentials.credentials
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid or missing token",
        )
    return token


# ---------------------------------------------------------------------------
# Pydantic models
# ---------------------------------------------------------------------------


class CreateRoomInput(BaseModel):
    name: str


class Room(BaseModel):
    id: str
    name: str
    created_by: str
    created_at: int


class Message(BaseModel):
    id: str
    room: str
    sender: str
    content: str
    timestamp: int


# ---------------------------------------------------------------------------
# In-memory stores
# ---------------------------------------------------------------------------

_rooms: dict[str, Room] = {}
_messages: dict[str, list[Message]] = {}  # room_id -> [Message]


# ---------------------------------------------------------------------------
# REST routes
# ---------------------------------------------------------------------------


@app.post("/api/rooms", response_model=Room, status_code=201)
async def create_room(
    body: CreateRoomInput,
    user: str = Depends(get_current_user),
):
    """Create a new chat room."""
    room_id = str(uuid.uuid4())
    room = Room(
        id=room_id,
        name=body.name,
        created_by=user,
        created_at=int(time.time()),
    )
    _rooms[room_id] = room
    _messages[room_id] = []
    return room


@app.get("/api/rooms", response_model=List[Room])
async def list_rooms(user: str = Depends(get_current_user)):
    """List all chat rooms."""
    return list(_rooms.values())


@app.get("/api/rooms/{room_id}/messages", response_model=List[Message])
async def list_messages(
    room_id: str,
    user: str = Depends(get_current_user),
):
    """Get message history for a specific room."""
    if room_id not in _rooms:
        raise HTTPException(status_code=404, detail="Room not found")
    return _messages.get(room_id, [])


# ---------------------------------------------------------------------------
# WebSocket – real-time messaging
# ---------------------------------------------------------------------------


class ConnectionManager:
    """Manages active WebSocket connections and broadcasts."""

    def __init__(self):
        self.active_connections: list[tuple[WebSocket, str]] = []

    async def connect(self, websocket: WebSocket, username: str):
        await websocket.accept()
        self.active_connections.append((websocket, username))
        await self.broadcast(f"User {username} joined the chat")

    async def disconnect(self, websocket: WebSocket, username: str):
        self.active_connections = [
            (ws, u) for ws, u in self.active_connections if ws is not websocket
        ]
        await self.broadcast(f"User {username} left the chat")

    async def broadcast(self, message: str):
        for ws, _ in self.active_connections:
            try:
                await ws.send_text(message)
            except Exception:
                pass


manager = ConnectionManager()


@app.websocket("/chat")
async def websocket_chat(websocket: WebSocket):
    """Real-time messaging via WebSocket.

    The first text message received after connection is treated as the
    username.  Subsequent messages are broadcast to all connected clients.
    """
    # Accept connection; first message is the username identification.
    await websocket.accept()

    try:
        username = await websocket.receive_text()
    except WebSocketDisconnect:
        return

    # Re-register through the manager (already accepted, so we just track).
    manager.active_connections.append((websocket, username))
    await manager.broadcast(f"User {username} joined the chat")

    try:
        while True:
            data = await websocket.receive_text()
            await manager.broadcast(data)
    except WebSocketDisconnect:
        await manager.disconnect(websocket, username)


# ---------------------------------------------------------------------------
# Custom exception handler to match a consistent error shape
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
