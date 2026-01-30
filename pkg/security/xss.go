package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/glyphlang/glyph/pkg/interpreter"
)

// XSSDetector detects Cross-Site Scripting vulnerabilities
type XSSDetector struct {
	warnings []SecurityWarning
}

// NewXSSDetector creates a new XSS detector
func NewXSSDetector() *XSSDetector {
	return &XSSDetector{
		warnings: make([]SecurityWarning, 0),
	}
}

// DetectXSS analyzes an expression for XSS vulnerabilities
func DetectXSS(expr interpreter.Expr) []SecurityWarning {
	detector := NewXSSDetector()
	detector.analyzeExpr(expr, false)
	return detector.warnings
}

// analyzeExpr recursively analyzes an expression for XSS patterns
func (d *XSSDetector) analyzeExpr(expr interpreter.Expr, inHTMLContext bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case interpreter.LiteralExpr:
		d.analyzeLiteral(e, inHTMLContext)

	case interpreter.VariableExpr:
		d.analyzeVariable(e, inHTMLContext)

	case interpreter.BinaryOpExpr:
		// String concatenation is particularly risky for XSS
		if e.Op == interpreter.Add {
			d.analyzeExpr(e.Left, inHTMLContext)
			d.analyzeExpr(e.Right, inHTMLContext)

			// Check if we're concatenating HTML tags with variables
			if d.containsHTMLPatterns(e) && d.containsUserInput(e) {
				d.addWarning(SecurityWarning{
					Type:       "XSS",
					Severity:   "HIGH",
					Message:    "Potential XSS: Concatenating user input with HTML content",
					Location:   "binary operation",
					Suggestion: SuggestHTMLEscape(e),
					Expr:       e,
				})
			}
		} else {
			d.analyzeExpr(e.Left, inHTMLContext)
			d.analyzeExpr(e.Right, inHTMLContext)
		}

	case interpreter.FunctionCallExpr:
		d.analyzeFunctionCall(e, inHTMLContext)
		for _, arg := range e.Args {
			d.analyzeExpr(arg, inHTMLContext)
		}

	case interpreter.ObjectExpr:
		d.analyzeObjectExpr(e)

	case interpreter.FieldAccessExpr:
		// Field access from request/input is user data
		if varExpr, ok := e.Object.(interpreter.VariableExpr); ok {
			if isUserInputSource(varExpr.Name) && inHTMLContext {
				d.addWarning(SecurityWarning{
					Type:       "XSS",
					Severity:   "HIGH",
					Message:    fmt.Sprintf("Potential XSS: User input '%s.%s' used without escaping", varExpr.Name, e.Field),
					Location:   "field access",
					Suggestion: SuggestHTMLEscape(e),
					Expr:       e,
				})
			}
		}
		d.analyzeExpr(e.Object, inHTMLContext)

	case interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			d.analyzeExpr(elem, inHTMLContext)
		}
	}
}

// analyzeLiteral checks string literals for HTML/JavaScript content
func (d *XSSDetector) analyzeLiteral(lit interpreter.LiteralExpr, inHTMLContext bool) {
	if strLit, ok := lit.Value.(interpreter.StringLiteral); ok {
		if containsHTMLTags(strLit.Value) {
			d.addWarning(SecurityWarning{
				Type:       "XSS",
				Severity:   "LOW",
				Message:    "HTML content detected in string literal",
				Location:   "string literal",
				Suggestion: "Ensure this HTML is intentional and properly escaped if it contains dynamic content",
				Expr:       lit,
			})
		}

		if containsScriptTags(strLit.Value) {
			d.addWarning(SecurityWarning{
				Type:       "XSS",
				Severity:   "MEDIUM",
				Message:    "JavaScript <script> tags detected in string literal",
				Location:   "string literal",
				Suggestion: "Avoid embedding script tags in strings; use external JS files",
				Expr:       lit,
			})
		}

		if containsEventHandlers(strLit.Value) {
			d.addWarning(SecurityWarning{
				Type:       "XSS",
				Severity:   "MEDIUM",
				Message:    "JavaScript event handlers (onclick, onerror, etc.) detected",
				Location:   "string literal",
				Suggestion: "Remove inline event handlers; use proper event listeners",
				Expr:       lit,
			})
		}
	}
}

// analyzeVariable checks if a variable might contain user input
func (d *XSSDetector) analyzeVariable(varExpr interpreter.VariableExpr, inHTMLContext bool) {
	if isUserInputSource(varExpr.Name) && inHTMLContext {
		d.addWarning(SecurityWarning{
			Type:       "XSS",
			Severity:   "HIGH",
			Message:    fmt.Sprintf("Potential XSS: User input variable '%s' used without escaping", varExpr.Name),
			Location:   "variable reference",
			Suggestion: SuggestHTMLEscape(varExpr),
			Expr:       varExpr,
		})
	}
}

