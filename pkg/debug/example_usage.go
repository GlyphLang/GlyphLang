package debug

// Example usage of the Glyph Debugger
//
// This file demonstrates how to use the debug package to debug Glyph bytecode execution.

/*
Basic Usage:

	import (
		"github.com/glyphlang/glyph/pkg/debug"
		"github.com/glyphlang/glyph/pkg/vm"
		"os"
	)

	// Create VM and debugger
	v := vm.NewVM()
	d := debug.NewDebugger(v)

	// Set breakpoints
	d.SetBreakpoint(100)  // Break at bytecode offset 100
	d.SetBreakpoint(200)  // Break at bytecode offset 200

	// List breakpoints
	bps := d.ListBreakpoints()
	for _, bp := range bps {
		fmt.Printf("Breakpoint #%d at 0x%04x\n", bp.ID, bp.Location)
	}

	// Control execution
	d.Continue()        // Run until next breakpoint
	d.StepInto()        // Step into next instruction
	d.StepOver()        // Step over function calls
	d.StepOut()         // Step out of current function

	// Inspect state
	locals := d.GetLocals()
	globals := d.GetGlobals()
	stack := d.GetStack()
	callStack := d.GetCallStack()

	// Formatted output
	fmt.Println(d.FormatLocals())
	fmt.Println(d.FormatStack())
	fmt.Println(d.FormatCallStack())

	// Variable inspection
	if val, err := d.GetVariable("x"); err == nil {
		fmt.Printf("x = %v\n", val)
	}

	// Detailed inspection
	if info, err := d.InspectVariable("myArray"); err == nil {
		fmt.Println(info)
	}

	// Disassemble instructions
	if instr, err := d.DisassembleInstruction(100); err == nil {
		fmt.Println(instr)
	}

Interactive REPL Usage:

	import (
		"github.com/glyphlang/glyph/pkg/debug"
		"github.com/glyphlang/glyph/pkg/vm"
		"os"
	)

	v := vm.NewVM()
	d := debug.NewDebugger(v)

	// Create REPL with stdin/stdout
	repl := debug.NewREPL(d, os.Stdin, os.Stdout)

	// Start interactive session
	repl.Start()

	// Or run commands programmatically
	repl.RunCommand("break 100")
	repl.RunCommand("continue")
	repl.RunCommand("locals")
	repl.RunCommand("quit")

REPL Commands:

Breakpoint Management:
  - break, b <location>      Set breakpoint at bytecode location
  - clear, cl <location>     Clear breakpoint at location
  - breakpoints, bp          List all breakpoints

Execution Control:
  - continue, c              Continue execution until next breakpoint
  - step, s                  Step into next instruction
  - next, n                  Step over (don't enter function calls)
  - out, o                   Step out of current function

Inspection:
  - print, p <var>           Print variable value
  - locals, l                Show local variables
  - globals, g               Show global variables
  - stack, st                Show value stack
  - callstack, cs, bt        Show call stack (backtrace)
  - inspect, i <var>         Detailed variable inspection

Evaluation:
  - eval, e <expr>           Evaluate expression (partial support)
  - disassemble, d [pc]      Disassemble instruction at PC

Utility:
  - reset, r                 Reset debugger state
  - help, h, ?               Show help message
  - quit, q, exit            Exit debugger

Integration with VM:

The debugger is designed to wrap VM execution. For full functionality, the VM
would need to expose:
  - Current program counter (PC)
  - Local variables map
  - Global variables map
  - Value stack
  - Execution hooks for breakpoints

Current implementation provides the debugging infrastructure. Future enhancements:
  1. VM hooks for shouldBreak() checks during execution
  2. VM state exposure for variable inspection
  3. Expression parser for eval command
  4. Source-level debugging with line number mapping
  5. Conditional breakpoints
  6. Watchpoints (break on variable change)
  7. Debug info in bytecode for better disassembly

Example Session:

	$ glyph debug myprogram.glyph
	Glyph Debugger REPL
	Type 'help' for available commands
	=====================================

	(glyph-debug:running) break 100
	Breakpoint 1 set at location 0x0064

	(glyph-debug:running) continue
	Continuing execution...
	Breakpoint 1 hit at 0x0064

	(glyph-debug:paused) locals
	Local Variables:
	  x = 42 (int)
	  name = "hello" (string)

	(glyph-debug:paused) stack
	Value Stack:
	  [2] 100 (int)
	  [1] "world" (string)
	  [0] true (bool)

	(glyph-debug:paused) inspect x
	Variable: x
	Type: int
	Value: 42

	(glyph-debug:paused) step
	Stepping into next instruction...

	(glyph-debug:paused) disassemble
	0x0065: ADD (0x10)

	Context:
	   0x0060: PUSH (0x01)
	   0x0061: PUSH (0x01)
	   0x0062: LOAD_VAR (0x40)
	   0x0063: PUSH (0x01)
	   0x0064: LOAD_VAR (0x40)
	=> 0x0065: ADD (0x10)
	   0x0066: STORE_VAR (0x41)
	   0x0067: LOAD_VAR (0x40)

	(glyph-debug:paused) quit
	Goodbye!
*/
