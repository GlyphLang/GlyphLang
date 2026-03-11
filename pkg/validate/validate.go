// Package validate provides AI-friendly validation for Glyph source files.
// It returns structured errors that AI agents can easily parse and act upon.
package validate

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
)

// ValidationResult contains the results of validating a Glyph file
type ValidationResult struct {
	Valid    bool               `json:"valid"`
	FilePath string             `json:"file_path"`
	Errors   []*ValidationError `json:"errors,omitempty"`
	Warnings []*ValidationError `json:"warnings,omitempty"`
	Stats    *ValidationStats   `json:"stats,omitempty"`
}

// ValidationStats contains statistics about the validated file
type ValidationStats struct {
	Types     int `json:"types"`
	Routes    int `json:"routes"`
	Functions int `json:"functions"`
	Commands  int `json:"commands"`
	Lines     int `json:"lines"`
}

// ValidationError represents a single validation error with context
type ValidationError struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Location  *Location `json:"location,omitempty"`
	FixHint   string    `json:"fix_hint,omitempty"`
	Context   string    `json:"context,omitempty"`
	Severity  string    `json:"severity"` // "error" or "warning"
	RelatedTo string    `json:"related_to,omitempty"`
}

// Location represents a source code location
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// ErrorType constants for structured error identification
const (
	ErrTypeSyntax       = "syntax_error"
	ErrTypeLexer        = "lexer_error"
	ErrTypeUndefined    = "undefined_reference"
	ErrTypeMismatch     = "type_mismatch"
	ErrTypeDuplicate    = "duplicate_definition"
	ErrTypeMissing      = "missing_required"
	ErrTypeUnused       = "unused_definition"
	ErrTypeDeprecated   = "deprecated_usage"
	ErrTypeInvalidRoute = "invalid_route"
	ErrTypeInvalidType  = "invalid_type"
)

// Validator validates Glyph source code
type Validator struct {
	source   string
	filePath string
	lines    []string
}

// NewValidator creates a new validator for the given source
func NewValidator(source, filePath string) *Validator {
	return &Validator{
		source:   source,
		filePath: filePath,
		lines:    strings.Split(source, "\n"),
	}
}

// Validate performs full validation and returns structured results
func (v *Validator) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		FilePath: v.filePath,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]*ValidationError, 0),
		Stats: &ValidationStats{
			Lines: len(v.lines),
		},
	}

	// Phase 1: Lexical analysis
	lexer := parser.NewLexer(v.source)
	tokens, err := lexer.Tokenize()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, v.createLexerError(err))
		return result
	}

	// Phase 2: Parsing
	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, v.createParseError(err))
		return result
	}

	// Phase 3: Semantic validation
	v.validateSemantics(module, result)

	// Update stats
	v.collectStats(module, result.Stats)

	return result
}

// createLexerError creates a structured error from a lexer error
func (v *Validator) createLexerError(err error) *ValidationError {
	errStr := err.Error()
	line, col := v.extractLocation(errStr)

	return &ValidationError{
		Type:     ErrTypeLexer,
		Message:  errStr,
		Severity: "error",
		Location: &Location{
			File:   v.filePath,
			Line:   line,
			Column: col,
		},
		Context: v.getLineContext(line),
		FixHint: v.suggestLexerFix(errStr),
	}
}

// createParseError creates a structured error from a parser error
func (v *Validator) createParseError(err error) *ValidationError {
	errStr := err.Error()
	line, col := v.extractLocation(errStr)

	return &ValidationError{
		Type:     ErrTypeSyntax,
		Message:  errStr,
		Severity: "error",
		Location: &Location{
			File:   v.filePath,
			Line:   line,
			Column: col,
		},
		Context: v.getLineContext(line),
		FixHint: v.suggestParseFix(errStr),
	}
}

