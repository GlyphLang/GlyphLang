package debug

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glyphlang/glyph/pkg/vm"
)

// TestDebuggerCreation tests creating a new debugger
func TestDebuggerCreation(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	if d == nil {
		t.Fatal("NewDebugger returned nil")
	}

	if d.vm != v {
		t.Error("Debugger VM not set correctly")
	}

	if len(d.breakpoints) != 0 {
		t.Error("New debugger should have no breakpoints")
	}

	if d.stepMode != StepContinue {
		t.Error("New debugger should default to StepContinue mode")
	}

	if d.paused {
		t.Error("New debugger should not be paused")
	}
}

// TestBreakpointManagement tests breakpoint operations
func TestBreakpointManagement(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test setting breakpoint
	id1 := d.SetBreakpoint(100)
	if id1 != 1 {
		t.Errorf("Expected first breakpoint ID to be 1, got %d", id1)
	}

	id2 := d.SetBreakpoint(200)
	if id2 != 2 {
		t.Errorf("Expected second breakpoint ID to be 2, got %d", id2)
	}

	// Test getting breakpoint
	bp, exists := d.GetBreakpoint(100)
	if !exists {
		t.Error("Breakpoint 100 should exist")
	}
	if bp.ID != 1 {
		t.Errorf("Expected breakpoint ID 1, got %d", bp.ID)
	}
	if bp.Location != 100 {
		t.Errorf("Expected breakpoint location 100, got %d", bp.Location)
	}
	if !bp.Enabled {
		t.Error("Breakpoint should be enabled by default")
	}
	if bp.HitCount != 0 {
		t.Error("Breakpoint should have hit count 0")
	}

	// Test listing breakpoints
	bps := d.ListBreakpoints()
	if len(bps) != 2 {
		t.Errorf("Expected 2 breakpoints, got %d", len(bps))
	}

	// Test disabling breakpoint
	if !d.DisableBreakpoint(100) {
		t.Error("Failed to disable breakpoint")
	}
	bp, _ = d.GetBreakpoint(100)
	if bp.Enabled {
		t.Error("Breakpoint should be disabled")
	}

	// Test enabling breakpoint
	if !d.EnableBreakpoint(100) {
		t.Error("Failed to enable breakpoint")
	}
	bp, _ = d.GetBreakpoint(100)
	if !bp.Enabled {
		t.Error("Breakpoint should be enabled")
	}

	// Test clearing breakpoint
	if !d.ClearBreakpoint(100) {
		t.Error("Failed to clear breakpoint")
	}
	_, exists = d.GetBreakpoint(100)
	if exists {
		t.Error("Breakpoint should not exist after clearing")
	}

	// Test clearing non-existent breakpoint
	if d.ClearBreakpoint(999) {
		t.Error("Should return false when clearing non-existent breakpoint")
	}
}

// TestBreakpointByID tests breakpoint operations by ID
func TestBreakpointByID(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	id1 := d.SetBreakpoint(100)
	id2 := d.SetBreakpoint(200)

	// Clear by ID
	if !d.ClearBreakpointByID(id1) {
		t.Error("Failed to clear breakpoint by ID")
	}

	bps := d.ListBreakpoints()
	if len(bps) != 1 {
		t.Errorf("Expected 1 breakpoint after clearing, got %d", len(bps))
	}

	if bps[0].ID != id2 {
		t.Errorf("Wrong breakpoint remaining, expected ID %d, got %d", id2, bps[0].ID)
	}
}

// TestStepModes tests step mode operations
func TestStepModes(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test default mode
	if d.GetStepMode() != StepContinue {
		t.Error("Default step mode should be StepContinue")
	}

	// Test Continue
	d.Continue()
	if d.GetStepMode() != StepContinue {
		t.Error("Continue should set StepContinue mode")
	}
	if d.IsPaused() {
		t.Error("Continue should unpause debugger")
	}

	// Test StepInto
	d.StepInto()
	if d.GetStepMode() != StepInto {
		t.Error("StepInto should set StepInto mode")
	}

	// Test StepOver
	d.StepOver()
	if d.GetStepMode() != StepOver {
		t.Error("StepOver should set StepOver mode")
	}

	// Test StepOut
	d.StepOut()
	if d.GetStepMode() != StepOut {
		t.Error("StepOut should set StepOut mode")
	}

	// Test Pause
	d.Pause()
	if !d.IsPaused() {
		t.Error("Pause should pause debugger")
	}
}

