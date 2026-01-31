package jit

import (
	"fmt"
	"github.com/glyphlang/glyph/pkg/interpreter"
	"sync"
	"time"

	"github.com/glyphlang/glyph/pkg/compiler"
)

// HotPathThreshold is the number of executions before code becomes "hot"
const (
	DefaultHotPathThreshold = 100
	DefaultRecompileWindow  = 10 * time.Second
)

// OptimizationTier represents different levels of optimization
type OptimizationTier int

const (
	TierInterpreted     OptimizationTier = iota // No compilation
	TierBaseline                                // Basic compilation, no optimization
	TierOptimized                               // Standard optimizations
	TierHighlyOptimized                         // Aggressive optimizations for hot code
)

// CompilationUnit represents a compiled route or function
type CompilationUnit struct {
	Name           string
	Bytecode       []byte
	Tier           OptimizationTier
	CompiledAt     time.Time
	ExecutionCount int64
	LastExecuted   time.Time
}

// JITCompiler manages just-in-time compilation
type JITCompiler struct {
	// Configuration
	hotPathThreshold int
	recompileWindow  time.Duration
	configMux        sync.RWMutex

	// Profiler for runtime data
	profiler *Profiler

	// Compiled units cache
	units    map[string]*CompilationUnit
	unitsMux sync.RWMutex

	// Type specialization
	specializationCache *SpecializationCache
	typeCompiler        *TypeSpecializedCompiler

	// Inlining decisions
	inliningOracle *InliningOracle

	// Adaptive recompilation
	recompileTrigger *AdaptiveRecompilationTrigger

	// Deoptimization tracking
	deoptTracker *DeoptimizationTracker

	// Statistics
	stats    JITStats
	statsMux sync.RWMutex
}

// JITStats tracks JIT compilation statistics
type JITStats struct {
	TotalCompilations      int64
	BaselineCompilations   int64
	OptimizedCompilations  int64
	AggressiveCompilations int64
	Recompilations         int64
	CacheHits              int64
	CacheMisses            int64
	TotalExecutionTime     time.Duration
	TotalCompilationTime   time.Duration
	SpecializationHits     int64
	SpecializationMisses   int64
	InlinedFunctions       int64
	Deoptimizations        int64
	AdaptiveRecompilations int64
}

// NewJITCompiler creates a new JIT compiler instance
func NewJITCompiler() *JITCompiler {
	profiler := NewProfiler()
	return &JITCompiler{
		hotPathThreshold:    DefaultHotPathThreshold,
		recompileWindow:     DefaultRecompileWindow,
		profiler:            profiler,
		units:               make(map[string]*CompilationUnit),
		specializationCache: NewSpecializationCache(),
		typeCompiler:        NewTypeSpecializedCompiler(profiler),
		inliningOracle:      NewInliningOracle(profiler),
		recompileTrigger:    NewAdaptiveRecompilationTrigger(profiler),
		deoptTracker:        NewDeoptimizationTracker(),
	}
}

// NewJITCompilerWithConfig creates a JIT compiler with custom configuration
func NewJITCompilerWithConfig(hotPathThreshold int, recompileWindow time.Duration) *JITCompiler {
	jit := NewJITCompiler()
	jit.hotPathThreshold = hotPathThreshold
	jit.recompileWindow = recompileWindow
	return jit
}

// CompileRoute compiles or recompiles a route based on profiling data
func (jit *JITCompiler) CompileRoute(name string, route *interpreter.Route) ([]byte, error) {
	startTime := time.Now()

	// Check if we have a cached compiled unit
	jit.unitsMux.RLock()
	unit, exists := jit.units[name]
	jit.unitsMux.RUnlock()

	if exists {
		// Update statistics
		jit.statsMux.Lock()
		jit.stats.CacheHits++
		jit.statsMux.Unlock()

		// Check if we should recompile to a higher tier
		if jit.shouldRecompile(unit) {
			return jit.recompileRoute(name, route, unit)
		}

		return unit.Bytecode, nil
	}

	// Cache miss - compile for the first time
	jit.statsMux.Lock()
	jit.stats.CacheMisses++
	jit.statsMux.Unlock()

	// Determine initial optimization tier
	tier := jit.determineInitialTier(name)

	// Compile with appropriate tier
	bytecode, err := jit.compileWithTier(route, tier)
	if err != nil {
		return nil, fmt.Errorf("JIT compilation failed for %s: %w", name, err)
	}

	// Create compilation unit
	unit = &CompilationUnit{
		Name:           name,
		Bytecode:       bytecode,
		Tier:           tier,
		CompiledAt:     time.Now(),
		ExecutionCount: 0,
		LastExecuted:   time.Now(),
	}

	// Cache the unit
	jit.unitsMux.Lock()
	jit.units[name] = unit
	jit.unitsMux.Unlock()

	// Update statistics
	jit.statsMux.Lock()
	jit.stats.TotalCompilations++
	jit.stats.TotalCompilationTime += time.Since(startTime)
	jit.updateTierStats(tier)
	jit.statsMux.Unlock()

	return bytecode, nil
}

