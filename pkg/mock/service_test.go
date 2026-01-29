package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_OnAndCall(t *testing.T) {
	svc := NewService()
	svc.On("GetUser", map[string]interface{}{"id": 1, "name": "Alice"})

	result, err := svc.CallMethod("GetUser", 1)
	require.NoError(t, err)

	user, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, user["id"])
	assert.Equal(t, "Alice", user["name"])
}

func TestService_OnFunc(t *testing.T) {
	svc := NewService()
	svc.OnFunc("Add", func(args ...interface{}) interface{} {
		a := args[0].(int)
		b := args[1].(int)
		return a + b
	})

	result, err := svc.CallMethod("Add", 3, 4)
	require.NoError(t, err)
	assert.Equal(t, 7, result)
}

func TestService_CallUnstubbedMethod(t *testing.T) {
	svc := NewService()
	_, err := svc.CallMethod("Unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no stub defined")
}

func TestService_Called(t *testing.T) {
	svc := NewService()
	svc.On("Ping", "pong")

	assert.False(t, svc.Called("Ping"))

	_, _ = svc.CallMethod("Ping")
	assert.True(t, svc.Called("Ping"))
}

func TestService_CalledOnce(t *testing.T) {
	svc := NewService()
	svc.On("Ping", "pong")

	_, _ = svc.CallMethod("Ping")
	assert.True(t, svc.CalledOnce("Ping"))

	_, _ = svc.CallMethod("Ping")
	assert.False(t, svc.CalledOnce("Ping"))
}

func TestService_CalledTimes(t *testing.T) {
	svc := NewService()
	svc.On("Fetch", "data")

	_, _ = svc.CallMethod("Fetch")
	_, _ = svc.CallMethod("Fetch")
	_, _ = svc.CallMethod("Fetch")

	assert.Equal(t, 3, svc.CalledTimes("Fetch"))
	assert.Equal(t, 0, svc.CalledTimes("Other"))
}

func TestService_Calls(t *testing.T) {
	svc := NewService()
	svc.On("Save", true)

	_, _ = svc.CallMethod("Save", "doc1")
	_, _ = svc.CallMethod("Save", "doc2")

	calls := svc.Calls("Save")
	require.Len(t, calls, 2)
	assert.Equal(t, []interface{}{"doc1"}, calls[0].Args)
	assert.Equal(t, []interface{}{"doc2"}, calls[1].Args)
}

func TestService_LastCall(t *testing.T) {
	svc := NewService()
	svc.On("Send", true)

	assert.Nil(t, svc.LastCall("Send"))

	_, _ = svc.CallMethod("Send", "msg1")
	_, _ = svc.CallMethod("Send", "msg2")

	last := svc.LastCall("Send")
	require.NotNil(t, last)
	assert.Equal(t, []interface{}{"msg2"}, last.Args)
}

func TestService_Reset(t *testing.T) {
	svc := NewService()
	svc.On("Ping", "pong")

	_, _ = svc.CallMethod("Ping")
	assert.True(t, svc.Called("Ping"))

	svc.Reset()
	assert.False(t, svc.Called("Ping"))

	// Stubs should still work after reset
	result, err := svc.CallMethod("Ping")
	require.NoError(t, err)
	assert.Equal(t, "pong", result)
}

func TestService_ResetAll(t *testing.T) {
	svc := NewService()
	svc.On("Ping", "pong")

	_, _ = svc.CallMethod("Ping")
	svc.ResetAll()

	assert.False(t, svc.Called("Ping"))

	// Stubs should be gone
	_, err := svc.CallMethod("Ping")
	assert.Error(t, err)
}

func TestService_Table(t *testing.T) {
	svc := NewService()
	users := svc.Table("users")

	users.On("Get", map[string]interface{}{"id": 1, "name": "Alice"})
	result := users.Get(1)

	user, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Alice", user["name"])
	assert.True(t, users.Called("Get"))
}

func TestService_Table_All(t *testing.T) {
	svc := NewService()
	users := svc.Table("users")

	allUsers := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
	}
	users.On("All", allUsers)

	result := users.All()
	assert.Equal(t, allUsers, result)
	assert.Equal(t, 1, users.CalledTimes("All"))
}

func TestService_Table_Create(t *testing.T) {
	svc := NewService()
	users := svc.Table("users")

	created := map[string]interface{}{"id": 3, "name": "Charlie"}
	users.On("Create", created)

	result := users.Create(map[string]interface{}{"name": "Charlie"})
	assert.Equal(t, created, result)
	assert.True(t, users.Called("Create"))
}

func TestService_Table_Delete(t *testing.T) {
	svc := NewService()
	users := svc.Table("users")

	result := users.Delete(1)
	assert.True(t, result)
	assert.True(t, users.Called("Delete"))
}

func TestService_Table_DeleteStubbed(t *testing.T) {
	svc := NewService()
	users := svc.Table("users")
	users.On("Delete", false)

	result := users.Delete(1)
	assert.False(t, result)
}

func TestService_Table_Reset(t *testing.T) {
	svc := NewService()
	users := svc.Table("users")

	users.Get(1)
	assert.True(t, users.Called("Get"))

	users.Reset()
	assert.False(t, users.Called("Get"))
}

func TestService_Table_Reuse(t *testing.T) {
	svc := NewService()

	t1 := svc.Table("users")
	t1.Get(1)

	t2 := svc.Table("users")
	assert.True(t, t2.Called("Get"))
}

func TestService_MultipleStubs(t *testing.T) {
	svc := NewService()
	svc.On("GetUser", map[string]interface{}{"name": "Alice"})
	svc.On("GetCount", int64(42))

	user, err := svc.CallMethod("GetUser")
	require.NoError(t, err)

	count, err := svc.CallMethod("GetCount")
	require.NoError(t, err)

	assert.Equal(t, map[string]interface{}{"name": "Alice"}, user)
	assert.Equal(t, int64(42), count)
}

func TestService_TrackUnstubbedCalls(t *testing.T) {
	svc := NewService()

	// Even unstubbed calls are tracked
	_, _ = svc.CallMethod("MissingMethod", "arg1")
	assert.True(t, svc.Called("MissingMethod"))
	assert.Equal(t, 1, svc.CalledTimes("MissingMethod"))
}