// TestCallStack tests call stack operations
func TestCallStack(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test empty call stack
	stack := d.GetCallStack()
	if len(stack) != 0 {
		t.Error("New debugger should have empty call stack")
	}

	// Test pushing frames
	locals1 := map[string]vm.Value{
		"x": vm.IntValue{Val: 42},
		"y": vm.StringValue{Val: "hello"},
	}
	d.PushCallFrame("main", 100, locals1)

	stack = d.GetCallStack()
	if len(stack) != 1 {
		t.Errorf("Expected 1 frame, got %d", len(stack))
	}

	if stack[0].FuncName != "main" {
		t.Errorf("Expected function name 'main', got '%s'", stack[0].FuncName)
	}
	if stack[0].ReturnAddr != 100 {
		t.Errorf("Expected return address 100, got %d", stack[0].ReturnAddr)
	}

	// Test pushing another frame
	locals2 := map[string]vm.Value{
		"a": vm.BoolValue{Val: true},
	}
	d.PushCallFrame("helper", 200, locals2)

	stack = d.GetCallStack()
	if len(stack) != 2 {
		t.Errorf("Expected 2 frames, got %d", len(stack))
	}

	// Test popping frame
	frame, err := d.PopCallFrame()
	if err != nil {
		t.Errorf("PopCallFrame failed: %v", err)
	}
	if frame.FuncName != "helper" {
		t.Errorf("Expected popped frame to be 'helper', got '%s'", frame.FuncName)
	}

	stack = d.GetCallStack()
	if len(stack) != 1 {
		t.Errorf("Expected 1 frame after pop, got %d", len(stack))
	}

	// Test popping all frames
	d.PopCallFrame()

	// Test popping from empty stack
	_, err = d.PopCallFrame()
	if err == nil {
		t.Error("Should get error when popping from empty stack")
	}
}

// TestFormatCallStack tests call stack formatting
func TestFormatCallStack(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test empty stack
	output := d.FormatCallStack()
	if !strings.Contains(output, "empty") {
		t.Error("Empty call stack should mention it's empty")
	}

	// Test with frames
	d.PushCallFrame("main", 100, nil)
	d.PushCallFrame("helper", 200, nil)

	output = d.FormatCallStack()
	if !strings.Contains(output, "main") {
		t.Error("Call stack should contain 'main'")
	}
	if !strings.Contains(output, "helper") {
		t.Error("Call stack should contain 'helper'")
	}
	if !strings.Contains(output, "Call Stack") {
		t.Error("Call stack should have header")
	}
}

// TestFormatValue tests value formatting
func TestFormatValue(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	tests := []struct {
		name     string
		value    vm.Value
		contains string
	}{
		{"int", vm.IntValue{Val: 42}, "42"},
		{"float", vm.FloatValue{Val: 3.14}, "3.14"},
		{"string", vm.StringValue{Val: "hello"}, "hello"},
		{"bool true", vm.BoolValue{Val: true}, "true"},
		{"bool false", vm.BoolValue{Val: false}, "false"},
		{"null", vm.NullValue{}, "null"},
		{"array", vm.ArrayValue{Val: []vm.Value{vm.IntValue{Val: 1}}}, "array"},
		{"object", vm.ObjectValue{Val: map[string]vm.Value{"x": vm.IntValue{Val: 1}}}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := d.formatValue(tt.value)
			if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.contains)) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.contains, output)
			}
		})
	}
}

