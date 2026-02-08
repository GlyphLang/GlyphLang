package compiler

import "testing"

func TestNewGlobalSymbolTable(t *testing.T) {
	st := NewGlobalSymbolTable()
	if st == nil {
		t.Fatal("NewGlobalSymbolTable() returned nil")
	}
	if st.Scope() != GlobalScope {
		t.Errorf("Scope() = %d, want GlobalScope (%d)", st.Scope(), GlobalScope)
	}
	if st.Parent() != nil {
		t.Error("Expected nil parent for global symbol table")
	}
	if len(st.Symbols()) != 0 {
		t.Errorf("Expected empty symbols, got %d", len(st.Symbols()))
	}
}

func TestNewSymbolTable(t *testing.T) {
	parent := NewGlobalSymbolTable()
	child := NewSymbolTable(parent, FunctionScope)

	if child.Scope() != FunctionScope {
		t.Errorf("Scope() = %d, want FunctionScope (%d)", child.Scope(), FunctionScope)
	}
	if child.Parent() != parent {
		t.Error("Parent() should return the parent symbol table")
	}
}

func TestSymbolTable_Define(t *testing.T) {
	st := NewGlobalSymbolTable()

	sym := st.Define("x", 0)
	if sym == nil {
		t.Fatal("Define() returned nil")
	}
	if sym.Name != "x" {
		t.Errorf("Name = %q, want %q", sym.Name, "x")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("Scope = %d, want GlobalScope", sym.Scope)
	}
	if sym.Index != 0 {
		t.Errorf("Index = %d, want 0", sym.Index)
	}
	if !sym.IsDefined {
		t.Error("IsDefined should be true")
	}
	if sym.IsConstant {
		t.Error("IsConstant should be false for non-constant")
	}
}

func TestSymbolTable_DefineConstant(t *testing.T) {
	st := NewGlobalSymbolTable()

	sym := st.DefineConstant("PI", 0, 1)
	if sym == nil {
		t.Fatal("DefineConstant() returned nil")
	}
	if sym.Name != "PI" {
		t.Errorf("Name = %q, want %q", sym.Name, "PI")
	}
	if !sym.IsConstant {
		t.Error("IsConstant should be true")
	}
	if !sym.IsDefined {
		t.Error("IsDefined should be true")
	}
	if sym.ConstantIdx != 1 {
		t.Errorf("ConstantIdx = %d, want 1", sym.ConstantIdx)
	}
	if sym.Index != 0 {
		t.Errorf("Index = %d, want 0", sym.Index)
	}
}

func TestSymbolTable_Resolve_CurrentScope(t *testing.T) {
	st := NewGlobalSymbolTable()
	st.Define("x", 0)

	sym, ok := st.Resolve("x")
	if !ok {
		t.Fatal("Resolve() returned false for defined symbol")
	}
	if sym.Name != "x" {
		t.Errorf("Name = %q, want %q", sym.Name, "x")
	}
}

func TestSymbolTable_Resolve_ParentScope(t *testing.T) {
	parent := NewGlobalSymbolTable()
	parent.Define("globalVar", 0)

	child := parent.EnterScope(FunctionScope)

	sym, ok := child.Resolve("globalVar")
	if !ok {
		t.Fatal("Resolve() should find symbol in parent scope")
	}
	if sym.Name != "globalVar" {
		t.Errorf("Name = %q, want %q", sym.Name, "globalVar")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("Scope = %d, want GlobalScope", sym.Scope)
	}
}

func TestSymbolTable_Resolve_NotFound(t *testing.T) {
	st := NewGlobalSymbolTable()

	sym, ok := st.Resolve("undefined")
	if ok {
		t.Error("Resolve() should return false for undefined symbol")
	}
	if sym != nil {
		t.Errorf("Expected nil symbol, got %v", sym)
	}
}

