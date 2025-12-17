package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestHelper provides utilities for GLYPH tests
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// LoadFixture loads a test fixture file
func (h *TestHelper) LoadFixture(name string) string {
	path := filepath.Join("fixtures", name)
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	return string(content)
}

// WriteFixture writes content to a test fixture file (for test generation)
func (h *TestHelper) WriteFixture(name, content string) {
	path := filepath.Join("fixtures", name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		h.t.Fatalf("Failed to write fixture %s: %v", name, err)
	}
}

// AssertEqual checks if two values are equal
func (h *TestHelper) AssertEqual(got, want interface{}, msg string) {
	if got != want {
		h.t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertNotNil checks if value is not nil
func (h *TestHelper) AssertNotNil(val interface{}, msg string) {
	if val == nil {
		h.t.Errorf("%s: expected non-nil value", msg)
	}
}

// AssertNil checks if value is nil
func (h *TestHelper) AssertNil(val interface{}, msg string) {
	if val != nil {
		h.t.Errorf("%s: expected nil, got %v", msg, val)
	}
}

// AssertNoError checks if error is nil
func (h *TestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.t.Fatalf("%s: unexpected error: %v", msg, err)
	}
}

// AssertError checks if error is not nil
func (h *TestHelper) AssertError(err error, msg string) {
	if err == nil {
		h.t.Errorf("%s: expected error, got nil", msg)
	}
}

// AssertContains checks if string contains substring
func (h *TestHelper) AssertContains(str, substr, msg string) {
	if !contains(str, substr) {
		h.t.Errorf("%s: %q does not contain %q", msg, str, substr)
	}
}

// MockServer represents a test HTTP server
type MockServer struct {
	Server *httptest.Server
	URL    string
}

// NewMockServer creates a test HTTP server
func NewMockServer(handler http.Handler) *MockServer {
	server := httptest.NewServer(handler)
	return &MockServer{
		Server: server,
		URL:    server.URL,
	}
}

// Close shuts down the mock server
func (m *MockServer) Close() {
	m.Server.Close()
}

// HTTPRequest represents a test HTTP request
type HTTPRequest struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// HTTPResponse represents a test HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       string
	JSON       map[string]interface{}
	Headers    http.Header
}

// MakeRequest makes an HTTP request to a test server
func MakeRequest(t *testing.T, serverURL string, req HTTPRequest) *HTTPResponse {
	var bodyReader io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	method := req.Method
	if method == "" {
		method = "GET"
	}

	httpReq, err := http.NewRequest(method, serverURL+req.Path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Make request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	response := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
		Headers:    resp.Header,
	}

	// Try to parse as JSON
	if resp.Header.Get("Content-Type") == "application/json" {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			response.JSON = jsonData
		}
	}

	return response
}

// CompilerMock represents a mock compiler for testing
type CompilerMock struct {
	CompileFunc     func(source string) ([]byte, error)
	CompileFileFunc func(filename string) ([]byte, error)
}

// Compile mocks the compilation process
func (m *CompilerMock) Compile(source string) ([]byte, error) {
	if m.CompileFunc != nil {
		return m.CompileFunc(source)
	}
	// Default mock implementation
	return []byte{0x41, 0x49, 0x42, 0x43, 0x01, 0x00, 0x00, 0x00}, nil
}

// CompileFile mocks file compilation
func (m *CompilerMock) CompileFile(filename string) ([]byte, error) {
	if m.CompileFileFunc != nil {
		return m.CompileFileFunc(filename)
	}
	return nil, fmt.Errorf("not implemented")
}

// ParseResult represents parsed GLYPH code for testing
type ParseResult struct {
	Success bool
	AST     interface{}
	Error   string
}

// InterpreterMock represents a mock interpreter for testing
type InterpreterMock struct {
	ExecuteFunc func(ast interface{}) (interface{}, error)
}

// Execute mocks AST execution
func (m *InterpreterMock) Execute(ast interface{}) (interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ast)
	}
	return map[string]interface{}{"status": "ok"}, nil
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TempFile creates a temporary file for testing
func TempFile(t *testing.T, name, content string) string {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return path
}

// AssertJSONField checks if JSON response contains expected field value
func AssertJSONField(t *testing.T, json map[string]interface{}, field string, expected interface{}) {
	value, ok := json[field]
	if !ok {
		t.Errorf("JSON missing field %q", field)
		return
	}
	if value != expected {
		t.Errorf("JSON field %q: got %v, want %v", field, value, expected)
	}
}

// RetryWithTimeout retries a function until it succeeds or timeout
func RetryWithTimeout(t *testing.T, timeout time.Duration, fn func() error) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf("timeout after %v: %v", timeout, lastErr)
}