// RecordExecution records an execution of a route for profiling
func (jit *JITCompiler) RecordExecution(name string, executionTime time.Duration) {
	// Update profiler
	jit.profiler.RecordExecution(name, executionTime)

	// Update compilation unit
	jit.unitsMux.Lock()
	if unit, exists := jit.units[name]; exists {
		unit.ExecutionCount++
		unit.LastExecuted = time.Now()
	}
	jit.unitsMux.Unlock()

	// Update statistics
	jit.statsMux.Lock()
	jit.stats.TotalExecutionTime += executionTime
	jit.statsMux.Unlock()
}

// shouldRecompile determines if a route should be recompiled to a higher tier
func (jit *JITCompiler) shouldRecompile(unit *CompilationUnit) bool {
	// Don't recompile if already at highest tier
	if unit.Tier >= TierHighlyOptimized {
		return false
	}

	// Check if execution count exceeds hot path threshold
	profile := jit.profiler.GetProfile(unit.Name)
	if profile == nil {
		return false
	}

	// Read config values with lock protection
	jit.configMux.RLock()
	hotPathThreshold := jit.hotPathThreshold
	recompileWindow := jit.recompileWindow
	jit.configMux.RUnlock()

	// Recompile if:
	// 1. Execution count is high enough
	// 2. It's been long enough since last compilation
	timeSinceCompile := time.Since(unit.CompiledAt)

	switch unit.Tier {
	case TierInterpreted, TierBaseline:
		// Upgrade to optimized if executed frequently
		return profile.ExecutionCount >= int64(hotPathThreshold/2) &&
			timeSinceCompile > recompileWindow
	case TierOptimized:
		// Upgrade to highly optimized if very hot
		return profile.ExecutionCount >= int64(hotPathThreshold) &&
			timeSinceCompile > recompileWindow
	default:
		return false
	}
}

// recompileRoute recompiles a route to a higher optimization tier
func (jit *JITCompiler) recompileRoute(name string, route *interpreter.Route, currentUnit *CompilationUnit) ([]byte, error) {
	startTime := time.Now()

	// Determine next tier
	nextTier := jit.getNextTier(currentUnit.Tier)

	// Compile with new tier
	bytecode, err := jit.compileWithTier(route, nextTier)
	if err != nil {
		return nil, fmt.Errorf("recompilation failed for %s: %w", name, err)
	}

	// Update compilation unit
	jit.unitsMux.Lock()
	currentUnit.Bytecode = bytecode
	currentUnit.Tier = nextTier
	currentUnit.CompiledAt = time.Now()
	jit.unitsMux.Unlock()

	// Update statistics
	jit.statsMux.Lock()
	jit.stats.Recompilations++
	jit.stats.TotalCompilations++
	jit.stats.TotalCompilationTime += time.Since(startTime)
	jit.updateTierStats(nextTier)
	jit.statsMux.Unlock()

	return bytecode, nil
}

// compileWithTier compiles a route with a specific optimization tier
func (jit *JITCompiler) compileWithTier(route *interpreter.Route, tier OptimizationTier) ([]byte, error) {
	var comp *compiler.Compiler

	// Create a new compiler instance for each compilation to avoid race conditions
	switch tier {
	case TierBaseline:
		comp = compiler.NewCompilerWithOptLevel(compiler.OptNone)
	case TierOptimized:
		comp = compiler.NewCompilerWithOptLevel(compiler.OptBasic)
	case TierHighlyOptimized:
		comp = compiler.NewCompilerWithOptLevel(compiler.OptAggressive)
	default:
		comp = compiler.NewCompilerWithOptLevel(compiler.OptNone)
	}

	return comp.CompileRoute(route)
}

