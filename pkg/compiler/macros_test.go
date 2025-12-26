package compiler

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

func TestMacroExpander_RegisterAndGet(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "test_macro",
		Params: []string{"x", "y"},
		Body:   []interpreter.Node{},
	}

	expander.RegisterMacro(macro)

	retrieved, ok := expander.GetMacro("test_macro")
	if !ok {
		t.Error("Expected to find registered macro")
	}
	if retrieved.Name != "test_macro" {
		t.Errorf("Expected macro name 'test_macro', got '%s'", retrieved.Name)
	}
	if len(retrieved.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(retrieved.Params))
	}

	// Test non-existent macro
	_, ok = expander.GetMacro("nonexistent")
	if ok {
		t.Error("Expected to not find non-existent macro")
	}
}

func TestMacroExpander_SimpleSubstitution(t *testing.T) {
	expander := NewMacroExpander()

	// Define a simple macro: macro! double(x) { > x + x }
	macro := &interpreter.MacroDef{
		Name:   "double",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  interpreter.VariableExpr{Name: "x"},
					Right: interpreter.VariableExpr{Name: "x"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke the macro: double!(5)
	invocation := &interpreter.MacroInvocation{
		Name: "double",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	// Check that the return statement has the value substituted
	retStmt, ok := expanded[0].(interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", expanded[0])
	}

	binOp, ok := retStmt.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr, got %T", retStmt.Value)
	}

	// Both left and right should be the substituted value (5)
	leftLit, ok := binOp.Left.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr on left, got %T", binOp.Left)
	}
	if leftLit.Value.(interpreter.IntLiteral).Value != 5 {
		t.Errorf("Expected left value 5, got %d", leftLit.Value.(interpreter.IntLiteral).Value)
	}

	rightLit, ok := binOp.Right.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr on right, got %T", binOp.Right)
	}
	if rightLit.Value.(interpreter.IntLiteral).Value != 5 {
		t.Errorf("Expected right value 5, got %d", rightLit.Value.(interpreter.IntLiteral).Value)
	}
}

