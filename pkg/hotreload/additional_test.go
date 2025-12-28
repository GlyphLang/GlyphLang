package hotreload

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestWithPatterns tests the WithPatterns option
func TestWithPatterns(t *testing.T) {
	w := NewFileWatcher([]string{"."}, nil, WithPatterns("*.go", "*.txt"))

	if len(w.patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(w.patterns))
	}

	if w.patterns[0] != "*.go" || w.patterns[1] != "*.txt" {
		t.Errorf("Unexpected patterns: %v", w.patterns)
	}
}

// TestWithExcludes tests the WithExcludes option
func TestWithExcludes(t *testing.T) {
	w := NewFileWatcher([]string{"."}, nil, WithExcludes("build", "dist"))

	if len(w.excludes) != 2 {
		t.Errorf("Expected 2 excludes, got %d", len(w.excludes))
	}

	if w.excludes[0] != "build" || w.excludes[1] != "dist" {
		t.Errorf("Unexpected excludes: %v", w.excludes)
	}
}

// TestFileWatcher_Stop tests the Stop method
func TestFileWatcher_Stop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hotreload-stop-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFileWatcher([]string{tmpDir}, func(c []FileChange) {
		// callback for testing
	}, WithPollInterval(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Wait a bit then stop
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	// Verify it's stopped by checking we can't start again until reset
	w.mu.RLock()
	running := w.running
	w.mu.RUnlock()

	if running {
		t.Error("Expected watcher to be stopped")
	}

	// Calling Stop again should be safe
	w.Stop()
}

// TestFileWatcher_AlreadyRunning tests starting an already running watcher
func TestFileWatcher_AlreadyRunning(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hotreload-running-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	w := NewFileWatcher([]string{tmpDir}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}
	defer w.Stop()

	// Try to start again
	err = w.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running watcher")
	}
}

// TestMatchesPattern_EmptyPatterns tests pattern matching with no patterns
func TestMatchesPattern_EmptyPatterns(t *testing.T) {
	w := &FileWatcher{
		patterns: []string{},
	}

	// Should match everything when no patterns are set
	if !w.matchesPattern("anyfile.txt") {
		t.Error("Expected empty patterns to match everything")
	}
}

// TestHashFile_NonExistent tests hashing a non-existent file
func TestHashFile_NonExistent(t *testing.T) {
	w := &FileWatcher{}

	_, err := w.hashFile("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// MockCompiler for testing
type MockCompiler struct {
	compileError error
	bytecode     []byte
	compileCalls int
}

func (m *MockCompiler) CompileFile(path string) ([]byte, error) {
	m.compileCalls++
	if m.compileError != nil {
		return nil, m.compileError
	}
	return m.bytecode, nil
}

// MockServer for testing
type MockServer struct {
	reloadError   error
	setStateError error
	state         map[string]interface{}
	reloadCalls   int
}

func (m *MockServer) Reload(bytecode []byte) error {
	m.reloadCalls++
	return m.reloadError
}

func (m *MockServer) GetState() map[string]interface{} {
	return m.state
}

func (m *MockServer) SetState(state map[string]interface{}) error {
	m.state = state
	return m.setStateError
}

// TestWithErrorHandler tests the WithErrorHandler option
func TestWithErrorHandler(t *testing.T) {
	handler := func(err error) {
		// error handler for testing
	}

	rm := NewReloadManager([]string{"."}, nil, nil, WithErrorHandler(handler))

	if rm.errorHandler == nil {
		t.Error("Expected error handler to be set")
	}
}

// TestWithOnReload tests the WithOnReload option
func TestWithOnReload(t *testing.T) {
	handler := func(event ReloadEvent) {
		// reload handler for testing
	}

	rm := NewReloadManager([]string{"."}, nil, nil, WithOnReload(handler))

	if rm.onReload == nil {
		t.Error("Expected onReload handler to be set")
	}
}

// TestNewReloadManager tests creating a new reload manager
func TestNewReloadManager(t *testing.T) {
	compiler := &MockCompiler{}
	server := &MockServer{}

	rm := NewReloadManager([]string{"."}, compiler, server)

	if rm.compiler != compiler {
		t.Error("Expected compiler to be set")
	}

	if rm.server != server {
		t.Error("Expected server to be set")
	}

	if rm.state == nil {
		t.Error("Expected state to be initialized")
	}

	if rm.watcher == nil {
		t.Error("Expected watcher to be initialized")
	}
}

// TestReloadManager_StartStop tests starting and stopping the reload manager
func TestReloadManager_StartStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hotreload-rm-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	compiler := &MockCompiler{bytecode: []byte("test")}
	server := &MockServer{}

	rm := NewReloadManager([]string{tmpDir}, compiler, server)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = rm.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start reload manager: %v", err)
	}

	// Let it run a bit
	time.Sleep(50 * time.Millisecond)

	// Stop should work
	rm.Stop()
}

