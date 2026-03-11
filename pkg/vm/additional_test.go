package vm

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

// TestValueToString tests the valueToString function
func TestValueToString(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"string", StringValue{Val: "hello"}, "hello"},
		{"int_positive", IntValue{Val: 42}, "42"},
		{"int_zero", IntValue{Val: 0}, "0"},
		{"int_negative", IntValue{Val: -123}, "-123"},
		{"float", FloatValue{Val: 3.14}, "3.14"},
		{"bool_true", BoolValue{Val: true}, "true"},
		{"bool_false", BoolValue{Val: false}, "false"},
		{"null", NullValue{}, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToString(tt.value)
			if result != tt.expected {
				t.Errorf("valueToString(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

// TestValuesEqual tests the valuesEqual method
func TestValuesEqual(t *testing.T) {
	vm := NewVM()

	tests := []struct {
		name     string
		a        Value
		b        Value
		expected bool
	}{
		// Int comparisons
		{"int_equal", IntValue{Val: 42}, IntValue{Val: 42}, true},
		{"int_not_equal", IntValue{Val: 42}, IntValue{Val: 24}, false},
		{"int_vs_float", IntValue{Val: 42}, FloatValue{Val: 42.0}, false},

		// Float comparisons
		{"float_equal", FloatValue{Val: 3.14}, FloatValue{Val: 3.14}, true},
		{"float_not_equal", FloatValue{Val: 3.14}, FloatValue{Val: 2.71}, false},

		// Bool comparisons
		{"bool_equal_true", BoolValue{Val: true}, BoolValue{Val: true}, true},
		{"bool_equal_false", BoolValue{Val: false}, BoolValue{Val: false}, true},
		{"bool_not_equal", BoolValue{Val: true}, BoolValue{Val: false}, false},

		// String comparisons
		{"string_equal", StringValue{Val: "hello"}, StringValue{Val: "hello"}, true},
		{"string_not_equal", StringValue{Val: "hello"}, StringValue{Val: "world"}, false},

		// Null comparisons
		{"null_equal", NullValue{}, NullValue{}, true},
		{"null_vs_int", NullValue{}, IntValue{Val: 0}, false},

		// Mixed type comparisons
		{"int_vs_string", IntValue{Val: 42}, StringValue{Val: "42"}, false},
		{"bool_vs_int", BoolValue{Val: true}, IntValue{Val: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestFloatComparisons tests comparison operations with floats
func TestFloatComparisons(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		op       Opcode
		expected bool
	}{
		// Float vs Float
		{"float_lt_true", 1.5, 2.5, OpLt, true},
		{"float_lt_false", 2.5, 1.5, OpLt, false},
		{"float_le_true", 1.5, 2.5, OpLe, true},
		{"float_le_equal", 2.5, 2.5, OpLe, true},
		{"float_le_false", 2.5, 1.5, OpLe, false},
		{"float_gt_true", 2.5, 1.5, OpGt, true},
		{"float_gt_false", 1.5, 2.5, OpGt, false},
		{"float_ge_true", 2.5, 1.5, OpGe, true},
		{"float_ge_equal", 2.5, 2.5, OpGe, true},
		{"float_ge_false", 1.5, 2.5, OpGe, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constants := []Value{FloatValue{Val: tt.a}, FloatValue{Val: tt.b}}
			bytecode := createBytecodeHeader(constants)

			operand0 := uint32(0)
			operand1 := uint32(1)
			bytecode = addInstruction(bytecode, OpPush, &operand0)
			bytecode = addInstruction(bytecode, OpPush, &operand1)
			bytecode = addInstruction(bytecode, tt.op, nil)
			bytecode = addInstruction(bytecode, OpHalt, nil)

			vm := NewVM()
			result, err := vm.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			boolVal, ok := result.(BoolValue)
			if !ok || boolVal.Val != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIntFloatComparisons tests comparisons between int and float
func TestIntFloatComparisons(t *testing.T) {
	// Test int < float
	constants := []Value{IntValue{Val: 5}, FloatValue{Val: 5.5}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	operand1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &operand0)
	bytecode = addInstruction(bytecode, OpPush, &operand1)
	bytecode = addInstruction(bytecode, OpLt, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if boolVal, ok := result.(BoolValue); !ok || !boolVal.Val {
		t.Errorf("Expected true, got %v", result)
	}
}

// TestFloatIntComparisons tests comparisons between float and int
func TestFloatIntComparisons(t *testing.T) {
	// Test float > int
	constants := []Value{FloatValue{Val: 5.5}, IntValue{Val: 5}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	operand1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &operand0)
	bytecode = addInstruction(bytecode, OpPush, &operand1)
	bytecode = addInstruction(bytecode, OpGt, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if boolVal, ok := result.(BoolValue); !ok || !boolVal.Val {
		t.Errorf("Expected true, got %v", result)
	}
}

// TestStringComparisons tests string comparison operations
func TestStringComparisons(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		op       Opcode
		expected bool
	}{
		{"string_lt_true", "apple", "banana", OpLt, true},
		{"string_lt_false", "banana", "apple", OpLt, false},
		{"string_le_true", "apple", "banana", OpLe, true},
		{"string_le_equal", "apple", "apple", OpLe, true},
		{"string_gt_true", "banana", "apple", OpGt, true},
		{"string_gt_false", "apple", "banana", OpGt, false},
		{"string_ge_true", "banana", "apple", OpGe, true},
		{"string_ge_equal", "banana", "banana", OpGe, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constants := []Value{StringValue{Val: tt.a}, StringValue{Val: tt.b}}
			bytecode := createBytecodeHeader(constants)

			operand0 := uint32(0)
			operand1 := uint32(1)
			bytecode = addInstruction(bytecode, OpPush, &operand0)
			bytecode = addInstruction(bytecode, OpPush, &operand1)
			bytecode = addInstruction(bytecode, tt.op, nil)
			bytecode = addInstruction(bytecode, OpHalt, nil)

			vm := NewVM()
			result, err := vm.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			boolVal, ok := result.(BoolValue)
			if !ok || boolVal.Val != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestFloatArithmeticOperations tests float arithmetic including mixed types
func TestFloatArithmeticOperations(t *testing.T) {
	// Test int + float
	t.Run("int_plus_float", func(t *testing.T) {
		constants := []Value{IntValue{Val: 10}, FloatValue{Val: 5.5}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpAdd, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-15.5) > 0.01 {
			t.Errorf("Expected 15.5, got %v", result)
		}
	})

	// Test float + int
	t.Run("float_plus_int", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 5.5}, IntValue{Val: 10}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpAdd, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-15.5) > 0.01 {
			t.Errorf("Expected 15.5, got %v", result)
		}
	})

	// Test float - int
	t.Run("float_minus_int", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 15.5}, IntValue{Val: 10}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpSub, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-5.5) > 0.01 {
			t.Errorf("Expected 5.5, got %v", result)
		}
	})

	// Test int - float
	t.Run("int_minus_float", func(t *testing.T) {
		constants := []Value{IntValue{Val: 10}, FloatValue{Val: 2.5}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpSub, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-7.5) > 0.01 {
			t.Errorf("Expected 7.5, got %v", result)
		}
	})

	// Test int * float
	t.Run("int_mul_float", func(t *testing.T) {
		constants := []Value{IntValue{Val: 4}, FloatValue{Val: 2.5}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpMul, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-10.0) > 0.01 {
			t.Errorf("Expected 10.0, got %v", result)
		}
	})

	// Test float * int
	t.Run("float_mul_int", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 2.5}, IntValue{Val: 4}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpMul, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-10.0) > 0.01 {
			t.Errorf("Expected 10.0, got %v", result)
		}
	})
}

// TestFloatDivision tests float division operations
func TestFloatDivision(t *testing.T) {
	// Test float / float
	t.Run("float_div_float", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 10.0}, FloatValue{Val: 4.0}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpDiv, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-2.5) > 0.01 {
			t.Errorf("Expected 2.5, got %v", result)
		}
	})

	// Test int / float
	t.Run("int_div_float", func(t *testing.T) {
		constants := []Value{IntValue{Val: 10}, FloatValue{Val: 4.0}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpDiv, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-2.5) > 0.01 {
			t.Errorf("Expected 2.5, got %v", result)
		}
	})

	// Test float / int
	t.Run("float_div_int", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 10.0}, IntValue{Val: 4}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpDiv, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-2.5) > 0.01 {
			t.Errorf("Expected 2.5, got %v", result)
		}
	})

	// Test float division by zero
	t.Run("float_div_by_zero", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 10.0}, FloatValue{Val: 0.0}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpDiv, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected division by zero error")
		}
	})

	// Test int / float division by zero
	t.Run("int_div_float_zero", func(t *testing.T) {
		constants := []Value{IntValue{Val: 10}, FloatValue{Val: 0.0}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpDiv, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected division by zero error")
		}
	})

	// Test float / int division by zero
	t.Run("float_div_int_zero", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 10.0}, IntValue{Val: 0}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpDiv, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected division by zero error")
		}
	})
}

