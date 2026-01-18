package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  LoggerConfig
		wantErr bool
	}{
		{
			name: "default config",
			config: LoggerConfig{
				MinLevel: INFO,
				Format:   TextFormat,
			},
			wantErr: false,
		},
		{
			name: "json format",
			config: LoggerConfig{
				MinLevel: DEBUG,
				Format:   JSONFormat,
			},
			wantErr: false,
		},
		{
			name: "with buffer size",
			config: LoggerConfig{
				MinLevel:   INFO,
				Format:     TextFormat,
				BufferSize: 500,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if logger != nil {
				defer logger.Close()
			}
		})
	}
}

func TestLoggerTextFormat(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   TextFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("test message")
	logger.Sync() // Wait for async processing

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected log to contain INFO, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected log to contain 'test message', got: %s", output)
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("test message")
	logger.Sync() // Wait for async processing

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	fields := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}

	logger.InfoWithFields("user action", fields)
	logger.Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Fields["user_id"] != float64(123) { // JSON unmarshals numbers as float64
		t.Errorf("Expected user_id 123, got %v", entry.Fields["user_id"])
	}
	if entry.Fields["action"] != "login" {
		t.Errorf("Expected action 'login', got %v", entry.Fields["action"])
	}
}

func TestLoggerLogLevels(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	tests := []struct {
		name     string
		logFunc  func(string)
		expected string
	}{
		{"debug", logger.Debug, "DEBUG"},
		{"info", logger.Info, "INFO"},
		{"warn", logger.Warn, "WARN"},
		{"error", logger.Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test")
			logger.Sync()

			var entry LogEntry
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("Failed to parse JSON log: %v", err)
			}

			if entry.Level != tt.expected {
				t.Errorf("Expected level %s, got %s", tt.expected, entry.Level)
			}
		})
	}
}

func TestLoggerMinLevel(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: WARN,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Sync()

	if buf.Len() > 0 {
		t.Errorf("Expected no logs for DEBUG and INFO when MinLevel is WARN, got: %s", buf.String())
	}

	logger.Warn("warning message")
	logger.Sync()

	if buf.Len() == 0 {
		t.Error("Expected WARN log to be written")
	}
}

func TestContextLogger(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	requestID := "test-request-123"
	ctxLogger := logger.WithRequestID(requestID)

	ctxLogger.Info("context test")
	logger.Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.RequestID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, entry.RequestID)
	}
}

func TestContextLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	fields := map[string]interface{}{
		"service": "api",
		"version": "1.0",
	}

	ctxLogger := logger.WithFields(fields)
	ctxLogger.Info("test")
	logger.Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Fields["service"] != "api" {
		t.Errorf("Expected service 'api', got %v", entry.Fields["service"])
	}
	if entry.Fields["version"] != "1.0" {
		t.Errorf("Expected version '1.0', got %v", entry.Fields["version"])
	}
}

func TestContextLoggerChaining(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	ctxLogger := logger.
		WithRequestID("req-123").
		WithField("user", "john").
		WithField("action", "update")

	ctxLogger.Info("chained context")
	logger.Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.RequestID != "req-123" {
		t.Errorf("Expected request ID 'req-123', got %s", entry.RequestID)
	}
	if entry.Fields["user"] != "john" {
		t.Errorf("Expected user 'john', got %v", entry.Fields["user"])
	}
	if entry.Fields["action"] != "update" {
		t.Errorf("Expected action 'update', got %v", entry.Fields["action"])
	}
}

func TestLoggerWithCaller(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel:      DEBUG,
		Format:        JSONFormat,
		IncludeCaller: true,
		Outputs:       []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("caller test")
	logger.Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Caller == "" {
		t.Error("Expected caller information to be present")
	}
	if !strings.Contains(entry.Caller, "logger_test.go") {
		t.Errorf("Expected caller to contain 'logger_test.go', got %s", entry.Caller)
	}
}

func TestLoggerWithStackTrace(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel:          DEBUG,
		Format:            JSONFormat,
		IncludeStackTrace: true,
		Outputs:           []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Error("error with stack")
	logger.Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.StackTrace == "" {
		t.Error("Expected stack trace to be present for ERROR level")
	}
}

func TestLogFileRotation(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogger(LoggerConfig{
		MinLevel:    INFO,
		Format:      TextFormat,
		FilePath:    logPath,
		MaxFileSize: 100, // Very small size to trigger rotation
		MaxBackups:  3,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write enough logs to trigger rotation
	for i := 0; i < 20; i++ {
		logger.Info("This is a test message that should trigger rotation")
		time.Sleep(10 * time.Millisecond)
	}

	logger.Close()

	// Check that rotation happened
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("Main log file should exist: %v", err)
	}

	rotatedPath := logPath + ".1"
	if _, err := os.Stat(rotatedPath); err != nil {
		t.Errorf("Rotated log file should exist: %v", err)
	}
}

func TestNewRequestID(t *testing.T) {
	id1 := NewRequestID()
	id2 := NewRequestID()

	if id1 == "" {
		t.Error("Request ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Request IDs should be unique")
	}

	// UUID format check (basic)
	if len(id1) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(id1))
	}
}