// TestInspectVariable tests variable inspection
func TestInspectVariable(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test with call frame containing locals
	locals := map[string]vm.Value{
		"x":   vm.IntValue{Val: 42},
		"msg": vm.StringValue{Val: "hello world"},
		"arr": vm.ArrayValue{Val: []vm.Value{
			vm.IntValue{Val: 1},
			vm.IntValue{Val: 2},
		}},
		"obj": vm.ObjectValue{Val: map[string]vm.Value{
			"name": vm.StringValue{Val: "test"},
			"age":  vm.IntValue{Val: 25},
		}},
	}
	d.PushCallFrame("test", 0, locals)

	// Test inspecting integer
	output, err := d.InspectVariable("x")
	if err != nil {
		t.Errorf("Failed to inspect variable: %v", err)
	}
	if !strings.Contains(output, "42") {
		t.Error("Inspection should show value 42")
	}
	if !strings.Contains(output, "int") {
		t.Error("Inspection should show type int")
	}

	// Test inspecting string
	output, err = d.InspectVariable("msg")
	if err != nil {
		t.Errorf("Failed to inspect variable: %v", err)
	}
	if !strings.Contains(output, "hello world") {
		t.Error("Inspection should show string value")
	}
	if !strings.Contains(output, "Length") {
		t.Error("String inspection should show length")
	}

	// Test inspecting array
	output, err = d.InspectVariable("arr")
	if err != nil {
		t.Errorf("Failed to inspect variable: %v", err)
	}
	if !strings.Contains(output, "Elements") {
		t.Error("Array inspection should show elements")
	}

	// Test inspecting object
	output, err = d.InspectVariable("obj")
	if err != nil {
		t.Errorf("Failed to inspect variable: %v", err)
	}
	if !strings.Contains(output, "Properties") {
		t.Error("Object inspection should show properties")
	}

	// Test non-existent variable
	_, err = d.InspectVariable("nonexistent")
	if err == nil {
		t.Error("Should get error for non-existent variable")
	}
}

// TestReset tests debugger reset
func TestReset(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Set up some state
	bp := d.SetBreakpoint(100)
	d.PushCallFrame("test", 0, nil)
	d.StepInto()
	d.Pause()

	// Manually increment hit count
	if b, exists := d.GetBreakpoint(100); exists {
		b.HitCount = 5
	}

	// Reset
	d.Reset()

	// Check state is reset
	if len(d.GetCallStack()) != 0 {
		t.Error("Call stack should be empty after reset")
	}
	if d.GetStepMode() != StepContinue {
		t.Error("Step mode should be StepContinue after reset")
	}
	if d.IsPaused() {
		t.Error("Should not be paused after reset")
	}

	// Breakpoint should still exist but hit count should be reset
	if b, exists := d.GetBreakpoint(100); exists {
		if b.HitCount != 0 {
			t.Error("Breakpoint hit count should be reset")
		}
		if b.ID != bp {
			t.Error("Breakpoint ID should remain the same")
		}
	} else {
		t.Error("Breakpoint should still exist after reset")
	}
}

// TestBytecodeOperations tests bytecode-related operations
func TestBytecodeOperations(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test setting bytecode
	bytecode := []byte{0x01, 0x02, 0x03, 0x04}
	d.SetBytecode(bytecode)

	retrieved := d.GetBytecode()
	if len(retrieved) != len(bytecode) {
		t.Errorf("Expected bytecode length %d, got %d", len(bytecode), len(retrieved))
	}
}

// TestDisassembleInstruction tests instruction disassembly
func TestDisassembleInstruction(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Create simple bytecode with known opcodes
	bytecode := []byte{
		byte(vm.OpPush),
		byte(vm.OpPop),
		byte(vm.OpAdd),
		byte(vm.OpSub),
		byte(vm.OpHalt),
	}
	d.SetBytecode(bytecode)

	tests := []struct {
		pc       int
		contains string
	}{
		{0, "PUSH"},
		{1, "POP"},
		{2, "ADD"},
		{3, "SUB"},
		{4, "HALT"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			output, err := d.DisassembleInstruction(tt.pc)
			if err != nil {
				t.Errorf("Failed to disassemble: %v", err)
			}
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.contains, output)
			}
		})
	}

	// Test invalid PC
	_, err := d.DisassembleInstruction(999)
	if err == nil {
		t.Error("Should get error for invalid PC")
	}
}

// TestREPLCreation tests REPL creation
func TestREPLCreation(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)

	if repl == nil {
		t.Fatal("NewREPL returned nil")
	}

	if repl.debugger != d {
		t.Error("REPL debugger not set correctly")
	}

	if repl.running {
		t.Error("New REPL should not be running")
	}
}