// analyzeFunctionCall checks for unsafe function calls
func (d *XSSDetector) analyzeFunctionCall(call interpreter.FunctionCallExpr, inHTMLContext bool) {
	// Check for functions that output HTML
	htmlOutputFunctions := map[string]bool{
		"render":     true,
		"renderHTML": true,
		"toHTML":     true,
		"innerHTML":  true,
		"write":      true,
		"writeHTML":  true,
		"sendHTML":   true,
	}

	if htmlOutputFunctions[call.Name] {
		// This function outputs HTML, check its arguments
		for _, arg := range call.Args {
			if RequiresHTMLEscape(arg) {
				d.addWarning(SecurityWarning{
					Type:       "XSS",
					Severity:   "HIGH",
					Message:    fmt.Sprintf("Potential XSS: Unescaped content passed to HTML rendering function '%s'", call.Name),
					Location:   "function call",
					Suggestion: SuggestHTMLEscape(arg),
					Expr:       call,
				})
			}
		}
	}

	// Check for safe escaping functions
	safeEscapeFunctions := map[string]bool{
		"escapeHTML": true,
		"htmlEscape": true,
		"sanitize":   true,
		"escapeJS":   true,
	}

	if !safeEscapeFunctions[call.Name] {
		// Not using a safe escape function, analyze arguments
		for _, arg := range call.Args {
			d.analyzeExpr(arg, inHTMLContext)
		}
	}
}

// analyzeObjectExpr checks object expressions for HTML content in response objects
func (d *XSSDetector) analyzeObjectExpr(obj interpreter.ObjectExpr) {
	for _, field := range obj.Fields {
		// Check for Content-Type field
		if strings.EqualFold(field.Key, "content-type") || strings.EqualFold(field.Key, "contentType") {
			if strLit, ok := field.Value.(interpreter.LiteralExpr); ok {
				if lit, ok := strLit.Value.(interpreter.StringLiteral); ok {
					if strings.Contains(strings.ToLower(lit.Value), "text/html") {
						d.addWarning(SecurityWarning{
							Type:       "XSS",
							Severity:   "MEDIUM",
							Message:    "HTML content type detected - ensure all dynamic content is properly escaped",
							Location:   "object field: " + field.Key,
							Suggestion: "Use JSON responses (application/json) when possible, or ensure HTML escaping",
							Expr:       obj,
						})
					}
				}
			}
		}

		// Check for body/html/content fields
		if isHTMLContentField(field.Key) {
			if RequiresHTMLEscape(field.Value) {
				d.addWarning(SecurityWarning{
					Type:       "XSS",
					Severity:   "HIGH",
					Message:    fmt.Sprintf("Potential XSS: Unescaped content in HTML field '%s'", field.Key),
					Location:   "object field: " + field.Key,
					Suggestion: SuggestHTMLEscape(field.Value),
					Expr:       obj,
				})
			}
			d.analyzeExpr(field.Value, true)
		} else {
			d.analyzeExpr(field.Value, false)
		}
	}
}

// RequiresHTMLEscape checks if an expression requires HTML escaping
func RequiresHTMLEscape(expr interpreter.Expr) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case interpreter.VariableExpr:
		return isUserInputSource(e.Name)

	case interpreter.FieldAccessExpr:
		if varExpr, ok := e.Object.(interpreter.VariableExpr); ok {
			return isUserInputSource(varExpr.Name)
		}
		return RequiresHTMLEscape(e.Object)

	case interpreter.BinaryOpExpr:
		// If any part contains user input, the whole expression needs escaping
		return RequiresHTMLEscape(e.Left) || RequiresHTMLEscape(e.Right)

	case interpreter.FunctionCallExpr:
		// Check if it's already an escape function
		safeEscapeFunctions := map[string]bool{
			"escapeHTML": true,
			"htmlEscape": true,
			"sanitize":   true,
			"escapeJS":   true,
		}
		if safeEscapeFunctions[e.Name] {
			return false
		}
		// Check if any argument needs escaping
		for _, arg := range e.Args {
			if RequiresHTMLEscape(arg) {
				return true
			}
		}
		return false

	case interpreter.ArrayExpr:
		for _, elem := range e.Elements {
			if RequiresHTMLEscape(elem) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// SuggestHTMLEscape generates a suggestion for HTML escaping
func SuggestHTMLEscape(expr interpreter.Expr) string {
	exprStr := exprToString(expr)
	return fmt.Sprintf("Use escapeHTML(%s) to prevent XSS attacks", exprStr)
}

// EscapeHTML escapes HTML special characters.
// Ampersand must be replaced first to avoid double-escaping.
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// EscapeJS escapes characters for JavaScript context.
// Backslash must be replaced first to avoid double-escaping.
func EscapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "<", "\\u003C")
	s = strings.ReplaceAll(s, ">", "\\u003E")
	s = strings.ReplaceAll(s, "&", "\\u0026")
	return s
}