// TestReloadManager_Stats tests the Stats method
func TestReloadManager_Stats(t *testing.T) {
	compiler := &MockCompiler{}
	server := &MockServer{}

	rm := NewReloadManager([]string{"."}, compiler, server)

	stats := rm.Stats()

	if stats.ReloadCount != 0 {
		t.Errorf("Expected initial reload count 0, got %d", stats.ReloadCount)
	}

	if !stats.LastReload.IsZero() {
		t.Error("Expected zero time for last reload initially")
	}
}

// TestReloadManager_HandleChanges tests the handleChanges method
func TestReloadManager_HandleChanges(t *testing.T) {
	compiler := &MockCompiler{bytecode: []byte("bytecode")}
	server := &MockServer{state: map[string]interface{}{"key": "value"}}

	var mu sync.Mutex
	var reloadEvents []ReloadEvent

	rm := NewReloadManager([]string{"."}, compiler, server,
		WithOnReload(func(event ReloadEvent) {
			mu.Lock()
			reloadEvents = append(reloadEvents, event)
			mu.Unlock()
		}),
	)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	rm.handleChanges(changes)

	mu.Lock()
	defer mu.Unlock()

	if len(reloadEvents) != 1 {
		t.Fatalf("Expected 1 reload event, got %d", len(reloadEvents))
	}

	if !reloadEvents[0].Success {
		t.Errorf("Expected successful reload, got error: %v", reloadEvents[0].Error)
	}

	if reloadEvents[0].ReloadCount != 1 {
		t.Errorf("Expected reload count 1, got %d", reloadEvents[0].ReloadCount)
	}
}

// TestReloadManager_HandleChanges_CompileError tests compilation failure
func TestReloadManager_HandleChanges_CompileError(t *testing.T) {
	compiler := &MockCompiler{compileError: os.ErrNotExist}
	server := &MockServer{}

	var mu sync.Mutex
	var errors []error

	rm := NewReloadManager([]string{"."}, compiler, server,
		WithErrorHandler(func(err error) {
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}),
	)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	rm.handleChanges(changes)

	mu.Lock()
	defer mu.Unlock()

	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
}

// TestReloadManager_HandleChanges_ReloadError tests server reload failure
func TestReloadManager_HandleChanges_ReloadError(t *testing.T) {
	compiler := &MockCompiler{bytecode: []byte("bytecode")}
	server := &MockServer{reloadError: os.ErrPermission}

	var mu sync.Mutex
	var errors []error

	rm := NewReloadManager([]string{"."}, compiler, server,
		WithErrorHandler(func(err error) {
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}),
	)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	rm.handleChanges(changes)

	mu.Lock()
	defer mu.Unlock()

	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
}

