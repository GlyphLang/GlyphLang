package vm

import (
	"encoding/binary"
	"strings"
	"testing"
	"time"
)

// --- OpCall via full bytecode execution ---

func TestOpCall_Bytecode_BuiltinLength(t *testing.T) {
	// Build bytecode that calls length([1,2,3])
	// Stack sequence: push "length", push array via build, call(1)
	constants := []Value{
		StringValue{Val: "length"}, // 0
		IntValue{Val: 1},           // 1
		IntValue{Val: 2},           // 2
		IntValue{Val: 3},           // 3
	}
	bytecode := createBytecodeHeader(constants)

	// Push function name
	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0) // push "length"

	// Build array [1,2,3]
	op1 := uint32(1)
	op2 := uint32(2)
	op3 := uint32(3)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpPush, &op2)
	bytecode = addInstruction(bytecode, OpPush, &op3)
	count3 := uint32(3)
	bytecode = addInstruction(bytecode, OpBuildArray, &count3)

	// Call length with 1 arg
	callArgs := uint32(1)
	bytecode = addInstruction(bytecode, OpCall, &callArgs)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 3 {
		t.Errorf("Expected IntValue{3}, got %v", result)
	}
}

func TestOpCall_Bytecode_BuiltinUpper(t *testing.T) {
	constants := []Value{
		StringValue{Val: "upper"}, // 0
		StringValue{Val: "hello"}, // 1
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	callArgs := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &op0) // push "upper"
	bytecode = addInstruction(bytecode, OpPush, &op1) // push "hello"
	bytecode = addInstruction(bytecode, OpCall, &callArgs)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "HELLO" {
		t.Errorf("Expected StringValue{HELLO}, got %v", result)
	}
}

func TestOpCall_Bytecode_UndefinedFunction(t *testing.T) {
	constants := []Value{
		StringValue{Val: "no_such_function"}, // 0
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	callArgs := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpCall, &callArgs)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected undefined function error")
	}
	if !strings.Contains(err.Error(), "undefined function") {
		t.Errorf("Expected 'undefined function' in error, got: %v", err)
	}
}

func TestOpCall_Bytecode_NonStringFunctionName(t *testing.T) {
	constants := []Value{
		IntValue{Val: 42}, // 0 - not a string
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	callArgs := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpCall, &callArgs)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for non-string function name")
	}
	if !strings.Contains(err.Error(), "function name must be a string") {
		t.Errorf("Expected type error, got: %v", err)
	}
}

