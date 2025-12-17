# Contributing to Glyph

Thank you for your interest in contributing to Glyph! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/glyph.git`
3. Create a branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.21 or later
- Rust 1.75 or later
- Make (optional, for convenience)

### Building

```bash
# Build everything
make build

# Or build individually
go build ./cmd/glyph          # Go CLI
cd glyph-core && cargo build  # Rust compiler
```

### Running Tests

```bash
# Run all tests
make test

# Or run individually
go test ./...                # Go tests
cd glyph-core && cargo test  # Rust tests
```

## Code Style

### Go Code
- Use `gofmt` to format code
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Run `golangci-lint` before submitting

### Rust Code
- Use `cargo fmt` to format code
- Follow [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- Run `cargo clippy` before submitting

## Pull Request Process

1. Update tests for any new functionality
2. Update documentation as needed
3. Ensure all tests pass
4. Update CHANGELOG.md with your changes
5. Submit PR with clear description of changes

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
feat: add new bytecode instruction
fix: resolve stack overflow in VM
docs: update README with examples
test: add compiler integration tests
```

## Areas for Contribution

- **Compiler**: Lexer, parser, type system implementation
- **VM**: Bytecode execution, optimization
- **Tooling**: CLI commands, IDE extensions
- **Documentation**: Tutorials, examples, guides
- **Testing**: Unit tests, integration tests, benchmarks

## Questions?

- Open an issue for discussion
- Join our Discord community
- Check existing documentation

Thank you for contributing!
