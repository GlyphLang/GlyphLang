package repl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/interpreter"
)

// executeCommand executes a REPL command (lines starting with :).
func (r *REPL) executeCommand(line string) error {
	// Parse command and arguments
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case ":help", ":h":
		return r.cmdHelp(args)
	case ":quit", ":q", ":exit":
		return r.cmdQuit(args)
	case ":type", ":t":
		return r.cmdType(args)
	case ":load", ":l":
		return r.cmdLoad(args)
	case ":reset", ":r":
		return r.cmdReset(args)
	case ":clear", ":cls":
		return r.cmdClear(args)
	case ":vars", ":v":
		return r.cmdVars(args)
	case ":types":
		return r.cmdTypes(args)
	case ":functions", ":fns":
		return r.cmdFunctions(args)
	default:
		return fmt.Errorf("unknown command: %s (type :help for available commands)", cmd)
	}
}

// cmdHelp displays help information.
func (r *REPL) cmdHelp(args []string) error {
	r.printf("Glyph REPL Commands:\n")
	r.printf("====================\n\n")
	r.printf("Commands:\n")
	r.printf("  :help, :h              - Show this help message\n")
	r.printf("  :quit, :q, :exit       - Exit the REPL\n")
	r.printf("  :type, :t <expr>       - Show the type of an expression\n")
	r.printf("  :load, :l <file>       - Load a .glyph file\n")
	r.printf("  :reset, :r             - Reset the REPL state\n")
	r.printf("  :clear, :cls           - Clear the screen\n")
	r.printf("  :vars, :v              - List all defined variables\n")
	r.printf("  :types                 - List all defined types\n")
	r.printf("  :functions, :fns       - List all defined functions\n")
	r.printf("\n")
	r.printf("Input Types:\n")
	r.printf("  Expressions            - e.g., 1 + 2, \"hello\", [1, 2, 3]\n")
	r.printf("  Statements             - e.g., $ x = 5, > x + 1\n")
	r.printf("  Type definitions       - e.g., : User { name: str!, age: int }\n")
	r.printf("  Function definitions   - e.g., ! add(a: int, b: int) -> int { > a + b }\n")
	r.printf("\n")
	r.printf("Examples:\n")
	r.printf("  glyph> 1 + 2 * 3\n")
	r.printf("  => 7\n")
	r.printf("  glyph> $ name = \"Alice\"\n")
	r.printf("  => \"Alice\"\n")
	r.printf("  glyph> \"Hello, \" + name + \"!\"\n")
	r.printf("  => \"Hello, Alice!\"\n")
	r.printf("  glyph> : Point { x: int!, y: int! }\n")
	r.printf("  Type 'Point' defined\n")
	r.printf("  glyph> ! double(n: int) -> int { > n * 2 }\n")
	r.printf("  Function 'double' defined\n")
	r.printf("  glyph> double(21)\n")
	r.printf("  => 42\n")
	r.printf("\n")
	return nil
}

// cmdQuit exits the REPL.
func (r *REPL) cmdQuit(args []string) error {
	r.running = false
	return nil
}

// cmdType shows the type of an expression.
func (r *REPL) cmdType(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: :type <expression>")
	}

	input := strings.Join(args, " ")

	// Parse and evaluate the expression
	expr, err := r.parseExpression(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	result, err := r.interp.EvaluateExpression(expr, r.env)
	if err != nil {
		return fmt.Errorf("evaluation error: %w", err)
	}

	// Determine the type
	typeName := getTypeName(result)
	r.printf("%s :: %s\n", input, typeName)
	return nil
}

// getTypeName returns the type name of a value.
func getTypeName(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case int64, int:
		return "int"
	case float64:
		return "float"
	case string:
		return "str"
	case bool:
		return "bool"
	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}
		// Try to infer element type from first element
		elemType := getTypeName(val[0])
		return "[" + elemType + "]"
	case map[string]interface{}:
		return "object"
	case ast.Function:
		return "function"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// cmdLoad loads a .glyph file.