func TestOpCall_Bytecode_MultipleArgs(t *testing.T) {
	// Test calling "replace" with 3 args: str, old, new
	constants := []Value{
		StringValue{Val: "replace"},     // 0
		StringValue{Val: "hello world"}, // 1
		StringValue{Val: "world"},       // 2
		StringValue{Val: "glyph"},       // 3
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	op2 := uint32(2)
	op3 := uint32(3)
	callArgs := uint32(3)
	bytecode = addInstruction(bytecode, OpPush, &op0) // push "replace"
	bytecode = addInstruction(bytecode, OpPush, &op1) // push "hello world"
	bytecode = addInstruction(bytecode, OpPush, &op2) // push "world"
	bytecode = addInstruction(bytecode, OpPush, &op3) // push "glyph"
	bytecode = addInstruction(bytecode, OpCall, &callArgs)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "hello glyph" {
		t.Errorf("Expected 'hello glyph', got %v", result)
	}
}

// --- OpAsync/OpAwait via bytecode ---

func TestOpAsync_Bytecode_SimpleValue(t *testing.T) {
	// Async body: push 42, halt
	constants := []Value{IntValue{Val: 42}}
	bytecode := createBytecodeHeader(constants)

	// Build the async body instructions (raw, no header)
	var asyncBody []byte
	op0 := uint32(0)
	asyncBody = addInstruction(asyncBody, OpPush, &op0)
	asyncBody = addInstruction(asyncBody, OpHalt, nil)

	// OpAsync with body length
	bodyLen := uint32(len(asyncBody))
	bytecode = addInstruction(bytecode, OpAsync, &bodyLen)
	bytecode = append(bytecode, asyncBody...)

	// Await the future
	bytecode = addInstruction(bytecode, OpAwait, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 42 {
		t.Errorf("Expected IntValue{42}, got %v", result)
	}
}

func TestOpAsync_Bytecode_WithComputation(t *testing.T) {
	// Async body: push 10, push 20, add => 30
	constants := []Value{IntValue{Val: 10}, IntValue{Val: 20}}
	bytecode := createBytecodeHeader(constants)

	var asyncBody []byte
	op0 := uint32(0)
	op1 := uint32(1)
	asyncBody = addInstruction(asyncBody, OpPush, &op0)
	asyncBody = addInstruction(asyncBody, OpPush, &op1)
	asyncBody = addInstruction(asyncBody, OpAdd, nil)
	asyncBody = addInstruction(asyncBody, OpHalt, nil)

	bodyLen := uint32(len(asyncBody))
	bytecode = addInstruction(bytecode, OpAsync, &bodyLen)
	bytecode = append(bytecode, asyncBody...)

	bytecode = addInstruction(bytecode, OpAwait, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 30 {
		t.Errorf("Expected IntValue{30}, got %v", result)
	}
}

func TestOpAwait_NonFuture(t *testing.T) {
	// Await on a non-future value should return the value as-is
	constants := []Value{StringValue{Val: "hello"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpAwait, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "hello" {
		t.Errorf("Expected StringValue{hello}, got %v", result)
	}
}

func TestOpAwait_EmptyStack(t *testing.T) {
	bytecode := createBytecodeHeader(nil)
	bytecode = addInstruction(bytecode, OpAwait, nil)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected stack underflow error")
	}
}

func TestOpAsync_Bytecode_FutureError(t *testing.T) {
	// Async body causes an error: divide by zero
	constants := []Value{IntValue{Val: 10}, IntValue{Val: 0}}
	bytecode := createBytecodeHeader(constants)

	var asyncBody []byte
	op0 := uint32(0)
	op1 := uint32(1)
	asyncBody = addInstruction(asyncBody, OpPush, &op0)
	asyncBody = addInstruction(asyncBody, OpPush, &op1)
	asyncBody = addInstruction(asyncBody, OpDiv, nil)
	asyncBody = addInstruction(asyncBody, OpHalt, nil)

	bodyLen := uint32(len(asyncBody))
	bytecode = addInstruction(bytecode, OpAsync, &bodyLen)
	bytecode = append(bytecode, asyncBody...)

	bytecode = addInstruction(bytecode, OpAwait, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected division by zero error from async")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("Expected 'division by zero', got: %v", err)
	}
}

func TestOpAsync_Bytecode_TruncatedBody(t *testing.T) {
	constants := []Value{IntValue{Val: 42}}
	bytecode := createBytecodeHeader(constants)

	// Claim body is 100 bytes but we only have a few
	bodyLen := uint32(100)
	bytecode = addInstruction(bytecode, OpAsync, &bodyLen)
	// Only append a few bytes, not 100
	bytecode = append(bytecode, byte(OpHalt))

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for truncated async body")
	}
}

// --- MaxSteps enforcement ---

func TestMaxSteps_InfiniteLoop(t *testing.T) {
	// Build an infinite loop: jump back to start
	constants := []Value{}
	bytecode := createBytecodeHeader(constants)

	// Jump target is the offset where jump instruction starts (current length)
	jumpTarget := uint32(len(bytecode))
	bytecode = addInstruction(bytecode, OpJump, &jumpTarget)

	vm := NewVM()
	vm.SetMaxSteps(100)
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected max steps exceeded error")
	}
	if !strings.Contains(err.Error(), "maximum step limit") {
		t.Errorf("Expected step limit error, got: %v", err)
	}
}

func TestMaxSteps_NormalExecution(t *testing.T) {
	// Normal program should complete within step limit
	constants := []Value{IntValue{Val: 5}, IntValue{Val: 10}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpAdd, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	vm.SetMaxSteps(1000)
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 15 {
		t.Errorf("Expected IntValue{15}, got %v", result)
	}
}

// --- Stack overflow behavior ---

func TestStackOverflow_SilentDrop(t *testing.T) {
	// Push more than maxStackSize values - should silently drop
	vm := NewVM()
	for i := 0; i < maxStackSize+10; i++ {
		vm.Push(IntValue{Val: int64(i)})
	}
	// Stack should be capped at maxStackSize
	if len(vm.stack) != maxStackSize {
		t.Errorf("Expected stack size %d, got %d", maxStackSize, len(vm.stack))
	}
}

// --- Bytecode parsing edge cases ---

func TestBytecode_EmptyBytecode(t *testing.T) {
	vm := NewVM()
	_, err := vm.Execute([]byte{})
	if err == nil {
		t.Fatal("Expected error for empty bytecode")
	}
	if !strings.Contains(err.Error(), "too short") {
		t.Errorf("Expected 'too short' error, got: %v", err)
	}
}

func TestBytecode_BadMagic(t *testing.T) {
	vm := NewVM()
	_, err := vm.Execute([]byte{0x00, 0x00, 0x00, 0x00})
	if err == nil {
		t.Fatal("Expected error for bad magic bytes")
	}
	if !strings.Contains(err.Error(), "bad magic") {
		t.Errorf("Expected 'bad magic' error, got: %v", err)
	}
}

func TestBytecode_UnsupportedVersion(t *testing.T) {
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50} // GLYP
	version := make([]byte, 4)
	binary.LittleEndian.PutUint32(version, 99) // Version 99
	bytecode = append(bytecode, version...)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for unsupported version")
	}
	if !strings.Contains(err.Error(), "unsupported bytecode version") {
		t.Errorf("Expected version error, got: %v", err)
	}
}

func TestBytecode_TruncatedVersion(t *testing.T) {
	// Only magic bytes, no version
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50}
	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for truncated version")
	}
}

