package compiler

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"
)

func TestMacroExpander_RegisterAndGet(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "test_macro",
		Params: []string{"x", "y"},
		Body:   []ast.Node{},
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
	macro := &ast.MacroDef{
		Name:   "double",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  ast.VariableExpr{Name: "x"},
					Right: ast.VariableExpr{Name: "x"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke the macro: double!(5)
	invocation := &ast.MacroInvocation{
		Name: "double",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
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
	retStmt, ok := expanded[0].(ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", expanded[0])
	}

	binOp, ok := retStmt.Value.(ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr, got %T", retStmt.Value)
	}

	// Both left and right should be the substituted value (5)
	leftLit, ok := binOp.Left.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr on left, got %T", binOp.Left)
	}
	if leftLit.Value.(ast.IntLiteral).Value != 5 {
		t.Errorf("Expected left value 5, got %d", leftLit.Value.(ast.IntLiteral).Value)
	}

	rightLit, ok := binOp.Right.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr on right, got %T", binOp.Right)
	}
	if rightLit.Value.(ast.IntLiteral).Value != 5 {
		t.Errorf("Expected right value 5, got %d", rightLit.Value.(ast.IntLiteral).Value)
	}
}

func TestMacroExpander_MultipleParams(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with multiple params: macro! add(a, b) { > a + b }
	macro := &ast.MacroDef{
		Name:   "add",
		Params: []string{"a", "b"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  ast.VariableExpr{Name: "a"},
					Right: ast.VariableExpr{Name: "b"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: add!(10, 20)
	invocation := &ast.MacroInvocation{
		Name: "add",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 20}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	binOp := retStmt.Value.(ast.BinaryOpExpr)

	leftLit := binOp.Left.(ast.LiteralExpr)
	if leftLit.Value.(ast.IntLiteral).Value != 10 {
		t.Errorf("Expected left value 10, got %d", leftLit.Value.(ast.IntLiteral).Value)
	}

	rightLit := binOp.Right.(ast.LiteralExpr)
	if rightLit.Value.(ast.IntLiteral).Value != 20 {
		t.Errorf("Expected right value 20, got %d", rightLit.Value.(ast.IntLiteral).Value)
	}
}

func TestMacroExpander_UndefinedMacro(t *testing.T) {
	expander := NewMacroExpander()

	invocation := &ast.MacroInvocation{
		Name: "undefined_macro",
		Args: []ast.Expr{},
	}

	_, err := expander.ExpandMacroInvocation(invocation)
	if err == nil {
		t.Error("Expected error for undefined macro")
	}
}

func TestMacroExpander_WrongArgCount(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "two_args",
		Params: []string{"a", "b"},
		Body:   []ast.Node{},
	}
	expander.RegisterMacro(macro)

	// Try to invoke with wrong number of args
	invocation := &ast.MacroInvocation{
		Name: "two_args",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
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
	macro := &ast.MacroDef{
		Name:   "greet",
		Params: []string{"name"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.LiteralExpr{
					Value: ast.StringLiteral{Value: "Hello, ${name}!"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: greet!("World")
	invocation := &ast.MacroInvocation{
		Name: "greet",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.StringLiteral{Value: "World"}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	litExpr := retStmt.Value.(ast.LiteralExpr)
	strLit := litExpr.Value.(ast.StringLiteral)

	if strLit.Value != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", strLit.Value)
	}
}

func TestMacroExpander_ExpandModule(t *testing.T) {
	expander := NewMacroExpander()

	// Create a module with a macro definition and invocation
	module := &ast.Module{
		Items: []ast.Item{
			// Define macro
			&ast.MacroDef{
				Name:   "make_route",
				Params: []string{"path"},
				Body: []ast.Node{
					&ast.Route{
						Path:   "/${path}",
						Method: ast.Get,
						Body: []ast.Statement{
							ast.ReturnStatement{
								Value: ast.LiteralExpr{
									Value: ast.StringLiteral{Value: "ok"},
								},
							},
						},
					},
				},
			},
			// Invoke macro
			&ast.MacroInvocation{
				Name: "make_route",
				Args: []ast.Expr{
					ast.LiteralExpr{Value: ast.StringLiteral{Value: "users"}},
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

	route, ok := expanded.Items[0].(*ast.Route)
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
	innerMacro := &ast.MacroDef{
		Name:   "inner",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.BinaryOpExpr{
					Op:    ast.Mul,
					Left:  ast.VariableExpr{Name: "x"},
					Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				},
			},
		},
	}
	expander.RegisterMacro(innerMacro)

	// Define outer macro that uses inner: macro! outer(y) { inner!(y) }
	outerMacro := &ast.MacroDef{
		Name:   "outer",
		Params: []string{"y"},
		Body: []ast.Node{
			&ast.MacroInvocation{
				Name: "inner",
				Args: []ast.Expr{
					ast.VariableExpr{Name: "y"},
				},
			},
		},
	}
	expander.RegisterMacro(outerMacro)

	// Invoke outer!(5) should expand to inner!(5) which expands to > 5 * 2
	invocation := &ast.MacroInvocation{
		Name: "outer",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	retStmt, ok := expanded[0].(ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", expanded[0])
	}

	binOp := retStmt.Value.(ast.BinaryOpExpr)
	leftLit := binOp.Left.(ast.LiteralExpr)
	if leftLit.Value.(ast.IntLiteral).Value != 5 {
		t.Errorf("Expected left value 5, got %d", leftLit.Value.(ast.IntLiteral).Value)
	}
}

func TestMacroExpander_FunctionCallSubstitution(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro that calls a function: macro! log(msg) { print(msg) }
	macro := &ast.MacroDef{
		Name:   "log",
		Params: []string{"msg"},
		Body: []ast.Node{
			ast.ExpressionStatement{
				Expr: ast.FunctionCallExpr{
					Name: "print",
					Args: []ast.Expr{
						ast.VariableExpr{Name: "msg"},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: log!("Hello")
	invocation := &ast.MacroInvocation{
		Name: "log",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.StringLiteral{Value: "Hello"}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	exprStmt := expanded[0].(ast.ExpressionStatement)
	funcCall := exprStmt.Expr.(ast.FunctionCallExpr)

	if funcCall.Name != "print" {
		t.Errorf("Expected function name 'print', got '%s'", funcCall.Name)
	}

	argLit := funcCall.Args[0].(ast.LiteralExpr)
	if argLit.Value.(ast.StringLiteral).Value != "Hello" {
		t.Errorf("Expected arg 'Hello', got '%s'", argLit.Value.(ast.StringLiteral).Value)
	}
}

func TestMacroExpander_IfStatementSubstitution(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro with if statement: macro! check(cond, val) { if cond { > val } }
	macro := &ast.MacroDef{
		Name:   "check",
		Params: []string{"cond", "val"},
		Body: []ast.Node{
			ast.IfStatement{
				Condition: ast.VariableExpr{Name: "cond"},
				ThenBlock: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.VariableExpr{Name: "val"},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Invoke: check!(true, 42)
	invocation := &ast.MacroInvocation{
		Name: "check",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ifStmt := expanded[0].(ast.IfStatement)
	condLit := ifStmt.Condition.(ast.LiteralExpr)
	if !condLit.Value.(ast.BoolLiteral).Value {
		t.Error("Expected condition to be true")
	}

	retStmt := ifStmt.ThenBlock[0].(ast.ReturnStatement)
	valLit := retStmt.Value.(ast.LiteralExpr)
	if valLit.Value.(ast.IntLiteral).Value != 42 {
		t.Errorf("Expected value 42, got %d", valLit.Value.(ast.IntLiteral).Value)
	}
}

// =================================================================
// Additional macro expander tests for expandStatement, expandItem,
// substituteNode, substituteExpr coverage
// =================================================================

func TestMacroExpander_ExpandStatementWhile(t *testing.T) {
	expander := NewMacroExpander()

	// Define a macro that contains a while statement
	macro := &ast.MacroDef{
		Name:   "loop_macro",
		Params: []string{"limit"},
		Body: []ast.Node{
			ast.WhileStatement{
				Condition: ast.BinaryOpExpr{
					Op:    ast.Lt,
					Left:  ast.VariableExpr{Name: "i"},
					Right: ast.VariableExpr{Name: "limit"},
				},
				Body: []ast.Statement{
					ast.AssignStatement{
						Target: "i",
						Value: ast.BinaryOpExpr{
							Op:    ast.Add,
							Left:  ast.VariableExpr{Name: "i"},
							Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "loop_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	whileStmt, ok := expanded[0].(ast.WhileStatement)
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
	macro := &ast.MacroDef{
		Name:   "iter_macro",
		Params: []string{"collection"},
		Body: []ast.Node{
			ast.ForStatement{
				ValueVar: "item",
				Iterable: ast.VariableExpr{Name: "collection"},
				Body: []ast.Statement{
					ast.ExpressionStatement{
						Expr: ast.FunctionCallExpr{
							Name: "print",
							Args: []ast.Expr{ast.VariableExpr{Name: "item"}},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "iter_macro",
		Args: []ast.Expr{
			ast.VariableExpr{Name: "myList"},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	forStmt, ok := expanded[0].(ast.ForStatement)
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
	macro := &ast.MacroDef{
		Name:   "switch_macro",
		Params: []string{"val"},
		Body: []ast.Node{
			ast.SwitchStatement{
				Value: ast.VariableExpr{Name: "val"},
				Cases: []ast.SwitchCase{
					{
						Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
						Body: []ast.Statement{
							ast.ReturnStatement{
								Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "one"}},
							},
						},
					},
				},
				Default: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "other"}},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "switch_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(expanded) != 1 {
		t.Fatalf("Expected 1 expanded node, got %d", len(expanded))
	}

	switchStmt, ok := expanded[0].(ast.SwitchStatement)
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

	macro := &ast.MacroDef{
		Name:   "reassign_macro",
		Params: []string{"val"},
		Body: []ast.Node{
			ast.ReassignStatement{
				Target: "x",
				Value:  ast.VariableExpr{Name: "val"},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "reassign_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	reassignStmt, ok := expanded[0].(ast.ReassignStatement)
	if !ok {
		t.Fatalf("Expected ReassignStatement, got %T", expanded[0])
	}

	litExpr, ok := reassignStmt.Value.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", reassignStmt.Value)
	}
	if litExpr.Value.(ast.IntLiteral).Value != 42 {
		t.Errorf("Expected value 42, got %v", litExpr.Value)
	}
}

func TestMacroExpander_SubstituteExpr_UnaryOp(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "neg_macro",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.UnaryOpExpr{
					Op:    ast.Neg,
					Right: ast.VariableExpr{Name: "x"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "neg_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	unaryOp, ok := retStmt.Value.(ast.UnaryOpExpr)
	if !ok {
		t.Fatalf("Expected UnaryOpExpr, got %T", retStmt.Value)
	}

	litExpr, ok := unaryOp.Right.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", unaryOp.Right)
	}
	if litExpr.Value.(ast.IntLiteral).Value != 5 {
		t.Errorf("Expected value 5, got %v", litExpr.Value)
	}
}

func TestMacroExpander_SubstituteExpr_FieldAccess(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "field_macro",
		Params: []string{"obj"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.FieldAccessExpr{
					Object: ast.VariableExpr{Name: "obj"},
					Field:  "name",
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "field_macro",
		Args: []ast.Expr{
			ast.VariableExpr{Name: "user"},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	fieldAccess, ok := retStmt.Value.(ast.FieldAccessExpr)
	if !ok {
		t.Fatalf("Expected FieldAccessExpr, got %T", retStmt.Value)
	}

	varExpr, ok := fieldAccess.Object.(ast.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr, got %T", fieldAccess.Object)
	}
	if varExpr.Name != "user" {
		t.Errorf("Expected variable 'user', got '%s'", varExpr.Name)
	}
}

func TestMacroExpander_SubstituteExpr_ArrayIndex(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "idx_macro",
		Params: []string{"arr", "i"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.ArrayIndexExpr{
					Array: ast.VariableExpr{Name: "arr"},
					Index: ast.VariableExpr{Name: "i"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "idx_macro",
		Args: []ast.Expr{
			ast.VariableExpr{Name: "myArr"},
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	idxExpr, ok := retStmt.Value.(ast.ArrayIndexExpr)
	if !ok {
		t.Fatalf("Expected ArrayIndexExpr, got %T", retStmt.Value)
	}

	arrVar, ok := idxExpr.Array.(ast.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr for array, got %T", idxExpr.Array)
	}
	if arrVar.Name != "myArr" {
		t.Errorf("Expected variable 'myArr', got '%s'", arrVar.Name)
	}
}

func TestMacroExpander_SubstituteExpr_ObjectExpr(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "obj_macro",
		Params: []string{"val"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{Key: "result", Value: ast.VariableExpr{Name: "val"}},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "obj_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	objExpr, ok := retStmt.Value.(ast.ObjectExpr)
	if !ok {
		t.Fatalf("Expected ObjectExpr, got %T", retStmt.Value)
	}

	if len(objExpr.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(objExpr.Fields))
	}

	litExpr, ok := objExpr.Fields[0].Value.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr, got %T", objExpr.Fields[0].Value)
	}
	if litExpr.Value.(ast.IntLiteral).Value != 42 {
		t.Errorf("Expected value 42, got %v", litExpr.Value)
	}
}

func TestMacroExpander_SubstituteExpr_ArrayExpr(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "arr_macro",
		Params: []string{"a", "b"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.ArrayExpr{
					Elements: []ast.Expr{
						ast.VariableExpr{Name: "a"},
						ast.VariableExpr{Name: "b"},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "arr_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	arrExpr, ok := retStmt.Value.(ast.ArrayExpr)
	if !ok {
		t.Fatalf("Expected ArrayExpr, got %T", retStmt.Value)
	}

	if len(arrExpr.Elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(arrExpr.Elements))
	}
}

func TestMacroExpander_SubstituteExpr_UnquoteExpr(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "unquote_macro",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.UnquoteExpr{
					Expr: ast.VariableExpr{Name: "x"},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "unquote_macro",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 7}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	litExpr, ok := retStmt.Value.(ast.LiteralExpr)
	if !ok {
		t.Fatalf("Expected LiteralExpr (unquote should substitute), got %T", retStmt.Value)
	}
	if litExpr.Value.(ast.IntLiteral).Value != 7 {
		t.Errorf("Expected value 7, got %v", litExpr.Value)
	}
}

func TestMacroExpander_ExpandItemCommand(t *testing.T) {
	expander := NewMacroExpander()

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Command{
				Name:        "test-cmd",
				Description: "A test command",
				Params:      []ast.CommandParam{},
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "done"}},
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

	cmd, ok := expanded.Items[0].(*ast.Command)
	if !ok {
		t.Fatalf("Expected Command, got %T", expanded.Items[0])
	}

	if cmd.Name != "test-cmd" {
		t.Errorf("Expected command name 'test-cmd', got '%s'", cmd.Name)
	}
}

func TestMacroExpander_ExpandItemCronTask(t *testing.T) {
	expander := NewMacroExpander()

	module := &ast.Module{
		Items: []ast.Item{
			&ast.CronTask{
				Name:     "cleanup",
				Schedule: "0 * * * *",
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "cleaned"}},
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

	cron, ok := expanded.Items[0].(*ast.CronTask)
	if !ok {
		t.Fatalf("Expected CronTask, got %T", expanded.Items[0])
	}

	if cron.Name != "cleanup" {
		t.Errorf("Expected task name 'cleanup', got '%s'", cron.Name)
	}
}

func TestMacroExpander_ExpandItemEventHandler(t *testing.T) {
	expander := NewMacroExpander()

	module := &ast.Module{
		Items: []ast.Item{
			&ast.EventHandler{
				EventType: "user.created",
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "handled"}},
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

	handler, ok := expanded.Items[0].(*ast.EventHandler)
	if !ok {
		t.Fatalf("Expected EventHandler, got %T", expanded.Items[0])
	}

	if handler.EventType != "user.created" {
		t.Errorf("Expected event type 'user.created', got '%s'", handler.EventType)
	}
}

func TestMacroExpander_ExpandItemQueueWorker(t *testing.T) {
	expander := NewMacroExpander()

	module := &ast.Module{
		Items: []ast.Item{
			&ast.QueueWorker{
				QueueName: "emails",
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "sent"}},
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

	worker, ok := expanded.Items[0].(*ast.QueueWorker)
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
	macro := &ast.MacroDef{
		Name:   "inc",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  ast.VariableExpr{Name: "x"},
					Right: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	// Module with a route that contains a macro invocation in body
	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/test",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.MacroInvocation{
						Name: "inc",
						Args: []ast.Expr{
							ast.LiteralExpr{Value: ast.IntLiteral{Value: 5}},
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

	route := expanded.Items[0].(*ast.Route)
	if len(route.Body) != 1 {
		t.Fatalf("Expected 1 body statement, got %d", len(route.Body))
	}

	retStmt, ok := route.Body[0].(ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", route.Body[0])
	}

	binOp, ok := retStmt.Value.(ast.BinaryOpExpr)
	if !ok {
		t.Fatalf("Expected BinaryOpExpr, got %T", retStmt.Value)
	}

	if binOp.Op != ast.Add {
		t.Errorf("Expected Add op, got %v", binOp.Op)
	}
}

func TestMacroExpander_SubstituteStringWithIntParam(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "port_route",
		Params: []string{"port"},
		Body: []ast.Node{
			&ast.Route{
				Path:   "/api/port/${port}",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{
							Value: ast.StringLiteral{Value: "port is ${port}"},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "port_route",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 8080}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded[0].(*ast.Route)
	if route.Path != "/api/port/8080" {
		t.Errorf("Expected path '/api/port/8080', got '%s'", route.Path)
	}
}

func TestMacroExpander_SubstituteStringWithVarParam(t *testing.T) {
	expander := NewMacroExpander()

	macro := &ast.MacroDef{
		Name:   "var_route",
		Params: []string{"resource"},
		Body: []ast.Node{
			&ast.Route{
				Path:   "/api/${resource}",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.ReturnStatement{
						Value: ast.LiteralExpr{
							Value: ast.StringLiteral{Value: "ok"},
						},
					},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "var_route",
		Args: []ast.Expr{
			ast.VariableExpr{Name: "items"},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded[0].(*ast.Route)
	if route.Path != "/api/items" {
		t.Errorf("Expected path '/api/items', got '%s'", route.Path)
	}
}

func TestMacroExpander_SubstituteExpr_LiteralNonStringPassthrough(t *testing.T) {
	expander := NewMacroExpander()

	// Non-string literal should pass through unchanged
	macro := &ast.MacroDef{
		Name:   "passthrough",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "passthrough",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 100}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	litExpr := retStmt.Value.(ast.LiteralExpr)
	intLit := litExpr.Value.(ast.IntLiteral)
	if intLit.Value != 42 {
		t.Errorf("Expected 42, got %d", intLit.Value)
	}
}

func TestMacroExpander_DefaultNodePassthrough(t *testing.T) {
	expander := NewMacroExpander()

	// Use a node type that falls through to default in substituteNode
	macro := &ast.MacroDef{
		Name:   "default_macro",
		Params: []string{},
		Body: []ast.Node{
			ast.ValidationStatement{},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "default_macro",
		Args: []ast.Expr{},
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
	macro := &ast.MacroDef{
		Name:   "match_macro",
		Params: []string{},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.MatchExpr{
					Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
					Cases: []ast.MatchCase{},
				},
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "match_macro",
		Args: []ast.Expr{},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	_, ok := retStmt.Value.(ast.MatchExpr)
	if !ok {
		t.Fatalf("Expected MatchExpr, got %T", retStmt.Value)
	}
}

func TestMacroExpander_ExpandItemDefault(t *testing.T) {
	expander := NewMacroExpander()

	// TypeDef falls through to default in expandItem
	module := &ast.Module{
		Items: []ast.Item{
			&ast.TypeDef{
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
	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/test",
				Method: ast.Get,
				Body: []ast.Statement{
					&ast.ValidationStatement{},
					&ast.ReturnStatement{
						Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*ast.Route)
	if len(route.Body) != 2 {
		t.Errorf("Expected 2 body statements, got %d", len(route.Body))
	}
}

func TestMacroExpander_ExpandStatement_IfValueType(t *testing.T) {
	// Tests expandStatement with IfStatement (value type, not pointer)
	expander := NewMacroExpander()

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/test",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.IfStatement{
						Condition: ast.LiteralExpr{Value: ast.BoolLiteral{Value: true}},
						ThenBlock: []ast.Statement{
							ast.ReturnStatement{
								Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
						ElseBlock: []ast.Statement{
							ast.ReturnStatement{
								Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
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

	route := expanded.Items[0].(*ast.Route)
	ifStmt, ok := route.Body[0].(ast.IfStatement)
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

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/test",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.WhileStatement{
						Condition: ast.LiteralExpr{Value: ast.BoolLiteral{Value: false}},
						Body: []ast.Statement{
							ast.AssignStatement{
								Target: "x",
								Value:  ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
							},
						},
					},
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*ast.Route)
	whileStmt, ok := route.Body[0].(ast.WhileStatement)
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

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/test",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.ForStatement{
						ValueVar: "item",
						Iterable: ast.VariableExpr{Name: "items"},
						Body: []ast.Statement{
							ast.ExpressionStatement{
								Expr: ast.FunctionCallExpr{
									Name: "print",
									Args: []ast.Expr{ast.VariableExpr{Name: "item"}},
								},
							},
						},
					},
					ast.ReturnStatement{
						Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
					},
				},
			},
		},
	}

	expanded, err := expander.ExpandModule(module)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	route := expanded.Items[0].(*ast.Route)
	forStmt, ok := route.Body[0].(ast.ForStatement)
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

	module := &ast.Module{
		Items: []ast.Item{
			&ast.Route{
				Path:   "/test",
				Method: ast.Get,
				Body: []ast.Statement{
					ast.SwitchStatement{
						Value: ast.VariableExpr{Name: "x"},
						Cases: []ast.SwitchCase{
							{
								Value: ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
								Body: []ast.Statement{
									ast.ReturnStatement{
										Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "one"}},
									},
								},
							},
						},
						Default: []ast.Statement{
							ast.ReturnStatement{
								Value: ast.LiteralExpr{Value: ast.StringLiteral{Value: "other"}},
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

	route := expanded.Items[0].(*ast.Route)
	switchStmt, ok := route.Body[0].(ast.SwitchStatement)
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
	macro := &ast.MacroDef{
		Name:   "keep_var",
		Params: []string{"x"},
		Body: []ast.Node{
			ast.ReturnStatement{
				Value: ast.VariableExpr{Name: "y"}, // y is not a param
			},
		},
	}
	expander.RegisterMacro(macro)

	invocation := &ast.MacroInvocation{
		Name: "keep_var",
		Args: []ast.Expr{
			ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
		},
	}

	expanded, err := expander.ExpandMacroInvocation(invocation)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retStmt := expanded[0].(ast.ReturnStatement)
	varExpr, ok := retStmt.Value.(ast.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr, got %T", retStmt.Value)
	}
	if varExpr.Name != "y" {
		t.Errorf("Expected variable 'y', got '%s'", varExpr.Name)
	}
}
