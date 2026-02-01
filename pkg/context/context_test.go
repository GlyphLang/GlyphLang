package context

import (
	"encoding/json"
	"github.com/glyphlang/glyph/pkg/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("/test/path")
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.rootDir != "/test/path" {
		t.Errorf("expected rootDir /test/path, got %s", g.rootDir)
	}
}

func TestComputeHash(t *testing.T) {
	tests := []struct {
		input    string
		wantLen  int
		wantSame bool
	}{
		{"hello", 64, true},
		{"world", 64, true},
		{"", 64, true},
	}

	for _, tt := range tests {
		hash := computeHash(tt.input)
		if len(hash) != tt.wantLen {
			t.Errorf("computeHash(%q) returned hash of length %d, want %d", tt.input, len(hash), tt.wantLen)
		}
		// Same input should produce same hash
		hash2 := computeHash(tt.input)
		if hash != hash2 {
			t.Errorf("computeHash(%q) not deterministic", tt.input)
		}
	}

	// Different inputs should produce different hashes
	h1 := computeHash("hello")
	h2 := computeHash("world")
	if h1 == h2 {
		t.Error("different inputs produced same hash")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice []string
		val   string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{nil, "a", false},
		{[]string{"GET", "POST", "PUT"}, "GET", true},
		{[]string{"GET", "POST", "PUT"}, "DELETE", false},
	}

	for _, tt := range tests {
		got := contains(tt.slice, tt.val)
		if got != tt.want {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.val, got, tt.want)
		}
	}
}

