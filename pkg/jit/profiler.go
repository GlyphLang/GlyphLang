package jit

import (
	"sort"
	"sync"
	"time"
)

// ExecutionProfile contains profiling data for a route or function
type ExecutionProfile struct {
	Name           string
	ExecutionCount int64
	TotalTime      time.Duration
	AverageTime    time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	LastExecuted   time.Time

	// Type profiling - track actual types seen at runtime
	TypeProfile    *TypeProfile

	// Call graph data
	CalledBy       map[string]int64 // caller -> count
	Calls          map[string]int64 // callee -> count
}

// TypeProfile tracks runtime type information for adaptive optimization
type TypeProfile struct {
	// Variable name -> type name -> count
	VariableTypes map[string]map[string]int64

	// Return type tracking
	ReturnTypes map[string]int64 // type -> count
}

// CallGraphNode represents a node in the call graph
type CallGraphNode struct {
	Name        string
	Callers     map[string]int64 // caller -> count
	Callees     map[string]int64 // callee -> count
	IsHot       bool             // true if this is a hot path
}

// Profiler collects runtime profiling data
type Profiler struct {
	profiles   map[string]*ExecutionProfile
	callGraph  map[string]*CallGraphNode
	mutex      sync.RWMutex

	// Configuration
	maxProfiles int // Maximum number of profiles to keep
}

// NewProfiler creates a new profiler instance
func NewProfiler() *Profiler {
	return &Profiler{
		profiles:    make(map[string]*ExecutionProfile),
		callGraph:   make(map[string]*CallGraphNode),
		maxProfiles: 1000,
	}
}

// RecordExecution records an execution of a route/function
func (p *Profiler) RecordExecution(name string, executionTime time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	profile, exists := p.profiles[name]
	if !exists {
		profile = &ExecutionProfile{
			Name:          name,
			ExecutionCount: 0,
			TotalTime:     0,
			MinTime:       executionTime,
			MaxTime:       executionTime,
			TypeProfile:   NewTypeProfile(),
			CalledBy:      make(map[string]int64),
			Calls:         make(map[string]int64),
		}
		p.profiles[name] = profile
	}

	// Update execution statistics
	profile.ExecutionCount++
	profile.TotalTime += executionTime
	profile.AverageTime = time.Duration(int64(profile.TotalTime) / profile.ExecutionCount)
	profile.LastExecuted = time.Now()

	// Update min/max times
	if executionTime < profile.MinTime {
		profile.MinTime = executionTime
	}
	if executionTime > profile.MaxTime {
		profile.MaxTime = executionTime
	}

	// Enforce max profiles limit (LRU-style eviction)
	if len(p.profiles) > p.maxProfiles {
		p.evictOldestProfile()
	}
}

// RecordTypeUsage records the type of a variable at runtime
func (p *Profiler) RecordTypeUsage(routeName, variableName, typeName string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	profile := p.getOrCreateProfile(routeName)

	if profile.TypeProfile.VariableTypes[variableName] == nil {
		profile.TypeProfile.VariableTypes[variableName] = make(map[string]int64)
	}

	profile.TypeProfile.VariableTypes[variableName][typeName]++
}

// RecordReturnType records the return type of a function
func (p *Profiler) RecordReturnType(routeName, typeName string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	profile := p.getOrCreateProfile(routeName)
	profile.TypeProfile.ReturnTypes[typeName]++
}

// RecordCall records a function call in the call graph
func (p *Profiler) RecordCall(caller, callee string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Update caller's profile
	callerProfile := p.getOrCreateProfile(caller)
	callerProfile.Calls[callee]++

	// Update callee's profile
	calleeProfile := p.getOrCreateProfile(callee)
	calleeProfile.CalledBy[caller]++

	// Update call graph
	callerNode := p.getOrCreateCallGraphNode(caller)
	callerNode.Callees[callee]++

	calleeNode := p.getOrCreateCallGraphNode(callee)
	calleeNode.Callers[caller]++
}

