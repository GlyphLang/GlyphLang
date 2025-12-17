package debug

import (
	"fmt"
	"strings"

	"github.com/glyphlang/glyph/pkg/vm"
)

// StepMode defines how the debugger should step through execution
type StepMode int

const (
	StepContinue StepMode = iota // Run until next breakpoint
	StepInto                      // Step into function calls
	StepOver                      // Step over function calls
	StepOut                       // Step out of current function
)

// Breakpoint represents a breakpoint in the code
type Breakpoint struct {
	ID       int
	Location int    // Bytecode offset (PC value)
	FuncName string // Optional: function name for breakpoints by function
	Enabled  bool
	HitCount int
}

// CallFrame represents a single function call on the call stack
type CallFrame struct {
	FuncName   string
	ReturnAddr int              // Return address (PC to return to)
	Locals     map[string]vm.Value // Local variables at this frame
}

// Debugger wraps VM execution with debugging capabilities
type Debugger struct {
	vm          *vm.VM
	bytecode    []byte
	breakpoints map[int]*Breakpoint
	nextBPID    int
	stepMode    StepMode
	callStack   []CallFrame
	paused      bool
	stepDepth   int // Track call depth for step over/out
}

// NewDebugger creates a new debugger instance
func NewDebugger(v *vm.VM) *Debugger {
	return &Debugger{
		vm:          v,
		breakpoints: make(map[int]*Breakpoint),
		nextBPID:    1,
		stepMode:    StepContinue,
		callStack:   make([]CallFrame, 0),
		paused:      false,
		stepDepth:   0,
	}
}

// SetBreakpoint sets a breakpoint at a specific bytecode location
func (d *Debugger) SetBreakpoint(location int) int {
	bp := &Breakpoint{
		ID:       d.nextBPID,
		Location: location,
		Enabled:  true,
		HitCount: 0,
	}
	d.breakpoints[location] = bp
	d.nextBPID++
	return bp.ID
}

// SetBreakpointByFunction sets a breakpoint by function name
// Note: This requires function metadata which may not be available in bytecode
// For now, this is a placeholder that would need additional compiler support
func (d *Debugger) SetBreakpointByFunction(funcName string) (int, error) {
	// This would require function symbol table from compiler
	// For now, return an error indicating it's not supported
	return 0, fmt.Errorf("breakpoint by function name not yet supported: need function metadata")
}

// ClearBreakpoint removes a breakpoint by location
func (d *Debugger) ClearBreakpoint(location int) bool {
	if _, exists := d.breakpoints[location]; exists {
		delete(d.breakpoints, location)
		return true
	}
	return false
}

// ClearBreakpointByID removes a breakpoint by ID
func (d *Debugger) ClearBreakpointByID(id int) bool {
	for loc, bp := range d.breakpoints {
		if bp.ID == id {
			delete(d.breakpoints, loc)
			return true
		}
	}
	return false
}

// EnableBreakpoint enables a breakpoint
func (d *Debugger) EnableBreakpoint(location int) bool {
	if bp, exists := d.breakpoints[location]; exists {
		bp.Enabled = true
		return true
	}
	return false
}

// DisableBreakpoint disables a breakpoint without removing it
func (d *Debugger) DisableBreakpoint(location int) bool {
	if bp, exists := d.breakpoints[location]; exists {
		bp.Enabled = false
		return true
	}
	return false
}

// ListBreakpoints returns all breakpoints
func (d *Debugger) ListBreakpoints() []*Breakpoint {
	bps := make([]*Breakpoint, 0, len(d.breakpoints))
	for _, bp := range d.breakpoints {
		bps = append(bps, bp)
	}
	return bps
}

// GetBreakpoint retrieves a breakpoint by location
func (d *Debugger) GetBreakpoint(location int) (*Breakpoint, bool) {
	bp, exists := d.breakpoints[location]
	return bp, exists
}

// SetStepMode sets the stepping mode for execution
func (d *Debugger) SetStepMode(mode StepMode) {
	d.stepMode = mode
}

