package decompiler

import (
	"testing"
)

func TestDecompileValidBytecode(t *testing.T) {
	// Create a simple valid bytecode:
	// Magic: GLYP
	// Version: 1
	// Constants: 2 (string "greeting", string "Hello!")
	// Instructions: PUSH 1, STORE_VAR 0, LOAD_VAR 0, RETURN, HALT
	bytecode := []byte{
		// Magic
		'G', 'L', 'Y', 'P',
		// Version (little-endian u32)
		0x01, 0x00, 0x00, 0x00,
		// Constant count (2)
		0x02, 0x00, 0x00, 0x00,
		// Constant 0: string "greeting" (type 0x04)
		0x04, 0x08, 0x00, 0x00, 0x00, 'g', 'r', 'e', 'e', 't', 'i', 'n', 'g',
		// Constant 1: string "Hello!" (type 0x04)
		0x04, 0x06, 0x00, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o', '!',
		// Instruction count (17 bytes)
		0x11, 0x00, 0x00, 0x00,
		// PUSH 1 (opcode 0x01, operand 1)
		0x01, 0x01, 0x00, 0x00, 0x00,
		// STORE_VAR 0 (opcode 0x41, operand 0)
		0x41, 0x00, 0x00, 0x00, 0x00,
		// LOAD_VAR 0 (opcode 0x40, operand 0)
		0x40, 0x00, 0x00, 0x00, 0x00,
		// RETURN (opcode 0x61)
		0x61,
		// HALT (opcode 0xFF)
		0xFF,
	}

	dec := NewDecompiler()
	result, err := dec.Decompile(bytecode)
	if err != nil {
		t.Fatalf("Decompile failed: %v", err)
	}

	// Check version
	if result.Version != 1 {
		t.Errorf("Expected version 1, got %d", result.Version)
	}

	// Check constants
	if len(result.Constants) != 2 {
		t.Errorf("Expected 2 constants, got %d", len(result.Constants))
	}

	if result.Constants[0].Value != `"greeting"` {
		t.Errorf("Expected constant 0 to be \"greeting\", got %s", result.Constants[0].Value)
	}

	if result.Constants[1].Value != `"Hello!"` {
		t.Errorf("Expected constant 1 to be \"Hello!\", got %s", result.Constants[1].Value)
	}

	// Check instructions
	if len(result.Instructions) != 5 {
		t.Errorf("Expected 5 instructions, got %d", len(result.Instructions))
	}

	expectedOps := []string{"PUSH", "STORE_VAR", "LOAD_VAR", "RETURN", "HALT"}
	for i, op := range expectedOps {
		if result.Instructions[i].Opcode != op {
			t.Errorf("Instruction %d: expected %s, got %s", i, op, result.Instructions[i].Opcode)
		}
	}
}

func TestDecompileInvalidMagic(t *testing.T) {
	bytecode := []byte{'B', 'A', 'D', '!', 0x01, 0x00, 0x00, 0x00}

	dec := NewDecompiler()
	_, err := dec.Decompile(bytecode)
	if err == nil {
		t.Error("Expected error for invalid magic bytes")
	}
}

func TestDecompileTooShort(t *testing.T) {
	bytecode := []byte{'G', 'L'}

	dec := NewDecompiler()
	_, err := dec.Decompile(bytecode)
	if err == nil {
		t.Error("Expected error for too short bytecode")
	}
}

func TestDecompileConstantTypes(t *testing.T) {
	// Test all constant types
	bytecode := []byte{
		// Magic
		'G', 'L', 'Y', 'P',
		// Version
		0x01, 0x00, 0x00, 0x00,
		// Constant count (5)
		0x05, 0x00, 0x00, 0x00,
		// Constant 0: null (type 0x00)
		0x00,
		// Constant 1: int 42 (type 0x01)
		0x01, 0x2A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Constant 2: float 3.14 (type 0x02) - approximation
		0x02, 0x1F, 0x85, 0xEB, 0x51, 0xB8, 0x1E, 0x09, 0x40,
		// Constant 3: bool true (type 0x03)
		0x03, 0x01,
		// Constant 4: bool false (type 0x03)
		0x03, 0x00,
		// Instruction count (1)
		0x01, 0x00, 0x00, 0x00,
		// HALT
		0xFF,
	}

	dec := NewDecompiler()
	result, err := dec.Decompile(bytecode)
	if err != nil {
		t.Fatalf("Decompile failed: %v", err)
	}

	if len(result.Constants) != 5 {
		t.Fatalf("Expected 5 constants, got %d", len(result.Constants))
	}

	// Check types
	expectedTypes := []string{"null", "int", "float", "bool", "bool"}
	for i, expectedType := range expectedTypes {
		if result.Constants[i].Type != expectedType {
			t.Errorf("Constant %d: expected type %s, got %s", i, expectedType, result.Constants[i].Type)
		}
	}

	// Check values
	if result.Constants[0].Value != "null" {
		t.Errorf("Constant 0: expected null, got %s", result.Constants[0].Value)
	}
	if result.Constants[1].Value != "42" {
		t.Errorf("Constant 1: expected 42, got %s", result.Constants[1].Value)
	}
	if result.Constants[3].Value != "true" {
		t.Errorf("Constant 3: expected true, got %s", result.Constants[3].Value)
	}
	if result.Constants[4].Value != "false" {
		t.Errorf("Constant 4: expected false, got %s", result.Constants[4].Value)
	}
}

func TestFormatDisassembly(t *testing.T) {
	bytecode := []byte{
		'G', 'L', 'Y', 'P',
		0x01, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x04, 0x04, 0x00, 0x00, 0x00, 't', 'e', 's', 't',
		0x02, 0x00, 0x00, 0x00,
		0x61, // RETURN
		0xFF, // HALT
	}

	dec := NewDecompiler()
	result, err := dec.Decompile(bytecode)
	if err != nil {
		t.Fatalf("Decompile failed: %v", err)
	}

	disasm := result.FormatDisassembly()
	if disasm == "" {
		t.Error("FormatDisassembly returned empty string")
	}

	// Check that it contains expected sections
	if !contains(disasm, "CONSTANT POOL") {
		t.Error("Disassembly missing CONSTANT POOL section")
	}
	if !contains(disasm, "INSTRUCTIONS") {
		t.Error("Disassembly missing INSTRUCTIONS section")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