// TestArrayConcatenation tests array + array
func TestArrayConcatenation(t *testing.T) {
	vm := NewVM()

	// Set up arrays on the stack
	vm.Push(ArrayValue{Val: []Value{IntValue{Val: 1}, IntValue{Val: 2}}})
	vm.Push(ArrayValue{Val: []Value{IntValue{Val: 3}, IntValue{Val: 4}}})

	// Execute add
	err := vm.execAdd()
	if err != nil {
		t.Fatalf("execAdd() error: %v", err)
	}

	result, err := vm.Pop()
	if err != nil {
		t.Fatalf("Pop() error: %v", err)
	}

	arrVal, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}

	if len(arrVal.Val) != 4 {
		t.Errorf("Expected array length 4, got %d", len(arrVal.Val))
	}

	expected := []int64{1, 2, 3, 4}
	for i, exp := range expected {
		if intVal, ok := arrVal.Val[i].(IntValue); !ok || intVal.Val != exp {
			t.Errorf("Element %d: expected %d, got %v", i, exp, arrVal.Val[i])
		}
	}
}

// TestOpNegation tests unary negation with additional cases
func TestOpNegation(t *testing.T) {
	t.Run("negate_int", func(t *testing.T) {
		constants := []Value{IntValue{Val: 42}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpNeg, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if intVal, ok := result.(IntValue); !ok || intVal.Val != -42 {
			t.Errorf("Expected -42, got %v", result)
		}
	})

	t.Run("negate_float", func(t *testing.T) {
		constants := []Value{FloatValue{Val: 3.14}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpNeg, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val+3.14) > 0.01 {
			t.Errorf("Expected -3.14, got %v", result)
		}
	})

	t.Run("negate_error", func(t *testing.T) {
		constants := []Value{StringValue{Val: "hello"}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpNeg, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected error negating string")
		}
	})
}

// TestOpJumpUnconditional tests unconditional jump
func TestOpJumpUnconditional(t *testing.T) {
	constants := []Value{IntValue{Val: 1}, IntValue{Val: 2}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	operand1 := uint32(1)

	startOffset := len(bytecode)
	// Jump over the "Push 1" instruction to "Push 2"
	jumpTarget := uint32(startOffset + 5 + 5)

	bytecode = addInstruction(bytecode, OpJump, &jumpTarget) // Jump to Push 2
	bytecode = addInstruction(bytecode, OpPush, &operand0)   // Push 1 (skipped)
	bytecode = addInstruction(bytecode, OpPush, &operand1)   // Push 2
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if intVal, ok := result.(IntValue); !ok || intVal.Val != 2 {
		t.Errorf("Expected 2, got %v", result)
	}
}

// TestOpPop tests pop operation
func TestOpPop(t *testing.T) {
	constants := []Value{IntValue{Val: 1}, IntValue{Val: 2}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	operand1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &operand0) // Push 1
	bytecode = addInstruction(bytecode, OpPush, &operand1) // Push 2
	bytecode = addInstruction(bytecode, OpPop, nil)        // Pop 2
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Should have 1 left on stack (2 was popped)
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 1 {
		t.Errorf("Expected 1, got %v", result)
	}
}

// TestOpReturn tests the return opcode
func TestOpReturn(t *testing.T) {
	constants := []Value{IntValue{Val: 42}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &operand0) // Push 42
	bytecode = addInstruction(bytecode, OpReturn, nil)     // Return
	bytecode = addInstruction(bytecode, OpPush, &operand0) // This should not be executed
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if intVal, ok := result.(IntValue); !ok || intVal.Val != 42 {
		t.Errorf("Expected 42, got %v", result)
	}
}

// TestOpGetIndexExtended tests array indexing with additional cases
func TestOpGetIndexExtended(t *testing.T) {
	t.Run("valid_index", func(t *testing.T) {
		constants := []Value{IntValue{Val: 10}, IntValue{Val: 20}, IntValue{Val: 30}, IntValue{Val: 1}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		operand2 := uint32(2)
		operand3 := uint32(3)
		operandCount := uint32(3)

		// Build array [10, 20, 30]
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpPush, &operand2)
		bytecode = addInstruction(bytecode, OpBuildArray, &operandCount)

		// Get index 1
		bytecode = addInstruction(bytecode, OpPush, &operand3)
		bytecode = addInstruction(bytecode, OpGetIndex, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		result, err := vm.Execute(bytecode)
		if err != nil {
			t.Fatalf("Execute() error: %v", err)
		}

		if intVal, ok := result.(IntValue); !ok || intVal.Val != 20 {
			t.Errorf("Expected 20, got %v", result)
		}
	})

	t.Run("out_of_bounds", func(t *testing.T) {
		constants := []Value{IntValue{Val: 10}, IntValue{Val: 5}}
		bytecode := createBytecodeHeader(constants)

		operand0 := uint32(0)
		operand1 := uint32(1)
		operandCount := uint32(1)

		// Build array [10]
		bytecode = addInstruction(bytecode, OpPush, &operand0)
		bytecode = addInstruction(bytecode, OpBuildArray, &operandCount)

		// Get index 5 (out of bounds)
		bytecode = addInstruction(bytecode, OpPush, &operand1)
		bytecode = addInstruction(bytecode, OpGetIndex, nil)
		bytecode = addInstruction(bytecode, OpHalt, nil)

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected index out of bounds error")
		}
	})

	t.Run("negative_index", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ArrayValue{Val: []Value{IntValue{Val: 10}}})
		vm.Push(IntValue{Val: -1})

		err := vm.execGetIndex()
		if err == nil {
			t.Error("Expected error for negative index")
		}
	})

	t.Run("non_array", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "hello"})
		vm.Push(IntValue{Val: 0})

		err := vm.execGetIndex()
		if err == nil {
			t.Error("Expected error for non-array indexing")
		}
	})

	t.Run("non_int_index", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ArrayValue{Val: []Value{IntValue{Val: 10}}})
		vm.Push(StringValue{Val: "0"})

		err := vm.execGetIndex()
		if err == nil {
			t.Error("Expected error for non-integer index")
		}
	})
}

