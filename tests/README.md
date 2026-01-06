# Glyph Test Suite

Comprehensive test infrastructure for the Glyph project.

## Test Structure

```
tests/
├── README.md                    # This file
├── helpers.go                   # Test utilities and helper functions
├── e2e_test.go                 # End-to-end integration tests
├── integration_test.go         # Component integration tests
├── benchmark_test.go           # Performance benchmarks
└── fixtures/                   # Test Glyph programs
    ├── simple_route.glyph
    ├── path_param.glyph
    ├── json_response.glyph
    ├── multiple_routes.glyph
    ├── with_auth.glyph
    ├── post_route.glyph
    ├── invalid_syntax.glyph
    └── error_handling.glyph
```

## Running Tests

### Run All Tests
```bash
go test ./tests/... -v
```

### Run Specific Test Files
```bash
# End-to-end tests only
go test ./tests/... -v -run TestE2E

# Integration tests only
go test ./tests/... -v -run TestIntegration

# Benchmarks only
go test ./tests/... -bench=. -benchmem
```

### Run Specific Tests
```bash
# Test hello-world example
go test ./tests/... -v -run TestHelloWorldExample

# Test compilation
go test ./tests/... -v -run TestCompiler

# Test VM execution
go test ./tests/... -v -run TestVM
```

### Run Benchmarks
```bash
# All benchmarks
go test ./tests/... -bench=. -benchmem

# Compilation benchmarks
go test ./tests/... -bench=BenchmarkCompilation -benchmem

# VM benchmarks
go test ./tests/... -bench=BenchmarkVM -benchmem

# Scaling benchmarks
go test ./tests/... -bench=BenchmarkScaling -benchmem
```

## Test Coverage

### End-to-End Tests (`e2e_test.go`)

Tests complete workflows from source code to execution:

- **TestHelloWorldExample**: Tests the hello-world example program
- **TestRestAPIExample**: Tests the REST API example program
- **TestSimpleRouteE2E**: Tests basic route compilation and execution
- **TestPathParametersE2E**: Tests routes with path parameters
- **TestJSONSerializationE2E**: Tests JSON response handling
- **TestMultipleRoutesE2E**: Tests programs with multiple routes
- **TestHTTPMethodsE2E**: Tests different HTTP methods (GET, POST, etc.)
- **TestAuthenticationE2E**: Tests authentication middleware
- **TestErrorHandlingE2E**: Tests error handling and error types
- **TestInvalidSyntaxE2E**: Tests that invalid syntax is rejected
- **TestServerStartupShutdownE2E**: Tests server lifecycle (skipped until implemented)
- **TestHotReloadE2E**: Tests development server hot reload (skipped)
- **TestConcurrentRequestsE2E**: Tests concurrent request handling (skipped)
- **TestDatabaseIntegrationE2E**: Tests database operations (skipped)
- **TestSecurityFeaturesE2E**: Tests security features (skipped)

### Integration Tests (`integration_test.go`)

Tests individual components and their integration:

- **TestParserIntegration**: Tests parser with various Glyph syntax
- **TestLexerIntegration**: Tests lexer token generation
- **TestTypeCheckerIntegration**: Tests type checking and validation
- **TestInterpreterIntegration**: Tests AST interpretation (skipped)
- **TestServerIntegration**: Tests HTTP server (skipped)
- **TestCompilerVMIntegration**: Tests compiler → VM pipeline
- **TestStackOperations**: Tests VM stack operations
- **TestValueTypes**: Tests VM value types
- **TestRouteMatching**: Tests route pattern matching (skipped)
- **TestMiddlewareChain**: Tests middleware execution (skipped)
- **TestDatabaseQuerySafety**: Tests SQL injection prevention (skipped)
- **TestInputValidation**: Tests input validation rules (skipped)
- **TestErrorPropagation**: Tests error handling (skipped)
- **TestBytecodeFormat**: Tests bytecode structure
- **TestCompilerErrorMessages**: Tests error reporting (skipped)

