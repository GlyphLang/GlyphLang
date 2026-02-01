package debug

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// REPL provides an interactive debugging interface
type REPL struct {
	debugger *Debugger
	reader   *bufio.Reader
	writer   io.Writer
	running  bool
}

// NewREPL creates a new REPL instance
func NewREPL(debugger *Debugger, reader io.Reader, writer io.Writer) *REPL {
	return &REPL{
		debugger: debugger,
		reader:   bufio.NewReader(reader),
		writer:   writer,
		running:  false,
	}
}

// Start starts the REPL loop
func (r *REPL) Start() {
	r.running = true
	r.printWelcome()

	for r.running {
		r.printPrompt()
		line, err := r.readLine()
		if err != nil {
			if err == io.EOF {
				r.running = false
				break
			}
			r.printf("Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := r.executeCommand(line); err != nil {
			r.printf("Error: %v\n", err)
		}
	}

	r.printGoodbye()
}

// Stop stops the REPL loop
func (r *REPL) Stop() {
	r.running = false
}

// printWelcome prints the welcome message
func (r *REPL) printWelcome() {
	r.printf("Glyph Debugger REPL\n")
	r.printf("Type 'help' for available commands\n")
	r.printf("=====================================\n\n")
}

// printGoodbye prints the goodbye message
func (r *REPL) printGoodbye() {
	r.printf("\nGoodbye!\n")
}

// printPrompt prints the command prompt
func (r *REPL) printPrompt() {
	status := "running"
	if r.debugger.IsPaused() {
		status = "paused"
	}
	r.printf("(glyph-debug:%s) ", status)
}

// readLine reads a line of input
func (r *REPL) readLine() (string, error) {
	return r.reader.ReadString('\n')
}

// printf writes formatted output
func (r *REPL) printf(format string, args ...interface{}) {
	fmt.Fprintf(r.writer, format, args...)
}

// executeCommand executes a REPL command
func (r *REPL) executeCommand(line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "help", "h", "?":
		return r.cmdHelp(args)
	case "break", "b":
		return r.cmdBreak(args)
	case "clear", "cl":
		return r.cmdClear(args)
	case "breakpoints", "bp":
		return r.cmdBreakpoints(args)
	case "continue", "c":
		return r.cmdContinue(args)
	case "step", "s":
		return r.cmdStep(args)
	case "next", "n":
		return r.cmdNext(args)
	case "out", "o":
		return r.cmdOut(args)
	case "print", "p":
		return r.cmdPrint(args)
	case "locals", "l":
		return r.cmdLocals(args)
	case "globals", "g":
		return r.cmdGlobals(args)
	case "stack", "st":
		return r.cmdStack(args)
	case "callstack", "cs", "backtrace", "bt":
		return r.cmdCallStack(args)
	case "inspect", "i":
		return r.cmdInspect(args)
	case "eval", "e":
		return r.cmdEval(args)
	case "disassemble", "disasm", "d":
		return r.cmdDisassemble(args)
	case "reset", "r":
		return r.cmdReset(args)
	case "quit", "q", "exit":
		return r.cmdQuit(args)
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", command)
	}
}

// cmdHelp displays help information
func (r *REPL) cmdHelp(args []string) error {
	r.printf("Available Commands:\n")
	r.printf("==================\n\n")
	r.printf("Breakpoint Management:\n")
	r.printf("  break, b <location>     - Set breakpoint at bytecode location\n")
	r.printf("  clear, cl <location>    - Clear breakpoint at location\n")
	r.printf("  breakpoints, bp         - List all breakpoints\n\n")
	r.printf("Execution Control:\n")
	r.printf("  continue, c             - Continue execution until next breakpoint\n")
	r.printf("  step, s                 - Step into next instruction\n")
	r.printf("  next, n                 - Step over (don't enter function calls)\n")
	r.printf("  out, o                  - Step out of current function\n\n")
	r.printf("Inspection:\n")
	r.printf("  print, p <var>          - Print variable value\n")
	r.printf("  locals, l               - Show local variables\n")
	r.printf("  globals, g              - Show global variables\n")
	r.printf("  stack, st               - Show value stack\n")
	r.printf("  callstack, cs, bt       - Show call stack (backtrace)\n")
	r.printf("  inspect, i <var>        - Detailed variable inspection\n\n")
	r.printf("Evaluation:\n")
	r.printf("  eval, e <expr>          - Evaluate expression in current context\n")
	r.printf("  disassemble, d [pc]     - Disassemble instruction at PC\n\n")
	r.printf("Utility:\n")
	r.printf("  reset, r                - Reset debugger state\n")
	r.printf("  help, h, ?              - Show this help message\n")
	r.printf("  quit, q, exit           - Exit debugger\n\n")
	return nil
}

// cmdBreak sets a breakpoint
func (r *REPL) cmdBreak(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: break <location>")
	}

	location, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid location: %s", args[0])
	}

	id := r.debugger.SetBreakpoint(location)
	r.printf("Breakpoint %d set at location 0x%04x\n", id, location)
	return nil
}

// cmdClear clears a breakpoint
func (r *REPL) cmdClear(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: clear <location>")
	}

	location, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid location: %s", args[0])
	}

	if r.debugger.ClearBreakpoint(location) {
		r.printf("Breakpoint at location 0x%04x cleared\n", location)
	} else {
		r.printf("No breakpoint at location 0x%04x\n", location)
	}
	return nil
}

