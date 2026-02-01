package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/glyphlang/glyph/pkg/ast"
)

// SecurityWarning represents a security issue found in code
type SecurityWarning struct {
	Type       string // "XSS", "SQL_INJECTION", etc.
	Severity   string // "HIGH", "MEDIUM", "LOW", "CRITICAL"
	Message    string
	Location   string
	Suggestion string
	UnsafeCode string   // For SQL injection context
	Expr       ast.Expr // For XSS context (can be nil)
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
func (d *SQLInjectionDetector) DetectInRoute(route *ast.Route) []SecurityWarning {
	d.warnings = make([]SecurityWarning, 0)
	for _, stmt := range route.Body {
		d.checkStatement(stmt, route.Path)
	}
	return d.warnings
}

// checkStatement checks a statement for SQL injection
func (d *SQLInjectionDetector) checkStatement(stmt ast.Statement, location string) {
	switch s := stmt.(type) {
	case ast.AssignStatement:
		d.checkExpression(s.Value, location)
	case ast.ReassignStatement:
		d.checkExpression(s.Value, location)
	case ast.ReturnStatement:
		d.checkExpression(s.Value, location)
	case ast.IfStatement:
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
func (d *SQLInjectionDetector) checkExpression(expr ast.Expr, location string) {
	switch e := expr.(type) {
	case ast.BinaryOpExpr:
		if e.Op == ast.Add {
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
	case ast.ObjectExpr:
		for _, field := range e.Fields {
			d.checkExpression(field.Value, location)
		}
	case ast.ArrayExpr:
		for _, elem := range e.Elements {
			d.checkExpression(elem, location)
		}
	}
}

// looksLikeSQL checks if an expression looks like a SQL query
func (d *SQLInjectionDetector) looksLikeSQL(expr ast.Expr) bool {
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
func (d *SQLInjectionDetector) expressionToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case ast.LiteralExpr:
		return fmt.Sprintf("%v", e.Value)
	case ast.VariableExpr:
		return e.Name
	case ast.BinaryOpExpr:
		return fmt.Sprintf("%s %s %s",
			d.expressionToString(e.Left),
			e.Op.String(),
			d.expressionToString(e.Right))
	default:
		return "<expression>"
	}
}

// IsSafeQuery checks if a query expression is safe from SQL injection
func IsSafeQuery(expr ast.Expr) bool {
	detector := NewSQLInjectionDetector()
	detector.checkExpression(expr, "query")
	return len(detector.warnings) == 0
}

// Pre-compiled regexps for StripSQLComments
var (
	sqlLineCommentPattern  = regexp.MustCompile(`--.*$`)
	sqlBlockCommentPattern = regexp.MustCompile(`/\*.*?\*/`)
)

// StripSQLComments removes SQL comments and escapes single quotes.
// WARNING: This is NOT a security measure against SQL injection.
// Always use parameterized queries for user-supplied values.
// This function only strips comments and performs basic escaping.
func StripSQLComments(input string) string {
	input = sqlLineCommentPattern.ReplaceAllString(input, "")
	input = sqlBlockCommentPattern.ReplaceAllString(input, "")
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\x00", "")
	return input
}

// EscapeSQLString performs basic escaping of a string value for SQL contexts.
// WARNING: This is NOT a security measure. Always use parameterized queries
// ($1, $2 placeholders) for user-provided values. This function only escapes
// single quotes for contexts where parameterized queries are not available
// (e.g., identifiers). It does NOT protect against SQL injection.
func EscapeSQLString(input string) string {
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\x00", "")
	return input
}

// SanitizeSQL is deprecated. Use parameterized queries instead.
// Deprecated: This function provides a false sense of security. Use EscapeSQLString
// for non-security escaping, or preferably use parameterized queries.
func SanitizeSQL(input string) string {
	return EscapeSQLString(input)
}
