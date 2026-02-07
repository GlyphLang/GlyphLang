package interpreter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuture_Cancel_Pending(t *testing.T) {
	f := NewFuture()
	f.Cancel()

	assert.True(t, f.IsRejected())
	assert.False(t, f.IsPending())
	assert.Contains(t, f.Error().Error(), "cancelled")
}

func TestFuture_Cancel_AlreadyResolved(t *testing.T) {
	f := NewFuture()
	f.Resolve("done")
	f.Cancel() // Should be ignored

	assert.True(t, f.IsResolved())
	assert.Equal(t, "done", f.Value())
}

func TestFuture_Cancel_AlreadyCancelled(t *testing.T) {
	f := NewFuture()
	f.Cancel()
	f.Cancel() // Should not panic

	assert.True(t, f.IsRejected())
}

func TestFuture_Cancelled_Channel(t *testing.T) {
	f := NewFuture()

	// Not cancelled yet
	select {
	case <-f.Cancelled():
		t.Fatal("Cancelled channel should not be closed yet")
	default:
	}

	f.Cancel()

	// Now cancelled
	select {
	case <-f.Cancelled():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Cancelled channel should be closed after Cancel()")
	}
}

func TestFuture_Cancel_StopsGoroutine(t *testing.T) {
	f := NewFuture()
	started := make(chan struct{})
	finished := make(chan struct{})

	go func() {
		close(started)
		// Simulate long work with cancellation check
		select {
		case <-f.Cancelled():
			close(finished)
			return
		case <-time.After(5 * time.Second):
			f.Resolve("completed")
			close(finished)
			return
		}
	}()

	// Wait for goroutine to start
	<-started

	// Cancel the future
	f.Cancel()

	// The goroutine should detect cancellation
	select {
	case <-finished:
		// Goroutine exited
	case <-time.After(1 * time.Second):
		t.Fatal("Goroutine should have exited after cancellation")
	}

	assert.True(t, f.IsRejected())
}

func TestFuture_AwaitWithContext_Success(t *testing.T) {
	f := NewFuture()
	f.Resolve("hello")

	ctx := context.Background()
	val, err := f.AwaitWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestFuture_AwaitWithContext_Cancelled(t *testing.T) {
	f := NewFuture() // Never resolved

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := f.AwaitWithContext(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.True(t, f.IsRejected())
}

func TestFuture_AwaitWithContext_ManualCancel(t *testing.T) {
	f := NewFuture() // Never resolved

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel from another goroutine after a short delay
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := f.AwaitWithContext(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRace_CancelsLosers(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()
	f3 := NewFuture()

	f1.Resolve("winner")

	result := Race(f1, f2, f3)
	val, err := result.AwaitWithTimeout(1 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, "winner", val)

	// Give the cancellation goroutine time to run
	time.Sleep(50 * time.Millisecond)

	// Losing futures should be cancelled
	assert.True(t, f2.IsRejected(), "f2 should be cancelled")
	assert.True(t, f3.IsRejected(), "f3 should be cancelled")
}

func TestAll_CancelsRemainingOnFailure(t *testing.T) {
	f1 := NewFuture()
	f2 := NewFuture()
	f3 := NewFuture()

	f1.Resolve("ok")
	f2.Reject(assert.AnError)
	// f3 is still pending

	result := All(f1, f2, f3)
	_, err := result.AwaitWithTimeout(1 * time.Second)
	require.Error(t, err)

	// f3 should be cancelled since f2 failed
	assert.True(t, f3.IsRejected(), "f3 should be cancelled after f2 failed")
}
