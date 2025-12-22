package vm

import (
	"math"
	"strings"
	"testing"
)

// Tests for OpNeg (unary negation)
func TestOpNeg(t *testing.T) {
	tests := []struct {
		name        string
		input       Value
		expected    Value
		expectError bool
		errorMsg    string
	}{
		{
			name:     "negate positive integer",
			input:    IntValue{Val: 42},
			expected: IntValue{Val: -42},
		},
		{
			name:     "negate negative integer",
			input:    IntValue{Val: -42},
			expected: IntValue{Val: 42},
		},
		{
			name:     "negate zero integer",
			input:    IntValue{Val: 0},
			expected: IntValue{Val: 0},
		},
		{
			name:     "negate positive float",
			input:    FloatValue{Val: 3.14},
			expected: FloatValue{Val: -3.14},
		},
		{
			name:     "negate negative float",
			input:    FloatValue{Val: -2.5},
			expected: FloatValue{Val: 2.5},
		},
		{
			name:     "negate zero float",
			input:    FloatValue{Val: 0.0},
			expected: FloatValue{Val: 0.0},
		},
		{
			name:        "negate string error",
			input:       StringValue{Val: "hello"},
			expectError: true,
			errorMsg:    "type error",
		},
		{
			name:        "negate boolean error",
			input:       BoolValue{Val: true},
			expectError: true,
			errorMsg:    "type error",
		},
		{
			name:        "negate null error",
			input:       NullValue{},
			expectError: true,
			errorMsg:    "type error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constants := []Value{tt.input}
			bytecode := createBytecodeHeader(constants)

			operand0 := uint32(0)
			bytecode = addInstruction(bytecode, OpPush, &operand0)
			bytecode = addInstruction(bytecode, OpNeg, nil)
			bytecode = addInstruction(bytecode, OpHalt, nil)

			vm := NewVM()
			result, err := vm.Execute(bytecode)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			switch expected := tt.expected.(type) {
			case IntValue:
				if intVal, ok := result.(IntValue); !ok || intVal.Val != expected.Val {
					t.Errorf("Expected IntValue{%d}, got %v", expected.Val, result)
				}
			case FloatValue:
				if floatVal, ok := result.(FloatValue); !ok || math.Abs(floatVal.Val-expected.Val) > 0.0001 {
					t.Errorf("Expected FloatValue{%f}, got %v", expected.Val, result)
				}
			}
		})
	}
}

// Tests for OpJump (unconditional jump)
func TestOpJump(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() []byte
		expected int64
	}{
		{
			name: "forward jump skips instruction",
			setup: func() []byte {
				constants := []Value{IntValue{Val: 1}, IntValue{Val: 2}}
				bytecode := createBytecodeHeader(constants)

				operand0 := uint32(0)
				operand1 := uint32(1)

				startOffset := len(bytecode)
				// Jump target calculation:
				// OpJump (1) + operand (4) = 5 bytes
				// OpPush (1) + operand (4) = 5 bytes (skipped)
				// So jump to startOffset + 5 + 5 = startOffset + 10
				jumpTarget := uint32(startOffset + 10)
				bytecode = addInstruction(bytecode, OpJump, &jumpTarget) // Jump over Push 1
				bytecode = addInstruction(bytecode, OpPush, &operand0)   // Push 1 (skipped)
				// Jump lands here
				bytecode = addInstruction(bytecode, OpPush, &operand1) // Push 2
				bytecode = addInstruction(bytecode, OpHalt, nil)

				return bytecode
			},
			expected: 2,
		},
		{
			name: "backward jump creates loop",
			setup: func() []byte {
				// Simpler test: just verify backward jump works
				// Push 10, then push 5, then subtract
				constants := []Value{IntValue{Val: 10}, IntValue{Val: 5}, BoolValue{Val: false}}
				bytecode := createBytecodeHeader(constants)

				operand0 := uint32(0)
				operand1 := uint32(1)
				operand2 := uint32(2)

				startOffset := len(bytecode)
				// Push 10
				bytecode = addInstruction(bytecode, OpPush, &operand0) // offset: startOffset
				// Push 5
				bytecode = addInstruction(bytecode, OpPush, &operand1) // offset: startOffset + 5
				// Subtract: 10 - 5 = 5
				bytecode = addInstruction(bytecode, OpSub, nil) // offset: startOffset + 10
				// Push false to skip the backward jump
				bytecode = addInstruction(bytecode, OpPush, &operand2) // offset: startOffset + 11
				// Jump back if true (won't jump since false)
				backJumpTarget := uint32(startOffset)
				bytecode = addInstruction(bytecode, OpJumpIfTrue, &backJumpTarget) // offset: startOffset + 16
				bytecode = addInstruction(bytecode, OpHalt, nil)

				return bytecode
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := tt.setup()

			vm := NewVM()
			result, err := vm.Execute(bytecode)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			if intVal, ok := result.(IntValue); !ok || intVal.Val != tt.expected {
				t.Errorf("Expected IntValue{%d}, got %v", tt.expected, result)
			}
		})
	}
}

