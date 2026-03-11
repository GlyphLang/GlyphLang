package interpreter

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewFuture(t *testing.T) {
	f := NewFuture()
	if f == nil {
		t.Fatal("NewFuture() returned nil")
	}
	if !f.IsPending() {
		t.Error("New future should be pending")
	}
	if f.IsResolved() {
		t.Error("New future should not be resolved")
	}
	if f.IsRejected() {
		t.Error("New future should not be rejected")
	}
	if f.State() != FuturePending {
		t.Errorf("State() = %v, want FuturePending", f.State())
	}
}

func TestFuture_Resolve(t *testing.T) {
	f := NewFuture()
	f.Resolve("hello")

	if !f.IsResolved() {
		t.Error("Future should be resolved")
	}
	if f.IsPending() {
		t.Error("Resolved future should not be pending")
	}
	if f.IsRejected() {
		t.Error("Resolved future should not be rejected")
	}
	if f.State() != FutureResolved {
		t.Errorf("State() = %v, want FutureResolved", f.State())
	}
	if f.Value() != "hello" {
		t.Errorf("Value() = %v, want %q", f.Value(), "hello")
	}
	if f.Error() != nil {
		t.Errorf("Error() should be nil, got %v", f.Error())
	}
}

func TestFuture_Reject(t *testing.T) {
	f := NewFuture()
	expectedErr := errors.New("something failed")
	f.Reject(expectedErr)

	if !f.IsRejected() {
		t.Error("Future should be rejected")
	}
	if f.IsPending() {
		t.Error("Rejected future should not be pending")
	}
	if f.IsResolved() {
		t.Error("Rejected future should not be resolved")
	}
	if f.State() != FutureRejected {
		t.Errorf("State() = %v, want FutureRejected", f.State())
	}
	if f.Error() != expectedErr {
		t.Errorf("Error() = %v, want %v", f.Error(), expectedErr)
	}
	if f.Value() != nil {
		t.Errorf("Value() should be nil for rejected future, got %v", f.Value())
	}
}

func TestFuture_DoubleResolve(t *testing.T) {
	f := NewFuture()
	f.Resolve("first")
	f.Resolve("second") // Should be ignored

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}
	if val != "first" {
		t.Errorf("Value should be 'first' (first resolution wins), got %v", val)
	}
}

func TestFuture_DoubleReject(t *testing.T) {
	f := NewFuture()
	f.Reject(errors.New("first error"))
	f.Reject(errors.New("second error")) // Should be ignored

	_, err := f.Await()
	if err == nil {
		t.Fatal("Await() should return error")
	}
	if err.Error() != "first error" {
		t.Errorf("Error should be 'first error' (first rejection wins), got %q", err.Error())
	}
}

func TestFuture_ResolveAfterReject(t *testing.T) {
	f := NewFuture()
	f.Reject(errors.New("rejected"))
	f.Resolve("too late") // Should be ignored

	_, err := f.Await()
	if err == nil {
		t.Fatal("Await() should return error from rejection")
	}
	if err.Error() != "rejected" {
		t.Errorf("Error should be 'rejected', got %q", err.Error())
	}
}

func TestFuture_RejectAfterResolve(t *testing.T) {
	f := NewFuture()
	f.Resolve("resolved")
	f.Reject(errors.New("too late")) // Should be ignored

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}
	if val != "resolved" {
		t.Errorf("Value should be 'resolved', got %v", val)
	}
}

func TestFuture_Await_Resolved(t *testing.T) {
	f := NewFuture()
	f.Resolve(42)

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}
}

func TestFuture_Await_Rejected(t *testing.T) {
	f := NewFuture()
	f.Reject(errors.New("failed"))

	val, err := f.Await()
	if err == nil {
		t.Fatal("Await() should return error")
	}
	if val != nil {
		t.Errorf("Expected nil value, got %v", val)
	}
}

func TestFuture_Await_Async(t *testing.T) {
	f := NewFuture()

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.Resolve("async")
	}()

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}
	if val != "async" {
		t.Errorf("Expected 'async', got %v", val)
	}
}

func TestFuture_AwaitWithTimeout_Success(t *testing.T) {
	f := NewFuture()
	f.Resolve("fast")

	val, err := f.AwaitWithTimeout(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("AwaitWithTimeout() returned error: %v", err)
	}
	if val != "fast" {
		t.Errorf("Expected 'fast', got %v", val)
	}
}

func TestFuture_AwaitWithTimeout_Timeout(t *testing.T) {
	f := NewFuture() // Never resolved

	_, err := f.AwaitWithTimeout(10 * time.Millisecond)
	if err == nil {
		t.Fatal("AwaitWithTimeout() should return timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got %q", err.Error())
	}
}

func TestFuture_AwaitWithTimeout_Rejected(t *testing.T) {
	f := NewFuture()
	f.Reject(errors.New("rejected"))

	_, err := f.AwaitWithTimeout(100 * time.Millisecond)
	if err == nil {
		t.Fatal("AwaitWithTimeout() should return rejection error")
	}
	if err.Error() != "rejected" {
		t.Errorf("Expected 'rejected', got %q", err.Error())
	}
}

func TestFuture_Done(t *testing.T) {
	f := NewFuture()
	done := f.Done()

	select {
	case <-done:
		t.Fatal("Done channel should not be closed yet")
	default:
		// Expected
	}

	f.Resolve("done")

	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Done channel should be closed after resolve")
	}
}