### Performance Benchmarks (`benchmark_test.go`)

Measures performance of various operations:

- **BenchmarkCompilation**: Basic compilation performance
- **BenchmarkCompilationSmallProgram**: Small program compilation
- **BenchmarkCompilationMediumProgram**: Medium program compilation
- **BenchmarkCompilationLargeProgram**: Large program (50 routes)
- **BenchmarkVMExecution**: VM bytecode execution
- **BenchmarkStackOperations**: VM stack push/pop performance
- **BenchmarkCompileAndExecute**: Full compile + execute cycle
- **BenchmarkParallelCompilation**: Concurrent compilation
- **BenchmarkParallelExecution**: Concurrent execution
- **BenchmarkMemoryAllocation**: Memory allocation patterns
- **BenchmarkComparisonWithExamples**: Real example programs
- **BenchmarkCompilationScaling**: Scaling behavior (1-100 routes)

## Test Fixtures

Test fixtures are small Glyph programs in `fixtures/` directory:

1. **simple_route.glyph**: Basic route returning JSON
2. **path_param.glyph**: Route with path parameter (`:name`)
3. **json_response.glyph**: Route with type definition and JSON response
4. **multiple_routes.glyph**: Multiple routes in one program
5. **with_auth.glyph**: Route with authentication and rate limiting
6. **post_route.glyph**: POST route with input validation
7. **invalid_syntax.glyph**: Invalid syntax (for error testing)
8. **error_handling.glyph**: Route with Result type (success | error)

## Test Helpers

The `helpers.go` file provides utilities:

### TestHelper
- `LoadFixture(name)`: Load test fixture file
- `AssertEqual()`: Compare values
- `AssertNoError()`: Check no error occurred
- `AssertError()`: Check error occurred
- `AssertContains()`: Check string contains substring

### HTTP Testing
- `MockServer`: Test HTTP server
- `MakeRequest()`: Make HTTP request to test server
- `HTTPRequest`: Request structure
- `HTTPResponse`: Response structure with JSON parsing

### Mocks
- `CompilerMock`: Mock compiler for testing
- `InterpreterMock`: Mock interpreter for testing

### Utilities
- `TempFile()`: Create temporary test file
- `RetryWithTimeout()`: Retry operation with timeout
- `AssertJSONField()`: Check JSON field value

## Test-Driven Development (TDD) Approach

Many tests are written before implementation, following TDD principles:

1. **Write Tests First**: Tests define expected behavior
2. **Tests Currently Skip**: Many tests skip until components are ready
3. **Remove Skips**: As components are implemented, remove `t.Skip()` calls
4. **Tests Guide Implementation**: Use failing tests to guide development

### Example: Enabling a Test

When a component is ready, remove the skip:

```go
// Before (skipped)
func TestServerIntegration(t *testing.T) {
    t.Skip("Skipping until server is implemented")
    // ... test code ...
}

// After (enabled)
func TestServerIntegration(t *testing.T) {
    // ... test code runs ...
}
```

## Current Status

### Passing Tests (Placeholder Implementation)
- ✅ TestCompilerBasicFlow: Compiler creates valid bytecode
- ✅ TestVMBasicFlow: VM executes valid bytecode
- ✅ TestStackOperations: VM stack push/pop works
- ✅ TestValueTypes: VM value types are correct
- ✅ TestBytecodeFormat: Bytecode has correct magic bytes
- ✅ TestMockServerBasic: Test infrastructure verified

### Pending Tests (Awaiting Implementation)
- ⏸️ Most E2E tests: Need lexer, parser, interpreter, server
- ⏸️ Parser tests: Need parser implementation
- ⏸️ Type checker tests: Need type checker
- ⏸️ Server tests: Need HTTP server
- ⏸️ Database tests: Need database integration
- ⏸️ Security tests: Need security features