// TestReloadManager_HandleChanges_NoGlyphFile tests handling non-glyph file changes
func TestReloadManager_HandleChanges_NoGlyphFile(t *testing.T) {
	compiler := &MockCompiler{bytecode: []byte("bytecode")}
	server := &MockServer{}

	rm := NewReloadManager([]string{"."}, compiler, server)

	changes := []FileChange{
		{Path: "readme.md", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	rm.handleChanges(changes)

	// Should not compile anything
	if compiler.compileCalls != 0 {
		t.Errorf("Expected 0 compile calls, got %d", compiler.compileCalls)
	}
}

// TestReloadManager_HandleChanges_NilCompiler tests handling changes with nil compiler
func TestReloadManager_HandleChanges_NilCompiler(t *testing.T) {
	server := &MockServer{}

	rm := NewReloadManager([]string{"."}, nil, server)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	// Should not panic
	rm.handleChanges(changes)
}

// TestReloadManager_HandleChanges_NilServer tests handling changes with nil server
func TestReloadManager_HandleChanges_NilServer(t *testing.T) {
	compiler := &MockCompiler{bytecode: []byte("bytecode")}

	rm := NewReloadManager([]string{"."}, compiler, nil)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	// Should not panic
	rm.handleChanges(changes)
}

// TestReloadManager_HandleError tests error handling
func TestReloadManager_HandleError(t *testing.T) {
	var handledError error

	rm := &ReloadManager{
		errorHandler: func(err error) {
			handledError = err
		},
	}

	rm.handleError(os.ErrNotExist)

	if handledError != os.ErrNotExist {
		t.Errorf("Expected error to be handled, got %v", handledError)
	}
}

// TestReloadManager_HandleError_NilHandler tests error handling with nil handler
func TestReloadManager_HandleError_NilHandler(t *testing.T) {
	rm := &ReloadManager{}

	// Should not panic
	rm.handleError(os.ErrNotExist)
}

// TestReloadManager_NotifyReload tests reload notification
func TestReloadManager_NotifyReload(t *testing.T) {
	var notifiedEvent ReloadEvent

	rm := &ReloadManager{
		onReload: func(event ReloadEvent) {
			notifiedEvent = event
		},
	}

	event := ReloadEvent{
		Success:     true,
		ReloadCount: 5,
	}

	rm.notifyReload(event)

	if notifiedEvent.ReloadCount != 5 {
		t.Errorf("Expected reload count 5, got %d", notifiedEvent.ReloadCount)
	}
}

// TestReloadManager_NotifyReload_NilHandler tests notification with nil handler
func TestReloadManager_NotifyReload_NilHandler(t *testing.T) {
	rm := &ReloadManager{}

	event := ReloadEvent{Success: true}

	// Should not panic
	rm.notifyReload(event)
}

// TestScanError tests scan with non-existent path
func TestScanError(t *testing.T) {
	w := NewFileWatcher([]string{"/non/existent/path"}, nil)

	// scan should not return error for walk failures, just skip
	err := w.scan()
	if err != nil {
		t.Logf("Note: scan returned error %v (acceptable for non-existent path)", err)
	}
}

// TestDetectChanges_DirectCall tests detectChanges directly
func TestDetectChanges_DirectCall(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hotreload-detect-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial file
	testFile := filepath.Join(tmpDir, "test.glyph")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	w := NewFileWatcher([]string{tmpDir}, nil)

	// Initial scan
	if err := w.scan(); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Modify file
	if err := os.WriteFile(testFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Detect changes
	changes := w.detectChanges()

	found := false
	for _, c := range changes {
		if c.Path == testFile && c.Type == ChangeTypeModified {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to detect modified file")
	}
}

// TestReloadEvent tests ReloadEvent struct
func TestReloadEvent(t *testing.T) {
	event := ReloadEvent{
		Changes:     []FileChange{{Path: "test.glyph", Type: ChangeTypeModified}},
		Success:     true,
		Error:       nil,
		Duration:    100 * time.Millisecond,
		ReloadCount: 1,
		Timestamp:   time.Now(),
	}

	if len(event.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(event.Changes))
	}

	if !event.Success {
		t.Error("Expected success to be true")
	}

	if event.Duration != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v", event.Duration)
	}

	if event.Error != nil {
		t.Error("Expected error to be nil")
	}

	if event.ReloadCount != 1 {
		t.Errorf("Expected reload count 1, got %d", event.ReloadCount)
	}

	if event.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

// TestReloadStats tests ReloadStats struct
func TestReloadStats(t *testing.T) {
	now := time.Now()
	stats := ReloadStats{
		ReloadCount: 10,
		LastReload:  now,
	}

	if stats.ReloadCount != 10 {
		t.Errorf("Expected reload count 10, got %d", stats.ReloadCount)
	}

	if stats.LastReload != now {
		t.Error("Expected last reload time to match")
	}
}

// TestDevServerConfig tests DevServerConfig struct
func TestDevServerConfig(t *testing.T) {
	config := DevServerConfig{
		WatchPaths:   []string{".", "src"},
		Port:         3000,
		Host:         "0.0.0.0",
		LiveReload:   true,
		OpenBrowser:  true,
		BuildOnStart: false,
	}

	if len(config.WatchPaths) != 2 {
		t.Errorf("Expected 2 watch paths, got %d", len(config.WatchPaths))
	}

	if config.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", config.Port)
	}

	if config.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Host)
	}

	if !config.LiveReload {
		t.Error("Expected LiveReload to be true")
	}

	if !config.OpenBrowser {
		t.Error("Expected OpenBrowser to be true")
	}

	if config.BuildOnStart {
		t.Error("Expected BuildOnStart to be false")
	}
}