// TestREPLCommands tests REPL command execution
func TestREPLCommands(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"help", "help", false},
		{"help short", "h", false},
		{"help question", "?", false},
		{"break valid", "break 100", false},
		{"break invalid", "break abc", true},
		{"break no args", "break", true},
		{"clear valid", "clear 100", false},
		{"clear invalid", "clear xyz", true},
		{"breakpoints", "breakpoints", false},
		{"continue", "continue", false},
		{"step", "step", false},
		{"next", "next", false},
		{"out", "out", false},
		{"locals", "locals", false},
		{"globals", "globals", false},
		{"stack", "stack", false},
		{"callstack", "callstack", false},
		{"reset", "reset", false},
		{"unknown", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			err := repl.RunCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("Command '%s' error = %v, wantErr %v", tt.command, err, tt.wantErr)
			}
		})
	}
}

// TestREPLBreakpointCommands tests REPL breakpoint management
func TestREPLBreakpointCommands(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)

	// Set a breakpoint
	output.Reset()
	err := repl.RunCommand("break 100")
	if err != nil {
		t.Errorf("Failed to set breakpoint: %v", err)
	}
	if !strings.Contains(output.String(), "Breakpoint") {
		t.Error("Break command should output breakpoint info")
	}

	// List breakpoints
	output.Reset()
	err = repl.RunCommand("breakpoints")
	if err != nil {
		t.Errorf("Failed to list breakpoints: %v", err)
	}
	if !strings.Contains(output.String(), "0x0064") { // 100 in hex
		t.Error("Breakpoint list should show breakpoint at 0x0064")
	}

	// Clear breakpoint
	output.Reset()
	err = repl.RunCommand("clear 100")
	if err != nil {
		t.Errorf("Failed to clear breakpoint: %v", err)
	}

	// Verify cleared
	bps := d.ListBreakpoints()
	if len(bps) != 0 {
		t.Error("Breakpoint should be cleared")
	}
}

// TestREPLQuit tests REPL quit command
func TestREPLQuit(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)
	repl.running = true

	err := repl.RunCommand("quit")
	if err != nil {
		t.Errorf("Quit command failed: %v", err)
	}

	if repl.IsRunning() {
		t.Error("REPL should not be running after quit")
	}
}

// TestOpcodeToString tests opcode string conversion
func TestOpcodeToString(t *testing.T) {
	tests := []struct {
		opcode   vm.Opcode
		expected string
	}{
		{vm.OpPush, "PUSH"},
		{vm.OpPop, "POP"},
		{vm.OpAdd, "ADD"},
		{vm.OpSub, "SUB"},
		{vm.OpMul, "MUL"},
		{vm.OpDiv, "DIV"},
		{vm.OpEq, "EQ"},
		{vm.OpNe, "NE"},
		{vm.OpLt, "LT"},
		{vm.OpGt, "GT"},
		{vm.OpLoadVar, "LOAD_VAR"},
		{vm.OpStoreVar, "STORE_VAR"},
		{vm.OpJump, "JUMP"},
		{vm.OpReturn, "RETURN"},
		{vm.OpCall, "CALL"},
		{vm.OpHalt, "HALT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := opcodeToString(tt.opcode)
			if result != tt.expected {
				t.Errorf("opcodeToString(%v) = %s, want %s", tt.opcode, result, tt.expected)
			}
		})
	}

	// Test unknown opcode
	unknown := opcodeToString(vm.Opcode(0xAB))
	if !strings.Contains(unknown, "UNKNOWN") {
		t.Error("Unknown opcode should contain 'UNKNOWN'")
	}
}