// validateSemantics performs semantic validation on the AST
func (v *Validator) validateSemantics(module *ast.Module, result *ValidationResult) {
	// Collect all defined types for reference checking
	definedTypes := make(map[string]bool)
	builtinTypes := map[string]bool{
		"int": true, "str": true, "string": true, "bool": true,
		"float": true, "timestamp": true, "any": true, "object": true,
		"List": true, "Map": true, "Result": true,
		"Database": true, "Redis": true, "MongoDB": true, "LLM": true,
	}

	// Process imports to collect types from imported modules
	v.processImports(module, definedTypes, result)

	// First pass: collect all type and provider definitions
	definedProviders := make(map[string]bool)
	for _, item := range module.Items {
		switch it := item.(type) {
		case *ast.TypeDef:
			if definedTypes[it.Name] {
				result.Errors = append(result.Errors, &ValidationError{
					Type:      ErrTypeDuplicate,
					Message:   fmt.Sprintf("duplicate type definition: %s", it.Name),
					Severity:  "error",
					RelatedTo: it.Name,
					FixHint:   fmt.Sprintf("rename one of the '%s' type definitions or remove the duplicate", it.Name),
				})
				result.Valid = false
			}
			definedTypes[it.Name] = true
		case *ast.ProviderDef:
			if definedProviders[it.Name] {
				result.Errors = append(result.Errors, &ValidationError{
					Type:      ErrTypeDuplicate,
					Message:   fmt.Sprintf("duplicate provider definition: %s", it.Name),
					Severity:  "error",
					RelatedTo: it.Name,
					FixHint:   fmt.Sprintf("rename one of the '%s' provider definitions or remove the duplicate", it.Name),
				})
				result.Valid = false
			}
			definedProviders[it.Name] = true
			// Provider names are valid types for injection
			definedTypes[it.Name] = true
		}
	}

	// Second pass: validate type references
	for _, item := range module.Items {
		switch node := item.(type) {
		case *ast.TypeDef:
			v.validateTypeFields(node, definedTypes, builtinTypes, result)
		case *ast.Route:
			// Routes receive definedProviders to validate injection types.
			// validateFunction does not need it since functions don't have injections.
			v.validateRoute(node, definedTypes, builtinTypes, definedProviders, result)
		case *ast.Function:
			v.validateFunction(node, definedTypes, builtinTypes, result)
		case *ast.ProviderDef:
			// validateProvider checks method param/return types only,
			// so it doesn't need definedProviders.
			v.validateProvider(node, definedTypes, builtinTypes, result)
		}
	}

	// Check for common issues
	v.checkCommonIssues(module, result)
}

// processImports processes import statements and adds imported types to the defined types map
func (v *Validator) processImports(module *ast.Module, definedTypes map[string]bool, result *ValidationResult) {
	// Get the base path for resolving relative imports
	basePath := filepath.Dir(v.filePath)

	// Create a module resolver
	resolver := interpreter.NewModuleResolver()
	resolver.AddSearchPath(basePath)

	// Set up the parse function for the resolver
	resolver.SetParseFunc(func(source string) (*ast.Module, error) {
		lexer := parser.NewLexer(source)
		tokens, err := lexer.Tokenize()
		if err != nil {
			return nil, err
		}
		p := parser.NewParser(tokens)
		return p.Parse()
	})

	// Process imports
	imports, err := resolver.ProcessImports(module, basePath)
	if err != nil {
		// Add a warning but don't fail validation - the module might still work at runtime
		result.Warnings = append(result.Warnings, &ValidationError{
			Type:     ErrTypeUndefined,
			Message:  fmt.Sprintf("failed to resolve imports: %s", err.Error()),
			Severity: "warning",
			FixHint:  "check that imported modules exist and are accessible",
		})
		return
	}

	// Add imported types to the defined types map with their aliases
	for alias, loadedModule := range imports {
		for name, item := range loadedModule.Exports {
			if _, isType := item.(*ast.TypeDef); isType {
				// Add the type with the alias prefix (e.g., "m.User")
				qualifiedName := alias + "." + name
				definedTypes[qualifiedName] = true
			}
		}
	}

	// Also process selective imports (from "./module" import { Type1, Type2 })
	for _, item := range module.Items {
		if importStmt, ok := item.(*ast.ImportStatement); ok && importStmt.Selective {
			// For selective imports, the imported names are used directly without prefix
			if loadedModule, exists := imports[importStmt.Path]; exists {
				for _, importName := range importStmt.Names {
					if item, ok := loadedModule.Exports[importName.Name]; ok {
						if _, isType := item.(*ast.TypeDef); isType {
							// Use alias if provided, otherwise use original name
							name := importName.Name
							if importName.Alias != "" {
								name = importName.Alias
							}
							definedTypes[name] = true
						}
					}
				}
			}
		}
	}
}

// validateTypeFields validates field types in a type definition
func (v *Validator) validateTypeFields(typeDef *ast.TypeDef, defined, builtin map[string]bool, result *ValidationResult) {
	for _, field := range typeDef.Fields {
		v.validateTypeRef(field.TypeAnnotation, defined, builtin, result, typeDef.Name)
	}
}