func TestMacroExpander_MultipleParams(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with multiple params: macro! add(a, b) { > a + b }
	macro := &interpreter.MacroDef{
		Name:   "add",
		Params: []string{"a", "b"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  interpreter.VariableExpr{Name: "a"},
					Right: interpreter.VariableExpr{Name: "b"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: add!(10, 20)
	invocation := &interpreter.MacroInvocation{
		Name: "add",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 20}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	binOp := retStmt.Value.(interpreter.BinaryOpExpr)

	leftLit := binOp.Left.(interpreter.LiteralExpr)
	if leftLit.Value.(interpreter.IntLiteral).Value != 10 {
		t.Errorf("Expected left value 10, got %d", leftLit.Value.(interpreter.IntLiteral).Value)
	}

	rightLit := binOp.Right.(interpreter.LiteralExpr)
	if rightLit.Value.(interpreter.IntLiteral).Value != 20 {
		t.Errorf("Expected right value 20, got %d", rightLit.Value.(interpreter.IntLiteral).Value)
	}
}

func TestMacroExpander_UndefinedMacro(t *testing.T) {
	expander := NewMacroExpander()

	invocation := &interpreter.MacroInvocation{
		Name: "undefined_macro",
		Args: []interpreter.Expr{},
	}

	_, err := expander.ExpandMacroInvocation(invocation)
	if err == nil {
		t.Error("Expected error for undefined macro")
	}
}

func TestMacroExpander_WrongArgCount(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "two_args",
		Params: []string{"a", "b"},
		Body:   []interpreter.Node{},
	}
	expander.RegisterMacro(macro)

	// Try to invoke with wrong number of args
	invocation := &interpreter.MacroInvocation{
		Name: "two_args",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
		},
	}

	_, err := expander.ExpandMacroInvocation(invocation)
	if err == nil {
		t.Error("Expected error for wrong argument count")
	}
}

func TestMacroExpander_StringInterpolation(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with string interpolation: macro! greet(name) { > "Hello, ${name}!" }
	macro := &interpreter.MacroDef{
		Name:   "greet",
		Params: []string{"name"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.LiteralExpr{
					Value: interpreter.StringLiteral{Value: "Hello, ${name}!"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: greet!("World")
	invocation := &interpreter.MacroInvocation{
		Name: "greet",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "World"}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	litExpr := retStmt.Value.(interpreter.LiteralExpr)
	strLit := litExpr.Value.(interpreter.StringLiteral)

	if strLit.Value != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", strLit.Value)
	}
}

func TestMacroExpander_ExpandModule(t *testing.T) {
	expander := NewMacroExpander()

	// Create a module with a macro definition and invocation
	module := &interpreter.Module{
		Items: []interpreter.Item{
			// Define macro
			&interpreter.MacroDef{
				Name:   "make_route",
				Params: []string{"path"},
				Body: []interpreter.Node{
					&interpreter.Route{
						Path:   "/${path}",
						Method: interpreter.Get,
						Body: []interpreter.Statement{
							interpreter.ReturnStatement{
								Value: interpreter.LiteralExpr{
									Value: interpreter.StringLiteral{Value: "ok"},
								},
							},
						},
					},
				},
			},
			// Invoke macro
			&interpreter.MacroInvocation{
				Name: "make_route",
				Args: []interpreter.Expr{
					interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "users"}},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have 1 item (the expanded route, macro def is stripped)
	if len(expanded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(expanded.Items))
	}

	route, ok := expanded.Items[0].(*interpreter.Route)
	if !ok {
		t.Fatalf("Expected Route, got %T", expanded.Items[0])
	}

	if route.Path != "/users" {
		t.Errorf("Expected path '/users', got '%s'", route.Path)
	}
}

func TestMacroExpander_NestedMacros(t *testing.T) {
	expander := NewMacroExpander()

	// Define inner macro: macro! inner(x) { > x * 2 }
	innerMacro := &interpreter.MacroDef{
		Name:   "inner",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.BinaryOpExpr{
					Op:    interpreter.Mul,
					Left:  interpreter.VariableExpr{Name: "x"},
					Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
				},
			},
		},
	}
	expander.RegisterMacro(innerMacro)

	// Define outer macro that uses inner: macro! outer(y) { inner!(y) }
	outerMacro := &interpreter.MacroDef{
		Name:   "outer",
		Params: []string{"y"},
		Body: []interpreter.Node{
			&interpreter.MacroInvocation{
				Name: "inner",
				Args: []interpreter.Expr{
					interpreter.VariableExpr{Name: "y"},
				},
			},
		},
	}
	expander.RegisterMacro(outerMacro)

	// Invoke outer!(5) should expand to inner!(5) which expands to > 5 * 2
	invocation := &interpreter.MacroInvocation{
		Name: "outer",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	retStmt, ok := expanded[0].(interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", expanded[0])
	}

	binOp := retStmt.Value.(interpreter.BinaryOpExpr)
	leftLit := binOp.Left.(interpreter.LiteralExpr)
	if leftLit.Value.(interpreter.IntLiteral).Value != 5 {
		t.Errorf("Expected left value 5, got %d", leftLit.Value.(interpreter.IntLiteral).Value)
	}
}

func TestMacroExpander_FunctionCallSubstitution(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro that calls a function: macro! log(msg) { print(msg) }
	macro := &interpreter.MacroDef{
		Name:   "log",
		Params: []string{"msg"},
		Body: []interpreter.Node{
			interpreter.ExpressionStatement{
				Expr: interpreter.FunctionCallExpr{
					Name: "print",
					Args: []interpreter.Expr{
						interpreter.VariableExpr{Name: "msg"},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: log!("Hello")
	invocation := &interpreter.MacroInvocation{
		Name: "log",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "Hello"}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	exprStmt := expanded[0].(interpreter.ExpressionStatement)
	funcCall := exprStmt.Expr.(interpreter.FunctionCallExpr)

	if funcCall.Name != "print" {
		t.Errorf("Expected function name 'print', got '%s'", funcCall.Name)
	}

	argLit := funcCall.Args[0].(interpreter.LiteralExpr)
	if argLit.Value.(interpreter.StringLiteral).Value != "Hello" {
		t.Errorf("Expected arg 'Hello', got '%s'", argLit.Value.(interpreter.StringLiteral).Value)
	}
}

func TestMacroExpander_IfStatementSubstitution(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with if statement: macro! check(cond, val) { if cond { > val } }
	macro := &interpreter.MacroDef{
		Name:   "check",
		Params: []string{"cond", "val"},
		Body: []interpreter.Node{
			interpreter.IfStatement{
				Condition: interpreter.VariableExpr{Name: "cond"},
				ThenBlock: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.VariableExpr{Name: "val"},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: check!(true, 42)
	invocation := &interpreter.MacroInvocation{
		Name: "check",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ifStmt := expanded[0].(interpreter.IfStatement)
	condLit := ifStmt.Condition.(interpreter.LiteralExpr)
	if !condLit.Value.(interpreter.BoolLiteral).Value {
		t.Error("Expected condition to be true")
	}

	retStmt := ifStmt.ThenBlock[0].(interpreter.ReturnStatement)
	valLit := retStmt.Value.(interpreter.LiteralExpr)
	if valLit.Value.(interpreter.IntLiteral).Value != 42 {
		t.Errorf("Expected value 42, got %d", valLit.Value.(interpreter.IntLiteral).Value)
	}
}
