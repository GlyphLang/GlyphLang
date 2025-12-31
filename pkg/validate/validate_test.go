package validate

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	source := "@ GET /hello {\n  > {message: \"Hello\"}\n}"
	v := NewValidator(source, "test.glyph")

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.source != source {
		t.Error("source not set correctly")
	}
	if v.filePath != "test.glyph" {
		t.Error("filePath not set correctly")
	}
	if len(v.lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(v.lines))
	}
}

func TestValidateValidSource(t *testing.T) {
	source := `
: User {
  id: int!
  name: string!
}

@ GET /users/:id -> User {
  $ user = db.find(id)
  > user
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if result.Stats.Types != 1 {
		t.Errorf("expected 1 type, got %d", result.Stats.Types)
	}
	if result.Stats.Routes != 1 {
		t.Errorf("expected 1 route, got %d", result.Stats.Routes)
	}
}

func TestValidateLexerError(t *testing.T) {
	// Unterminated string
	source := `@ GET /test {
  $ message = "unterminated`

	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for lexer error")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error")
	}
	if result.Errors[0].Type != ErrTypeLexer {
		t.Errorf("expected lexer error type, got %s", result.Errors[0].Type)
	}
}

func TestValidateParserError(t *testing.T) {
	// Missing closing brace
	source := `: User {
  id: int!
  name: string!
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for parser error")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error")
	}
	if result.Errors[0].Type != ErrTypeSyntax {
		t.Errorf("expected syntax error type, got %s", result.Errors[0].Type)
	}
}

func TestValidateDuplicateType(t *testing.T) {
	source := `
: User {
  id: int!
}

: User {
  name: string!
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for duplicate type")
	}

	hasDuplicateError := false
	for _, err := range result.Errors {
		if err.Type == ErrTypeDuplicate && strings.Contains(err.Message, "User") {
			hasDuplicateError = true
			break
		}
	}
	if !hasDuplicateError {
		t.Error("expected duplicate type error")
	}
}

func TestValidateUndefinedType(t *testing.T) {
	source := `
@ GET /users -> NonExistentType {
  > {}
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for undefined type")
	}

	hasUndefinedError := false
	for _, err := range result.Errors {
		if err.Type == ErrTypeUndefined && strings.Contains(err.Message, "NonExistentType") {
			hasUndefinedError = true
			break
		}
	}
	if !hasUndefinedError {
		t.Error("expected undefined type error")
	}
}

func TestValidateInvalidRoutePath(t *testing.T) {
	source := `
@ GET users {
  > {}
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for invalid route path")
	}

	hasPathError := false
	for _, err := range result.Errors {
		if err.Type == ErrTypeInvalidRoute {
			hasPathError = true
			break
		}
	}
	if !hasPathError {
		t.Error("expected invalid route error")
	}
}

