package interpreter

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestFutureBasicOperations tests basic Future functionality
func TestFutureBasicOperations(t *testing.T) {
	t.Run("NewFuture creates pending future", func(t *testing.T) {
		f := NewFuture()
		if !f.IsPending() {
			t.Error("Expected new future to be pending")
		}
		if f.IsResolved() {
			t.Error("Expected new future to not be resolved")
		}
		if f.IsRejected() {
			t.Error("Expected new future to not be rejected")
		}
	})

	t.Run("Resolve sets value", func(t *testing.T) {
		f := NewFuture()
		f.Resolve(42)

		if !f.IsResolved() {
			t.Error("Expected future to be resolved")
		}
		if f.IsPending() {
			t.Error("Expected future to not be pending")
		}
		if f.Value() != 42 {
			t.Errorf("Expected value 42, got %v", f.Value())
		}
	})

	t.Run("Reject sets error", func(t *testing.T) {
		f := NewFuture()
		expectedErr := errors.New("test error")
		f.Reject(expectedErr)

		if !f.IsRejected() {
			t.Error("Expected future to be rejected")
		}
		if f.IsPending() {
			t.Error("Expected future to not be pending")
		}
		if f.Error() != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, f.Error())
		}
	})

	t.Run("Double resolve is ignored", func(t *testing.T) {
		f := NewFuture()
		f.Resolve(1)
		f.Resolve(2) // Should be ignored

		if f.Value() != 1 {
			t.Errorf("Expected value 1 (first resolve), got %v", f.Value())
		}
	})

	t.Run("Reject after resolve is ignored", func(t *testing.T) {
		f := NewFuture()
		f.Resolve(1)
		f.Reject(errors.New("error")) // Should be ignored

		if !f.IsResolved() {
			t.Error("Expected future to remain resolved")
		}
		if f.IsRejected() {
			t.Error("Expected future to not be rejected")
		}
	})
}

// TestFutureAwait tests the Await method
func TestFutureAwait(t *testing.T) {
	t.Run("Await returns resolved value", func(t *testing.T) {
		f := NewFuture()

		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Resolve("hello")
		}()

		val, err := f.Await()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != "hello" {
			t.Errorf("Expected 'hello', got %v", val)
		}
	})

	t.Run("Await returns rejected error", func(t *testing.T) {
		f := NewFuture()
		expectedErr := errors.New("async error")

		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Reject(expectedErr)
		}()

		val, err := f.Await()
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
		if val != nil {
			t.Errorf("Expected nil value, got %v", val)
		}
	})

	t.Run("AwaitWithTimeout succeeds before timeout", func(t *testing.T) {
		f := NewFuture()

		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Resolve(100)
		}()

		val, err := f.AwaitWithTimeout(1 * time.Second)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != 100 {
			t.Errorf("Expected 100, got %v", val)
		}
	})

	t.Run("AwaitWithTimeout times out", func(t *testing.T) {
		f := NewFuture()
		// Don't resolve - let it timeout

		_, err := f.AwaitWithTimeout(50 * time.Millisecond)
		if err == nil {
			t.Error("Expected timeout error")
		}
	})
}

