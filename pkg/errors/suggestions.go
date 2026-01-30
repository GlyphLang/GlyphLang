package errors

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// SuggestionConfig controls suggestion behavior
type SuggestionConfig struct {
	MaxSuggestions          int
	MaxDistance             int
	MinSimilarityScore      float64
	ShowMultipleSuggestions bool
}

// DefaultSuggestionConfig returns the default configuration
func DefaultSuggestionConfig() *SuggestionConfig {
	return &SuggestionConfig{
		MaxSuggestions:          3,
		MaxDistance:             3,
		MinSimilarityScore:      0.5,
		ShowMultipleSuggestions: true,
	}
}

// SuggestionResult contains a suggestion with its confidence score
type SuggestionResult struct {
	Suggestion string
	Distance   int
	Score      float64
}

// Common typo patterns
var commonTypos = map[string]string{
	"fucntion":  "function",
	"functoin":  "function",
	"funtion":   "function",
	"funciton":  "function",
	"retrun":    "return",
	"reutrn":    "return",
	"retur":     "return",
	"slect":     "select",
	"inser":     "insert",
	"updaet":    "update",
	"delte":     "delete",
	"whre":      "where",
	"form":      "from",
	"joim":      "join",
	"lmit":      "limit",
	"ofset":     "offset",
	"ordr":      "order",
	"grup":      "group",
	"havin":     "having",
	"treu":      "true",
	"flase":     "false",
	"nill":      "nil",
	"nul":       "null",
	"undifined": "undefined",
	"lenght":    "length",
	"strng":     "string",
	"integr":    "integer",
	"boolen":    "boolean",
	"arry":      "array",
	"ojbect":    "object",
	"rquest":    "request",
	"respone":   "response",
	"reequest":  "request",
	"resonse":   "response",
}

// SyntaxPattern represents a common syntax error pattern
type SyntaxPattern struct {
	Pattern     string
	Description string
	Suggestion  string
}

// Common syntax patterns
var syntaxPatterns = []SyntaxPattern{
	{
		Pattern:     "missing opening",
		Description: "Missing opening bracket, brace, or parenthesis",
		Suggestion:  "Check that all brackets, braces, and parentheses are properly opened",
	},
	{
		Pattern:     "missing closing",
		Description: "Missing closing bracket, brace, or parenthesis",
		Suggestion:  "Check that all brackets, braces, and parentheses are properly closed",
	},
	{
		Pattern:     "unclosed string",
		Description: "String literal is not properly closed",
		Suggestion:  "Make sure all string literals have matching quotes (\" or ')",
	},
	{
		Pattern:     "unexpected eof",
		Description: "Unexpected end of file",
		Suggestion:  "Check for unclosed blocks, missing closing brackets, or incomplete statements",
	},
}

// FindBestSuggestions finds the best matching suggestions for a given name
func FindBestSuggestions(target string, candidates []string, config *SuggestionConfig) []SuggestionResult {
	if config == nil {
		config = DefaultSuggestionConfig()
	}

	// Check for common typos first
	if correction, ok := commonTypos[strings.ToLower(target)]; ok {
		return []SuggestionResult{{
			Suggestion: correction,
			Distance:   0,
			Score:      1.0,
		}}
	}

	var results []SuggestionResult

	for _, candidate := range candidates {
		// Skip exact matches
		if candidate == target {
			continue
		}

		// Calculate similarity
		distance := levenshteinDistance(target, candidate)
		score := calculateSimilarityScore(target, candidate, distance)

		// Filter by distance and score
		if distance <= config.MaxDistance && score >= config.MinSimilarityScore {
			results = append(results, SuggestionResult{
				Suggestion: candidate,
				Distance:   distance,
				Score:      score,
			})
		}
	}

	// Sort by score (highest first), then by distance (lowest first)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Distance < results[j].Distance
	})

	// Limit results
	if len(results) > config.MaxSuggestions {
		results = results[:config.MaxSuggestions]
	}

	return results
}

