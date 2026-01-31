package security

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"
)

func TestSQLInjectionDetector_DetectConcatenation(t *testing.T) {
	detector := NewSQLInjectionDetector()

	// Test SQL query with string concatenation
	route := &ast.Route{
		Path: "/api/users",
		Body: []ast.Statement{
			ast.AssignStatement{
				Target: "query",
				Value: ast.BinaryOpExpr{
					Left: ast.LiteralExpr{
						Value: ast.StringLiteral{Value: "SELECT * FROM users WHERE id = "},
					},
					Op:    ast.Add,
					Right: ast.VariableExpr{Name: "userId"},
				},
			},
		},
	}

	warnings := detector.DetectInRoute(route)

	if len(warnings) == 0 {
		t.Fatal("Expected SQL injection warning for string concatenation")
	}

	if warnings[0].Type != "sql_injection" {
		t.Errorf("Expected type 'sql_injection', got '%s'", warnings[0].Type)
	}

	if warnings[0].Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", warnings[0].Severity)
	}
}

func TestSQLInjectionDetector_SafeQuery(t *testing.T) {
	detector := NewSQLInjectionDetector()

	// Test safe query (no concatenation)
	route := &ast.Route{
		Path: "/api/users",
		Body: []ast.Statement{
			ast.ReturnStatement{
				Value: ast.ObjectExpr{
					Fields: []ast.ObjectField{
						{
							Key: "message",
							Value: ast.LiteralExpr{
								Value: ast.StringLiteral{Value: "Hello"},
							},
						},
					},
				},
			},
		},
	}

	warnings := detector.DetectInRoute(route)

	if len(warnings) > 0 {
		t.Errorf("Expected no warnings for safe query, got %d warnings", len(warnings))
	}
}

func TestIsSafeQuery_OldTests(t *testing.T) {
	// Safe query - just a literal
	safeExpr := ast.LiteralExpr{
		Value: ast.StringLiteral{Value: "SELECT * FROM users"},
	}
	if !IsSafeQuery(safeExpr) {
		t.Error("Expected literal query to be safe")
	}

	// Unsafe query - concatenation with SQL
	unsafeExpr := ast.BinaryOpExpr{
		Left: ast.LiteralExpr{
			Value: ast.StringLiteral{Value: "SELECT * FROM users WHERE id = "},
		},
		Op:    ast.Add,
		Right: ast.VariableExpr{Name: "id"},
	}
	if IsSafeQuery(unsafeExpr) {
		t.Error("Expected concatenated query to be unsafe")
	}
}

func TestSanitizeSQL_OldTests(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{
			input:    "user' OR '1'='1",
			expected: "user'' OR ''1''=''1",
			desc:     "single quotes",
		},
		{
			input:    "DROP TABLE users--",
			expected: "DROP TABLE users--",
			desc:     "SQL comments are no longer stripped (use parameterized queries)",
		},
		{
			input:    "test\x00value",
			expected: "testvalue",
			desc:     "null bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := SanitizeSQL(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeSQL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSQLInjectionDetector_NestedExpressions(t *testing.T) {
	detector := NewSQLInjectionDetector()

	// Test nested in if statement
	route := &ast.Route{
		Path: "/api/search",
		Body: []ast.Statement{
			ast.IfStatement{
				Condition: ast.LiteralExpr{
					Value: ast.BoolLiteral{Value: true},
				},
				ThenBlock: []ast.Statement{
					ast.AssignStatement{
						Target: "query",
						Value: ast.BinaryOpExpr{
							Left: ast.LiteralExpr{
								Value: ast.StringLiteral{Value: "DELETE FROM users WHERE name = "},
							},
							Op:    ast.Add,
							Right: ast.VariableExpr{Name: "name"},
						},
					},
				},
			},
		},
	}

	warnings := detector.DetectInRoute(route)

	if len(warnings) == 0 {
		t.Fatal("Expected SQL injection warning in nested if statement")
	}
}
