package compiler

import (
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
)

func TestGetUsedVariablesInExpr_UnaryOpExpr(t *testing.T) {
	used := make(map[string]bool)
	expr := &ast.UnaryOpExpr{
		Op:    ast.Neg,
		Right: &ast.VariableExpr{Name: "x"},
	}
	getUsedVariablesInExpr(expr, used)

	if !used["x"] {
		t.Error("expected variable 'x' to be tracked through UnaryOpExpr")
	}
}

func TestGetUsedVariablesInExpr_UnaryOpExpr_Nested(t *testing.T) {
	used := make(map[string]bool)
	// !(-x) - nested unary
	expr := &ast.UnaryOpExpr{
		Op: ast.Not,
		Right: &ast.UnaryOpExpr{
			Op:    ast.Neg,
			Right: &ast.VariableExpr{Name: "y"},
		},
	}
	getUsedVariablesInExpr(expr, used)

	if !used["y"] {
		t.Error("expected variable 'y' to be tracked through nested UnaryOpExpr")
	}
}

func TestGetUsedVariablesInExpr_ArrayIndexExpr(t *testing.T) {
	used := make(map[string]bool)
	expr := &ast.ArrayIndexExpr{
		Array: &ast.VariableExpr{Name: "arr"},
		Index: &ast.VariableExpr{Name: "idx"},
	}
	getUsedVariablesInExpr(expr, used)

	if !used["arr"] {
		t.Error("expected variable 'arr' to be tracked from ArrayIndexExpr.Array")
	}
	if !used["idx"] {
		t.Error("expected variable 'idx' to be tracked from ArrayIndexExpr.Index")
	}
}

func TestGetUsedVariablesInExpr_ArrayIndexExpr_LiteralIndex(t *testing.T) {
	used := make(map[string]bool)
	expr := &ast.ArrayIndexExpr{
		Array: &ast.VariableExpr{Name: "items"},
		Index: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 0}},
	}
	getUsedVariablesInExpr(expr, used)

	if !used["items"] {
		t.Error("expected variable 'items' to be tracked from ArrayIndexExpr.Array")
	}
	if len(used) != 1 {
		t.Errorf("expected exactly 1 variable tracked, got %d", len(used))
	}
}

func TestGetUsedVariablesInExpr_FunctionCallExpr(t *testing.T) {
	used := make(map[string]bool)
	expr := &ast.FunctionCallExpr{
		Name: "myFunc",
		Args: []ast.Expr{
			&ast.VariableExpr{Name: "a"},
			&ast.VariableExpr{Name: "b"},
			&ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
		},
	}
	getUsedVariablesInExpr(expr, used)

	if !used["a"] {
		t.Error("expected variable 'a' to be tracked from FunctionCallExpr args")
	}
	if !used["b"] {
		t.Error("expected variable 'b' to be tracked from FunctionCallExpr args")
	}
	if len(used) != 2 {
		t.Errorf("expected 2 variables tracked, got %d", len(used))
	}
}

func TestGetUsedVariablesInExpr_FunctionCallExpr_NoArgs(t *testing.T) {
	used := make(map[string]bool)
	expr := &ast.FunctionCallExpr{
		Name: "noArgs",
		Args: []ast.Expr{},
	}
	getUsedVariablesInExpr(expr, used)

	if len(used) != 0 {
		t.Errorf("expected 0 variables tracked for no-arg function call, got %d", len(used))
	}
}

func TestGetUsedVariablesInExpr_ComplexNested(t *testing.T) {
	used := make(map[string]bool)
	// Represents: myFunc(-arr[idx], other)
	expr := &ast.FunctionCallExpr{
		Name: "myFunc",
		Args: []ast.Expr{
			&ast.UnaryOpExpr{
				Op: ast.Neg,
				Right: &ast.ArrayIndexExpr{
					Array: &ast.VariableExpr{Name: "arr"},
					Index: &ast.VariableExpr{Name: "idx"},
				},
			},
			&ast.VariableExpr{Name: "other"},
		},
	}
	getUsedVariablesInExpr(expr, used)

	expected := []string{"arr", "idx", "other"}
	for _, name := range expected {
		if !used[name] {
			t.Errorf("expected variable '%s' to be tracked in complex nested expression", name)
		}
	}
	if len(used) != 3 {
		t.Errorf("expected 3 variables tracked, got %d", len(used))
	}
}
