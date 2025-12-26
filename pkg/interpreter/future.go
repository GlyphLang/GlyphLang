package interpreter

import (
	"fmt"
	"sync"
	"time"
)

// FutureState represents the current state of a Future
type FutureState int

const (
	FuturePending FutureState = iota
	FutureResolved
	FutureRejected
)

// Future represents a value that will be available in the future.
// It provides non-blocking I/O and concurrent request handling using Go channels.
type Future struct {
	value    interface{}
	err      error
	state    FutureState
	done     chan struct{}
	mu       sync.RWMutex
	resolved bool
}

// NewFuture creates a new pending Future
func NewFuture() *Future {
	return &Future{
		state: FuturePending,
		done:  make(chan struct{}),
	}
}

// Resolve sets the value of the Future and marks it as resolved
func (f *Future) Resolve(value interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.resolved {
		return // Already resolved or rejected
	}

	f.value = value
	f.state = FutureResolved
	f.resolved = true
	close(f.done)
}

// Reject sets an error on the Future and marks it as rejected
func (f *Future) Reject(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.resolved {
		return // Already resolved or rejected
	}

	f.err = err
	f.state = FutureRejected
	f.resolved = true
	close(f.done)
}

// Await blocks until the Future is resolved or rejected, then returns the value or error
func (f *Future) Await() (interface{}, error) {
	<-f.done

	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.state == FutureRejected {
		return nil, f.err
	}
	return f.value, nil
}

// AwaitWithTimeout blocks until the Future is resolved, rejected, or timeout occurs
func (f *Future) AwaitWithTimeout(timeout time.Duration) (interface{}, error) {
	select {
	case <-f.done:
		f.mu.RLock()
		defer f.mu.RUnlock()

		if f.state == FutureRejected {
			return nil, f.err
		}
		return f.value, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("future timed out after %v", timeout)
	}
}

// IsResolved returns true if the Future has been resolved (not rejected)
func (f *Future) IsResolved() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == FutureResolved
}

// IsRejected returns true if the Future has been rejected
func (f *Future) IsRejected() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == FutureRejected
}

// IsPending returns true if the Future is still pending
func (f *Future) IsPending() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == FuturePending
}

// Done returns a channel that is closed when the Future completes
func (f *Future) Done() <-chan struct{} {
	return f.done
}

// State returns the current state of the Future
func (f *Future) State() FutureState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// Value returns the resolved value (nil if pending or rejected)
func (f *Future) Value() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.value
}

// Error returns the rejection error (nil if pending or resolved)
func (f *Future) Error() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.err
}

// String returns a string representation of the Future state
func (f *Future) String() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	switch f.state {
	case FuturePending:
		return "Future<pending>"
	case FutureResolved:
		return fmt.Sprintf("Future<resolved: %v>", f.value)
	case FutureRejected:
		return fmt.Sprintf("Future<rejected: %v>", f.err)
	default:
		return "Future<unknown>"
	}
}

// IsFuture checks if a value is a Future
func IsFuture(v interface{}) bool {
	_, ok := v.(*Future)
	return ok
}

// RunAsync executes a function asynchronously and returns a Future
// This is a helper function that can be used to wrap synchronous operations
func RunAsync(fn func() (interface{}, error)) *Future {
	future := NewFuture()

	go func() {
		result, err := fn()
		if err != nil {
			future.Reject(err)
		} else {
			future.Resolve(result)
		}
	}()

	return future
}

// All waits for all futures to complete and returns a slice of their values
// If any future is rejected, the first error is returned
func All(futures ...*Future) ([]interface{}, error) {
	results := make([]interface{}, len(futures))

	for i, f := range futures {
		val, err := f.Await()
		if err != nil {
			return nil, err
		}
		results[i] = val
	}

	return results, nil
}

// Race returns the value of the first future to complete
// If the first to complete is rejected, its error is returned
func Race(futures ...*Future) (interface{}, error) {
	if len(futures) == 0 {
		return nil, fmt.Errorf("Race requires at least one future")
	}

	result := make(chan struct {
		value interface{}
		err   error
	}, 1)

	for _, f := range futures {
		go func(fut *Future) {
			val, err := fut.Await()
			select {
			case result <- struct {
				value interface{}
				err   error
			}{val, err}:
			default:
			}
		}(f)
	}

	r := <-result
	return r.value, r.err
}

// Any returns the value of the first future to resolve successfully
// If all futures are rejected, returns an error
func Any(futures ...*Future) (interface{}, error) {
	if len(futures) == 0 {
		return nil, fmt.Errorf("Any requires at least one future")
	}

	var wg sync.WaitGroup
	successChan := make(chan interface{}, 1)
	errorCount := 0
	var mu sync.Mutex

	for _, f := range futures {
		wg.Add(1)
		go func(fut *Future) {
			defer wg.Done()
			val, err := fut.Await()
			if err == nil {
				select {
				case successChan <- val:
				default:
				}
			} else {
				mu.Lock()
				errorCount++
				mu.Unlock()
			}
		}(f)
	}

	// Wait for completion in a separate goroutine
	go func() {
		wg.Wait()
		close(successChan)
	}()

	// Wait for first success or all failures
	if val, ok := <-successChan; ok {
		return val, nil
	}

	return nil, fmt.Errorf("all futures were rejected")
}
