package jit

import (
	"testing"
	"time"
)

// TestNewProfiler tests profiler creation
func TestNewProfiler(t *testing.T) {
	p := NewProfiler()

	if p == nil {
		t.Fatal("NewProfiler returned nil")
	}

	if p.profiles == nil {
		t.Error("Profiles map not initialized")
	}

	if p.callGraph == nil {
		t.Error("Call graph not initialized")
	}

	if p.maxProfiles <= 0 {
		t.Error("Max profiles should be positive")
	}
}

// TestRecordExecution tests execution recording
func TestRecordExecutionProfiler(t *testing.T) {
	p := NewProfiler()

	// Record first execution
	p.RecordExecution("test_route", 10*time.Millisecond)

	profile := p.GetProfile("test_route")
	if profile == nil {
		t.Fatal("Profile not created")
	}

	if profile.ExecutionCount != 1 {
		t.Errorf("Expected execution count 1, got %d", profile.ExecutionCount)
	}

	if profile.TotalTime != 10*time.Millisecond {
		t.Errorf("Expected total time 10ms, got %v", profile.TotalTime)
	}

	if profile.AverageTime != 10*time.Millisecond {
		t.Errorf("Expected average time 10ms, got %v", profile.AverageTime)
	}

	// Record second execution
	p.RecordExecution("test_route", 20*time.Millisecond)

	profile = p.GetProfile("test_route")
	if profile.ExecutionCount != 2 {
		t.Errorf("Expected execution count 2, got %d", profile.ExecutionCount)
	}

	if profile.TotalTime != 30*time.Millisecond {
		t.Errorf("Expected total time 30ms, got %v", profile.TotalTime)
	}

	if profile.AverageTime != 15*time.Millisecond {
		t.Errorf("Expected average time 15ms, got %v", profile.AverageTime)
	}
}

// TestMinMaxTime tests min/max execution time tracking
func TestMinMaxTime(t *testing.T) {
	p := NewProfiler()

	p.RecordExecution("test_route", 10*time.Millisecond)
	p.RecordExecution("test_route", 5*time.Millisecond)
	p.RecordExecution("test_route", 15*time.Millisecond)

	profile := p.GetProfile("test_route")

	if profile.MinTime != 5*time.Millisecond {
		t.Errorf("Expected min time 5ms, got %v", profile.MinTime)
	}

	if profile.MaxTime != 15*time.Millisecond {
		t.Errorf("Expected max time 15ms, got %v", profile.MaxTime)
	}
}

// TestRecordTypeUsage tests type usage recording
func TestRecordTypeUsage(t *testing.T) {
	p := NewProfiler()

	p.RecordTypeUsage("test_route", "x", "int")
	p.RecordTypeUsage("test_route", "x", "int")
	p.RecordTypeUsage("test_route", "x", "string")
	p.RecordTypeUsage("test_route", "y", "float")

	profile := p.GetProfile("test_route")
	if profile == nil {
		t.Fatal("Profile not created")
	}

	// Check type counts for variable x
	xTypes := profile.TypeProfile.VariableTypes["x"]
	if xTypes["int"] != 2 {
		t.Errorf("Expected int type count 2 for x, got %d", xTypes["int"])
	}

	if xTypes["string"] != 1 {
		t.Errorf("Expected string type count 1 for x, got %d", xTypes["string"])
	}

	// Check type counts for variable y
	yTypes := profile.TypeProfile.VariableTypes["y"]
	if yTypes["float"] != 1 {
		t.Errorf("Expected float type count 1 for y, got %d", yTypes["float"])
	}
}

// TestRecordReturnType tests return type recording
func TestRecordReturnType(t *testing.T) {
	p := NewProfiler()

	p.RecordReturnType("test_route", "int")
	p.RecordReturnType("test_route", "int")
	p.RecordReturnType("test_route", "string")

	profile := p.GetProfile("test_route")
	if profile == nil {
		t.Fatal("Profile not created")
	}

	if profile.TypeProfile.ReturnTypes["int"] != 2 {
		t.Errorf("Expected int return type count 2, got %d", profile.TypeProfile.ReturnTypes["int"])
	}

	if profile.TypeProfile.ReturnTypes["string"] != 1 {
		t.Errorf("Expected string return type count 1, got %d", profile.TypeProfile.ReturnTypes["string"])
	}
}

// TestRecordCall tests call recording
func TestRecordCall(t *testing.T) {
	p := NewProfiler()

	p.RecordCall("route_a", "route_b")
	p.RecordCall("route_a", "route_b")
	p.RecordCall("route_a", "route_c")

	// Check caller profile
	profileA := p.GetProfile("route_a")
	if profileA.Calls["route_b"] != 2 {
		t.Errorf("Expected route_a to call route_b 2 times, got %d", profileA.Calls["route_b"])
	}

	if profileA.Calls["route_c"] != 1 {
		t.Errorf("Expected route_a to call route_c 1 time, got %d", profileA.Calls["route_c"])
	}

	// Check callee profile
	profileB := p.GetProfile("route_b")
	if profileB.CalledBy["route_a"] != 2 {
		t.Errorf("Expected route_b to be called by route_a 2 times, got %d", profileB.CalledBy["route_a"])
	}
}

