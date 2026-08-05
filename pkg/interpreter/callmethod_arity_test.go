package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// arityTarget stands in for a provider handler: one method taking a single
// argument, one variadic.
type arityTarget struct{}

func (arityTarget) Post(args interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": args}, nil
}

func (arityTarget) Del(keys ...interface{}) (int64, error) {
	return int64(len(keys)), nil
}

// TestCallMethodRejectsWrongArity covers the crash path: reflect.Call panics on
// an argument-count mismatch, which took down the whole request handler instead
// of surfacing an error the route could report. Calling http.post with the
// two-argument form its own docs used to show was enough to trigger it.
func TestCallMethodRejectsWrongArity(t *testing.T) {
	_, err := CallMethod(arityTarget{}, "Post", "https://example.com", map[string]interface{}{"body": 1})
	require.Error(t, err, "a two-argument call to a one-argument method must error, not panic")
	assert.Contains(t, err.Error(), "expects 1 argument(s), got 2")

	_, err = CallMethod(arityTarget{}, "Post")
	require.Error(t, err, "a zero-argument call to a one-argument method must error")
}

// TestCallMethodAcceptsCorrectArity keeps the guard from over-rejecting, and
// covers variadic methods, which Redis uses heavily (Del, Exists, HSet).
func TestCallMethodAcceptsCorrectArity(t *testing.T) {
	result, err := CallMethod(arityTarget{}, "Post", map[string]interface{}{"url": "https://example.com"})
	require.NoError(t, err)
	assert.NotNil(t, result)

	for _, args := range [][]interface{}{
		{},
		{"one"},
		{"one", "two", "three"},
	} {
		count, err := CallMethod(arityTarget{}, "Del", args...)
		require.NoError(t, err, "variadic call with %d args", len(args))
		assert.Equal(t, int64(len(args)), count)
	}
}
