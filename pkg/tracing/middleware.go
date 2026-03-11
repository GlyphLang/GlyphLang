package tracing

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ServerContext is a minimal interface for the server context
// This allows the middleware to work with the GLYPHLANG server package
type ServerContext interface {
	GetRequest() *http.Request
	GetResponseWriter() http.ResponseWriter
	GetStatusCode() int
	SetContext(ctx context.Context)
	GetContext() context.Context
}

// MiddlewareConfig holds configuration for the tracing middleware
type MiddlewareConfig struct {
	// SpanNameFormatter is a function that formats the span name
	// Default format is "HTTP {method} {path}"
	SpanNameFormatter func(req *http.Request) string

	// ExcludePaths is a list of paths to exclude from tracing
	ExcludePaths map[string]bool

	// RecordRequestBody determines if request body should be recorded
	RecordRequestBody bool

	// RecordResponseBody determines if response body should be recorded
	RecordResponseBody bool

	// CustomAttributes is a function that returns custom attributes to add to the span
	CustomAttributes func(req *http.Request) []attribute.KeyValue
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SpanNameFormatter: func(req *http.Request) string {
			return fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)
		},
		ExcludePaths: map[string]bool{
			"/health":  true,
			"/metrics": true,
			"/ping":    true,
		},
		RecordRequestBody:  false,
		RecordResponseBody: false,
	}
}

// HTTPTracingMiddleware creates a middleware that traces HTTP requests
// This is designed to work with the GLYPHLANG server middleware pattern
func HTTPTracingMiddleware(config *MiddlewareConfig) func(next http.Handler) http.Handler {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be excluded
			if config.ExcludePaths != nil && config.ExcludePaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract trace context from incoming request
			ctx := ExtractContext(r.Context(), r)

			// Generate span name
			spanName := config.SpanNameFormatter(r)

			// Start span
			ctx, span := StartSpan(ctx, spanName, SpanKind.Server)
			defer span.End()

			// Create a custom response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Record start time
			start := time.Now()

			// Add request attributes
			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.target", r.URL.Path),
				attribute.String("http.host", r.Host),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("net.peer.ip", r.RemoteAddr),
			}

			if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
				attrs = append(attrs, attribute.String("http.user_agent", userAgent))
			}

			if config.CustomAttributes != nil {
				attrs = append(attrs, config.CustomAttributes(r)...)
			}

			span.SetAttributes(attrs...)

			// Add trace IDs to response headers for debugging
			if traceID := GetTraceID(ctx); traceID != "" {
				wrapped.Header().Set("X-Trace-ID", traceID)
			}
			if spanID := GetSpanID(ctx); spanID != "" {
				wrapped.Header().Set("X-Span-ID", spanID)
			}

			// Create new request with traced context
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Record response attributes
			duration := time.Since(start)
			span.SetAttributes(
				attribute.Int("http.status_code", wrapped.statusCode),
				attribute.Int64("http.response_size", wrapped.bytesWritten),
				attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			)

			// Set span status based on HTTP status code
			if wrapped.statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Log completion
			log.Printf("[TRACE] %s %s - %d (%v) [trace_id=%s span_id=%s]", // #nosec G706 -- sanitized
				sanitizeLog(r.Method),
				sanitizeLog(r.URL.Path),
				wrapped.statusCode,
				duration,
				GetTraceID(ctx),
				GetSpanID(ctx),
			)
		})
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// TraceOutgoingRequest traces an outgoing HTTP request
// This should be called before making HTTP client requests
func TraceOutgoingRequest(ctx context.Context, req *http.Request, spanName string) (context.Context, trace.Span) {
	// Start client span
	ctx, span := StartSpan(ctx, spanName, SpanKind.Client)

	// Add request attributes
	span.SetAttributes(
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.target", req.URL.Path),
		attribute.String("http.host", req.Host),
	)

	// Inject trace context into request headers
	InjectContext(ctx, req)

	return ctx, span
}

// RecordOutgoingResponse records the response from an outgoing HTTP request
func RecordOutgoingResponse(ctx context.Context, resp *http.Response, err error) {
	span := SpanFromContext(ctx)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if resp != nil {
		span.SetAttributes(
			attribute.Int("http.status_code", resp.StatusCode),
		)

		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}

// WithHTTPClientTrace is a helper for tracing HTTP client requests
func WithHTTPClientTrace(ctx context.Context, req *http.Request, client *http.Client) (*http.Response, error) {
	spanName := fmt.Sprintf("HTTP Client %s %s", req.Method, req.URL.Path)
	ctx, span := TraceOutgoingRequest(ctx, req, spanName)
	defer span.End()

	// Make the request with the traced context
	req = req.WithContext(ctx)
	resp, err := client.Do(req)

	// Record the response
	RecordOutgoingResponse(ctx, resp, err)

	return resp, err
}
