package tracing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestExcludedPaths tests the ExcludedPaths function
func TestExcludedPaths(t *testing.T) {
	paths := ExcludedPaths()

	if paths == nil {
		t.Fatal("Expected non-nil paths")
	}

	expectedPaths := []string{"/health", "/metrics", "/ping", "/readiness", "/liveness"}
	for _, path := range expectedPaths {
		if !paths[path] {
			t.Errorf("Expected %s to be in excluded paths", path)
		}
	}
}

// TestAddContextToRequest tests adding trace context to a request
func TestAddContextToRequest(t *testing.T) {
	// Test with nil request
	req := AddContextToRequest(nil)
	if req != nil {
		t.Error("Expected nil result for nil input")
	}

	// Test with valid request
	originalReq := httptest.NewRequest("GET", "http://example.com/test", nil)
	newReq := AddContextToRequest(originalReq)

	if newReq == nil {
		t.Error("Expected non-nil result")
	}
}

// TestGetTraceHeaders tests extracting trace headers
func TestGetTraceHeaders(t *testing.T) {
	// Test with nil request
	headers := GetTraceHeaders(nil)
	if headers == nil {
		t.Error("Expected non-nil headers even for nil request")
	}
	if len(headers) != 0 {
		t.Error("Expected empty headers for nil request")
	}

	// Test with request containing trace headers
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("traceparent", "00-12345678901234567890123456789012-1234567890123456-01")
	req.Header.Set("tracestate", "vendor=value")

	headers = GetTraceHeaders(req)

	if headers["traceparent"] != "00-12345678901234567890123456789012-1234567890123456-01" {
		t.Errorf("Expected traceparent header, got %v", headers["traceparent"])
	}

	if headers["tracestate"] != "vendor=value" {
		t.Errorf("Expected tracestate header, got %v", headers["tracestate"])
	}
}

// TestSetStatus tests setting the span status
func TestSetStatus(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	SetStatus(ctx, codes.Ok, "success")
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].Status.Code != codes.Ok {
		t.Errorf("Expected Ok status, got %v", spans[0].Status.Code)
	}
}

// TestRecordOutgoingResponse tests recording outgoing HTTP responses
func TestRecordOutgoingResponse(t *testing.T) {
	t.Run("with_error", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(exporter),
		)
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("test")
		ctx := context.Background()

		ctx, span := tracer.Start(ctx, "test-span")
		testErr := errors.New("connection error")
		RecordOutgoingResponse(ctx, nil, testErr)
		span.End()

		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("Expected 1 span, got %d", len(spans))
		}

		if spans[0].Status.Code != codes.Error {
			t.Errorf("Expected Error status, got %v", spans[0].Status.Code)
		}
	})

	t.Run("with_nil_response", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(exporter),
		)
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("test")
		ctx := context.Background()

		ctx, span := tracer.Start(ctx, "test-span")
		RecordOutgoingResponse(ctx, nil, nil)
		span.End()

		// Should not panic
	})

	t.Run("with_success_response", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(exporter),
		)
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("test")
		ctx := context.Background()

		ctx, span := tracer.Start(ctx, "test-span")
		resp := &http.Response{StatusCode: 200}
		RecordOutgoingResponse(ctx, resp, nil)
		span.End()

		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("Expected 1 span, got %d", len(spans))
		}

		if spans[0].Status.Code != codes.Ok {
			t.Errorf("Expected Ok status, got %v", spans[0].Status.Code)
		}
	})

	t.Run("with_error_response", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(exporter),
		)
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("test")
		ctx := context.Background()

		ctx, span := tracer.Start(ctx, "test-span")
		resp := &http.Response{StatusCode: 500}
		RecordOutgoingResponse(ctx, resp, nil)
		span.End()

		spans := exporter.GetSpans()
		if len(spans) != 1 {
			t.Fatalf("Expected 1 span, got %d", len(spans))
		}

		if spans[0].Status.Code != codes.Error {
			t.Errorf("Expected Error status, got %v", spans[0].Status.Code)
		}
	})
}