// calculateSimilarityScore computes a normalized similarity score between 0 and 1
func calculateSimilarityScore(s1, s2 string, distance int) float64 {
	maxLen := max(len(s1), len(s2))
	if maxLen == 0 {
		return 1.0
	}

	// Base score from Levenshtein distance
	baseScore := 1.0 - float64(distance)/float64(maxLen)

	// Bonus for common prefixes
	prefixBonus := 0.0
	minLen := min2(len(s1), len(s2))
	for i := 0; i < minLen && i < 3; i++ {
		if strings.ToLower(string(s1[i])) == strings.ToLower(string(s2[i])) {
			prefixBonus += 0.1
		} else {
			break
		}
	}

	// Bonus for common suffixes
	suffixBonus := 0.0
	for i := 1; i <= minLen && i <= 2; i++ {
		if strings.ToLower(string(s1[len(s1)-i])) == strings.ToLower(string(s2[len(s2)-i])) {
			suffixBonus += 0.05
		} else {
			break
		}
	}

	// Bonus for substring matches
	substringBonus := 0.0
	if strings.Contains(strings.ToLower(s1), strings.ToLower(s2)) ||
		strings.Contains(strings.ToLower(s2), strings.ToLower(s1)) {
		substringBonus = 0.2
	}

	// Bonus for case-insensitive matches
	caseBonus := 0.0
	if strings.ToLower(s1) == strings.ToLower(s2) {
		caseBonus = 0.3
	}

	totalScore := baseScore + prefixBonus + suffixBonus + substringBonus + caseBonus

	// Cap at 1.0
	if totalScore > 1.0 {
		totalScore = 1.0
	}

	return totalScore
}

// FormatSuggestions formats suggestion results into a human-readable string
func FormatSuggestions(results []SuggestionResult, multipleAllowed bool) string {
	if len(results) == 0 {
		return ""
	}

	if len(results) == 1 {
		return fmt.Sprintf("Did you mean '%s'?", results[0].Suggestion)
	}

	if !multipleAllowed {
		return fmt.Sprintf("Did you mean '%s'?", results[0].Suggestion)
	}

	var suggestions []string
	for _, r := range results {
		suggestions = append(suggestions, fmt.Sprintf("'%s'", r.Suggestion))
	}

	if len(suggestions) == 2 {
		return fmt.Sprintf("Did you mean %s or %s?", suggestions[0], suggestions[1])
	}

	lastIdx := len(suggestions) - 1
	return fmt.Sprintf("Did you mean %s, or %s?",
		strings.Join(suggestions[:lastIdx], ", "),
		suggestions[lastIdx])
}

// GetVariableSuggestion suggests variable names with enhanced fuzzy matching
func GetVariableSuggestion(varName string, availableVars []string) string {
	config := DefaultSuggestionConfig()
	results := FindBestSuggestions(varName, availableVars, config)

	if len(results) > 0 {
		suggestion := FormatSuggestions(results, config.ShowMultipleSuggestions)
		return fmt.Sprintf("%s Or make sure to define the variable with '$ %s = value'",
			suggestion, varName)
	}

	return fmt.Sprintf("Make sure to define the variable before using it: $ %s = value", varName)
}

// GetFunctionSuggestion suggests function names
func GetFunctionSuggestion(funcName string, availableFuncs []string) string {
	config := DefaultSuggestionConfig()
	results := FindBestSuggestions(funcName, availableFuncs, config)

	if len(results) > 0 {
		return FormatSuggestions(results, config.ShowMultipleSuggestions)
	}

	return fmt.Sprintf("Function '%s' is not defined. Check the function name or define it.", funcName)
}

