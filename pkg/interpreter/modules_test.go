package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test ModuleResolver

func TestModuleResolver_NewModuleResolver(t *testing.T) {
	resolver := NewModuleResolver()

	assert.NotNil(t, resolver)
	assert.NotNil(t, resolver.ModuleCache)
	assert.Equal(t, []string{"."}, resolver.SearchPaths)
}

func TestModuleResolver_AddSearchPath(t *testing.T) {
	resolver := NewModuleResolver()
	resolver.AddSearchPath("/custom/path")

	assert.Contains(t, resolver.SearchPaths, "/custom/path")
}

func TestModuleResolver_SetParseFunc(t *testing.T) {
	resolver := NewModuleResolver()

	parseFunc := func(source string) (*Module, error) {
		return &Module{}, nil
	}

	resolver.SetParseFunc(parseFunc)
	assert.NotNil(t, resolver.ParseFunc)
}

func TestModuleResolver_ClearCache(t *testing.T) {
	resolver := NewModuleResolver()
	resolver.ModuleCache["test"] = &LoadedModule{Path: "test"}

	resolver.ClearCache()

	assert.Empty(t, resolver.ModuleCache)
}

func TestModuleResolver_GetCachedModule(t *testing.T) {
	resolver := NewModuleResolver()
	loadedModule := &LoadedModule{Path: "test.glyph"}
	resolver.ModuleCache["test.glyph"] = loadedModule

	cached, ok := resolver.GetCachedModule("test.glyph")
	assert.True(t, ok)
	assert.Equal(t, loadedModule, cached)

	_, ok = resolver.GetCachedModule("nonexistent")
	assert.False(t, ok)
}

func TestModuleResolver_ExtractExports(t *testing.T) {
	resolver := NewModuleResolver()

	module := &Module{
		Items: []Item{
			&Function{Name: "myFunc"},
			&TypeDef{Name: "MyType"},
			&Route{Method: Get, Path: "/test"},
			&Command{Name: "myCommand"},
		},
	}

	exports := resolver.extractExports(module)

	assert.Contains(t, exports, "myFunc")
	assert.Contains(t, exports, "MyType")
	assert.Contains(t, exports, "GET:/test")
	assert.Contains(t, exports, "myCommand")
}

func TestModuleResolver_ExtractNamespace(t *testing.T) {
	resolver := NewModuleResolver()

	// Module with namespace declaration
	moduleWithNs := &Module{
		Items: []Item{
			&ModuleDecl{Name: "myapp/utils"},
			&Function{Name: "helper"},
		},
	}

	ns := resolver.extractNamespace(moduleWithNs)
	assert.Equal(t, "myapp/utils", ns)

	// Module without namespace
	moduleWithoutNs := &Module{
		Items: []Item{
			&Function{Name: "helper"},
		},
	}

	ns = resolver.extractNamespace(moduleWithoutNs)
	assert.Empty(t, ns)
}

func TestModuleResolver_ResolveModule_Cached(t *testing.T) {
	resolver := NewModuleResolver()

	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.glyph")
	err := os.WriteFile(testFile, []byte("! hello() { }"), 0644)
	require.NoError(t, err)

	absPath, _ := filepath.Abs(testFile)

	// Pre-cache the module
	cachedModule := &LoadedModule{
		Path:    absPath,
		Module:  &Module{},
		Exports: map[string]interface{}{"hello": &Function{Name: "hello"}},
	}
	resolver.ModuleCache[absPath] = cachedModule

	// Set a parse function (shouldn't be called if cached)
	parseCallCount := 0
	resolver.SetParseFunc(func(source string) (*Module, error) {
		parseCallCount++
		return &Module{}, nil
	})

	// Resolve should return cached module
	loaded, err := resolver.ResolveModule(testFile, ".")
	require.NoError(t, err)
	assert.Equal(t, cachedModule, loaded)
	assert.Equal(t, 0, parseCallCount) // Parser should not have been called
}

func TestModuleResolver_CircularDependency(t *testing.T) {
	resolver := NewModuleResolver()

	// Simulate circular dependency by adding to loading stack
	resolver.loadingStack = []string{"/path/to/module.glyph"}

	formatted := resolver.formatCircularDependency("/path/to/module.glyph")
	assert.Contains(t, formatted, "/path/to/module.glyph")
}