// GetStepMode returns the current stepping mode
func (d *Debugger) GetStepMode() StepMode {
	return d.stepMode
}

// Continue resumes execution until next breakpoint
func (d *Debugger) Continue() {
	d.stepMode = StepContinue
	d.paused = false
}

// StepInto steps into the next instruction (including function calls)
func (d *Debugger) StepInto() {
	d.stepMode = StepInto
	d.paused = false
}

// StepOver steps over function calls
func (d *Debugger) StepOver() {
	d.stepMode = StepOver
	d.stepDepth = len(d.callStack)
	d.paused = false
}

// StepOut steps out of the current function
func (d *Debugger) StepOut() {
	d.stepMode = StepOut
	if len(d.callStack) > 0 {
		d.stepDepth = len(d.callStack) - 1
	}
	d.paused = false
}

// IsPaused returns whether the debugger is currently paused
func (d *Debugger) IsPaused() bool {
	return d.paused
}

// Pause pauses execution at the next instruction
func (d *Debugger) Pause() {
	d.paused = true
}

// GetPC returns the current program counter
func (d *Debugger) GetPC() int {
	// We need to expose this through the VM
	// For now, this is a limitation - we'd need to modify VM to expose PC
	return 0 // Placeholder
}

// GetLocals returns the current local variables
func (d *Debugger) GetLocals() map[string]vm.Value {
	// Return copy to prevent modification
	locals := make(map[string]vm.Value)
	// This would need VM to expose its locals map
	// For now, return empty map
	return locals
}

// GetGlobals returns the current global variables
func (d *Debugger) GetGlobals() map[string]vm.Value {
	// Return copy to prevent modification
	globals := make(map[string]vm.Value)
	// This would need VM to expose its globals map
	// For now, return empty map
	return globals
}

// GetVariable retrieves a variable by name from locals or globals
func (d *Debugger) GetVariable(name string) (vm.Value, error) {
	// Check current frame locals
	if len(d.callStack) > 0 {
		frame := d.callStack[len(d.callStack)-1]
		if val, exists := frame.Locals[name]; exists {
			return val, nil
		}
	}

	// Check VM locals (would need VM API)
	// Check VM globals (would need VM API)

	return nil, fmt.Errorf("variable not found: %s", name)
}

// GetStack returns the current value stack
func (d *Debugger) GetStack() []vm.Value {
	// This would need VM to expose its stack
	// For now, return empty slice
	return []vm.Value{}
}

// GetCallStack returns the current call stack
func (d *Debugger) GetCallStack() []CallFrame {
	// Return copy to prevent modification
	stack := make([]CallFrame, len(d.callStack))
	copy(stack, d.callStack)
	return stack
}

// FormatCallStack returns a formatted string representation of the call stack
func (d *Debugger) FormatCallStack() string {
	if len(d.callStack) == 0 {
		return "Call stack is empty"
	}

	var sb strings.Builder
	sb.WriteString("Call Stack:\n")
	for i := len(d.callStack) - 1; i >= 0; i-- {
		frame := d.callStack[i]
		sb.WriteString(fmt.Sprintf("  #%d %s (return addr: %d)\n", len(d.callStack)-1-i, frame.FuncName, frame.ReturnAddr))
	}
	return sb.String()
}

// FormatLocals returns a formatted string representation of local variables
func (d *Debugger) FormatLocals() string {
	locals := d.GetLocals()
	if len(locals) == 0 {
		return "No local variables"
	}

	var sb strings.Builder
	sb.WriteString("Local Variables:\n")
	for name, val := range locals {
		sb.WriteString(fmt.Sprintf("  %s = %s\n", name, d.formatValue(val)))
	}
	return sb.String()
}

