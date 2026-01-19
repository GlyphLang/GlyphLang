package hotreload

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/config"
)

func TestFileWatcher_MatchesPattern(t *testing.T) {
	w := &FileWatcher{
		patterns: []string{"*.glyph", "*.glyph"},
	}

	tests := []struct {
		path    string
		matches bool
	}{
		{"main.glyph", true},
		{"app.glyph", true},
		{"test.go", false},
		{"readme.md", false},
		{"path/to/file.glyph", true},
	}

	for _, tt := range tests {
		result := w.matchesPattern(tt.path)
		if result != tt.matches {
			t.Errorf("matchesPattern(%q) = %v, want %v", tt.path, result, tt.matches)
		}
	}
}

func TestFileWatcher_ShouldExclude(t *testing.T) {
	w := &FileWatcher{
		excludes: []string{"node_modules", ".git", "vendor"},
	}

	tests := []struct {
		path     string
		excluded bool
	}{
		{"src/main.glyph", false},
		{"node_modules/pkg/file.js", true},
		{".git/config", true},
		{"vendor/lib/code.go", true},
		{"app/main.glyph", false},
	}

	for _, tt := range tests {
		result := w.shouldExclude(tt.path)
		if result != tt.excluded {
			t.Errorf("shouldExclude(%q) = %v, want %v", tt.path, result, tt.excluded)
		}
	}
}

func TestFileWatcher_DetectChanges(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "hotreload-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial file
	testFile := filepath.Join(tmpDir, "test.glyph")
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create watcher
	var changes []FileChange
	var mu sync.Mutex

	w := NewFileWatcher([]string{tmpDir}, func(c []FileChange) {
		mu.Lock()
		changes = append(changes, c...)
		mu.Unlock()
	}, WithPollInterval(50*time.Millisecond), WithDebounce(10*time.Millisecond))

	// Start watching
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Modify file
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for detection
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(changes) == 0 {
		t.Error("Expected at least one change, got none")
	}

	found := false
	for _, c := range changes {
		if c.Path == testFile && c.Type == ChangeTypeModified {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected modified change for test file")
	}
}

func TestFileWatcher_NewFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "hotreload-newfile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	var changes []FileChange
	var mu sync.Mutex

	w := NewFileWatcher([]string{tmpDir}, func(c []FileChange) {
		mu.Lock()
		changes = append(changes, c...)
		mu.Unlock()
	}, WithPollInterval(50*time.Millisecond), WithDebounce(10*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create new file
	time.Sleep(100 * time.Millisecond)
	newFile := filepath.Join(tmpDir, "new.glyph")
	if err := os.WriteFile(newFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// Wait for detection
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	found := false
	for _, c := range changes {
		if c.Path == newFile && c.Type == ChangeTypeCreated {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected created change for new file")
	}
}

func TestFileWatcher_DeleteFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "hotreload-delete-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial file
	testFile := filepath.Join(tmpDir, "delete.glyph")
	if err := os.WriteFile(testFile, []byte("content to delete"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var changes []FileChange
	var mu sync.Mutex

	w := NewFileWatcher([]string{tmpDir}, func(c []FileChange) {
		mu.Lock()
		changes = append(changes, c...)
		mu.Unlock()
	}, WithPollInterval(50*time.Millisecond), WithDebounce(10*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Delete file
	time.Sleep(100 * time.Millisecond)
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Wait for detection
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	found := false
	for _, c := range changes {
		if c.Path == testFile && c.Type == ChangeTypeDeleted {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected deleted change for test file")
	}
}

func TestApplicationState(t *testing.T) {
	state := NewApplicationState()

	// Test Set and Get
	state.Set("key1", "value1")
	state.Set("key2", 42)

	val, ok := state.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected key1=value1, got %v", val)
	}

	val, ok = state.Get("key2")
	if !ok || val != 42 {
		t.Errorf("Expected key2=42, got %v", val)
	}

	// Test non-existent key
	_, ok = state.Get("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Test GetAll
	all := state.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 items, got %d", len(all))
	}

	// Test SetAll
	state.SetAll(map[string]interface{}{"new": "data"})
	all = state.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 item after SetAll, got %d", len(all))
	}
	if all["new"] != "data" {
		t.Error("Expected new=data")
	}
}

func TestChangeType_String(t *testing.T) {
	tests := []struct {
		ct       ChangeType
		expected string
	}{
		{ChangeTypeModified, "modified"},
		{ChangeTypeCreated, "created"},
		{ChangeTypeDeleted, "deleted"},
		{ChangeType(999), "unknown"},
	}

	for _, tt := range tests {
		result := tt.ct.String()
		if result != tt.expected {
			t.Errorf("ChangeType(%d).String() = %q, want %q", tt.ct, result, tt.expected)
		}
	}
}

func TestLiveReloadScript(t *testing.T) {
	script := LiveReloadScript(8080)

	if len(script) == 0 {
		t.Error("Expected non-empty script")
	}

	if !contains(script, "WebSocket") {
		t.Error("Script should contain WebSocket")
	}

	if !contains(script, "8080") {
		t.Error("Script should contain port number")
	}

	if !contains(script, "__livereload") {
		t.Error("Script should contain livereload endpoint")
	}
}

func TestDefaultDevServerConfig(t *testing.T) {
	cfg := DefaultDevServerConfig()

	if cfg.Port != config.DefaultPort {
		t.Errorf("Expected port %d, got %d", config.DefaultPort, cfg.Port)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", cfg.Host)
	}

	if !cfg.LiveReload {
		t.Error("Expected LiveReload to be true")
	}

	if len(cfg.WatchPaths) == 0 {
		t.Error("Expected at least one watch path")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests

func BenchmarkFileWatcher_HashFile(b *testing.B) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "bench-hash-*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some content
	content := make([]byte, 1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	if _, err := tmpFile.Write(content); err != nil {
		b.Fatalf("Failed to write content: %v", err)
	}
	tmpFile.Close()

	w := &FileWatcher{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := w.hashFile(tmpFile.Name())
		if err != nil {
			b.Fatalf("hashFile failed: %v", err)
		}
	}
}

func BenchmarkApplicationState(b *testing.B) {
	state := NewApplicationState()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			state.Set("key", i)
		}
	})

	b.Run("Get", func(b *testing.B) {
		state.Set("key", "value")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state.Get("key")
		}
	})

	b.Run("GetAll", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			state.Set(string(rune(i)), i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			state.GetAll()
		}
	})
}
