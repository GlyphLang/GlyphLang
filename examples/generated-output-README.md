# Generated Output Examples

The `generated/` directories in each example contain pre-generated server code produced by `glyph codegen`. These files are checked in so you can see what the codegen produces without installing the tool.

## Directory Structure

Each example with generated output follows this layout:

```
examples/<name>/
  main.glyph              # Source .glyph file
  generated/
    python/
      main.py             # FastAPI application
      requirements.txt    # Python dependencies
    typescript/
      src/app.ts          # Express application
      package.json        # Node.js dependencies
      tsconfig.json       # TypeScript configuration
```

## Available Examples

| Example | Description | Source |
|---------|-------------|--------|
| `hello-world/` | Minimal API with 3 routes, no types or providers | `main.glyph` |
| `intent-tests/` | CRUD Todo API with types, database provider, error handling | `01-crud-api.glyph` |
| `custom-provider/` | Custom provider contracts (EmailService, PaymentGateway, Notifier) | `main.glyph` |

## Regenerating

To regenerate these files, build the Glyph compiler and run:

```bash
go build -o glyph ./cmd/glyph

# Hello World
./glyph codegen examples/hello-world/main.glyph --lang python -o examples/hello-world/generated/python
./glyph codegen examples/hello-world/main.glyph --lang typescript -o examples/hello-world/generated/typescript

# CRUD API
./glyph codegen examples/intent-tests/01-crud-api.glyph --lang python -o examples/intent-tests/generated/python
./glyph codegen examples/intent-tests/01-crud-api.glyph --lang typescript -o examples/intent-tests/generated/typescript

# Custom Provider
./glyph codegen examples/custom-provider/main.glyph --lang python -o examples/custom-provider/generated/python
./glyph codegen examples/custom-provider/main.glyph --lang typescript -o examples/custom-provider/generated/typescript
```

## Running the Generated Code

### Python/FastAPI

```bash
cd examples/hello-world/generated/python
pip install -r requirements.txt
uvicorn main:app --reload
```

### TypeScript/Express

```bash
cd examples/hello-world/generated/typescript
npm install
npm run dev
```