func TestRunAsync_Success(t *testing.T) {
	f := RunAsync(func() (interface{}, error) {
		return "result", nil
	})

	val, err := f.AwaitWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Await returned error: %v", err)
	}
	if val != "result" {
		t.Errorf("Expected 'result', got %v", val)
	}
	if !f.IsResolved() {
		t.Error("Future should be resolved")
	}
}

func TestRunAsync_Error(t *testing.T) {
	f := RunAsync(func() (interface{}, error) {
		return nil, errors.New("async failure")
	})

	_, err := f.AwaitWithTimeout(1 * time.Second)
	if err == nil {
		t.Fatal("Expected error from RunAsync")
	}
	if err.Error() != "async failure" {
		t.Errorf("Expected 'async failure', got %q", err.Error())
	}
	if !f.IsRejected() {
		t.Error("Future should be rejected")
	}
}

func TestAll_Success(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()
	f3 := NewFuture()

	f1.Resolve("a")
	f2.Resolve("b")
	f3.Resolve("c")

	result := All(f1, f2, f3)
	val, err := result.AwaitWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("All() returned error: %v", err)
	}

	values, ok := val.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", val)
	}
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}
	if values[0] != "a" || values[1] != "b" || values[2] != "c" {
		t.Errorf("Expected [a, b, c], got %v", values)
	}
}

func TestAll_OneRejects(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()

	f1.Resolve("ok")
	f2.Reject(errors.New("fail"))

	result := All(f1, f2)
	_, err := result.AwaitWithTimeout(1 * time.Second)
	if err == nil {
		t.Fatal("All() should reject when any future rejects")
	}
	if err.Error() != "fail" {
		t.Errorf("Expected 'fail', got %q", err.Error())
	}
}

func TestRace_FirstResolves(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()

	f1.Resolve("winner")

	result := Race(f1, f2)
	val, err := result.AwaitWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Race() returned error: %v", err)
	}
	if val != "winner" {
		t.Errorf("Expected 'winner', got %v", val)
	}
}

func TestRace_FirstRejects(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()

	f1.Reject(errors.New("fast failure"))

	result := Race(f1, f2)
	_, err := result.AwaitWithTimeout(1 * time.Second)
	if err == nil {
		t.Fatal("Race() should propagate rejection")
	}
	if err.Error() != "fast failure" {
		t.Errorf("Expected 'fast failure', got %q", err.Error())
	}
}

func TestRace_Empty(t *testing.T) {
	result := Race()
	_, err := result.AwaitWithTimeout(1 * time.Second)
	if err == nil {
		t.Fatal("Race() with no futures should reject")
	}
	if !strings.Contains(err.Error(), "no futures") {
		t.Errorf("Expected 'no futures' error, got %q", err.Error())
	}
}

func TestAny_FirstResolves(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()

	f1.Reject(errors.New("fail 1"))
	f2.Resolve("success")

	result := Any(f1, f2)
	val, err := result.AwaitWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Any() returned error: %v", err)
	}
	if val != "success" {
		t.Errorf("Expected 'success', got %v", val)
	}
}

func TestAny_AllReject(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()

	f1.Reject(errors.New("fail 1"))
	f2.Reject(errors.New("fail 2"))

	result := Any(f1, f2)
	_, err := result.AwaitWithTimeout(1 * time.Second)
	if err == nil {
		t.Fatal("Any() should reject when all futures reject")
	}
	if !strings.Contains(err.Error(), "all futures rejected") {
		t.Errorf("Expected 'all futures rejected' error, got %q", err.Error())
	}
}

func TestAny_Empty(t *testing.T) {
	result := Any()
	_, err := result.AwaitWithTimeout(1 * time.Second)
	if err == nil {
		t.Fatal("Any() with no futures should reject")
	}
	if !strings.Contains(err.Error(), "no futures") {
		t.Errorf("Expected 'no futures' error, got %q", err.Error())
	}
}

func TestIsFuture_Extended(t *testing.T) {
	f := NewFuture()
	if !IsFuture(f) {
		t.Error("IsFuture should return true for *Future")
	}

	if IsFuture("not a future") {
		t.Error("IsFuture should return false for string")
	}

	if IsFuture(42) {
		t.Error("IsFuture should return false for int")
	}

	if IsFuture(nil) {
		t.Error("IsFuture should return false for nil")
	}
}

func TestFutureState_StringValues(t *testing.T) {
	tests := []struct {
		state    FutureState
		expected string
	}{
		{FuturePending, "pending"},
		{FutureResolved, "resolved"},
		{FutureRejected, "rejected"},
		{FutureState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("String() = %q, want %q", tt.state.String(), tt.expected)
			}
		})
	}
}

func TestFuture_ConcurrentAccess(t *testing.T) {
	f := NewFuture()

	// Spawn multiple goroutines trying to resolve/reject concurrently
	for i := 0; i < 10; i++ {
		go func(val int) {
			f.Resolve(val)
		}(i)
		go func(val int) {
			f.Reject(errors.New("error"))
		}(i)
	}

	// Should not panic
	val, err := f.AwaitWithTimeout(1 * time.Second)
	// One of them should have won
	if err == nil && val == nil {
		t.Error("Expected either a value or error")
	}

	// State should be consistent
	if f.IsPending() {
		t.Error("Future should not be pending after concurrent resolve/reject")
	}
}