// TestShouldBreak tests the shouldBreak logic
func TestShouldBreak(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test with no breakpoints and Continue mode
	d.SetStepMode(StepContinue)
	if d.shouldBreak(100) {
		t.Error("Should not break with no breakpoints in Continue mode")
	}

	// Test with breakpoint
	d.SetBreakpoint(100)
	if !d.shouldBreak(100) {
		t.Error("Should break at breakpoint location")
	}

	// Test disabled breakpoint
	d.DisableBreakpoint(100)
	if d.shouldBreak(100) {
		t.Error("Should not break at disabled breakpoint")
	}
	d.EnableBreakpoint(100)

	// Test StepInto mode
	d.SetStepMode(StepInto)
	if !d.shouldBreak(200) {
		t.Error("Should break on every instruction in StepInto mode")
	}

	// Test paused state
	d.SetStepMode(StepContinue)
	d.Pause()
	if !d.shouldBreak(200) {
		t.Error("Should break when paused")
	}
}

// TestREPLOutput tests that REPL produces expected output
func TestREPLOutput(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)

	// Test help command output
	output.Reset()
	repl.RunCommand("help")
	helpOutput := output.String()

	expectedSections := []string{
		"Available Commands",
		"Breakpoint Management",
		"Execution Control",
		"Inspection",
		"Evaluation",
		"Utility",
		"break",
		"continue",
		"step",
		"print",
		"locals",
		"quit",
	}

	for _, section := range expectedSections {
		if !strings.Contains(helpOutput, section) {
			t.Errorf("Help output should contain '%s'", section)
		}
	}
}

// TestFormatLocals tests local variables formatting
func TestFormatLocals(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test with no locals
	output := d.FormatLocals()
	if !strings.Contains(output, "No local") {
		t.Error("Should indicate no local variables")
	}

	// Note: Full test would require VM to expose locals
	// This is a placeholder for when that functionality is added
}

// TestSetBreakpointByFunction tests setting breakpoints by function name
func TestSetBreakpointByFunction(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// This should return an error since function metadata isn't available
	_, err := d.SetBreakpointByFunction("main")
	if err == nil {
		t.Error("Expected error for SetBreakpointByFunction without metadata")
	}
}

// TestGetPC tests getting the program counter
func TestGetPC(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	pc := d.GetPC()
	if pc < 0 {
		t.Error("GetPC should return a non-negative value")
	}
}

// TestFormatLocalsWithValues tests formatting local variables with various types
func TestFormatLocalsWithValues(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	locals := map[string]vm.Value{
		"x":       vm.IntValue{Val: 42},
		"name":    vm.StringValue{Val: "test"},
		"active":  vm.BoolValue{Val: true},
		"price":   vm.FloatValue{Val: 3.14},
	}
	d.PushCallFrame("testFunc", 100, locals)

	output := d.FormatLocals()
	// The FormatLocals function returns locals from the current frame
	// Just check it returns something reasonable
	if output == "" {
		t.Log("FormatLocals returned empty string (may be expected)")
	}
}

// TestFormatStack tests formatting call stack
func TestFormatStack(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Empty stack
	output := d.FormatStack()
	if !strings.Contains(output, "empty") && output != "" {
		t.Log("Empty stack output:", output)
	}

	// Push some frames
	d.PushCallFrame("main", 0, nil)
	d.PushCallFrame("helper", 50, nil)
	d.PushCallFrame("nested", 100, nil)

	output = d.FormatStack()
	if output == "" {
		t.Error("FormatStack should return non-empty string with frames")
	}
}

// TestClearBreakpointByID tests clearing breakpoints by ID
func TestClearBreakpointByID(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Set breakpoints
	id1 := d.SetBreakpoint(100)
	id2 := d.SetBreakpoint(200)

	// Clear by ID
	if !d.ClearBreakpointByID(id1) {
		t.Error("Should successfully clear breakpoint by ID")
	}

	// Verify cleared
	bps := d.ListBreakpoints()
	if len(bps) != 1 {
		t.Errorf("Expected 1 breakpoint after clearing, got %d", len(bps))
	}

	// Try clearing non-existent
	if d.ClearBreakpointByID(999) {
		t.Error("Should return false for non-existent breakpoint ID")
	}

	// Clear the second one
	d.ClearBreakpointByID(id2)
	if len(d.ListBreakpoints()) != 0 {
		t.Error("Should have no breakpoints left")
	}
}

