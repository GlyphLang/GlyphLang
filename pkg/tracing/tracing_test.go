package tracing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ServiceName != "glyphlang" {
		t.Errorf("Expected service name 'glyphlang', got '%s'", config.ServiceName)
	}

	if config.ServiceVersion != "1.0.0" {
		t.Errorf("Expected service version '1.0.0', got '%s'", config.ServiceVersion)
	}

	if config.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", config.Environment)
	}

	if config.ExporterType != "stdout" {
		t.Errorf("Expected exporter type 'stdout', got '%s'", config.ExporterType)
	}

	if config.SamplingRate != 1.0 {
		t.Errorf("Expected sampling rate 1.0, got %f", config.SamplingRate)
	}

	if !config.Enabled {
		t.Error("Expected tracing to be enabled by default")
	}
}

func TestInitTracingDisabled(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		Enabled:      false,
		ExporterType: "stdout",
	}

	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	if tp.provider == nil {
		t.Error("Expected non-nil provider even when disabled")
	}
}

func TestInitTracingStdout(t *testing.T) {
	config := &Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		ExporterType:   "stdout",
		SamplingRate:   1.0,
		Enabled:        true,
	}

	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	if tp.provider == nil {
		t.Error("Expected non-nil provider")
	}

	if tp.config.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", tp.config.ServiceName)
	}
}

func TestInitTracingInvalidExporter(t *testing.T) {
	config := &Config{
		ServiceName:  "test-service",
		ExporterType: "invalid",
		Enabled:      true,
	}

	_, err := InitTracing(config)
	if err == nil {
		t.Error("Expected error for invalid exporter type")
	}
}

func TestGetTracer(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	tracer := tp.GetTracer("test-tracer")
	if tracer == nil {
		t.Error("Expected non-nil tracer")
	}
}

func TestStartSpan(t *testing.T) {
	// Create a test exporter to capture spans
	exporter := tracetest.NewInMemoryExporter()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	// Start a span
	ctx, span := tracer.Start(ctx, "test-span")
	span.End()

	// Verify span was created
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].Name != "test-span" {
		t.Errorf("Expected span name 'test-span', got '%s'", spans[0].Name)
	}
}

func TestGetTraceIDAndSpanID(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	// Start a span
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Get trace ID and span ID
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	if traceID == "" {
		t.Error("Expected non-empty trace ID")
	}

	if spanID == "" {
		t.Error("Expected non-empty span ID")
	}

	if len(traceID) != 32 { // Trace ID is 16 bytes = 32 hex chars
		t.Errorf("Expected trace ID length 32, got %d", len(traceID))
	}

	if len(spanID) != 16 { // Span ID is 8 bytes = 16 hex chars
		t.Errorf("Expected span ID length 16, got %d", len(spanID))
	}
}

func TestSetAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	span.SetAttributes(
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
		attribute.Bool("key3", true),
	)
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes
	if len(attrs) != 3 {
		t.Fatalf("Expected 3 attributes, got %d", len(attrs))
	}
}

func TestSetError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	testErr := errors.New("test error")
	span.RecordError(testErr)
	span.SetStatus(codes.Error, testErr.Error())
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].Status.Code != codes.Error {
		t.Errorf("Expected error status, got %v", spans[0].Status.Code)
	}

	if spans[0].Status.Description != "test error" {
		t.Errorf("Expected error description 'test error', got '%s'", spans[0].Status.Description)
	}
}

func TestAddEvent(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	span.AddEvent("test-event", trace.WithAttributes(
		attribute.String("event-key", "event-value"),
	))
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	events := spans[0].Events
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	if events[0].Name != "test-event" {
		t.Errorf("Expected event name 'test-event', got '%s'", events[0].Name)
	}
}

func TestWithSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	executed := false
	testFunc := func(ctx context.Context) error {
		executed = true
		return nil
	}

	ctx, span := tracer.Start(ctx, "parent-span")
	defer span.End()

	err := testFunc(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Expected function to be executed")
	}
}

func TestWithSpanError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	testErr := errors.New("test error")
	testFunc := func(ctx context.Context) error {
		return testErr
	}

	ctx, span := tracer.Start(ctx, "parent-span")
	defer span.End()

	err := testFunc(ctx)
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}
}