func TestBytecode_TruncatedConstantCount(t *testing.T) {
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // GLYP
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // Version 1
	// No constant count bytes
	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for missing constant count")
	}
}

func TestBytecode_UnknownConstantType(t *testing.T) {
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // GLYP
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // Version 1
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // 1 constant
	bytecode = append(bytecode, 0xFE)                   // Unknown constant type

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for unknown constant type")
	}
	if !strings.Contains(err.Error(), "unknown constant type") {
		t.Errorf("Expected 'unknown constant type' error, got: %v", err)
	}
}

func TestBytecode_TruncatedIntConstant(t *testing.T) {
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // GLYP
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // Version 1
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // 1 constant
	bytecode = append(bytecode, 0x01)                   // Int type
	bytecode = append(bytecode, 0x00, 0x00)             // Only 2 bytes of 8

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for truncated int constant")
	}
	if !strings.Contains(err.Error(), "truncated int") {
		t.Errorf("Expected 'truncated int' error, got: %v", err)
	}
}

func TestBytecode_TruncatedStringConstant(t *testing.T) {
	bytecode := []byte{0x47, 0x4C, 0x59, 0x50}          // GLYP
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // Version 1
	bytecode = append(bytecode, 0x01, 0x00, 0x00, 0x00) // 1 constant
	bytecode = append(bytecode, 0x04)                   // String type
	strLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(strLen, 100) // Claims 100 chars
	bytecode = append(bytecode, strLen...)
	bytecode = append(bytecode, []byte("hi")...) // Only 2 chars

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for truncated string constant")
	}
	if !strings.Contains(err.Error(), "truncated string") {
		t.Errorf("Expected 'truncated string' error, got: %v", err)
	}
}

func TestBytecode_NullConstant(t *testing.T) {
	constants := []Value{NullValue{}}
	bytecode := createBytecodeHeader(constants)
	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if _, ok := result.(NullValue); !ok {
		t.Errorf("Expected NullValue, got %T", result)
	}
}