// TestOpCall tests function calls
func TestOpCall(t *testing.T) {
	t.Run("builtin_length", func(t *testing.T) {
		vm := NewVM()
		// Push function name, then argument
		vm.Push(StringValue{Val: "length"})
		vm.Push(ArrayValue{Val: []Value{IntValue{Val: 1}, IntValue{Val: 2}, IntValue{Val: 3}}})
		// Set up code buffer for reading argument count
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 1) // 1 argument

		err := vm.execCall()
		if err != nil {
			t.Fatalf("execCall() error: %v", err)
		}

		result, _ := vm.Pop()
		if intVal, ok := result.(IntValue); !ok || intVal.Val != 3 {
			t.Errorf("Expected 3, got %v", result)
		}
	})

	t.Run("undefined_function", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "undefined_func"})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0) // 0 arguments

		err := vm.execCall()
		if err == nil {
			t.Error("Expected undefined function error")
		}
	})
}

// TestIterators tests iterator operations
func TestIterators(t *testing.T) {
	t.Run("array_iterator", func(t *testing.T) {
		vm := NewVM()
		arr := ArrayValue{Val: []Value{IntValue{Val: 10}, IntValue{Val: 20}, IntValue{Val: 30}}}
		vm.Push(arr)

		// GetIter
		err := vm.execGetIter()
		if err != nil {
			t.Fatalf("execGetIter() error: %v", err)
		}

		iterIDVal, _ := vm.Pop()
		iterID := iterIDVal.(IntValue)

		// IterHasNext should be true
		vm.Push(iterIDVal)
		err = vm.execIterHasNext()
		if err != nil {
			t.Fatalf("execIterHasNext() error: %v", err)
		}

		hasNext, _ := vm.Pop()
		if !hasNext.(BoolValue).Val {
			t.Error("Expected hasNext to be true")
		}

		// IterNext (without key)
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0) // operand = 0 (no key)

		vm.Push(IntValue{Val: iterID.Val})
		err = vm.execIterNext()
		if err != nil {
			t.Fatalf("execIterNext() error: %v", err)
		}

		firstVal, _ := vm.Pop()
		if firstVal.(IntValue).Val != 10 {
			t.Errorf("Expected first value 10, got %v", firstVal)
		}
	})

	t.Run("object_iterator", func(t *testing.T) {
		vm := NewVM()
		obj := ObjectValue{Val: map[string]Value{
			"a": IntValue{Val: 1},
			"b": IntValue{Val: 2},
		}}
		vm.Push(obj)

		// GetIter
		err := vm.execGetIter()
		if err != nil {
			t.Fatalf("execGetIter() error: %v", err)
		}

		iterIDVal, _ := vm.Pop()

		// IterHasNext should be true
		vm.Push(iterIDVal)
		err = vm.execIterHasNext()
		if err != nil {
			t.Fatalf("execIterHasNext() error: %v", err)
		}

		hasNext, _ := vm.Pop()
		if !hasNext.(BoolValue).Val {
			t.Error("Expected hasNext to be true")
		}
	})

	t.Run("invalid_iterator_id", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 999})

		err := vm.execIterHasNext()
		if err == nil {
			t.Error("Expected error for invalid iterator ID")
		}
	})

	t.Run("iterator_on_non_collection", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 42})
		vm.iterators[0] = &Iterator{collection: IntValue{Val: 42}, index: 0}
		vm.Push(IntValue{Val: 0})

		err := vm.execIterHasNext()
		if err == nil {
			t.Error("Expected error for iterating over non-collection")
		}
	})
}