### Benchmarks Status
- ✅ Basic benchmarks run with placeholder implementation
- ⏸️ Advanced benchmarks skipped until components ready

## Integration with CI/CD

### GitHub Actions (Future)
```yaml
- name: Run tests
  run: go test ./tests/... -v -race -coverprofile=coverage.txt

- name: Run benchmarks
  run: go test ./tests/... -bench=. -benchmem

- name: Upload coverage
  uses: codecov/codecov-action@v3
```

### Local Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit
go test ./tests/... -v
```

## Adding New Tests

### 1. Add to Appropriate File

**End-to-end test** (tests complete workflow):
```go
// In e2e_test.go
func TestNewFeatureE2E(t *testing.T) {
    helper := NewTestHelper(t)
    // ... test implementation ...
}
```

**Integration test** (tests component):
```go
// In integration_test.go
func TestNewComponentIntegration(t *testing.T) {
    helper := NewTestHelper(t)
    // ... test implementation ...
}
```

**Benchmark** (measures performance):
```go
// In benchmark_test.go
func BenchmarkNewOperation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // ... operation to benchmark ...
    }
}
```

### 2. Add Test Fixture (if needed)

```bash
# Create new fixture
echo '@ route /new-feature
  > {status: "ok"}' > tests/fixtures/new_feature.glyph
```

### 3. Use Test Helpers

```go
helper := NewTestHelper(t)

// Load fixture
source := helper.LoadFixture("new_feature.glyph")

// Make assertions
helper.AssertEqual(got, want, "value comparison")
helper.AssertNoError(err, "operation should succeed")
helper.AssertContains(text, "expected", "text should contain")
```

## Test Conventions

1. **Test Names**: Use descriptive names with context
   - Good: `TestHelloWorldExample`
   - Bad: `TestExample1`

2. **Skip vs Comment**: Use `t.Skip()` for tests awaiting implementation
   - Provides clear feedback when running tests
   - Easy to find and enable later

3. **Assertions**: Use helper methods for consistent error messages
   - `helper.AssertEqual()` instead of manual `if` checks

4. **Fixtures**: Create reusable test programs in `fixtures/`
   - One file per test scenario
   - Clear, minimal examples

5. **Documentation**: Add comments explaining what tests validate
   - Especially important for TDD tests written before implementation

## Performance Goals

Based on CONTEXT.md goals:

- **Compilation Speed**: Target 3 seconds for typical program
- **HTTP Response Time**: Target < 10ms for simple routes
- **Throughput**: Target 10,000 requests/second
- **Memory**: Efficient stack usage, no leaks

Benchmarks track progress toward these goals.

## Troubleshooting

### Tests Won't Run
```bash
# Check Go installation
go version

# Ensure dependencies are installed
go mod tidy

# Run from project root
cd /path/to/glyph
go test ./tests/... -v
```

### Skipped Tests
- Normal during development
- Tests skip until components are implemented
- Check test output for skip messages

### Benchmark Variation
```bash
# Run multiple times for stable results
go test ./tests/... -bench=BenchmarkName -count=5

# Use benchstat for comparison
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

## Contributing

When adding a new component:

1. Check if tests exist (they might be skipped)
2. Remove `t.Skip()` from relevant tests
3. Run tests to verify implementation
4. Add new tests for edge cases
5. Update this README if adding new test categories

## Future Test Additions

Planned test areas:

- [ ] Language Server Protocol (LSP) tests
- [ ] VS Code extension tests
- [ ] Context cache tests
- [ ] Pattern marketplace tests
- [ ] Cloud deployment tests
- [ ] Load testing with realistic workloads
- [ ] Fuzzing tests for parser robustness
- [ ] Property-based tests with rapid
- [ ] Mutation testing for test quality

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Benchmarking](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Glyph Project Context](../CONTEXT.md)

---

**Last Updated**: 2024-12-04
**Test Infrastructure Version**: 1.0.0
**Status**: Comprehensive test suite ready for component implementation
