from datetime import datetime, timezone
from typing import Optional

from fastapi import Depends, FastAPI, HTTPException
from pydantic import BaseModel


app = FastAPI()


class Todo(BaseModel):
    id: int
    title: str
    completed: bool
    created_at: Optional[str] = None


class CreateTodoInput(BaseModel):
    title: str
    completed: bool = False


class UpdateTodoInput(BaseModel):
    title: Optional[str] = None
    completed: Optional[bool] = None


class Database:
    def __init__(self):
        self._todos: dict[int, Todo] = {}
        self._next_id: int = 1

    def list_todos(self) -> list[Todo]:
        return list(self._todos.values())

    def get_todo(self, todo_id: int) -> Optional[Todo]:
        return self._todos.get(todo_id)

    def create_todo(self, data: CreateTodoInput) -> Todo:
        todo = Todo(
            id=self._next_id,
            title=data.title,
            completed=data.completed,
            created_at=datetime.now(timezone.utc).isoformat(),
        )
        self._todos[self._next_id] = todo
        self._next_id += 1
        return todo

    def update_todo(self, todo_id: int, data: UpdateTodoInput) -> Optional[Todo]:
        todo = self._todos.get(todo_id)
        if todo is None:
            return None
        updated = todo.model_copy(
            update={k: v for k, v in data.model_dump().items() if v is not None}
        )
        self._todos[todo_id] = updated
        return updated

    def delete_todo(self, todo_id: int) -> bool:
        if todo_id in self._todos:
            del self._todos[todo_id]
            return True
        return False


db = Database()


def get_db() -> Database:
    return db


@app.get("/api/todos")
async def list_todos(database: Database = Depends(get_db)) -> list[Todo]:
    return database.list_todos()


@app.get("/api/todos/{todo_id}")
async def get_todo(todo_id: int, database: Database = Depends(get_db)) -> Todo:
    todo = database.get_todo(todo_id)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo


@app.post("/api/todos", status_code=201)
async def create_todo(
    data: CreateTodoInput, database: Database = Depends(get_db)
) -> Todo:
    return database.create_todo(data)


@app.put("/api/todos/{todo_id}")
async def update_todo(
    todo_id: int, data: UpdateTodoInput, database: Database = Depends(get_db)
) -> Todo:
    todo = database.update_todo(todo_id, data)
    if todo is None:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo


@app.delete("/api/todos/{todo_id}")
async def delete_todo(
    todo_id: int, database: Database = Depends(get_db)
) -> dict:
    if not database.delete_todo(todo_id):
        raise HTTPException(status_code=404, detail="Todo not found")
    return {"detail": "Todo deleted"}
