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

// =================================================================
// Additional macro expander tests for expandStatement, expandItem,
// substituteNode, substituteExpr coverage
// =================================================================

func TestMacroExpander_ExpandStatementWhile(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro that contains a while statement
	macro := &interpreter.MacroDef{
		Name:   "loop_macro",
		Params: []string{"limit"},
		Body: []interpreter.Node{
			interpreter.WhileStatement{
				Condition: interpreter.BinaryOpExpr{
					Op:    interpreter.Lt,
					Left:  interpreter.VariableExpr{Name: "i"},
					Right: interpreter.VariableExpr{Name: "limit"},
				},
				Body: []interpreter.Statement{
					interpreter.AssignStatement{
						Target: "i",
						Value: interpreter.BinaryOpExpr{
							Op:    interpreter.Add,
							Left:  interpreter.VariableExpr{Name: "i"},
							Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "loop_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 10}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	whileStmt, ok := expanded[0].(interpreter.WhileStatement)
	if !ok {
		t.Fatalf("Expected WhileStatement, got %T", expanded[0])
	}

	if len(whileStmt.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestMacroExpander_ExpandStatementFor(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with a for statement
	macro := &interpreter.MacroDef{
		Name:   "iter_macro",
		Params: []string{"collection"},
		Body: []interpreter.Node{
			interpreter.ForStatement{
				ValueVar: "item",
				Iterable: interpreter.VariableExpr{Name: "collection"},
				Body: []interpreter.Statement{
					interpreter.ExpressionStatement{
						Expr: interpreter.FunctionCallExpr{
							Name: "print",
							Args: []interpreter.Expr{interpreter.VariableExpr{Name: "item"}},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "iter_macro",
		Args: []interpreter.Expr{
			interpreter.VariableExpr{Name: "myList"},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	forStmt, ok := expanded[0].(interpreter.ForStatement)
	if !ok {
		t.Fatalf("Expected ForStatement, got %T", expanded[0])
	}

	if forStmt.ValueVar != "item" {
		t.Errorf("Expected ValueVar 'item', got '%s'", forStmt.ValueVar)
	}
}

func TestMacroExpander_ExpandStatementSwitch(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with a switch statement
	macro := &interpreter.MacroDef{
		Name:   "switch_macro",
		Params: []string{"val"},
		Body: []interpreter.Node{
			interpreter.SwitchStatement{
				Value: interpreter.VariableExpr{Name: "val"},
				Cases: []interpreter.SwitchCase{
					{
						Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
						Body: []interpreter.Statement{
							interpreter.ReturnStatement{
								Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
							},
						},
					},
				},
				Default: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "other"}},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "switch_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	switchStmt, ok := expanded[0].(interpreter.SwitchStatement)
	if !ok {
		t.Fatalf("Expected SwitchStatement, got %T", expanded[0])
	}

	if len(switchStmt.Cases) != 1 {
		t.Errorf("Expected 1 case, got %d", len(switchStmt.Cases))
	}
	if len(switchStmt.Default) != 1 {
		t.Errorf("Expected 1 default statement, got %d", len(switchStmt.Default))
	}
}

func TestMacroExpander_SubstituteNode_ReassignStatement(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "reassign_macro",
		Params: []string{"val"},
		Body: []interpreter.Node{
			interpreter.ReassignStatement{
				Target: "x",
				Value:  interpreter.VariableExpr{Name: "val"},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "reassign_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	reassignStmt, ok := expanded[0].(interpreter.ReassignStatement)
	if !ok {
		t.Fatalf("Expected ReassignStatement, got %T", expanded[0])
	}

	litExpr, ok := reassignStmt.Value.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", reassignStmt.Value)
	}
	if litExpr.Value.(interpreter.IntLiteral).Value != 42 {
		t.Errorf("Expected value 42, got %v", litExpr.Value)
	}
}

func TestMacroExpander_SubstituteExpr_UnaryOp(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "neg_macro",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.UnaryOpExpr{
					Op:    interpreter.Neg,
					Right: interpreter.VariableExpr{Name: "x"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "neg_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	unaryOp, ok := retStmt.Value.(interpreter.UnaryOpExpr)
	if !ok {
		t.Fatalf("Expected UnaryOpExpr, got %T", retStmt.Value)
	}

	litExpr, ok := unaryOp.Right.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", unaryOp.Right)
	}
	if litExpr.Value.(interpreter.IntLiteral).Value != 5 {
		t.Errorf("Expected value 5, got %v", litExpr.Value)
	}
}

func TestMacroExpander_SubstituteExpr_FieldAccess(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "field_macro",
		Params: []string{"obj"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.FieldAccessExpr{
					Object: interpreter.VariableExpr{Name: "obj"},
					Field:  "name",
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "field_macro",
		Args: []interpreter.Expr{
			interpreter.VariableExpr{Name: "user"},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	fieldAccess, ok := retStmt.Value.(interpreter.FieldAccessExpr)
	if !ok {
		t.Fatalf("Expected FieldAccessExpr, got %T", retStmt.Value)
	}

	varExpr, ok := fieldAccess.Object.(interpreter.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr, got %T", fieldAccess.Object)
	}
	if varExpr.Name != "user" {
		t.Errorf("Expected variable 'user', got '%s'", varExpr.Name)
	}
}

func TestMacroExpander_SubstituteExpr_ArrayIndex(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "idx_macro",
		Params: []string{"arr", "i"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.ArrayIndexExpr{
					Array: interpreter.VariableExpr{Name: "arr"},
					Index: interpreter.VariableExpr{Name: "i"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "idx_macro",
		Args: []interpreter.Expr{
			interpreter.VariableExpr{Name: "myArr"},
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	idxExpr, ok := retStmt.Value.(interpreter.ArrayIndexExpr)
	if !ok {
		t.Fatalf("Expected ArrayIndexExpr, got %T", retStmt.Value)
	}

	arrVar, ok := idxExpr.Array.(interpreter.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr for array, got %T", idxExpr.Array)
	}
	if arrVar.Name != "myArr" {
		t.Errorf("Expected variable 'myArr', got '%s'", arrVar.Name)
	}
}

func TestMacroExpander_SubstituteExpr_ObjectExpr(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "obj_macro",
		Params: []string{"val"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.ObjectExpr{
					Fields: []interpreter.ObjectField{
						{Key: "result", Value: interpreter.VariableExpr{Name: "val"}},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "obj_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 42}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	objExpr, ok := retStmt.Value.(interpreter.ObjectExpr)
	if !ok {
		t.Fatalf("Expected ObjectExpr, got %T", retStmt.Value)
	}

	if len(objExpr.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(objExpr.Fields))
	}

	litExpr, ok := objExpr.Fields[0].Value.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", objExpr.Fields[0].Value)
	}
	if litExpr.Value.(interpreter.IntLiteral).Value != 42 {
		t.Errorf("Expected value 42, got %v", litExpr.Value)
	}
}

func TestMacroExpander_SubstituteExpr_ArrayExpr(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "arr_macro",
		Params: []string{"a", "b"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.ArrayExpr{
					Elements: []interpreter.Expr{
						interpreter.VariableExpr{Name: "a"},
						interpreter.VariableExpr{Name: "b"},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "arr_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	arrExpr, ok := retStmt.Value.(interpreter.ArrayExpr)
	if !ok {
		t.Fatalf("Expected ArrayExpr, got %T", retStmt.Value)
	}

	if len(arrExpr.Elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(arrExpr.Elements))
	}
}

func TestMacroExpander_SubstituteExpr_UnquoteExpr(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "unquote_macro",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.UnquoteExpr{
					Expr: interpreter.VariableExpr{Name: "x"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "unquote_macro",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 7}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	litExpr, ok := retStmt.Value.(interpreter.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr (unquote should substitute), got %T", retStmt.Value)
	}
	if litExpr.Value.(interpreter.IntLiteral).Value != 7 {
		t.Errorf("Expected value 7, got %v", litExpr.Value)
	}
}

func TestMacroExpander_ExpandItemCommand(t *testing.T) {
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Command{
				Name:        "test-cmd",
				Description: "A test command",
				Params:      []interpreter.CommandParam{},
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "done"}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(expanded.Items))
	}

	cmd, ok := expanded.Items[0].(*interpreter.Command)
	if !ok {
		t.Fatalf("Expected Command, got %T", expanded.Items[0])
	}

	if cmd.Name != "test-cmd" {
		t.Errorf("Expected command name 'test-cmd', got '%s'", cmd.Name)
	}
}

func TestMacroExpander_ExpandItemCronTask(t *testing.T) {
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.CronTask{
				Name:     "cleanup",
				Schedule: "0 * * * *",
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "cleaned"}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(expanded.Items))
	}

	cron, ok := expanded.Items[0].(*interpreter.CronTask)
	if !ok {
		t.Fatalf("Expected CronTask, got %T", expanded.Items[0])
	}

	if cron.Name != "cleanup" {
		t.Errorf("Expected task name 'cleanup', got '%s'", cron.Name)
	}
}

func TestMacroExpander_ExpandItemEventHandler(t *testing.T) {
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.EventHandler{
				EventType: "user.created",
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "handled"}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(expanded.Items))
	}

	handler, ok := expanded.Items[0].(*interpreter.EventHandler)
	if !ok {
		t.Fatalf("Expected EventHandler, got %T", expanded.Items[0])
	}

	if handler.EventType != "user.created" {
		t.Errorf("Expected event type 'user.created', got '%s'", handler.EventType)
	}
}

func TestMacroExpander_ExpandItemQueueWorker(t *testing.T) {
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.QueueWorker{
				QueueName: "emails",
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "sent"}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(expanded.Items))
	}

	worker, ok := expanded.Items[0].(*interpreter.QueueWorker)
	if !ok {
		t.Fatalf("Expected QueueWorker, got %T", expanded.Items[0])
	}

	if worker.QueueName != "emails" {
		t.Errorf("Expected queue name 'emails', got '%s'", worker.QueueName)
	}
}

func TestMacroExpander_ExpandStatementWithMacroInvocation(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro
	macro := &interpreter.MacroDef{
		Name:   "inc",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.BinaryOpExpr{
					Op:    interpreter.Add,
					Left:  interpreter.VariableExpr{Name: "x"},
					Right: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Module with a route that contains a macro invocation in body
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/test",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					interpreter.MacroInvocation{
						Name: "inc",
						Args: []interpreter.Expr{
							interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 5}},
						},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*interpreter.Route)
	if len(route.Body) != 1 {
		t.Fatalf("Expected 1 body statement, got %d", len(route.Body))
	}

	retStmt, ok := route.Body[0].(interpreter.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", route.Body[0])
	}

	binOp, ok := retStmt.Value.(interpreter.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr, got %T", retStmt.Value)
	}

	if binOp.Op != interpreter.Add {
		t.Errorf("Expected Add op, got %v", binOp.Op)
	}
}

func TestMacroExpander_SubstituteStringWithIntParam(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "port_route",
		Params: []string{"port"},
		Body: []interpreter.Node{
			&interpreter.Route{
				Path:   "/api/port/${port}",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{
							Value: interpreter.StringLiteral{Value: "port is ${port}"},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "port_route",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 8080}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded[0].(*interpreter.Route)
	if route.Path != "/api/port/8080" {
		t.Errorf("Expected path '/api/port/8080', got '%s'", route.Path)
	}
}

func TestMacroExpander_SubstituteStringWithVarParam(t *testing.T) {
	expander := NewMacroExpander()

	macro := &interpreter.MacroDef{
		Name:   "var_route",
		Params: []string{"resource"},
		Body: []interpreter.Node{
			&interpreter.Route{
				Path:   "/api/${resource}",
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
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "var_route",
		Args: []interpreter.Expr{
			interpreter.VariableExpr{Name: "items"},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded[0].(*interpreter.Route)
	if route.Path != "/api/items" {
		t.Errorf("Expected path '/api/items', got '%s'", route.Path)
	}
}

func TestMacroExpander_SubstituteExpr_LiteralNonStringPassthrough(t *testing.T) {
	expander := NewMacroExpander()

	// Non-string literal should pass through unchanged
	macro := &interpreter.MacroDef{
		Name:   "passthrough",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.LiteralExpr{
					Value: interpreter.IntLiteral{Value: 42},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "passthrough",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 100}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	litExpr := retStmt.Value.(interpreter.LiteralExpr)
	intLit := litExpr.Value.(interpreter.IntLiteral)
	if intLit.Value != 42 {
		t.Errorf("Expected 42, got %d", intLit.Value)
	}
}

func TestMacroExpander_DefaultNodePassthrough(t *testing.T) {
	expander := NewMacroExpander()

	// Use a node type that falls through to default in substituteNode
	macro := &interpreter.MacroDef{
		Name:   "default_macro",
		Params: []string{},
		Body: []interpreter.Node{
			interpreter.ValidationStatement{},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "default_macro",
		Args: []interpreter.Expr{},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}
}

func TestMacroExpander_DefaultExprPassthrough(t *testing.T) {
	expander := NewMacroExpander()

	// Use a MatchExpr which falls through to default in substituteExpr
	macro := &interpreter.MacroDef{
		Name:   "match_macro",
		Params: []string{},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.MatchExpr{
					Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
					Cases: []interpreter.MatchCase{},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "match_macro",
		Args: []interpreter.Expr{},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	_, ok := retStmt.Value.(interpreter.MatchExpr)
	if !ok {
		t.Fatalf("Expected MatchExpr, got %T", retStmt.Value)
	}
}

func TestMacroExpander_ExpandItemDefault(t *testing.T) {
	expander := NewMacroExpander()

	// TypeDef falls through to default in expandItem
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.TypeDef{
				Name: "MyType",
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(expanded.Items))
	}
}

func TestMacroExpander_ExpandStatementDefault(t *testing.T) {
	expander := NewMacroExpander()

	// Route with a validation statement - falls through default in expandStatement
	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/test",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					&interpreter.ValidationStatement{},
					&interpreter.ReturnStatement{
						Value: &interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*interpreter.Route)
	if len(route.Body) != 2 {
		t.Errorf("Expected 2 body statements, got %d", len(route.Body))
	}
}

func TestMacroExpander_ExpandStatement_IfValueType(t *testing.T) {
	// Tests expandStatement with IfStatement (value type, not pointer)
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/test",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					interpreter.IfStatement{
						Condition: interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: true}},
						ThenBlock: []interpreter.Statement{
							interpreter.ReturnStatement{
								Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
							},
						},
						ElseBlock: []interpreter.Statement{
							interpreter.ReturnStatement{
								Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 2}},
							},
						},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*interpreter.Route)
	ifStmt, ok := route.Body[0].(interpreter.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", route.Body[0])
	}
	if len(ifStmt.ThenBlock) != 1 {
		t.Errorf("Expected 1 then block statement, got %d", len(ifStmt.ThenBlock))
	}
	if len(ifStmt.ElseBlock) != 1 {
		t.Errorf("Expected 1 else block statement, got %d", len(ifStmt.ElseBlock))
	}
}

func TestMacroExpander_ExpandStatement_WhileValueType(t *testing.T) {
	// Tests expandStatement with WhileStatement (value type)
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/test",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					interpreter.WhileStatement{
						Condition: interpreter.LiteralExpr{Value: interpreter.BoolLiteral{Value: false}},
						Body: []interpreter.Statement{
							interpreter.AssignStatement{
								Target: "x",
								Value:  interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
							},
						},
					},
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*interpreter.Route)
	whileStmt, ok := route.Body[0].(interpreter.WhileStatement)
	if !ok {
		t.Fatalf("Expected WhileStatement, got %T", route.Body[0])
	}
	if len(whileStmt.Body) != 1 {
		t.Errorf("Expected 1 body statement, got %d", len(whileStmt.Body))
	}
}

func TestMacroExpander_ExpandStatement_ForValueType(t *testing.T) {
	// Tests expandStatement with ForStatement (value type)
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/test",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					interpreter.ForStatement{
						ValueVar: "item",
						Iterable: interpreter.VariableExpr{Name: "items"},
						Body: []interpreter.Statement{
							interpreter.ExpressionStatement{
								Expr: interpreter.FunctionCallExpr{
									Name: "print",
									Args: []interpreter.Expr{interpreter.VariableExpr{Name: "item"}},
								},
							},
						},
					},
					interpreter.ReturnStatement{
						Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 0}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*interpreter.Route)
	forStmt, ok := route.Body[0].(interpreter.ForStatement)
	if !ok {
		t.Fatalf("Expected ForStatement, got %T", route.Body[0])
	}
	if forStmt.ValueVar != "item" {
		t.Errorf("Expected ValueVar 'item', got '%s'", forStmt.ValueVar)
	}
}

func TestMacroExpander_ExpandStatement_SwitchValueType(t *testing.T) {
	// Tests expandStatement with SwitchStatement (value type)
	expander := NewMacroExpander()

	module := &interpreter.Module{
		Items: []interpreter.Item{
			&interpreter.Route{
				Path:   "/test",
				Method: interpreter.Get,
				Body: []interpreter.Statement{
					interpreter.SwitchStatement{
						Value: interpreter.VariableExpr{Name: "x"},
						Cases: []interpreter.SwitchCase{
							{
								Value: interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
								Body: []interpreter.Statement{
									interpreter.ReturnStatement{
										Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "one"}},
									},
								},
							},
						},
						Default: []interpreter.Statement{
							interpreter.ReturnStatement{
								Value: interpreter.LiteralExpr{Value: interpreter.StringLiteral{Value: "other"}},
							},
						},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*interpreter.Route)
	switchStmt, ok := route.Body[0].(interpreter.SwitchStatement)
	if !ok {
		t.Fatalf("Expected SwitchStatement, got %T", route.Body[0])
	}
	if len(switchStmt.Cases) != 1 {
		t.Errorf("Expected 1 case, got %d", len(switchStmt.Cases))
	}
	if len(switchStmt.Default) != 1 {
		t.Errorf("Expected 1 default statement, got %d", len(switchStmt.Default))
	}
}

func TestMacroExpander_SubstituteExpr_NonSubstitutedVariable(t *testing.T) {
	expander := NewMacroExpander()

	// Variable that is NOT in the substitution map should pass through
	macro := &interpreter.MacroDef{
		Name:   "keep_var",
		Params: []string{"x"},
		Body: []interpreter.Node{
			interpreter.ReturnStatement{
				Value: interpreter.VariableExpr{Name: "y"}, // y is not a param
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &interpreter.MacroInvocation{
		Name: "keep_var",
		Args: []interpreter.Expr{
			interpreter.LiteralExpr{Value: interpreter.IntLiteral{Value: 1}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(interpreter.ReturnStatement)
	varExpr, ok := retStmt.Value.(interpreter.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr, got %T", retStmt.Value)
	}
	if varExpr.Name != "y" {
		t.Errorf("Expected variable 'y', got '%s'", varExpr.Name)
	}
}
