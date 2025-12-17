package tracing_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/glyphlang/glyph/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// ExampleInitTracing demonstrates basic tracing initialization
func ExampleInitTracing() {
	// Create configuration
	config := &tracing.Config{
		ServiceName:    "my-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		ExporterType:   "stdout",
		SamplingRate:   1.0,
		Enabled:        true,
	}

	// Initialize tracing
	tp, err := tracing.InitTracing(config)
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	fmt.Println("Tracing initialized successfully")
	// Output: Tracing initialized successfully
}

// ExampleStartSpan demonstrates creating a span
func ExampleStartSpan() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()

	// Start a span
	ctx, span := tracing.StartSpan(ctx, "process-order", tracing.SpanKind.Internal)
	defer span.End()

	// Do some work
	tracing.SetAttributes(ctx,
		attribute.String("order.id", "12345"),
		attribute.Int("order.items", 3),
	)

	fmt.Println("Span created successfully")
	// Output: Span created successfully
}

// ExampleWithSpan demonstrates using WithSpan helper
func ExampleWithSpan() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()

	err := tracing.WithSpan(ctx, "database-query", func(ctx context.Context) error {
		// Simulate database operation
		tracing.AddEvent(ctx, "query-started")
		// ... execute query ...
		tracing.AddEvent(ctx, "query-completed")
		return nil
	}, tracing.SpanKind.Internal)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database query traced successfully")
	// Output: Database query traced successfully
}

// ExampleHTTPTracingMiddleware demonstrates HTTP middleware usage
func ExampleHTTPTracingMiddleware() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	// Create middleware
	middlewareConfig := tracing.DefaultMiddlewareConfig()
	middleware := tracing.HTTPTracingMiddleware(middlewareConfig)

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Wrap handler with tracing middleware
	tracedHandler := middleware(handler)

	// Make request
	req := httptest.NewRequest("GET", "http://example.com/api/users", nil)
	w := httptest.NewRecorder()
	tracedHandler.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output: Status: 200
}

// ExampleTraceOutgoingRequest demonstrates tracing HTTP client requests
func ExampleTraceOutgoingRequest() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()

	// Start parent span
	ctx, parentSpan := tracing.StartSpan(ctx, "handle-request", tracing.SpanKind.Server)
	defer parentSpan.End()

	// Create outgoing request
	req := httptest.NewRequest("GET", "http://api.example.com/data", nil)

	// Trace outgoing request
	ctx, span := tracing.TraceOutgoingRequest(ctx, req, "GET /data")
	defer span.End()

	// Make the request (simulated)
	resp := &http.Response{StatusCode: 200}

	// Record the response
	tracing.RecordOutgoingResponse(ctx, resp, nil)

	fmt.Println("Outgoing request traced")
	// Output: Outgoing request traced
}

// ExampleGetTracingInfo demonstrates extracting trace IDs for logging
func ExampleGetTracingInfo() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := tracing.StartSpan(ctx, "operation", tracing.SpanKind.Internal)
	defer span.End()

	// Get trace info for structured logging
	info := tracing.GetTracingInfo(ctx)

	if info["trace_id"] != "" && info["span_id"] != "" {
		fmt.Println("Trace IDs extracted successfully")
	}
	// Output: Trace IDs extracted successfully
}

// ExampleSetError demonstrates error recording
func ExampleSetError() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := tracing.StartSpan(ctx, "risky-operation", tracing.SpanKind.Internal)
	defer span.End()

	err := fmt.Errorf("something went wrong")
	if err != nil {
		tracing.SetError(ctx, err)
		fmt.Println("Error recorded in span")
	}
	// Output: Error recorded in span
}

// ExampleInjectContext demonstrates trace context propagation
func ExampleInjectContext() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := tracing.StartSpan(ctx, "parent-operation", tracing.SpanKind.Server)
	defer span.End()

	// Create outgoing HTTP request
	req, _ := http.NewRequest("GET", "http://api.example.com/data", nil)

	// Inject trace context into request headers
	tracing.InjectContext(ctx, req)

	// The request now carries W3C Trace Context headers
	if req.Header.Get("traceparent") != "" {
		fmt.Println("Trace context injected into request headers")
	}
	// Output: Trace context injected into request headers
}

// ExampleExtractContext demonstrates extracting trace context from incoming requests
func ExampleExtractContext() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	// Simulate incoming request with trace context
	req := httptest.NewRequest("GET", "http://example.com/api", nil)

	// Extract trace context from request
	ctx := tracing.ExtractContext(context.Background(), req)

	// Start a span with the extracted context
	ctx, span := tracing.StartSpan(ctx, "handle-api-request", tracing.SpanKind.Server)
	defer span.End()

	fmt.Println("Trace context extracted from request")
	// Output: Trace context extracted from request
}

// Example_customAttributes demonstrates adding custom attributes to spans
func Example_customAttributes() {
	config := tracing.DefaultConfig()
	tp, _ := tracing.InitTracing(config)
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	ctx, span := tracing.StartSpan(ctx, "process-payment", tracing.SpanKind.Internal)
	defer span.End()

	// Add custom attributes
	tracing.SetAttributes(ctx,
		attribute.String("payment.method", "credit_card"),
		attribute.Float64("payment.amount", 99.99),
		attribute.String("payment.currency", "USD"),
		attribute.String("customer.id", "cust_12345"),
	)

	// Add an event with attributes
	tracing.AddEvent(ctx, "payment-processed",
		attribute.String("transaction.id", "txn_67890"),
	)

	fmt.Println("Custom attributes added to span")
	// Output: Custom attributes added to span
}

// Example demonstrating complete server setup with tracing
func Example_serverSetup() {
	// Initialize tracing
	config := &tracing.Config{
		ServiceName:    "glyphlang-api",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		ExporterType:   "otlp",
		OTLPEndpoint:   "localhost:4317",
		SamplingRate:   1.0,
		Enabled:        true,
	}

	tp, err := tracing.InitTracing(config)
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	// Create tracing middleware
	middlewareConfig := &tracing.MiddlewareConfig{
		SpanNameFormatter: func(req *http.Request) string {
			return fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)
		},
		ExcludePaths: map[string]bool{
			"/health":  true,
			"/metrics": true,
		},
		CustomAttributes: func(req *http.Request) []attribute.KeyValue {
			return []attribute.KeyValue{
				attribute.String("service.name", "glyphlang-api"),
			}
		},
	}

	middleware := tracing.HTTPTracingMiddleware(middlewareConfig)

	// Create HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context
		ctx := r.Context()

		// Create a child span for business logic
		ctx, span := tracing.StartSpan(ctx, "business-logic", tracing.SpanKind.Internal)
		defer span.End()

		// Add custom attributes
		tracing.SetAttributes(ctx,
			attribute.String("user.id", "user123"),
		)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap handler with tracing middleware
	_ = middleware(handler)

	fmt.Println("Server configured with tracing")
	// Output: Server configured with tracing
}