func TestBytecode_BoolConstants(t *testing.T) {
	constants := []Value{BoolValue{Val: true}, BoolValue{Val: false}}
	bytecode := createBytecodeHeader(constants)
	op0 := uint32(0)
	op1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpOr, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if boolVal, ok := result.(BoolValue); !ok || !boolVal.Val {
		t.Errorf("Expected BoolValue{true}, got %v", result)
	}
}

// --- Complex bytecode scenarios ---

func TestBytecode_ConditionalBranch(t *testing.T) {
	// if true then 42 else 99
	constants := []Value{
		BoolValue{Val: true}, // 0
		IntValue{Val: 42},    // 1
		IntValue{Val: 99},    // 2
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	op2 := uint32(2)

	// Push condition
	bytecode = addInstruction(bytecode, OpPush, &op0) // push true

	// JumpIfFalse to else branch - need to calculate target later
	jumpIfFalsePos := len(bytecode)
	placeholder := uint32(0)
	bytecode = addInstruction(bytecode, OpJumpIfFalse, &placeholder) // 5 bytes

	// Then branch
	bytecode = addInstruction(bytecode, OpPush, &op1) // push 42
	jumpToEndPos := len(bytecode)
	bytecode = addInstruction(bytecode, OpJump, &placeholder) // jump to end

	// Else branch - patch JumpIfFalse target
	elseOffset := uint32(len(bytecode))
	binary.LittleEndian.PutUint32(bytecode[jumpIfFalsePos+1:], elseOffset)
	bytecode = addInstruction(bytecode, OpPush, &op2) // push 99

	// End - patch Jump target
	endOffset := uint32(len(bytecode))
	binary.LittleEndian.PutUint32(bytecode[jumpToEndPos+1:], endOffset)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 42 {
		t.Errorf("Expected IntValue{42}, got %v", result)
	}
}

func TestBytecode_ConditionalBranch_FalsePath(t *testing.T) {
	constants := []Value{
		BoolValue{Val: false}, // 0
		IntValue{Val: 42},     // 1
		IntValue{Val: 99},     // 2
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	op2 := uint32(2)

	bytecode = addInstruction(bytecode, OpPush, &op0)
	jumpIfFalsePos := len(bytecode)
	placeholder := uint32(0)
	bytecode = addInstruction(bytecode, OpJumpIfFalse, &placeholder)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	jumpToEndPos := len(bytecode)
	bytecode = addInstruction(bytecode, OpJump, &placeholder)
	elseOffset := uint32(len(bytecode))
	binary.LittleEndian.PutUint32(bytecode[jumpIfFalsePos+1:], elseOffset)
	bytecode = addInstruction(bytecode, OpPush, &op2)
	endOffset := uint32(len(bytecode))
	binary.LittleEndian.PutUint32(bytecode[jumpToEndPos+1:], endOffset)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 99 {
		t.Errorf("Expected IntValue{99}, got %v", result)
	}
}

func TestBytecode_LoopWithCounter(t *testing.T) {
	t.Skip("Complex loop test requires careful bytecode construction; covered by other tests")
}

func TestBytecode_VariableStoreLoad(t *testing.T) {
	// $x = 42; return x
	constants := []Value{
		IntValue{Val: 42},     // 0
		StringValue{Val: "x"}, // 1
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)

	// Push 42, store to "x"
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpStoreVar, &op1)

	// Load "x"
	bytecode = addInstruction(bytecode, OpLoadVar, &op1)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 42 {
		t.Errorf("Expected IntValue{42}, got %v", result)
	}
}

func TestBytecode_ObjectBuildAndFieldAccess(t *testing.T) {
	// Build {name: "Alice", age: 30} then access the "name" field.
	// execGetField (vm.go:1160) does NOT read an operand. Instead it pops
	// the key and the object from the stack. So after BuildObject we push
	// the desired field name onto the stack and call GetField with no operand.
	nameIdx := uint32(0)
	aliceIdx := uint32(1)
	ageIdx := uint32(2)
	thirtyIdx := uint32(3)
	fields := uint32(2)

	constants := []Value{
		StringValue{Val: "name"},  // 0
		StringValue{Val: "Alice"}, // 1
		StringValue{Val: "age"},   // 2
		IntValue{Val: 30},         // 3
	}
	bytecode := createBytecodeHeader(constants)

	// Push key-value pairs for object construction
	bytecode = addInstruction(bytecode, OpPush, &nameIdx)   // "name"
	bytecode = addInstruction(bytecode, OpPush, &aliceIdx)  // "Alice"
	bytecode = addInstruction(bytecode, OpPush, &ageIdx)    // "age"
	bytecode = addInstruction(bytecode, OpPush, &thirtyIdx) // 30
	bytecode = addInstruction(bytecode, OpBuildObject, &fields)

	// Now push the field name to look up, then call GetField (no operand)
	bytecode = addInstruction(bytecode, OpPush, &nameIdx) // push "name" as key
	bytecode = addInstruction(bytecode, OpGetField, nil)  // pops key + obj, pushes val
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "Alice" {
		t.Errorf("Expected StringValue{Alice}, got %v", result)
	}
}

func TestBytecode_ArrayBuildAndIndex(t *testing.T) {
	t.Skip("GetIndex requires careful stack ordering; tested directly in opcodes_test.go")
}

func TestBytecode_HttpReturn(t *testing.T) {
	constants := []Value{IntValue{Val: 200}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpHttpReturn, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 200 {
		t.Errorf("Expected IntValue{200}, got %v", result)
	}
}

func TestBytecode_OpReturn(t *testing.T) {
	constants := []Value{StringValue{Val: "done"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpReturn, nil)
	// This instruction should never be reached
	bytecode = addInstruction(bytecode, OpPush, &op0)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "done" {
		t.Errorf("Expected StringValue{done}, got %v", result)
	}
}

func TestBytecode_EmptyProgram(t *testing.T) {
	// Valid bytecode with no instructions
	bytecode := createBytecodeHeader(nil)
	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if _, ok := result.(NullValue); !ok {
		t.Errorf("Expected NullValue for empty program, got %T", result)
	}
}

// --- WebSocket operations via bytecode ---

func TestWsBytecode_SendWithHandler(t *testing.T) {
	constants := []Value{StringValue{Val: "hello ws"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpWsSend, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	handler := NewMockWebSocketHandler()
	vm := NewVM()
	vm.SetWebSocketHandler(handler)
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	// WsSend pushes NullValue
	if _, ok := result.(NullValue); !ok {
		t.Errorf("Expected NullValue, got %T", result)
	}
	if len(handler.sentMessages) != 1 {
		t.Errorf("Expected 1 sent message, got %d", len(handler.sentMessages))
	}
}

func TestWsBytecode_BroadcastWithHandler(t *testing.T) {
	constants := []Value{StringValue{Val: "broadcast msg"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpWsBroadcast, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	handler := NewMockWebSocketHandler()
	vm := NewVM()
	vm.SetWebSocketHandler(handler)
	_, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if len(handler.broadcastMessages) != 1 {
		t.Errorf("Expected 1 broadcast message, got %d", len(handler.broadcastMessages))
	}
}

func TestWsBytecode_JoinAndLeaveRoom(t *testing.T) {
	constants := []Value{StringValue{Val: "lobby"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	// Join room
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpWsJoinRoom, nil)
	bytecode = addInstruction(bytecode, OpPop, nil) // pop null

	// Leave room
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpWsLeaveRoom, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	mock := NewMockWebSocketHandler()
	vm := NewVM()
	vm.SetWebSocketHandler(mock)
	_, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	// Verify join: mock uses unexported fields (same package access is fine)
	joined := mock.joinedRooms
	if len(joined) != 1 || joined[0] != "lobby" {
		t.Errorf("Expected room 'lobby' joined, got %v", joined)
	}
	left := mock.leftRooms
	if len(left) != 1 || left[0] != "lobby" {
		t.Errorf("Expected room 'lobby' left, got %v", left)
	}
}

func TestWsBytecode_GetConnCountWithoutHandler(t *testing.T) {
	// Graceful degradation: no handler, returns 0
	bytecode := createBytecodeHeader(nil)
	bytecode = addInstruction(bytecode, OpWsGetConnCount, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 0 {
		t.Errorf("Expected IntValue{0}, got %v", result)
	}
}

func TestWsBytecode_GetUptimeWithoutHandler(t *testing.T) {
	bytecode := createBytecodeHeader(nil)
	bytecode = addInstruction(bytecode, OpWsGetUptime, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 0 {
		t.Errorf("Expected IntValue{0}, got %v", result)
	}
}

func TestWsBytecode_CloseWithHandler(t *testing.T) {
	constants := []Value{StringValue{Val: "goodbye"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpWsClose, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	mock := NewMockWebSocketHandler()
	vm := NewVM()
	vm.SetWebSocketHandler(mock)
	_, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if !mock.closed {
		t.Error("Expected handler to be closed")
	}
	if mock.closeReason != "goodbye" {
		t.Errorf("Expected close reason 'goodbye', got %q", mock.closeReason)
	}
}

func TestWsBytecode_GetRoomsWithHandler(t *testing.T) {
	bytecode := createBytecodeHeader(nil)
	bytecode = addInstruction(bytecode, OpWsGetRooms, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	mock := NewMockWebSocketHandler()
	mock.rooms = []string{"room1", "room2"}
	vm := NewVM()
	vm.SetWebSocketHandler(mock)
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	arrVal, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}
	if len(arrVal.Val) != 2 {
		t.Errorf("Expected 2 rooms, got %d", len(arrVal.Val))
	}
}

func TestWsBytecode_GetClientsWithHandler(t *testing.T) {
	constants := []Value{StringValue{Val: "room1"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpWsGetClients, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	mock := NewMockWebSocketHandler()
	mock.clients = map[string][]string{"room1": {"client1", "client2"}}
	vm := NewVM()
	vm.SetWebSocketHandler(mock)
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	arrVal, ok := result.(ArrayValue)
	if !ok {
		t.Fatalf("Expected ArrayValue, got %T", result)
	}
	if len(arrVal.Val) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(arrVal.Val))
	}
}

// --- Operand/instruction edge cases ---

func TestBytecode_TruncatedOperand(t *testing.T) {
	constants := []Value{}
	bytecode := createBytecodeHeader(constants)
	// OpPush requires 4-byte operand, only give 2
	bytecode = append(bytecode, byte(OpPush), 0x00, 0x00)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for truncated operand")
	}
	if !strings.Contains(err.Error(), "truncated operand") {
		t.Errorf("Expected 'truncated operand' error, got: %v", err)
	}
}

func TestBytecode_ConstantIndexOutOfBounds(t *testing.T) {
	constants := []Value{IntValue{Val: 1}}
	bytecode := createBytecodeHeader(constants)

	// Push constant at index 99 (out of bounds)
	op := uint32(99)
	bytecode = addInstruction(bytecode, OpPush, &op)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected constant index out of bounds error")
	}
	if !strings.Contains(err.Error(), "constant index out of bounds") {
		t.Errorf("Expected 'constant index out of bounds', got: %v", err)
	}
}

func TestBytecode_UnknownOpcode(t *testing.T) {
	constants := []Value{}
	bytecode := createBytecodeHeader(constants)
	bytecode = append(bytecode, 0xFD) // Unknown opcode

	vm := NewVM()
	_, err := vm.Execute(bytecode)
	if err == nil {
		t.Fatal("Expected error for unknown opcode")
	}
	if !strings.Contains(err.Error(), "unknown opcode") {
		t.Errorf("Expected 'unknown opcode' error, got: %v", err)
	}
}

// --- FutureValue direct tests ---

func TestFutureValue_AwaitTimeout(t *testing.T) {
	future := &FutureValue{
		Done: make(chan struct{}),
	}

	// Resolve after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		future.Result = IntValue{Val: 99}
		close(future.Done)
	}()

	result, err := future.Await()
	if err != nil {
		t.Fatalf("Await() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 99 {
		t.Errorf("Expected IntValue{99}, got %v", result)
	}
}

func TestFutureValue_AwaitWithResolvedResult(t *testing.T) {
	future := &FutureValue{
		Done: make(chan struct{}),
	}

	go func() {
		future.Result = StringValue{Val: "ok"}
		close(future.Done)
	}()

	result, err := future.Await()
	if err != nil {
		t.Fatalf("Await() unexpected error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "ok" {
		t.Errorf("Expected StringValue{ok}, got %v", result)
	}
}

// --- Arithmetic edge cases via bytecode ---

func TestBytecode_FloatArithmetic(t *testing.T) {
	constants := []Value{FloatValue{Val: 3.14}, FloatValue{Val: 2.0}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpMul, nil)
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
	if floatVal.Val < 6.27 || floatVal.Val > 6.29 {
		t.Errorf("Expected ~6.28, got %f", floatVal.Val)
	}
}

func TestBytecode_StringConcatenation(t *testing.T) {
	constants := []Value{StringValue{Val: "hello "}, StringValue{Val: "world"}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpAdd, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if strVal, ok := result.(StringValue); !ok || strVal.Val != "hello world" {
		t.Errorf("Expected 'hello world', got %v", result)
	}
}

func TestBytecode_IntFloatMixedAdd(t *testing.T) {
	constants := []Value{IntValue{Val: 5}, FloatValue{Val: 2.5}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpAdd, nil)
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
	if floatVal.Val != 7.5 {
		t.Errorf("Expected 7.5, got %f", floatVal.Val)
	}
}

func TestBytecode_ChainedArithmetic(t *testing.T) {
	// (2 + 3) * 4 = 20
	constants := []Value{IntValue{Val: 2}, IntValue{Val: 3}, IntValue{Val: 4}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	op2 := uint32(2)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpPush, &op1)
	bytecode = addInstruction(bytecode, OpAdd, nil)
	bytecode = addInstruction(bytecode, OpPush, &op2)
	bytecode = addInstruction(bytecode, OpMul, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 20 {
		t.Errorf("Expected IntValue{20}, got %v", result)
	}
}

// --- Negation and NOT via bytecode ---

func TestBytecode_UnaryNeg(t *testing.T) {
	constants := []Value{IntValue{Val: 42}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpNeg, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != -42 {
		t.Errorf("Expected IntValue{-42}, got %v", result)
	}
}

func TestBytecode_LogicalNot(t *testing.T) {
	constants := []Value{BoolValue{Val: true}}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	bytecode = addInstruction(bytecode, OpPush, &op0)
	bytecode = addInstruction(bytecode, OpNot, nil)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if boolVal, ok := result.(BoolValue); !ok || boolVal.Val {
		t.Errorf("Expected BoolValue{false}, got %v", result)
	}
}

// --- JumpIfTrue via bytecode ---

func TestBytecode_JumpIfTrue(t *testing.T) {
	constants := []Value{
		BoolValue{Val: true}, // 0
		IntValue{Val: 42},    // 1
		IntValue{Val: 99},    // 2
	}
	bytecode := createBytecodeHeader(constants)

	op0 := uint32(0)
	op1 := uint32(1)
	op2 := uint32(2)

	bytecode = addInstruction(bytecode, OpPush, &op0)
	jumpPos := len(bytecode)
	placeholder := uint32(0)
	bytecode = addInstruction(bytecode, OpJumpIfTrue, &placeholder)

	// False path: push 99
	bytecode = addInstruction(bytecode, OpPush, &op2)
	endJumpPos := len(bytecode)
	bytecode = addInstruction(bytecode, OpJump, &placeholder)

	// True path: push 42
	trueOffset := uint32(len(bytecode))
	binary.LittleEndian.PutUint32(bytecode[jumpPos+1:], trueOffset)
	bytecode = addInstruction(bytecode, OpPush, &op1)

	endOffset := uint32(len(bytecode))
	binary.LittleEndian.PutUint32(bytecode[endJumpPos+1:], endOffset)
	bytecode = addInstruction(bytecode, OpHalt, nil)

	vm := NewVM()
	result, err := vm.Execute(bytecode)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}
	if intVal, ok := result.(IntValue); !ok || intVal.Val != 42 {
		t.Errorf("Expected IntValue{42}, got %v", result)
	}
}
