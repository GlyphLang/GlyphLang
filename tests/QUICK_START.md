# Glyph Tests - Quick Start Guide

Get up and running with Glyph tests in 5 minutes.

## Prerequisites

```bash
# Verify Go is installed
go version  # Should be 1.21+

# Navigate to project root
cd /path/to/glyph
```

## Run All Tests

```bash
# Run all tests (verbose output)
go test ./tests/... -v

# Run all tests (quiet - only failures)
go test ./tests/...

# Run tests with race detector
go test ./tests/... -race
```

## Expected Output (Current State)

```
=== RUN   TestHelloWorldExample
--- PASS: TestHelloWorldExample (0.00s)
=== RUN   TestRestAPIExample
--- PASS: TestRestAPIExample (0.00s)
=== RUN   TestSimpleRouteE2E
--- PASS: TestSimpleRouteE2E (0.00s)
=== RUN   TestCompilerBasicFlow
--- PASS: TestCompilerBasicFlow (0.00s)
=== RUN   TestVMBasicFlow
--- PASS: TestVMBasicFlow (0.00s)
=== RUN   TestStackOperations
--- PASS: TestStackOperations (0.00s)
... [many skipped tests] ...
PASS
ok      github.com/glyphlang/glyph/tests   0.123s
```

## Run Specific Tests

```bash
# Run only E2E tests
go test ./tests/... -v -run TestE2E

# Run only integration tests
go test ./tests/... -v -run TestIntegration

# Run specific test
go test ./tests/... -v -run TestHelloWorldExample

# Run tests matching pattern
go test ./tests/... -v -run "Test.*Compiler"
```

## Run Benchmarks

```bash
# Run all benchmarks
go test ./tests/... -bench=.

# Run with memory statistics
go test ./tests/... -bench=. -benchmem

# Run specific benchmark
go test ./tests/... -bench=BenchmarkCompilation

# Run benchmark multiple times for accuracy
go test ./tests/... -bench=BenchmarkCompilation -count=5
```

## Example Benchmark Output

```
BenchmarkCompilation-8                  100000      12345 ns/op      1024 B/op      10 allocs/op
BenchmarkCompilationSmallProgram-8       50000      23456 ns/op      2048 B/op      20 allocs/op
BenchmarkVMExecution-8                  500000       2345 ns/op       256 B/op       5 allocs/op
```

## Understanding Test Output

### Test Status Symbols
- `PASS` - Test passed
- `FAIL` - Test failed
- `SKIP` - Test skipped (awaiting implementation)

### Common Output
```
=== RUN   TestName
    test_file.go:123: Test message
--- PASS: TestName (0.01s)
```

## Testing Workflow for Developers

### 1. Before Making Changes
```bash
# Ensure tests pass before starting
go test ./tests/... -v
```

### 2. While Developing
```bash
# Run tests related to your component
go test ./tests/... -v -run TestParser  # If working on parser

# Watch mode (requires entr or similar)
find . -name "*.go" | entr -c go test ./tests/... -v
```

### 3. Before Committing
```bash
# Run full test suite
go test ./tests/... -v

# Run with race detector
go test ./tests/... -race

# Run benchmarks to check for performance regressions
go test ./tests/... -bench=. > bench-new.txt
# Compare with previous: benchstat bench-old.txt bench-new.txt
```

## Enabling Skipped Tests

When your component is ready, enable its tests:

### Step 1: Find Skipped Tests
```bash
# Find all skipped tests
grep -r "t.Skip" tests/

# Find skipped tests for specific component
grep -r "parser is implemented" tests/
```

### Step 2: Remove Skip Statement
```go
// Before
func TestParserIntegration(t *testing.T) {
    t.Skip("Skipping until parser is implemented")
    // ... test code ...
}

// After
func TestParserIntegration(t *testing.T) {
    // ... test code runs ...
}
```

### Step 3: Run the Test
```bash
go test ./tests/... -v -run TestParserIntegration
```

### Step 4: Fix Failures
```bash
# Run with more verbose output
go test ./tests/... -v -run TestParserIntegration

# Debug with print statements (they'll show in output)
# Add t.Logf("Debug: %v", value) in test code
```

## Test Coverage

```bash
# Generate coverage report
go test ./tests/... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out
```

## Debugging Tests