// TestGetHotPathsProfiler tests hot path retrieval
func TestGetHotPathsProfiler(t *testing.T) {
	p := NewProfiler()

	// Record executions
	for i := 0; i < 150; i++ {
		p.RecordExecution("hot_route", time.Millisecond)
	}

	for i := 0; i < 50; i++ {
		p.RecordExecution("warm_route", time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		p.RecordExecution("cold_route", time.Millisecond)
	}

	// Get hot paths with threshold of 100
	hotPaths := p.GetHotPaths(100)

	if len(hotPaths) != 1 {
		t.Errorf("Expected 1 hot path, got %d", len(hotPaths))
	}

	if len(hotPaths) > 0 && hotPaths[0] != "hot_route" {
		t.Errorf("Expected hot_route to be hot, got %s", hotPaths[0])
	}

	// Get hot paths with lower threshold
	hotPaths = p.GetHotPaths(40)

	if len(hotPaths) != 2 {
		t.Errorf("Expected 2 hot paths with threshold 40, got %d", len(hotPaths))
	}
}

// TestGetTopNByExecutionCount tests top N retrieval by execution count
func TestGetTopNByExecutionCount(t *testing.T) {
	p := NewProfiler()

	p.RecordExecution("route1", time.Millisecond)
	for i := 0; i < 5; i++ {
		p.RecordExecution("route2", time.Millisecond)
	}
	for i := 0; i < 10; i++ {
		p.RecordExecution("route3", time.Millisecond)
	}

	topN := p.GetTopNByExecutionCount(2)

	if len(topN) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(topN))
	}

	// Should be sorted by execution count
	if topN[0].Name != "route3" {
		t.Errorf("Expected route3 to be first, got %s", topN[0].Name)
	}

	if topN[1].Name != "route2" {
		t.Errorf("Expected route2 to be second, got %s", topN[1].Name)
	}
}

// TestGetTopNByTotalTime tests top N retrieval by total time
func TestGetTopNByTotalTime(t *testing.T) {
	p := NewProfiler()

	p.RecordExecution("route1", 1*time.Millisecond)
	p.RecordExecution("route2", 10*time.Millisecond)
	p.RecordExecution("route3", 5*time.Millisecond)

	topN := p.GetTopNByTotalTime(2)

	if len(topN) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(topN))
	}

	// Should be sorted by total time
	if topN[0].Name != "route2" {
		t.Errorf("Expected route2 to be first, got %s", topN[0].Name)
	}

	if topN[1].Name != "route3" {
		t.Errorf("Expected route3 to be second, got %s", topN[1].Name)
	}
}

// TestAnalyzeTypeStability tests type stability analysis
func TestAnalyzeTypeStability(t *testing.T) {
	p := NewProfiler()

	// Monomorphic variable (stable)
	for i := 0; i < 100; i++ {
		p.RecordTypeUsage("route1", "x", "int")
	}

	isStable, dominantType := p.AnalyzeTypeStability("route1", "x")
	if !isStable {
		t.Error("Expected x to be type-stable")
	}

	if dominantType != "int" {
		t.Errorf("Expected dominant type to be int, got %s", dominantType)
	}

	// Polymorphic variable (not stable)
	for i := 0; i < 50; i++ {
		p.RecordTypeUsage("route2", "y", "int")
		p.RecordTypeUsage("route2", "y", "string")
	}

	isStable, _ = p.AnalyzeTypeStability("route2", "y")
	if isStable {
		t.Error("Expected y to be polymorphic (not stable)")
	}
}

// TestGetMonomorphicVariables tests monomorphic variable detection
func TestGetMonomorphicVariables(t *testing.T) {
	p := NewProfiler()

	// Record stable types
	for i := 0; i < 100; i++ {
		p.RecordTypeUsage("test_route", "x", "int")
		p.RecordTypeUsage("test_route", "y", "string")
	}

	// Record unstable type
	for i := 0; i < 50; i++ {
		p.RecordTypeUsage("test_route", "z", "int")
		p.RecordTypeUsage("test_route", "z", "float")
	}

	monomorphic := p.GetMonomorphicVariables("test_route")

	if len(monomorphic) != 2 {
		t.Errorf("Expected 2 monomorphic variables, got %d", len(monomorphic))
	}

	if monomorphic["x"] != "int" {
		t.Errorf("Expected x to be monomorphic int, got %s", monomorphic["x"])
	}

	if monomorphic["y"] != "string" {
		t.Errorf("Expected y to be monomorphic string, got %s", monomorphic["y"])
	}

	if _, exists := monomorphic["z"]; exists {
		t.Error("Expected z to not be monomorphic")
	}
}