func (r *REPL) cmdLoad(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: :load <filename>")
	}

	filepath := args[0]

	// Add .glyph extension if not present
	if !strings.HasSuffix(filepath, ".glyph") {
		filepath += ".glyph"
	}

	r.printf("Loading %s...\n", filepath)

	if err := r.LoadFile(filepath); err != nil {
		return err
	}

	r.printf("Loaded successfully\n")
	return nil
}

// cmdReset resets the REPL state.
func (r *REPL) cmdReset(args []string) error {
	r.Reset()
	r.printf("REPL state reset\n")
	return nil
}

// cmdClear clears the screen.
func (r *REPL) cmdClear(args []string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		// Fallback: print newlines
		for i := 0; i < 50; i++ {
			r.printf("\n")
		}
	}
	return nil
}

// cmdVars lists all defined variables.
func (r *REPL) cmdVars(args []string) error {
	vars := r.getEnvVars(r.env)

	if len(vars) == 0 {
		r.printf("No variables defined\n")
		return nil
	}

	r.printf("Variables:\n")

	// Sort variable names for consistent output
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		val := vars[name]
		// Skip functions
		if _, ok := val.(ast.Function); ok {
			continue
		}
		typeName := getTypeName(val)
		r.printf("  %s :: %s = %s\n", name, typeName, formatValue(val))
	}

	return nil
}

// cmdTypes lists all defined types.
func (r *REPL) cmdTypes(args []string) error {
	// Get types from the interpreter
	// Note: This requires access to interpreter internals
	// For now, we'll show a message about checking via :load
	r.printf("Type definitions are tracked internally.\n")
	r.printf("Use :load <file> to load type definitions from a file.\n")
	r.printf("Types can be defined inline: : TypeName { field: type! }\n")
	return nil
}

// cmdFunctions lists all defined functions.
func (r *REPL) cmdFunctions(args []string) error {
	vars := r.getEnvVars(r.env)

	var fns []string
	for name, val := range vars {
		if fn, ok := val.(ast.Function); ok {
			sig := formatFunctionSignature(name, fn)
			fns = append(fns, sig)
		}
	}

	if len(fns) == 0 {
		r.printf("No functions defined\n")
		return nil
	}

	r.printf("Functions:\n")

	// Sort function names for consistent output
	sort.Strings(fns)

	for _, sig := range fns {
		r.printf("  %s\n", sig)
	}

	return nil
}

// formatFunctionSignature formats a function signature for display.
func formatFunctionSignature(name string, fn ast.Function) string {
	var params []string
	for _, param := range fn.Params {
		paramStr := param.Name
		if param.TypeAnnotation != nil {
			paramStr += ": " + formatType(param.TypeAnnotation)
		}
		if param.Required {
			paramStr += "!"
		}
		params = append(params, paramStr)
	}

	sig := name + "(" + strings.Join(params, ", ") + ")"

	if fn.ReturnType != nil {
		sig += " -> " + formatType(fn.ReturnType)
	}

	return sig
}

// formatType formats a type for display.
func formatType(t ast.Type) string {
	if t == nil {
		return "any"
	}

	switch typ := t.(type) {
	case ast.IntType:
		return "int"
	case ast.StringType:
		return "str"
	case ast.BoolType:
		return "bool"
	case ast.FloatType:
		return "float"
	case ast.ArrayType:
		return "[" + formatType(typ.ElementType) + "]"
	case ast.OptionalType:
		return formatType(typ.InnerType) + "?"
	case ast.NamedType:
		return typ.Name
	case ast.GenericType:
		base := formatType(typ.BaseType)
		if len(typ.TypeArgs) > 0 {
			var args []string
			for _, arg := range typ.TypeArgs {
				args = append(args, formatType(arg))
			}
			return base + "<" + strings.Join(args, ", ") + ">"
		}
		return base
	case ast.FunctionType:
		var params []string
		for _, param := range typ.ParamTypes {
			params = append(params, formatType(param))
		}
		return "(" + strings.Join(params, ", ") + ") -> " + formatType(typ.ReturnType)
	default:
		return fmt.Sprintf("%T", t)
	}
}

// getEnvVars returns all variables in the environment (including parent scopes).
func (r *REPL) getEnvVars(env *interpreter.Environment) map[string]interface{} {
	return env.GetAll()
}
