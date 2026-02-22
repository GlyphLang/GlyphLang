```glyph
@ GET /status/:code {
  $ result = match code {
    200 => "OK"
    201 => "Created"
    400 => "Bad Request"
    404 => "Not Found"
    n when n >= 500 => "Server Error"
    _ => "Unknown"
  }
  > {status: code, message: result}
}
```