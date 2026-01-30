// Package context provides AI-optimized context generation for Glyph projects.
// This enables AI agents to efficiently understand and work with Glyph codebases
// by providing compact, cacheable representations of project structure.
package context

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/glyphlang/glyph/pkg/interpreter"
	"github.com/glyphlang/glyph/pkg/parser"
)

// ProjectContext represents the complete AI-optimized context for a Glyph project
type ProjectContext struct {
	Version     string                   `json:"version"`
	Generated   time.Time                `json:"generated"`
	ProjectHash string                   `json:"project_hash"`
	Files       map[string]*FileContext  `json:"files"`
	Types       map[string]*TypeInfo     `json:"types"`
	Routes      []*RouteInfo             `json:"routes"`
	Functions   map[string]*FunctionInfo `json:"functions"`
	Commands    map[string]*CommandInfo  `json:"commands"`
	Patterns    []string                 `json:"patterns"`
}

// FileContext represents context for a single file
type FileContext struct {
	Path     string    `json:"path"`
	Hash     string    `json:"hash"`
	Modified time.Time `json:"modified"`
	Summary  string    `json:"summary"`
}

// TypeInfo represents a compact type definition
type TypeInfo struct {
	Name       string   `json:"name"`
	Fields     []string `json:"fields"`                // Compact field representations: "name: type" or "name: type!"
	TypeParams []string `json:"type_params,omitempty"` // Generic type parameters
	Hash       string   `json:"hash"`
}

// RouteInfo represents a compact route definition
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Params      []string `json:"params,omitempty"`       // Path parameters
	QueryParams []string `json:"query_params,omitempty"` // Query parameters
	Returns     string   `json:"returns,omitempty"`
	Auth        string   `json:"auth,omitempty"`
	Injects     []string `json:"injects,omitempty"`
	Hash        string   `json:"hash"`
}

// FunctionInfo represents a compact function definition
type FunctionInfo struct {
	Name       string   `json:"name"`
	Params     []string `json:"params"` // "name: type"
	Returns    string   `json:"returns"`
	TypeParams []string `json:"type_params,omitempty"`
	Hash       string   `json:"hash"`
}

// CommandInfo represents a compact CLI command definition
type CommandInfo struct {
	Name        string   `json:"name"`
	Params      []string `json:"params"` // "name: type" or "--flag: type"
	Description string   `json:"description,omitempty"`
	Hash        string   `json:"hash"`
}

// Generator generates AI context from Glyph source files
type Generator struct {
	rootDir string
}

// NewGenerator creates a new context generator
func NewGenerator(rootDir string) *Generator {
	return &Generator{rootDir: rootDir}
}

// Generate generates the complete project context
func (g *Generator) Generate() (*ProjectContext, error) {
	ctx := &ProjectContext{
		Version:   "1.0",
		Generated: time.Now(),
		Files:     make(map[string]*FileContext),
		Types:     make(map[string]*TypeInfo),
		Routes:    make([]*RouteInfo, 0),
		Functions: make(map[string]*FunctionInfo),
		Commands:  make(map[string]*CommandInfo),
		Patterns:  make([]string, 0),
	}

	// Find all .glyph files
	files, err := g.findGlyphFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find glyph files: %w", err)
	}

	// Process each file
	var allHashes []string
	for _, file := range files {
		fileCtx, err := g.processFile(file, ctx)
		if err != nil {
			// Log warning but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s: %v\n", file, err)
			continue
		}
		ctx.Files[file] = fileCtx
		allHashes = append(allHashes, fileCtx.Hash)
	}

	// Compute project hash from all file hashes
	sort.Strings(allHashes)
	ctx.ProjectHash = computeHash(strings.Join(allHashes, ""))

	// Detect common patterns
	ctx.Patterns = g.detectPatterns(ctx)

	return ctx, nil
}

// GenerateForFile generates context for a single file
func (g *Generator) GenerateForFile(filePath string) (*ProjectContext, error) {
	ctx := &ProjectContext{
		Version:   "1.0",
		Generated: time.Now(),
		Files:     make(map[string]*FileContext),
		Types:     make(map[string]*TypeInfo),
		Routes:    make([]*RouteInfo, 0),
		Functions: make(map[string]*FunctionInfo),
		Commands:  make(map[string]*CommandInfo),
		Patterns:  make([]string, 0),
	}

	fileCtx, err := g.processFile(filePath, ctx)
	if err != nil {
		return nil, err
	}

	ctx.Files[filePath] = fileCtx
	ctx.ProjectHash = fileCtx.Hash

	return ctx, nil
}

