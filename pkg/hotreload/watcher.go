// Package hotreload provides development hot-reload functionality
package hotreload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/glyphlang/glyph/pkg/config"
)

// ========================================
// File Watcher
// ========================================

// FileWatcher watches for file changes
type FileWatcher struct {
	mu           sync.RWMutex
	watchPaths   []string
	patterns     []string
	excludes     []string
	debounceTime time.Duration
	onChange     func([]FileChange)
	fileHashes   map[string]string
	stop         chan struct{}
	running      bool
	pollInterval time.Duration
}

// FileChange represents a change to a file
type FileChange struct {
	Path      string
	Type      ChangeType
	Timestamp time.Time
}

// ChangeType represents the type of file change
type ChangeType int

const (
	ChangeTypeModified ChangeType = iota
	ChangeTypeCreated
	ChangeTypeDeleted
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeModified:
		return "modified"
	case ChangeTypeCreated:
		return "created"
	case ChangeTypeDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

// WatcherOption configures the file watcher
type WatcherOption func(*FileWatcher)

// WithPatterns sets file patterns to watch (e.g., "*.glyph", "*.glyph")
func WithPatterns(patterns ...string) WatcherOption {
	return func(w *FileWatcher) {
		w.patterns = patterns
	}
}

// WithExcludes sets paths to exclude
func WithExcludes(excludes ...string) WatcherOption {
	return func(w *FileWatcher) {
		w.excludes = excludes
	}
}

// WithDebounce sets the debounce duration
func WithDebounce(d time.Duration) WatcherOption {
	return func(w *FileWatcher) {
		w.debounceTime = d
	}
}

// WithPollInterval sets the poll interval for checking files
func WithPollInterval(d time.Duration) WatcherOption {
	return func(w *FileWatcher) {
		w.pollInterval = d
	}
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(paths []string, onChange func([]FileChange), opts ...WatcherOption) *FileWatcher {
	w := &FileWatcher{
		watchPaths:   paths,
		patterns:     []string{"*.glyph", "*.glyph"},
		excludes:     []string{"node_modules", ".git", "vendor"},
		debounceTime: 100 * time.Millisecond,
		pollInterval: 500 * time.Millisecond,
		onChange:     onChange,
		fileHashes:   make(map[string]string),
		stop:         make(chan struct{}),
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Start begins watching for file changes
func (w *FileWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watcher already running")
	}
	w.running = true
	w.mu.Unlock()

	// Initial scan to populate file hashes
	if err := w.scan(); err != nil {
		return fmt.Errorf("initial scan failed: %w", err)
	}

	// Start watch loop
	go w.watchLoop(ctx)

	return nil
}

// Stop stops the file watcher
func (w *FileWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		close(w.stop)
		w.running = false
	}
}

// watchLoop is the main watch loop
func (w *FileWatcher) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	var pendingChanges []FileChange
	var debounceTimer *time.Timer

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			changes := w.detectChanges()
			if len(changes) > 0 {
				pendingChanges = append(pendingChanges, changes...)

				// Reset debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(w.debounceTime, func() {
					w.mu.Lock()
					changesToNotify := pendingChanges
					pendingChanges = nil
					w.mu.Unlock()

					if len(changesToNotify) > 0 && w.onChange != nil {
						w.onChange(changesToNotify)
					}
				})
			}
		}
	}
}

// scan scans all watch paths and builds initial file hash map
func (w *FileWatcher) scan() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, watchPath := range w.watchPaths {
		err := filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip inaccessible files
			}

			// Skip excluded paths
			if w.shouldExclude(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check if file matches patterns
			if !w.matchesPattern(path) {
				return nil
			}

			// Calculate hash
			hash, err := w.hashFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			w.fileHashes[path] = hash
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// detectChanges detects file changes since last scan
func (w *FileWatcher) detectChanges() []FileChange {
	w.mu.Lock()
	defer w.mu.Unlock()

	var changes []FileChange
	currentFiles := make(map[string]bool)

	for _, watchPath := range w.watchPaths {
		_ = filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if w.shouldExclude(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if !w.matchesPattern(path) {
				return nil
			}

			currentFiles[path] = true

			hash, err := w.hashFile(path)
			if err != nil {
				return nil
			}

			oldHash, exists := w.fileHashes[path]
			if !exists {
				// New file
				changes = append(changes, FileChange{
					Path:      path,
					Type:      ChangeTypeCreated,
					Timestamp: time.Now(),
				})
				w.fileHashes[path] = hash
			} else if oldHash != hash {
				// Modified file
				changes = append(changes, FileChange{
					Path:      path,
					Type:      ChangeTypeModified,
					Timestamp: time.Now(),
				})
				w.fileHashes[path] = hash
			}

			return nil
		})
	}

	// Check for deleted files
	for path := range w.fileHashes {
		if !currentFiles[path] {
			changes = append(changes, FileChange{
				Path:      path,
				Type:      ChangeTypeDeleted,
				Timestamp: time.Now(),
			})
			delete(w.fileHashes, path)
		}
	}

	return changes
}

