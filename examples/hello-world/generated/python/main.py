# Auto-generated Python/FastAPI server from GlyphLang
# Do not edit manually

from fastapi import FastAPI, HTTPException, Depends, Request
from pydantic import BaseModel
from typing import Optional, List, Any
import uuid
import time


app = FastAPI()


@app.get("/")
async def get_root():
    return {"text": "Hello, World!", "timestamp": 1234567890}


@app.get("/hello/{name}")
async def get_hello_name(name: str):
    greeting = (("Hello, " + name) + "!")
    return {"message": greeting}


@app.get("/health")
async def get_health():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