// findGlyphFiles finds all .glyph files in the project
func (g *Generator) findGlyphFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(g.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories and node_modules
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		// Include .glyph files
		if strings.HasSuffix(path, ".glyph") {
			relPath, _ := filepath.Rel(g.rootDir, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

// processFile processes a single Glyph file and extracts context
func (g *Generator) processFile(relPath string, ctx *ProjectContext) (*FileContext, error) {
	fullPath := filepath.Join(g.rootDir, relPath)

	// Read file
	source, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Compute file hash
	fileHash := computeHash(string(source))

	// Parse file
	lexer := parser.NewLexer(string(source))
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %w", err)
	}

	p := parser.NewParser(tokens)
	module, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser error: %w", err)
	}

	// Extract information from AST
	var typeCount, routeCount, funcCount, cmdCount int

	for _, item := range module.Items {
		switch v := item.(type) {
		case *interpreter.TypeDef:
			typeInfo := extractTypeInfo(v)
			ctx.Types[v.Name] = typeInfo
			typeCount++

		case *interpreter.Route:
			routeInfo := extractRouteInfo(v)
			ctx.Routes = append(ctx.Routes, routeInfo)
			routeCount++

		case *interpreter.Function:
			funcInfo := extractFunctionInfo(v)
			ctx.Functions[v.Name] = funcInfo
			funcCount++

		case *interpreter.Command:
			cmdInfo := extractCommandInfo(v)
			ctx.Commands[v.Name] = cmdInfo
			cmdCount++
		}
	}

	// Generate summary
	var parts []string
	if typeCount > 0 {
		parts = append(parts, fmt.Sprintf("%d types", typeCount))
	}
	if routeCount > 0 {
		parts = append(parts, fmt.Sprintf("%d routes", routeCount))
	}
	if funcCount > 0 {
		parts = append(parts, fmt.Sprintf("%d functions", funcCount))
	}
	if cmdCount > 0 {
		parts = append(parts, fmt.Sprintf("%d commands", cmdCount))
	}

	summary := strings.Join(parts, ", ")
	if summary == "" {
		summary = "empty"
	}

	return &FileContext{
		Path:     relPath,
		Hash:     fileHash,
		Modified: info.ModTime(),
		Summary:  summary,
	}, nil
}

// extractTypeInfo extracts compact type information
func extractTypeInfo(t *interpreter.TypeDef) *TypeInfo {
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fieldStr := f.Name + ": " + typeToString(f.TypeAnnotation)
		if f.Required {
			fieldStr += "!"
		}
		fields[i] = fieldStr
	}

	typeParams := make([]string, len(t.TypeParams))
	for i, tp := range t.TypeParams {
		typeParams[i] = tp.Name
	}

	// Compute hash of type definition
	hashInput := t.Name + ":" + strings.Join(fields, ",")

	return &TypeInfo{
		Name:       t.Name,
		Fields:     fields,
		TypeParams: typeParams,
		Hash:       computeHash(hashInput)[:8],
	}
}

// extractRouteInfo extracts compact route information
func extractRouteInfo(r *interpreter.Route) *RouteInfo {
	// Extract path parameters (e.g., :id from /users/:id)
	var params []string
	parts := strings.Split(r.Path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:]+": string")
		}
	}

	// Extract query parameters
	var queryParams []string
	for _, qp := range r.QueryParams {
		qpStr := qp.Name + ": " + typeToString(qp.Type)
		if qp.Required {
			qpStr += "!"
		}
		queryParams = append(queryParams, qpStr)
	}

	// Extract injections
	var injects []string
	for _, inj := range r.Injections {
		injects = append(injects, inj.Name)
	}

	// Auth info
	var auth string
	if r.Auth != nil {
		auth = r.Auth.AuthType
		if r.Auth.Required {
			auth += "!"
		}
	}

	// Return type
	returns := ""
	if r.ReturnType != nil {
		returns = typeToString(r.ReturnType)
	}

	// Compute hash
	hashInput := r.Method.String() + r.Path + returns

	return &RouteInfo{
		Method:      r.Method.String(),
		Path:        r.Path,
		Params:      params,
		QueryParams: queryParams,
		Returns:     returns,
		Auth:        auth,
		Injects:     injects,
		Hash:        computeHash(hashInput)[:8],
	}
}

// extractFunctionInfo extracts compact function information
func extractFunctionInfo(f *interpreter.Function) *FunctionInfo {
	params := make([]string, len(f.Params))
	for i, p := range f.Params {
		paramStr := p.Name + ": " + typeToString(p.TypeAnnotation)
		if p.Required {
			paramStr += "!"
		}
		params[i] = paramStr
	}

	typeParams := make([]string, len(f.TypeParams))
	for i, tp := range f.TypeParams {
		typeParams[i] = tp.Name
	}

	returns := ""
	if f.ReturnType != nil {
		returns = typeToString(f.ReturnType)
	}

	hashInput := f.Name + ":" + strings.Join(params, ",") + "->" + returns

	return &FunctionInfo{
		Name:       f.Name,
		Params:     params,
		Returns:    returns,
		TypeParams: typeParams,
		Hash:       computeHash(hashInput)[:8],
	}
}