// determineInitialTier determines the initial optimization tier for a route
func (jit *JITCompiler) determineInitialTier(name string) OptimizationTier {
	// Check if we have profiling data from previous runs
	profile := jit.profiler.GetProfile(name)
	if profile == nil {
		return TierBaseline
	}

	// Read config value with lock protection
	jit.configMux.RLock()
	hotPathThreshold := jit.hotPathThreshold
	jit.configMux.RUnlock()

	// If we have historical data showing this is hot, start optimized
	if profile.ExecutionCount >= int64(hotPathThreshold) {
		return TierOptimized
	}

	return TierBaseline
}

// getNextTier returns the next optimization tier
func (jit *JITCompiler) getNextTier(current OptimizationTier) OptimizationTier {
	switch current {
	case TierInterpreted:
		return TierBaseline
	case TierBaseline:
		return TierOptimized
	case TierOptimized:
		return TierHighlyOptimized
	default:
		return current
	}
}

// updateTierStats updates tier-specific statistics
func (jit *JITCompiler) updateTierStats(tier OptimizationTier) {
	switch tier {
	case TierBaseline:
		jit.stats.BaselineCompilations++
	case TierOptimized:
		jit.stats.OptimizedCompilations++
	case TierHighlyOptimized:
		jit.stats.AggressiveCompilations++
	}
}

// GetStats returns current JIT statistics
func (jit *JITCompiler) GetStats() JITStats {
	jit.statsMux.RLock()
	defer jit.statsMux.RUnlock()
	return jit.stats
}

// GetUnit returns a copy of a compilation unit by name
func (jit *JITCompiler) GetUnit(name string) (*CompilationUnit, bool) {
	jit.unitsMux.RLock()
	defer jit.unitsMux.RUnlock()

	unit, exists := jit.units[name]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions when caller reads while recompileRoute writes
	unitCopy := &CompilationUnit{
		Name:           unit.Name,
		Bytecode:       make([]byte, len(unit.Bytecode)),
		Tier:           unit.Tier,
		CompiledAt:     unit.CompiledAt,
		ExecutionCount: unit.ExecutionCount,
		LastExecuted:   unit.LastExecuted,
	}
	copy(unitCopy.Bytecode, unit.Bytecode)

	return unitCopy, true
}

// GetHotPaths returns a list of hot paths (frequently executed routes)
func (jit *JITCompiler) GetHotPaths() []string {
	jit.configMux.RLock()
	hotPathThreshold := jit.hotPathThreshold
	jit.configMux.RUnlock()

	hotPaths := jit.profiler.GetHotPaths(hotPathThreshold)
	return hotPaths
}

// InvalidateCache removes a compilation unit from the cache
func (jit *JITCompiler) InvalidateCache(name string) {
	jit.unitsMux.Lock()
	delete(jit.units, name)
	jit.unitsMux.Unlock()
}

// ClearCache removes all compilation units from the cache
func (jit *JITCompiler) ClearCache() {
	jit.unitsMux.Lock()
	jit.units = make(map[string]*CompilationUnit)
	jit.unitsMux.Unlock()
}

// GetProfiler returns the profiler instance
func (jit *JITCompiler) GetProfiler() *Profiler {
	return jit.profiler
}

// SetHotPathThreshold sets the hot path threshold
func (jit *JITCompiler) SetHotPathThreshold(threshold int) {
	jit.configMux.Lock()
	jit.hotPathThreshold = threshold
	jit.configMux.Unlock()
}

// SetRecompileWindow sets the recompilation window
func (jit *JITCompiler) SetRecompileWindow(window time.Duration) {
	jit.configMux.Lock()
	jit.recompileWindow = window
	jit.configMux.Unlock()
}

// CompileRouteWithTypes compiles a route with type specialization
func (jit *JITCompiler) CompileRouteWithTypes(name string, route *interpreter.Route, types map[string]string) ([]byte, error) {
	// Check for existing specialization
	if spec := jit.specializationCache.GetSpecialization(name, types); spec != nil {
		jit.statsMux.Lock()
		jit.stats.SpecializationHits++
		jit.statsMux.Unlock()
		return spec.Bytecode, nil
	}

	jit.statsMux.Lock()
	jit.stats.SpecializationMisses++
	jit.statsMux.Unlock()

	// Compile with type information
	bytecode, err := jit.typeCompiler.CompileWithTypeInfo(route, types)
	if err != nil {
		return nil, err
	}

	// Cache the specialization
	jit.specializationCache.AddSpecialization(name, types, bytecode)

	return bytecode, nil
}

