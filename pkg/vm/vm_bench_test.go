package vm

import (
	"testing"
)

// BenchmarkVMSimpleArithmetic benchmarks simple arithmetic operations
func BenchmarkVMSimpleArithmetic(b *testing.B) {
	vm := NewVM()

	// Prepare test values
	val1 := IntValue{Val: 10}
	val2 := IntValue{Val: 20}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		vm.Push(val1)
		vm.Push(val2)

		// Pop and add
		v2, _ := vm.Pop()
		v1, _ := vm.Pop()

		// Simulate addition
		result := IntValue{Val: v1.(IntValue).Val + v2.(IntValue).Val}
		vm.Push(result)

		// Clean up
		vm.Pop()
	}
}

// BenchmarkVMObjectCreation benchmarks object creation
func BenchmarkVMObjectCreation(b *testing.B) {
	vm := NewVM()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate creating an object with multiple fields
		vm.Push(IntValue{Val: 1})
		vm.Push(StringValue{Val: "John Doe"})
		vm.Push(StringValue{Val: "john@example.com"})

		// Pop all fields (simulating object construction)
		vm.Pop()
		vm.Pop()
		vm.Pop()

		// Push result object (in real implementation, this would be an ObjectValue)
		vm.Push(StringValue{Val: "object"})
		vm.Pop()
	}
}

// BenchmarkVMArrayCreation benchmarks array creation and operations
func BenchmarkVMArrayCreation(b *testing.B) {
	vm := NewVM()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate creating an array with 10 elements
		for j := 0; j < 10; j++ {
			vm.Push(IntValue{Val: int64(j)})
		}

		// Pop all elements
		for j := 0; j < 10; j++ {
			vm.Pop()
		}
	}
}

// BenchmarkVMRouteExecution benchmarks executing a simple route with valid bytecode
func BenchmarkVMRouteExecution(b *testing.B) {
	vm := NewVM()

	// Create valid bytecode that matches VM's expected format:
	// Magic: GLYP (4 bytes)
	// Version: 1 (4 bytes, little-endian)
	// Constant count: 1 (4 bytes, little-endian)
	// Constant 0: String "Hello" (type=0x04, len=5, data)
	// Instruction count: 2 (4 bytes, little-endian)
	// Instructions: PUSH 0, RETURN
	bytecode := []byte{
		// Magic bytes
		0x47, 0x4C, 0x59, 0x50, // "GLYP"
		// Version (uint32 LE)
		0x01, 0x00, 0x00, 0x00, // Version 1
		// Constant count (uint32 LE)
		0x01, 0x00, 0x00, 0x00, // 1 constant
		// Constant 0: String "Hello"
		0x04,                   // Type: String
		0x05, 0x00, 0x00, 0x00, // Length: 5 (uint32 LE)
		0x48, 0x65, 0x6c, 0x6c, 0x6f, // "Hello"
		// Instruction count (uint32 LE)
		0x06, 0x00, 0x00, 0x00, // 6 bytes of instructions
		// Instructions
		0x01,                   // OpPush
		0x00, 0x00, 0x00, 0x00, // Constant index 0
		0x61, // OpReturn (0x61)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := vm.Execute(bytecode)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
		if result == nil {
			b.Fatal("Result is nil")
		}
	}
}

// BenchmarkVMStackOperations benchmarks raw stack push/pop operations
func BenchmarkVMStackOperations(b *testing.B) {
	vm := NewVM()
	val := IntValue{Val: 42}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		vm.Push(val)
		vm.Pop()
	}
}

// BenchmarkVMComplexOperation benchmarks a complex operation with multiple steps
func BenchmarkVMComplexOperation(b *testing.B) {
	vm := NewVM()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate: result = (a + b) * (c - d)
		vm.Push(IntValue{Val: 10}) // a
		vm.Push(IntValue{Val: 20}) // b
		v2, _ := vm.Pop()
		v1, _ := vm.Pop()
		sum := IntValue{Val: v1.(IntValue).Val + v2.(IntValue).Val}
		vm.Push(sum)

		vm.Push(IntValue{Val: 30}) // c
		vm.Push(IntValue{Val: 5})  // d
		v4, _ := vm.Pop()
		v3, _ := vm.Pop()
		diff := IntValue{Val: v3.(IntValue).Val - v4.(IntValue).Val}
		vm.Push(diff)

		v6, _ := vm.Pop()
		v5, _ := vm.Pop()
		result := IntValue{Val: v5.(IntValue).Val * v6.(IntValue).Val}
		vm.Push(result)

		vm.Pop() // Clean up
	}
}

// BenchmarkVMGlobalVarAccess benchmarks global variable access
func BenchmarkVMGlobalVarAccess(b *testing.B) {
	vm := NewVM()
	vm.globals["test_var"] = IntValue{Val: 42}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		val := vm.globals["test_var"]
		vm.Push(val)
		vm.Pop()
	}
}

// BenchmarkVMStringConcatenation benchmarks string operations
func BenchmarkVMStringConcatenation(b *testing.B) {
	vm := NewVM()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		vm.Push(StringValue{Val: "Hello, "})
		vm.Push(StringValue{Val: "World!"})

		v2, _ := vm.Pop()
		v1, _ := vm.Pop()

		// Simulate concatenation
		result := StringValue{Val: v1.(StringValue).Val + v2.(StringValue).Val}
		vm.Push(result)
		vm.Pop()
	}
}

// BenchmarkVMBooleanOperations benchmarks boolean operations
func BenchmarkVMBooleanOperations(b *testing.B) {
	vm := NewVM()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		vm.Push(IntValue{Val: 10})
		vm.Push(IntValue{Val: 20})

		v2, _ := vm.Pop()
		v1, _ := vm.Pop()

		// Simulate comparison (less than)
		result := BoolValue{Val: v1.(IntValue).Val < v2.(IntValue).Val}
		vm.Push(result)
		vm.Pop()
	}
}