func TestTypeToString(t *testing.T) {
	tests := []struct {
		name string
		typ  ast.Type
		want string
	}{
		{"nil type", nil, "any"},
		{"int type", ast.IntType{}, "int"},
		{"string type", ast.StringType{}, "string"},
		{"bool type", ast.BoolType{}, "bool"},
		{"float type", ast.FloatType{}, "float"},
		{"named type", ast.NamedType{Name: "User"}, "User"},
		{"array type", ast.ArrayType{ElementType: ast.StringType{}}, "[string]"},
		{"optional type", ast.OptionalType{InnerType: ast.IntType{}}, "int?"},
		{"database type", ast.DatabaseType{}, "Database"},
		{
			"union type",
			ast.UnionType{Types: []ast.Type{ast.StringType{}, ast.IntType{}}},
			"string | int",
		},
		{
			"function type",
			ast.FunctionType{
				ParamTypes: []ast.Type{ast.StringType{}, ast.IntType{}},
				ReturnType: ast.BoolType{},
			},
			"(string, int) -> bool",
		},
		{
			"nested array",
			ast.ArrayType{ElementType: ast.ArrayType{ElementType: ast.IntType{}}},
			"[[int]]",
		},
		{
			"optional array",
			ast.OptionalType{InnerType: ast.ArrayType{ElementType: ast.StringType{}}},
			"[string]?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeToString(tt.typ)
			if got != tt.want {
				t.Errorf("typeToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeToStringGeneric(t *testing.T) {
	genericType := ast.GenericType{
		BaseType: ast.NamedType{Name: "List"},
		TypeArgs: []ast.Type{ast.StringType{}},
	}
	got := typeToString(genericType)
	want := "List<string>"
	if got != want {
		t.Errorf("typeToString(GenericType) = %q, want %q", got, want)
	}
}

func TestExtractTypeInfo(t *testing.T) {
	typeDef := &ast.TypeDef{
		Name: "User",
		Fields: []ast.Field{
			{Name: "id", TypeAnnotation: ast.IntType{}, Required: true},
			{Name: "name", TypeAnnotation: ast.StringType{}, Required: true},
			{Name: "email", TypeAnnotation: ast.StringType{}, Required: false},
		},
		TypeParams: []ast.TypeParameter{},
	}

	info := extractTypeInfo(typeDef)

	if info.Name != "User" {
		t.Errorf("expected name User, got %s", info.Name)
	}
	if len(info.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(info.Fields))
	}
	if info.Fields[0] != "id: int!" {
		t.Errorf("expected 'id: int!', got %q", info.Fields[0])
	}
	if info.Fields[1] != "name: string!" {
		t.Errorf("expected 'name: string!', got %q", info.Fields[1])
	}
	if info.Fields[2] != "email: string" {
		t.Errorf("expected 'email: string', got %q", info.Fields[2])
	}
	if info.Hash == "" {
		t.Error("hash should not be empty")
	}
	if len(info.Hash) != 8 {
		t.Errorf("hash should be 8 chars, got %d", len(info.Hash))
	}
}

func TestExtractTypeInfoWithTypeParams(t *testing.T) {
	typeDef := &ast.TypeDef{
		Name: "Result",
		Fields: []ast.Field{
			{Name: "value", TypeAnnotation: ast.NamedType{Name: "T"}, Required: false},
			{Name: "error", TypeAnnotation: ast.StringType{}, Required: false},
		},
		TypeParams: []ast.TypeParameter{
			{Name: "T"},
			{Name: "E"},
		},
	}

	info := extractTypeInfo(typeDef)

	if len(info.TypeParams) != 2 {
		t.Fatalf("expected 2 type params, got %d", len(info.TypeParams))
	}
	if info.TypeParams[0] != "T" || info.TypeParams[1] != "E" {
		t.Errorf("unexpected type params: %v", info.TypeParams)
	}
}

func TestExtractRouteInfo(t *testing.T) {
	route := &ast.Route{
		Method:     ast.Get,
		Path:       "/users/:id",
		ReturnType: ast.NamedType{Name: "User"},
		Auth: &ast.AuthConfig{
			AuthType: "jwt",
			Required: true,
		},
		Injections: []ast.Injection{
			{Name: "db"},
		},
		QueryParams: []ast.QueryParamDecl{
			{Name: "include", Type: ast.StringType{}, Required: false},
		},
	}

	info := extractRouteInfo(route)

	if info.Method != "GET" {
		t.Errorf("expected method GET, got %s", info.Method)
	}
	if info.Path != "/users/:id" {
		t.Errorf("expected path /users/:id, got %s", info.Path)
	}
	if info.Returns != "User" {
		t.Errorf("expected returns User, got %s", info.Returns)
	}
	if info.Auth != "jwt!" {
		t.Errorf("expected auth jwt!, got %s", info.Auth)
	}
	if len(info.Params) != 1 || info.Params[0] != "id: string" {
		t.Errorf("expected params [id: string], got %v", info.Params)
	}
	if len(info.Injects) != 1 || info.Injects[0] != "db" {
		t.Errorf("expected injects [db], got %v", info.Injects)
	}
	if len(info.QueryParams) != 1 || info.QueryParams[0] != "include: string" {
		t.Errorf("expected query params [include: string], got %v", info.QueryParams)
	}
	if info.Hash == "" || len(info.Hash) != 8 {
		t.Errorf("invalid hash: %s", info.Hash)
	}
}

func TestExtractFunctionInfo(t *testing.T) {
	fn := &ast.Function{
		Name: "formatUser",
		Params: []ast.Field{
			{Name: "user", TypeAnnotation: ast.NamedType{Name: "User"}, Required: true},
			{Name: "format", TypeAnnotation: ast.StringType{}, Required: false},
		},
		ReturnType: ast.StringType{},
		TypeParams: []ast.TypeParameter{},
	}

	info := extractFunctionInfo(fn)

	if info.Name != "formatUser" {
		t.Errorf("expected name formatUser, got %s", info.Name)
	}
	if info.Returns != "string" {
		t.Errorf("expected returns string, got %s", info.Returns)
	}
	if len(info.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(info.Params))
	}
	if info.Params[0] != "user: User!" {
		t.Errorf("expected 'user: User!', got %q", info.Params[0])
	}
	if info.Params[1] != "format: string" {
		t.Errorf("expected 'format: string', got %q", info.Params[1])
	}
}

func TestExtractCommandInfo(t *testing.T) {
	cmd := &ast.Command{
		Name:        "deploy",
		Description: "Deploy the application",
		Params: []ast.CommandParam{
			{Name: "env", Type: ast.StringType{}, Required: true, IsFlag: false},
			{Name: "force", Type: ast.BoolType{}, Required: false, IsFlag: true},
		},
	}

	info := extractCommandInfo(cmd)

	if info.Name != "deploy" {
		t.Errorf("expected name deploy, got %s", info.Name)
	}
	if info.Description != "Deploy the application" {
		t.Errorf("expected description, got %s", info.Description)
	}
	if len(info.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(info.Params))
	}
	if info.Params[0] != "env: string!" {
		t.Errorf("expected 'env: string!', got %q", info.Params[0])
	}
	if info.Params[1] != "--force: bool" {
		t.Errorf("expected '--force: bool', got %q", info.Params[1])
	}
}

func TestProjectContextToJSON(t *testing.T) {
	ctx := &ProjectContext{
		Version:     "1.0",
		Generated:   time.Now(),
		ProjectHash: "abc123",
		Files:       make(map[string]*FileContext),
		Types:       make(map[string]*TypeInfo),
		Routes:      make([]*RouteInfo, 0),
		Functions:   make(map[string]*FunctionInfo),
		Commands:    make(map[string]*CommandInfo),
		Patterns:    []string{"crud(users)"},
	}

	// Test compact JSON
	compactJSON, err := ctx.ToJSON(false)
	if err != nil {
		t.Fatalf("ToJSON(false) error: %v", err)
	}
	if strings.Contains(string(compactJSON), "\n") {
		t.Error("compact JSON should not contain newlines")
	}

	// Test pretty JSON
	prettyJSON, err := ctx.ToJSON(true)
	if err != nil {
		t.Fatalf("ToJSON(true) error: %v", err)
	}
	if !strings.Contains(string(prettyJSON), "\n") {
		t.Error("pretty JSON should contain newlines")
	}

	// Verify JSON is valid
	var parsed ProjectContext
	if err := json.Unmarshal(compactJSON, &parsed); err != nil {
		t.Errorf("invalid JSON produced: %v", err)
	}
}

func TestProjectContextToCompact(t *testing.T) {
	ctx := &ProjectContext{
		Version:     "1.0",
		ProjectHash: "abcdef123456789",
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Fields: []string{"id: int!", "name: string!"}},
		},
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users/:id", Returns: "User", Auth: "jwt"},
		},
		Functions: map[string]*FunctionInfo{
			"formatName": {Name: "formatName", Params: []string{"first: string!", "last: string!"}, Returns: "string"},
		},
		Commands: map[string]*CommandInfo{
			"build": {Name: "build", Params: []string{"--prod: bool"}},
		},
		Patterns: []string{"crud(users)"},
	}

	compact := ctx.ToCompact()

	// Verify sections are present
	if !strings.Contains(compact, "# Glyph Project Context") {
		t.Error("missing header")
	}
	if !strings.Contains(compact, "## Types") {
		t.Error("missing Types section")
	}
	if !strings.Contains(compact, "## Routes") {
		t.Error("missing Routes section")
	}
	if !strings.Contains(compact, "## Functions") {
		t.Error("missing Functions section")
	}
	if !strings.Contains(compact, "## Commands") {
		t.Error("missing Commands section")
	}
	if !strings.Contains(compact, "## Patterns") {
		t.Error("missing Patterns section")
	}
	if !strings.Contains(compact, ": User") {
		t.Error("missing User type")
	}
	if !strings.Contains(compact, "@ GET /users/:id") {
		t.Error("missing route")
	}
}