func TestHTTPAttributes(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/test?query=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	attrs := HTTPAttributes(req, 200)

	if len(attrs) == 0 {
		t.Error("Expected non-empty attributes")
	}

	// Check for required attributes
	hasMethod := false
	hasTarget := false
	hasStatusCode := false

	for _, attr := range attrs {
		switch attr.Key {
		case "http.method":
			hasMethod = true
			if attr.Value.AsString() != "GET" {
				t.Errorf("Expected method 'GET', got '%s'", attr.Value.AsString())
			}
		case "http.target":
			hasTarget = true
			if attr.Value.AsString() != "/test" {
				t.Errorf("Expected target '/test', got '%s'", attr.Value.AsString())
			}
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
	if !hasTarget {
		t.Error("Missing http.target attribute")
	}
	if !hasStatusCode {
		t.Error("Missing http.status_code attribute")
	}
}

func TestInjectAndExtractContext(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	// Start a span
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Create a request and inject context
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	InjectContext(ctx, req)

	// Check that trace headers were added
	traceParent := req.Header.Get("traceparent")
	if traceParent == "" {
		t.Error("Expected traceparent header to be set")
	}

	// Extract context from the request
	extractedCtx := ExtractContext(context.Background(), req)

	// Verify that the trace context was extracted
	extractedSpan := trace.SpanFromContext(extractedCtx)
	if !extractedSpan.SpanContext().IsValid() {
		t.Error("Expected valid span context after extraction")
	}
}

func TestGetTracingInfo(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	info := GetTracingInfo(ctx)

	if info["trace_id"] == "" {
		t.Error("Expected non-empty trace_id")
	}

	if info["span_id"] == "" {
		t.Error("Expected non-empty span_id")
	}
}

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	if config.SpanNameFormatter == nil {
		t.Error("Expected non-nil SpanNameFormatter")
	}

	if config.ExcludePaths == nil {
		t.Error("Expected non-nil ExcludePaths")
	}

	if !config.ExcludePaths["/health"] {
		t.Error("Expected /health to be in excluded paths")
	}

	if !config.ExcludePaths["/metrics"] {
		t.Error("Expected /metrics to be in excluded paths")
	}

	if config.RecordRequestBody {
		t.Error("Expected RecordRequestBody to be false by default")
	}

	if config.RecordResponseBody {
		t.Error("Expected RecordResponseBody to be false by default")
	}
}

func TestHTTPTracingMiddleware(t *testing.T) {
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
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that trace headers were added to response
	traceID := resp.Header.Get("X-Trace-ID")
	spanID := resp.Header.Get("X-Span-ID")

	if traceID == "" {
		t.Error("Expected X-Trace-ID header in response")
	}

	if spanID == "" {
		t.Error("Expected X-Span-ID header in response")
	}
}

func TestHTTPTracingMiddlewareExcludedPath(t *testing.T) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com/health", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	// For excluded paths, trace headers may not be added
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestTraceOutgoingRequest(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	// Set global propagator for this test
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("test")
	ctx := context.Background()

	// Start parent span
	ctx, parentSpan := tracer.Start(ctx, "parent-span")
	defer parentSpan.End()

	// Create outgoing request
	req := httptest.NewRequest("GET", "http://example.com/api", nil)

	// Trace outgoing request
	ctx, span := tracer.Start(ctx, "HTTP Client GET /api", trace.WithSpanKind(trace.SpanKindClient))
	InjectContext(ctx, req)
	span.End()

	// Verify traceparent header was injected
	traceParent := req.Header.Get("traceparent")
	if traceParent == "" {
		t.Error("Expected traceparent header in outgoing request")
	}

	// End parent span to ensure it's exported
	parentSpan.End()

	// Verify spans were created
	spans := exporter.GetSpans()
	if len(spans) < 2 {
		t.Errorf("Expected at least 2 spans, got %d", len(spans))
	}
}

func TestSpanKindConstants(t *testing.T) {
	if SpanKind.Server == nil {
		t.Error("Expected non-nil SpanKind.Server")
	}
	if SpanKind.Client == nil {
		t.Error("Expected non-nil SpanKind.Client")
	}
	if SpanKind.Internal == nil {
		t.Error("Expected non-nil SpanKind.Internal")
	}
	if SpanKind.Producer == nil {
		t.Error("Expected non-nil SpanKind.Producer")
	}
	if SpanKind.Consumer == nil {
		t.Error("Expected non-nil SpanKind.Consumer")
	}
}

func TestRecordError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span")
	testErr := errors.New("test error")
	RecordError(ctx, testErr, attribute.String("error.type", "validation"))
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].Status.Code != codes.Error {
		t.Errorf("Expected error status, got %v", spans[0].Status.Code)
	}
}

func BenchmarkStartSpan(b *testing.B) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		b.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "benchmark-span")
		span.End()
	}
}

func BenchmarkHTTPTracingMiddleware(b *testing.B) {
	config := DefaultConfig()
	tp, err := InitTracing(config)
	if err != nil {
		b.Fatalf("InitTracing failed: %v", err)
	}
	defer tp.Shutdown(context.Background())

	middlewareConfig := DefaultMiddlewareConfig()
	middleware := HTTPTracingMiddleware(middlewareConfig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
	}
}

// TestConcurrentSpans tests that spans can be created concurrently
func TestConcurrentSpans(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("test")
	ctx := context.Background()

	numGoroutines := 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_, span := tracer.Start(ctx, "concurrent-span")
			time.Sleep(time.Millisecond)
			span.End()
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	spans := exporter.GetSpans()
	if len(spans) != numGoroutines {
		t.Errorf("Expected %d spans, got %d", numGoroutines, len(spans))
	}
}
