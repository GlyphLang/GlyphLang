package interpreter

import (
	"context"
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

// Future represents an asynchronous computation that will complete in the future.
// It can be in one of three states: pending, resolved, or rejected.
type Future struct {
	state    FutureState
	value    interface{}
	err      error
	done     chan struct{}
	cancel   chan struct{}
	mu       sync.RWMutex
	resolved bool // Double-resolve protection
}

// NewFuture creates a new pending Future
func NewFuture() *Future {
	return &Future{
		state:  FuturePending,
		done:   make(chan struct{}),
		cancel: make(chan struct{}),
	}
}

// Resolve completes the Future with a successful value.
// Can only be called once; subsequent calls are ignored.
func (f *Future) Resolve(value interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.resolved {
		return // Already resolved/rejected
	}

	f.resolved = true
	f.state = FutureResolved
	f.value = value
	close(f.done)
}

// Reject completes the Future with an error.
// Can only be called once; subsequent calls are ignored.
func (f *Future) Reject(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.resolved {
		return // Already resolved/rejected
	}

	f.resolved = true
	f.state = FutureRejected
	f.err = err
	close(f.done)
}

// Await blocks until the Future completes and returns the value or error.
func (f *Future) Await() (interface{}, error) {
	<-f.done

	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.state == FutureRejected {
		return nil, f.err
	}
	return f.value, nil
}

// AwaitWithTimeout blocks until the Future completes or the timeout expires.
// Returns an error if the timeout is reached before completion.
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

// Cancel signals the goroutine associated with this Future to stop.
// If the Future is still pending, it will be rejected with a cancellation error.
func (f *Future) Cancel() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.resolved {
		return
	}

	// Signal cancellation to the goroutine
	select {
	case <-f.cancel:
		// Already cancelled
	default:
		close(f.cancel)
	}

	f.resolved = true
	f.state = FutureRejected
	f.err = fmt.Errorf("future cancelled")
	close(f.done)
}

// Cancelled returns a channel that is closed when the Future is cancelled.
// Goroutines should select on this to detect cancellation.
func (f *Future) Cancelled() <-chan struct{} {
	return f.cancel
}

// AwaitWithContext blocks until the Future completes or the context is cancelled.
func (f *Future) AwaitWithContext(ctx context.Context) (interface{}, error) {
	select {
	case <-f.done:
		f.mu.RLock()
		defer f.mu.RUnlock()

		if f.state == FutureRejected {
			return nil, f.err
		}
		return f.value, nil
	case <-ctx.Done():
		f.Cancel()
		return nil, ctx.Err()
	}
}

// IsResolved returns true if the Future has completed successfully
func (f *Future) IsResolved() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == FutureResolved
}

// IsRejected returns true if the Future has completed with an error
func (f *Future) IsRejected() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == FutureRejected
}

// IsPending returns true if the Future has not yet completed
func (f *Future) IsPending() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state == FuturePending
}

// State returns the current state of the Future
func (f *Future) State() FutureState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// Value returns the resolved value. Panics if the Future is not resolved.
func (f *Future) Value() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.state != FutureResolved {
		return nil
	}
	return f.value
}

// Error returns the rejection error. Returns nil if the Future is not rejected.
func (f *Future) Error() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.state != FutureRejected {
		return nil
	}
	return f.err
}

// Done returns a channel that is closed when the Future completes.
// This allows for select-based waiting on multiple Futures.
func (f *Future) Done() <-chan struct{} {
	return f.done
}

// RunAsync executes a function asynchronously and returns a Future for its result.
func RunAsync(fn func() (interface{}, error)) *Future {
	future := NewFuture()

	go func() {
		// Check for cancellation before running
		select {
		case <-future.Cancelled():
			return
		default:
		}

		value, err := fn()

		// Check for cancellation after running
		select {
		case <-future.Cancelled():
			return
		default:
		}

		if err != nil {
			future.Reject(err)
		} else {
			future.Resolve(value)
		}
	}()

	return future
}

// All waits for all Futures to complete and returns a Future containing
// a slice of all values. If any Future rejects, the returned Future rejects
// with that error and remaining futures are cancelled.
func All(futures ...*Future) *Future {
	result := NewFuture()

	go func() {
		values := make([]interface{}, len(futures))

		for idx, f := range futures {
			value, err := f.Await()
			if err != nil {
				// Cancel remaining futures
				for _, remaining := range futures[idx+1:] {
					remaining.Cancel()
				}
				result.Reject(err)
				return
			}
			values[idx] = value
		}

		result.Resolve(values)
	}()

	return result
}

// Race returns a Future that completes with the first Future to complete,
// whether it resolves or rejects. Losing futures are cancelled.
func Race(futures ...*Future) *Future {
	if len(futures) == 0 {
		result := NewFuture()
		result.Reject(fmt.Errorf("Race called with no futures"))
		return result
	}

	result := NewFuture()

	for _, f := range futures {
		go func(future *Future) {
			value, err := future.Await()
			if err != nil {
				result.Reject(err)
			} else {
				result.Resolve(value)
			}
		}(f)
	}

	// Cancel losing futures once the race is decided
	go func() {
		<-result.Done()
		for _, f := range futures {
			if f.IsPending() {
				f.Cancel()
			}
		}
	}()

	return result
}

// Any returns a Future that completes with the first Future to resolve
// successfully. If all Futures reject, the returned Future rejects with
// an error containing all rejection errors.
func Any(futures ...*Future) *Future {
	if len(futures) == 0 {
		result := NewFuture()
		result.Reject(fmt.Errorf("Any called with no futures"))
		return result
	}

	result := NewFuture()
	errors := make([]error, len(futures))
	var errorCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(futures))

	for i, f := range futures {
		go func(index int, future *Future) {
			defer wg.Done()

			value, err := future.Await()

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors[index] = err
				errorCount++

				// If all futures have failed, reject with aggregate error
				if errorCount == len(futures) {
					result.Reject(fmt.Errorf("all futures rejected: %v", errors))
				}
			} else {
				// First successful resolution wins
				result.Resolve(value)
			}
		}(i, f)
	}

	return result
}

// IsFuture checks if a value is a Future
func IsFuture(v interface{}) bool {
	_, ok := v.(*Future)
	return ok
}

// String returns a string representation of the FutureState
func (s FutureState) String() string {
	switch s {
	case FuturePending:
		return "pending"
	case FutureResolved:
		return "resolved"
	case FutureRejected:
		return "rejected"
	default:
		return "unknown"
	}
}