func TestTruncateHash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "(none)"},
		{"abc", "abc"},
		{"abcdefghijkl", "abcdefghijkl"},
		{"abcdefghijklmno", "abcdefghijkl"},
	}

	for _, tt := range tests {
		got := truncateHash(tt.input)
		if got != tt.want {
			t.Errorf("truncateHash(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDefaultContextPath(t *testing.T) {
	path := DefaultContextPath("/project")
	expected := filepath.Join("/project", ".glyph", "context.json")
	if path != expected {
		t.Errorf("DefaultContextPath = %q, want %q", path, expected)
	}
}

func TestContextDiff(t *testing.T) {
	prev := &ProjectContext{
		ProjectHash: "hash1",
		Files: map[string]*FileContext{
			"a.glyph": {Path: "a.glyph", Hash: "h1"},
			"b.glyph": {Path: "b.glyph", Hash: "h2"},
		},
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Hash: "t1"},
		},
	}

	curr := &ProjectContext{
		ProjectHash: "hash2",
		Files: map[string]*FileContext{
			"a.glyph": {Path: "a.glyph", Hash: "h1_changed"},
			"c.glyph": {Path: "c.glyph", Hash: "h3"},
		},
		Types: map[string]*TypeInfo{
			"User":    {Name: "User", Hash: "t1_changed"},
			"Product": {Name: "Product", Hash: "t2"},
		},
	}

	diff := curr.Diff(prev)

	if !diff.HasChanges {
		t.Error("expected HasChanges to be true")
	}
	if diff.PreviousHash != "hash1" {
		t.Errorf("expected previous hash hash1, got %s", diff.PreviousHash)
	}
	if diff.CurrentHash != "hash2" {
		t.Errorf("expected current hash hash2, got %s", diff.CurrentHash)
	}

	// Check file diff
	if diff.Files == nil {
		t.Fatal("expected Files diff")
	}
	if len(diff.Files.Added) != 1 || diff.Files.Added[0] != "c.glyph" {
		t.Errorf("expected c.glyph added, got %v", diff.Files.Added)
	}
	if len(diff.Files.Removed) != 1 || diff.Files.Removed[0] != "b.glyph" {
		t.Errorf("expected b.glyph removed, got %v", diff.Files.Removed)
	}
	if len(diff.Files.Modified) != 1 || diff.Files.Modified[0] != "a.glyph" {
		t.Errorf("expected a.glyph modified, got %v", diff.Files.Modified)
	}

	// Check types diff
	if diff.Types == nil {
		t.Fatal("expected Types diff")
	}
	if !contains(diff.Types.Added, "Product") {
		t.Error("expected Product added")
	}
	if !contains(diff.Types.Modified, "User") {
		t.Error("expected User modified")
	}
}