// TestFileWatcher_MultipleOptions tests multiple options
func TestFileWatcher_MultipleOptions(t *testing.T) {
	w := NewFileWatcher(
		[]string{"."},
		nil,
		WithPatterns("*.go"),
		WithExcludes("vendor"),
		WithDebounce(500*time.Millisecond),
		WithPollInterval(1*time.Second),
	)

	if len(w.patterns) != 1 || w.patterns[0] != "*.go" {
		t.Error("Expected patterns to be set correctly")
	}

	if len(w.excludes) != 1 || w.excludes[0] != "vendor" {
		t.Error("Expected excludes to be set correctly")
	}

	if w.debounceTime != 500*time.Millisecond {
		t.Error("Expected debounce to be set correctly")
	}

	if w.pollInterval != 1*time.Second {
		t.Error("Expected poll interval to be set correctly")
	}
}

// TestReloadManager_SetStateError tests state restoration failure
func TestReloadManager_SetStateError(t *testing.T) {
	compiler := &MockCompiler{bytecode: []byte("bytecode")}
	server := &MockServer{
		state:         map[string]interface{}{"key": "value"},
		setStateError: os.ErrPermission,
	}

	rm := NewReloadManager([]string{"."}, compiler, server)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	// Should not panic, just log warning
	rm.handleChanges(changes)

	// Verify reload still completed
	stats := rm.Stats()
	if stats.ReloadCount != 1 {
		t.Errorf("Expected reload count 1, got %d", stats.ReloadCount)
	}
}

// TestMultipleReloads tests multiple consecutive reloads
func TestMultipleReloads(t *testing.T) {
	compiler := &MockCompiler{bytecode: []byte("bytecode")}
	server := &MockServer{}

	var reloadCount int
	var mu sync.Mutex

	rm := NewReloadManager([]string{"."}, compiler, server,
		WithOnReload(func(event ReloadEvent) {
			mu.Lock()
			reloadCount++
			mu.Unlock()
		}),
	)

	changes := []FileChange{
		{Path: "main.glyph", Type: ChangeTypeModified, Timestamp: time.Now()},
	}

	// Trigger multiple reloads
	for i := 0; i < 3; i++ {
		rm.handleChanges(changes)
	}

	mu.Lock()
	defer mu.Unlock()

	if reloadCount != 3 {
		t.Errorf("Expected 3 reloads, got %d", reloadCount)
	}

	stats := rm.Stats()
	if stats.ReloadCount != 3 {
		t.Errorf("Expected stats reload count 3, got %d", stats.ReloadCount)
	}
}

// TestApplicationState_Concurrent tests concurrent access to ApplicationState
func TestApplicationState_Concurrent(t *testing.T) {
	state := NewApplicationState()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key"
			state.Set(key, i)
			state.Get(key)
			state.GetAll()
		}(i)
	}
	wg.Wait()
}