// shouldExclude checks if a path should be excluded
func (w *FileWatcher) shouldExclude(path string) bool {
	for _, exclude := range w.excludes {
		if strings.Contains(path, exclude) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a file matches the watch patterns
func (w *FileWatcher) matchesPattern(path string) bool {
	if len(w.patterns) == 0 {
		return true
	}

	for _, pattern := range w.patterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}
	return false
}

// hashFile calculates the SHA256 hash of a file
func (w *FileWatcher) hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ========================================
// Hot Reload Manager
// ========================================

// ReloadManager manages hot-reloading of Glyph applications
type ReloadManager struct {
	mu           sync.RWMutex
	watcher      *FileWatcher
	compiler     CompilerInterface
	server       ServerInterface
	state        *ApplicationState
	reloadCount  int
	lastReload   time.Time
	errorHandler func(error)
	onReload     func(ReloadEvent)
}

// CompilerInterface defines the interface for the Glyph compiler
type CompilerInterface interface {
	CompileFile(path string) ([]byte, error)
}

// ServerInterface defines the interface for the Glyph server
type ServerInterface interface {
	Reload(bytecode []byte) error
	GetState() map[string]interface{}
	SetState(state map[string]interface{}) error
}

// ApplicationState preserves application state across reloads
type ApplicationState struct {
	mu       sync.RWMutex
	data     map[string]interface{}
	sessions map[string]interface{}
	cache    map[string]interface{}
}

// NewApplicationState creates a new application state container
func NewApplicationState() *ApplicationState {
	return &ApplicationState{
		data:     make(map[string]interface{}),
		sessions: make(map[string]interface{}),
		cache:    make(map[string]interface{}),
	}
}

// Get retrieves a value from the state
func (s *ApplicationState) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Set stores a value in the state
func (s *ApplicationState) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// GetAll returns all state data
func (s *ApplicationState) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// SetAll sets all state data
func (s *ApplicationState) SetAll(data map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]interface{})
	for k, v := range data {
		s.data[k] = v
	}
}

// ReloadEvent represents a reload event
type ReloadEvent struct {
	Changes     []FileChange
	Success     bool
	Error       error
	Duration    time.Duration
	ReloadCount int
	Timestamp   time.Time
}

// ReloadManagerOption configures the reload manager
type ReloadManagerOption func(*ReloadManager)

// WithErrorHandler sets the error handler
func WithErrorHandler(handler func(error)) ReloadManagerOption {
	return func(rm *ReloadManager) {
		rm.errorHandler = handler
	}
}

// WithOnReload sets the reload callback
func WithOnReload(handler func(ReloadEvent)) ReloadManagerOption {
	return func(rm *ReloadManager) {
		rm.onReload = handler
	}
}

// NewReloadManager creates a new reload manager
func NewReloadManager(watchPaths []string, compiler CompilerInterface, server ServerInterface, opts ...ReloadManagerOption) *ReloadManager {
	rm := &ReloadManager{
		compiler: compiler,
		server:   server,
		state:    NewApplicationState(),
	}

	for _, opt := range opts {
		opt(rm)
	}

	// Create file watcher with reload callback
	rm.watcher = NewFileWatcher(watchPaths, rm.handleChanges,
		WithDebounce(200*time.Millisecond),
	)

	return rm
}

// Start starts the hot reload manager
func (rm *ReloadManager) Start(ctx context.Context) error {
	log.Println("[HotReload] Starting hot reload manager...")
	return rm.watcher.Start(ctx)
}

