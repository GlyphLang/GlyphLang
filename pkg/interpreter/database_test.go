package interpreter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestType is a test type for method calls
type TestType struct {
	value string
}

func (t TestType) String() string {
	return t.value
}

func (t TestType) Len() int {
	return len(t.value)
}

func (t TestType) IsZero() bool {
	return t.value == ""
}

func (t TestType) Get(key string) (string, error) {
	if key == "error" {
		return "", errors.New("test error")
	}
	return t.value + key, nil
}

func (t TestType) DangerousMethod() string {
	return "should not be callable"
}

func (t TestType) ExecuteCommand(cmd string) string {
	return "should not be callable: " + cmd
}

func TestCallMethod_AllowedMethods(t *testing.T) {
	tests := []struct {
		name       string
		methodName string
		wantErr    bool
	}{
		{name: "String is allowed", methodName: "String", wantErr: false},
		{name: "Get is allowed", methodName: "Get", wantErr: false},
		{name: "Len is allowed", methodName: "Len", wantErr: false},
		{name: "IsZero is allowed", methodName: "IsZero", wantErr: false},
		{name: "Find is allowed", methodName: "Find", wantErr: false},
		{name: "Create is allowed", methodName: "Create", wantErr: false},
		{name: "Update is allowed", methodName: "Update", wantErr: false},
		{name: "Delete is allowed", methodName: "Delete", wantErr: false},
		{name: "First is allowed", methodName: "First", wantErr: false},
		{name: "All is allowed", methodName: "All", wantErr: false},
		{name: "Where is allowed", methodName: "Where", wantErr: false},
		{name: "Count is allowed", methodName: "Count", wantErr: false},
		{name: "Save is allowed", methodName: "Save", wantErr: false},
		{name: "Insert is allowed", methodName: "Insert", wantErr: false},
		{name: "Select is allowed", methodName: "Select", wantErr: false},
		{name: "Limit is allowed", methodName: "Limit", wantErr: false},
		{name: "Offset is allowed", methodName: "Offset", wantErr: false},
		{name: "Order is allowed", methodName: "Order", wantErr: false},
		{name: "Int is allowed", methodName: "Int", wantErr: false},
		{name: "Bool is allowed", methodName: "Bool", wantErr: false},
		{name: "Float is allowed", methodName: "Float", wantErr: false},
	}

	obj := TestType{value: "test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For methods that exist on TestType, test them
			switch tt.methodName {
			case "String", "Len", "IsZero":
				_, err := CallMethod(obj, tt.methodName)
				assert.NoError(t, err)
			case "Get":
				// Get requires an argument
				_, err := CallMethod(obj, "Get", "key")
				assert.NoError(t, err)
			default:
				// Method doesn't exist on TestType but should be in whitelist
				_, err := CallMethod(obj, tt.methodName)
				// Should fail because method not found, NOT because it's blocked
				if err != nil {
					assert.Contains(t, err.Error(), "not found")
				}
			}
		})
	}
}

func TestCallMethod_BlockedMethods(t *testing.T) {
	tests := []struct {
		name       string
		methodName string
	}{
		{name: "DangerousMethod is blocked", methodName: "DangerousMethod"},
		{name: "ExecuteCommand is blocked", methodName: "ExecuteCommand"},
		{name: "Exec is blocked", methodName: "Exec"},
		{name: "Run is blocked", methodName: "Run"},
		{name: "Call is blocked", methodName: "Call"},
		{name: "Invoke is blocked", methodName: "Invoke"},
		{name: "System is blocked", methodName: "System"},
		{name: "Shell is blocked", methodName: "Shell"},
		{name: "Eval is blocked", methodName: "Eval"},
		{name: "Execute is blocked", methodName: "Execute"},
		{name: "arbitrary method is blocked", methodName: "ArbitraryMethod"},
	}

	obj := TestType{value: "test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CallMethod(obj, tt.methodName)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not allowed")
		})
	}
}

func TestCallMethod_WithValidMethod(t *testing.T) {
	obj := TestType{value: "hello"}

	// Test String method
	result, err := CallMethod(obj, "String")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)

	// Test Len method
	result, err = CallMethod(obj, "Len")
	require.NoError(t, err)
	assert.Equal(t, 5, result)

	// Test IsZero method
	result, err = CallMethod(obj, "IsZero")
	require.NoError(t, err)
	assert.Equal(t, false, result)

	// Test with empty value
	emptyObj := TestType{value: ""}
	result, err = CallMethod(emptyObj, "IsZero")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestCallMethod_WithErrorReturn(t *testing.T) {
	obj := TestType{value: "test"}

	// Test Get method with valid key
	result, err := CallMethod(obj, "Get", "suffix")
	require.NoError(t, err)
	assert.Equal(t, "testsuffix", result)

	// Test Get method that returns error
	_, err = CallMethod(obj, "Get", "error")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestHasMethod_SecurityContext(t *testing.T) {
	obj := TestType{value: "test"}

	assert.True(t, HasMethod(obj, "String"))
	assert.True(t, HasMethod(obj, "Len"))
	assert.True(t, HasMethod(obj, "Get"))
	assert.True(t, HasMethod(obj, "IsZero"))
	assert.True(t, HasMethod(obj, "DangerousMethod"))
	assert.False(t, HasMethod(obj, "NonExistentMethod"))
}

func TestGetMethodNames_SecurityContext(t *testing.T) {
	obj := TestType{value: "test"}

	methods := GetMethodNames(obj)
	assert.Contains(t, methods, "String")
	assert.Contains(t, methods, "Len")
	assert.Contains(t, methods, "Get")
	assert.Contains(t, methods, "IsZero")
	assert.Contains(t, methods, "DangerousMethod")
	assert.Contains(t, methods, "ExecuteCommand")
}

func TestAllowedMethodsWhitelist(t *testing.T) {
	// Verify the whitelist contains expected database/ORM methods
	expectedMethods := []string{
		"Get", "Find", "Create", "Update", "Delete",
		"First", "All", "Where", "Count",
		"Save", "Insert", "Select", "Limit", "Offset", "Order",
		"String", "Int", "Bool", "Float", "Len", "IsZero",
	}

	for _, method := range expectedMethods {
		assert.True(t, allowedMethods[method], "Expected %s to be in whitelist", method)
	}

	// Verify dangerous methods are NOT in whitelist
	dangerousMethods := []string{
		"Exec", "Run", "Call", "Invoke", "System", "Shell", "Eval", "Execute",
		"DangerousMethod", "ArbitraryMethod", "UnsafeCall", "Query",
	}

	for _, method := range dangerousMethods {
		assert.False(t, allowedMethods[method], "Expected %s to NOT be in whitelist", method)
	}
}
