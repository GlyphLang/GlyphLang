package vm

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestFutureValue_Type(t *testing.T) {
	f := &FutureValue{
		Done: make(chan struct{}),
	}
	if f.Type() != "future" {
		t.Errorf("FutureValue.Type() = %q, want %q", f.Type(), "future")
	}
}

func TestFutureValue_MarshalJSON_Pending(t *testing.T) {
	f := &FutureValue{
		Done: make(chan struct{}),
	}

	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `{"pending":true}`
	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}

func TestFutureValue_MarshalJSON_Resolved(t *testing.T) {
	f := &FutureValue{
		Result: IntValue{Val: 42},
		Done:   make(chan struct{}),
	}
	close(f.Done)

	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := "42"
	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}

func TestFutureValue_MarshalJSON_ResolvedString(t *testing.T) {
	f := &FutureValue{
		Result: StringValue{Val: "hello"},
		Done:   make(chan struct{}),
	}
	close(f.Done)

	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"hello"`
	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}

func TestFutureValue_MarshalJSON_ResolvedNull(t *testing.T) {
	f := &FutureValue{
		Result: nil,
		Done:   make(chan struct{}),
	}
	close(f.Done)

	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := "null"
	if string(b) != expected {
		t.Errorf("got %s, want %s", string(b), expected)
	}
}

func TestFutureValue_Await_Success(t *testing.T) {
	f := &FutureValue{
		Result: IntValue{Val: 42},
		Done:   make(chan struct{}),
	}
	close(f.Done)

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}

	intVal, ok := val.(IntValue)
	if !ok {
		t.Fatalf("Expected IntValue, got %T", val)
	}
	if intVal.Val != 42 {
		t.Errorf("Expected 42, got %d", intVal.Val)
	}
}

func TestFutureValue_Await_Error(t *testing.T) {
	expectedErr := errors.New("async operation failed")
	f := &FutureValue{
		Error: expectedErr,
		Done:  make(chan struct{}),
	}
	close(f.Done)

	val, err := f.Await()
	if err == nil {
		t.Fatal("Await() should have returned an error")
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if val != nil {
		t.Errorf("Expected nil value, got %v", val)
	}
}

func TestFutureValue_Await_NilDone(t *testing.T) {
	f := &FutureValue{
		Result: IntValue{Val: 99},
		Done:   nil,
	}

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}

	intVal, ok := val.(IntValue)
	if !ok {
		t.Fatalf("Expected IntValue, got %T", val)
	}
	if intVal.Val != 99 {
		t.Errorf("Expected 99, got %d", intVal.Val)
	}
}

func TestFutureValue_Await_AsyncResolve(t *testing.T) {
	f := &FutureValue{
		Done: make(chan struct{}),
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.Result = StringValue{Val: "async result"}
		close(f.Done)
	}()

	val, err := f.Await()
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}

	strVal, ok := val.(StringValue)
	if !ok {
		t.Fatalf("Expected StringValue, got %T", val)
	}
	if strVal.Val != "async result" {
		t.Errorf("Expected 'async result', got %q", strVal.Val)
	}
}

func TestFutureValue_Await_AsyncError(t *testing.T) {
	f := &FutureValue{
		Done: make(chan struct{}),
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.Error = errors.New("async error")
		close(f.Done)
	}()

	val, err := f.Await()
	if err == nil {
		t.Fatal("Await() should have returned an error")
	}
	if err.Error() != "async error" {
		t.Errorf("Expected 'async error', got %q", err.Error())
	}
	if val != nil {
		t.Errorf("Expected nil value, got %v", val)
	}
}

func TestDefaultAsyncTimeout(t *testing.T) {
	if DefaultAsyncTimeout != 30*time.Second {
		t.Errorf("DefaultAsyncTimeout = %v, want 30s", DefaultAsyncTimeout)
	}
}