// FormatStack returns a formatted string representation of the value stack
func (d *Debugger) FormatStack() string {
	stack := d.GetStack()
	if len(stack) == 0 {
		return "Stack is empty"
	}

	var sb strings.Builder
	sb.WriteString("Value Stack:\n")
	for i := len(stack) - 1; i >= 0; i-- {
		sb.WriteString(fmt.Sprintf("  [%d] %s\n", i, d.formatValue(stack[i])))
	}
	return sb.String()
}

// formatValue formats a VM value for display
func (d *Debugger) formatValue(val vm.Value) string {
	switch v := val.(type) {
	case vm.IntValue:
		return fmt.Sprintf("%d (int)", v.Val)
	case vm.FloatValue:
		return fmt.Sprintf("%f (float)", v.Val)
	case vm.StringValue:
		return fmt.Sprintf("\"%s\" (string)", v.Val)
	case vm.BoolValue:
		return fmt.Sprintf("%t (bool)", v.Val)
	case vm.NullValue:
		return "null"
	case vm.ArrayValue:
		return fmt.Sprintf("[%d elements] (array)", len(v.Val))
	case vm.ObjectValue:
		return fmt.Sprintf("{%d fields} (object)", len(v.Val))
	default:
		return fmt.Sprintf("%v (%s)", val, val.Type())
	}
}

// InspectVariable provides detailed information about a variable
func (d *Debugger) InspectVariable(name string) (string, error) {
	val, err := d.GetVariable(name)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Variable: %s\n", name))
	sb.WriteString(fmt.Sprintf("Type: %s\n", val.Type()))

	switch v := val.(type) {
	case vm.IntValue:
		sb.WriteString(fmt.Sprintf("Value: %d\n", v.Val))
	case vm.FloatValue:
		sb.WriteString(fmt.Sprintf("Value: %f\n", v.Val))
	case vm.StringValue:
		sb.WriteString(fmt.Sprintf("Value: \"%s\"\n", v.Val))
		sb.WriteString(fmt.Sprintf("Length: %d\n", len(v.Val)))
	case vm.BoolValue:
		sb.WriteString(fmt.Sprintf("Value: %t\n", v.Val))
	case vm.NullValue:
		sb.WriteString("Value: null\n")
	case vm.ArrayValue:
		sb.WriteString(fmt.Sprintf("Length: %d\n", len(v.Val)))
		sb.WriteString("Elements:\n")
		for i, elem := range v.Val {
			sb.WriteString(fmt.Sprintf("  [%d] = %s\n", i, d.formatValue(elem)))
		}
	case vm.ObjectValue:
		sb.WriteString(fmt.Sprintf("Fields: %d\n", len(v.Val)))
		sb.WriteString("Properties:\n")
		for k, elem := range v.Val {
			sb.WriteString(fmt.Sprintf("  %s = %s\n", k, d.formatValue(elem)))
		}
	}

	return sb.String(), nil
}

// shouldBreak determines if execution should break at the current PC
func (d *Debugger) shouldBreak(pc int) bool {
	// Check if paused
	if d.paused {
		return true
	}

	// Check breakpoints
	if bp, exists := d.breakpoints[pc]; exists && bp.Enabled {
		bp.HitCount++
		return true
	}

	// Check step modes
	switch d.stepMode {
	case StepInto:
		return true // Break on every instruction
	case StepOver:
		// Break when we're back at the same depth or shallower
		return len(d.callStack) <= d.stepDepth
	case StepOut:
		// Break when we're back at a shallower depth
		return len(d.callStack) <= d.stepDepth
	case StepContinue:
		return false // Only break on breakpoints
	}

	return false
}

// PushCallFrame adds a new call frame to the stack
func (d *Debugger) PushCallFrame(funcName string, returnAddr int, locals map[string]vm.Value) {
	frame := CallFrame{
		FuncName:   funcName,
		ReturnAddr: returnAddr,
		Locals:     locals,
	}
	d.callStack = append(d.callStack, frame)
}

// PopCallFrame removes the top call frame from the stack
func (d *Debugger) PopCallFrame() (CallFrame, error) {
	if len(d.callStack) == 0 {
		return CallFrame{}, fmt.Errorf("call stack is empty")
	}

	frame := d.callStack[len(d.callStack)-1]
	d.callStack = d.callStack[:len(d.callStack)-1]
	return frame, nil
}

