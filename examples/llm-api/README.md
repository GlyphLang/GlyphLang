# Document Q&A API

A document store with AI summarization and grounded question answering, showing the LLM provider used alongside the database provider in one file.

## Features

- LLM provider injection (`% ai: LLM`) next to database injection (`% db: Database`)
- Provider configuration from the environment, so the source carries no credentials
- API key authentication and rate limiting on the model-backed routes
- Guards for missing documents and empty model responses
- No client library, no async wiring, no glue code

## Configuration

The LLM provider is read from the environment at startup:

| Variable | Purpose |
|---|---|
| `GLYPH_LLM_PROVIDER` | `anthropic`, `openai`, or `ollama` |
| `GLYPH_LLM_API_KEY` | Provider credential. Not needed for `ollama` |
| `GLYPH_LLM_BASE_URL` | Optional. Self-hosted or proxied endpoints |

Starting without a provider configured prints a warning and the model-backed routes fail; the document routes still work.

```bash
export GLYPH_LLM_PROVIDER=anthropic
export GLYPH_LLM_API_KEY=sk-ant-...
glyph run examples/llm-api/main.glyph --port 8080
```

The database provider uses the built-in in-memory mock, so documents live only as long as the process.

## API Endpoints

### Create a Document
```
POST /documents
```

**Request:**
```json
{
  "title": "Ponytail",
  "body": "The best code is the code never written."
}
```

**Response:** `201`
```json
{
  "id": 1,
  "title": "Ponytail",
  "body": "The best code is the code never written."
}
```

### Get a Document
```
GET /documents/:id
```

Returns the stored document, or `404` with `{"error": "document not found"}`.

### Summarize a Document
```
POST /documents/:id/summarize
```

Sends the document body to the model, stores the summary, and returns it. Requires an API key and is limited to 20 requests per minute.

**Response:** `200`
```json
{
  "id": 1,
  "title": "Ponytail",
  "summary": "A two sentence summary produced by the model.",
  "model": "claude-opus-5"
}
```

Responds `404` if the document is missing and `502` if the model returns no content.

### Ask a Question
```
POST /documents/:id/ask
```

Answers using only the stored document, and says so when the answer is not in it.

**Request:**
```json
{
  "question": "What is the best code?"
}
```

**Response:** `200`
```json
{
  "question": "What is the best code?",
  "answer": "The code never written.",
  "document_id": 1
}
```

Responds `400` if `question` is missing, `404` if the document is missing, and `502` if the model returns no content.

## Running Without a Provider Account

Any endpoint that speaks the Ollama protocol works, including a local Ollama install or a stub:

```bash
export GLYPH_LLM_PROVIDER=ollama
export GLYPH_LLM_BASE_URL=http://127.0.0.1:11434
glyph run examples/llm-api/main.glyph --port 8080
```

## Authentication

The two model-backed routes declare `+ auth(apikey)`, so they need a key list before they will serve anyone:

```bash
export GLYPH_API_KEYS="key-one,key-two"
```

Callers send it as `X-API-Key: key-one` or `Authorization: Bearer key-one`. Without the variable set, those routes reject every request with `401` and a warning is printed at startup - auth fails closed, so an unconfigured deployment denies rather than serves. That matters here because an open model-backed route spends tokens for anonymous callers.

`+ auth(jwt)` elsewhere uses `GLYPH_JWT_SECRET` the same way.

## Rate Limiting

`+ ratelimit(20/min)` on the model-backed routes and `60/min` on the document routes are enforced per client IP; requests past the budget get `429`. No configuration needed.

## Notes

- The `Document` type declares `summary` as optional, so a document reads back fine before it has been summarized.
- The database provider is the in-memory mock, so documents live only as long as the process.