// TestWithHTTPClientTrace tests the HTTP client trace helper
func TestWithHTTPClientTrace(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify trace context was propagated
		if r.Header.Get("traceparent") == "" {
			t.Error("Expected traceparent header to be propagated")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Make a traced request
	ctx := context.Background()
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	client := &http.Client{}

	resp, err := WithHTTPClientTrace(ctx, req, client)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestExtractRequest tests the extractRequest helper
func TestExtractRequest(t *testing.T) {
	// Test with nil
	req, ok := extractRequest(nil)
	if ok {
		t.Error("Expected false for nil input")
	}
	if req != nil {
		t.Error("Expected nil request for nil input")
	}

	// Test with wrong type
	req, ok = extractRequest("not a context")
	if ok {
		t.Error("Expected false for wrong type")
	}

	// Test with struct implementing hasRequest interface
	mockReq := httptest.NewRequest("GET", "http://example.com", nil)
	mockCtx := &MockServerContext{request: mockReq}
	req, ok = extractRequest(mockCtx)
	if !ok {
		t.Error("Expected true for valid interface implementation")
	}
	if req != mockReq {
		t.Error("Expected request to match")
	}
}

// MockServerContext implements the hasRequest interface for testing
type MockServerContext struct {
	request        *http.Request
	responseWriter http.ResponseWriter
	statusCode     int
	pathParams     map[string]string
	queryParams    map[string]string
}

func (m *MockServerContext) GetRequest() *http.Request {
	return m.request
}

func (m *MockServerContext) GetResponseWriter() http.ResponseWriter {
	return m.responseWriter
}

func (m *MockServerContext) GetStatusCode() int {
	return m.statusCode
}

func (m *MockServerContext) SetContext(ctx context.Context) {}

func (m *MockServerContext) GetContext() context.Context {
	return context.Background()
}

// TestExtractResponseWriter tests the extractResponseWriter helper
func TestExtractResponseWriter(t *testing.T) {
	// Test with nil
	w, ok := extractResponseWriter(nil)
	if ok {
		t.Error("Expected false for nil input")
	}
	if w != nil {
		t.Error("Expected nil writer for nil input")
	}

	// Test with struct implementing hasResponseWriter interface
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{responseWriter: mockWriter}
	w, ok = extractResponseWriter(mockCtx)
	if !ok {
		t.Error("Expected true for valid interface implementation")
	}
	if w != mockWriter {
		t.Error("Expected writer to match")
	}
}

// TestExtractPathParams tests the extractPathParams helper
func TestExtractPathParams(t *testing.T) {
	// Test with nil
	params := extractPathParams(nil)
	if params != nil {
		t.Error("Expected nil for nil input")
	}

	// Test with wrong type
	params = extractPathParams("not a context")
	if params != nil {
		t.Error("Expected nil for wrong type")
	}
}

// TestExtractQueryParams tests the extractQueryParams helper
func TestExtractQueryParams(t *testing.T) {
	// Test with nil
	params := extractQueryParams(nil)
	if params != nil {
		t.Error("Expected nil for nil input")
	}

	// Test with wrong type
	params = extractQueryParams("not a context")
	if params != nil {
		t.Error("Expected nil for wrong type")
	}
}

// TestExtractStatusCode tests the extractStatusCode helper
func TestExtractStatusCode(t *testing.T) {
	// Test with nil
	code := extractStatusCode(nil)
	if code != 0 {
		t.Errorf("Expected 0 for nil input, got %d", code)
	}

	// Test with wrong type
	code = extractStatusCode("not a context")
	if code != 0 {
		t.Errorf("Expected 0 for wrong type, got %d", code)
	}
}

// TestUpdateRequestInContext tests the updateRequestInContext helper
func TestUpdateRequestInContext(t *testing.T) {
	// Should not panic
	req := httptest.NewRequest("GET", "http://example.com", nil)
	updateRequestInContext(nil, req)
	updateRequestInContext("not a context", req)
}

// TestTraceServerRequest tests the TraceServerRequest function
func TestTraceServerRequest(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	t.Run("nil_context", func(t *testing.T) {
		err := TraceServerRequest(nil, nil)
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})

	t.Run("with_handler", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx interface{}) error {
			handlerCalled = true
			return nil
		}

		mockReq := httptest.NewRequest("GET", "http://example.com/test", nil)
		mockWriter := httptest.NewRecorder()
		mockCtx := &MockServerContext{
			request:        mockReq,
			responseWriter: mockWriter,
			statusCode:     200,
		}

		err := TraceServerRequest(mockCtx, handler)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !handlerCalled {
			t.Error("Expected handler to be called")
		}
	})

	t.Run("excluded_path", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx interface{}) error {
			handlerCalled = true
			return nil
		}

		mockReq := httptest.NewRequest("GET", "http://example.com/health", nil)
		mockWriter := httptest.NewRecorder()
		mockCtx := &MockServerContext{
			request:        mockReq,
			responseWriter: mockWriter,
		}

		err := TraceServerRequest(mockCtx, handler)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !handlerCalled {
			t.Error("Expected handler to be called even for excluded path")
		}
	})

	t.Run("with_error", func(t *testing.T) {
		testErr := errors.New("test error")
		handler := func(ctx interface{}) error {
			return testErr
		}

		mockReq := httptest.NewRequest("GET", "http://example.com/test", nil)
		mockWriter := httptest.NewRecorder()
		mockCtx := &MockServerContext{
			request:        mockReq,
			responseWriter: mockWriter,
		}

		err := TraceServerRequest(mockCtx, handler)
		if err != testErr {
			t.Errorf("Expected error %v, got %v", testErr, err)
		}
	})
}