// validateRoute validates a route definition
func (v *Validator) validateRoute(route *ast.Route, defined, builtin, providers map[string]bool, result *ValidationResult) {
	// Validate return type
	if route.ReturnType != nil {
		v.validateTypeRef(route.ReturnType, defined, builtin, result, fmt.Sprintf("route %s %s", route.Method, route.Path))
	}

	// Validate path
	if !strings.HasPrefix(route.Path, "/") {
		result.Errors = append(result.Errors, &ValidationError{
			Type:      ErrTypeInvalidRoute,
			Message:   fmt.Sprintf("route path must start with /: %s", route.Path),
			Severity:  "error",
			RelatedTo: route.Path,
			FixHint:   fmt.Sprintf("change path to '/%s'", strings.TrimPrefix(route.Path, "/")),
		})
		result.Valid = false
	}

	// Check for duplicate path parameters
	params := make(map[string]bool)
	parts := strings.Split(route.Path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			param := part[1:]
			if params[param] {
				result.Warnings = append(result.Warnings, &ValidationError{
					Type:      ErrTypeDuplicate,
					Message:   fmt.Sprintf("duplicate path parameter: %s", param),
					Severity:  "warning",
					RelatedTo: route.Path,
					FixHint:   "use unique names for path parameters",
				})
			}
			params[param] = true
		}
	}

	// Validate injected provider types
	for _, inj := range route.Injections {
		provType := resolveProviderTypeName(inj.Type)
		if provType != "" && !providers[provType] && !isBuiltinProvider(provType) {
			result.Errors = append(result.Errors, &ValidationError{
				Type:      ErrTypeUndefined,
				Message:   fmt.Sprintf("undefined provider type: %s", provType),
				Severity:  "error",
				RelatedTo: fmt.Sprintf("route %s %s", route.Method, route.Path),
				FixHint:   fmt.Sprintf("define 'provider %s { ... }' or use a builtin provider (Database, Redis, MongoDB, LLM)", provType),
			})
			result.Valid = false
		}
	}
}

// validateProvider validates a provider definition's method types
func (v *Validator) validateProvider(prov *ast.ProviderDef, defined, builtin map[string]bool, result *ValidationResult) {
	for _, method := range prov.Methods {
		context := fmt.Sprintf("provider %s method %s", prov.Name, method.Name)
		if method.ReturnType != nil {
			v.validateTypeRef(method.ReturnType, defined, builtin, result, context)
		}
		for _, param := range method.Params {
			v.validateTypeRef(param.TypeAnnotation, defined, builtin, result, context)
		}
	}
}

// resolveProviderTypeName extracts the provider type name from an AST type
func resolveProviderTypeName(t ast.Type) string {
	switch typ := t.(type) {
	case ast.DatabaseType:
		return "Database"
	case ast.RedisType:
		return "Redis"
	case ast.MongoDBType:
		return "MongoDB"
	case ast.LLMType:
		return "LLM"
	case ast.NamedType:
		return typ.Name
	default:
		return ""
	}
}

// isBuiltinProvider returns true for the four standard provider types
func isBuiltinProvider(name string) bool {
	switch name {
	case "Database", "Redis", "MongoDB", "LLM":
		return true
	default:
		return false
	}
}

// validateFunction validates a function definition
func (v *Validator) validateFunction(fn *ast.Function, defined, builtin map[string]bool, result *ValidationResult) {
	// Validate return type
	if fn.ReturnType != nil {
		v.validateTypeRef(fn.ReturnType, defined, builtin, result, fmt.Sprintf("function %s", fn.Name))
	}

	// Validate parameter types
	for _, param := range fn.Params {
		v.validateTypeRef(param.TypeAnnotation, defined, builtin, result, fmt.Sprintf("function %s parameter %s", fn.Name, param.Name))
	}
}

// validateTypeRef validates a type reference
func (v *Validator) validateTypeRef(t ast.Type, defined, builtin map[string]bool, result *ValidationResult, context string) {
	if t == nil {
		return
	}

	switch typ := t.(type) {
	case ast.NamedType:
		if !defined[typ.Name] && !builtin[typ.Name] {
			result.Errors = append(result.Errors, &ValidationError{
				Type:      ErrTypeUndefined,
				Message:   fmt.Sprintf("undefined type: %s", typ.Name),
				Severity:  "error",
				RelatedTo: context,
				FixHint:   fmt.Sprintf("define type '%s' or check for typos", typ.Name),
			})
			result.Valid = false
		}
	case ast.ArrayType:
		v.validateTypeRef(typ.ElementType, defined, builtin, result, context)
	case ast.OptionalType:
		v.validateTypeRef(typ.InnerType, defined, builtin, result, context)
	case ast.GenericType:
		v.validateTypeRef(typ.BaseType, defined, builtin, result, context)
		for _, arg := range typ.TypeArgs {
			v.validateTypeRef(arg, defined, builtin, result, context)
		}
	}
}

// checkCommonIssues checks for common coding issues
func (v *Validator) checkCommonIssues(module *ast.Module, result *ValidationResult) {
	routePaths := make(map[string]bool)

	for _, item := range module.Items {
		if route, ok := item.(*ast.Route); ok {
			key := fmt.Sprintf("%s %s", route.Method, route.Path)
			if routePaths[key] {
				result.Errors = append(result.Errors, &ValidationError{
					Type:      ErrTypeDuplicate,
					Message:   fmt.Sprintf("duplicate route: %s", key),
					Severity:  "error",
					RelatedTo: key,
					FixHint:   "remove duplicate route or change the path/method",
				})
				result.Valid = false
			}
			routePaths[key] = true
		}
	}
}

