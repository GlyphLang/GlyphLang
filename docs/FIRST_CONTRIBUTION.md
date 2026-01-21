# Your First Contribution to GlyphLang

Welcome to the GlyphLang project! This guide will help you make your first contribution. We are excited to have you join our community.

## Finding Your First Issue

Look for issues labeled `good-first-issue` in our GitHub repository:

1. Go to [github.com/GlyphLang/GlyphLang/issues](https://github.com/GlyphLang/GlyphLang/issues)
2. Filter by the `good-first-issue` label
3. Comment on the issue to let maintainers know you want to work on it
4. Wait for assignment before starting work

Good first issues typically involve documentation fixes, adding test cases, improving error messages, or small bug fixes.

---

## Understanding the Codebase

GlyphLang is written in Go and follows a modular architecture. Here is an overview of the main packages:

### pkg/parser - Lexer and Parser

Handles the first stage of compilation. The lexer tokenizes source code and the parser builds an Abstract Syntax Tree (AST). Good for learning about language design.

### pkg/compiler - Bytecode Compiler

Transforms the AST into bytecode with multiple optimization levels (0-3). Handles constant folding and dead code elimination.

### pkg/vm - Virtual Machine

A stack-based VM that executes compiled bytecode. Manages memory, implements all opcodes, and handles garbage collection.

### pkg/interpreter - AST Interpreter

An alternative execution mode that walks the AST directly. Easier to understand than bytecode execution and useful for debugging.

### pkg/server - HTTP Server

Handles HTTP routing, middleware chains, request/response handling, and WebSocket support.

### pkg/lsp - Language Server Protocol

Provides IDE features: code completion, go-to-definition, error diagnostics, and hover information.

### pkg/database - Database Integration

Handles PostgreSQL operations: connection pooling, query execution, transactions, and ORM-like operations.

---

## Beginner-Friendly Areas

Not sure where to start? These areas are welcoming to newcomers:

### Documentation Improvements
- Fix typos or unclear explanations in `docs/` or package READMEs
- Add examples to existing documentation
- Update outdated information

### Adding Test Cases
- Add tests for edge cases in any package
- Improve test coverage with table-driven tests
- Test error conditions and failure modes

Run tests with: `go test ./...`

### Parser Error Messages
- Improve error message clarity in `pkg/parser/` and `pkg/errors/`
- Add suggestions for common mistakes
- Include better context in syntax errors

### CLI Improvements
- Improve help text in `cmd/glyph/`
- Add better error messages for invalid arguments
- Enhance output formatting

---

## Step-by-Step: Your First PR

### 1. Fork and Clone

```bash
git clone https://github.com/YOUR_USERNAME/GlyphLang.git
cd GlyphLang
git remote add upstream https://github.com/GlyphLang/GlyphLang.git
```

### 2. Set Up Your Environment

```bash
go mod tidy
go build -o glyph ./cmd/glyph
./glyph version
```

### 3. Create a Branch

Always branch from main:

```bash
git fetch upstream
git checkout main
git merge upstream/main
git checkout -b feature/your-feature-name
```

### 4. Make Your Changes

- Follow existing code style
- Run `gofmt` to format your code
- Add tests for new functionality

### 5. Run Tests

```bash
go test ./...
go vet ./...
```

### 6. Commit and Push

```bash
git add .
git commit -m "fix: improve error message for missing semicolon"
git push origin feature/your-feature-name
```

Use prefixes: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`

### 7. Create the Pull Request

1. Go to GitHub and click "Compare & pull request"
2. Fill in the PR template with a clear description
3. Reference related issues (e.g., "Fixes #123")
4. Submit and respond to reviewer feedback

---

## Getting Help

### Ask in Your PR
If you have questions while working, ask in the PR comments. Maintainers are happy to help.

### Open a Discussion
For general questions, use the Discussions tab on GitHub.

### Review Existing Resources
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Detailed contribution guidelines
- [docs/ARCHITECTURE_DESIGN.md](ARCHITECTURE_DESIGN.md) - System architecture
- [docs/LANGUAGE_SPECIFICATION.md](LANGUAGE_SPECIFICATION.md) - Language syntax

### Tips for Getting Help
- Describe what you are trying to accomplish
- Share what you have already tried
- Include relevant error messages or code snippets

---

## What to Expect

After submitting your PR:

1. **Automated checks**: CI runs tests and linting
2. **CLA signature**: Sign the Contributor License Agreement
3. **Code review**: A maintainer reviews your changes
4. **Feedback**: You may receive requests for changes
5. **Merge**: Once approved, your PR is merged

Reviews typically take a few days. Feel free to ping the PR if you have not heard back in a week.

---

Thank you for contributing to GlyphLang! Every contribution helps make the project better. We look forward to your first PR.