// GetTypeSuggestion suggests type names
func GetTypeSuggestion(typeName string, availableTypes []string) string {
	// Add built-in types
	builtInTypes := []string{"int", "str", "bool", "float", "array", "object", "null"}
	allTypes := append(builtInTypes, availableTypes...)

	config := DefaultSuggestionConfig()
	results := FindBestSuggestions(typeName, allTypes, config)

	if len(results) > 0 {
		return FormatSuggestions(results, config.ShowMultipleSuggestions)
	}

	return fmt.Sprintf("Unknown type '%s'. Valid types are: int, str, bool, float, array, object", typeName)
}

// GetRouteSuggestion suggests route paths
func GetRouteSuggestion(routePath string, availableRoutes []string) string {
	config := DefaultSuggestionConfig()
	config.MaxDistance = 5 // Allow more distance for route paths
	results := FindBestSuggestions(routePath, availableRoutes, config)

	if len(results) > 0 {
		return FormatSuggestions(results, config.ShowMultipleSuggestions)
	}

	return fmt.Sprintf("Route '%s' is not defined. Check the route path.", routePath)
}

// DetectMissingBracket detects missing brackets, braces, or parentheses
func DetectMissingBracket(source string, line, column int) string {
	lines := strings.Split(source, "\n")
	if line <= 0 || line > len(lines) {
		return ""
	}

	// Count brackets up to the error position
	openBrackets := 0
	openBraces := 0
	openParens := 0

	for i := 0; i < line; i++ {
		lineText := lines[i]
		for _, ch := range lineText {
			switch ch {
			case '[':
				openBrackets++
			case ']':
				openBrackets--
			case '{':
				openBraces++
			case '}':
				openBraces--
			case '(':
				openParens++
			case ')':
				openParens--
			}
		}
	}

	// Check the error line up to the column
	if line > 0 && column > 0 {
		lineText := lines[line-1]
		for i := 0; i < column && i < len(lineText); i++ {
			switch lineText[i] {
			case '[':
				openBrackets++
			case ']':
				openBrackets--
			case '{':
				openBraces++
			case '}':
				openBraces--
			case '(':
				openParens++
			case ')':
				openParens--
			}
		}
	}

	var suggestions []string
	if openBrackets > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Missing %d closing bracket(s) ']'", openBrackets))
	} else if openBrackets < 0 {
		suggestions = append(suggestions, fmt.Sprintf("Unexpected closing bracket ']' (no matching opening bracket)"))
	}

	if openBraces > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Missing %d closing brace(s) '}'", openBraces))
	} else if openBraces < 0 {
		suggestions = append(suggestions, fmt.Sprintf("Unexpected closing brace '}' (no matching opening brace)"))
	}

	if openParens > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Missing %d closing parenthesis(es) ')'", openParens))
	} else if openParens < 0 {
		suggestions = append(suggestions, fmt.Sprintf("Unexpected closing parenthesis ')' (no matching opening parenthesis)"))
	}

	if len(suggestions) > 0 {
		return strings.Join(suggestions, "; ")
	}

	return ""
}

// DetectUnclosedString detects unclosed string literals
func DetectUnclosedString(source string, line int) string {
	lines := strings.Split(source, "\n")
	if line <= 0 || line > len(lines) {
		return ""
	}

	lineText := lines[line-1]

	// Track quote states
	var openQuote rune
	escaped := false

	for i, ch := range lineText {
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' || ch == '\'' {
			if openQuote == 0 {
				openQuote = ch
			} else if openQuote == ch {
				openQuote = 0
			}
		}

		// Check if we reached end of line with open quote
		if i == len(lineText)-1 && openQuote != 0 {
			return fmt.Sprintf("Unclosed string literal (missing closing %c)", openQuote)
		}
	}

	if openQuote != 0 {
		return fmt.Sprintf("Unclosed string literal (missing closing %c)", openQuote)
	}

	return ""
}