// Reset resets the debugger state
func (d *Debugger) Reset() {
	d.callStack = make([]CallFrame, 0)
	d.stepMode = StepContinue
	d.paused = false
	d.stepDepth = 0
	// Reset breakpoint hit counts
	for _, bp := range d.breakpoints {
		bp.HitCount = 0
	}
}

// SetBytecode sets the bytecode being debugged
func (d *Debugger) SetBytecode(bytecode []byte) {
	d.bytecode = bytecode
}

// GetBytecode returns the current bytecode
func (d *Debugger) GetBytecode() []byte {
	return d.bytecode
}

// DisassembleInstruction disassembles a single instruction at the given PC
// This is a helper for debugging and displaying what instruction will execute
func (d *Debugger) DisassembleInstruction(pc int) (string, error) {
	if d.bytecode == nil || pc >= len(d.bytecode) {
		return "", fmt.Errorf("invalid PC or no bytecode loaded")
	}

	opcode := vm.Opcode(d.bytecode[pc])
	return fmt.Sprintf("0x%04x: %s (0x%02x)", pc, opcodeToString(opcode), byte(opcode)), nil
}

// opcodeToString converts an opcode to a readable string
func opcodeToString(opcode vm.Opcode) string {
	switch opcode {
	case vm.OpPush:
		return "PUSH"
	case vm.OpPop:
		return "POP"
	case vm.OpAdd:
		return "ADD"
	case vm.OpSub:
		return "SUB"
	case vm.OpMul:
		return "MUL"
	case vm.OpDiv:
		return "DIV"
	case vm.OpEq:
		return "EQ"
	case vm.OpNe:
		return "NE"
	case vm.OpLt:
		return "LT"
	case vm.OpGt:
		return "GT"
	case vm.OpGe:
		return "GE"
	case vm.OpLe:
		return "LE"
	case vm.OpAnd:
		return "AND"
	case vm.OpOr:
		return "OR"
	case vm.OpNot:
		return "NOT"
	case vm.OpLoadVar:
		return "LOAD_VAR"
	case vm.OpStoreVar:
		return "STORE_VAR"
	case vm.OpJump:
		return "JUMP"
	case vm.OpJumpIfFalse:
		return "JUMP_IF_FALSE"
	case vm.OpJumpIfTrue:
		return "JUMP_IF_TRUE"
	case vm.OpGetIter:
		return "GET_ITER"
	case vm.OpIterNext:
		return "ITER_NEXT"
	case vm.OpIterHasNext:
		return "ITER_HAS_NEXT"
	case vm.OpGetIndex:
		return "GET_INDEX"
	case vm.OpReturn:
		return "RETURN"
	case vm.OpCall:
		return "CALL"
	case vm.OpBuildObject:
		return "BUILD_OBJECT"
	case vm.OpGetField:
		return "GET_FIELD"
	case vm.OpBuildArray:
		return "BUILD_ARRAY"
	case vm.OpHttpReturn:
		return "HTTP_RETURN"
	case vm.OpWsSend:
		return "WS_SEND"
	case vm.OpWsBroadcast:
		return "WS_BROADCAST"
	case vm.OpWsBroadcastRoom:
		return "WS_BROADCAST_ROOM"
	case vm.OpWsJoinRoom:
		return "WS_JOIN_ROOM"
	case vm.OpWsLeaveRoom:
		return "WS_LEAVE_ROOM"
	case vm.OpWsClose:
		return "WS_CLOSE"
	case vm.OpWsGetRooms:
		return "WS_GET_ROOMS"
	case vm.OpWsGetClients:
		return "WS_GET_CLIENTS"
	case vm.OpHalt:
		return "HALT"
	default:
		return fmt.Sprintf("UNKNOWN(0x%02x)", byte(opcode))
	}
}