// extractCommandInfo extracts compact command information
func extractCommandInfo(c *interpreter.Command) *CommandInfo {
	params := make([]string, len(c.Params))
	for i, p := range c.Params {
		prefix := ""
		if p.IsFlag {
			prefix = "--"
		}
		paramStr := prefix + p.Name
		if p.Type != nil {
			paramStr += ": " + typeToString(p.Type)
		}
		if p.Required {
			paramStr += "!"
		}
		params[i] = paramStr
	}

	hashInput := c.Name + ":" + strings.Join(params, ",")

	return &CommandInfo{
		Name:        c.Name,
		Params:      params,
		Description: c.Description,
		Hash:        computeHash(hashInput)[:8],
	}
}

// detectPatterns detects common patterns in the project
func (g *Generator) detectPatterns(ctx *ProjectContext) []string {
	var patterns []string

	// Detect CRUD patterns
	routesByResource := make(map[string][]string)
	for _, r := range ctx.Routes {
		// Extract resource name from path (e.g., /users -> users, /api/v1/products -> products)
		parts := strings.Split(strings.Trim(r.Path, "/"), "/")
		if len(parts) > 0 {
			resource := parts[len(parts)-1]
			// Remove path params
			if !strings.HasPrefix(resource, ":") {
				routesByResource[resource] = append(routesByResource[resource], r.Method)
			} else if len(parts) > 1 {
				resource = parts[len(parts)-2]
				routesByResource[resource] = append(routesByResource[resource], r.Method)
			}
		}
	}

	for resource, methods := range routesByResource {
		hasGet := contains(methods, "GET")
		hasPost := contains(methods, "POST")
		hasPut := contains(methods, "PUT") || contains(methods, "PATCH")
		hasDelete := contains(methods, "DELETE")

		if hasGet && hasPost && hasPut && hasDelete {
			patterns = append(patterns, fmt.Sprintf("crud(%s)", resource))
		} else if hasGet && hasPost {
			patterns = append(patterns, fmt.Sprintf("read_create(%s)", resource))
		}
	}

	// Detect auth usage
	authRoutes := 0
	for _, r := range ctx.Routes {
		if r.Auth != "" {
			authRoutes++
		}
	}
	if authRoutes > 0 {
		patterns = append(patterns, fmt.Sprintf("auth_routes(%d)", authRoutes))
	}

	// Detect database usage
	dbRoutes := 0
	for _, r := range ctx.Routes {
		if contains(r.Injects, "db") || contains(r.Injects, "database") {
			dbRoutes++
		}
	}
	if dbRoutes > 0 {
		patterns = append(patterns, fmt.Sprintf("database_routes(%d)", dbRoutes))
	}

	return patterns
}

// typeToString converts a Type to a compact string representation
func typeToString(t interpreter.Type) string {
	if t == nil {
		return "any"
	}

	switch v := t.(type) {
	case interpreter.IntType:
		return "int"
	case interpreter.StringType:
		return "string"
	case interpreter.BoolType:
		return "bool"
	case interpreter.FloatType:
		return "float"
	case interpreter.NamedType:
		return v.Name
	case interpreter.ArrayType:
		return "[" + typeToString(v.ElementType) + "]"
	case interpreter.OptionalType:
		return typeToString(v.InnerType) + "?"
	case interpreter.GenericType:
		args := make([]string, len(v.TypeArgs))
		for i, arg := range v.TypeArgs {
			args[i] = typeToString(arg)
		}
		return typeToString(v.BaseType) + "<" + strings.Join(args, ", ") + ">"
	case interpreter.UnionType:
		types := make([]string, len(v.Types))
		for i, ut := range v.Types {
			types[i] = typeToString(ut)
		}
		return strings.Join(types, " | ")
	case interpreter.FunctionType:
		params := make([]string, len(v.ParamTypes))
		for i, pt := range v.ParamTypes {
			params[i] = typeToString(pt)
		}
		return "(" + strings.Join(params, ", ") + ") -> " + typeToString(v.ReturnType)
	case interpreter.DatabaseType:
		return "Database"
	default:
		return "any"
	}
}

// computeHash computes a SHA256 hash and returns hex string
func computeHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// ToJSON serializes the context to JSON
func (ctx *ProjectContext) ToJSON(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(ctx, "", "  ")
	}
	return json.Marshal(ctx)
}