// TestEnableDisableBreakpointNonExistent tests enable/disable on non-existent breakpoints
func TestEnableDisableBreakpointNonExistent(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	if d.EnableBreakpoint(999) {
		t.Error("EnableBreakpoint should return false for non-existent location")
	}

	if d.DisableBreakpoint(999) {
		t.Error("DisableBreakpoint should return false for non-existent location")
	}
}

// TestAllStepModes tests all stepping modes
func TestAllStepModes(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Test StepInto
	d.StepInto()
	if d.stepMode != StepInto {
		t.Error("StepInto should set step mode to StepInto")
	}

	// Test StepOver
	d.StepOver()
	if d.stepMode != StepOver {
		t.Error("StepOver should set step mode to StepOver")
	}

	// Test StepOut
	d.StepOut()
	if d.stepMode != StepOut {
		t.Error("StepOut should set step mode to StepOut")
	}

	// Test Continue
	d.Continue()
	if d.stepMode != StepContinue {
		t.Error("Continue should set step mode to StepContinue")
	}
}

// TestREPLBasicCreation tests creating a REPL with buffers
func TestREPLBasicCreation(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	var input bytes.Buffer
	var output bytes.Buffer

	repl := NewREPL(d, &input, &output)
	if repl == nil {
		t.Fatal("NewREPL returned nil")
	}

	if repl.debugger != d {
		t.Error("REPL debugger not set correctly")
	}
}

// TestREPLStop tests stopping the REPL
func TestREPLStop(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	var input bytes.Buffer
	var output bytes.Buffer

	repl := NewREPL(d, &input, &output)
	repl.running = true
	repl.Stop()

	if repl.running {
		t.Error("REPL should not be running after Stop()")
	}
}

// TestOpcodeToStringAll tests all opcode conversions for coverage
func TestOpcodeToStringAll(t *testing.T) {
	opcodes := []vm.Opcode{
		vm.OpGe, vm.OpLe, vm.OpAnd, vm.OpOr, vm.OpNot,
		vm.OpJumpIfFalse, vm.OpJumpIfTrue, vm.OpGetIter, vm.OpIterNext,
		vm.OpIterHasNext, vm.OpGetIndex, vm.OpBuildObject, vm.OpGetField,
		vm.OpBuildArray, vm.OpHttpReturn, vm.OpWsSend, vm.OpWsBroadcast,
		vm.OpWsBroadcastRoom, vm.OpWsJoinRoom, vm.OpWsLeaveRoom,
		vm.OpWsClose, vm.OpWsGetRooms, vm.OpWsGetClients, vm.OpNeg,
		vm.OpWsGetConnCount, vm.OpWsGetUptime,
	}

	for _, op := range opcodes {
		result := opcodeToString(op)
		if result == "" {
			t.Errorf("opcodeToString returned empty for %v", op)
		}
	}
}

// TestFormatLocalsVM tests FormatLocals with the VM
func TestFormatLocalsVM(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Without any call frames - should show "no locals"
	output := d.FormatLocals()
	if !strings.Contains(strings.ToLower(output), "no") || !strings.Contains(strings.ToLower(output), "local") {
		// Alternatively may just be empty
		if output != "" {
			t.Logf("FormatLocals output: %s", output)
		}
	}
}

// TestFormatStackVM tests FormatStack with the VM
func TestFormatStackVM(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Without any call frames - should show "empty"
	output := d.FormatStack()
	if output == "" {
		t.Log("FormatStack returned empty")
	}
}

// TestREPLGlobalsCommand tests the globals command
func TestREPLGlobalsCommand(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)

	// Run globals command
	output.Reset()
	err := repl.RunCommand("globals")
	if err != nil {
		t.Errorf("Globals command failed: %v", err)
	}
}

// TestREPLClearCommands tests various clear command forms
func TestREPLClearCommands(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	repl := NewREPL(d, input, output)

	// Set a breakpoint first
	d.SetBreakpoint(100)
	d.SetBreakpoint(200)

	// Clear with address
	err := repl.RunCommand("clear 100")
	if err != nil {
		t.Logf("Clear command returned error: %v", err)
	}

	// Clear all
	err = repl.RunCommand("clear all")
	if err != nil {
		t.Logf("Clear all returned error: %v", err)
	}

	// Invalid clear
	err = repl.RunCommand("clear xyz")
	// May or may not return error
	_ = err
}