// Tests for Iterator operations (OpGetIter, OpIterNext, OpIterHasNext)
func TestIteratorOperationsArray(t *testing.T) {
	tests := []struct {
		name          string
		arrayElements []Value
		expectSum     int64
	}{
		{
			name:          "iterate over array of integers",
			arrayElements: []Value{IntValue{Val: 1}, IntValue{Val: 2}, IntValue{Val: 3}},
			expectSum:     6, // 1 + 2 + 3
		},
		{
			name:          "iterate over empty array",
			arrayElements: []Value{},
			expectSum:     0,
		},
		{
			name:          "iterate over single element array",
			arrayElements: []Value{IntValue{Val: 42}},
			expectSum:     42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the array using VM operations and iterate
			vm := NewVM()

			// Manually set up and test iterator
			arr := ArrayValue{Val: tt.arrayElements}
			vm.Push(arr)

			// Get iterator
			err := vm.execGetIter()
			if err != nil {
				t.Fatalf("execGetIter() error: %v", err)
			}

			iterID, err := vm.Pop()
			if err != nil {
				t.Fatalf("Pop() error: %v", err)
			}

			sum := int64(0)

			// Iterate manually
			for {
				vm.Push(iterID)
				err := vm.execIterHasNext()
				if err != nil {
					t.Fatalf("execIterHasNext() error: %v", err)
				}

				hasNext, _ := vm.Pop()
				if !hasNext.(BoolValue).Val {
					break
				}

				vm.Push(iterID)
				// Set up operand for iterNext (0 = no key)
				vm.code = []byte{0x00, 0x00, 0x00, 0x00}
				vm.pc = 0
				err = vm.execIterNext()
				if err != nil {
					t.Fatalf("execIterNext() error: %v", err)
				}

				val, _ := vm.Pop()
				if intVal, ok := val.(IntValue); ok {
					sum += intVal.Val
				}
			}

			if sum != tt.expectSum {
				t.Errorf("Expected sum %d, got %d", tt.expectSum, sum)
			}
		})
	}
}

