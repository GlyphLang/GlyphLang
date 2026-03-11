// Package tracing provides OpenTelemetry distributed tracing functionality for GLYPHLANG.
// It supports W3C Trace Context propagation, span creation and management, HTTP request tracing,
// and configurable exporters for both development and production environments.
package tracing

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// sanitizeLog replaces newlines and carriage returns in user-controlled values
// to prevent log injection attacks (gosec G706).
func sanitizeLog(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// Config holds the configuration for the tracing system
type Config struct {
	// ServiceName is the name of the service being traced
	ServiceName string

	// ServiceVersion is the version of the service
	ServiceVersion string

	// Environment specifies the deployment environment (dev, staging, prod)
	Environment string

	// ExporterType specifies which exporter to use ("stdout" or "otlp")
	ExporterType string

	// OTLPEndpoint is the endpoint for the OTLP exporter (e.g., "localhost:4317")
	OTLPEndpoint string

	// SamplingRate is the rate at which traces are sampled (0.0 to 1.0)
	// 1.0 means all traces are sampled, 0.5 means 50% are sampled
	SamplingRate float64

	// Enabled determines if tracing is enabled
	Enabled bool
}

// DefaultConfig returns a default configuration for development
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    "glyphlang",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		ExporterType:   "stdout",
		SamplingRate:   1.0,
		Enabled:        true,
	}
}

// TracerProvider wraps the OpenTelemetry tracer provider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	config   *Config
}

// InitTracing initializes the OpenTelemetry tracing system
// It returns a TracerProvider that should be shut down when the application exits
func InitTracing(config *Config) (*TracerProvider, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if !config.Enabled {
		// Return a no-op provider if tracing is disabled
		return &TracerProvider{
			provider: sdktrace.NewTracerProvider(),
			config:   config,
		}, nil
	}

	// Create exporter based on configuration
	var exporter sdktrace.SpanExporter
	var err error

	switch config.ExporterType {
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	case "otlp":
		if config.OTLPEndpoint == "" {
			config.OTLPEndpoint = "localhost:4317"
		}
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		exporter, err = otlptrace.New(context.Background(), client)
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", config.ExporterType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create sampler based on sampling rate
	var sampler sdktrace.Sampler
	if config.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SamplingRate)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to W3C Trace Context
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &TracerProvider{
		provider: tp,
		config:   config,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}
	return tp.provider.Shutdown(ctx)
}

// GetTracer returns a tracer for the given name
func (tp *TracerProvider) GetTracer(name string) trace.Tracer {
	if tp.provider == nil {
		return otel.Tracer(name)
	}
	return tp.provider.Tracer(name)
}

// Tracer returns the global tracer for GLYPHLANG
func Tracer() trace.Tracer {
	return otel.Tracer("glyphlang")
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, spanName, opts...)
}

// SpanFromContext returns the current span from the context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// InjectContext injects trace context into HTTP headers for outgoing requests
func InjectContext(ctx context.Context, req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractContext extracts trace context from HTTP headers for incoming requests
func ExtractContext(ctx context.Context, req *http.Request) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))
}

// GetTraceID extracts the trace ID from the context
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extracts the span ID from the context
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// SetError marks the current span as having an error
func SetError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetStatus sets the status of the current span
func SetStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(code, description)
}

// HTTPAttributes returns common HTTP attributes for a request
func HTTPAttributes(req *http.Request, statusCode int) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPMethod(req.Method),
		semconv.HTTPTarget(req.URL.Path),
		semconv.HTTPScheme(req.URL.Scheme),
		semconv.HTTPStatusCode(statusCode),
		semconv.NetHostName(req.Host),
	}

	if req.URL.RawQuery != "" {
		attrs = append(attrs, attribute.String("http.query", req.URL.RawQuery))
	}

	if userAgent := req.Header.Get("User-Agent"); userAgent != "" {
		attrs = append(attrs, attribute.String("http.user_agent", userAgent))
	}

	if clientIP := req.Header.Get("X-Forwarded-For"); clientIP != "" {
		attrs = append(attrs, attribute.String("http.client_ip", clientIP))
	} else if clientIP := req.RemoteAddr; clientIP != "" {
		attrs = append(attrs, attribute.String("http.client_ip", clientIP))
	}

	return attrs
}

// HTTPClientAttributes returns common HTTP attributes for outgoing requests
func HTTPClientAttributes(req *http.Request, statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPMethod(req.Method),
		semconv.HTTPURL(req.URL.String()),
		semconv.HTTPStatusCode(statusCode),
		attribute.String("http.target", req.URL.Path),
	}
}

// WithSpan is a helper function that creates a span, executes a function, and properly closes the span
func WithSpan(ctx context.Context, spanName string, fn func(context.Context) error, opts ...trace.SpanStartOption) error {
	ctx, span := StartSpan(ctx, spanName, opts...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		SetError(ctx, err)
	}

	return err
}

// RecordError records an error with additional context
func RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(attrs...))
	span.SetStatus(codes.Error, err.Error())
}

// GetTracingInfo returns trace ID and span ID as a map for logging
func GetTracingInfo(ctx context.Context) map[string]string {
	return map[string]string{
		"trace_id": GetTraceID(ctx),
		"span_id":  GetSpanID(ctx),
	}
}

// IsTracingEnabled checks if tracing is enabled in the environment
func IsTracingEnabled() bool {
	enabled := os.Getenv("OTEL_SDK_DISABLED")
	return enabled != "true"
}

// SpanKind returns span kind options for common scenarios
var SpanKind = struct {
	Server   trace.SpanStartOption
	Client   trace.SpanStartOption
	Internal trace.SpanStartOption
	Producer trace.SpanStartOption
	Consumer trace.SpanStartOption
}{
	Server:   trace.WithSpanKind(trace.SpanKindServer),
	Client:   trace.WithSpanKind(trace.SpanKindClient),
	Internal: trace.WithSpanKind(trace.SpanKindInternal),
	Producer: trace.WithSpanKind(trace.SpanKindProducer),
	Consumer: trace.WithSpanKind(trace.SpanKindConsumer),
}