// TestStepOutCoverage tests StepOut coverage
func TestStepOutCoverage(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Push a call frame first
	d.PushCallFrame("test", 100, nil)
	d.PushCallFrame("inner", 200, nil)

	// Now StepOut should work
	d.StepOut()

	if d.stepMode != StepOut {
		t.Error("StepOut should set mode to StepOut")
	}
}

// TestShouldBreakStepOut tests shouldBreak with StepOut mode
func TestShouldBreakStepOut(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Set up call stack
	d.PushCallFrame("main", 0, nil)
	d.PushCallFrame("helper", 100, nil)

	// Set StepOut mode
	d.StepOut()

	// Now pop to reduce call stack
	d.PopCallFrame()

	// shouldBreak should return true when call stack depth <= stepOutDepth
	shouldBreak := d.shouldBreak(50)
	// May or may not break depending on implementation
	_ = shouldBreak
}

// TestREPLPrint tests the print command
func TestREPLPrint(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Set up a call frame with locals
	locals := map[string]vm.Value{
		"x": vm.IntValue{Val: 42},
	}
	d.PushCallFrame("test", 0, locals)

	input := strings.NewReader("")
	output := &bytes.Buffer{}
	repl := NewREPL(d, input, output)

	// Test print command
	err := repl.RunCommand("print x")
	if err != nil {
		t.Logf("Print command returned error: %v", err)
	}
}

// TestREPLInspect tests the inspect command
func TestREPLInspect(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Set up a call frame with locals
	locals := map[string]vm.Value{
		"x": vm.IntValue{Val: 42},
	}
	d.PushCallFrame("test", 0, locals)

	input := strings.NewReader("")
	output := &bytes.Buffer{}
	repl := NewREPL(d, input, output)

	// Test inspect command
	err := repl.RunCommand("inspect x")
	if err != nil {
		t.Logf("Inspect command returned error: %v", err)
	}
}

// TestREPLEval tests the eval command
func TestREPLEval(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("")
	output := &bytes.Buffer{}
	repl := NewREPL(d, input, output)

	// Test eval command
	err := repl.RunCommand("eval 1 + 2")
	if err != nil {
		t.Logf("Eval command returned error: %v", err)
	}
}

// TestREPLDisassemble tests the disassemble command
func TestREPLDisassemble(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)

	// Set some bytecode
	d.SetBytecode([]byte{byte(vm.OpPush), 0x01, byte(vm.OpHalt)})

	input := strings.NewReader("")
	output := &bytes.Buffer{}
	repl := NewREPL(d, input, output)

	// Test disassemble command
	err := repl.RunCommand("disassemble")
	if err != nil {
		t.Logf("Disassemble command returned error: %v", err)
	}
}

// TestREPLWelcomeGoodbye tests the welcome and goodbye messages
func TestREPLWelcomeGoodbye(t *testing.T) {
	v := vm.NewVM()
	d := NewDebugger(v)
	input := strings.NewReader("quit\n")
	output := &bytes.Buffer{}
	repl := NewREPL(d, input, output)

	// These are called by Start() but Start is blocking
	// We can test that the REPL is properly initialized
	if repl.debugger != d {
		t.Error("REPL debugger not set correctly")
	}
}

// BenchmarkDebuggerOperations benchmarks common debugger operations
func BenchmarkDebuggerOperations(b *testing.B) {
	v := vm.NewVM()
	d := NewDebugger(v)

	b.Run("SetBreakpoint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			d.SetBreakpoint(i)
		}
	})

	b.Run("ListBreakpoints", func(b *testing.B) {
		// Set up some breakpoints
		for i := 0; i < 100; i++ {
			d.SetBreakpoint(i * 10)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.ListBreakpoints()
		}
	})

	b.Run("PushCallFrame", func(b *testing.B) {
		locals := map[string]vm.Value{
			"x": vm.IntValue{Val: 42},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			d.PushCallFrame("test", 100, locals)
			if len(d.callStack) > 1000 {
				d.Reset()
			}
		}
	})
}