// TestRunAsync tests the RunAsync helper
func TestRunAsync(t *testing.T) {
	t.Run("RunAsync resolves with return value", func(t *testing.T) {
		f := RunAsync(func() (interface{}, error) {
			return 42, nil
		})

		val, err := f.Await()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != 42 {
			t.Errorf("Expected 42, got %v", val)
		}
	})

	t.Run("RunAsync rejects with error", func(t *testing.T) {
		expectedErr := errors.New("async failure")
		f := RunAsync(func() (interface{}, error) {
			return nil, expectedErr
		})

		_, err := f.Await()
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestFutureAll tests the All combinator
func TestFutureAll(t *testing.T) {
	t.Run("All resolves with all values", func(t *testing.T) {
		f1 := NewFuture()
		f2 := NewFuture()
		f3 := NewFuture()

		go func() {
			f1.Resolve(1)
			f2.Resolve(2)
			f3.Resolve(3)
		}()

		result := All(f1, f2, f3)
		val, err := result.Await()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		values, ok := val.([]interface{})
		if !ok {
			t.Fatalf("Expected []interface{}, got %T", val)
		}
		if len(values) != 3 {
			t.Errorf("Expected 3 values, got %d", len(values))
		}
		if values[0] != 1 || values[1] != 2 || values[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", values)
		}
	})

	t.Run("All rejects on first error", func(t *testing.T) {
		f1 := NewFuture()
		f2 := NewFuture()
		expectedErr := errors.New("failure")

		go func() {
			f1.Resolve(1)
			f2.Reject(expectedErr)
		}()

		result := All(f1, f2)
		_, err := result.Await()
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestFutureRace tests the Race combinator
func TestFutureRace(t *testing.T) {
	t.Run("Race resolves with first completion", func(t *testing.T) {
		f1 := NewFuture()
		f2 := NewFuture()

		go func() {
			time.Sleep(10 * time.Millisecond)
			f1.Resolve("slow")
		}()

		go func() {
			f2.Resolve("fast")
		}()

		result := Race(f1, f2)
		val, err := result.Await()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != "fast" {
			t.Errorf("Expected 'fast', got %v", val)
		}
	})

	t.Run("Race with no futures rejects", func(t *testing.T) {
		result := Race()
		_, err := result.Await()
		if err == nil {
			t.Error("Expected error for empty Race")
		}
	})
}

// TestFutureAny tests the Any combinator
func TestFutureAny(t *testing.T) {
	t.Run("Any resolves with first success", func(t *testing.T) {
		f1 := NewFuture()
		f2 := NewFuture()

		go func() {
			f1.Reject(errors.New("error 1"))
			f2.Resolve("success")
		}()

		result := Any(f1, f2)
		val, err := result.Await()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if val != "success" {
			t.Errorf("Expected 'success', got %v", val)
		}
	})

	t.Run("Any rejects when all fail", func(t *testing.T) {
		f1 := NewFuture()
		f2 := NewFuture()

		go func() {
			f1.Reject(errors.New("error 1"))
			f2.Reject(errors.New("error 2"))
		}()

		result := Any(f1, f2)
		_, err := result.Await()
		if err == nil {
			t.Error("Expected error when all futures reject")
		}
	})
}

// TestIsFuture tests the type check utility
func TestIsFuture(t *testing.T) {
	f := NewFuture()
	if !IsFuture(f) {
		t.Error("Expected IsFuture to return true for *Future")
	}
	if IsFuture(42) {
		t.Error("Expected IsFuture to return false for non-Future")
	}
	if IsFuture("string") {
		t.Error("Expected IsFuture to return false for string")
	}
}

// TestFutureStateString tests the String method
func TestFutureStateString(t *testing.T) {
	if FuturePending.String() != "pending" {
		t.Errorf("Expected 'pending', got %s", FuturePending.String())
	}
	if FutureResolved.String() != "resolved" {
		t.Errorf("Expected 'resolved', got %s", FutureResolved.String())
	}
	if FutureRejected.String() != "rejected" {
		t.Errorf("Expected 'rejected', got %s", FutureRejected.String())
	}
}

// TestConcurrentFutures tests that futures actually run concurrently
func TestConcurrentFutures(t *testing.T) {
	var counter int64

	f1 := RunAsync(func() (interface{}, error) {
		atomic.AddInt64(&counter, 1)
		time.Sleep(50 * time.Millisecond)
		return "f1", nil
	})

	f2 := RunAsync(func() (interface{}, error) {
		atomic.AddInt64(&counter, 1)
		time.Sleep(50 * time.Millisecond)
		return "f2", nil
	})

	f3 := RunAsync(func() (interface{}, error) {
		atomic.AddInt64(&counter, 1)
		time.Sleep(50 * time.Millisecond)
		return "f3", nil
	})

	// Small delay to let all goroutines start
	time.Sleep(10 * time.Millisecond)

	// All three should have started by now (counter == 3)
	if atomic.LoadInt64(&counter) != 3 {
		t.Errorf("Expected all futures to start concurrently, only %d started", counter)
	}

	// Wait for all
	result := All(f1, f2, f3)
	_, err := result.Await()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestAsyncExprEvaluation tests evaluation of async/await using direct AST construction
func TestAsyncExprEvaluation(t *testing.T) {
	t.Run("Direct async block evaluation", func(t *testing.T) {
		// Test the evaluator directly with a constructed async expression
		interp := NewInterpreter()
		env := NewEnvironment()

		// Create an async expression that returns 42
		asyncExpr := AsyncExpr{
			Body: []Statement{
				ReturnStatement{
					Value: LiteralExpr{Value: IntLiteral{Value: 42}},
				},
			},
		}

		// Evaluate the async expression
		result, err := interp.evaluateAsyncExpr(asyncExpr, env)
		if err != nil {
			t.Fatalf("Async evaluation error: %v", err)
		}

		// Result should be a Future
		future, ok := result.(*Future)
		if !ok {
			t.Fatalf("Expected *Future, got %T", result)
		}

		// Await the future
		value, err := future.Await()
		if err != nil {
			t.Fatalf("Await error: %v", err)
		}

		if value != int64(42) {
			t.Errorf("Expected 42, got %v", value)
		}
	})

	t.Run("Direct await expression evaluation", func(t *testing.T) {
		interp := NewInterpreter()
		env := NewEnvironment()

		// Create a future and store it in the environment
		future := NewFuture()
		go func() {
			time.Sleep(10 * time.Millisecond)
			future.Resolve("async result")
		}()
		env.Define("myFuture", future)

		// Create an await expression
		awaitExpr := AwaitExpr{
			Expr: VariableExpr{Name: "myFuture"},
		}

		// Evaluate the await expression
		result, err := interp.evaluateAwaitExpr(awaitExpr, env)
		if err != nil {
			t.Fatalf("Await evaluation error: %v", err)
		}

		if result != "async result" {
			t.Errorf("Expected 'async result', got %v", result)
		}
	})

	t.Run("Async with computation", func(t *testing.T) {
		interp := NewInterpreter()
		env := NewEnvironment()

		// Create an async block that does computation
		asyncExpr := AsyncExpr{
			Body: []Statement{
				AssignStatement{
					Target: "x",
					Value:  LiteralExpr{Value: IntLiteral{Value: 10}},
				},
				AssignStatement{
					Target: "y",
					Value:  LiteralExpr{Value: IntLiteral{Value: 20}},
				},
				ReturnStatement{
					Value: BinaryOpExpr{
						Op:    Add,
						Left:  VariableExpr{Name: "x"},
						Right: VariableExpr{Name: "y"},
					},
				},
			},
		}

		result, err := interp.evaluateAsyncExpr(asyncExpr, env)
		if err != nil {
			t.Fatalf("Async evaluation error: %v", err)
		}

		future, ok := result.(*Future)
		if !ok {
			t.Fatalf("Expected *Future, got %T", result)
		}

		value, err := future.Await()
		if err != nil {
			t.Fatalf("Await error: %v", err)
		}

		// The result should be 30 (10 + 20)
		if value != int64(30) {
			t.Errorf("Expected 30, got %v (type: %T)", value, value)
		}
	})
}