func TestContextDiffNoChanges(t *testing.T) {
	ctx := &ProjectContext{
		ProjectHash: "same_hash",
	}

	diff := ctx.Diff(ctx)

	if diff.HasChanges {
		t.Error("expected no changes when comparing same context")
	}
}

func TestContextDiffNilPrevious(t *testing.T) {
	ctx := &ProjectContext{
		ProjectHash: "hash1",
		Files: map[string]*FileContext{
			"a.glyph": {Path: "a.glyph", Hash: "h1"},
		},
	}

	diff := ctx.Diff(nil)

	if !diff.HasChanges {
		t.Error("expected changes when previous is nil")
	}
	if diff.PreviousHash != "" {
		t.Error("expected empty previous hash")
	}
}

func TestContextDiffToCompact(t *testing.T) {
	ctx := &ProjectContext{
		ProjectHash: "current123456",
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Fields: []string{"id: int!"}},
		},
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Returns: "User"},
		},
		Functions: map[string]*FunctionInfo{
			"test": {Name: "test", Params: []string{}, Returns: "void"},
		},
	}

	diff := &ContextDiff{
		PreviousHash: "prev123456",
		CurrentHash:  "current123456",
		HasChanges:   true,
		Files: &FilesDiff{
			Added:    []string{"new.glyph"},
			Removed:  []string{"old.glyph"},
			Modified: []string{"changed.glyph"},
		},
		Types: &ItemsDiff{
			Added: []string{"User"},
		},
		Routes: &RoutesDiff{
			Added: []*RouteInfo{{Method: "GET", Path: "/users", Returns: "User"}},
		},
		Functions: &ItemsDiff{
			Added: []string{"test"},
		},
	}

	compact := diff.ToCompact(ctx)

	if !strings.Contains(compact, "# Glyph Context Changes") {
		t.Error("missing header")
	}
	if !strings.Contains(compact, "+ new.glyph") {
		t.Error("missing added file")
	}
	if !strings.Contains(compact, "- old.glyph") {
		t.Error("missing removed file")
	}
	if !strings.Contains(compact, "~ changed.glyph") {
		t.Error("missing modified file")
	}
}

func TestContextDiffToCompactNoChanges(t *testing.T) {
	diff := &ContextDiff{
		HasChanges: false,
	}

	compact := diff.ToCompact(nil)

	if !strings.Contains(compact, "No changes detected") {
		t.Error("should indicate no changes")
	}
}

func TestTargetedContextForRoute(t *testing.T) {
	ctx := &ProjectContext{
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Fields: []string{"id: int!"}},
		},
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Injects: []string{"db", "cache"}},
		},
		Functions: map[string]*FunctionInfo{
			"validate": {Name: "validate", Returns: "bool"},
		},
		Patterns: []string{"crud(users)"},
	}

	tc := ctx.ForRoute()

	if tc.Task != "route" {
		t.Errorf("expected task 'route', got %s", tc.Task)
	}
	if tc.Types == nil || len(tc.Types) != 1 {
		t.Error("expected types to be included")
	}
	if tc.Syntax == nil || len(tc.Syntax.Examples) == 0 {
		t.Error("expected syntax guide with examples")
	}
	if len(tc.Injections) != 2 {
		t.Errorf("expected 2 injections, got %d", len(tc.Injections))
	}
}

func TestTargetedContextForType(t *testing.T) {
	ctx := &ProjectContext{
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Fields: []string{"id: int!"}},
		},
	}

	tc := ctx.ForType()

	if tc.Task != "type" {
		t.Errorf("expected task 'type', got %s", tc.Task)
	}
	if tc.Syntax == nil {
		t.Error("expected syntax guide")
	}
	if len(tc.Syntax.Notes) == 0 {
		t.Error("expected syntax notes")
	}
}

func TestTargetedContextForFunction(t *testing.T) {
	ctx := &ProjectContext{
		Types: map[string]*TypeInfo{
			"User": {Name: "User"},
		},
		Functions: map[string]*FunctionInfo{
			"test": {Name: "test"},
		},
	}

	tc := ctx.ForFunction()

	if tc.Task != "function" {
		t.Errorf("expected task 'function', got %s", tc.Task)
	}
	if tc.Types == nil {
		t.Error("expected types")
	}
	if tc.Functions == nil {
		t.Error("expected functions")
	}
}