// collectStats collects statistics about the module
func (v *Validator) collectStats(module *ast.Module, stats *ValidationStats) {
	for _, item := range module.Items {
		switch item.(type) {
		case *ast.TypeDef:
			stats.Types++
		case *ast.Route:
			stats.Routes++
		case *ast.Function:
			stats.Functions++
		case *ast.Command:
			stats.Commands++
		}
	}
}

// extractLocation attempts to extract line/column from error message
func (v *Validator) extractLocation(errStr string) (int, int) {
	// Try to parse "line X" or "at line X"
	line := 1
	col := 1

	// Common patterns in error messages
	if idx := strings.Index(errStr, "line "); idx != -1 {
		fmt.Sscanf(errStr[idx:], "line %d", &line)
	}
	if idx := strings.Index(errStr, "column "); idx != -1 {
		fmt.Sscanf(errStr[idx:], "column %d", &col)
	}

	return line, col
}

// getLineContext returns the source line for context
func (v *Validator) getLineContext(line int) string {
	if line < 1 || line > len(v.lines) {
		return ""
	}
	return strings.TrimSpace(v.lines[line-1])
}

// suggestLexerFix suggests fixes for common lexer errors
func (v *Validator) suggestLexerFix(errStr string) string {
	errLower := strings.ToLower(errStr)

	if strings.Contains(errLower, "unterminated string") {
		return "add closing quote to string literal"
	}
	if strings.Contains(errLower, "unexpected character") {
		return "check for invalid characters or typos"
	}
	if strings.Contains(errLower, "invalid number") {
		return "check number format (e.g., 123, 3.14)"
	}

	return "check syntax near the error location"
}

// suggestParseFix suggests fixes for common parser errors
func (v *Validator) suggestParseFix(errStr string) string {
	errLower := strings.ToLower(errStr)

	if strings.Contains(errLower, "expected") {
		// Extract what was expected
		if strings.Contains(errLower, "expected '{'") {
			return "add opening brace '{' after type or route declaration"
		}
		if strings.Contains(errLower, "expected '}'") {
			return "add closing brace '}' to complete the block"
		}
		if strings.Contains(errLower, "expected ':'") {
			return "add colon ':' between field name and type"
		}
		if strings.Contains(errLower, "expected identifier") {
			return "add a valid name (letters, numbers, underscores)"
		}
	}

	if strings.Contains(errLower, "unexpected token") {
		return "remove unexpected token or check syntax"
	}
	if strings.Contains(errLower, "unexpected end") {
		return "complete the statement or block"
	}

	return "review Glyph syntax documentation"
}

// ToJSON serializes the result to JSON
func (r *ValidationResult) ToJSON(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(r, "", "  ")
	}
	return json.Marshal(r)
}

// ToHuman generates human-readable output
func (r *ValidationResult) ToHuman() string {
	var sb strings.Builder

	if r.Valid {
		sb.WriteString(fmt.Sprintf("✓ %s is valid\n", r.FilePath))
		sb.WriteString(fmt.Sprintf("  %d types, %d routes, %d functions, %d commands\n",
			r.Stats.Types, r.Stats.Routes, r.Stats.Functions, r.Stats.Commands))
	} else {
		sb.WriteString(fmt.Sprintf("✗ %s has errors\n\n", r.FilePath))
	}

	for _, err := range r.Errors {
		sb.WriteString(fmt.Sprintf("ERROR [%s]: %s\n", err.Type, err.Message))
		if err.Location != nil {
			sb.WriteString(fmt.Sprintf("  at %s:%d:%d\n", err.Location.File, err.Location.Line, err.Location.Column))
		}
		if err.Context != "" {
			sb.WriteString(fmt.Sprintf("  > %s\n", err.Context))
		}
		if err.FixHint != "" {
			sb.WriteString(fmt.Sprintf("  hint: %s\n", err.FixHint))
		}
		sb.WriteString("\n")
	}

	for _, warn := range r.Warnings {
		sb.WriteString(fmt.Sprintf("WARNING [%s]: %s\n", warn.Type, warn.Message))
		if warn.FixHint != "" {
			sb.WriteString(fmt.Sprintf("  hint: %s\n", warn.FixHint))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Summary returns a brief summary of the validation result
func (r *ValidationResult) Summary() string {
	if r.Valid {
		return fmt.Sprintf("valid: %d types, %d routes", r.Stats.Types, r.Stats.Routes)
	}
	return fmt.Sprintf("invalid: %d errors, %d warnings", len(r.Errors), len(r.Warnings))
}