// CheckAdaptiveRecompilation checks if a route should be recompiled based on profiling
func (jit *JITCompiler) CheckAdaptiveRecompilation(name string, route *interpreter.Route) (bool, error) {
	jit.unitsMux.RLock()
	unit, exists := jit.units[name]
	jit.unitsMux.RUnlock()

	if !exists {
		return false, nil
	}

	trigger := jit.recompileTrigger.ShouldRecompile(name, unit.Tier)
	if !trigger.ShouldRecompile {
		return false, nil
	}

	// Recompile with the next tier
	bytecode, err := jit.recompileRoute(name, route, unit)
	if err != nil {
		return false, err
	}

	jit.statsMux.Lock()
	jit.stats.AdaptiveRecompilations++
	jit.statsMux.Unlock()

	_ = bytecode // Already cached by recompileRoute
	return true, nil
}

// RecordDeoptimization records when specialized code had to deoptimize
func (jit *JITCompiler) RecordDeoptimization(routeName string, reason string, typeMismatch map[string]string) {
	jit.unitsMux.RLock()
	unit, exists := jit.units[routeName]
	jit.unitsMux.RUnlock()

	var fromTier OptimizationTier
	if exists {
		fromTier = unit.Tier
	}

	record := DeoptimizationRecord{
		RouteName:    routeName,
		Reason:       reason,
		OccurredAt:   time.Now().Unix(),
		FromTier:     fromTier,
		TypeMismatch: typeMismatch,
	}

	jit.deoptTracker.RecordDeoptimization(record)

	jit.statsMux.Lock()
	jit.stats.Deoptimizations++
	jit.statsMux.Unlock()

	// Invalidate specializations for this route
	jit.specializationCache.InvalidateSpecializations(routeName)
}

// GetInlineCandidates returns functions that are good candidates for inlining
func (jit *JITCompiler) GetInlineCandidates() []InlineCandidate {
	return jit.inliningOracle.GetInlineCandidates()
}

// ShouldInline checks if a function should be inlined at a specific call site
func (jit *JITCompiler) ShouldInline(caller, callee string, bodySize int) InliningDecision {
	return jit.inliningOracle.ShouldInline(caller, callee, bodySize)
}

// GetSpecializationStats returns type specialization statistics
func (jit *JITCompiler) GetSpecializationStats() map[string]SpecializationStats {
	return jit.specializationCache.GetStats()
}

// GetDeoptimizationRecords returns recent deoptimization records
func (jit *JITCompiler) GetDeoptimizationRecords(limit int) []DeoptimizationRecord {
	return jit.deoptTracker.GetRecords(limit)
}

// GetDetailedStats returns comprehensive JIT statistics
func (jit *JITCompiler) GetDetailedStats() map[string]interface{} {
	jit.statsMux.RLock()
	stats := jit.stats
	jit.statsMux.RUnlock()

	profilerStats := jit.profiler.GetStats()
	specStats := jit.specializationCache.GetStats()

	return map[string]interface{}{
		"compilations": map[string]int64{
			"total":      stats.TotalCompilations,
			"baseline":   stats.BaselineCompilations,
			"optimized":  stats.OptimizedCompilations,
			"aggressive": stats.AggressiveCompilations,
			"adaptive":   stats.AdaptiveRecompilations,
		},
		"cache": map[string]int64{
			"hits":   stats.CacheHits,
			"misses": stats.CacheMisses,
		},
		"specialization": map[string]interface{}{
			"hits":    stats.SpecializationHits,
			"misses":  stats.SpecializationMisses,
			"details": specStats,
		},
		"profiler": map[string]interface{}{
			"totalProfiles":   profilerStats.TotalProfiles,
			"totalExecutions": profilerStats.TotalExecutions,
			"callGraphSize":   profilerStats.CallGraphSize,
		},
		"deoptimizations":  stats.Deoptimizations,
		"inlinedFunctions": stats.InlinedFunctions,
		"timing": map[string]string{
			"totalCompilationTime": stats.TotalCompilationTime.String(),
			"totalExecutionTime":   stats.TotalExecutionTime.String(),
		},
	}
}
