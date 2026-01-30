package jit

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/glyphlang/glyph/pkg/compiler"
	"github.com/glyphlang/glyph/pkg/interpreter"
)

// TypeSpecialization represents specialized code for specific type combinations
type TypeSpecialization struct {
	Name      string
	Types     map[string]string // variable -> type
	Bytecode  []byte
	HitCount  int64
	MissCount int64
	IsValid   bool
}

// SpecializationCache caches type-specialized code
type SpecializationCache struct {
	specializations map[string][]*TypeSpecialization // route -> list of specializations
	maxPerRoute     int
	mutex           sync.RWMutex
}

// NewSpecializationCache creates a new specialization cache
func NewSpecializationCache() *SpecializationCache {
	return &SpecializationCache{
		specializations: make(map[string][]*TypeSpecialization),
		maxPerRoute:     5, // Maximum specializations per route
	}
}

// GetSpecialization finds a matching specialization for the given types
func (sc *SpecializationCache) GetSpecialization(routeName string, types map[string]string) *TypeSpecialization {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	specs, ok := sc.specializations[routeName]
	if !ok {
		return nil
	}

	for _, spec := range specs {
		if spec.IsValid && typesMatch(spec.Types, types) {
			atomic.AddInt64(&spec.HitCount, 1)
			return spec
		}
	}

	return nil
}

// AddSpecialization adds a new specialization
func (sc *SpecializationCache) AddSpecialization(routeName string, types map[string]string, bytecode []byte) *TypeSpecialization {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	spec := &TypeSpecialization{
		Name:     fmt.Sprintf("%s<%s>", routeName, typeSignature(types)),
		Types:    types,
		Bytecode: bytecode,
		HitCount: 0,
		IsValid:  true,
	}

	if _, ok := sc.specializations[routeName]; !ok {
		sc.specializations[routeName] = make([]*TypeSpecialization, 0)
	}

	// Check if we've hit the limit
	if len(sc.specializations[routeName]) >= sc.maxPerRoute {
		// Evict the least-used specialization
		sc.evictLeastUsed(routeName)
	}

	sc.specializations[routeName] = append(sc.specializations[routeName], spec)
	return spec
}

// InvalidateSpecializations invalidates all specializations for a route
func (sc *SpecializationCache) InvalidateSpecializations(routeName string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if specs, ok := sc.specializations[routeName]; ok {
		for _, spec := range specs {
			spec.IsValid = false
		}
	}
}

// RecordMiss records a specialization miss
func (sc *SpecializationCache) RecordMiss(routeName string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if specs, ok := sc.specializations[routeName]; ok {
		for _, spec := range specs {
			spec.MissCount++
		}
	}
}

// GetStats returns specialization statistics
func (sc *SpecializationCache) GetStats() map[string]SpecializationStats {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	stats := make(map[string]SpecializationStats)
	for routeName, specs := range sc.specializations {
		var totalHits, totalMisses int64
		validCount := 0
		for _, spec := range specs {
			totalHits += spec.HitCount
			totalMisses += spec.MissCount
			if spec.IsValid {
				validCount++
			}
		}
		stats[routeName] = SpecializationStats{
			TotalSpecializations: len(specs),
			ValidSpecializations: validCount,
			TotalHits:            totalHits,
			TotalMisses:          totalMisses,
		}
	}
	return stats
}

func (sc *SpecializationCache) evictLeastUsed(routeName string) {
	specs := sc.specializations[routeName]
	if len(specs) == 0 {
		return
	}

	minIdx := 0
	minHits := specs[0].HitCount

	for i, spec := range specs {
		if spec.HitCount < minHits {
			minIdx = i
			minHits = spec.HitCount
		}
	}

	// Remove the least-used specialization
	sc.specializations[routeName] = append(specs[:minIdx], specs[minIdx+1:]...)
}

// SpecializationStats contains statistics for a route's specializations
type SpecializationStats struct {
	TotalSpecializations int
	ValidSpecializations int
	TotalHits            int64
	TotalMisses          int64
}