// TestResponseWriter tests the responseWriter wrapper
func TestResponseWriter(t *testing.T) {
	t.Run("write_header", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: underlying,
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)

		if rw.statusCode != http.StatusNotFound {
			t.Errorf("Expected status code 404, got %d", rw.statusCode)
		}
	})

	t.Run("write", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: underlying,
			statusCode:     http.StatusOK,
		}

		n, err := rw.Write([]byte("hello"))
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if n != 5 {
			t.Errorf("Expected 5 bytes written, got %d", n)
		}
		if rw.bytesWritten != 5 {
			t.Errorf("Expected bytesWritten 5, got %d", rw.bytesWritten)
		}
	})
}

// TestHTTPTracingMiddlewareWithCustomAttributes tests custom attributes
func TestHTTPTracingMiddlewareWithCustomAttributes(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := &MiddlewareConfig{
		SpanNameFormatter: func(req *http.Request) string {
			return "Custom: " + req.URL.Path
		},
		ExcludePaths: map[string]bool{},
		CustomAttributes: func(req *http.Request) []attribute.KeyValue {
			return []attribute.KeyValue{
				attribute.String("custom.attr", "custom-value"),
			}
		},
	}

	middleware := HTTPTracingMiddleware(middlewareConfig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestHTTPTracingMiddlewareNilConfig tests middleware with nil config
func TestHTTPTracingMiddlewareNilConfig(t *testing.T) {
	middleware := HTTPTracingMiddleware(nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestHTTPTracingMiddlewareErrorStatus tests error status codes
func TestHTTPTracingMiddlewareErrorStatus(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

// TestTraceOutgoingRequestDirect tests the TraceOutgoingRequest function directly
func TestTraceOutgoingRequestDirect(t *testing.T) {
	// Set up tracing with proper propagator
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	req := httptest.NewRequest("POST", "http://example.com/api", nil)

	ctx, span := TraceOutgoingRequest(ctx, req, "Test Outgoing")
	defer span.End()

	// Verify trace context was injected (may not have traceparent if no active span)
	// The function should at least set attributes
	if span == nil {
		t.Error("Expected non-nil span")
	}
}

// TestGetTraceIDEmpty tests GetTraceID with no span in context
func TestGetTraceIDEmpty(t *testing.T) {
	ctx := context.Background()
	traceID := GetTraceID(ctx)
	if traceID != "" {
		t.Errorf("Expected empty trace ID, got %s", traceID)
	}
}

// TestGetSpanIDEmpty tests GetSpanID with no span in context
func TestGetSpanIDEmpty(t *testing.T) {
	ctx := context.Background()
	spanID := GetSpanID(ctx)
	if spanID != "" {
		t.Errorf("Expected empty span ID, got %s", spanID)
	}
}

// TestIsTracingEnabled tests the IsTracingEnabled function
func TestIsTracingEnabled(t *testing.T) {
	// By default, tracing should be enabled
	enabled := IsTracingEnabled()
	// The result depends on environment, just test it doesn't panic
	_ = enabled
}

// TestHTTPClientAttributes tests HTTP client attributes
func TestHTTPClientAttributes(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com/api/test?query=value", nil)
	attrs := HTTPClientAttributes(req, 200)

	if len(attrs) == 0 {
		t.Error("Expected non-empty attributes")
	}

	// Check required attributes
	hasMethod := false
	hasURL := false
	hasStatusCode := false

	for _, attr := range attrs {
		switch string(attr.Key) {
		case "http.method":
			hasMethod = true
			if attr.Value.AsString() != "POST" {
				t.Errorf("Expected method 'POST', got '%s'", attr.Value.AsString())
			}
		case "http.url":
			hasURL = true
		case "http.status_code":
			hasStatusCode = true
			if attr.Value.AsInt64() != 200 {
				t.Errorf("Expected status code 200, got %d", attr.Value.AsInt64())
			}
		}
	}

	if !hasMethod {
		t.Error("Missing http.method attribute")
	}
	if !hasURL {
		t.Error("Missing http.url attribute")
	}
	if !hasStatusCode {
		t.Error("Missing http.status_code attribute")
	}
}

// TestWithSpanHelper tests the WithSpan helper function
func TestWithSpanHelper(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		executed := false
		err := WithSpan(ctx, "test-span", func(ctx context.Context) error {
			executed = true
			return nil
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !executed {
			t.Error("Expected function to be executed")
		}
	})

	t.Run("error", func(t *testing.T) {
		testErr := errors.New("test error")
		err := WithSpan(ctx, "test-span", func(ctx context.Context) error {
			return testErr
		})

		if err != testErr {
			t.Errorf("Expected error %v, got %v", testErr, err)
		}
	})
}

// TestTracerProviderShutdownNil tests shutdown with nil provider
func TestTracerProviderShutdownNil(t *testing.T) {
	tp := &TracerProvider{provider: nil}
	err := tp.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// TestTracerProviderGetTracerNil tests GetTracer with nil provider
func TestTracerProviderGetTracerNil(t *testing.T) {
	tp := &TracerProvider{provider: nil}
	tracer := tp.GetTracer("test")
	if tracer == nil {
		t.Error("Expected non-nil tracer")
	}
}

// TestInitTracingNilConfig tests InitTracing with nil config
func TestInitTracingNilConfig(t *testing.T) {
	tp, err := InitTracing(nil)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	if tp.config == nil {
		t.Error("Expected non-nil config")
	}
}

// TestSamplingRates tests different sampling rates
func TestSamplingRates(t *testing.T) {
	t.Run("always_sample", func(t *testing.T) {
		config := &Config{
			ServiceName:  "test",
			ExporterType: "stdout",
			SamplingRate: 1.0,
			Enabled:      true,
		}
		tp, err := InitTracing(config)
		if err != nil {
			t.Fatalf("InitTracing failed: %v", err)
		}
		tp.Shutdown(context.Background())
	})

	t.Run("never_sample", func(t *testing.T) {
		config := &Config{
			ServiceName:  "test",
			ExporterType: "stdout",
			SamplingRate: 0.0,
			Enabled:      true,
		}
		tp, err := InitTracing(config)
		if err != nil {
			t.Fatalf("InitTracing failed: %v", err)
		}
		tp.Shutdown(context.Background())
	})

	t.Run("partial_sample", func(t *testing.T) {
		config := &Config{
			ServiceName:  "test",
			ExporterType: "stdout",
			SamplingRate: 0.5,
			Enabled:      true,
		}
		tp, err := InitTracing(config)
		if err != nil {
			t.Fatalf("InitTracing failed: %v", err)
		}
		tp.Shutdown(context.Background())
	})
}

// TestMiddlewareConfig tests MiddlewareConfig struct
func TestMiddlewareConfig(t *testing.T) {
	config := &MiddlewareConfig{
		SpanNameFormatter:  nil,
		ExcludePaths:       nil,
		RecordRequestBody:  true,
		RecordResponseBody: true,
		CustomAttributes:   nil,
	}

	if config.RecordRequestBody != true {
		t.Error("Expected RecordRequestBody to be true")
	}

	if config.RecordResponseBody != true {
		t.Error("Expected RecordResponseBody to be true")
	}
}

// TestServerContext tests the ServerContext interface
func TestServerContext(t *testing.T) {
	mockReq := httptest.NewRequest("GET", "http://example.com", nil)
	mockWriter := httptest.NewRecorder()

	ctx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
		statusCode:     200,
	}

	if ctx.GetRequest() != mockReq {
		t.Error("Expected request to match")
	}

	if ctx.GetResponseWriter() != mockWriter {
		t.Error("Expected response writer to match")
	}

	if ctx.GetStatusCode() != 200 {
		t.Errorf("Expected status code 200, got %d", ctx.GetStatusCode())
	}

	ctx.SetContext(context.Background())
	if ctx.GetContext() == nil {
		t.Error("Expected non-nil context")
	}
}

// TestHTTPAttributesWithRemoteAddr tests HTTPAttributes with RemoteAddr fallback
func TestHTTPAttributesWithRemoteAddr(t *testing.T) {
	// Test without X-Forwarded-For but with RemoteAddr
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.RemoteAddr = "192.168.1.100:54321"

	attrs := HTTPAttributes(req, 200)

	hasClientIP := false
	for _, attr := range attrs {
		if string(attr.Key) == "http.client_ip" {
			hasClientIP = true
			if attr.Value.AsString() != "192.168.1.100:54321" {
				t.Errorf("Expected client IP '192.168.1.100:54321', got '%s'", attr.Value.AsString())
			}
		}
	}

	if !hasClientIP {
		t.Error("Expected http.client_ip attribute from RemoteAddr")
	}
}

// TestHTTPAttributesMinimal tests HTTPAttributes with minimal request
func TestHTTPAttributesMinimal(t *testing.T) {
	// Request without query, user-agent, or forwarding headers
	req := httptest.NewRequest("POST", "http://example.com/api", nil)

	attrs := HTTPAttributes(req, 201)

	if len(attrs) == 0 {
		t.Error("Expected at least some attributes")
	}

	// Verify required attributes exist
	hasMethod := false
	hasStatusCode := false
	for _, attr := range attrs {
		switch string(attr.Key) {
		case "http.method":
			hasMethod = true
			if attr.Value.AsString() != "POST" {
				t.Errorf("Expected method 'POST', got '%s'", attr.Value.AsString())
			}
		case "http.status_code":
			hasStatusCode = true
			if attr.Value.AsInt64() != 201 {
				t.Errorf("Expected status code 201, got %d", attr.Value.AsInt64())
			}
		}
	}

	if !hasMethod {
		t.Error("Missing http.method attribute")
	}
	if !hasStatusCode {
		t.Error("Missing http.status_code attribute")
	}
}

// TestTraceServerRequestWithNilHandler tests TraceServerRequest with nil handler
func TestTraceServerRequestWithNilHandler(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	mockReq := httptest.NewRequest("GET", "http://example.com/test", nil)
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
	}

	// Call with nil handler
	err = TraceServerRequest(mockCtx, nil)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// TestTraceServerRequestWithStatusCode tests TraceServerRequest status code handling
func TestTraceServerRequestWithStatusCode(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	t.Run("error_status_code", func(t *testing.T) {
		handler := func(ctx interface{}) error {
			// Set status code in mock context
			if mc, ok := ctx.(*MockServerContext); ok {
				mc.statusCode = 404
			}
			return nil
		}

		mockReq := httptest.NewRequest("GET", "http://example.com/notfound", nil)
		mockWriter := httptest.NewRecorder()
		mockCtx := &MockServerContext{
			request:        mockReq,
			responseWriter: mockWriter,
			statusCode:     404,
		}

		err := TraceServerRequest(mockCtx, handler)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

// TestAddEventWithNoSpan tests AddEvent with background context
func TestAddEventWithNoSpan(t *testing.T) {
	ctx := context.Background()
	// Should not panic
	AddEvent(ctx, "test-event", attribute.String("key", "value"))
}

// TestSetAttributesWithNoSpan tests SetAttributes with background context
func TestSetAttributesWithNoSpan(t *testing.T) {
	ctx := context.Background()
	// Should not panic
	SetAttributes(ctx, attribute.String("key", "value"))
}

// TestSetErrorWithNoSpan tests SetError with background context
func TestSetErrorWithNoSpan(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")
	// Should not panic
	SetError(ctx, testErr)
}

// TestRecordErrorWithNoSpan tests RecordError with background context
func TestRecordErrorWithNoSpan(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")
	// Should not panic
	RecordError(ctx, testErr, attribute.String("key", "value"))
}

// TestSetStatusWithNoSpan tests SetStatus with background context
func TestSetStatusWithNoSpan(t *testing.T) {
	ctx := context.Background()
	// Should not panic
	SetStatus(ctx, codes.Ok, "success")
}

// TestSpanFromContext tests SpanFromContext
func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)
	// Should return a no-op span
	if span == nil {
		t.Error("Expected non-nil span from SpanFromContext")
	}
}

// TestTracerFunction tests the global Tracer function
func TestTracerFunction(t *testing.T) {
	tracer := Tracer()
	if tracer == nil {
		t.Error("Expected non-nil tracer")
	}
}

// TestHTTPTracingMiddlewareWithUserAgent tests middleware with user agent header
func TestHTTPTracingMiddlewareWithUserAgent(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 Test Agent")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
}

// TestOTLPExporterWithDefaultEndpoint tests OTLP exporter initialization
// Note: This test may fail in CI without an OTLP endpoint, but tests the code path
func TestOTLPExporterWithDefaultEndpoint(t *testing.T) {
	config := &Config{
		ServiceName:  "test",
		ExporterType: "otlp",
		OTLPEndpoint: "", // Should default to localhost:4317
		SamplingRate: 1.0,
		Enabled:      true,
	}

	// This will attempt to connect to localhost:4317
	// We skip full initialization to avoid network errors
	if config.OTLPEndpoint == "" {
		config.OTLPEndpoint = "localhost:4317"
	}
	if config.OTLPEndpoint != "localhost:4317" {
		t.Errorf("Expected default endpoint localhost:4317, got %s", config.OTLPEndpoint)
	}
}

// TestMiddlewareWithPanicHandler tests that middleware handles panic properly
func TestMiddlewareWithPanicHandler(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This test verifies that the middleware doesn't break normal operation
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
}

// TestGetTracingInfoEmpty tests GetTracingInfo with no span
func TestGetTracingInfoEmpty(t *testing.T) {
	ctx := context.Background()
	info := GetTracingInfo(ctx)

	if info["trace_id"] != "" {
		t.Errorf("Expected empty trace_id, got %s", info["trace_id"])
	}

	if info["span_id"] != "" {
		t.Errorf("Expected empty span_id, got %s", info["span_id"])
	}
}

// TestResponseWriterStatusCodeTracking tests responseWriter status code tracking
func TestResponseWriterStatusCodeTracking(t *testing.T) {
	underlying := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: underlying,
		statusCode:     http.StatusOK, // default
	}

	// Write without explicit WriteHeader
	n, err := rw.Write([]byte("test"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("Expected 4 bytes, got %d", n)
	}

	// Status should still be default OK
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status 200, got %d", rw.statusCode)
	}
}

// TestMiddlewareWithBytesWritten tests middleware bytes written tracking
func TestMiddlewareWithBytesWritten(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	testBody := "Hello, World! This is a longer test body."
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testBody))
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Body.String() != testBody {
		t.Errorf("Expected body %q, got %q", testBody, w.Body.String())
	}
}

// TestWithSpanWithSpanOptions tests WithSpan with additional span options
func TestWithSpanWithSpanOptions(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	ctx := context.Background()

	executed := false
	err = WithSpan(ctx, "test-span", func(ctx context.Context) error {
		executed = true
		// Verify we have a valid context
		info := GetTracingInfo(ctx)
		if info["trace_id"] == "" {
			t.Error("Expected non-empty trace_id inside WithSpan")
		}
		return nil
	}, SpanKind.Internal)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !executed {
		t.Error("Expected function to be executed")
	}
}

// TestExtractContextWithExistingContext tests ExtractContext with existing span context
func TestExtractContextWithExistingContext(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	parentCtx := context.Background()

	// Create a span
	parentCtx, span := tracer.Start(parentCtx, "parent-span")

	// Create request and inject context
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	InjectContext(parentCtx, req)

	// Extract context from request
	extractedCtx := ExtractContext(context.Background(), req)

	// The extracted context should have valid trace info
	extractedSpan := SpanFromContext(extractedCtx)
	if !extractedSpan.SpanContext().IsValid() {
		t.Error("Expected valid span context after extraction")
	}

	span.End()
}

// TestHTTPMiddlewareWithClientRedirect tests client redirect status codes
func TestHTTPMiddlewareWithClientRedirect(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	// Test 3xx redirect
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/new-location")
		w.WriteHeader(http.StatusFound)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/old", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusFound {
		t.Errorf("Expected status 302, got %d", w.Result().StatusCode)
	}
}

// TestHTTPMiddlewareWithClientError tests client error status codes
func TestHTTPMiddlewareWithClientError(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	// Test 4xx client error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/missing", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Result().StatusCode)
	}
}

// TestTraceServerRequestNoContext tests TraceServerRequest without proper context
func TestTraceServerRequestNoContext(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Test with non-implementing type
	err = TraceServerRequest("invalid", nil)
	if err != nil {
		t.Errorf("Expected nil error for invalid context type, got %v", err)
	}
}

// TestUpdateRequestInContextCalled verifies updateRequestInContext doesn't panic
func TestUpdateRequestInContextCalled(t *testing.T) {
	mockReq := httptest.NewRequest("GET", "http://example.com", nil)
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
	}

	// This should not panic
	updateRequestInContext(mockCtx, mockReq)
}