func TestTargetedContextForCommand(t *testing.T) {
	ctx := &ProjectContext{
		Commands: map[string]*CommandInfo{
			"deploy": {Name: "deploy"},
		},
	}

	tc := ctx.ForCommand()

	if tc.Task != "command" {
		t.Errorf("expected task 'command', got %s", tc.Task)
	}
	if tc.Commands == nil {
		t.Error("expected commands")
	}
}

func TestTargetedContextToJSON(t *testing.T) {
	tc := &TargetedContext{
		Task:        "route",
		Description: "Test context",
	}

	data, err := tc.ToJSON(true)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}
}

func TestTargetedContextToCompact(t *testing.T) {
	tc := &TargetedContext{
		Task:        "route",
		Description: "Context for routes",
		Syntax: &SyntaxGuide{
			Examples: []string{"@ GET /path"},
			Notes:    []string{"Use @ for routes"},
		},
		Injections: []string{"db"},
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Fields: []string{"id: int!"}},
		},
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Returns: "User"},
		},
		Functions: map[string]*FunctionInfo{
			"test": {Name: "test", Params: []string{}, Returns: "void"},
		},
		Commands: map[string]*CommandInfo{
			"build": {Name: "build", Params: []string{}},
		},
		Patterns: []string{"crud(users)"},
	}

	compact := tc.ToCompact()

	if !strings.Contains(compact, "# Glyph Context: route") {
		t.Error("missing header")
	}
	if !strings.Contains(compact, "## Syntax") {
		t.Error("missing syntax section")
	}
	if !strings.Contains(compact, "## Available Injections") {
		t.Error("missing injections section")
	}
	if !strings.Contains(compact, "## Types") {
		t.Error("missing types section")
	}
}

func TestGenerateStubs(t *testing.T) {
	ctx := &ProjectContext{
		Generated:   time.Now(),
		ProjectHash: "abc123456789def",
		Types: map[string]*TypeInfo{
			"User": {Name: "User", Fields: []string{"id: int!", "name: string!"}},
		},
		Functions: map[string]*FunctionInfo{
			"validate": {Name: "validate", Params: []string{"data: any!"}, Returns: "bool"},
		},
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Returns: "User", Auth: "jwt", Injects: []string{"db"}},
		},
	}

	stubs := ctx.GenerateStubs()

	if !strings.Contains(stubs, "# Auto-generated Glyph type stubs") {
		t.Error("missing header")
	}
	if !strings.Contains(stubs, ": User") {
		t.Error("missing User type stub")
	}
	if !strings.Contains(stubs, "! validate") {
		t.Error("missing validate function stub")
	}
	if !strings.Contains(stubs, "@ GET /users") {
		t.Error("missing route stub")
	}
}

func TestDetectPatterns(t *testing.T) {
	ctx := &ProjectContext{
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users"},
			{Method: "POST", Path: "/users"},
			{Method: "PUT", Path: "/users/:id"},
			{Method: "DELETE", Path: "/users/:id"},
			{Method: "GET", Path: "/posts", Auth: "jwt"},
			{Method: "POST", Path: "/posts", Auth: "jwt"},
			{Method: "GET", Path: "/comments", Injects: []string{"db"}},
		},
	}

	g := NewGenerator(".")
	patterns := g.detectPatterns(ctx)

	// Should detect CRUD pattern for users
	hasCrud := false
	hasAuth := false
	hasDb := false
	for _, p := range patterns {
		if strings.Contains(p, "crud(users)") {
			hasCrud = true
		}
		if strings.Contains(p, "auth_routes") {
			hasAuth = true
		}
		if strings.Contains(p, "database_routes") {
			hasDb = true
		}
	}

	if !hasCrud {
		t.Error("expected crud(users) pattern")
	}
	if !hasAuth {
		t.Error("expected auth_routes pattern")
	}
	if !hasDb {
		t.Error("expected database_routes pattern")
	}
}

func TestContextDiffToJSON(t *testing.T) {
	diff := &ContextDiff{
		PreviousHash: "prev",
		CurrentHash:  "curr",
		HasChanges:   true,
	}

	data, err := diff.ToJSON(true)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}

	if parsed["has_changes"] != true {
		t.Error("expected has_changes to be true")
	}
}