func TestValidateDuplicateRoute(t *testing.T) {
	source := `
@ GET /users {
  > {}
}

@ GET /users {
  > {}
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for duplicate route")
	}

	hasDuplicateError := false
	for _, err := range result.Errors {
		if err.Type == ErrTypeDuplicate && strings.Contains(err.Message, "route") {
			hasDuplicateError = true
			break
		}
	}
	if !hasDuplicateError {
		t.Error("expected duplicate route error")
	}
}

func TestValidateDuplicatePathParam(t *testing.T) {
	source := `
@ GET /users/:id/posts/:id {
  > {}
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	// This should produce a warning, not an error
	hasWarning := false
	for _, warn := range result.Warnings {
		if warn.Type == ErrTypeDuplicate && strings.Contains(warn.Message, "path parameter") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected duplicate path parameter warning")
	}
}

func TestValidateBuiltinTypes(t *testing.T) {
	source := `
: Response {
  count: int!
  message: str!
  flag: bool!
  value: float!
  created: timestamp!
  data: any!
}

@ GET /test -> Response {
  > {}
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateArrayType(t *testing.T) {
	source := `
: User {
  id: int!
}

: UserList {
  users: [User]!
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateOptionalType(t *testing.T) {
	source := `
: User {
  id: int!
  email: string?
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateGenericType(t *testing.T) {
	source := `
: User {
  id: int!
}

: Response {
  data: List<User>!
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateFunction(t *testing.T) {
	source := `
: User {
  id: int!
}

! getUser(id: int!): User {
  $ user = db.find(id)
  > user
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if result.Stats.Functions != 1 {
		t.Errorf("expected 1 function, got %d", result.Stats.Functions)
	}
}

func TestValidateFunctionUndefinedReturnType(t *testing.T) {
	source := `
! getUser(id: int!): NonExistent {
  > {}
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for undefined return type")
	}
}

func TestValidateFunctionUndefinedParamType(t *testing.T) {
	source := `
! process(data: UnknownType!): string {
  > "done"
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for undefined param type")
	}
}

func TestValidationResultToJSON(t *testing.T) {
	result := &ValidationResult{
		Valid:    true,
		FilePath: "test.glyph",
		Errors:   []*ValidationError{},
		Warnings: []*ValidationError{},
		Stats: &ValidationStats{
			Types:     2,
			Routes:    3,
			Functions: 1,
			Commands:  0,
			Lines:     50,
		},
	}

	// Test compact JSON
	compact, err := result.ToJSON(false)
	if err != nil {
		t.Fatalf("ToJSON(false) error: %v", err)
	}
	if strings.Contains(string(compact), "\n") {
		t.Error("compact JSON should not contain newlines")
	}

	// Test pretty JSON
	pretty, err := result.ToJSON(true)
	if err != nil {
		t.Fatalf("ToJSON(true) error: %v", err)
	}
	if !strings.Contains(string(pretty), "\n") {
		t.Error("pretty JSON should contain newlines")
	}

	// Verify JSON is valid
	var parsed ValidationResult
	if err := json.Unmarshal(compact, &parsed); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}
	if parsed.FilePath != "test.glyph" {
		t.Error("file path not preserved in JSON")
	}
}

func TestValidationResultToHuman(t *testing.T) {
	// Test valid result
	validResult := &ValidationResult{
		Valid:    true,
		FilePath: "test.glyph",
		Errors:   []*ValidationError{},
		Warnings: []*ValidationError{},
		Stats: &ValidationStats{
			Types:     2,
			Routes:    3,
			Functions: 1,
			Commands:  0,
			Lines:     50,
		},
	}

	human := validResult.ToHuman()
	if !strings.Contains(human, "is valid") {
		t.Error("valid result should contain 'is valid'")
	}
	if !strings.Contains(human, "2 types") {
		t.Error("should contain type count")
	}

	// Test invalid result
	invalidResult := &ValidationResult{
		Valid:    false,
		FilePath: "broken.glyph",
		Errors: []*ValidationError{
			{
				Type:     ErrTypeSyntax,
				Message:  "unexpected token",
				Severity: "error",
				Location: &Location{File: "broken.glyph", Line: 5, Column: 10},
				Context:  "$ foo = bar",
				FixHint:  "check syntax",
			},
		},
		Warnings: []*ValidationError{
			{
				Type:     ErrTypeUnused,
				Message:  "unused variable",
				Severity: "warning",
				FixHint:  "remove unused variable",
			},
		},
		Stats: &ValidationStats{Lines: 10},
	}

	human = invalidResult.ToHuman()
	if !strings.Contains(human, "has errors") {
		t.Error("invalid result should contain 'has errors'")
	}
	if !strings.Contains(human, "ERROR") {
		t.Error("should contain ERROR label")
	}
	if !strings.Contains(human, "WARNING") {
		t.Error("should contain WARNING label")
	}
	if !strings.Contains(human, "broken.glyph:5:10") {
		t.Error("should contain location")
	}
	if !strings.Contains(human, "hint:") {
		t.Error("should contain hint")
	}
}

func TestValidationResultSummary(t *testing.T) {
	validResult := &ValidationResult{
		Valid: true,
		Stats: &ValidationStats{Types: 2, Routes: 3},
	}

	summary := validResult.Summary()
	if !strings.Contains(summary, "valid") {
		t.Error("valid summary should contain 'valid'")
	}
	if !strings.Contains(summary, "2 types") {
		t.Error("should contain type count")
	}

	invalidResult := &ValidationResult{
		Valid:    false,
		Errors:   make([]*ValidationError, 3),
		Warnings: make([]*ValidationError, 2),
	}

	summary = invalidResult.Summary()
	if !strings.Contains(summary, "invalid") {
		t.Error("invalid summary should contain 'invalid'")
	}
	if !strings.Contains(summary, "3 errors") {
		t.Error("should contain error count")
	}
	if !strings.Contains(summary, "2 warnings") {
		t.Error("should contain warning count")
	}
}

func TestExtractLocation(t *testing.T) {
	v := NewValidator("", "test.glyph")

	tests := []struct {
		errStr   string
		wantLine int
		wantCol  int
	}{
		{"error at line 5", 5, 1},
		{"error at line 10, column 15", 10, 15},
		{"syntax error line 3 column 8", 3, 8},
		{"unknown error", 1, 1},
	}

	for _, tt := range tests {
		line, col := v.extractLocation(tt.errStr)
		if line != tt.wantLine {
			t.Errorf("extractLocation(%q) line = %d, want %d", tt.errStr, line, tt.wantLine)
		}
		if col != tt.wantCol {
			t.Errorf("extractLocation(%q) col = %d, want %d", tt.errStr, col, tt.wantCol)
		}
	}
}

func TestGetLineContext(t *testing.T) {
	source := "line 1\n  line 2  \nline 3"
	v := NewValidator(source, "test.glyph")

	tests := []struct {
		line int
		want string
	}{
		{1, "line 1"},
		{2, "line 2"},
		{3, "line 3"},
		{0, ""},  // out of range
		{10, ""}, // out of range
	}

	for _, tt := range tests {
		got := v.getLineContext(tt.line)
		if got != tt.want {
			t.Errorf("getLineContext(%d) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestSuggestLexerFix(t *testing.T) {
	v := NewValidator("", "test.glyph")

	tests := []struct {
		errStr      string
		wantContain string
	}{
		{"unterminated string literal", "closing quote"},
		{"unexpected character", "invalid characters"},
		{"invalid number format", "number format"},
		{"some random error", "syntax"},
	}

	for _, tt := range tests {
		got := v.suggestLexerFix(tt.errStr)
		if !strings.Contains(strings.ToLower(got), strings.ToLower(tt.wantContain)) {
			t.Errorf("suggestLexerFix(%q) = %q, want to contain %q", tt.errStr, got, tt.wantContain)
		}
	}
}

func TestSuggestParseFix(t *testing.T) {
	v := NewValidator("", "test.glyph")

	tests := []struct {
		errStr      string
		wantContain string
	}{
		{"expected '{'", "brace"},
		{"expected '}'", "brace"},
		{"expected ':'", "colon"},
		{"expected identifier", "name"},
		{"unexpected token at position 5", "remove"},
		{"unexpected end of file", "complete"},
		{"some random parse error", "documentation"},
	}

	for _, tt := range tests {
		got := v.suggestParseFix(tt.errStr)
		if !strings.Contains(strings.ToLower(got), strings.ToLower(tt.wantContain)) {
			t.Errorf("suggestParseFix(%q) = %q, want to contain %q", tt.errStr, got, tt.wantContain)
		}
	}
}

func TestErrorTypeConstants(t *testing.T) {
	// Verify error type constants are properly defined
	constants := []string{
		ErrTypeSyntax,
		ErrTypeLexer,
		ErrTypeUndefined,
		ErrTypeMismatch,
		ErrTypeDuplicate,
		ErrTypeMissing,
		ErrTypeUnused,
		ErrTypeDeprecated,
		ErrTypeInvalidRoute,
		ErrTypeInvalidType,
	}

	for _, c := range constants {
		if c == "" {
			t.Error("error type constant should not be empty")
		}
	}

	// Verify uniqueness
	seen := make(map[string]bool)
	for _, c := range constants {
		if seen[c] {
			t.Errorf("duplicate error type constant: %s", c)
		}
		seen[c] = true
	}
}

func TestValidateComplexFile(t *testing.T) {
	source := `
: User {
  id: int!
  name: string!
  email: string?
}

: Post {
  id: int!
  title: string!
}

: ApiResponse {
  success: bool!
}

@ GET /users {
  > []
}

@ GET /users/:id {
  > {}
}

@ POST /users {
  > {}
}
`
	v := NewValidator(source, "complex.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if result.Stats.Types != 3 {
		t.Errorf("expected 3 types, got %d", result.Stats.Types)
	}
	if result.Stats.Routes != 3 {
		t.Errorf("expected 3 routes, got %d", result.Stats.Routes)
	}
}

func TestValidateEmptySource(t *testing.T) {
	v := NewValidator("", "empty.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Error("empty source should be valid")
	}
	if result.Stats.Types != 0 {
		t.Error("empty source should have 0 types")
	}
	if result.Stats.Routes != 0 {
		t.Error("empty source should have 0 routes")
	}
}

func TestValidateCommentsOnly(t *testing.T) {
	source := `# This is a comment
# Another comment
# More comments
`
	v := NewValidator(source, "comments.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Error("comments-only source should be valid")
	}
}

func TestValidateNestedTypes(t *testing.T) {
	source := `
: Address {
  street: string!
  city: string!
}

: Company {
  name: string!
  address: Address!
}

: Employee {
  name: string!
  company: Company!
}
`
	v := NewValidator(source, "nested.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateUndefinedNestedType(t *testing.T) {
	source := `
: Employee {
  name: string!
  company: NonExistentCompany!
}
`
	v := NewValidator(source, "broken.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for undefined nested type")
	}

	hasError := false
	for _, err := range result.Errors {
		if err.Type == ErrTypeUndefined && strings.Contains(err.Message, "NonExistentCompany") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected undefined type error for NonExistentCompany")
	}
}

func TestValidationErrorFields(t *testing.T) {
	source := `
: User {
  id: int!
}

: User {
  name: string!
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}

	err := result.Errors[0]
	if err.Type == "" {
		t.Error("error type should not be empty")
	}
	if err.Message == "" {
		t.Error("error message should not be empty")
	}
	if err.Severity == "" {
		t.Error("error severity should not be empty")
	}
	if err.FixHint == "" {
		t.Error("error fix hint should not be empty")
	}
}

func TestValidateWithDatabaseType(t *testing.T) {
	source := `
@ GET /data {
  % db: Database
  $ result = db.query("SELECT * FROM data")
  > result
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result, Database should be a builtin type: %v", result.Errors)
	}
}

func TestCreateLexerError(t *testing.T) {
	source := "line 1\nline 2\nline 3"
	v := NewValidator(source, "test.glyph")

	testErr := fmt.Errorf("unterminated string at line 2")
	err := v.createLexerError(testErr)
	if err == nil {
		t.Fatal("createLexerError returned nil")
	}
	if err.Type != ErrTypeLexer {
		t.Errorf("expected type %s, got %s", ErrTypeLexer, err.Type)
	}
	if err.Severity != "error" {
		t.Errorf("expected severity 'error', got %s", err.Severity)
	}
	if err.Location == nil {
		t.Error("expected location to be set")
	}
	if err.Location != nil && err.Location.Line != 2 {
		t.Errorf("expected line 2, got %d", err.Location.Line)
	}
}

func TestCreateParseError(t *testing.T) {
	source := "line 1\nline 2\nline 3"
	v := NewValidator(source, "test.glyph")

	testErr := fmt.Errorf("unexpected token at line 3 column 5")
	err := v.createParseError(testErr)
	if err == nil {
		t.Fatal("createParseError returned nil")
	}
	if err.Type != ErrTypeSyntax {
		t.Errorf("expected type %s, got %s", ErrTypeSyntax, err.Type)
	}
	if err.Location == nil {
		t.Error("expected location to be set")
	}
}

func TestValidateGenericTypeArgs(t *testing.T) {
	source := `
: User {
  id: int!
}

: Response {
  data: Result<User, string>!
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if !result.Valid {
		t.Errorf("expected valid result for generic type with args: %v", result.Errors)
	}
}

func TestValidateUndefinedGenericArg(t *testing.T) {
	source := `
: Response {
  data: Result<UnknownType, string>!
}
`
	v := NewValidator(source, "test.glyph")
	result := v.Validate()

	if result.Valid {
		t.Error("expected invalid result for undefined generic type argument")
	}
}

func TestValidationResultJSONSerialization(t *testing.T) {
	result := &ValidationResult{
		Valid:    false,
		FilePath: "test.glyph",
		Errors: []*ValidationError{
			{
				Type:     ErrTypeSyntax,
				Message:  "test error",
				Severity: "error",
				Location: &Location{
					File:   "test.glyph",
					Line:   10,
					Column: 5,
				},
				FixHint:   "fix it",
				Context:   "some code",
				RelatedTo: "User",
			},
		},
		Warnings: []*ValidationError{},
		Stats: &ValidationStats{
			Types:     1,
			Routes:    2,
			Functions: 3,
			Commands:  4,
			Lines:     100,
		},
	}

	data, err := result.ToJSON(true)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	// Verify all fields are serialized
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed["valid"] != false {
		t.Error("valid field not correct")
	}
	if parsed["file_path"] != "test.glyph" {
		t.Error("file_path field not correct")
	}

	errors := parsed["errors"].([]interface{})
	if len(errors) != 1 {
		t.Error("errors not serialized correctly")
	}

	errObj := errors[0].(map[string]interface{})
	if errObj["type"] != ErrTypeSyntax {
		t.Error("error type not correct")
	}

	location := errObj["location"].(map[string]interface{})
	if location["line"].(float64) != 10 {
		t.Error("location line not correct")
	}
}