// ToCompact generates a compact text representation optimized for AI consumption
func (ctx *ProjectContext) ToCompact() string {
	var sb strings.Builder

	sb.WriteString("# Glyph Project Context\n")
	sb.WriteString(fmt.Sprintf("# Hash: %s\n\n", ctx.ProjectHash[:12]))

	// Types section
	if len(ctx.Types) > 0 {
		sb.WriteString("## Types\n")
		for name, t := range ctx.Types {
			if len(t.TypeParams) > 0 {
				sb.WriteString(fmt.Sprintf(": %s<%s> { %s }\n", name, strings.Join(t.TypeParams, ", "), strings.Join(t.Fields, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf(": %s { %s }\n", name, strings.Join(t.Fields, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	// Routes section
	if len(ctx.Routes) > 0 {
		sb.WriteString("## Routes\n")
		for _, r := range ctx.Routes {
			line := fmt.Sprintf("@ %s %s", r.Method, r.Path)
			if r.Auth != "" {
				line += fmt.Sprintf(" [%s]", r.Auth)
			}
			if r.Returns != "" {
				line += fmt.Sprintf(" -> %s", r.Returns)
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Functions section
	if len(ctx.Functions) > 0 {
		sb.WriteString("## Functions\n")
		for name, f := range ctx.Functions {
			if len(f.TypeParams) > 0 {
				sb.WriteString(fmt.Sprintf("fn %s<%s>(%s) -> %s\n", name, strings.Join(f.TypeParams, ", "), strings.Join(f.Params, ", "), f.Returns))
			} else {
				sb.WriteString(fmt.Sprintf("fn %s(%s) -> %s\n", name, strings.Join(f.Params, ", "), f.Returns))
			}
		}
		sb.WriteString("\n")
	}

	// Commands section
	if len(ctx.Commands) > 0 {
		sb.WriteString("## Commands\n")
		for name, c := range ctx.Commands {
			sb.WriteString(fmt.Sprintf("@ command %s %s\n", name, strings.Join(c.Params, " ")))
		}
		sb.WriteString("\n")
	}

	// Patterns section
	if len(ctx.Patterns) > 0 {
		sb.WriteString("## Patterns\n")
		for _, p := range ctx.Patterns {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
	}

	return sb.String()
}

// ContextDiff represents changes between two context versions
type ContextDiff struct {
	PreviousHash string      `json:"previous_hash"`
	CurrentHash  string      `json:"current_hash"`
	HasChanges   bool        `json:"has_changes"`
	Files        *FilesDiff  `json:"files,omitempty"`
	Types        *ItemsDiff  `json:"types,omitempty"`
	Routes       *RoutesDiff `json:"routes,omitempty"`
	Functions    *ItemsDiff  `json:"functions,omitempty"`
	Commands     *ItemsDiff  `json:"commands,omitempty"`
}

// FilesDiff represents file-level changes
type FilesDiff struct {
	Added    []string `json:"added,omitempty"`
	Removed  []string `json:"removed,omitempty"`
	Modified []string `json:"modified,omitempty"`
}

// ItemsDiff represents changes to named items (types, functions, commands)
type ItemsDiff struct {
	Added    []string `json:"added,omitempty"`
	Removed  []string `json:"removed,omitempty"`
	Modified []string `json:"modified,omitempty"`
}

// RoutesDiff represents changes to routes
type RoutesDiff struct {
	Added    []*RouteInfo `json:"added,omitempty"`
	Removed  []*RouteInfo `json:"removed,omitempty"`
	Modified []*RouteInfo `json:"modified,omitempty"`
}

// DefaultContextPath returns the default path for context storage
func DefaultContextPath(rootDir string) string {
	return filepath.Join(rootDir, ".glyph", "context.json")
}

// LoadContext loads a previously saved context from disk
func LoadContext(path string) (*ProjectContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ctx ProjectContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to parse context: %w", err)
	}

	return &ctx, nil
}

// SaveContext saves the context to disk
func (ctx *ProjectContext) SaveContext(path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := ctx.ToJSON(true)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Diff compares the current context with a previous context and returns the differences
func (ctx *ProjectContext) Diff(previous *ProjectContext) *ContextDiff {
	diff := &ContextDiff{
		CurrentHash: ctx.ProjectHash,
	}

	if previous != nil {
		diff.PreviousHash = previous.ProjectHash
	}

	// If hashes match, no changes
	if previous != nil && ctx.ProjectHash == previous.ProjectHash {
		diff.HasChanges = false
		return diff
	}

	diff.HasChanges = true

	// Compare files
	diff.Files = diffFiles(previous, ctx)

	// Compare types
	diff.Types = diffItems(
		getTypeNames(previous),
		getTypeNames(ctx),
		getTypeHashes(previous),
		getTypeHashes(ctx),
	)

	// Compare routes
	diff.Routes = diffRoutes(previous, ctx)

	// Compare functions
	diff.Functions = diffItems(
		getFunctionNames(previous),
		getFunctionNames(ctx),
		getFunctionHashes(previous),
		getFunctionHashes(ctx),
	)

	// Compare commands
	diff.Commands = diffItems(
		getCommandNames(previous),
		getCommandNames(ctx),
		getCommandHashes(previous),
		getCommandHashes(ctx),
	)

	return diff
}

// diffFiles compares files between two contexts
func diffFiles(prev, curr *ProjectContext) *FilesDiff {
	diff := &FilesDiff{}

	prevFiles := make(map[string]string)
	if prev != nil {
		for path, f := range prev.Files {
			prevFiles[path] = f.Hash
		}
	}

	currFiles := make(map[string]string)
	for path, f := range curr.Files {
		currFiles[path] = f.Hash
	}

	// Find added and modified
	for path, hash := range currFiles {
		if prevHash, exists := prevFiles[path]; !exists {
			diff.Added = append(diff.Added, path)
		} else if prevHash != hash {
			diff.Modified = append(diff.Modified, path)
		}
	}

	// Find removed
	for path := range prevFiles {
		if _, exists := currFiles[path]; !exists {
			diff.Removed = append(diff.Removed, path)
		}
	}

	if len(diff.Added) == 0 && len(diff.Removed) == 0 && len(diff.Modified) == 0 {
		return nil
	}
	return diff
}

// diffItems compares named items between two contexts
func diffItems(prevNames, currNames []string, prevHashes, currHashes map[string]string) *ItemsDiff {
	diff := &ItemsDiff{}

	prevSet := make(map[string]bool)
	for _, name := range prevNames {
		prevSet[name] = true
	}

	currSet := make(map[string]bool)
	for _, name := range currNames {
		currSet[name] = true
	}

	// Find added and modified
	for _, name := range currNames {
		if !prevSet[name] {
			diff.Added = append(diff.Added, name)
		} else if prevHashes[name] != currHashes[name] {
			diff.Modified = append(diff.Modified, name)
		}
	}

	// Find removed
	for _, name := range prevNames {
		if !currSet[name] {
			diff.Removed = append(diff.Removed, name)
		}
	}

	if len(diff.Added) == 0 && len(diff.Removed) == 0 && len(diff.Modified) == 0 {
		return nil
	}
	return diff
}

// diffRoutes compares routes between two contexts
func diffRoutes(prev, curr *ProjectContext) *RoutesDiff {
	diff := &RoutesDiff{}

	// Create route keys for comparison (method + path)
	prevRoutes := make(map[string]*RouteInfo)
	if prev != nil {
		for _, r := range prev.Routes {
			key := r.Method + " " + r.Path
			prevRoutes[key] = r
		}
	}

	currRoutes := make(map[string]*RouteInfo)
	for _, r := range curr.Routes {
		key := r.Method + " " + r.Path
		currRoutes[key] = r
	}

	// Find added and modified
	for key, route := range currRoutes {
		if prevRoute, exists := prevRoutes[key]; !exists {
			diff.Added = append(diff.Added, route)
		} else if prevRoute.Hash != route.Hash {
			diff.Modified = append(diff.Modified, route)
		}
	}

	// Find removed
	for key, route := range prevRoutes {
		if _, exists := currRoutes[key]; !exists {
			diff.Removed = append(diff.Removed, route)
		}
	}

	if len(diff.Added) == 0 && len(diff.Removed) == 0 && len(diff.Modified) == 0 {
		return nil
	}
	return diff
}

// Helper functions for extracting names and hashes

func getTypeNames(ctx *ProjectContext) []string {
	if ctx == nil {
		return nil
	}
	names := make([]string, 0, len(ctx.Types))
	for name := range ctx.Types {
		names = append(names, name)
	}
	return names
}

func getTypeHashes(ctx *ProjectContext) map[string]string {
	if ctx == nil {
		return nil
	}
	hashes := make(map[string]string)
	for name, t := range ctx.Types {
		hashes[name] = t.Hash
	}
	return hashes
}

func getFunctionNames(ctx *ProjectContext) []string {
	if ctx == nil {
		return nil
	}
	names := make([]string, 0, len(ctx.Functions))
	for name := range ctx.Functions {
		names = append(names, name)
	}
	return names
}

func getFunctionHashes(ctx *ProjectContext) map[string]string {
	if ctx == nil {
		return nil
	}
	hashes := make(map[string]string)
	for name, f := range ctx.Functions {
		hashes[name] = f.Hash
	}
	return hashes
}

func getCommandNames(ctx *ProjectContext) []string {
	if ctx == nil {
		return nil
	}
	names := make([]string, 0, len(ctx.Commands))
	for name := range ctx.Commands {
		names = append(names, name)
	}
	return names
}

func getCommandHashes(ctx *ProjectContext) map[string]string {
	if ctx == nil {
		return nil
	}
	hashes := make(map[string]string)
	for name, c := range ctx.Commands {
		hashes[name] = c.Hash
	}
	return hashes
}

// ToJSON serializes the diff to JSON
func (d *ContextDiff) ToJSON(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(d, "", "  ")
	}
	return json.Marshal(d)
}

// ToCompact generates a compact text representation of the diff
func (d *ContextDiff) ToCompact(ctx *ProjectContext) string {
	var sb strings.Builder

	sb.WriteString("# Glyph Context Changes\n")
	sb.WriteString(fmt.Sprintf("# Previous: %s\n", truncateHash(d.PreviousHash)))
	sb.WriteString(fmt.Sprintf("# Current:  %s\n\n", truncateHash(d.CurrentHash)))

	if !d.HasChanges {
		sb.WriteString("No changes detected.\n")
		return sb.String()
	}

	// Files changes
	if d.Files != nil {
		sb.WriteString("## Files\n")
		for _, f := range d.Files.Added {
			sb.WriteString(fmt.Sprintf("+ %s\n", f))
		}
		for _, f := range d.Files.Removed {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		for _, f := range d.Files.Modified {
			sb.WriteString(fmt.Sprintf("~ %s\n", f))
		}
		sb.WriteString("\n")
	}

	// Type changes
	if d.Types != nil {
		sb.WriteString("## Types\n")
		for _, name := range d.Types.Added {
			if t, ok := ctx.Types[name]; ok {
				sb.WriteString(fmt.Sprintf("+ : %s { %s }\n", name, strings.Join(t.Fields, ", ")))
			}
		}
		for _, name := range d.Types.Removed {
			sb.WriteString(fmt.Sprintf("- : %s\n", name))
		}
		for _, name := range d.Types.Modified {
			if t, ok := ctx.Types[name]; ok {
				sb.WriteString(fmt.Sprintf("~ : %s { %s }\n", name, strings.Join(t.Fields, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	// Route changes
	if d.Routes != nil {
		sb.WriteString("## Routes\n")
		for _, r := range d.Routes.Added {
			line := fmt.Sprintf("+ @ %s %s", r.Method, r.Path)
			if r.Returns != "" {
				line += fmt.Sprintf(" -> %s", r.Returns)
			}
			sb.WriteString(line + "\n")
		}
		for _, r := range d.Routes.Removed {
			sb.WriteString(fmt.Sprintf("- @ %s %s\n", r.Method, r.Path))
		}
		for _, r := range d.Routes.Modified {
			line := fmt.Sprintf("~ @ %s %s", r.Method, r.Path)
			if r.Returns != "" {
				line += fmt.Sprintf(" -> %s", r.Returns)
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Function changes
	if d.Functions != nil {
		sb.WriteString("## Functions\n")
		for _, name := range d.Functions.Added {
			if f, ok := ctx.Functions[name]; ok {
				sb.WriteString(fmt.Sprintf("+ fn %s(%s) -> %s\n", name, strings.Join(f.Params, ", "), f.Returns))
			}
		}
		for _, name := range d.Functions.Removed {
			sb.WriteString(fmt.Sprintf("- fn %s\n", name))
		}
		for _, name := range d.Functions.Modified {
			if f, ok := ctx.Functions[name]; ok {
				sb.WriteString(fmt.Sprintf("~ fn %s(%s) -> %s\n", name, strings.Join(f.Params, ", "), f.Returns))
			}
		}
		sb.WriteString("\n")
	}

	// Command changes
	if d.Commands != nil {
		sb.WriteString("## Commands\n")
		for _, name := range d.Commands.Added {
			sb.WriteString(fmt.Sprintf("+ @ command %s\n", name))
		}
		for _, name := range d.Commands.Removed {
			sb.WriteString(fmt.Sprintf("- @ command %s\n", name))
		}
		for _, name := range d.Commands.Modified {
			sb.WriteString(fmt.Sprintf("~ @ command %s\n", name))
		}
	}

	return sb.String()
}

// truncateHash returns a truncated hash for display
func truncateHash(hash string) string {
	if hash == "" {
		return "(none)"
	}
	if len(hash) > 12 {
		return hash[:12]
	}
	return hash
}

// TargetedContext represents context optimized for a specific task
type TargetedContext struct {
	Task        string                   `json:"task"`
	Description string                   `json:"description"`
	Types       map[string]*TypeInfo     `json:"types,omitempty"`
	Routes      []*RouteInfo             `json:"routes,omitempty"`
	Functions   map[string]*FunctionInfo `json:"functions,omitempty"`
	Commands    map[string]*CommandInfo  `json:"commands,omitempty"`
	Patterns    []string                 `json:"patterns,omitempty"`
	Injections  []string                 `json:"available_injections,omitempty"`
	Syntax      *SyntaxGuide             `json:"syntax,omitempty"`
}

// SyntaxGuide provides quick reference for Glyph syntax
type SyntaxGuide struct {
	Examples []string `json:"examples"`
	Notes    []string `json:"notes,omitempty"`
}

// ForRoute generates context optimized for writing a new route
func (ctx *ProjectContext) ForRoute() *TargetedContext {
	tc := &TargetedContext{
		Task:        "route",
		Description: "Context for creating a new HTTP route",
		Types:       ctx.Types,
		Routes:      ctx.Routes,
		Functions:   ctx.Functions,
		Patterns:    ctx.Patterns,
	}

	// Extract available injections from existing routes
	injectionSet := make(map[string]bool)
	for _, r := range ctx.Routes {
		for _, inj := range r.Injects {
			injectionSet[inj] = true
		}
	}
	for inj := range injectionSet {
		tc.Injections = append(tc.Injections, inj)
	}

	// Add syntax guide for routes
	tc.Syntax = &SyntaxGuide{
		Examples: []string{
			"@ GET /path -> ReturnType",
			"@ POST /path/:param -> ReturnType",
			"  + ratelimit(100/min)",
			"  + auth(jwt)",
			"  % db: Database",
			"  $ result = db.collection.find()",
			"  > { field: value }",
		},
		Notes: []string{
			"Use : for path parameters (e.g., /users/:id)",
			"Use + for middleware (auth, ratelimit)",
			"Use % for dependency injection",
			"Use $ for variable assignment",
			"Use > for return statement",
			"Use input to access request body (POST/PUT)",
			"Use query to access query parameters",
		},
	}

	return tc
}

// ForType generates context optimized for writing a new type
func (ctx *ProjectContext) ForType() *TargetedContext {
	tc := &TargetedContext{
		Task:        "type",
		Description: "Context for creating a new type definition",
		Types:       ctx.Types,
	}

	tc.Syntax = &SyntaxGuide{
		Examples: []string{
			": TypeName {",
			"  field: str!        # required string",
			"  optional: int      # optional int",
			"  list: [OtherType]  # array of types",
			"  nested: SubType    # reference to other type",
			"}",
		},
		Notes: []string{
			"Use ! suffix for required fields",
			"Basic types: str, int, bool, float, timestamp",
			"Use [Type] for arrays",
			"Use Type? for optional types",
			"Generic types: List<T>, Map<K,V>, Result<T,E>",
		},
	}

	return tc
}

// ForFunction generates context optimized for writing a new function
func (ctx *ProjectContext) ForFunction() *TargetedContext {
	tc := &TargetedContext{
		Task:        "function",
		Description: "Context for creating a new function",
		Types:       ctx.Types,
		Functions:   ctx.Functions,
	}

	tc.Syntax = &SyntaxGuide{
		Examples: []string{
			"! functionName(param: Type): ReturnType",
			"  $ result = someOperation()",
			"  > result",
			"",
			"! genericFn<T>(items: [T]): T",
			"  > items[0]",
		},
		Notes: []string{
			"Use ! to define a function",
			"Parameters: name: Type",
			"Return type after :",
			"Use <T> for generic type parameters",
			"Use $ for variable assignment",
			"Use > for return statement",
		},
	}

	return tc
}

// ForCommand generates context optimized for writing a new CLI command
func (ctx *ProjectContext) ForCommand() *TargetedContext {
	tc := &TargetedContext{
		Task:        "command",
		Description: "Context for creating a new CLI command",
		Types:       ctx.Types,
		Commands:    ctx.Commands,
		Functions:   ctx.Functions,
	}

	tc.Syntax = &SyntaxGuide{
		Examples: []string{
			"@ command name arg: str! --flag: bool",
			"  $ result = processArg(arg)",
			"  if flag {",
			"    > { verbose: true, result: result }",
			"  }",
			"  > result",
		},
		Notes: []string{
			"Use @ command to define a CLI command",
			"Positional args: name: type",
			"Flags: --name: type",
			"Use ! suffix for required args/flags",
			"Use $ for variable assignment",
			"Use > for return/output",
		},
	}

	return tc
}

// ToJSON serializes the targeted context to JSON
func (tc *TargetedContext) ToJSON(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(tc, "", "  ")
	}
	return json.Marshal(tc)
}

// ToCompact generates a compact text representation
func (tc *TargetedContext) ToCompact() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Glyph Context: %s\n", tc.Task))
	sb.WriteString(fmt.Sprintf("# %s\n\n", tc.Description))

	// Syntax guide first (most important for AI)
	if tc.Syntax != nil {
		sb.WriteString("## Syntax\n```\n")
		for _, ex := range tc.Syntax.Examples {
			sb.WriteString(ex + "\n")
		}
		sb.WriteString("```\n")
		if len(tc.Syntax.Notes) > 0 {
			for _, note := range tc.Syntax.Notes {
				sb.WriteString(fmt.Sprintf("- %s\n", note))
			}
		}
		sb.WriteString("\n")
	}

	// Available injections
	if len(tc.Injections) > 0 {
		sb.WriteString("## Available Injections\n")
		for _, inj := range tc.Injections {
			sb.WriteString(fmt.Sprintf("- %s\n", inj))
		}
		sb.WriteString("\n")
	}

	// Types
	if len(tc.Types) > 0 {
		sb.WriteString("## Types\n")
		for name, t := range tc.Types {
			if len(t.TypeParams) > 0 {
				sb.WriteString(fmt.Sprintf(": %s<%s> { %s }\n", name, strings.Join(t.TypeParams, ", "), strings.Join(t.Fields, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf(": %s { %s }\n", name, strings.Join(t.Fields, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	// Existing routes (for reference)
	if len(tc.Routes) > 0 {
		sb.WriteString("## Existing Routes\n")
		for _, r := range tc.Routes {
			line := fmt.Sprintf("@ %s %s", r.Method, r.Path)
			if r.Auth != "" {
				line += fmt.Sprintf(" [%s]", r.Auth)
			}
			if r.Returns != "" {
				line += fmt.Sprintf(" -> %s", r.Returns)
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Functions
	if len(tc.Functions) > 0 {
		sb.WriteString("## Functions\n")
		for name, f := range tc.Functions {
			if len(f.TypeParams) > 0 {
				sb.WriteString(fmt.Sprintf("! %s<%s>(%s): %s\n", name, strings.Join(f.TypeParams, ", "), strings.Join(f.Params, ", "), f.Returns))
			} else {
				sb.WriteString(fmt.Sprintf("! %s(%s): %s\n", name, strings.Join(f.Params, ", "), f.Returns))
			}
		}
		sb.WriteString("\n")
	}

	// Commands
	if len(tc.Commands) > 0 {
		sb.WriteString("## Commands\n")
		for name, c := range tc.Commands {
			sb.WriteString(fmt.Sprintf("@ command %s %s\n", name, strings.Join(c.Params, " ")))
		}
		sb.WriteString("\n")
	}

	// Patterns
	if len(tc.Patterns) > 0 {
		sb.WriteString("## Patterns\n")
		for _, p := range tc.Patterns {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
	}

	return sb.String()
}

// GenerateStubs generates type stub content (.glyph.d format)
func (ctx *ProjectContext) GenerateStubs() string {
	var sb strings.Builder

	sb.WriteString("# Auto-generated Glyph type stubs\n")
	sb.WriteString(fmt.Sprintf("# Generated: %s\n", ctx.Generated.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("# Hash: %s\n\n", ctx.ProjectHash[:12]))

	// Types
	for _, t := range ctx.Types {
		if len(t.TypeParams) > 0 {
			sb.WriteString(fmt.Sprintf(": %s<%s>\n", t.Name, strings.Join(t.TypeParams, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf(": %s\n", t.Name))
		}
		for _, field := range t.Fields {
			sb.WriteString(fmt.Sprintf("  %s\n", field))
		}
		sb.WriteString("\n")
	}

	// Function signatures
	for _, f := range ctx.Functions {
		params := strings.Join(f.Params, ", ")
		if len(f.TypeParams) > 0 {
			sb.WriteString(fmt.Sprintf("! %s<%s>(%s): %s\n", f.Name, strings.Join(f.TypeParams, ", "), params, f.Returns))
		} else {
			sb.WriteString(fmt.Sprintf("! %s(%s): %s\n", f.Name, params, f.Returns))
		}
	}

	if len(ctx.Functions) > 0 {
		sb.WriteString("\n")
	}

	// Route signatures
	for _, r := range ctx.Routes {
		line := fmt.Sprintf("@ %s %s", r.Method, r.Path)
		if len(r.Injects) > 0 {
			line += fmt.Sprintf(" +%s", strings.Join(r.Injects, " +"))
		}
		if r.Auth != "" {
			line += fmt.Sprintf(" [%s]", r.Auth)
		}
		if r.Returns != "" {
			line += fmt.Sprintf(" -> %s", r.Returns)
		}
		sb.WriteString(line + "\n")
	}

	return sb.String()
}