// GetProfile returns a copy of the execution profile for a route
func (p *Profiler) GetProfile(name string) *ExecutionProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profile, exists := p.profiles[name]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions when caller reads while RecordExecution writes
	profileCopy := &ExecutionProfile{
		Name:           profile.Name,
		ExecutionCount: profile.ExecutionCount,
		TotalTime:      profile.TotalTime,
		AverageTime:    profile.AverageTime,
		MinTime:        profile.MinTime,
		MaxTime:        profile.MaxTime,
		LastExecuted:   profile.LastExecuted,
	}

	// Deep copy TypeProfile
	if profile.TypeProfile != nil {
		profileCopy.TypeProfile = &TypeProfile{
			VariableTypes: make(map[string]map[string]int64),
			ReturnTypes:   make(map[string]int64),
		}
		for varName, types := range profile.TypeProfile.VariableTypes {
			profileCopy.TypeProfile.VariableTypes[varName] = make(map[string]int64)
			for typeName, count := range types {
				profileCopy.TypeProfile.VariableTypes[varName][typeName] = count
			}
		}
		for typeName, count := range profile.TypeProfile.ReturnTypes {
			profileCopy.TypeProfile.ReturnTypes[typeName] = count
		}
	}

	// Deep copy CalledBy
	profileCopy.CalledBy = make(map[string]int64)
	for k, v := range profile.CalledBy {
		profileCopy.CalledBy[k] = v
	}

	// Deep copy Calls
	profileCopy.Calls = make(map[string]int64)
	for k, v := range profile.Calls {
		profileCopy.Calls[k] = v
	}

	return profileCopy
}

// GetAllProfiles returns all execution profiles
func (p *Profiler) GetAllProfiles() map[string]*ExecutionProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Create a copy to avoid race conditions
	profiles := make(map[string]*ExecutionProfile, len(p.profiles))
	for k, v := range p.profiles {
		profiles[k] = v
	}
	return profiles
}

// GetHotPaths returns routes that have been executed more than the threshold
func (p *Profiler) GetHotPaths(threshold int) []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	hotPaths := make([]string, 0)
	for name, profile := range p.profiles {
		if profile.ExecutionCount >= int64(threshold) {
			hotPaths = append(hotPaths, name)
		}
	}

	// Sort by execution count (descending)
	sort.Slice(hotPaths, func(i, j int) bool {
		return p.profiles[hotPaths[i]].ExecutionCount > p.profiles[hotPaths[j]].ExecutionCount
	})

	return hotPaths
}

// GetTopNByExecutionCount returns the top N routes by execution count
func (p *Profiler) GetTopNByExecutionCount(n int) []*ExecutionProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profiles := make([]*ExecutionProfile, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, profile)
	}

	// Sort by execution count (descending)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].ExecutionCount > profiles[j].ExecutionCount
	})

	if n > len(profiles) {
		n = len(profiles)
	}

	return profiles[:n]
}

// GetTopNByTotalTime returns the top N routes by total execution time
func (p *Profiler) GetTopNByTotalTime(n int) []*ExecutionProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profiles := make([]*ExecutionProfile, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, profile)
	}

	// Sort by total time (descending)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].TotalTime > profiles[j].TotalTime
	})

	if n > len(profiles) {
		n = len(profiles)
	}

	return profiles[:n]
}

// GetCallGraph returns the call graph
func (p *Profiler) GetCallGraph() map[string]*CallGraphNode {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Create a copy
	graph := make(map[string]*CallGraphNode, len(p.callGraph))
	for k, v := range p.callGraph {
		graph[k] = v
	}
	return graph
}

// AnalyzeTypeStability analyzes if a variable has stable types (monomorphic)
func (p *Profiler) AnalyzeTypeStability(routeName, variableName string) (isStable bool, dominantType string) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profile, exists := p.profiles[routeName]
	if !exists {
		return false, ""
	}

	types, exists := profile.TypeProfile.VariableTypes[variableName]
	if !exists || len(types) == 0 {
		return false, ""
	}

	// A variable is type-stable if one type dominates (>95% of uses)
	var maxType string
	var maxCount int64
	var totalCount int64

	for typeName, count := range types {
		totalCount += count
		if count > maxCount {
			maxCount = count
			maxType = typeName
		}
	}

	// Type is stable if one type accounts for >95% of uses
	stability := float64(maxCount) / float64(totalCount)
	return stability > 0.95, maxType
}