// Test ImportStatement AST

func TestImportStatement_Basic(t *testing.T) {
	stmt := ImportStatement{
		Path:      "./utils",
		Alias:     "",
		Selective: false,
		Names:     nil,
	}

	assert.Equal(t, "./utils", stmt.Path)
	assert.False(t, stmt.Selective)
}

func TestImportStatement_WithAlias(t *testing.T) {
	stmt := ImportStatement{
		Path:      "./utils",
		Alias:     "u",
		Selective: false,
		Names:     nil,
	}

	assert.Equal(t, "u", stmt.Alias)
}

func TestImportStatement_Selective(t *testing.T) {
	stmt := ImportStatement{
		Path:      "./utils",
		Alias:     "",
		Selective: true,
		Names: []ImportName{
			{Name: "funcA", Alias: ""},
			{Name: "funcB", Alias: "fb"},
		},
	}

	assert.True(t, stmt.Selective)
	assert.Len(t, stmt.Names, 2)
	assert.Equal(t, "funcA", stmt.Names[0].Name)
	assert.Equal(t, "fb", stmt.Names[1].Alias)
}

// Test ModuleDecl AST

func TestModuleDecl(t *testing.T) {
	decl := ModuleDecl{Name: "myapp/utils"}

	assert.Equal(t, "myapp/utils", decl.Name)
}

// Test extractModuleName helper

func TestExtractModuleName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"./utils", "utils"},
		{"./path/to/module", "module"},
		{"module", "module"},
		{"./utils.glyph", "utils"},
		{"path/to/module.glyph", "module"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractModuleName(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test Interpreter module loading

func TestInterpreter_GetModuleResolver(t *testing.T) {
	interp := NewInterpreter()

	resolver := interp.GetModuleResolver()
	assert.NotNil(t, resolver)
}

func TestInterpreter_SetModuleResolver(t *testing.T) {
	interp := NewInterpreter()
	customResolver := NewModuleResolver()
	customResolver.AddSearchPath("/custom")

	interp.SetModuleResolver(customResolver)

	resolver := interp.GetModuleResolver()
	assert.Contains(t, resolver.SearchPaths, "/custom")
}

func TestInterpreter_LoadModuleWithPath(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&Function{
				Name:   "testFunc",
				Params: []Field{},
				Body: []Statement{
					ReturnStatement{
						Value: LiteralExpr{Value: IntLiteral{Value: 42}},
					},
				},
			},
		},
	}

	err := interp.LoadModuleWithPath(module, ".")
	require.NoError(t, err)

	fn, exists := interp.GetFunction("testFunc")
	assert.True(t, exists)
	assert.Equal(t, "testFunc", fn.Name)
}

func TestInterpreter_LoadModuleWithModuleDecl(t *testing.T) {
	interp := NewInterpreter()

	module := Module{
		Items: []Item{
			&ModuleDecl{Name: "myapp/utils"},
			&Function{Name: "helper"},
		},
	}

	err := interp.LoadModuleWithPath(module, ".")
	require.NoError(t, err)

	// ModuleDecl should be ignored (just stored for reference)
	fn, exists := interp.GetFunction("helper")
	assert.True(t, exists)
	assert.Equal(t, "helper", fn.Name)
}

// Test interface implementations

func TestImportStatement_IsItem(t *testing.T) {
	stmt := ImportStatement{}
	var _ Item = stmt // Compile-time check that ImportStatement implements Item
}

func TestModuleDecl_IsItem(t *testing.T) {
	decl := ModuleDecl{}
	var _ Item = decl // Compile-time check that ModuleDecl implements Item
}

func TestImportStatement_IsNode(t *testing.T) {
	stmt := ImportStatement{}
	var _ Node = stmt // Compile-time check that ImportStatement implements Node
}

func TestModuleDecl_IsNode(t *testing.T) {
	decl := ModuleDecl{}
	var _ Node = decl // Compile-time check that ModuleDecl implements Node
}
