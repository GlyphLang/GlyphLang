package vm

import (
	"encoding/json"
	"testing"
)

func TestValueMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"null", NullValue{}, "null"},
		{"int", IntValue{Val: 42}, "42"},
		{"float", FloatValue{Val: 3.14}, "3.14"},
		{"string", StringValue{Val: "hello"}, `"hello"`},
		{"bool_true", BoolValue{Val: true}, "true"},
		{"bool_false", BoolValue{Val: false}, "false"},
		{"array", ArrayValue{Val: []Value{IntValue{Val: 1}, IntValue{Val: 2}}}, "[1,2]"},
		{"object", ObjectValue{Val: map[string]Value{"x": IntValue{Val: 10}}}, `{"x":10}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}
			if string(b) != tt.expected {
				t.Errorf("got %s, want %s", string(b), tt.expected)
			}
		})
	}
}

func TestNestedObjectMarshalJSON(t *testing.T) {
	obj := ObjectValue{
		Val: map[string]Value{
			"user": ObjectValue{
				Val: map[string]Value{
					"name": StringValue{Val: "John"},
					"age":  IntValue{Val: 30},
				},
			},
			"scores": ArrayValue{
				Val: []Value{IntValue{Val: 95}, IntValue{Val: 87}},
			},
		},
	}

	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify no "Val" keys appear
	if _, hasVal := result["Val"]; hasVal {
		t.Error("JSON should not contain 'Val' key at top level")
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user should be an object")
	}
	if _, hasVal := user["Val"]; hasVal {
		t.Error("nested object should not contain 'Val' key")
	}
	if user["name"] != "John" {
		t.Errorf("user.name = %v, want John", user["name"])
	}
}