// TestTypeErrors tests type error cases
func TestTypeErrors(t *testing.T) {
	t.Run("add_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 42})
		vm.Push(BoolValue{Val: true})

		err := vm.execAdd()
		if err == nil {
			t.Error("Expected type error for int + bool")
		}
	})

	t.Run("sub_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "hello"})
		vm.Push(IntValue{Val: 5})

		err := vm.execSub()
		if err == nil {
			t.Error("Expected type error for string - int")
		}
	})

	t.Run("mul_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "hello"})
		vm.Push(IntValue{Val: 5})

		err := vm.execMul()
		if err == nil {
			t.Error("Expected type error for string * int")
		}
	})

	t.Run("div_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "hello"})
		vm.Push(IntValue{Val: 5})

		err := vm.execDiv()
		if err == nil {
			t.Error("Expected type error for string / int")
		}
	})

	t.Run("lt_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ArrayValue{Val: []Value{}})
		vm.Push(IntValue{Val: 5})

		err := vm.execLt()
		if err == nil {
			t.Error("Expected type error for array < int")
		}
	})

	t.Run("gt_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ArrayValue{Val: []Value{}})
		vm.Push(IntValue{Val: 5})

		err := vm.execGt()
		if err == nil {
			t.Error("Expected type error for array > int")
		}
	})

	t.Run("le_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ArrayValue{Val: []Value{}})
		vm.Push(IntValue{Val: 5})

		err := vm.execLe()
		if err == nil {
			t.Error("Expected type error for array <= int")
		}
	})

	t.Run("ge_incompatible", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ArrayValue{Val: []Value{}})
		vm.Push(IntValue{Val: 5})

		err := vm.execGe()
		if err == nil {
			t.Error("Expected type error for array >= int")
		}
	})

	t.Run("and_non_bool", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 1})
		vm.Push(IntValue{Val: 0})

		err := vm.execAnd()
		if err == nil {
			t.Error("Expected type error for int AND int")
		}
	})

	t.Run("or_non_bool", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 1})
		vm.Push(IntValue{Val: 0})

		err := vm.execOr()
		if err == nil {
			t.Error("Expected type error for int OR int")
		}
	})

	t.Run("not_non_bool", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 1})

		err := vm.execNot()
		if err == nil {
			t.Error("Expected type error for NOT int")
		}
	})

	t.Run("jump_if_false_non_bool", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 0})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execJumpIfFalse()
		if err == nil {
			t.Error("Expected type error for JumpIfFalse with non-bool")
		}
	})

	t.Run("jump_if_true_non_bool", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 1})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execJumpIfTrue()
		if err == nil {
			t.Error("Expected type error for JumpIfTrue with non-bool")
		}
	})
}

// TestWebSocketOperations tests WebSocket opcodes without handler
func TestWebSocketOperationsNoHandler(t *testing.T) {
	t.Run("ws_send_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "hello"})
		err := vm.execWsSend()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_broadcast_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "hello"})
		err := vm.execWsBroadcast()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_broadcast_room_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "room1"})
		vm.Push(StringValue{Val: "hello"})
		err := vm.execWsBroadcastRoom()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_join_room_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "room1"})
		err := vm.execWsJoinRoom()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_leave_room_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "room1"})
		err := vm.execWsLeaveRoom()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_close_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "goodbye"})
		err := vm.execWsClose()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_get_rooms_no_handler", func(t *testing.T) {
		vm := NewVM()
		err := vm.execWsGetRooms()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_get_clients_no_handler", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "room1"})
		err := vm.execWsGetClients()
		if err == nil {
			t.Error("Expected error when WebSocket handler not available")
		}
	})

	t.Run("ws_get_conn_count_no_handler", func(t *testing.T) {
		vm := NewVM()
		err := vm.execWsGetConnCount()
		if err != nil {
			t.Fatalf("Expected graceful degradation, got error: %v", err)
		}
		result, _ := vm.Pop()
		if result.(IntValue).Val != 0 {
			t.Errorf("Expected 0 for no handler, got %v", result)
		}
	})

	t.Run("ws_get_uptime_no_handler", func(t *testing.T) {
		vm := NewVM()
		err := vm.execWsGetUptime()
		if err != nil {
			t.Fatalf("Expected graceful degradation, got error: %v", err)
		}
		result, _ := vm.Pop()
		if result.(IntValue).Val != 0 {
			t.Errorf("Expected 0 for no handler, got %v", result)
		}
	})
}

// MockWebSocketHandler implements WebSocketHandler for testing
type MockWebSocketHandler struct {
	sentMessages      []interface{}
	broadcastMessages []interface{}
	roomMessages      map[string][]interface{}
	joinedRooms       []string
	leftRooms         []string
	closeReason       string
	closed            bool
	rooms             []string
	clients           map[string][]string
	connectionCount   int
	uptime            int64
}

func NewMockWebSocketHandler() *MockWebSocketHandler {
	return &MockWebSocketHandler{
		roomMessages:    make(map[string][]interface{}),
		clients:         make(map[string][]string),
		rooms:           []string{"room1", "room2"},
		connectionCount: 5,
		uptime:          1000,
	}
}

func (m *MockWebSocketHandler) Send(message interface{}) error {
	m.sentMessages = append(m.sentMessages, message)
	return nil
}

func (m *MockWebSocketHandler) Broadcast(message interface{}) error {
	m.broadcastMessages = append(m.broadcastMessages, message)
	return nil
}

func (m *MockWebSocketHandler) BroadcastToRoom(room string, message interface{}) error {
	m.roomMessages[room] = append(m.roomMessages[room], message)
	return nil
}