// GetMonomorphicVariables returns variables that have stable types
func (p *Profiler) GetMonomorphicVariables(routeName string) map[string]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := make(map[string]string)

	profile, exists := p.profiles[routeName]
	if !exists {
		return result
	}

	for varName := range profile.TypeProfile.VariableTypes {
		if isStable, dominantType := p.AnalyzeTypeStability(routeName, varName); isStable {
			result[varName] = dominantType
		}
	}

	return result
}

// MarkHotPaths marks hot paths in the call graph
func (p *Profiler) MarkHotPaths(threshold int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for name, node := range p.callGraph {
		profile, exists := p.profiles[name]
		if exists && profile.ExecutionCount >= int64(threshold) {
			node.IsHot = true
		} else {
			node.IsHot = false
		}
	}
}

// GetInlineCandidates returns functions that are good candidates for inlining
func (p *Profiler) GetInlineCandidates(minCallCount int64) []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	candidates := make([]string, 0)

	for name, profile := range p.profiles {
		// A function is a good inline candidate if:
		// 1. It's called frequently
		// 2. It has few callers (reduces code bloat)
		totalCalls := int64(0)
		for _, count := range profile.CalledBy {
			totalCalls += count
		}

		if totalCalls >= minCallCount && len(profile.CalledBy) <= 3 {
			candidates = append(candidates, name)
		}
	}

	return candidates
}

// Reset clears all profiling data
func (p *Profiler) Reset() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.profiles = make(map[string]*ExecutionProfile)
	p.callGraph = make(map[string]*CallGraphNode)
}

// Helper functions

func (p *Profiler) getOrCreateProfile(name string) *ExecutionProfile {
	profile, exists := p.profiles[name]
	if !exists {
		profile = &ExecutionProfile{
			Name:          name,
			ExecutionCount: 0,
			TypeProfile:   NewTypeProfile(),
			CalledBy:      make(map[string]int64),
			Calls:         make(map[string]int64),
		}
		p.profiles[name] = profile
	}
	return profile
}

func (p *Profiler) getOrCreateCallGraphNode(name string) *CallGraphNode {
	node, exists := p.callGraph[name]
	if !exists {
		node = &CallGraphNode{
			Name:    name,
			Callers: make(map[string]int64),
			Callees: make(map[string]int64),
			IsHot:   false,
		}
		p.callGraph[name] = node
	}
	return node
}

func (p *Profiler) evictOldestProfile() {
	var oldestName string
	var oldestTime time.Time
	first := true

	for name, profile := range p.profiles {
		if first || profile.LastExecuted.Before(oldestTime) {
			oldestName = name
			oldestTime = profile.LastExecuted
			first = false
		}
	}

	if oldestName != "" {
		delete(p.profiles, oldestName)
	}
}

// NewTypeProfile creates a new type profile
func NewTypeProfile() *TypeProfile {
	return &TypeProfile{
		VariableTypes: make(map[string]map[string]int64),
		ReturnTypes:   make(map[string]int64),
	}
}

// ProfilerStats returns statistics about the profiler itself
type ProfilerStats struct {
	TotalProfiles     int
	HotPathCount      int
	TotalExecutions   int64
	TotalExecutionTime time.Duration
	CallGraphSize     int
}

// GetStats returns profiler statistics
func (p *Profiler) GetStats() ProfilerStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := ProfilerStats{
		TotalProfiles: len(p.profiles),
		CallGraphSize: len(p.callGraph),
	}

	for _, profile := range p.profiles {
		stats.TotalExecutions += profile.ExecutionCount
		stats.TotalExecutionTime += profile.TotalTime
	}

	return stats
}