// cmdBreakpoints lists all breakpoints
func (r *REPL) cmdBreakpoints(args []string) error {
	bps := r.debugger.ListBreakpoints()
	if len(bps) == 0 {
		r.printf("No breakpoints set\n")
		return nil
	}

	r.printf("Breakpoints:\n")
	for _, bp := range bps {
		status := "enabled"
		if !bp.Enabled {
			status = "disabled"
		}
		r.printf("  #%d: 0x%04x (%s, hit %d times)\n", bp.ID, bp.Location, status, bp.HitCount)
	}
	return nil
}

// cmdContinue continues execution
func (r *REPL) cmdContinue(args []string) error {
	r.debugger.Continue()
	r.printf("Continuing execution...\n")
	return nil
}

// cmdStep steps into next instruction
func (r *REPL) cmdStep(args []string) error {
	r.debugger.StepInto()
	r.printf("Stepping into next instruction...\n")
	return nil
}

// cmdNext steps over (doesn't enter function calls)
func (r *REPL) cmdNext(args []string) error {
	r.debugger.StepOver()
	r.printf("Stepping over...\n")
	return nil
}

// cmdOut steps out of current function
func (r *REPL) cmdOut(args []string) error {
	r.debugger.StepOut()
	r.printf("Stepping out...\n")
	return nil
}

// cmdPrint prints a variable value
func (r *REPL) cmdPrint(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: print <variable>")
	}

	varName := args[0]
	val, err := r.debugger.GetVariable(varName)
	if err != nil {
		return err
	}

	r.printf("%s = %s\n", varName, r.debugger.formatValue(val))
	return nil
}

// cmdLocals shows local variables
func (r *REPL) cmdLocals(args []string) error {
	r.printf("%s\n", r.debugger.FormatLocals())
	return nil
}

// cmdGlobals shows global variables
func (r *REPL) cmdGlobals(args []string) error {
	globals := r.debugger.GetGlobals()
	if len(globals) == 0 {
		r.printf("No global variables\n")
		return nil
	}

	r.printf("Global Variables:\n")
	for name, val := range globals {
		r.printf("  %s = %s\n", name, r.debugger.formatValue(val))
	}
	return nil
}

// cmdStack shows the value stack
func (r *REPL) cmdStack(args []string) error {
	r.printf("%s\n", r.debugger.FormatStack())
	return nil
}

// cmdCallStack shows the call stack
func (r *REPL) cmdCallStack(args []string) error {
	r.printf("%s\n", r.debugger.FormatCallStack())
	return nil
}

// cmdInspect provides detailed variable inspection
func (r *REPL) cmdInspect(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: inspect <variable>")
	}

	varName := args[0]
	info, err := r.debugger.InspectVariable(varName)
	if err != nil {
		return err
	}

	r.printf("%s", info)
	return nil
}

// cmdEval evaluates an expression in the current context
func (r *REPL) cmdEval(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: eval <expression>")
	}

	exprStr := strings.Join(args, " ")

	// For now, expression evaluation requires full parsing support
	// This is a placeholder that would need integration with the interpreter
	// when VM exposes its state properly

	// Simple evaluation: try to parse as a route and extract the expression
	// This is a simplified version - full support would require exposing
	// the parseExpr method from the parser or creating a separate expression parser

	r.printf("Expression evaluation not yet fully implemented\n")
	r.printf("Expression: %s\n", exprStr)
	r.printf("Note: This feature requires VM state exposure and expression-only parsing\n")

	// Expression evaluation requires:
	// 1. VM exposing locals/globals for state inspection
	// 2. A public parseExpr method or standalone expression parser
	// 3. An interpreter that can evaluate with the VM's current state

	return nil
}

// cmdDisassemble disassembles instruction at PC
func (r *REPL) cmdDisassemble(args []string) error {
	var pc int
	var err error

	if len(args) == 0 {
		// Use current PC
		pc = r.debugger.GetPC()
	} else {
		pc, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PC: %s", args[0])
		}
	}

	instr, err := r.debugger.DisassembleInstruction(pc)
	if err != nil {
		return err
	}

	r.printf("%s\n", instr)

	// Show surrounding instructions for context
	r.printf("\nContext:\n")
	for i := pc - 5; i <= pc+5; i++ {
		if i < 0 || i >= len(r.debugger.GetBytecode()) {
			continue
		}
		marker := "  "
		if i == pc {
			marker = "=>"
		}
		if instr, err := r.debugger.DisassembleInstruction(i); err == nil {
			r.printf("%s %s\n", marker, instr)
		}
	}

	return nil
}

// cmdReset resets the debugger state
func (r *REPL) cmdReset(args []string) error {
	r.debugger.Reset()
	r.printf("Debugger state reset\n")
	return nil
}

// cmdQuit exits the debugger
func (r *REPL) cmdQuit(args []string) error {
	r.running = false
	return nil
}

// RunCommand executes a single command (useful for programmatic control)
func (r *REPL) RunCommand(command string) error {
	return r.executeCommand(command)
}

// IsRunning returns whether the REPL is currently running
func (r *REPL) IsRunning() bool {
	return r.running
}