func (m *MockWebSocketHandler) JoinRoom(room string) error {
	m.joinedRooms = append(m.joinedRooms, room)
	return nil
}

func (m *MockWebSocketHandler) LeaveRoom(room string) error {
	m.leftRooms = append(m.leftRooms, room)
	return nil
}

func (m *MockWebSocketHandler) Close(reason string) error {
	m.closeReason = reason
	m.closed = true
	return nil
}

func (m *MockWebSocketHandler) GetRooms() []string {
	return m.rooms
}

func (m *MockWebSocketHandler) GetRoomClients(room string) []string {
	if clients, ok := m.clients[room]; ok {
		return clients
	}
	return []string{}
}

func (m *MockWebSocketHandler) GetConnectionID() string {
	return "mock-conn-id"
}

func (m *MockWebSocketHandler) GetConnectionCount() int {
	return m.connectionCount
}

func (m *MockWebSocketHandler) GetUptime() int64 {
	return m.uptime
}

// TestWebSocketOperationsWithHandler tests WebSocket opcodes with a mock handler
func TestWebSocketOperationsWithHandler(t *testing.T) {
	t.Run("ws_send", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "hello"})
		err := vm.execWsSend()
		if err != nil {
			t.Fatalf("execWsSend() error: %v", err)
		}

		if len(handler.sentMessages) != 1 || handler.sentMessages[0] != "hello" {
			t.Errorf("Expected sent message 'hello', got %v", handler.sentMessages)
		}
	})

	t.Run("ws_broadcast", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "broadcast msg"})
		err := vm.execWsBroadcast()
		if err != nil {
			t.Fatalf("execWsBroadcast() error: %v", err)
		}

		if len(handler.broadcastMessages) != 1 {
			t.Errorf("Expected 1 broadcast message, got %d", len(handler.broadcastMessages))
		}
	})

	t.Run("ws_broadcast_room", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "room1"})
		vm.Push(StringValue{Val: "room msg"})
		err := vm.execWsBroadcastRoom()
		if err != nil {
			t.Fatalf("execWsBroadcastRoom() error: %v", err)
		}

		if len(handler.roomMessages["room1"]) != 1 {
			t.Errorf("Expected 1 room message, got %d", len(handler.roomMessages["room1"]))
		}
	})

	t.Run("ws_broadcast_room_non_string", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(IntValue{Val: 123}) // Non-string room name
		vm.Push(StringValue{Val: "msg"})
		err := vm.execWsBroadcastRoom()
		if err == nil {
			t.Error("Expected error for non-string room name")
		}
	})

	t.Run("ws_join_room", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "newroom"})
		err := vm.execWsJoinRoom()
		if err != nil {
			t.Fatalf("execWsJoinRoom() error: %v", err)
		}

		if len(handler.joinedRooms) != 1 || handler.joinedRooms[0] != "newroom" {
			t.Errorf("Expected joined room 'newroom', got %v", handler.joinedRooms)
		}
	})

	t.Run("ws_join_room_non_string", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(IntValue{Val: 123})
		err := vm.execWsJoinRoom()
		if err == nil {
			t.Error("Expected error for non-string room name")
		}
	})

	t.Run("ws_leave_room", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "oldroom"})
		err := vm.execWsLeaveRoom()
		if err != nil {
			t.Fatalf("execWsLeaveRoom() error: %v", err)
		}

		if len(handler.leftRooms) != 1 || handler.leftRooms[0] != "oldroom" {
			t.Errorf("Expected left room 'oldroom', got %v", handler.leftRooms)
		}
	})

	t.Run("ws_leave_room_non_string", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(IntValue{Val: 123})
		err := vm.execWsLeaveRoom()
		if err == nil {
			t.Error("Expected error for non-string room name")
		}
	})

	t.Run("ws_close", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "goodbye"})
		err := vm.execWsClose()
		if err != nil {
			t.Fatalf("execWsClose() error: %v", err)
		}

		if !handler.closed || handler.closeReason != "goodbye" {
			t.Errorf("Expected closed with reason 'goodbye', got closed=%v, reason=%s",
				handler.closed, handler.closeReason)
		}
	})

	t.Run("ws_close_non_string", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(IntValue{Val: 123}) // Non-string but should still work with empty reason
		err := vm.execWsClose()
		if err != nil {
			t.Fatalf("execWsClose() error: %v", err)
		}
	})

	t.Run("ws_get_rooms", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		err := vm.execWsGetRooms()
		if err != nil {
			t.Fatalf("execWsGetRooms() error: %v", err)
		}

		result, _ := vm.Pop()
		arrVal, ok := result.(ArrayValue)
		if !ok {
			t.Fatalf("Expected ArrayValue, got %T", result)
		}

		if len(arrVal.Val) != 2 {
			t.Errorf("Expected 2 rooms, got %d", len(arrVal.Val))
		}
	})

	t.Run("ws_get_clients", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		handler.clients["room1"] = []string{"client1", "client2"}
		vm.SetWebSocketHandler(handler)

		vm.Push(StringValue{Val: "room1"})
		err := vm.execWsGetClients()
		if err != nil {
			t.Fatalf("execWsGetClients() error: %v", err)
		}

		result, _ := vm.Pop()
		arrVal, ok := result.(ArrayValue)
		if !ok {
			t.Fatalf("Expected ArrayValue, got %T", result)
		}

		if len(arrVal.Val) != 2 {
			t.Errorf("Expected 2 clients, got %d", len(arrVal.Val))
		}
	})

	t.Run("ws_get_clients_non_string", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		vm.Push(IntValue{Val: 123})
		err := vm.execWsGetClients()
		if err == nil {
			t.Error("Expected error for non-string room name")
		}
	})

	t.Run("ws_get_conn_count", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		err := vm.execWsGetConnCount()
		if err != nil {
			t.Fatalf("execWsGetConnCount() error: %v", err)
		}

		result, _ := vm.Pop()
		if result.(IntValue).Val != 5 {
			t.Errorf("Expected 5, got %v", result)
		}
	})

	t.Run("ws_get_uptime", func(t *testing.T) {
		vm := NewVM()
		handler := NewMockWebSocketHandler()
		vm.SetWebSocketHandler(handler)

		err := vm.execWsGetUptime()
		if err != nil {
			t.Fatalf("execWsGetUptime() error: %v", err)
		}

		result, _ := vm.Pop()
		if result.(IntValue).Val != 1000 {
			t.Errorf("Expected 1000, got %v", result)
		}
	})
}