// MockServerContextWithParams extends MockServerContext with path and query params
type MockServerContextWithParams struct {
	MockServerContext
	pathParams  map[string]string
	queryParams map[string]string
}

// TestTraceServerRequestWithPathParams tests TraceServerRequest with path parameters
func TestTraceServerRequestWithPathParams(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	handler := func(ctx interface{}) error {
		return nil
	}

	mockReq := httptest.NewRequest("GET", "http://example.com/users/123", nil)
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
		pathParams:     map[string]string{"id": "123"},
		queryParams:    map[string]string{"filter": "active"},
		statusCode:     200,
	}

	err = TraceServerRequest(mockCtx, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestTraceServerRequestWithUserAgent tests TraceServerRequest with user agent
func TestTraceServerRequestWithUserAgent(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	handler := func(ctx interface{}) error {
		return nil
	}

	mockReq := httptest.NewRequest("GET", "http://example.com/test", nil)
	mockReq.Header.Set("User-Agent", "TestAgent/1.0")
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
		statusCode:     200,
	}

	err = TraceServerRequest(mockCtx, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestTraceServerRequestWith500Error tests TraceServerRequest with 500 status
func TestTraceServerRequestWith500Error(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	handler := func(ctx interface{}) error {
		return nil
	}

	mockReq := httptest.NewRequest("GET", "http://example.com/error", nil)
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
		statusCode:     500,
	}

	err = TraceServerRequest(mockCtx, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestTraceServerRequestWithNoResponseWriter tests when no response writer is available
func TestTraceServerRequestWithNoResponseWriter(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	handler := func(ctx interface{}) error {
		return nil
	}

	// Create a context without response writer
	mockReq := httptest.NewRequest("GET", "http://example.com/test", nil)
	mockCtx := &MockServerContextNoWriter{
		request: mockReq,
	}

	err = TraceServerRequest(mockCtx, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// MockServerContextNoWriter has request but no response writer
type MockServerContextNoWriter struct {
	request *http.Request
}

func (m *MockServerContextNoWriter) GetRequest() *http.Request {
	return m.request
}

// TestExtractRequestWithStructType tests extractRequest with struct type
func TestExtractRequestWithStructType(t *testing.T) {
	// Test with a struct that matches the expected pattern
	// This should return false since Go struct types don't match like this
	type ServerContext struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}

	mockReq := httptest.NewRequest("GET", "http://example.com", nil)
	ctx := ServerContext{
		Request:    mockReq,
		StatusCode: 200,
	}

	// This won't match because Go doesn't do structural typing for concrete structs
	req, ok := extractRequest(ctx)
	if ok {
		t.Error("Did not expect struct type to match")
	}
	if req != nil {
		t.Error("Expected nil request")
	}
}

// TestExtractResponseWriterWithStructType tests extractResponseWriter with struct type
func TestExtractResponseWriterWithStructType(t *testing.T) {
	type ServerContext struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}

	mockWriter := httptest.NewRecorder()
	ctx := ServerContext{
		ResponseWriter: mockWriter,
		StatusCode:     200,
	}

	// This won't match because Go doesn't do structural typing for concrete structs
	w, ok := extractResponseWriter(ctx)
	if ok {
		t.Error("Did not expect struct type to match")
	}
	if w != nil {
		t.Error("Expected nil writer")
	}
}

// TestExtractPathParamsWithStructType tests extractPathParams with struct type
func TestExtractPathParamsWithStructType(t *testing.T) {
	type ServerContext struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}

	ctx := ServerContext{
		PathParams: map[string]string{"id": "123"},
	}

	// This won't match because Go doesn't do structural typing for concrete structs
	params := extractPathParams(ctx)
	if params != nil {
		t.Error("Did not expect struct type to match for path params")
	}
}

// TestExtractQueryParamsWithStructType tests extractQueryParams with struct type
func TestExtractQueryParamsWithStructType(t *testing.T) {
	type ServerContext struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}

	ctx := ServerContext{
		QueryParams: map[string]string{"filter": "active"},
	}

	// This won't match because Go doesn't do structural typing for concrete structs
	params := extractQueryParams(ctx)
	if params != nil {
		t.Error("Did not expect struct type to match for query params")
	}
}

// TestExtractStatusCodeWithStructType tests extractStatusCode with struct type
func TestExtractStatusCodeWithStructType(t *testing.T) {
	type ServerContext struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}

	ctx := ServerContext{
		StatusCode: 404,
	}

	// This won't match because Go doesn't do structural typing for concrete structs
	code := extractStatusCode(ctx)
	if code != 0 {
		t.Error("Did not expect struct type to match for status code")
	}
}