func TestDefaultLogger(t *testing.T) {
	// Reset default logger
	defaultLoggerMu.Lock()
	if defaultLogger != nil {
		defaultLogger.Close()
		defaultLogger = nil
	}
	defaultLoggerMu.Unlock()

	logger := GetDefaultLogger()
	if logger == nil {
		t.Fatal("Default logger should not be nil")
	}
	defer logger.Close()

	// Test that we get the same instance
	logger2 := GetDefaultLogger()
	if logger != logger2 {
		t.Error("Should return the same default logger instance")
	}
}

func TestInitDefaultLogger(t *testing.T) {
	var buf bytes.Buffer

	err := InitDefaultLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to init default logger: %v", err)
	}
	defer GetDefaultLogger().Close()

	Info("test default logger")
	GetDefaultLogger().Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Message != "test default logger" {
		t.Errorf("Expected message 'test default logger', got %s", entry.Message)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	var buf bytes.Buffer

	err := InitDefaultLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to init default logger: %v", err)
	}
	defer GetDefaultLogger().Close()

	tests := []struct {
		name     string
		logFunc  func(string)
		expected string
	}{
		{"debug", Debug, "DEBUG"},
		{"info", Info, "INFO"},
		{"warn", Warn, "WARN"},
		{"error", Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test")
			GetDefaultLogger().Sync()

			var entry LogEntry
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("Failed to parse JSON log: %v", err)
			}

			if entry.Level != tt.expected {
				t.Errorf("Expected level %s, got %s", tt.expected, entry.Level)
			}
		})
	}
}

func TestWithRequestIDConvenience(t *testing.T) {
	var buf bytes.Buffer

	err := InitDefaultLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to init default logger: %v", err)
	}
	defer GetDefaultLogger().Close()

	requestID := "convenience-test-123"
	ctxLogger := WithRequestID(requestID)
	ctxLogger.Info("test")
	GetDefaultLogger().Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.RequestID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, entry.RequestID)
	}
}

func TestWithFieldsConvenience(t *testing.T) {
	var buf bytes.Buffer

	err := InitDefaultLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to init default logger: %v", err)
	}
	defer GetDefaultLogger().Close()

	fields := map[string]interface{}{
		"test": "value",
	}

	ctxLogger := WithFields(fields)
	ctxLogger.Info("test")
	GetDefaultLogger().Sync()

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if entry.Fields["test"] != "value" {
		t.Errorf("Expected field test='value', got %v", entry.Fields["test"])
	}
}

func TestLoggerConcurrency(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel:   DEBUG,
		Format:     JSONFormat,
		Outputs:    []io.Writer{&buf},
		BufferSize: 1000,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test concurrent logging
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				logger.InfoWithFields("concurrent test", map[string]interface{}{
					"goroutine": id,
					"iteration": j,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	logger.Sync()

	// Count log entries
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 100 {
		t.Errorf("Expected 100 log entries, got %d", len(lines))
	}
}

func TestContextLoggerConcurrency(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
		Outputs:  []io.Writer{&buf},
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	baseCtx := logger.WithRequestID("base-req")

	// Test concurrent field additions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			ctx := baseCtx.WithField("goroutine", id)
			ctx.Info("concurrent context test")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	logger.Sync()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 log entries, got %d", len(lines))
	}
}

func BenchmarkLoggerInfo(b *testing.B) {
	logger, _ := NewLogger(LoggerConfig{
		MinLevel:   INFO,
		Format:     TextFormat,
		Outputs:    []io.Writer{io.Discard},
		BufferSize: 10000,
	})
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

func BenchmarkLoggerInfoWithFields(b *testing.B) {
	logger, _ := NewLogger(LoggerConfig{
		MinLevel:   INFO,
		Format:     JSONFormat,
		Outputs:    []io.Writer{io.Discard},
		BufferSize: 10000,
	})
	defer logger.Close()

	fields := map[string]interface{}{
		"user_id": 123,
		"action":  "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoWithFields("benchmark message", fields)
	}
}

func BenchmarkContextLogger(b *testing.B) {
	logger, _ := NewLogger(LoggerConfig{
		MinLevel:   INFO,
		Format:     JSONFormat,
		Outputs:    []io.Writer{io.Discard},
		BufferSize: 10000,
	})
	defer logger.Close()

	ctx := logger.WithRequestID("bench-req").WithField("service", "api")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Info("benchmark message")
	}
}
