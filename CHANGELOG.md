# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-12-16

### Changed
- **Project Rename**: Migrated from `ai_lang` to `GlyphLang/glyph`
- Updated all documentation references to use new project name
- Version bump to 1.0.0 - Production Ready release

### Documentation
- Fixed outdated `ai_lang` references in README.md, CLI_DEPLOYMENT.md, QUICK_START.md
- Updated project structure documentation
- Updated test output examples to reflect correct module path

## [0.2.0-beta] - 2025-12-13

### Fixed
- **WebSocket Tests**: Fixed timeout issues by adding proper goroutine tracking and shutdown sequence
  - Added WaitGroup tracking for connection goroutines (ReadPump/WritePump)
  - Fixed shutdown deadlock by closing connections before hub shutdown
  - Fixed mutex deadlock in Handler.Clear() method
  - Added nil check for mock connections in tests
  - WebSocket tests now complete in 3.3s (previously timed out after 600s)
- **Test Package Build**: Updated all tests to use new compiler API
  - Created parseSource() helper function for parsing source strings to AST Modules
  - Fixed 50+ occurrences across 8 test files (benchmark_test.go, bytecode_integration_test.go, e2e_test.go, etc.)
  - Tests package now builds successfully
- **Package Conflicts**: Removed conflicting debug files from root directory
  - Moved debug_*.go files to prevent main() function conflicts
  - All core packages now build and test successfully

### Changed
- Compiler API now accepts `*interpreter.Module` instead of raw source strings
- Test infrastructure updated to parse source before compilation

### Test Status
- ✅ All core package tests passing (compiler, vm, parser, websocket, etc.)
- ✅ 1000+ test assertions passing
- ⚠️ Some integration tests pending full server implementation (expected for beta)

## [0.1.0-alpha] - Previous

### Added
- Interactive showcase website with live code editor (`website/`)
- Dark futuristic UI with glassmorphism design
- Real-time Glyph syntax highlighting with Prism.js
- Animated performance metrics and scroll effects
- Initial project structure
- Go CLI with basic commands (compile, run, dev, init)
- Basic VM implementation in Go
- Example hello-world program
- Documentation (README, CONTRIBUTING, Quick Start)
- Build system (Makefile)
- Test infrastructure

### Changed
- **BREAKING**: Migrated from hybrid Rust+Go to 100% pure Go implementation
- Removed `glyph-core` Rust compiler in favor of native Go parser/compiler
- Updated build process - no longer requires Rust/Cargo toolchain
- Single binary deployment with zero external dependencies
- All 492+ tests passing with pure Go implementation

### Removed
- Rust compiler core (`glyph-core/` directory - 7,402 lines)
- FFI bridge between Go and Rust (`pkg/ffi/`)
- Cargo.toml and Rust build dependencies

## [0.1.0-alpha] - 2024-12-03

### Added
- Project initialization
- Hybrid Go + Rust architecture
- Basic project scaffolding