### Print Debug Information
```go
func TestExample(t *testing.T) {
    t.Logf("Debug info: %v", someValue)  // Always prints
    fmt.Printf("Debug: %v\n", someValue) // Only with -v
}
```

### Run with Verbose Output
```bash
go test ./tests/... -v -run TestExample
```

### Use Test Helpers
```go
helper := NewTestHelper(t)
helper.AssertEqual(got, want, "descriptive message")  # Clear error on failure
```

## Common Issues & Solutions

### Issue: Tests Not Found
```
no test files
```
**Solution**: Run from project root, not tests/ directory
```bash
cd /path/to/glyph  # Project root
go test ./tests/... -v  # Note the ./tests/...
```

### Issue: Import Errors
```
cannot find package "github.com/glyphlang/glyph/pkg/..."
```
**Solution**: Ensure go.mod is correct
```bash
go mod tidy
go test ./tests/... -v
```

### Issue: All Tests Skip
```
--- SKIP: TestName (0.00s)
```
**Solution**: This is expected! Tests skip until components are implemented.
Remove `t.Skip()` when component is ready.

### Issue: Test Timeout
```
panic: test timed out after 10m
```
**Solution**: Increase timeout or fix hanging test
```bash
go test ./tests/... -v -timeout 30m
```

## Test Files Overview

| File | Purpose | Test Count |
|------|---------|------------|
| `e2e_test.go` | Complete workflows | 17 tests |
| `integration_test.go` | Component integration | 15 tests |
| `benchmark_test.go` | Performance | 30+ benchmarks |
| `helpers.go` | Test utilities | N/A (utilities) |

## Test Fixtures

Test programs in `fixtures/`:

```bash
# List all fixtures
ls tests/fixtures/

# View a fixture
cat tests/fixtures/simple_route.glyph

# Use in tests
helper.LoadFixture("simple_route.glyph")
```

## Continuous Integration

### Local Pre-commit Check
```bash
#!/bin/bash
# Save as .git/hooks/pre-commit
go test ./tests/... || exit 1
echo "Tests passed!"
```

### CI Pipeline (GitHub Actions)
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test ./tests/... -v -race
```

## Performance Tracking

### Save Baseline
```bash
# Run benchmarks and save results
go test ./tests/... -bench=. -benchmem > bench-baseline.txt
```

### Compare Performance
```bash
# After changes, run again
go test ./tests/... -bench=. -benchmem > bench-new.txt

# Compare (requires benchstat)
go install golang.org/x/perf/cmd/benchstat@latest
benchstat bench-baseline.txt bench-new.txt
```

### Example Comparison Output
```
name                    old time/op  new time/op  delta
Compilation-8           12.3µs ± 2%  11.8µs ± 1%  -4.07%  (p=0.000 n=10+10)
VMExecution-8           2.34µs ± 1%  2.20µs ± 2%  -5.98%  (p=0.000 n=10+10)
```

## Next Steps

1. **Read Documentation**
   - `README.md` - Comprehensive test documentation
   - `TEST_SUMMARY.md` - Implementation summary
   - `TEST_ARCHITECTURE.md` - Visual test architecture

2. **Start Development**
   - Find tests for your component
   - Remove `t.Skip()` statements
   - Implement until tests pass

3. **Add New Tests**
   - Add fixtures in `fixtures/`
   - Add tests in appropriate file
   - Use test helpers for consistency

## Quick Command Reference

```bash
# All tests
go test ./tests/... -v

# Specific test
go test ./tests/... -v -run TestName

# Benchmarks
go test ./tests/... -bench=. -benchmem

# Coverage
go test ./tests/... -coverprofile=coverage.out

# Race detector
go test ./tests/... -race

# Skip cache
go test ./tests/... -count=1

# Parallel execution
go test ./tests/... -v -parallel 4

# Timeout
go test ./tests/... -v -timeout 30m
```

## Getting Help

- Check test output for error messages
- Look at similar passing tests for examples
- Read test documentation in `README.md`
- Review test helpers in `helpers.go`

## Success!

You're ready to start testing Glyph! As components are implemented, enable tests and watch the project come to life.

```bash
# Watch progress with:
go test ./tests/... -v | grep -E "(PASS|SKIP|FAIL)"
```

Good luck! 🚀
