package jit

import (
	"github.com/glyphlang/glyph/pkg/ast"
	"testing"
	"time"
)

// TestNewJITCompiler tests JIT compiler creation
func TestNewJITCompiler(t *testing.T) {
	jit := NewJITCompiler()

	if jit == nil {
		t.Fatal("NewJITCompiler returned nil")
	}

	if jit.hotPathThreshold != DefaultHotPathThreshold {
		t.Errorf("Expected hot path threshold %d, got %d", DefaultHotPathThreshold, jit.hotPathThreshold)
	}

	if jit.profiler == nil {
		t.Error("Profiler not initialized")
	}

	if jit.units == nil {
		t.Error("Units map not initialized")
	}
}

// TestNewJITCompilerWithConfig tests JIT compiler creation with custom config
func TestNewJITCompilerWithConfig(t *testing.T) {
	threshold := 50
	window := 5 * time.Second

	jit := NewJITCompilerWithConfig(threshold, window)

	if jit.hotPathThreshold != threshold {
		t.Errorf("Expected hot path threshold %d, got %d", threshold, jit.hotPathThreshold)
	}

	if jit.recompileWindow != window {
		t.Errorf("Expected recompile window %v, got %v", window, jit.recompileWindow)
	}
}

// TestCompileRoute tests basic route compilation
func TestCompileRoute(t *testing.T) {
	jit := NewJITCompiler()

	// Create a simple route
	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	bytecode, err := jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("CompileRoute failed: %v", err)
	}

	if len(bytecode) == 0 {
		t.Error("Expected non-empty bytecode")
	}

	// Verify unit was cached
	unit, exists := jit.GetUnit("test_route")
	if !exists {
		t.Error("Expected unit to be cached")
	}

	if unit.Name != "test_route" {
		t.Errorf("Expected unit name 'test_route', got '%s'", unit.Name)
	}

	if unit.Tier != TierBaseline {
		t.Errorf("Expected initial tier to be TierBaseline, got %v", unit.Tier)
	}
}

// TestCacheHit tests that cached bytecode is reused
func TestCacheHit(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// First compilation
	bytecode1, err := jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("First compilation failed: %v", err)
	}

	stats1 := jit.GetStats()

	// Second compilation (should hit cache)
	bytecode2, err := jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("Second compilation failed: %v", err)
	}

	stats2 := jit.GetStats()

	// Verify cache hit
	if stats2.CacheHits != stats1.CacheHits+1 {
		t.Errorf("Expected cache hit count to increase by 1")
	}

	// Bytecode should be identical
	if len(bytecode1) != len(bytecode2) {
		t.Errorf("Cached bytecode length mismatch")
	}
}

// TestRecordExecution tests execution recording
func TestRecordExecution(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// Compile the route first
	_, err := jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Record some executions
	for i := 0; i < 10; i++ {
		jit.RecordExecution("test_route", time.Millisecond)
	}

	// Check unit execution count
	unit, exists := jit.GetUnit("test_route")
	if !exists {
		t.Fatal("Unit not found")
	}

	if unit.ExecutionCount != 10 {
		t.Errorf("Expected execution count 10, got %d", unit.ExecutionCount)
	}

	// Check profiler data
	profile := jit.profiler.GetProfile("test_route")
	if profile == nil {
		t.Fatal("Profile not found")
	}

	if profile.ExecutionCount != 10 {
		t.Errorf("Expected profile execution count 10, got %d", profile.ExecutionCount)
	}
}

// TestHotPathDetection tests hot path detection and recompilation
func TestHotPathDetection(t *testing.T) {
	// Use a low threshold for testing
	jit := NewJITCompilerWithConfig(10, 1*time.Millisecond)

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// Compile the route
	_, err := jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify initial tier
	unit, _ := jit.GetUnit("test_route")
	if unit.Tier != TierBaseline {
		t.Errorf("Expected initial tier TierBaseline, got %v", unit.Tier)
	}

	// Record enough executions to trigger recompilation
	for i := 0; i < 6; i++ {
		jit.RecordExecution("test_route", time.Millisecond)
	}

	// Wait for recompile window
	time.Sleep(2 * time.Millisecond)

	// Trigger recompilation by compiling again
	_, err = jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("Recompilation failed: %v", err)
	}

	// Verify tier was upgraded
	unit, _ = jit.GetUnit("test_route")
	if unit.Tier < TierOptimized {
		t.Errorf("Expected tier to be upgraded to at least TierOptimized, got %v", unit.Tier)
	}
}

// TestGetHotPaths tests retrieving hot paths
func TestGetHotPaths(t *testing.T) {
	jit := NewJITCompiler()

	// Record executions for multiple routes
	jit.RecordExecution("route1", time.Millisecond)
	jit.RecordExecution("route2", time.Millisecond)

	for i := 0; i < 150; i++ {
		jit.RecordExecution("route1", time.Millisecond)
	}

	for i := 0; i < 200; i++ {
		jit.RecordExecution("route3", time.Millisecond)
	}

	// Get hot paths with threshold of 100
	hotPaths := jit.GetHotPaths()

	if len(hotPaths) != 2 {
		t.Errorf("Expected 2 hot paths, got %d", len(hotPaths))
	}

	// Verify they're sorted by execution count (route3 should be first)
	if len(hotPaths) >= 2 && hotPaths[0] != "route3" {
		t.Errorf("Expected route3 to be the hottest path, got %s", hotPaths[0])
	}
}

// TestInvalidateCache tests cache invalidation
func TestInvalidateCache(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// Compile and cache
	_, err := jit.CompileRoute("test_route", route)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify cached
	_, exists := jit.GetUnit("test_route")
	if !exists {
		t.Fatal("Expected unit to be cached")
	}

	// Invalidate
	jit.InvalidateCache("test_route")

	// Verify removed
	_, exists = jit.GetUnit("test_route")
	if exists {
		t.Error("Expected unit to be removed from cache")
	}
}