func TestIteratorOperationsObject(t *testing.T) {
	vm := NewVM()

	// Create an object
	obj := ObjectValue{Val: map[string]Value{
		"a": IntValue{Val: 10},
		"b": IntValue{Val: 20},
		"c": IntValue{Val: 30},
	}}
	vm.Push(obj)

	// Get iterator
	err := vm.execGetIter()
	if err != nil {
		t.Fatalf("execGetIter() error: %v", err)
	}

	iterID, err := vm.Pop()
	if err != nil {
		t.Fatalf("Pop() error: %v", err)
	}

	sum := int64(0)
	count := 0

	// Iterate
	for {
		vm.Push(iterID)
		err := vm.execIterHasNext()
		if err != nil {
			t.Fatalf("execIterHasNext() error: %v", err)
		}

		hasNext, _ := vm.Pop()
		if !hasNext.(BoolValue).Val {
			break
		}

		vm.Push(iterID)
		// Set up operand for iterNext (0 = no key)
		vm.code = []byte{0x00, 0x00, 0x00, 0x00}
		vm.pc = 0
		err = vm.execIterNext()
		if err != nil {
			t.Fatalf("execIterNext() error: %v", err)
		}

		val, _ := vm.Pop()
		if intVal, ok := val.(IntValue); ok {
			sum += intVal.Val
			count++
		}
	}

	if sum != 60 {
		t.Errorf("Expected sum 60, got %d", sum)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestIteratorWithKey(t *testing.T) {
	vm := NewVM()

	// Create an array
	arr := ArrayValue{Val: []Value{IntValue{Val: 100}, IntValue{Val: 200}}}
	vm.Push(arr)

	// Get iterator
	err := vm.execGetIter()
	if err != nil {
		t.Fatalf("execGetIter() error: %v", err)
	}

	iterID, err := vm.Pop()
	if err != nil {
		t.Fatalf("Pop() error: %v", err)
	}

	// Get first element with key
	vm.Push(iterID)
	// Set up operand for iterNext (1 = with key)
	vm.code = []byte{0x01, 0x00, 0x00, 0x00}
	vm.pc = 0
	err = vm.execIterNext()
	if err != nil {
		t.Fatalf("execIterNext() error: %v", err)
	}

	// Value should be 100, key should be 0
	val, _ := vm.Pop()
	key, _ := vm.Pop()

	if intVal, ok := val.(IntValue); !ok || intVal.Val != 100 {
		t.Errorf("Expected value 100, got %v", val)
	}
	if intKey, ok := key.(IntValue); !ok || intKey.Val != 0 {
		t.Errorf("Expected key 0, got %v", key)
	}
}

func TestIteratorInvalidType(t *testing.T) {
	vm := NewVM()

	// Try to iterate over a string (not supported)
	vm.Push(StringValue{Val: "hello"})

	err := vm.execGetIter()
	if err != nil {
		t.Fatalf("execGetIter() unexpected error: %v", err)
	}

	iterID, _ := vm.Pop()
	vm.Push(iterID)

	err = vm.execIterHasNext()
	if err == nil {
		t.Error("Expected error when iterating over string")
	}
}

// Tests for OpGetIndex
func TestOpGetIndex(t *testing.T) {
	tests := []struct {
		name        string
		array       ArrayValue
		index       Value
		expected    Value
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid index first element",
			array:    ArrayValue{Val: []Value{IntValue{Val: 10}, IntValue{Val: 20}, IntValue{Val: 30}}},
			index:    IntValue{Val: 0},
			expected: IntValue{Val: 10},
		},
		{
			name:     "valid index middle element",
			array:    ArrayValue{Val: []Value{IntValue{Val: 10}, IntValue{Val: 20}, IntValue{Val: 30}}},
			index:    IntValue{Val: 1},
			expected: IntValue{Val: 20},
		},
		{
			name:     "valid index last element",
			array:    ArrayValue{Val: []Value{IntValue{Val: 10}, IntValue{Val: 20}, IntValue{Val: 30}}},
			index:    IntValue{Val: 2},
			expected: IntValue{Val: 30},
		},
		{
			name:        "out of bounds positive index",
			array:       ArrayValue{Val: []Value{IntValue{Val: 10}}},
			index:       IntValue{Val: 5},
			expectError: true,
			errorMsg:    "index out of bounds",
		},
		{
			name:        "negative index error",
			array:       ArrayValue{Val: []Value{IntValue{Val: 10}}},
			index:       IntValue{Val: -1},
			expectError: true,
			errorMsg:    "index out of bounds",
		},
		{
			name:        "string index error",
			array:       ArrayValue{Val: []Value{IntValue{Val: 10}}},
			index:       StringValue{Val: "0"},
			expectError: true,
			errorMsg:    "array index must be an integer",
		},
		{
			name:        "float index error",
			array:       ArrayValue{Val: []Value{IntValue{Val: 10}}},
			index:       FloatValue{Val: 0.0},
			expectError: true,
			errorMsg:    "array index must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(tt.array)
			vm.Push(tt.index)

			err := vm.execGetIndex()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("execGetIndex() error: %v", err)
			}

			result, _ := vm.Pop()
			if intVal, ok := result.(IntValue); !ok || intVal.Val != tt.expected.(IntValue).Val {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestOpGetIndexNonArray(t *testing.T) {
	vm := NewVM()
	vm.Push(StringValue{Val: "hello"})
	vm.Push(IntValue{Val: 0})

	err := vm.execGetIndex()
	if err == nil {
		t.Error("Expected error when indexing non-array")
	}
	if !strings.Contains(err.Error(), "can only index arrays") {
		t.Errorf("Expected 'can only index arrays' error, got '%s'", err.Error())
	}
}

// Tests for type compatibility errors
func TestTypeCompatibilityErrors(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(vm *VM)
		execFunc func(vm *VM) error
		errorMsg string
	}{
		{
			name: "add string and integer",
			setup: func(vm *VM) {
				vm.Push(StringValue{Val: "hello"})
				vm.Push(IntValue{Val: 42})
			},
			execFunc: func(vm *VM) error { return vm.execAdd() },
			errorMsg: "cannot add",
		},
		{
			name: "subtract string from string",
			setup: func(vm *VM) {
				vm.Push(StringValue{Val: "hello"})
				vm.Push(StringValue{Val: "world"})
			},
			execFunc: func(vm *VM) error { return vm.execSub() },
			errorMsg: "cannot subtract",
		},
		{
			name: "multiply boolean and integer",
			setup: func(vm *VM) {
				vm.Push(BoolValue{Val: true})
				vm.Push(IntValue{Val: 5})
			},
			execFunc: func(vm *VM) error { return vm.execMul() },
			errorMsg: "cannot multiply",
		},
		{
			name: "divide string by integer",
			setup: func(vm *VM) {
				vm.Push(StringValue{Val: "hello"})
				vm.Push(IntValue{Val: 2})
			},
			execFunc: func(vm *VM) error { return vm.execDiv() },
			errorMsg: "cannot divide",
		},
		{
			name: "compare integer and boolean with less than",
			setup: func(vm *VM) {
				vm.Push(IntValue{Val: 5})
				vm.Push(BoolValue{Val: true})
			},
			execFunc: func(vm *VM) error { return vm.execLt() },
			errorMsg: "cannot compare",
		},
		{
			name: "compare integer and boolean with greater than",
			setup: func(vm *VM) {
				vm.Push(IntValue{Val: 5})
				vm.Push(BoolValue{Val: true})
			},
			execFunc: func(vm *VM) error { return vm.execGt() },
			errorMsg: "cannot compare",
		},
		{
			name: "logical AND with integers",
			setup: func(vm *VM) {
				vm.Push(IntValue{Val: 1})
				vm.Push(IntValue{Val: 1})
			},
			execFunc: func(vm *VM) error { return vm.execAnd() },
			errorMsg: "requires boolean",
		},
		{
			name: "logical OR with strings",
			setup: func(vm *VM) {
				vm.Push(StringValue{Val: "true"})
				vm.Push(StringValue{Val: "false"})
			},
			execFunc: func(vm *VM) error { return vm.execOr() },
			errorMsg: "requires boolean",
		},
		{
			name: "logical NOT with integer",
			setup: func(vm *VM) {
				vm.Push(IntValue{Val: 0})
			},
			execFunc: func(vm *VM) error { return vm.execNot() },
			errorMsg: "requires boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			tt.setup(vm)

			err := tt.execFunc(vm)
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
			} else if !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

// Additional edge case tests
func TestOpNegStackUnderflow(t *testing.T) {
	vm := NewVM()
	// Empty stack
	err := vm.execNeg()
	if err == nil {
		t.Error("Expected stack underflow error")
	}
	if !strings.Contains(err.Error(), "underflow") {
		t.Errorf("Expected underflow error, got '%s'", err.Error())
	}
}

func TestOpGetIndexEmptyArray(t *testing.T) {
	vm := NewVM()
	vm.Push(ArrayValue{Val: []Value{}})
	vm.Push(IntValue{Val: 0})

	err := vm.execGetIndex()
	if err == nil {
		t.Error("Expected error for empty array index")
	}
	if !strings.Contains(err.Error(), "index out of bounds") {
		t.Errorf("Expected 'index out of bounds' error, got '%s'", err.Error())
	}
}

func TestIteratorExhausted(t *testing.T) {
	vm := NewVM()

	// Create a single element array
	arr := ArrayValue{Val: []Value{IntValue{Val: 1}}}
	vm.Push(arr)

	err := vm.execGetIter()
	if err != nil {
		t.Fatalf("execGetIter() error: %v", err)
	}

	iterID, _ := vm.Pop()

	// Get the first element
	vm.Push(iterID)
	vm.code = []byte{0x00, 0x00, 0x00, 0x00}
	vm.pc = 0
	err = vm.execIterNext()
	if err != nil {
		t.Fatalf("execIterNext() error: %v", err)
	}
	vm.Pop() // discard the value

	// Try to get another element (should fail)
	vm.Push(iterID)
	vm.code = []byte{0x00, 0x00, 0x00, 0x00}
	vm.pc = 0
	err = vm.execIterNext()
	if err == nil {
		t.Error("Expected error when iterator is exhausted")
	}
	if !strings.Contains(err.Error(), "exhausted") {
		t.Errorf("Expected 'exhausted' error, got '%s'", err.Error())
	}
}

func TestInvalidIteratorID(t *testing.T) {
	vm := NewVM()

	// Push a non-existent iterator ID
	vm.Push(IntValue{Val: 999})

	err := vm.execIterHasNext()
	if err == nil {
		t.Error("Expected error for invalid iterator ID")
	}
	if !strings.Contains(err.Error(), "invalid iterator ID") {
		t.Errorf("Expected 'invalid iterator ID' error, got '%s'", err.Error())
	}
}

func TestIteratorIDNotInteger(t *testing.T) {
	vm := NewVM()

	// Push a string instead of iterator ID
	vm.Push(StringValue{Val: "not an iterator"})

	err := vm.execIterHasNext()
	if err == nil {
		t.Error("Expected error for non-integer iterator ID")
	}
	if !strings.Contains(err.Error(), "must be an integer") {
		t.Errorf("Expected 'must be an integer' error, got '%s'", err.Error())
	}
}

// Test OpJump with bytecode execution
func TestOpJumpBytecodeExecution(t *testing.T) {
	constants := []Value{IntValue{Val: 100}, IntValue{Val: 200}}
	bytecode := createBytecodeHeader(constants)

	operand0 := uint32(0)
	operand1 := uint32(1)

	startOffset := len(bytecode)
	// Calculate jump target: skip the Push 100 instruction (5 bytes for OpJump + 5 bytes for OpPush)
	jumpTarget := uint32(startOffset + 10)

	bytecode = addInstruction(bytecode, OpJump, &jumpTarget) // Jump over next instruction
	bytecode = addInstruction(bytecode, OpPush, &operand0)   // Push 100 (skipped)
	bytecode = addInstruction(bytecode, OpPush, &operand1)   // Push 200
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if intVal, ok := result.(IntValue); !ok || intVal.Val != 200 {
		t.Errorf("Expected IntValue{200}, got %v", result)
	}
}

// Test mixed type arithmetic with floats
func TestMixedTypeArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		a        Value
		b        Value
		op       func(vm *VM) error
		expected float64
	}{
		{
			name:     "int + float",
			a:        IntValue{Val: 10},
			b:        FloatValue{Val: 2.5},
			op:       func(vm *VM) error { return vm.execAdd() },
			expected: 12.5,
		},
		{
			name:     "float - int",
			a:        FloatValue{Val: 10.5},
			b:        IntValue{Val: 3},
			op:       func(vm *VM) error { return vm.execSub() },
			expected: 7.5,
		},
		{
			name:     "int * float",
			a:        IntValue{Val: 4},
			b:        FloatValue{Val: 2.5},
			op:       func(vm *VM) error { return vm.execMul() },
			expected: 10.0,
		},
		{
			name:     "float / int",
			a:        FloatValue{Val: 15.0},
			b:        IntValue{Val: 3},
			op:       func(vm *VM) error { return vm.execDiv() },
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(tt.a)
			vm.Push(tt.b)

			err := tt.op(vm)
			if err != nil {
				t.Fatalf("Operation error: %v", err)
			}

			result, _ := vm.Pop()
			floatVal, ok := result.(FloatValue)
			if !ok {
				t.Fatalf("Expected FloatValue, got %T", result)
			}
			if math.Abs(floatVal.Val-tt.expected) > 0.0001 {
				t.Errorf("Expected %f, got %f", tt.expected, floatVal.Val)
			}
		})
	}
}

// Test comparison operators with mixed types
func TestMixedTypeComparisons(t *testing.T) {
	tests := []struct {
		name     string
		a        Value
		b        Value
		op       func(vm *VM) error
		expected bool
	}{
		{
			name:     "int < float (true)",
			a:        IntValue{Val: 5},
			b:        FloatValue{Val: 5.5},
			op:       func(vm *VM) error { return vm.execLt() },
			expected: true,
		},
		{
			name:     "float > int (true)",
			a:        FloatValue{Val: 10.5},
			b:        IntValue{Val: 10},
			op:       func(vm *VM) error { return vm.execGt() },
			expected: true,
		},
		{
			name:     "int >= float (equal values)",
			a:        IntValue{Val: 5},
			b:        FloatValue{Val: 5.0},
			op:       func(vm *VM) error { return vm.execGe() },
			expected: true,
		},
		{
			name:     "float <= int (equal values)",
			a:        FloatValue{Val: 3.0},
			b:        IntValue{Val: 3},
			op:       func(vm *VM) error { return vm.execLe() },
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.Push(tt.a)
			vm.Push(tt.b)

			err := tt.op(vm)
			if err != nil {
				t.Fatalf("Operation error: %v", err)
			}

			result, _ := vm.Pop()
			boolVal, ok := result.(BoolValue)
			if !ok {
				t.Fatalf("Expected BoolValue, got %T", result)
			}
			if boolVal.Val != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, boolVal.Val)
			}
		})
	}
}