// TestValueToInterface tests valueToInterface function
func TestValueToInterface(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected interface{}
	}{
		{"int", IntValue{Val: 42}, int64(42)},
		{"float", FloatValue{Val: 3.14}, float64(3.14)},
		{"string", StringValue{Val: "hello"}, "hello"},
		{"bool_true", BoolValue{Val: true}, true},
		{"bool_false", BoolValue{Val: false}, false},
		{"null", NullValue{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToInterface(tt.value)
			if result != tt.expected {
				t.Errorf("valueToInterface(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}

	// Test array
	t.Run("array", func(t *testing.T) {
		arr := ArrayValue{Val: []Value{IntValue{Val: 1}, IntValue{Val: 2}}}
		result := valueToInterface(arr).([]interface{})
		if len(result) != 2 || result[0] != int64(1) || result[1] != int64(2) {
			t.Errorf("Expected [1, 2], got %v", result)
		}
	})

	// Test object
	t.Run("object", func(t *testing.T) {
		obj := ObjectValue{Val: map[string]Value{"key": StringValue{Val: "value"}}}
		result := valueToInterface(obj).(map[string]interface{})
		if result["key"] != "value" {
			t.Errorf("Expected {key: value}, got %v", result)
		}
	})
}

// TestUnknownOpcode tests handling of unknown opcodes
func TestUnknownOpcode(t *testing.T) {
	constants := []Value{}
	bytecode := createBytecodeHeader(constants)

	// Add an unknown opcode
	bytecode = append(bytecode, 0xFE) // Unknown opcode
	bytecode = append(bytecode, byte(OpHalt))

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Error("Expected error for unknown opcode")
	}
}

// TestParseBytecodeErrors tests bytecode parsing error cases
func TestParseBytecodeErrors(t *testing.T) {
	t.Run("unsupported_version", func(t *testing.T) {
		bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // Magic
		bytecode = append(bytecode, 0x02, 0x00, 0x00, 0x00) // Version 2 (unsupported)

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected error for unsupported version")
		}
	})

	t.Run("truncated_constant", func(t *testing.T) {
		bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // Magic
		bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // Version 1
		bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // 1 constant
		bytecode = append(bytecode, 0x01)                   // Int type, but no value

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected error for truncated constant")
		}
	})

	t.Run("unknown_constant_type", func(t *testing.T) {
		bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // Magic
		bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // Version 1
		bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // 1 constant
		bytecode = append(bytecode, 0xFF)                   // Unknown constant type

		vm := NewVM()
		_, err := vm.Execute(bytecode)
		if err == nil {
			t.Error("Expected error for unknown constant type")
		}
	})
}

// TestBuiltinFunctionErrors tests builtin function error cases
func TestBuiltinFunctionErrors(t *testing.T) {
	vm := NewVM()

	t.Run("time.now_with_args", func(t *testing.T) {
		_, err := vm.builtins["time.now"]([]Value{IntValue{Val: 1}})
		if err == nil {
			t.Error("Expected error for time.now with arguments")
		}
	})

	t.Run("now_with_args", func(t *testing.T) {
		_, err := vm.builtins["now"]([]Value{IntValue{Val: 1}})
		if err == nil {
			t.Error("Expected error for now with arguments")
		}
	})

	t.Run("length_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["length"]([]Value{})
		if err == nil {
			t.Error("Expected error for length with no arguments")
		}
	})

	t.Run("length_wrong_type", func(t *testing.T) {
		_, err := vm.builtins["length"]([]Value{IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for length with int")
		}
	})

	t.Run("upper_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["upper"]([]Value{})
		if err == nil {
			t.Error("Expected error for upper with no arguments")
		}
	})

	t.Run("upper_wrong_type", func(t *testing.T) {
		_, err := vm.builtins["upper"]([]Value{IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for upper with int")
		}
	})

	t.Run("lower_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["lower"]([]Value{})
		if err == nil {
			t.Error("Expected error for lower with no arguments")
		}
	})

	t.Run("lower_wrong_type", func(t *testing.T) {
		_, err := vm.builtins["lower"]([]Value{IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for lower with int")
		}
	})

	t.Run("trim_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["trim"]([]Value{})
		if err == nil {
			t.Error("Expected error for trim with no arguments")
		}
	})

	t.Run("trim_wrong_type", func(t *testing.T) {
		_, err := vm.builtins["trim"]([]Value{IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for trim with int")
		}
	})

	t.Run("split_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["split"]([]Value{StringValue{Val: "a,b,c"}})
		if err == nil {
			t.Error("Expected error for split with 1 argument")
		}
	})

	t.Run("split_wrong_first_type", func(t *testing.T) {
		_, err := vm.builtins["split"]([]Value{IntValue{Val: 42}, StringValue{Val: ","}})
		if err == nil {
			t.Error("Expected error for split with int first arg")
		}
	})

	t.Run("split_wrong_second_type", func(t *testing.T) {
		_, err := vm.builtins["split"]([]Value{StringValue{Val: "a,b,c"}, IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for split with int second arg")
		}
	})

	t.Run("join_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["join"]([]Value{ArrayValue{}})
		if err == nil {
			t.Error("Expected error for join with 1 argument")
		}
	})

	t.Run("join_wrong_first_type", func(t *testing.T) {
		_, err := vm.builtins["join"]([]Value{StringValue{Val: "test"}, StringValue{Val: ","}})
		if err == nil {
			t.Error("Expected error for join with string first arg")
		}
	})

	t.Run("join_wrong_second_type", func(t *testing.T) {
		_, err := vm.builtins["join"]([]Value{ArrayValue{}, IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for join with int second arg")
		}
	})

	t.Run("contains_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["contains"]([]Value{StringValue{Val: "test"}})
		if err == nil {
			t.Error("Expected error for contains with 1 argument")
		}
	})

	t.Run("contains_wrong_first_type", func(t *testing.T) {
		_, err := vm.builtins["contains"]([]Value{IntValue{Val: 42}, StringValue{Val: "test"}})
		if err == nil {
			t.Error("Expected error for contains with int first arg")
		}
	})

	t.Run("contains_wrong_second_type", func(t *testing.T) {
		_, err := vm.builtins["contains"]([]Value{StringValue{Val: "test"}, IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for contains with int second arg")
		}
	})

	t.Run("replace_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["replace"]([]Value{StringValue{Val: "test"}, StringValue{Val: "t"}})
		if err == nil {
			t.Error("Expected error for replace with 2 arguments")
		}
	})

	t.Run("replace_wrong_first_type", func(t *testing.T) {
		_, err := vm.builtins["replace"]([]Value{IntValue{Val: 42}, StringValue{Val: "t"}, StringValue{Val: "x"}})
		if err == nil {
			t.Error("Expected error for replace with int first arg")
		}
	})

	t.Run("replace_wrong_second_type", func(t *testing.T) {
		_, err := vm.builtins["replace"]([]Value{StringValue{Val: "test"}, IntValue{Val: 42}, StringValue{Val: "x"}})
		if err == nil {
			t.Error("Expected error for replace with int second arg")
		}
	})

	t.Run("replace_wrong_third_type", func(t *testing.T) {
		_, err := vm.builtins["replace"]([]Value{StringValue{Val: "test"}, StringValue{Val: "t"}, IntValue{Val: 42}})
		if err == nil {
			t.Error("Expected error for replace with int third arg")
		}
	})

	t.Run("substring_wrong_args", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{StringValue{Val: "test"}, IntValue{Val: 0}})
		if err == nil {
			t.Error("Expected error for substring with 2 arguments")
		}
	})

	t.Run("substring_wrong_first_type", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{IntValue{Val: 42}, IntValue{Val: 0}, IntValue{Val: 2}})
		if err == nil {
			t.Error("Expected error for substring with int first arg")
		}
	})

	t.Run("substring_wrong_second_type", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{StringValue{Val: "test"}, StringValue{Val: "0"}, IntValue{Val: 2}})
		if err == nil {
			t.Error("Expected error for substring with string second arg")
		}
	})

	t.Run("substring_wrong_third_type", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{StringValue{Val: "test"}, IntValue{Val: 0}, StringValue{Val: "2"}})
		if err == nil {
			t.Error("Expected error for substring with string third arg")
		}
	})

	t.Run("substring_negative_start", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{StringValue{Val: "test"}, IntValue{Val: -1}, IntValue{Val: 2}})
		if err == nil {
			t.Error("Expected error for substring with negative start")
		}
	})

	t.Run("substring_negative_end", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{StringValue{Val: "test"}, IntValue{Val: 0}, IntValue{Val: -1}})
		if err == nil {
			t.Error("Expected error for substring with negative end")
		}
	})

	t.Run("substring_start_greater_than_end", func(t *testing.T) {
		_, err := vm.builtins["substring"]([]Value{StringValue{Val: "test"}, IntValue{Val: 3}, IntValue{Val: 1}})
		if err == nil {
			t.Error("Expected error for substring with start > end")
		}
	})
}

