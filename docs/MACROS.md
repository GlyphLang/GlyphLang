
```glyph
# Define reusable patterns with macros
macro! validate_required(field) {
  if field == null {
    > {error: "field is required", status: 400}
  }
}

macro! json_response(data, status) {
  > {data: data, status: status, timestamp: now()}
}

# CRUD pattern - what macro expansion produces:
@ GET /users {
  > db.query("SELECT * FROM users")
}

@ GET /users/:id {
  > db.query("SELECT * FROM users WHERE id = ?", id)
}

@ POST /users {
  > db.insert("users", input)
}

@ DELETE /users/:id {
  > db.query("DELETE FROM users WHERE id = ?", id)
}
```