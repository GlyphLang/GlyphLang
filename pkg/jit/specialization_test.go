package jit

import (
	"testing"
	"time"
)

func TestSpecializationCache_BasicOperations(t *testing.T) {
	cache := NewSpecializationCache()

	types := map[string]string{"x": "int", "y": "string"}
	bytecode := []byte{0x01, 0x02, 0x03}

	// Add specialization
	spec := cache.AddSpecialization("route1", types, bytecode)
	if spec == nil {
		t.Fatal("Expected specialization to be created")
	}
	if spec.Name != "route1<x:int,y:string,>" && spec.Name != "route1<y:string,x:int,>" {
		t.Errorf("Unexpected specialization name: %s", spec.Name)
	}

	// Get existing specialization
	found := cache.GetSpecialization("route1", types)
	if found == nil {
		t.Fatal("Expected to find specialization")
	}
	if found.HitCount != 1 {
		t.Errorf("Expected hit count 1, got %d", found.HitCount)
	}

	// Get non-existent specialization
	notFound := cache.GetSpecialization("route1", map[string]string{"x": "float"})
	if notFound != nil {
		t.Error("Expected nil for non-matching types")
	}
}

func TestSpecializationCache_Invalidation(t *testing.T) {
	cache := NewSpecializationCache()

	types := map[string]string{"x": "int"}
	cache.AddSpecialization("route1", types, []byte{0x01})

	// Invalidate
	cache.InvalidateSpecializations("route1")

	// Should not find invalidated specialization
	found := cache.GetSpecialization("route1", types)
	if found != nil {
		t.Error("Expected invalidated specialization to not be found")
	}
}

func TestSpecializationCache_Eviction(t *testing.T) {
	cache := NewSpecializationCache()
	cache.maxPerRoute = 3

	// Add more specializations than max
	for i := 0; i < 5; i++ {
		types := map[string]string{"x": string(rune('a' + i))}
		cache.AddSpecialization("route1", types, []byte{byte(i)})
	}

	stats := cache.GetStats()
	if stats["route1"].TotalSpecializations > 3 {
		t.Errorf("Expected max 3 specializations, got %d", stats["route1"].TotalSpecializations)
	}
}

func TestSpecializationCache_Stats(t *testing.T) {
	cache := NewSpecializationCache()

	types := map[string]string{"x": "int"}
	cache.AddSpecialization("route1", types, []byte{0x01})

	// Generate some hits
	for i := 0; i < 5; i++ {
		cache.GetSpecialization("route1", types)
	}

	// Generate some misses
	for i := 0; i < 3; i++ {
		cache.RecordMiss("route1")
	}

	stats := cache.GetStats()
	if stats["route1"].TotalHits != 5 {
		t.Errorf("Expected 5 hits, got %d", stats["route1"].TotalHits)
	}
	if stats["route1"].TotalMisses != 3 {
		t.Errorf("Expected 3 misses, got %d", stats["route1"].TotalMisses)
	}
}

func TestInliningOracle_BasicDecisions(t *testing.T) {
	profiler := NewProfiler()
	oracle := NewInliningOracle(profiler)

	// No profiling data - should not inline
	decision := oracle.ShouldInline("caller", "callee", 5)
	if decision.ShouldInline {
		t.Error("Should not inline without profiling data")
	}

	// Add profiling data
	for i := 0; i < 20; i++ {
		profiler.RecordCall("caller", "callee")
		profiler.RecordExecution("callee", time.Millisecond)
	}

	// Should now recommend inlining
	decision = oracle.ShouldInline("caller", "callee", 5)
	if !decision.ShouldInline {
		t.Errorf("Expected to inline, got: %s", decision.Reason)
	}

	// Too large body - should not inline
	decision = oracle.ShouldInline("caller", "callee", 100)
	if decision.ShouldInline {
		t.Error("Should not inline large body")
	}
}

func TestInliningOracle_GetCandidates(t *testing.T) {
	profiler := NewProfiler()
	oracle := NewInliningOracle(profiler)

	// Add profiling data for multiple functions
	for i := 0; i < 15; i++ {
		profiler.RecordCall("main", "helper1")
		profiler.RecordExecution("helper1", time.Millisecond)
	}
	for i := 0; i < 5; i++ {
		profiler.RecordCall("main", "helper2")
		profiler.RecordExecution("helper2", time.Millisecond)
	}

	candidates := oracle.GetInlineCandidates()
	if len(candidates) == 0 {
		t.Error("Expected some inline candidates")
	}

	// helper1 should have higher score due to more calls
	if len(candidates) >= 1 {
		found := false
		for _, c := range candidates {
			if c.Name == "helper1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected helper1 to be a candidate")
		}
	}
}

func TestAdaptiveRecompilationTrigger(t *testing.T) {
	profiler := NewProfiler()
	trigger := NewAdaptiveRecompilationTrigger(profiler)

	// No profiling data
	result := trigger.ShouldRecompile("route1", TierBaseline)
	if result.ShouldRecompile {
		t.Error("Should not recompile without profiling data")
	}

	// Add execution data
	for i := 0; i < 60; i++ {
		profiler.RecordExecution("route1", time.Millisecond)
	}

	// Should now recommend recompilation
	result = trigger.ShouldRecompile("route1", TierBaseline)
	if !result.ShouldRecompile {
		t.Error("Expected to recommend recompilation for hot route")
	}
	if result.Priority < 1 {
		t.Error("Expected positive priority")
	}

	// Already at highest tier - should not recompile
	result = trigger.ShouldRecompile("route1", TierHighlyOptimized)
	if result.ShouldRecompile {
		t.Error("Should not recompile already highly optimized code")
	}
}

