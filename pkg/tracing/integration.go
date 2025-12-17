package tracing

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Integration helpers for GLYPHLANG server package
// These functions help integrate tracing with the existing server middleware

// NewServerMiddleware creates a tracing middleware compatible with GLYPHLANG server
// This is a helper function that demonstrates how to create a tracing middleware
// that works with the server.Middleware type signature
//
// Example usage:
//
//	import (
//	    "github.com/glyphlang/glyph/pkg/server"
//	    "github.com/glyphlang/glyph/pkg/tracing"
//	)
//
//	// Initialize tracing
//	tp, _ := tracing.InitTracing(tracing.DefaultConfig())
//	defer tp.Shutdown(context.Background())
//
//	// Create the tracing middleware function
//	tracingMiddleware := func(next server.RouteHandler) server.RouteHandler {
//	    return func(ctx *server.Context) error {
//	        return tracing.TraceServerRequest(ctx, next)
//	    }
//	}
//
//	// Use it in the server
//	s := server.NewServer(
//	    server.WithMiddleware(tracingMiddleware),
//	)
func TraceServerRequest(ctx interface{}, next interface{}) error {
	// Type assertions to work with server.Context
	// This is a generic approach that avoids import cycles

	// Use reflection or type assertion to extract fields
	// For now, we'll use a simplified approach with interface{}

	// Extract *http.Request from context
	req, ok := extractRequest(ctx)
	if !ok {
		// If we can't extract the request, just pass through
		if handler, ok := next.(func(interface{}) error); ok {
			return handler(ctx)
		}
		return nil
	}

	w, ok := extractResponseWriter(ctx)
	if !ok {
		if handler, ok := next.(func(interface{}) error); ok {
			return handler(ctx)
		}
		return nil
	}

	// Check if this path should be excluded
	excludedPaths := map[string]bool{
		"/health":  true,
		"/metrics": true,
		"/ping":    true,
	}

	if excludedPaths[req.URL.Path] {
		if handler, ok := next.(func(interface{}) error); ok {
			return handler(ctx)
		}
		return nil
	}

	// Extract trace context from incoming request
	traceCtx := ExtractContext(req.Context(), req)

	// Generate span name
	spanName := fmt.Sprintf("HTTP %s %s", req.Method, req.URL.Path)

	// Start span
	traceCtx, span := StartSpan(traceCtx, spanName, SpanKind.Server)
	defer span.End()

	// Update request context
	req = req.WithContext(traceCtx)
	updateRequestInContext(ctx, req)

	// Record start time
	start := time.Now()

	// Add request attributes
	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.String("http.target", req.URL.Path),
		attribute.String("http.host", req.Host),
		attribute.String("http.scheme", req.URL.Scheme),
		attribute.String("net.peer.ip", req.RemoteAddr),
	}

	if userAgent := req.Header.Get("User-Agent"); userAgent != "" {
		attrs = append(attrs, attribute.String("http.user_agent", userAgent))
	}

	// Add path and query parameters
	if pathParams := extractPathParams(ctx); len(pathParams) > 0 {
		for key, value := range pathParams {
			attrs = append(attrs, attribute.String(fmt.Sprintf("http.path_param.%s", key), value))
		}
	}

	if queryParams := extractQueryParams(ctx); len(queryParams) > 0 {
		for key, value := range queryParams {
			attrs = append(attrs, attribute.String(fmt.Sprintf("http.query_param.%s", key), value))
		}
	}

	span.SetAttributes(attrs...)

	// Add trace IDs to response headers for debugging
	if traceID := GetTraceID(traceCtx); traceID != "" {
		w.Header().Set("X-Trace-ID", traceID)
	}
	if spanID := GetSpanID(traceCtx); spanID != "" {
		w.Header().Set("X-Span-ID", spanID)
	}

	// Call next handler
	var err error
	if handler, ok := next.(func(interface{}) error); ok {
		err = handler(ctx)
	}

	// Record response attributes
	duration := time.Since(start)
	statusCode := extractStatusCode(ctx)
	if err != nil && statusCode == 0 {
		statusCode = 500
	}

	span.SetAttributes(
		attribute.Int("http.status_code", statusCode),
		attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
	)

	// Record error if present
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else if statusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}

	// Log completion with trace IDs
	log.Printf("[TRACE] %s %s - %d (%v) [trace_id=%s span_id=%s]",
		req.Method,
		req.URL.Path,
		statusCode,
		duration,
		GetTraceID(traceCtx),
		GetSpanID(traceCtx),
	)

	return err
}

// Helper functions to extract fields from server.Context using reflection or type assertion

func extractRequest(ctx interface{}) (*http.Request, bool) {
	// Try to extract Request field using type assertion
	type hasRequest interface {
		GetRequest() *http.Request
	}

	// Try direct field access using reflection
	if ctxStruct, ok := ctx.(struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}); ok {
		return ctxStruct.Request, true
	}

	// Try via interface method
	if r, ok := ctx.(hasRequest); ok {
		return r.GetRequest(), true
	}

	return nil, false
}

func extractResponseWriter(ctx interface{}) (http.ResponseWriter, bool) {
	type hasResponseWriter interface {
		GetResponseWriter() http.ResponseWriter
	}

	// Try direct field access
	if ctxStruct, ok := ctx.(struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}); ok {
		return ctxStruct.ResponseWriter, true
	}

	// Try via interface method
	if w, ok := ctx.(hasResponseWriter); ok {
		return w.GetResponseWriter(), true
	}

	return nil, false
}

func extractPathParams(ctx interface{}) map[string]string {
	// Try direct field access using type assertion
	if ctxStruct, ok := ctx.(struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}); ok {
		return ctxStruct.PathParams
	}

	return nil
}

func extractQueryParams(ctx interface{}) map[string]string {
	if ctxStruct, ok := ctx.(struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}); ok {
		return ctxStruct.QueryParams
	}

	return nil
}

func extractStatusCode(ctx interface{}) int {
	if ctxStruct, ok := ctx.(struct {
		Request        *http.Request
		ResponseWriter http.ResponseWriter
		PathParams     map[string]string
		QueryParams    map[string]string
		Body           map[string]interface{}
		StatusCode     int
	}); ok {
		return ctxStruct.StatusCode
	}

	return 0
}

func updateRequestInContext(ctx interface{}, req *http.Request) {
	// This is a no-op for now since we can't easily update struct fields
	// In practice, the context is passed by pointer in the middleware chain
	// so the updates are reflected
}

// ExcludedPaths returns the default list of paths to exclude from tracing
func ExcludedPaths() map[string]bool {
	return map[string]bool{
		"/health":    true,
		"/metrics":   true,
		"/ping":      true,
		"/readiness": true,
		"/liveness":  true,
	}
}

// AddContextToRequest is a helper to add trace context to a request within a handler
func AddContextToRequest(req *http.Request) *http.Request {
	if req == nil {
		return req
	}

	// Extract trace context from the request
	ctx := ExtractContext(req.Context(), req)

	// Return request with updated context
	return req.WithContext(ctx)
}

// GetTraceHeaders extracts trace headers from the current context
func GetTraceHeaders(req *http.Request) map[string]string {
	if req == nil {
		return map[string]string{}
	}

	return map[string]string{
		"traceparent": req.Header.Get("traceparent"),
		"tracestate":  req.Header.Get("tracestate"),
	}
}
