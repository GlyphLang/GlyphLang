from datetime import datetime, timezone
from typing import Optional

from fastapi import Depends, FastAPI, HTTPException
from pydantic import BaseModel


# --- Models ---

class Todo(BaseModel):
    id: int
    title: str
    completed: bool
    created_at: str


class CreateTodoInput(BaseModel):
    title: str
    completed: bool = False


class UpdateTodoInput(BaseModel):
    title: Optional[str] = None
    completed: Optional[bool] = None


# --- Database ---

class TodoCollection:
    def __init__(self) -> None:
        self._store: dict[int, Todo] = {}
        self._next_id: int = 1

    def Find(self) -> list[Todo]:
        return list(self._store.values())

    def Get(self, id: int) -> Optional[Todo]:
        return self._store.get(id)

    def Create(self, input: CreateTodoInput) -> Todo:
        todo = Todo(
            id=self._next_id,
            title=input.title,
            completed=input.completed,
            created_at=datetime.now(timezone.utc).isoformat(),
        )
        self._store[todo.id] = todo
        self._next_id += 1
        return todo

    def Update(self, id: int, input: UpdateTodoInput) -> Todo:
        existing = self._store[id]
        updated_data = existing.model_copy(
            update={k: v for k, v in input.model_dump().items() if v is not None}
        )
        self._store[id] = updated_data
        return updated_data

    def Delete(self, id: int) -> None:
        del self._store[id]


class Database:
    def __init__(self) -> None:
        self.todos = TodoCollection()


# --- App ---

app = FastAPI()

_database = Database()


def get_db() -> Database:
    return _database


# --- Routes ---

@app.get("/api/todos")
async def list_todos(db: Database = Depends(get_db)) -> list[Todo]:
    todos = db.todos.Find()
    return todos


@app.get("/api/todos/{id}")
async def get_todo(id: int, db: Database = Depends(get_db)) -> Todo | dict:
    todo = db.todos.Get(id)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo


@app.post("/api/todos", status_code=201)
async def create_todo(input: CreateTodoInput, db: Database = Depends(get_db)) -> Todo:
    todo = db.todos.Create(input)
    return todo


@app.put("/api/todos/{id}")
async def update_todo(
    id: int, input: UpdateTodoInput, db: Database = Depends(get_db)
) -> Todo | dict:
    todo = db.todos.Get(id)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")
    updated = db.todos.Update(id, input)
    return updated


@app.delete("/api/todos/{id}")
async def delete_todo(id: int, db: Database = Depends(get_db)) -> dict:
    todo = db.todos.Get(id)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")
    db.todos.Delete(id)
    return {"deleted": True}