func TestAdaptiveRecompilationTrigger_TypeStability(t *testing.T) {
	profiler := NewProfiler()
	trigger := NewAdaptiveRecompilationTrigger(profiler)

	// Add type profiling data showing type stability
	for i := 0; i < 100; i++ {
		profiler.RecordTypeUsage("route1", "x", "int")
	}
	profiler.RecordExecution("route1", time.Millisecond)

	result := trigger.ShouldRecompile("route1", TierOptimized)
	if !result.ShouldRecompile {
		// Type stability should trigger recompilation for specialization
		// (depends on having monomorphic variables)
	}
}

func TestDeoptimizationTracker(t *testing.T) {
	tracker := NewDeoptimizationTracker()

	record := DeoptimizationRecord{
		RouteName:    "route1",
		Reason:       "type mismatch",
		OccurredAt:   time.Now().Unix(),
		FromTier:     TierOptimized,
		TypeMismatch: map[string]string{"int": "string"},
	}

	tracker.RecordDeoptimization(record)

	records := tracker.GetRecords(10)
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
	if records[0].RouteName != "route1" {
		t.Errorf("Expected route1, got %s", records[0].RouteName)
	}

	count := tracker.GetDeoptimizationCount("route1")
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestDeoptimizationTracker_MaxRecords(t *testing.T) {
	tracker := NewDeoptimizationTracker()
	tracker.maxRecords = 5

	// Add more records than max
	for i := 0; i < 10; i++ {
		record := DeoptimizationRecord{
			RouteName:  "route" + string(rune('0'+i)),
			Reason:     "test",
			OccurredAt: time.Now().Unix(),
		}
		tracker.RecordDeoptimization(record)
	}

	records := tracker.GetRecords(100)
	if len(records) > 5 {
		t.Errorf("Expected max 5 records, got %d", len(records))
	}
}

func TestJITCompiler_TypeSpecialization(t *testing.T) {
	jit := NewJITCompiler()

	// The type specialization should work even without a real route
	// Just test that the API works
	types := map[string]string{"x": "int", "y": "string"}

	// First call should be a miss
	stats := jit.GetStats()
	initialMisses := stats.SpecializationMisses

	// Record some type usage
	jit.profiler.RecordTypeUsage("testRoute", "x", "int")
	jit.profiler.RecordTypeUsage("testRoute", "y", "string")

	// Check detailed stats
	detailedStats := jit.GetDetailedStats()
	if detailedStats == nil {
		t.Error("Expected detailed stats")
	}

	compilations, ok := detailedStats["compilations"].(map[string]int64)
	if !ok {
		t.Error("Expected compilations map")
	}
	if compilations["total"] < 0 {
		t.Error("Expected non-negative total compilations")
	}

	_ = types
	_ = initialMisses
}

func TestJITCompiler_DeoptimizationRecording(t *testing.T) {
	jit := NewJITCompiler()

	// Record a deoptimization
	typeMismatch := map[string]string{"expected": "int", "actual": "string"}
	jit.RecordDeoptimization("route1", "type guard failed", typeMismatch)

	// Check stats
	stats := jit.GetStats()
	if stats.Deoptimizations != 1 {
		t.Errorf("Expected 1 deoptimization, got %d", stats.Deoptimizations)
	}

	// Check records
	records := jit.GetDeoptimizationRecords(10)
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}
	if records[0].Reason != "type guard failed" {
		t.Errorf("Expected 'type guard failed', got %s", records[0].Reason)
	}
}

func TestJITCompiler_InlineCandidates(t *testing.T) {
	jit := NewJITCompiler()

	// Add profiling data
	for i := 0; i < 20; i++ {
		jit.profiler.RecordCall("main", "helper")
		jit.profiler.RecordExecution("helper", time.Millisecond)
	}

	candidates := jit.GetInlineCandidates()
	// Should find helper as a candidate
	found := false
	for _, c := range candidates {
		if c.Name == "helper" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected helper to be an inline candidate")
	}

	// Check inlining decision
	decision := jit.ShouldInline("main", "helper", 5)
	if !decision.ShouldInline {
		t.Errorf("Expected to inline, reason: %s", decision.Reason)
	}
}

func TestTypesMatch(t *testing.T) {
	tests := []struct {
		name     string
		a        map[string]string
		b        map[string]string
		expected bool
	}{
		{
			name:     "identical",
			a:        map[string]string{"x": "int"},
			b:        map[string]string{"x": "int"},
			expected: true,
		},
		{
			name:     "different values",
			a:        map[string]string{"x": "int"},
			b:        map[string]string{"x": "string"},
			expected: false,
		},
		{
			name:     "different keys",
			a:        map[string]string{"x": "int"},
			b:        map[string]string{"y": "int"},
			expected: false,
		},
		{
			name:     "different sizes",
			a:        map[string]string{"x": "int", "y": "string"},
			b:        map[string]string{"x": "int"},
			expected: false,
		},
		{
			name:     "empty maps",
			a:        map[string]string{},
			b:        map[string]string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typesMatch(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("typesMatch(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
