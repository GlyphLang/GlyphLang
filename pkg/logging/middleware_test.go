package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

func TestStructuredLoggingMiddleware(t *testing.T) {
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

	middleware := StructuredLoggingMiddleware(logger)

	handler := middleware(func(ctx *server.Context) error {
		ctx.StatusCode = 200
		return nil
	})

	req := httptest.NewRequest("GET", "/test?query=param", nil)
	w := httptest.NewRecorder()

	ctx := &server.Context{
		Request:        req,
		ResponseWriter: w,
		PathParams:     make(map[string]string),
		QueryParams:    make(map[string][]string),
		StatusCode:     0,
	}

	err = handler(ctx)
	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Check that request ID header was set
	requestID := w.Header().Get(RequestIDHeader)
	if requestID == "" {
		t.Error("Expected request ID header to be set")
	}

	// Parse log entries
	logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries (start and complete), got %d", len(logs))
	}

	// Check first log (request started)
	var startEntry LogEntry
	if err := json.Unmarshal([]byte(logs[0]), &startEntry); err != nil {
		t.Fatalf("Failed to parse start log entry: %v", err)
	}

	if startEntry.Message != "request started" {
		t.Errorf("Expected message 'request started', got %s", startEntry.Message)
	}
	if startEntry.Fields["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", startEntry.Fields["method"])
	}
	if startEntry.Fields["path"] != "/test" {
		t.Errorf("Expected path /test, got %v", startEntry.Fields["path"])
	}

	// Check second log (request completed)
	var completeEntry LogEntry
	if err := json.Unmarshal([]byte(logs[1]), &completeEntry); err != nil {
		t.Fatalf("Failed to parse complete log entry: %v", err)
	}

	if completeEntry.Message != "request completed" {
		t.Errorf("Expected message 'request completed', got %s", completeEntry.Message)
	}
	if completeEntry.Fields["status"] != float64(200) {
		t.Errorf("Expected status 200, got %v", completeEntry.Fields["status"])
	}
	if completeEntry.Fields["duration_ms"] == nil {
		t.Error("Expected duration_ms to be present")
	}
}

func TestStructuredLoggingMiddlewareWithExistingRequestID(t *testing.T) {
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

	middleware := StructuredLoggingMiddleware(logger)

	handler := middleware(func(ctx *server.Context) error {
		ctx.StatusCode = 200
		return nil
	})

	existingRequestID := "existing-request-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingRequestID)
	w := httptest.NewRecorder()

	ctx := &server.Context{
		Request:        req,
		ResponseWriter: w,
		PathParams:     make(map[string]string),
		QueryParams:    make(map[string][]string),
		StatusCode:     0,
	}

	handler(ctx)
	time.Sleep(50 * time.Millisecond)

	// Check that existing request ID was preserved
	if w.Header().Get(RequestIDHeader) != existingRequestID {
		t.Errorf("Expected request ID %s, got %s", existingRequestID, w.Header().Get(RequestIDHeader))
	}

	// Parse log entries
	logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var entry LogEntry
	if err := json.Unmarshal([]byte(logs[0]), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.RequestID != existingRequestID {
		t.Errorf("Expected request ID %s in logs, got %s", existingRequestID, entry.RequestID)
	}
}

func TestStructuredLoggingMiddlewareError(t *testing.T) {
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

	middleware := StructuredLoggingMiddleware(logger)

	testError := errors.New("test error")
	handler := middleware(func(ctx *server.Context) error {
		return testError
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	ctx := &server.Context{
		Request:        req,
		ResponseWriter: w,
		PathParams:     make(map[string]string),
		QueryParams:    make(map[string][]string),
		StatusCode:     0,
	}

	err = handler(ctx)
	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}

	time.Sleep(50 * time.Millisecond)

	// Parse log entries
	logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var completeEntry LogEntry
	if err := json.Unmarshal([]byte(logs[1]), &completeEntry); err != nil {
		t.Fatalf("Failed to parse complete log entry: %v", err)
	}

	if completeEntry.Level != "ERROR" {
		t.Errorf("Expected ERROR level for failed request, got %s", completeEntry.Level)
	}
	if completeEntry.Message != "request failed" {
		t.Errorf("Expected message 'request failed', got %s", completeEntry.Message)
	}
	if completeEntry.Fields["error"] != "test error" {
		t.Errorf("Expected error field 'test error', got %v", completeEntry.Fields["error"])
	}
}

func TestStructuredLoggingMiddlewareStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedLevel  string
		expectedMsg    string
	}{
		{
			name:          "success",
			statusCode:    200,
			expectedLevel: "INFO",
			expectedMsg:   "request completed",
		},
		{
			name:          "client error",
			statusCode:    404,
			expectedLevel: "WARN",
			expectedMsg:   "request completed with client error",
		},
		{
			name:          "server error",
			statusCode:    500,
			expectedLevel: "ERROR",
			expectedMsg:   "request completed with server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			middleware := StructuredLoggingMiddleware(logger)

			handler := middleware(func(ctx *server.Context) error {
				ctx.StatusCode = tt.statusCode
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			ctx := &server.Context{
				Request:        req,
				ResponseWriter: w,
				PathParams:     make(map[string]string),
				QueryParams:    make(map[string][]string),
				StatusCode:     0,
			}

			handler(ctx)
			time.Sleep(50 * time.Millisecond)

			// Parse completion log entry
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			var completeEntry LogEntry
			if err := json.Unmarshal([]byte(logs[1]), &completeEntry); err != nil {
				t.Fatalf("Failed to parse complete log entry: %v", err)
			}

			if completeEntry.Level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %s", tt.expectedLevel, completeEntry.Level)
			}
			if completeEntry.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, completeEntry.Message)
			}
		})
	}
}

