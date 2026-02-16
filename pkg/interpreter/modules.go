package interpreter

import (
	. "github.com/glyphlang/glyph/pkg/ast"

	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModuleResolver handles module resolution and loading
type ModuleResolver struct {
	// SearchPaths contains directories to search for modules
	SearchPaths []string

	// ModuleCache stores already parsed modules to avoid re-parsing
	ModuleCache map[string]*LoadedModule

	// loadingStack tracks modules currently being loaded for circular dependency detection
	loadingStack []string

	// Parser function to parse source code (injected to avoid import cycles)
	ParseFunc func(source string) (*Module, error)
}

// LoadedModule represents a parsed and loaded module
type LoadedModule struct {
	Path      string                 // The resolved file path
	Module    *Module                // The parsed AST
	Exports   map[string]interface{} // Exported items (functions, types, etc.)
	Namespace string                 // Module namespace if declared
}

// NewModuleResolver creates a new module resolver
func NewModuleResolver() *ModuleResolver {
	return &ModuleResolver{
		SearchPaths:  []string{"."},
		ModuleCache:  make(map[string]*LoadedModule),
		loadingStack: []string{},
	}
}

// AddSearchPath adds a directory to the search paths
func (r *ModuleResolver) AddSearchPath(path string) {
	r.SearchPaths = append(r.SearchPaths, path)
}

// SetParseFunc sets the parsing function
func (r *ModuleResolver) SetParseFunc(fn func(source string) (*Module, error)) {
	r.ParseFunc = fn
}

// ResolveModule resolves and loads a module from a path
// The basePath is the directory of the importing file
func (r *ModuleResolver) ResolveModule(importPath string, basePath string) (*LoadedModule, error) {
	// Resolve the full file path
	fullPath, err := r.resolvePath(importPath, basePath)
	if err != nil {
		return nil, err
	}

	// Check cache first
	if cached, ok := r.ModuleCache[fullPath]; ok {
		return cached, nil
	}

	// Check for circular dependency
	for _, loading := range r.loadingStack {
		if loading == fullPath {
			return nil, fmt.Errorf("circular dependency detected: %s", r.formatCircularDependency(fullPath))
		}
	}

	// Add to loading stack
	r.loadingStack = append(r.loadingStack, fullPath)
	defer func() {
		r.loadingStack = r.loadingStack[:len(r.loadingStack)-1]
	}()

	// Load and parse the module
	loadedModule, err := r.loadModule(fullPath)
	if err != nil {
		return nil, err
	}

	// Cache the module
	r.ModuleCache[fullPath] = loadedModule

	return loadedModule, nil
}

// resolvePath resolves an import path to a full file path
func (r *ModuleResolver) resolvePath(importPath string, basePath string) (string, error) {
	// Handle relative paths (starting with ./ or ../)
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		// Resolve relative to the base path
		fullPath := filepath.Join(basePath, importPath)
		return r.findGlyphFile(fullPath)
	}

	// Handle absolute paths
	if filepath.IsAbs(importPath) {
		return r.findGlyphFile(importPath)
	}

	// Search in search paths for package imports
	for _, searchPath := range r.SearchPaths {
		fullPath := filepath.Join(searchPath, importPath)
		resolved, err := r.findGlyphFile(fullPath)
		if err == nil {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("module not found: %s", importPath)
}

// findGlyphFile finds a .glyph file, handling both with and without extension
func (r *ModuleResolver) findGlyphFile(path string) (string, error) {
	// If path already has .glyph extension
	if strings.HasSuffix(path, ".glyph") {
		if _, err := os.Stat(path); err == nil {
			return filepath.Abs(path)
		}
		return "", fmt.Errorf("file not found: %s", path)
	}

	// Try with .glyph extension
	withExt := path + ".glyph"
	if _, err := os.Stat(withExt); err == nil {
		return filepath.Abs(withExt)
	}

	// Try as directory with main.glyph
	mainFile := filepath.Join(path, "main.glyph")
	if _, err := os.Stat(mainFile); err == nil {
		return filepath.Abs(mainFile)
	}

	// Try as directory with index.glyph
	indexFile := filepath.Join(path, "index.glyph")
	if _, err := os.Stat(indexFile); err == nil {
		return filepath.Abs(indexFile)
	}

	return "", fmt.Errorf("module not found: %s (tried %s, %s, %s)", path, withExt, mainFile, indexFile)
}

// loadModule loads and parses a module file
func (r *ModuleResolver) loadModule(fullPath string) (*LoadedModule, error) {
	// Read the file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module %s: %w", fullPath, err)
	}

	// Check if we have a parser function
	if r.ParseFunc == nil {
		return nil, fmt.Errorf("no parser function set in ModuleResolver")
	}

	// Parse the module
	module, err := r.ParseFunc(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse module %s: %w", fullPath, err)
	}

	// Extract exports from the module
	exports := r.extractExports(module)

	// Extract module namespace if declared
	namespace := r.extractNamespace(module)

	return &LoadedModule{
		Path:      fullPath,
		Module:    module,
		Exports:   exports,
		Namespace: namespace,
	}, nil
}

// extractExports extracts exportable items from a module
func (r *ModuleResolver) extractExports(module *Module) map[string]interface{} {
	exports := make(map[string]interface{})

	for _, item := range module.Items {
		switch it := item.(type) {
		case *Function:
			// All top-level functions are exported
			exports[it.Name] = it
		case *TypeDef:
			// All top-level types are exported
			exports[it.Name] = it
		case *Route:
			// Routes are exported by their path+method
			key := fmt.Sprintf("%s:%s", it.Method.String(), it.Path)
			exports[key] = it
		case *Command:
			// Commands are exported by name
			exports[it.Name] = it
		case *ConstDecl:
			// All top-level constants are exported
			exports[it.Name] = it
		case *ProviderDef:
			// Provider contracts are exported by name
			exports[it.Name] = it
		}
	}

	return exports
}

// extractNamespace extracts the module namespace if declared
func (r *ModuleResolver) extractNamespace(module *Module) string {
	for _, item := range module.Items {
		if decl, ok := item.(*ModuleDecl); ok {
			return decl.Name
		}
	}
	return ""
}

// formatCircularDependency formats the circular dependency chain for error message
func (r *ModuleResolver) formatCircularDependency(lastPath string) string {
	chain := make([]string, len(r.loadingStack)+1)
	copy(chain, r.loadingStack)
	chain[len(chain)-1] = lastPath
	return strings.Join(chain, " -> ")
}

// GetCachedModule retrieves a module from cache by its resolved path
func (r *ModuleResolver) GetCachedModule(path string) (*LoadedModule, bool) {
	module, ok := r.ModuleCache[path]
	return module, ok
}

// ClearCache clears the module cache
func (r *ModuleResolver) ClearCache() {
	r.ModuleCache = make(map[string]*LoadedModule)
}

// ProcessImports processes all import statements in a module
// Returns a map of alias/name -> loaded module
func (r *ModuleResolver) ProcessImports(module *Module, basePath string) (map[string]*LoadedModule, error) {
	imports := make(map[string]*LoadedModule)

	for _, item := range module.Items {
		importStmt, ok := item.(*ImportStatement)
		if !ok {
			continue
		}

		// Resolve and load the imported module
		loadedModule, err := r.ResolveModule(importStmt.Path, basePath)
		if err != nil {
			return nil, fmt.Errorf("failed to import %s: %w", importStmt.Path, err)
		}

		if importStmt.Selective {
			// For selective imports, validate that all names exist
			for _, name := range importStmt.Names {
				if _, exists := loadedModule.Exports[name.Name]; !exists {
					return nil, fmt.Errorf("'%s' is not exported from module '%s'", name.Name, importStmt.Path)
				}
			}
		}

		// Determine the key for this import
		var key string
		if importStmt.Alias != "" {
			key = importStmt.Alias
		} else if importStmt.Selective {
			// For selective imports, we don't create a module-level key
			// The names are directly imported
			key = importStmt.Path
		} else {
			// Use the last segment of the path as the default name
			key = filepath.Base(importStmt.Path)
			key = strings.TrimSuffix(key, ".glyph")
		}

		imports[key] = loadedModule
	}

	return imports, nil
}
