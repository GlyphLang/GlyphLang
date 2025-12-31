# Contributing to GlyphLang

Thank you for your interest in contributing to GlyphLang! This document outlines the process for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Commit Messages](#commit-messages)
- [Pull Requests](#pull-requests)
- [Reporting Issues](#reporting-issues)
- [Feature Requests](#feature-requests)
- [Contributor License Agreement](#contributor-license-agreement)
- [Code of Conduct](#code-of-conduct)

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up the development environment
4. Create a branch for your changes
5. Make your changes and test them
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git

### Building from Source

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/GlyphLang.git
cd GlyphLang

# Add upstream remote
git remote add upstream https://github.com/GlyphLang/GlyphLang.git

# Install dependencies
go mod tidy

# Build the CLI
go build -o glyph ./cmd/glyph

# Verify the build
./glyph version
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./pkg/parser/...

# Run tests with verbose output
go test -v ./...
```

### Running Linters

```bash
# Run go vet
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

## Making Changes

1. **Sync with upstream** before starting work:
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes** in small, logical commits

4. **Test your changes** thoroughly

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

## Code Style

### Go Code

- Follow standard Go conventions and idioms
- Use `gofmt` to format code (run automatically on save in most editors)
- Run `go vet` before committing
- Keep functions focused and reasonably sized
- Add comments for exported functions, types, and non-obvious logic
- Handle errors explicitly; avoid ignoring error returns

### Naming Conventions

- Use camelCase for unexported identifiers
- Use PascalCase for exported identifiers
- Use descriptive names that convey purpose
- Avoid single-letter names except for loop indices or very short scopes

### File Organization

- One package per directory
- Test files should be named `*_test.go`
- Keep related functionality together

## Testing

- Write tests for new functionality
- Ensure existing tests pass before submitting
- Aim for meaningful test coverage, not just high percentages
- Use table-driven tests where appropriate
- Test edge cases and error conditions

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {"valid input", validInput, expectedOutput, false},
        {"invalid input", invalidInput, OutputType{}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if !reflect.DeepEqual(got, tt.expected) {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

## Commit Messages

Write clear, concise commit messages that explain what and why:

### Format

```
<type>: <short summary>

<optional body explaining the change in more detail>

<optional footer with references>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring without behavior change
- `perf`: Performance improvement
- `chore`: Maintenance tasks

### Examples

```
feat: Add support for array destructuring in match expressions

fix: Resolve panic when parsing empty object literals

docs: Update quickstart guide with database examples

test: Add coverage for edge cases in lexer
```

## Pull Requests

### Before Submitting

- [ ] Tests pass locally (`go test ./...`)
- [ ] Code is formatted (`gofmt`)
- [ ] No linter warnings (`go vet ./...`)
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventions

### PR Description

Use the pull request template and include:

- Clear description of the change
- Motivation and context
- How you tested the changes
- Any breaking changes or migration notes

### Review Process

1. A maintainer will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge the PR

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- GlyphLang version (`glyph version`)
- Operating system and version
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Minimal code example if applicable
- Error messages or stack traces

### Security Issues

For security vulnerabilities, please do NOT open a public issue. Instead, email the maintainers directly or use GitHub's private security reporting feature.

## Feature Requests

We welcome feature requests! When proposing a feature:

- Check existing issues to avoid duplicates
- Describe the problem you're trying to solve
- Explain your proposed solution
- Consider alternatives you've thought about
- Discuss any potential drawbacks

## Contributor License Agreement

All contributors must sign our CLA before their pull request can be merged. When you open a pull request, the CLA Assistant bot will automatically check if you have signed. If not, it will provide a link to sign electronically.

**[View the full CLA](https://gist.github.com/GlyphLang/daf9e93df782faadd67c22afc3661934)**

By signing the CLA, you agree to the following terms:

### Grant of Rights

1. **Copyright License**: You grant the GlyphLang maintainers a perpetual, worldwide, non-exclusive, royalty-free, irrevocable license to use, reproduce, modify, display, perform, sublicense, and distribute your contributions as part of the project.

2. **Patent License**: You grant the GlyphLang maintainers a perpetual, worldwide, non-exclusive, royalty-free, irrevocable patent license to make, have made, use, sell, offer for sale, import, and otherwise transfer your contributions.

3. **Relicensing Rights**: You grant the GlyphLang maintainers the right to relicense your contributions under any license, including proprietary licenses, without requiring additional permission from you.

### Representations

You represent that:

1. You are the original author of your contributions, or you have the right to submit them under this agreement.
2. Your contributions do not violate any third-party rights, including intellectual property rights.
3. Your contributions are provided "as is" without any warranty of any kind.

### Retention of Rights

You retain all rights to your contributions. This agreement does not transfer ownership of your contributions to the project maintainers. You are free to use your contributions for any other purpose.

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the technical merits of contributions
- Help maintain a welcoming environment for all contributors
- Assume good intentions
- Be patient with newcomers

## Questions

If you have questions about contributing:

- Open a discussion on GitHub
- Check existing issues and documentation
- Ask in your pull request

Thank you for contributing to GlyphLang!