func TestStructuredLoggingMiddlewareWithBodyLogging(t *testing.T) {
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

	middleware := StructuredLoggingMiddlewareWithBodyLogging(logger, 1024)

	handler := middleware(func(ctx *server.Context) error {
		ctx.StatusCode = 200
		ctx.ResponseWriter.Write([]byte(`{"result":"success"}`))
		return nil
	})

	requestBody := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := &server.Context{
		Request:        req,
		ResponseWriter: w,
		PathParams:     make(map[string]string),
		QueryParams:    make(map[string][]string),
		StatusCode:     0,
	}

	handler(ctx)
	time.Sleep(50 * time.Millisecond)

	// Parse log entries
	logs := strings.Split(strings.TrimSpace(buf.String()), "\n")

	// Check request log
	var startEntry LogEntry
	if err := json.Unmarshal([]byte(logs[0]), &startEntry); err != nil {
		t.Fatalf("Failed to parse start log entry: %v", err)
	}

	if startEntry.Fields["request_body"] == nil {
		t.Error("Expected request_body to be logged")
	}

	// Check response log
	var completeEntry LogEntry
	if err := json.Unmarshal([]byte(logs[1]), &completeEntry); err != nil {
		t.Fatalf("Failed to parse complete log entry: %v", err)
	}

	if completeEntry.Fields["response_body"] == nil {
		t.Error("Expected response_body to be logged")
	}
}

func TestStructuredRecoveryMiddleware(t *testing.T) {
	var buf bytes.Buffer

	logger, err := NewLogger(LoggerConfig{
		MinLevel:          DEBUG,
		Format:            JSONFormat,
		Outputs:           []io.Writer{&buf},
		IncludeStackTrace: true,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	middleware := StructuredRecoveryMiddleware(logger)

	handler := middleware(func(ctx *server.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, "panic-test-123")
	w := httptest.NewRecorder()

	ctx := &server.Context{
		Request:        req,
		ResponseWriter: w,
		PathParams:     make(map[string]string),
		QueryParams:    make(map[string][]string),
		StatusCode:     0,
	}

	err = handler(ctx)
	if err == nil {
		t.Error("Expected error to be returned after panic recovery")
	} else {
		t.Logf("Error returned: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Parse log entry
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("Expected ERROR level, got %s", entry.Level)
	}
	if entry.Message != "panic recovered" {
		t.Errorf("Expected message 'panic recovered', got %s", entry.Message)
	}
	if entry.Fields["panic"] != "test panic" {
		t.Errorf("Expected panic field 'test panic', got %v", entry.Fields["panic"])
	}
	if entry.RequestID != "panic-test-123" {
		t.Errorf("Expected request ID 'panic-test-123', got %s", entry.RequestID)
	}
}

func TestGetRequestLogger(t *testing.T) {
	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, "get-logger-test")

	ctxLogger := GetRequestLogger(logger, req)

	if ctxLogger.requestID != "get-logger-test" {
		t.Errorf("Expected request ID 'get-logger-test', got %s", ctxLogger.requestID)
	}

	if ctxLogger.fields["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", ctxLogger.fields["method"])
	}

	if ctxLogger.fields["path"] != "/test" {
		t.Errorf("Expected path /test, got %v", ctxLogger.fields["path"])
	}
}

func TestGetRequestLoggerWithoutRequestID(t *testing.T) {
	logger, err := NewLogger(LoggerConfig{
		MinLevel: DEBUG,
		Format:   JSONFormat,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	req := httptest.NewRequest("POST", "/api/test", nil)

	ctxLogger := GetRequestLogger(logger, req)

	if ctxLogger.requestID == "" {
		t.Error("Expected request ID to be generated")
	}

	if ctxLogger.fields["method"] != "POST" {
		t.Errorf("Expected method POST, got %v", ctxLogger.fields["method"])
	}

	if ctxLogger.fields["path"] != "/api/test" {
		t.Errorf("Expected path /api/test, got %v", ctxLogger.fields["path"])
	}
}

func BenchmarkStructuredLoggingMiddleware(b *testing.B) {
	logger, _ := NewLogger(LoggerConfig{
		MinLevel:   INFO,
		Format:     JSONFormat,
		Outputs:    []io.Writer{io.Discard},
		BufferSize: 10000,
	})
	defer logger.Close()

	middleware := StructuredLoggingMiddleware(logger)

	handler := middleware(func(ctx *server.Context) error {
		ctx.StatusCode = 200
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ctx := &server.Context{
			Request:        req,
			ResponseWriter: w,
			PathParams:     make(map[string]string),
			QueryParams:    make(map[string][]string),
			StatusCode:     0,
		}
		handler(ctx)
	}
}