// GetTypeMismatchSuggestion provides enhanced type mismatch suggestions
func GetTypeMismatchSuggestion(expected, actual, context string) string {
	// Build base message
	var suggestion strings.Builder

	// Type-specific suggestions
	switch {
	case expected == "int" && actual == "string":
		suggestion.WriteString("Convert the string to an integer using parseInt() or ensure the value is numeric")
	case expected == "int" && actual == "float":
		suggestion.WriteString("The value is a float but an integer is expected. Consider rounding or truncating the value")
	case expected == "string" && actual == "int":
		suggestion.WriteString("Convert the integer to a string using toString() or string concatenation")
	case expected == "string" && actual == "float":
		suggestion.WriteString("Convert the float to a string using toString() or string formatting")
	case expected == "bool" && actual == "int":
		suggestion.WriteString("Use a boolean value (true or false) instead of an integer, or convert using a comparison (e.g., value != 0)")
	case expected == "bool" && actual == "string":
		suggestion.WriteString("Use a boolean value (true or false) instead of a string, or check if string is non-empty")
	case expected == "float" && actual == "int":
		suggestion.WriteString("The integer will be automatically converted to float, but you can make it explicit")
	case expected == "array" && actual != "array":
		suggestion.WriteString("Wrap the value in square brackets [] to create an array, or ensure the expression returns an array")
	case expected == "object" && actual != "object":
		suggestion.WriteString("Use curly braces {} to create an object, or ensure the expression returns an object")
	default:
		suggestion.WriteString(fmt.Sprintf("Expected type '%s' but got '%s'", expected, actual))
		if context != "" {
			suggestion.WriteString(fmt.Sprintf(" in %s", context))
		}
		suggestion.WriteString(". Check your type annotations or the value being assigned")
	}

	return suggestion.String()
}

// GetRuntimeSuggestion provides context-aware runtime error suggestions
func GetRuntimeSuggestion(errorType string, context map[string]interface{}) string {
	switch errorType {
	case "division_by_zero":
		return "Add a check to ensure the divisor is not zero before dividing: if (divisor != 0) { ... }"

	case "null_reference":
		varName := context["variable"]
		if varName != nil {
			return fmt.Sprintf("Variable '%v' is null. Check if it has been initialized or add a null check: if (%v != null) { ... }", varName, varName)
		}
		return "Check if the variable is null before accessing its properties or methods"

	case "index_out_of_bounds":
		index := context["index"]
		length := context["length"]
		if index != nil && length != nil {
			return fmt.Sprintf("Array index %v is out of bounds (array length: %v). Valid indices are 0 to %v",
				index, length, context["length"].(int)-1)
		}
		return "Check that the array index is within bounds before accessing: if (index >= 0 && index < array.length) { ... }"

	case "type_error":
		return "The operation is not supported for this type. Check the types of your operands"

	case "undefined_property":
		propName := context["property"]
		if propName != nil {
			return fmt.Sprintf("Property '%v' does not exist. Check the property name or add it to the object", propName)
		}
		return "The property does not exist on this object. Check the property name"

	case "sql_injection":
		return "Potential SQL injection vulnerability detected. Use parameterized queries or prepared statements instead of string concatenation"

	case "xss":
		return "Potential XSS vulnerability detected. Sanitize user input before rendering it in HTML"

	default:
		return "Check your code for potential issues and ensure all values are properly initialized"
	}
}