func TestSaveAndLoadContext(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "glyph-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := &ProjectContext{
		Version:     "1.0",
		Generated:   time.Now().Truncate(time.Second), // Truncate for comparison
		ProjectHash: "testhash123",
		Files: map[string]*FileContext{
			"main.glyph": {Path: "main.glyph", Hash: "filehash"},
		},
		Types:     make(map[string]*TypeInfo),
		Routes:    make([]*RouteInfo, 0),
		Functions: make(map[string]*FunctionInfo),
		Commands:  make(map[string]*CommandInfo),
		Patterns:  []string{"test_pattern"},
	}

	// Save context
	savePath := filepath.Join(tmpDir, ".glyph", "context.json")
	if err := ctx.SaveContext(savePath); err != nil {
		t.Fatalf("SaveContext error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Fatal("context file was not created")
	}

	// Load context
	loaded, err := LoadContext(savePath)
	if err != nil {
		t.Fatalf("LoadContext error: %v", err)
	}

	// Verify loaded data
	if loaded.Version != ctx.Version {
		t.Errorf("version mismatch: got %s, want %s", loaded.Version, ctx.Version)
	}
	if loaded.ProjectHash != ctx.ProjectHash {
		t.Errorf("hash mismatch: got %s, want %s", loaded.ProjectHash, ctx.ProjectHash)
	}
	if len(loaded.Files) != len(ctx.Files) {
		t.Errorf("files count mismatch: got %d, want %d", len(loaded.Files), len(ctx.Files))
	}
}

func TestLoadContextNotExists(t *testing.T) {
	_, err := LoadContext("/nonexistent/path/context.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadContextInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	invalidPath := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidPath, []byte("not valid json"), 0644)

	_, err = LoadContext(invalidPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHelperFunctionsWithNilContext(t *testing.T) {
	// Test all helper functions with nil context
	if names := getTypeNames(nil); names != nil {
		t.Error("getTypeNames(nil) should return nil")
	}
	if hashes := getTypeHashes(nil); hashes != nil {
		t.Error("getTypeHashes(nil) should return nil")
	}
	if names := getFunctionNames(nil); names != nil {
		t.Error("getFunctionNames(nil) should return nil")
	}
	if hashes := getFunctionHashes(nil); hashes != nil {
		t.Error("getFunctionHashes(nil) should return nil")
	}
	if names := getCommandNames(nil); names != nil {
		t.Error("getCommandNames(nil) should return nil")
	}
	if hashes := getCommandHashes(nil); hashes != nil {
		t.Error("getCommandHashes(nil) should return nil")
	}
}

func TestDiffFilesNilContext(t *testing.T) {
	curr := &ProjectContext{
		Files: map[string]*FileContext{
			"a.glyph": {Path: "a.glyph", Hash: "h1"},
		},
	}

	diff := diffFiles(nil, curr)

	if diff == nil {
		t.Fatal("expected diff result")
	}
	if len(diff.Added) != 1 {
		t.Errorf("expected 1 added file, got %d", len(diff.Added))
	}
}

func TestDiffRoutesNilContext(t *testing.T) {
	curr := &ProjectContext{
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/test", Hash: "h1"},
		},
	}

	diff := diffRoutes(nil, curr)

	if diff == nil {
		t.Fatal("expected diff result")
	}
	if len(diff.Added) != 1 {
		t.Errorf("expected 1 added route, got %d", len(diff.Added))
	}
}

func TestExtractRouteInfoNoAuth(t *testing.T) {
	route := &ast.Route{
		Method: ast.Post,
		Path:   "/api/data",
		Auth:   nil,
	}

	info := extractRouteInfo(route)

	if info.Auth != "" {
		t.Errorf("expected empty auth, got %s", info.Auth)
	}
}

func TestExtractFunctionInfoNoReturn(t *testing.T) {
	fn := &ast.Function{
		Name:       "doSomething",
		Params:     []ast.Field{},
		ReturnType: nil,
	}

	info := extractFunctionInfo(fn)

	if info.Returns != "" {
		t.Errorf("expected empty returns, got %s", info.Returns)
	}
}

// Integration tests for file-based operations

func TestGenerateForFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test .glyph file
	glyphContent := `
: User {
  id: int!
  name: string!
}

@ GET /users/:id -> User {
  > {}
}
`
	testFile := filepath.Join(tmpDir, "test.glyph")
	if err := os.WriteFile(testFile, []byte(glyphContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Generate context for the file
	g := NewGenerator(tmpDir)
	ctx, err := g.GenerateForFile("test.glyph")
	if err != nil {
		t.Fatalf("GenerateForFile error: %v", err)
	}

	// Verify context
	if ctx.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", ctx.Version)
	}
	if len(ctx.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(ctx.Files))
	}
	if len(ctx.Types) != 1 {
		t.Errorf("expected 1 type, got %d", len(ctx.Types))
	}
	if _, ok := ctx.Types["User"]; !ok {
		t.Error("expected User type")
	}
	if len(ctx.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(ctx.Routes))
	}
	if ctx.ProjectHash == "" {
		t.Error("expected project hash to be set")
	}
}

func TestGenerateForFileNotExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	_, err = g.GenerateForFile("nonexistent.glyph")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestGenerateForFileParseError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid glyph file
	invalidContent := `: User {
  id: int!
  missing closing brace
`
	testFile := filepath.Join(tmpDir, "invalid.glyph")
	if err := os.WriteFile(testFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	g := NewGenerator(tmpDir)
	_, err = g.GenerateForFile("invalid.glyph")
	if err == nil {
		t.Error("expected error for invalid glyph file")
	}
}

func TestGenerate(t *testing.T) {
	// Create temp directory with multiple .glyph files
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create types.glyph
	typesContent := `
: User {
  id: int!
  name: string!
}

: Post {
  id: int!
  title: string!
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "types.glyph"), []byte(typesContent), 0644); err != nil {
		t.Fatalf("failed to write types.glyph: %v", err)
	}

	// Create routes.glyph
	routesContent := `
@ GET /users {
  > []
}

@ POST /users {
  > {}
}

@ GET /posts {
  > []
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "routes.glyph"), []byte(routesContent), 0644); err != nil {
		t.Fatalf("failed to write routes.glyph: %v", err)
	}

	// Generate context
	g := NewGenerator(tmpDir)
	ctx, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Verify context
	if len(ctx.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(ctx.Files))
	}
	if len(ctx.Types) != 2 {
		t.Errorf("expected 2 types, got %d", len(ctx.Types))
	}
	if len(ctx.Routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(ctx.Routes))
	}
	if ctx.ProjectHash == "" {
		t.Error("expected project hash")
	}
}

func TestGenerateSkipsHiddenDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a regular file
	if err := os.WriteFile(filepath.Join(tmpDir, "main.glyph"), []byte("@ GET /test {\n  > {}\n}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create hidden directory with a glyph file
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "hidden.glyph"), []byte("@ GET /hidden {\n  > {}\n}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create node_modules directory
	nodeModules := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(nodeModules, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nodeModules, "module.glyph"), []byte("@ GET /module {\n  > {}\n}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create vendor directory
	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.Mkdir(vendorDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vendorDir, "vendor.glyph"), []byte("@ GET /vendor {\n  > {}\n}"), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator(tmpDir)
	ctx, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Should only have 1 file (main.glyph), hidden/node_modules/vendor should be skipped
	if len(ctx.Files) != 1 {
		t.Errorf("expected 1 file (others should be skipped), got %d", len(ctx.Files))
	}
	if len(ctx.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(ctx.Routes))
	}
}

func TestGenerateWithSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "api")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files in root and subdirectory
	if err := os.WriteFile(filepath.Join(tmpDir, "main.glyph"), []byte("@ GET / {\n  > {}\n}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "users.glyph"), []byte("@ GET /users {\n  > []\n}"), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator(tmpDir)
	ctx, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if len(ctx.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(ctx.Files))
	}
}

func TestGenerateWithInvalidFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid file
	if err := os.WriteFile(filepath.Join(tmpDir, "valid.glyph"), []byte("@ GET /test {\n  > {}\n}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create invalid file (should be skipped with warning)
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.glyph"), []byte("invalid { syntax"), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator(tmpDir)
	ctx, err := g.Generate()
	// Should not error, but skip the invalid file
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Should only have the valid file
	if len(ctx.Files) != 1 {
		t.Errorf("expected 1 file (invalid should be skipped), got %d", len(ctx.Files))
	}
}

func TestGenerateEmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glyph-context-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	g := NewGenerator(tmpDir)
	ctx, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if len(ctx.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(ctx.Files))
	}
	if len(ctx.Patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(ctx.Patterns))
	}
}

func TestDiffRoutesModified(t *testing.T) {
	prev := &ProjectContext{
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Hash: "hash1", Returns: "User"},
		},
	}

	curr := &ProjectContext{
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Hash: "hash2", Returns: "[User]"}, // Modified
		},
	}

	diff := diffRoutes(prev, curr)

	if diff == nil {
		t.Fatal("expected diff result")
	}
	if len(diff.Modified) != 1 {
		t.Errorf("expected 1 modified route, got %d", len(diff.Modified))
	}
	if len(diff.Added) != 0 {
		t.Errorf("expected 0 added routes, got %d", len(diff.Added))
	}
}

func TestDiffRoutesRemoved(t *testing.T) {
	prev := &ProjectContext{
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Hash: "hash1"},
			{Method: "DELETE", Path: "/users/:id", Hash: "hash2"},
		},
	}

	curr := &ProjectContext{
		Routes: []*RouteInfo{
			{Method: "GET", Path: "/users", Hash: "hash1"},
		},
	}

	diff := diffRoutes(prev, curr)

	if diff == nil {
		t.Fatal("expected diff result")
	}
	if len(diff.Removed) != 1 {
		t.Errorf("expected 1 removed route, got %d", len(diff.Removed))
	}
}