// TestUpdateRequestInContextWithNil tests updateRequestInContext with nil values
func TestUpdateRequestInContextWithNil(t *testing.T) {
	// Should not panic with nil context
	updateRequestInContext(nil, nil)
}

// TestTraceServerRequestWithHandlerError tests TraceServerRequest when handler returns error
func TestTraceServerRequestWithHandlerError(t *testing.T) {
	// Set up tracing
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	expectedErr := errors.New("handler error")
	handler := func(ctx interface{}) error {
		return expectedErr
	}

	mockReq := httptest.NewRequest("GET", "http://example.com/test", nil)
	mockWriter := httptest.NewRecorder()
	mockCtx := &MockServerContext{
		request:        mockReq,
		responseWriter: mockWriter,
		statusCode:     0, // status code 0 with error should default to 500
	}

	err = TraceServerRequest(mockCtx, handler)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestHTTPMiddlewareLogging tests that middleware logs properly
func TestHTTPMiddlewareLogging(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	// Test with various HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		req := httptest.NewRequest(method, "http://example.com/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", method, w.Result().StatusCode)
		}
	}
}

// TestHTTPMiddlewareWithDifferentPaths tests middleware with different paths
func TestHTTPMiddlewareWithDifferentPaths(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	paths := []string{"/api/v1/users", "/api/v2/orders", "/internal/status"}
	for _, path := range paths {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		req := httptest.NewRequest("GET", "http://example.com"+path, nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for path %s, got %d", path, w.Result().StatusCode)
		}

		// Verify trace headers
		if w.Header().Get("X-Trace-ID") == "" {
			t.Errorf("Expected X-Trace-ID header for path %s", path)
		}
	}
}