// TestStringBuiltins tests that VM string builtins use Go stdlib correctly
func TestStringBuiltins(t *testing.T) {
	t.Run("split_empty_delim", func(t *testing.T) {
		result := strings.Split("abc", "")
		if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
			t.Errorf("Expected [a, b, c], got %v", result)
		}
	})

	t.Run("join_empty", func(t *testing.T) {
		result := strings.Join([]string{}, ",")
		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})

	t.Run("join_single", func(t *testing.T) {
		result := strings.Join([]string{"hello"}, ",")
		if result != "hello" {
			t.Errorf("Expected 'hello', got %q", result)
		}
	})

	t.Run("unicode_toUpper", func(t *testing.T) {
		result := strings.ToUpper("café")
		if result != "CAFÉ" {
			t.Errorf("Expected 'CAFÉ', got %q", result)
		}
	})
}

// TestGetField tests field access on non-objects
func TestGetFieldErrors(t *testing.T) {
	t.Run("non_string_key", func(t *testing.T) {
		vm := NewVM()
		vm.Push(ObjectValue{Val: map[string]Value{"key": IntValue{Val: 42}}})
		vm.Push(IntValue{Val: 0})

		err := vm.execGetField()
		if err == nil {
			t.Error("Expected error for non-string field name")
		}
	})

	t.Run("non_object", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 42})
		vm.Push(StringValue{Val: "key"})

		err := vm.execGetField()
		if err == nil {
			t.Error("Expected error for getting field from non-object")
		}
	})
}

// TestBuildObjectErrors tests object building error cases
func TestBuildObjectErrors(t *testing.T) {
	t.Run("non_string_key", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 123}) // Non-string key
		vm.Push(StringValue{Val: "value"})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 1)

		err := vm.execBuildObject()
		if err == nil {
			t.Error("Expected error for non-string object key")
		}
	})
}

// TestStoreLoadVarErrors tests variable storage/loading error cases
func TestStoreLoadVarErrors(t *testing.T) {
	t.Run("store_non_string_name", func(t *testing.T) {
		vm := NewVM()
		vm.constants = []Value{IntValue{Val: 42}}
		vm.Push(StringValue{Val: "value"})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execStoreVar()
		if err == nil {
			t.Error("Expected error for non-string variable name")
		}
	})

	t.Run("load_non_string_name", func(t *testing.T) {
		vm := NewVM()
		vm.constants = []Value{IntValue{Val: 42}}
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execLoadVar()
		if err == nil {
			t.Error("Expected error for non-string variable name")
		}
	})

	t.Run("load_from_globals", func(t *testing.T) {
		vm := NewVM()
		vm.constants = []Value{StringValue{Val: "globalVar"}}
		vm.globals["globalVar"] = IntValue{Val: 100}
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execLoadVar()
		if err != nil {
			t.Fatalf("execLoadVar() error: %v", err)
		}

		result, _ := vm.Pop()
		if result.(IntValue).Val != 100 {
			t.Errorf("Expected 100, got %v", result)
		}
	})
}