// TestClearCache tests clearing entire cache
func TestClearCache(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// Compile multiple routes
	jit.CompileRoute("route1", route)
	jit.CompileRoute("route2", route)
	jit.CompileRoute("route3", route)

	// Clear cache
	jit.ClearCache()

	// Verify all removed
	_, exists1 := jit.GetUnit("route1")
	_, exists2 := jit.GetUnit("route2")
	_, exists3 := jit.GetUnit("route3")

	if exists1 || exists2 || exists3 {
		t.Error("Expected all units to be removed from cache")
	}
}

// TestGetStats tests statistics collection
func TestGetStats(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// Perform some compilations
	jit.CompileRoute("route1", route)
	jit.CompileRoute("route2", route)

	stats := jit.GetStats()

	if stats.TotalCompilations != 2 {
		t.Errorf("Expected 2 total compilations, got %d", stats.TotalCompilations)
	}

	if stats.BaselineCompilations != 2 {
		t.Errorf("Expected 2 baseline compilations, got %d", stats.BaselineCompilations)
	}

	// Record executions
	jit.RecordExecution("route1", time.Millisecond)
	jit.RecordExecution("route1", time.Millisecond)

	stats = jit.GetStats()

	if stats.TotalExecutionTime != 2*time.Millisecond {
		t.Errorf("Expected total execution time of 2ms, got %v", stats.TotalExecutionTime)
	}
}

// TestOptimizationTiers tests different optimization tiers
func TestOptimizationTiers(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 1}},
			},
			&ast.AssignStatement{
				Target: "y",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
			},
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op:    ast.Add,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.VariableExpr{Name: "y"},
				},
			},
		},
	}

	// Test baseline compilation
	bytecodeBaseline, err := jit.compileWithTier(route, TierBaseline)
	if err != nil {
		t.Fatalf("Baseline compilation failed: %v", err)
	}

	// Test optimized compilation
	bytecodeOptimized, err := jit.compileWithTier(route, TierOptimized)
	if err != nil {
		t.Fatalf("Optimized compilation failed: %v", err)
	}

	// Test highly optimized compilation
	bytecodeAggressive, err := jit.compileWithTier(route, TierHighlyOptimized)
	if err != nil {
		t.Fatalf("Highly optimized compilation failed: %v", err)
	}

	// All should produce valid bytecode
	if len(bytecodeBaseline) == 0 || len(bytecodeOptimized) == 0 || len(bytecodeAggressive) == 0 {
		t.Error("Expected non-empty bytecode for all tiers")
	}

	// Optimized/aggressive bytecode might be smaller due to optimizations
	// (this is not guaranteed, but generally expected for simple code)
	t.Logf("Baseline size: %d, Optimized size: %d, Aggressive size: %d",
		len(bytecodeBaseline), len(bytecodeOptimized), len(bytecodeAggressive))
}

// TestConcurrentCompilation tests thread safety of JIT compiler
func TestConcurrentCompilation(t *testing.T) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{
					Value: ast.IntLiteral{Value: 42},
				},
			},
		},
	}

	// Spawn multiple goroutines to compile concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			routeName := "route"
			_, err := jit.CompileRoute(routeName, route)
			if err != nil {
				t.Errorf("Concurrent compilation failed: %v", err)
			}
			jit.RecordExecution(routeName, time.Microsecond)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify stats are consistent
	stats := jit.GetStats()
	if stats.TotalCompilations == 0 {
		t.Error("Expected some compilations to have occurred")
	}
}

// TestSettersAndGetters tests configuration setters
func TestSettersAndGetters(t *testing.T) {
	jit := NewJITCompiler()

	// Test SetHotPathThreshold
	jit.SetHotPathThreshold(500)
	if jit.hotPathThreshold != 500 {
		t.Errorf("Expected hot path threshold 500, got %d", jit.hotPathThreshold)
	}

	// Test SetRecompileWindow
	window := 30 * time.Second
	jit.SetRecompileWindow(window)
	if jit.recompileWindow != window {
		t.Errorf("Expected recompile window %v, got %v", window, jit.recompileWindow)
	}

	// Test GetProfiler
	profiler := jit.GetProfiler()
	if profiler == nil {
		t.Error("Expected non-nil profiler")
	}
}

// Benchmark tests

// BenchmarkCompileRoute benchmarks route compilation
func BenchmarkCompileRoute(b *testing.B) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.AssignStatement{
				Target: "x",
				Value:  &ast.LiteralExpr{Value: ast.IntLiteral{Value: 10}},
			},
			&ast.ReturnStatement{
				Value: &ast.BinaryOpExpr{
					Op:    ast.Mul,
					Left:  &ast.VariableExpr{Name: "x"},
					Right: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 2}},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit.InvalidateCache("test_route") // Force recompilation
		_, err := jit.CompileRoute("test_route", route)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkCacheHit benchmarks cached bytecode retrieval
func BenchmarkCacheHit(b *testing.B) {
	jit := NewJITCompiler()

	route := &ast.Route{
		Method: ast.Get,
		Path:   "/test",
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Value: &ast.LiteralExpr{Value: ast.IntLiteral{Value: 42}},
			},
		},
	}

	// Compile once
	jit.CompileRoute("test_route", route)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jit.CompileRoute("test_route", route)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}

// BenchmarkRecordExecution benchmarks execution recording
func BenchmarkRecordExecution(b *testing.B) {
	jit := NewJITCompiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jit.RecordExecution("test_route", time.Microsecond)
	}
}
