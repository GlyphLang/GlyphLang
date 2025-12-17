package metrics_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/glyphlang/glyph/pkg/metrics"
	"github.com/glyphlang/glyph/pkg/server"
)

// ExampleNewMetrics demonstrates basic metrics usage
func ExampleNewMetrics() {
	// Create metrics with default config
	m := metrics.NewMetrics(metrics.DefaultConfig())

	// Record some requests
	m.RecordRequest("GET", "/api/users", 200, 50*time.Millisecond)
	m.RecordRequest("POST", "/api/users", 201, 100*time.Millisecond)
	m.RecordRequest("GET", "/api/users/123", 404, 10*time.Millisecond)

	// Update runtime metrics manually (usually done automatically)
	m.UpdateRuntimeMetrics()

	fmt.Println("Metrics recorded successfully")
	// Output: Metrics recorded successfully
}

// ExampleMetrics_RegisterCustomCounter demonstrates custom counter registration
func ExampleMetrics_RegisterCustomCounter() {
	m := metrics.NewMetrics(metrics.DefaultConfig())

	// Register a custom counter for tracking user signups
	err := m.RegisterCustomCounter(
		"user_signups_total",
		"Total number of user signups",
		[]string{"plan", "region"},
	)
	if err != nil {
		panic(err)
	}

	// Increment the counter
	m.IncrementCustomCounter("user_signups_total", map[string]string{
		"plan":   "premium",
		"region": "us-east",
	})

	fmt.Println("Custom counter registered and incremented")
	// Output: Custom counter registered and incremented
}

// ExampleMetrics_RegisterCustomGauge demonstrates custom gauge registration
func ExampleMetrics_RegisterCustomGauge() {
	m := metrics.NewMetrics(metrics.DefaultConfig())

	// Register a custom gauge for queue size
	err := m.RegisterCustomGauge(
		"queue_size",
		"Current size of processing queue",
		[]string{"queue_name"},
	)
	if err != nil {
		panic(err)
	}

	// Set the gauge value
	m.SetCustomGauge("queue_size", 42.0, map[string]string{
		"queue_name": "email",
	})

	fmt.Println("Custom gauge registered and set")
	// Output: Custom gauge registered and set
}

// ExampleMetrics_RegisterCustomHistogram demonstrates custom histogram registration
func ExampleMetrics_RegisterCustomHistogram() {
	m := metrics.NewMetrics(metrics.DefaultConfig())

	// Register a custom histogram for database query times
	buckets := []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0}
	err := m.RegisterCustomHistogram(
		"db_query_duration_seconds",
		"Database query execution time in seconds",
		[]string{"query_type"},
		buckets,
	)
	if err != nil {
		panic(err)
	}

	// Observe query durations
	m.ObserveCustomHistogram("db_query_duration_seconds", 0.023, map[string]string{
		"query_type": "select",
	})
	m.ObserveCustomHistogram("db_query_duration_seconds", 0.045, map[string]string{
		"query_type": "insert",
	})

	fmt.Println("Custom histogram registered and observations recorded")
	// Output: Custom histogram registered and observations recorded
}

// ExampleMetricsMiddleware demonstrates middleware integration
func ExampleMetricsMiddleware() {
	// Create metrics instance
	m := metrics.NewMetrics(metrics.DefaultConfig())

	// Create the metrics middleware
	metricsMiddleware := metrics.MetricsMiddleware(m)

	// Create a sample handler
	handler := func(ctx *server.Context) error {
		ctx.StatusCode = 200
		return nil
	}

	// Wrap handler with middleware
	wrappedHandler := metricsMiddleware(handler)

	// Use the wrapped handler in your routes
	_ = wrappedHandler

	fmt.Println("Metrics middleware created and applied")
	// Output: Metrics middleware created and applied
}

// ExampleMetrics_Handler demonstrates setting up the /metrics endpoint
func ExampleMetrics_Handler() {
	// Create metrics instance
	m := metrics.NewMetrics(metrics.DefaultConfig())

	// Record some metrics
	m.RecordRequest("GET", "/api/test", 200, 50*time.Millisecond)

	// Get the HTTP handler for the /metrics endpoint
	metricsHandler := m.Handler()

	// Register it with your HTTP server
	http.Handle("/metrics", metricsHandler)

	fmt.Println("Metrics endpoint handler created")
	// Output: Metrics endpoint handler created
}

// ExampleConfig demonstrates custom configuration
func ExampleConfig() {
	// Create custom configuration
	config := metrics.Config{
		Namespace: "myapp",
		Subsystem: "api",
		DurationBuckets: []float64{
			0.001, 0.01, 0.1, 1.0, 10.0,
		},
	}

	// Create metrics with custom config
	m := metrics.NewMetrics(config)

	// Record requests - metrics will use "myapp_api" prefix
	m.RecordRequest("GET", "/health", 200, 5*time.Millisecond)

	fmt.Println("Custom metrics configuration applied")
	// Output: Custom metrics configuration applied
}