// FormatCodeSnippetWithFix formats source code with a suggested fix
func FormatCodeSnippetWithFix(source string, line, column int, fixedLine string) string {
	lines := strings.Split(source, "\n")
	if line <= 0 || line > len(lines) {
		return ""
	}

	var builder strings.Builder
	lineNum := line

	// Show previous line for context
	if lineNum > 1 {
		prevLineNum := lineNum - 1
		builder.WriteString(fmt.Sprintf("  %s%4d |%s %s\n", Gray, prevLineNum, Reset, lines[prevLineNum-1]))
	}

	// Show the error line
	errorLine := lines[lineNum-1]
	builder.WriteString(fmt.Sprintf("  %s%4d |%s %s\n", Red, lineNum, Reset, errorLine))

	// Show caret pointing to error column
	if column > 0 {
		spaces := strings.Repeat(" ", column-1)
		builder.WriteString(fmt.Sprintf("       %s|%s %s%s^ error here%s\n", Gray, Reset, Red, spaces, Reset))
	}

	// Show the suggested fix
	if fixedLine != "" {
		builder.WriteString(fmt.Sprintf("  %s%4d |%s %s %s(suggested fix)%s\n",
			Green, lineNum, Reset, fixedLine, Gray, Reset))
	}

	// Show next line for context
	if lineNum < len(lines) {
		builder.WriteString(fmt.Sprintf("  %s%4d |%s %s\n", Gray, lineNum+1, Reset, lines[lineNum]))
	}

	return builder.String()
}

// SuggestMissingSemicolon checks if a semicolon might be missing (if language requires it)
func SuggestMissingSemicolon(source string, line int) bool {
	lines := strings.Split(source, "\n")
	if line <= 0 || line > len(lines) {
		return false
	}

	lineText := strings.TrimSpace(lines[line-1])

	// Check if line ends with certain keywords or patterns that might need semicolon
	needsSemicolon := []string{"return", "break", "continue", "="}

	for _, pattern := range needsSemicolon {
		if strings.HasPrefix(lineText, pattern) || strings.Contains(lineText, pattern) {
			// Check if it already ends with semicolon, comma, or brace
			if !strings.HasSuffix(lineText, ";") &&
				!strings.HasSuffix(lineText, ",") &&
				!strings.HasSuffix(lineText, "{") &&
				!strings.HasSuffix(lineText, "}") {
				return true
			}
		}
	}

	return false
}

// DetectCommonSyntaxErrors detects common syntax error patterns
func DetectCommonSyntaxErrors(source string, line int, errorMsg string) string {
	errorMsgLower := strings.ToLower(errorMsg)

	// Check for bracket/brace/paren issues
	if strings.Contains(errorMsgLower, "expect") &&
		(strings.Contains(errorMsgLower, "}") ||
			strings.Contains(errorMsgLower, "]") ||
			strings.Contains(errorMsgLower, ")")) {
		return DetectMissingBracket(source, line, 0)
	}

	// Check for string issues
	if strings.Contains(errorMsgLower, "string") ||
		strings.Contains(errorMsgLower, "unterminated") {
		return DetectUnclosedString(source, line)
	}

	// Check for common patterns
	for _, pattern := range syntaxPatterns {
		if strings.Contains(errorMsgLower, pattern.Pattern) {
			return pattern.Suggestion
		}
	}

	return ""
}

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsValidIdentifier checks if a string is a valid identifier
func IsValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// First character must be letter or underscore
	first := rune(s[0])
	if !unicode.IsLetter(first) && first != '_' {
		return false
	}

	// Remaining characters must be letter, digit, or underscore
	for _, ch := range s[1:] {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}

	return true
}

// SuggestValidIdentifier suggests corrections to make an invalid identifier valid
func SuggestValidIdentifier(s string) string {
	if len(s) == 0 {
		return "Identifier cannot be empty. Use a letter or underscore to start"
	}

	// Check first character
	first := rune(s[0])
	if unicode.IsDigit(first) {
		return fmt.Sprintf("Identifier cannot start with a digit. Try '_%s' or use a letter", s)
	}

	// Check for special characters
	hasSpecial := false
	for _, ch := range s {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			hasSpecial = true
			break
		}
	}

	if hasSpecial {
		// Suggest removing special characters
		var cleaned strings.Builder
		for _, ch := range s {
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
				cleaned.WriteRune(ch)
			}
		}
		return fmt.Sprintf("Remove special characters. Try '%s'", cleaned.String())
	}

	return "Use only letters, digits, and underscores in identifiers"
}
