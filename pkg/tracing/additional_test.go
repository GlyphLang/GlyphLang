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