// TestMarkHotPaths tests hot path marking in call graph
func TestMarkHotPaths(t *testing.T) {
	p := NewProfiler()

	// Create some routes with different execution counts
	for i := 0; i < 150; i++ {
		p.RecordExecution("hot_route", time.Millisecond)
	}

	for i := 0; i < 50; i++ {
		p.RecordExecution("cold_route", time.Millisecond)
	}

	// Build call graph
	p.RecordCall("main", "hot_route")
	p.RecordCall("main", "cold_route")

	// Mark hot paths with threshold 100
	p.MarkHotPaths(100)

	callGraph := p.GetCallGraph()

	if !callGraph["hot_route"].IsHot {
		t.Error("Expected hot_route to be marked as hot")
	}

	if callGraph["cold_route"].IsHot {
		t.Error("Expected cold_route to not be marked as hot")
	}
}

// TestGetInlineCandidates tests inline candidate detection
func TestGetInlineCandidates(t *testing.T) {
	p := NewProfiler()

	// Record frequent calls from limited callers (good inline candidate)
	for i := 0; i < 100; i++ {
		p.RecordCall("caller1", "candidate1")
	}

	// Record frequent calls from many callers (bad inline candidate - code bloat)
	for i := 0; i < 100; i++ {
		p.RecordCall("caller1", "candidate2")
		p.RecordCall("caller2", "candidate2")
		p.RecordCall("caller3", "candidate2")
		p.RecordCall("caller4", "candidate2")
	}

	// Record infrequent calls (not a good candidate)
	p.RecordCall("caller1", "candidate3")
	p.RecordCall("caller1", "candidate3")

	candidates := p.GetInlineCandidates(50)

	// candidate1 should be a good candidate (frequent calls, few callers)
	found := false
	for _, name := range candidates {
		if name == "candidate1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected candidate1 to be an inline candidate")
	}
}

// TestReset tests profiler reset
func TestReset(t *testing.T) {
	p := NewProfiler()

	// Add some data
	p.RecordExecution("route1", time.Millisecond)
	p.RecordExecution("route2", time.Millisecond)
	p.RecordCall("route1", "route2")

	// Reset
	p.Reset()

	// Verify everything is cleared
	if len(p.GetAllProfiles()) != 0 {
		t.Error("Expected profiles to be cleared")
	}

	if len(p.GetCallGraph()) != 0 {
		t.Error("Expected call graph to be cleared")
	}
}

// TestGetStats tests profiler statistics
func TestGetStatsProfiler(t *testing.T) {
	p := NewProfiler()

	// Record some executions
	p.RecordExecution("route1", 10*time.Millisecond)
	p.RecordExecution("route1", 20*time.Millisecond)
	p.RecordExecution("route2", 5*time.Millisecond)

	stats := p.GetStats()

	if stats.TotalProfiles != 2 {
		t.Errorf("Expected 2 profiles, got %d", stats.TotalProfiles)
	}

	if stats.TotalExecutions != 3 {
		t.Errorf("Expected 3 total executions, got %d", stats.TotalExecutions)
	}

	if stats.TotalExecutionTime != 35*time.Millisecond {
		t.Errorf("Expected total execution time 35ms, got %v", stats.TotalExecutionTime)
	}
}

// TestConcurrentProfiling tests thread safety
func TestConcurrentProfiling(t *testing.T) {
	p := NewProfiler()

	done := make(chan bool)

	// Spawn multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				p.RecordExecution("test_route", time.Microsecond)
				p.RecordTypeUsage("test_route", "x", "int")
				p.RecordCall("caller", "callee")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify data consistency
	profile := p.GetProfile("test_route")
	if profile.ExecutionCount != 1000 {
		t.Errorf("Expected 1000 executions, got %d", profile.ExecutionCount)
	}
}

// TestMaxProfilesLimit tests LRU eviction
func TestMaxProfilesLimit(t *testing.T) {
	p := NewProfiler()
	p.maxProfiles = 5 // Set a low limit for testing

	// Create more profiles than the limit
	for i := 0; i < 10; i++ {
		routeName := "route" + string(rune('0'+i))
		p.RecordExecution(routeName, time.Millisecond)
		time.Sleep(time.Microsecond) // Ensure different timestamps
	}

	// Should only have maxProfiles entries
	if len(p.GetAllProfiles()) > p.maxProfiles {
		t.Errorf("Expected at most %d profiles, got %d", p.maxProfiles, len(p.GetAllProfiles()))
	}
}

// Benchmark tests

// BenchmarkRecordExecution benchmarks execution recording
func BenchmarkRecordExecutionProfiler(b *testing.B) {
	p := NewProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.RecordExecution("test_route", time.Microsecond)
	}
}

// BenchmarkRecordTypeUsage benchmarks type usage recording
func BenchmarkRecordTypeUsage(b *testing.B) {
	p := NewProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.RecordTypeUsage("test_route", "x", "int")
	}
}

// BenchmarkRecordCall benchmarks call recording
func BenchmarkRecordCall(b *testing.B) {
	p := NewProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.RecordCall("caller", "callee")
	}
}

// BenchmarkGetProfile benchmarks profile retrieval
func BenchmarkGetProfile(b *testing.B) {
	p := NewProfiler()
	p.RecordExecution("test_route", time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.GetProfile("test_route")
	}
}