// Helper functions

func (d *XSSDetector) addWarning(warning SecurityWarning) {
	d.warnings = append(d.warnings, warning)
}

func isUserInputSource(varName string) bool {
	userInputSources := []string{
		"input", "request", "req", "body", "params", "query",
		"form", "data", "payload", "ctx", "context",
	}

	lowerName := strings.ToLower(varName)
	for _, source := range userInputSources {
		if strings.Contains(lowerName, source) {
			return true
		}
	}
	return false
}

func isHTMLContentField(fieldName string) bool {
	htmlFields := []string{
		"html", "body", "content", "message", "description",
		"text", "markup", "template", "page",
	}

	lowerName := strings.ToLower(fieldName)
	for _, field := range htmlFields {
		if strings.Contains(lowerName, field) {
			return true
		}
	}
	return false
}

func containsHTMLTags(s string) bool {
	// Simple regex to detect HTML tags
	htmlTagPattern := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	return htmlTagPattern.MatchString(s)
}

func containsScriptTags(s string) bool {
	lowerS := strings.ToLower(s)
	return strings.Contains(lowerS, "<script") || strings.Contains(lowerS, "</script>")
}

func containsEventHandlers(s string) bool {
	eventHandlers := []string{
		"onclick", "onload", "onerror", "onmouseover", "onmouseout",
		"onfocus", "onblur", "onchange", "onsubmit", "onkeypress",
		"onkeydown", "onkeyup",
	}

	lowerS := strings.ToLower(s)
	for _, handler := range eventHandlers {
		if strings.Contains(lowerS, handler) {
			return true
		}
	}
	return false
}

func (d *XSSDetector) containsHTMLPatterns(expr interpreter.Expr) bool {
	switch e := expr.(type) {
	case interpreter.LiteralExpr:
		if strLit, ok := e.Value.(interpreter.StringLiteral); ok {
			return containsHTMLTags(strLit.Value)
		}
	case interpreter.BinaryOpExpr:
		return d.containsHTMLPatterns(e.Left) || d.containsHTMLPatterns(e.Right)
	}
	return false
}

func (d *XSSDetector) containsUserInput(expr interpreter.Expr) bool {
	switch e := expr.(type) {
	case interpreter.VariableExpr:
		return isUserInputSource(e.Name)
	case interpreter.FieldAccessExpr:
		if varExpr, ok := e.Object.(interpreter.VariableExpr); ok {
			return isUserInputSource(varExpr.Name)
		}
		return d.containsUserInput(e.Object)
	case interpreter.BinaryOpExpr:
		return d.containsUserInput(e.Left) || d.containsUserInput(e.Right)
	}
	return false
}

func exprToString(expr interpreter.Expr) string {
	if expr == nil {
		return "nil"
	}

	switch e := expr.(type) {
	case interpreter.LiteralExpr:
		if strLit, ok := e.Value.(interpreter.StringLiteral); ok {
			return fmt.Sprintf("\"%s\"", strLit.Value)
		}
		if intLit, ok := e.Value.(interpreter.IntLiteral); ok {
			return fmt.Sprintf("%d", intLit.Value)
		}
		if boolLit, ok := e.Value.(interpreter.BoolLiteral); ok {
			return fmt.Sprintf("%t", boolLit.Value)
		}
		if floatLit, ok := e.Value.(interpreter.FloatLiteral); ok {
			return fmt.Sprintf("%f", floatLit.Value)
		}

	case interpreter.VariableExpr:
		return e.Name

	case interpreter.FieldAccessExpr:
		return fmt.Sprintf("%s.%s", exprToString(e.Object), e.Field)

	case interpreter.BinaryOpExpr:
		return fmt.Sprintf("(%s %s %s)", exprToString(e.Left), e.Op.String(), exprToString(e.Right))

	case interpreter.FunctionCallExpr:
		args := make([]string, len(e.Args))
		for i, arg := range e.Args {
			args[i] = exprToString(arg)
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))

	case interpreter.ArrayExpr:
		elements := make([]string, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = exprToString(elem)
		}
		return fmt.Sprintf("[%s]", strings.Join(elements, ", "))

	case interpreter.ObjectExpr:
		return "{...}"
	}

	return "unknown"
}