// Stop stops the hot reload manager
func (rm *ReloadManager) Stop() {
	log.Println("[HotReload] Stopping hot reload manager...")
	rm.watcher.Stop()
}

// handleChanges handles file changes
func (rm *ReloadManager) handleChanges(changes []FileChange) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	start := time.Now()
	rm.reloadCount++

	// Log changes
	for _, change := range changes {
		log.Printf("[HotReload] File %s: %s", change.Type, change.Path)
	}

	// Find main file to compile
	mainFile := ""
	for _, change := range changes {
		if strings.HasSuffix(change.Path, ".glyph") {
			mainFile = change.Path
			break
		}
	}

	if mainFile == "" {
		log.Println("[HotReload] No compilable file found in changes")
		return
	}

	// Preserve server state
	var savedState map[string]interface{}
	if rm.server != nil {
		savedState = rm.server.GetState()
	}

	// Compile the file
	var bytecode []byte
	var err error
	if rm.compiler != nil {
		bytecode, err = rm.compiler.CompileFile(mainFile)
		if err != nil {
			rm.handleError(fmt.Errorf("compilation failed: %w", err))
			rm.notifyReload(ReloadEvent{
				Changes:     changes,
				Success:     false,
				Error:       err,
				Duration:    time.Since(start),
				ReloadCount: rm.reloadCount,
				Timestamp:   time.Now(),
			})
			return
		}
	}

	// Reload the server
	if rm.server != nil && bytecode != nil {
		if err := rm.server.Reload(bytecode); err != nil {
			rm.handleError(fmt.Errorf("reload failed: %w", err))
			rm.notifyReload(ReloadEvent{
				Changes:     changes,
				Success:     false,
				Error:       err,
				Duration:    time.Since(start),
				ReloadCount: rm.reloadCount,
				Timestamp:   time.Now(),
			})
			return
		}

		// Restore server state
		if savedState != nil {
			if err := rm.server.SetState(savedState); err != nil {
				log.Printf("[HotReload] Warning: failed to restore state: %v", err)
			}
		}
	}

	rm.lastReload = time.Now()
	duration := time.Since(start)

	log.Printf("[HotReload] Reload successful (took %v)", duration)

	rm.notifyReload(ReloadEvent{
		Changes:     changes,
		Success:     true,
		Duration:    duration,
		ReloadCount: rm.reloadCount,
		Timestamp:   time.Now(),
	})
}

// handleError handles errors during reload
func (rm *ReloadManager) handleError(err error) {
	log.Printf("[HotReload] Error: %v", err)
	if rm.errorHandler != nil {
		rm.errorHandler(err)
	}
}

// notifyReload notifies about a reload event
func (rm *ReloadManager) notifyReload(event ReloadEvent) {
	if rm.onReload != nil {
		rm.onReload(event)
	}
}

// Stats returns reload statistics
func (rm *ReloadManager) Stats() ReloadStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return ReloadStats{
		ReloadCount: rm.reloadCount,
		LastReload:  rm.lastReload,
	}
}

// ReloadStats contains reload statistics
type ReloadStats struct {
	ReloadCount int
	LastReload  time.Time
}

// ========================================
// Development Server Integration
// ========================================

// DevServerConfig configures the development server
type DevServerConfig struct {
	WatchPaths   []string
	Port         int
	Host         string
	LiveReload   bool
	OpenBrowser  bool
	BuildOnStart bool
}

// DefaultDevServerConfig returns the default development server configuration
func DefaultDevServerConfig() DevServerConfig {
	return DevServerConfig{
		WatchPaths:   []string{"."},
		Port:         config.DefaultPort,
		Host:         "localhost",
		LiveReload:   true,
		OpenBrowser:  false,
		BuildOnStart: true,
	}
}

// LiveReloadScript returns JavaScript for browser live reload
func LiveReloadScript(wsPort int) string {
	return fmt.Sprintf(`<script>
(function() {
    var ws = new WebSocket('ws://localhost:%d/__livereload');
    ws.onmessage = function(e) {
        if (e.data === 'reload') {
            console.log('[LiveReload] Reloading...');
            window.location.reload();
        }
    };
    ws.onclose = function() {
        console.log('[LiveReload] Connection lost. Retrying...');
        setTimeout(function() { window.location.reload(); }, 1000);
    };
})();
</script>`, wsPort)
}
