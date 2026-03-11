package vm

import (
	"math"
	"strings"
	"testing"
)

// TestOpModIntInt tests integer % integer modulo operations.
func TestOpModIntInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		expected int64
	}{
		{"basic modulo 10 % 3 = 1", 10, 3, 1},
		{"basic modulo 7 % 4 = 3", 7, 4, 3},
		{"exact divisor 9 % 3 = 0", 9, 3, 0},
		{"modulo by 1", 42, 1, 0},
		{"self modulo n % n = 0", 7, 7, 0},
		{"zero modulo 0 % n = 0", 0, 5, 0},
		{"negative dividend -10 % 3", -10, 3, -1},
		{"negative divisor 10 % -3", 10, -3, 1},
		{"both negative -10 % -3", -10, -3, -1},
		{"large values", 1000000007, 1000000, 7},
		{"one % large", 1, 999999, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constants := []Value{IntValue{Val: tt.a}, IntValue{Val: tt.b}}
			bytecode := createBytecodeHeader(constants)

			operand0 := uint32(0)
			operand1 := uint32(1)
			bytecode = addInstruction(bytecode, OpPush, &operand0)
			bytecode = addInstruction(bytecode, OpPush, &operand1)
			bytecode = addInstruction(bytecode, OpMod, nil)
			bytecode = addInstruction(bytecode, OpHalt, nil)

			vm := NewVM()
			result, err := vm.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			intVal, ok := result.(IntValue)
			if !ok {
				t.Fatalf("Expected IntValue, got %T", result)
			}
			if intVal.Val != tt.expected {
				t.Errorf("Expected %d %% %d = %d, got %d", tt.a, tt.b, tt.expected, intVal.Val)
			}
		})
	}
}

// TestOpModFloatFloat tests float % float modulo operations.
func TestOpModFloatFloat(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"basic float modulo 10.5 % 3.0", 10.5, 3.0, 1.5},
		{"float modulo 7.5 % 2.5", 7.5, 2.5, 0.0},
		{"float modulo with remainder", 10.0, 3.0, 1.0},
		{"negative float dividend", -10.5, 3.0, -1.5},
		{"negative float divisor", 10.5, -3.0, 1.5},
		{"both negative floats", -10.5, -3.0, -1.5},
		{"zero dividend float", 0.0, 3.14, 0.0},
		{"self modulo float", 3.14, 3.14, 0.0},
		{"small float values", 0.7, 0.3, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constants := []Value{FloatValue{Val: tt.a}, FloatValue{Val: tt.b}}
			bytecode := createBytecodeHeader(constants)

			operand0 := uint32(0)
			operand1 := uint32(1)
			bytecode = addInstruction(bytecode, OpPush, &operand0)
			bytecode = addInstruction(bytecode, OpPush, &operand1)
			bytecode = addInstruction(bytecode, OpMod, nil)
			bytecode = addInstruction(bytecode, OpHalt, nil)

			vm := NewVM()
			result, err := vm.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			floatVal, ok := result.(FloatValue)
			if !ok {
				t.Fatalf("Expected FloatValue, got %T", result)
			}
			if math.Abs(floatVal.Val-tt.expected) > 0.0001 {
				t.Errorf("Expected %f %% %f = %f, got %f", tt.a, tt.b, tt.expected, floatVal.Val)
			}
		})
	}
}

// TestOpModIntFloat tests int % float mixed-type modulo operations.
func TestOpModIntFloat(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        float64
		expected float64
	}{
		{"int % float basic", 10, 3.0, 1.0},
		{"int % float with remainder", 10, 3.5, 3.0},
		{"int % float exact", 9, 4.5, 0.0},
		{"negative int % float", -10, 3.0, -1.0},
		{"int % negative float", 10, -3.0, 1.0},
		{"zero int % float", 0, 3.5, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(IntValue{Val: tt.a})
			vm.Push(FloatValue{Val: tt.b})

			err := vm.execMod()
			if err != nil {
				t.Fatalf("execMod() error: %v", err)
			}

			result, err := vm.Pop()
			if err != nil {
				t.Fatalf("Pop() error: %v", err)
			}

			floatVal, ok := result.(FloatValue)
			if !ok {
				t.Fatalf("Expected FloatValue, got %T", result)
			}
			if math.Abs(floatVal.Val-tt.expected) > 0.0001 {
				t.Errorf("Expected %d %% %f = %f, got %f", tt.a, tt.b, tt.expected, floatVal.Val)
			}
		})
	}
}

// TestOpModFloatInt tests float % int mixed-type modulo operations.
func TestOpModFloatInt(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        int64
		expected float64
	}{
		{"float % int basic", 10.5, 3, 1.5},
		{"float % int exact", 9.0, 3, 0.0},
		{"negative float % int", -10.5, 3, -1.5},
		{"float % negative int", 10.5, -3, 1.5},
		{"zero float % int", 0.0, 5, 0.0},
		{"float % 1", 3.7, 1, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(FloatValue{Val: tt.a})
			vm.Push(IntValue{Val: tt.b})

			err := vm.execMod()
			if err != nil {
				t.Fatalf("execMod() error: %v", err)
			}

			result, err := vm.Pop()
			if err != nil {
				t.Fatalf("Pop() error: %v", err)
			}

			floatVal, ok := result.(FloatValue)
			if !ok {
				t.Fatalf("Expected FloatValue, got %T", result)
			}
			if math.Abs(floatVal.Val-tt.expected) > 0.0001 {
				t.Errorf("Expected %f %% %d = %f, got %f", tt.a, tt.b, tt.expected, floatVal.Val)
			}
		})
	}
}