func TestExtractRouteInfoRequiredQueryParam(t *testing.T) {
	route := &ast.Route{
		Method: ast.Get,
		Path:   "/search",
		QueryParams: []ast.QueryParamDecl{
			{Name: "q", Type: ast.StringType{}, Required: true},
			{Name: "limit", Type: ast.IntType{}, Required: false},
		},
	}

	info := extractRouteInfo(route)

	if len(info.QueryParams) != 2 {
		t.Fatalf("expected 2 query params, got %d", len(info.QueryParams))
	}
	if info.QueryParams[0] != "q: string!" {
		t.Errorf("expected 'q: string!', got %q", info.QueryParams[0])
	}
	if info.QueryParams[1] != "limit: int" {
		t.Errorf("expected 'limit: int', got %q", info.QueryParams[1])
	}
}

func TestExtractFunctionInfoWithTypeParams(t *testing.T) {
	fn := &ast.Function{
		Name: "map",
		Params: []ast.Field{
			{Name: "arr", TypeAnnotation: ast.ArrayType{ElementType: ast.NamedType{Name: "T"}}, Required: true},
		},
		ReturnType: ast.ArrayType{ElementType: ast.NamedType{Name: "U"}},
		TypeParams: []ast.TypeParameter{
			{Name: "T"},
			{Name: "U"},
		},
	}

	info := extractFunctionInfo(fn)

	if len(info.TypeParams) != 2 {
		t.Errorf("expected 2 type params, got %d", len(info.TypeParams))
	}
	if info.TypeParams[0] != "T" || info.TypeParams[1] != "U" {
		t.Errorf("unexpected type params: %v", info.TypeParams)
	}
}

func TestTypeToStringTypeParameterType(t *testing.T) {
	// Test TypeParameterType (used inside generic definitions)
	typ := ast.TypeParameterType{Name: "T"}
	// TypeParameterType should fall through to default case
	got := typeToString(typ)
	if got != "any" {
		t.Errorf("typeToString(TypeParameterType) = %q, want 'any'", got)
	}
}

func TestContextDiffToJSONCompact(t *testing.T) {
	diff := &ContextDiff{
		PreviousHash: "prev",
		CurrentHash:  "curr",
		HasChanges:   true,
	}

	data, err := diff.ToJSON(false)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	if strings.Contains(string(data), "\n") {
		t.Error("compact JSON should not contain newlines")
	}
}

func TestProjectContextToCompactWithTypeParams(t *testing.T) {
	ctx := &ProjectContext{
		ProjectHash: "abcdef123456789",
		Types: map[string]*TypeInfo{
			"Result": {Name: "Result", Fields: []string{"value: T?"}, TypeParams: []string{"T", "E"}},
		},
		Functions: map[string]*FunctionInfo{
			"map": {Name: "map", Params: []string{"arr: [T]!"}, Returns: "[U]", TypeParams: []string{"T", "U"}},
		},
	}

	compact := ctx.ToCompact()

	if !strings.Contains(compact, "Result<T, E>") {
		t.Error("expected generic type with type params in compact output")
	}
	if !strings.Contains(compact, "map<T, U>") {
		t.Error("expected generic function with type params in compact output")
	}
}

func TestTargetedContextToCompactWithTypeParams(t *testing.T) {
	tc := &TargetedContext{
		Task: "type",
		Types: map[string]*TypeInfo{
			"Result": {Name: "Result", Fields: []string{"value: T?"}, TypeParams: []string{"T"}},
		},
	}

	compact := tc.ToCompact()

	if !strings.Contains(compact, "Result<T>") {
		t.Error("expected generic type in targeted context")
	}
}

func TestGenerateStubsWithTypeParams(t *testing.T) {
	ctx := &ProjectContext{
		Generated:   time.Now(),
		ProjectHash: "abc123456789def",
		Types: map[string]*TypeInfo{
			"Result": {Name: "Result", Fields: []string{"value: T?", "error: E?"}, TypeParams: []string{"T", "E"}},
		},
		Functions: map[string]*FunctionInfo{
			"identity": {Name: "identity", Params: []string{"x: T!"}, Returns: "T", TypeParams: []string{"T"}},
		},
	}

	stubs := ctx.GenerateStubs()

	if !strings.Contains(stubs, ": Result<T, E>") {
		t.Error("expected generic type stub")
	}
	if !strings.Contains(stubs, "! identity<T>") {
		t.Error("expected generic function stub")
	}
}