func TestSymbolTable_Resolve_NestedScopes(t *testing.T) {
	global := NewGlobalSymbolTable()
	global.Define("a", 0)

	route := global.EnterScope(RouteScope)
	route.Define("b", 1)

	fn := route.EnterScope(FunctionScope)
	fn.Define("c", 2)

	block := fn.EnterScope(BlockScope)
	block.Define("d", 3)

	// Block should see all symbols
	for _, name := range []string{"a", "b", "c", "d"} {
		if _, ok := block.Resolve(name); !ok {
			t.Errorf("Block scope should resolve %q", name)
		}
	}

	// Function should see a, b, c but not d
	for _, name := range []string{"a", "b", "c"} {
		if _, ok := fn.Resolve(name); !ok {
			t.Errorf("Function scope should resolve %q", name)
		}
	}
	if _, ok := fn.Resolve("d"); ok {
		t.Error("Function scope should not resolve 'd' from child block scope")
	}

	// Global should only see a
	if _, ok := global.Resolve("a"); !ok {
		t.Error("Global scope should resolve 'a'")
	}
	if _, ok := global.Resolve("b"); ok {
		t.Error("Global scope should not resolve 'b' from child scope")
	}
}

func TestSymbolTable_ResolveLocal(t *testing.T) {
	parent := NewGlobalSymbolTable()
	parent.Define("globalVar", 0)

	child := parent.EnterScope(FunctionScope)
	child.Define("localVar", 1)

	// ResolveLocal should find localVar in child
	sym, ok := child.ResolveLocal("localVar")
	if !ok {
		t.Fatal("ResolveLocal() should find symbol in current scope")
	}
	if sym.Name != "localVar" {
		t.Errorf("Name = %q, want %q", sym.Name, "localVar")
	}

	// ResolveLocal should NOT find globalVar in child (only current scope)
	_, ok = child.ResolveLocal("globalVar")
	if ok {
		t.Error("ResolveLocal() should not find symbols in parent scope")
	}
}

func TestSymbolTable_EnterScope(t *testing.T) {
	global := NewGlobalSymbolTable()

	route := global.EnterScope(RouteScope)
	if route.Scope() != RouteScope {
		t.Errorf("Scope() = %d, want RouteScope", route.Scope())
	}
	if route.Parent() != global {
		t.Error("Parent() should be the global table")
	}

	fn := route.EnterScope(FunctionScope)
	if fn.Scope() != FunctionScope {
		t.Errorf("Scope() = %d, want FunctionScope", fn.Scope())
	}
	if fn.Parent() != route {
		t.Error("Parent() should be the route table")
	}

	block := fn.EnterScope(BlockScope)
	if block.Scope() != BlockScope {
		t.Errorf("Scope() = %d, want BlockScope", block.Scope())
	}
	if block.Parent() != fn {
		t.Error("Parent() should be the function table")
	}
}

func TestSymbolTable_Shadowing(t *testing.T) {
	global := NewGlobalSymbolTable()
	global.Define("x", 0)

	child := global.EnterScope(FunctionScope)
	child.Define("x", 1) // Shadow global x

	// Child should resolve to its own x
	sym, ok := child.Resolve("x")
	if !ok {
		t.Fatal("Resolve() should find shadowed symbol")
	}
	if sym.Scope != FunctionScope {
		t.Errorf("Expected FunctionScope for shadowed symbol, got %d", sym.Scope)
	}
	if sym.Index != 1 {
		t.Errorf("Expected Index 1 for shadowed symbol, got %d", sym.Index)
	}

	// Parent should resolve to its own x
	sym, ok = global.Resolve("x")
	if !ok {
		t.Fatal("Parent should still resolve its own x")
	}
	if sym.Scope != GlobalScope {
		t.Errorf("Expected GlobalScope for parent symbol, got %d", sym.Scope)
	}
	if sym.Index != 0 {
		t.Errorf("Expected Index 0 for parent symbol, got %d", sym.Index)
	}
}

func TestSymbolTable_Symbols(t *testing.T) {
	st := NewGlobalSymbolTable()
	st.Define("a", 0)
	st.Define("b", 1)
	st.Define("c", 2)

	symbols := st.Symbols()
	if len(symbols) != 3 {
		t.Errorf("Expected 3 symbols, got %d", len(symbols))
	}

	for _, name := range []string{"a", "b", "c"} {
		if _, ok := symbols[name]; !ok {
			t.Errorf("Missing symbol %q", name)
		}
	}
}

func TestSymbolTable_Redefine(t *testing.T) {
	st := NewGlobalSymbolTable()
	st.Define("x", 0)
	st.Define("x", 5) // Redefine

	sym, ok := st.Resolve("x")
	if !ok {
		t.Fatal("Resolve() should find redefined symbol")
	}
	if sym.Index != 5 {
		t.Errorf("Expected Index 5 after redefine, got %d", sym.Index)
	}
}