// TestExecCallErrors tests function call error cases
func TestExecCallErrors(t *testing.T) {
	t.Run("non_string_function_name", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 42}) // Non-string function name
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0) // 0 args

		err := vm.execCall()
		if err == nil {
			t.Error("Expected error for non-string function name")
		}
	})
}

// TestIterNextErrors tests iterator next errors
func TestIterNextErrors(t *testing.T) {
	t.Run("non_int_iterator_id", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "not an int"})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execIterNext()
		if err == nil {
			t.Error("Expected error for non-int iterator ID")
		}
	})

	t.Run("invalid_iterator_id", func(t *testing.T) {
		vm := NewVM()
		vm.Push(IntValue{Val: 999})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execIterNext()
		if err == nil {
			t.Error("Expected error for invalid iterator ID")
		}
	})

	t.Run("exhausted_array_iterator", func(t *testing.T) {
		vm := NewVM()
		vm.iterators[0] = &Iterator{
			collection: ArrayValue{Val: []Value{}},
			index:      0,
		}
		vm.Push(IntValue{Val: 0})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execIterNext()
		if err == nil {
			t.Error("Expected error for exhausted iterator")
		}
	})

	t.Run("exhausted_object_iterator", func(t *testing.T) {
		vm := NewVM()
		vm.iterators[0] = &Iterator{
			collection: ObjectValue{Val: map[string]Value{}},
			index:      0,
			keys:       []string{},
		}
		vm.Push(IntValue{Val: 0})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execIterNext()
		if err == nil {
			t.Error("Expected error for exhausted iterator")
		}
	})

	t.Run("iterate_non_collection", func(t *testing.T) {
		vm := NewVM()
		vm.iterators[0] = &Iterator{
			collection: IntValue{Val: 42},
			index:      0,
		}
		vm.Push(IntValue{Val: 0})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 0)

		err := vm.execIterNext()
		if err == nil {
			t.Error("Expected error for iterating non-collection")
		}
	})
}

// TestIterWithKey tests iteration with key
func TestIterWithKey(t *testing.T) {
	t.Run("array_with_key", func(t *testing.T) {
		vm := NewVM()
		vm.iterators[0] = &Iterator{
			collection: ArrayValue{Val: []Value{StringValue{Val: "first"}, StringValue{Val: "second"}}},
			index:      0,
		}
		vm.Push(IntValue{Val: 0})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 1) // hasKey = true

		err := vm.execIterNext()
		if err != nil {
			t.Fatalf("execIterNext() error: %v", err)
		}

		// Pop value
		val, _ := vm.Pop()
		if val.(StringValue).Val != "first" {
			t.Errorf("Expected 'first', got %v", val)
		}

		// Pop key (index)
		key, _ := vm.Pop()
		if key.(IntValue).Val != 0 {
			t.Errorf("Expected index 0, got %v", key)
		}
	})

	t.Run("object_with_key", func(t *testing.T) {
		vm := NewVM()
		vm.iterators[0] = &Iterator{
			collection: ObjectValue{Val: map[string]Value{"name": StringValue{Val: "Alice"}}},
			index:      0,
			keys:       []string{"name"},
		}
		vm.Push(IntValue{Val: 0})
		vm.pc = 0
		vm.code = make([]byte, 4)
		binary.LittleEndian.PutUint32(vm.code, 1) // hasKey = true

		err := vm.execIterNext()
		if err != nil {
			t.Fatalf("execIterNext() error: %v", err)
		}

		// Pop value
		val, _ := vm.Pop()
		if val.(StringValue).Val != "Alice" {
			t.Errorf("Expected 'Alice', got %v", val)
		}

		// Pop key
		key, _ := vm.Pop()
		if key.(StringValue).Val != "name" {
			t.Errorf("Expected 'name', got %v", key)
		}
	})
}

// TestIterHasNextErrors tests iterator hasNext errors
func TestIterHasNextErrors(t *testing.T) {
	t.Run("non_int_iterator_id", func(t *testing.T) {
		vm := NewVM()
		vm.Push(StringValue{Val: "not an int"})

		err := vm.execIterHasNext()
		if err == nil {
			t.Error("Expected error for non-int iterator ID")
		}
	})
}

// TestSetLocal tests the SetLocal method
func TestSetLocal(t *testing.T) {
	vm := NewVM()
	vm.SetLocal("test", IntValue{Val: 42})

	if val, ok := vm.locals["test"]; !ok || val.(IntValue).Val != 42 {
		t.Errorf("Expected local 'test' to be 42, got %v", vm.locals["test"])
	}
}

// TestEmptyStackResult tests execution that ends with empty stack
func TestEmptyStackResult(t *testing.T) {
	constants := []Value{IntValue{Val: 42}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &operand0)
	bytecode = addInstruction(bytecode, OpPop, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Should return NullValue when stack is empty
	if _, ok := result.(NullValue); !ok {
		t.Errorf("Expected NullValue, got %T", result)
	}
}

// TestValueTypes tests the Type() method of each value type
func TestValueTypes(t *testing.T) {
	tests := []struct {
		value    Value
		expected string
	}{
		{NullValue{}, "null"},
		{IntValue{Val: 42}, "int"},
		{FloatValue{Val: 3.14}, "float"},
		{StringValue{Val: "hello"}, "string"},
		{BoolValue{Val: true}, "bool"},
		{ArrayValue{Val: []Value{}}, "array"},
		{ObjectValue{Val: map[string]Value{}}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.value.Type() != tt.expected {
				t.Errorf("Expected type %q, got %q", tt.expected, tt.value.Type())
			}
		})
	}
}

// TestIntToString tests intToString helper
func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{-42, "-42"},
		{1234567890, "1234567890"},
		{-1234567890, "-1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := intToString(tt.input)
			if result != tt.expected {
				t.Errorf("intToString(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