// TestOpModByZero tests that modulo by zero returns appropriate errors.
func TestOpModByZero(t *testing.T) {
	tests := []struct {
		name string
		a    Value
		b    Value
	}{
		{"int % 0", IntValue{Val: 10}, IntValue{Val: 0}},
		{"int % 0.0", IntValue{Val: 10}, FloatValue{Val: 0.0}},
		{"float % 0.0", FloatValue{Val: 10.5}, FloatValue{Val: 0.0}},
		{"float % 0", FloatValue{Val: 10.5}, IntValue{Val: 0}},
		{"negative int % 0", IntValue{Val: -5}, IntValue{Val: 0}},
		{"zero % 0", IntValue{Val: 0}, IntValue{Val: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(tt.a)
			vm.Push(tt.b)

			err := vm.execMod()
			if err == nil {
				t.Error("Expected modulo by zero error, got nil")
				return
			}
			if !strings.Contains(err.Error(), "modulo by zero") {
				t.Errorf("Expected error containing 'modulo by zero', got '%s'", err.Error())
			}
		})
	}
}

// TestOpModByZeroBytecode tests modulo by zero through full bytecode execution.
func TestOpModByZeroBytecode(t *testing.T) {
	constants := []Value{IntValue{Val: 42}, IntValue{Val: 0}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	operand1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &operand0)
	bytecode = addInstruction(bytecode, OpPush, &operand1)
	bytecode = addInstruction(bytecode, OpMod, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Error("Expected modulo by zero error")
	}
	if !strings.Contains(err.Error(), "modulo by zero") {
		t.Errorf("Expected error containing 'modulo by zero', got '%s'", err.Error())
	}
}

// TestOpModTypeError tests that modulo with incompatible types returns errors.
func TestOpModTypeError(t *testing.T) {
	tests := []struct {
		name     string
		a        Value
		b        Value
		errorMsg string
	}{
		{"string % int", StringValue{Val: "hello"}, IntValue{Val: 3}, "cannot compute modulo"},
		{"int % string", IntValue{Val: 10}, StringValue{Val: "3"}, "cannot compute modulo"},
		{"bool % int", BoolValue{Val: true}, IntValue{Val: 2}, "cannot compute modulo"},
		{"int % bool", IntValue{Val: 10}, BoolValue{Val: true}, "cannot compute modulo"},
		{"string % string", StringValue{Val: "a"}, StringValue{Val: "b"}, "cannot compute modulo"},
		{"null % int", NullValue{}, IntValue{Val: 5}, "cannot compute modulo"},
		{"int % null", IntValue{Val: 5}, NullValue{}, "cannot compute modulo"},
		{"float % string", FloatValue{Val: 3.14}, StringValue{Val: "2"}, "cannot compute modulo"},
		{"string % float", StringValue{Val: "10"}, FloatValue{Val: 3.0}, "cannot compute modulo"},
		{"bool % float", BoolValue{Val: false}, FloatValue{Val: 1.0}, "cannot compute modulo"},
		{"float % bool", FloatValue{Val: 5.5}, BoolValue{Val: true}, "cannot compute modulo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(tt.a)
			vm.Push(tt.b)

			err := vm.execMod()
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				return
			}
			if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

// TestOpModStackUnderflow tests that modulo with insufficient stack values returns an error.
func TestOpModStackUnderflow(t *testing.T) {
	t.Run("empty stack", func(t *testing.T) {
		vm := NewVM()
		err := vm.execMod()
		if err == nil {
			t.Error("Expected stack underflow error")
		}
		if !strings.Contains(err.Error(), "underflow") {
			t.Errorf("Expected underflow error, got '%s'", err.Error())
		}
	})

	t.Run("single value on stack", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 10})
		err := vm.execMod()
		if err == nil {
			t.Error("Expected stack underflow error")
		}
		if !strings.Contains(err.Error(), "underflow") {
			t.Errorf("Expected underflow error, got '%s'", err.Error())
		}
	})
}

// TestOpModBytecodeExecution tests OpMod through full bytecode execution for
// various type combinations to ensure the opcode dispatch works correctly.
func TestOpModBytecodeExecution(t *testing.T) {
	t.Run("int mod via bytecode", func(t *testing.T) {
		constants := []Value{IntValue{Val: 17}, IntValue{Val: 5}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpMod, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		intVal, ok := result.(IntValue)
		if !ok {
			t.Fatalf("Expected IntValue, got %T", result)
		}
		if intVal.Val != 2 {
			t.Errorf("Expected 17 %% 5 = 2, got %d", intVal.Val)
		}
	})

	t.Run("float mod via bytecode", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 10.5}, FloatValue{Val: 3.0}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpMod, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		floatVal, ok := result.(FloatValue)
		if !ok {
			t.Fatalf("Expected FloatValue, got %T", result)
		}
		if math.Abs(floatVal.Val-1.5) > 0.0001 {
			t.Errorf("Expected 10.5 %% 3.0 = 1.5, got %f", floatVal.Val)
		}
	})

	t.Run("chained mod operations via bytecode", func(t *testing.T) {
		// Compute (100 % 7) % 3 = 2 % 3 = 2
		constants := []Value{IntValue{Val: 100}, IntValue{Val: 7}, IntValue{Val: 3}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		operand2 := uint32(2)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpMod, nil)
		bytecode = addInstruction(bytecode, OpPush, &operand2)
		bytecode = addInstruction(bytecode, OpMod, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		intVal, ok := result.(IntValue)
		if !ok {
			t.Fatalf("Expected IntValue, got %T", result)
		}
		if intVal.Val != 2 {
			t.Errorf("Expected (100 %% 7) %% 3 = 2, got %d", intVal.Val)
		}
	})
}
