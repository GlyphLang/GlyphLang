from datetime import datetime, timezone
from typing import Optional

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field

app = FastAPI(
    title="Todo CRUD API",
    description="REST API for managing Todo items with full CRUD operations.",
    version="1.0.0",
)

# ---------------------------------------------------------------------------
# Pydantic models
# ---------------------------------------------------------------------------

class CreateTodoInput(BaseModel):
    title: str
    completed: bool = False


class UpdateTodoInput(BaseModel):
    title: Optional[str] = None
    completed: Optional[bool] = None


class Todo(BaseModel):
    id: int
    title: str
    completed: bool
    created_at: Optional[str] = None


class Error(BaseModel):
    error: str


# ---------------------------------------------------------------------------
# In-memory store
# ---------------------------------------------------------------------------

_todos: dict[int, Todo] = {}
_next_id: int = 1


def _generate_id() -> int:
    global _next_id
    current = _next_id
    _next_id += 1
    return current


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@app.get("/api/todos", response_model=list[Todo])
async def list_todos():
    """List all todos from the database."""
    return list(_todos.values())


@app.post("/api/todos", response_model=Todo, status_code=201)
async def create_todo(body: CreateTodoInput):
    """Create a new todo."""
    todo_id = _generate_id()
    todo = Todo(
        id=todo_id,
        title=body.title,
        completed=body.completed,
        created_at=datetime.now(timezone.utc).isoformat(),
    )
    _todos[todo_id] = todo
    return todo


@app.get("/api/todos/{id}", response_model=Todo, responses={404: {"model": Error}})
async def get_todo(id: int):
    """Get a single todo by ID."""
    todo = _todos.get(id)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo


@app.put("/api/todos/{id}", response_model=Todo, responses={404: {"model": Error}})
async def update_todo(id: int, body: UpdateTodoInput):
    """Update an existing todo's title and/or completed status."""
    todo = _todos.get(id)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")

    if body.title is not None:
        todo.title = body.title
    if body.completed is not None:
        todo.completed = body.completed

    _todos[id] = todo
    return todo


@app.delete("/api/todos/{id}", responses={404: {"model": Error}})
async def delete_todo(id: int):
    """Delete a todo by ID."""
    if id not in _todos:
        raise HTTPException(status_code=404, detail="Todo not found")

    del _todos[id]
    return JSONResponse(content={"deleted": True})


# ---------------------------------------------------------------------------
# Custom exception handler to match the Error schema
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
