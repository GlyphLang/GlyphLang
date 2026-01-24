package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// SecurityWarning represents a security issue found in code
type SecurityWarning struct {
	Type        string // "XSS", "SQL_INJECTION", etc.
	Severity    string // "HIGH", "MEDIUM", "LOW", "CRITICAL"
	Message     string
	Location    string
	Suggestion  string
	UnsafeCode  string // For SQL injection context
	Expr        interpreter.Expr // For XSS context (can be nil)
}

// SQLInjectionDetector detects potential SQL injection vulnerabilities
type SQLInjectionDetector struct {
	warnings []SecurityWarning
}

// NewSQLInjectionDetector creates a new SQL injection detector
func NewSQLInjectionDetector() *SQLInjectionDetector {
	return &SQLInjectionDetector{
		warnings: make([]SecurityWarning, 0),
	}
}

// DetectInRoute analyzes a route for SQL injection vulnerabilities
func (d *SQLInjectionDetector) DetectInRoute(route *interpreter.Route) []SecurityWarning {
	d.warnings = make([]SecurityWarning, 0)
	for _, stmt := range route.Body {
		d.checkStatement(stmt, route.Path)
	}
	return d.warnings
}

// checkStatement checks a statement for SQL injection
func (d *SQLInjectionDetector) checkStatement(stmt interpreter.Statement, location string) {
	switch s := stmt.(type) {
	case interpreter.AssignStatement:
		d.checkExpression(s.Value, location)
	case interpreter.ReassignStatement:
		d.checkExpression(s.Value, location)
	case interpreter.ReturnStatement:
		d.checkExpression(s.Value, location)
	case interpreter.IfStatement:
		d.checkExpression(s.Condition, location)
		for _, thenStmt := range s.ThenBlock {
			d.checkStatement(thenStmt, location)
		}
		for _, elseStmt := range s.ElseBlock {
			d.checkStatement(elseStmt, location)
		}
	}
}

// checkExpression checks an expression for SQL injection
func (d *SQLInjectionDetector) checkExpression(expr interpreter.Expr, location string) {
	switch e := expr.(type) {
	case interpreter.BinaryOpExpr:
		if e.Op == interpreter.Add {
			if d.looksLikeSQL(e.Left) || d.looksLikeSQL(e.Right) {
				d.warnings = append(d.warnings, SecurityWarning{
					Type:       "sql_injection",
					Severity:   "critical",
					Message:    "Potential SQL injection: string concatenation in SQL query",
					Location:   location,
					Suggestion: "Use parameterized queries instead of string concatenation",
					UnsafeCode: d.expressionToString(expr),
					Expr:       expr,
				})
			}
		}
		d.checkExpression(e.Left, location)
		d.checkExpression(e.Right, location)
	case interpreter.ObjectExpr:
		for _, field := range e.Fields {
			d.checkExpression(field.Value, location)
		}
	case interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			d.checkExpression(elem, location)
		}
	}
}

// looksLikeSQL checks if an expression looks like a SQL query
func (d *SQLInjectionDetector) looksLikeSQL(expr interpreter.Expr) bool {
	str := d.expressionToString(expr)
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TRUNCATE", "FROM", "WHERE", "JOIN", "UNION",
	}
	upperStr := strings.ToUpper(str)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperStr, keyword) {
			return true
		}
	}
	return false
}

// expressionToString converts an expression to a string representation
func (d *SQLInjectionDetector) expressionToString(expr interpreter.Expr) string {
	switch e := expr.(type) {
	case interpreter.LiteralExpr:
		return fmt.Sprintf("%v", e.Value)
	case interpreter.VariableExpr:
		return e.Name
	case interpreter.BinaryOpExpr:
		return fmt.Sprintf("%s %s %s",
			d.expressionToString(e.Left),
			e.Op.String(),
			d.expressionToString(e.Right))
	default:
		return "<expression>"
	}
}

// IsSafeQuery checks if a query expression is safe from SQL injection
func IsSafeQuery(expr interpreter.Expr) bool {
	detector := NewSQLInjectionDetector()
	detector.checkExpression(expr, "query")
	return len(detector.warnings) == 0
}

// SanitizeSQL escapes dangerous characters
func SanitizeSQL(input string) string {
	input = regexp.MustCompile(`--.*$`).ReplaceAllString(input, "")
	input = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(input, "")
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\x00", "")
	return input
}