// typesMatch checks if two type maps match
func typesMatch(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// typeSignature creates a string signature for a type map
func typeSignature(types map[string]string) string {
	// Create a deterministic signature
	sig := ""
	for k, v := range types {
		sig += k + ":" + v + ","
	}
	return sig
}

// TypeSpecializedCompiler generates type-specialized bytecode
type TypeSpecializedCompiler struct {
	profiler *Profiler
}

// NewTypeSpecializedCompiler creates a new type-specialized compiler
func NewTypeSpecializedCompiler(profiler *Profiler) *TypeSpecializedCompiler {
	return &TypeSpecializedCompiler{
		profiler: profiler,
	}
}

// CompileWithTypeInfo compiles a route with type specialization hints
func (tsc *TypeSpecializedCompiler) CompileWithTypeInfo(route *interpreter.Route, types map[string]string) ([]byte, error) {
	// For now, compile with aggressive optimizations
	// In a full implementation, we would:
	// 1. Insert type guards at entry
	// 2. Specialize operations based on known types
	// 3. Inline type checks that we know will succeed

	// Create a new compiler instance for each compilation to avoid race conditions
	comp := compiler.NewCompilerWithOptLevel(compiler.OptAggressive)
	return comp.CompileRoute(route)
}

// InlineCandidate represents a function that could be inlined
type InlineCandidate struct {
	Name        string
	CallCount   int64
	CallerCount int
	BodySize    int
	Score       float64 // Higher = better candidate for inlining
}

// InliningDecision contains the decision about whether to inline a function
type InliningDecision struct {
	ShouldInline bool
	Reason       string
	Benefit      float64
}

// InliningOracle decides whether functions should be inlined
type InliningOracle struct {
	profiler        *Profiler
	maxInlineSize   int
	minCallCount    int64
	inlineThreshold float64
}

// NewInliningOracle creates a new inlining oracle
func NewInliningOracle(profiler *Profiler) *InliningOracle {
	return &InliningOracle{
		profiler:        profiler,
		maxInlineSize:   20,  // Max statements to inline
		minCallCount:    10,  // Min calls before considering inlining
		inlineThreshold: 0.5, // Score threshold for inlining
	}
}

// ShouldInline decides if a function should be inlined at a call site
func (io *InliningOracle) ShouldInline(caller, callee string, bodySize int) InliningDecision {
	// Get profiling data
	calleeProfile := io.profiler.GetProfile(callee)
	if calleeProfile == nil {
		return InliningDecision{
			ShouldInline: false,
			Reason:       "no profiling data",
		}
	}

	// Check call count
	callCount := calleeProfile.CalledBy[caller]
	if callCount < io.minCallCount {
		return InliningDecision{
			ShouldInline: false,
			Reason:       fmt.Sprintf("insufficient calls (%d < %d)", callCount, io.minCallCount),
		}
	}

	// Check body size
	if bodySize > io.maxInlineSize {
		return InliningDecision{
			ShouldInline: false,
			Reason:       fmt.Sprintf("body too large (%d > %d)", bodySize, io.maxInlineSize),
		}
	}

	// Calculate benefit score
	// Higher call count and smaller body = higher score
	score := float64(callCount) / float64(bodySize+1)

	if score < io.inlineThreshold {
		return InliningDecision{
			ShouldInline: false,
			Reason:       fmt.Sprintf("score too low (%.2f < %.2f)", score, io.inlineThreshold),
		}
	}

	return InliningDecision{
		ShouldInline: true,
		Reason:       "hot call site with small body",
		Benefit:      score,
	}
}

// GetInlineCandidates returns functions ranked by inlining benefit
func (io *InliningOracle) GetInlineCandidates() []InlineCandidate {
	candidates := io.profiler.GetInlineCandidates(io.minCallCount)

	result := make([]InlineCandidate, 0, len(candidates))
	for _, name := range candidates {
		profile := io.profiler.GetProfile(name)
		if profile == nil {
			continue
		}

		// Calculate total calls
		var totalCalls int64
		for _, count := range profile.CalledBy {
			totalCalls += count
		}

		candidate := InlineCandidate{
			Name:        name,
			CallCount:   totalCalls,
			CallerCount: len(profile.CalledBy),
			Score:       float64(totalCalls) / float64(len(profile.CalledBy)+1),
		}
		result = append(result, candidate)
	}

	return result
}

// AdaptiveRecompilationTrigger determines when to trigger recompilation
type AdaptiveRecompilationTrigger struct {
	profiler            *Profiler
	executionThreshold  int64
	timeThreshold       float64 // Average time increase threshold (ratio)
	typeChangeThreshold int
}

// NewAdaptiveRecompilationTrigger creates a new adaptive trigger
func NewAdaptiveRecompilationTrigger(profiler *Profiler) *AdaptiveRecompilationTrigger {
	return &AdaptiveRecompilationTrigger{
		profiler:            profiler,
		executionThreshold:  50,  // Recompile after this many executions
		timeThreshold:       1.5, // Recompile if time increases by 50%
		typeChangeThreshold: 5,   // Recompile after this many type changes
	}
}

// RecompilationTrigger describes why recompilation was triggered
type RecompilationTrigger struct {
	ShouldRecompile bool
	Reason          string
	Priority        int // Higher = more urgent
}

// ShouldRecompile checks if a route should be recompiled
func (art *AdaptiveRecompilationTrigger) ShouldRecompile(routeName string, currentTier OptimizationTier) RecompilationTrigger {
	profile := art.profiler.GetProfile(routeName)
	if profile == nil {
		return RecompilationTrigger{ShouldRecompile: false}
	}

	// Check execution count
	if profile.ExecutionCount >= art.executionThreshold && currentTier < TierOptimized {
		return RecompilationTrigger{
			ShouldRecompile: true,
			Reason:          "high execution count",
			Priority:        2,
		}
	}

	// Check if very hot (for aggressive optimization)
	if profile.ExecutionCount >= art.executionThreshold*10 && currentTier < TierHighlyOptimized {
		return RecompilationTrigger{
			ShouldRecompile: true,
			Reason:          "very high execution count",
			Priority:        3,
		}
	}

	// Check type stability for specialization opportunity
	monomorphicVars := art.profiler.GetMonomorphicVariables(routeName)
	if len(monomorphicVars) > 0 && currentTier < TierHighlyOptimized {
		return RecompilationTrigger{
			ShouldRecompile: true,
			Reason:          fmt.Sprintf("type-stable variables detected (%d)", len(monomorphicVars)),
			Priority:        1,
		}
	}

	return RecompilationTrigger{ShouldRecompile: false}
}

// DeoptimizationRecord tracks why deoptimization occurred
type DeoptimizationRecord struct {
	RouteName    string
	Reason       string
	OccurredAt   int64 // Timestamp
	FromTier     OptimizationTier
	TypeMismatch map[string]string // Expected -> Actual
}

// DeoptimizationTracker tracks deoptimization events
type DeoptimizationTracker struct {
	records    []DeoptimizationRecord
	mutex      sync.RWMutex
	maxRecords int
}

// NewDeoptimizationTracker creates a new tracker
func NewDeoptimizationTracker() *DeoptimizationTracker {
	return &DeoptimizationTracker{
		records:    make([]DeoptimizationRecord, 0),
		maxRecords: 100,
	}
}

// RecordDeoptimization records a deoptimization event
func (dt *DeoptimizationTracker) RecordDeoptimization(record DeoptimizationRecord) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	dt.records = append(dt.records, record)
	if len(dt.records) > dt.maxRecords {
		dt.records = dt.records[1:]
	}
}

// GetRecords returns recent deoptimization records
func (dt *DeoptimizationTracker) GetRecords(limit int) []DeoptimizationRecord {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	if limit > len(dt.records) {
		limit = len(dt.records)
	}

	result := make([]DeoptimizationRecord, limit)
	copy(result, dt.records[len(dt.records)-limit:])
	return result
}

// GetDeoptimizationCount returns total deoptimization count for a route
func (dt *DeoptimizationTracker) GetDeoptimizationCount(routeName string) int {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	count := 0
	for _, record := range dt.records {
		if record.RouteName == routeName {
			count++
		}
	}
	return count
}
