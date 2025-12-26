package metrics

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics(DefaultConfig())
	assert.NotNil(t, m)
	assert.NotNil(t, m.registry)
	assert.NotNil(t, m.requestsTotal)
	assert.NotNil(t, m.requestDuration)
	assert.NotNil(t, m.requestErrors)
	assert.NotNil(t, m.goroutines)
	assert.NotNil(t, m.memoryAlloc)
	assert.NotNil(t, m.customCounters)
	assert.NotNil(t, m.customGauges)
	assert.NotNil(t, m.customHistograms)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "glyphlang", config.Namespace)
	assert.Equal(t, "http", config.Subsystem)
	assert.NotEmpty(t, config.DurationBuckets)
	assert.Len(t, config.DurationBuckets, 12)
}

func TestRecordRequest(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		duration   time.Duration
	}{
		{
			name:       "successful GET request",
			method:     "GET",
			path:       "/api/users",
			statusCode: 200,
			duration:   50 * time.Millisecond,
		},
		{
			name:       "POST request",
			method:     "POST",
			path:       "/api/users",
			statusCode: 201,
			duration:   100 * time.Millisecond,
		},
		{
			name:       "error request",
			method:     "GET",
			path:       "/api/users/123",
			statusCode: 404,
			duration:   10 * time.Millisecond,
		},
		{
			name:       "server error",
			method:     "POST",
			path:       "/api/posts",
			statusCode: 500,
			duration:   200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordRequest(tt.method, tt.path, tt.statusCode, tt.duration)

			// Verify counter was incremented
			status := strconv.Itoa(tt.statusCode)
			count := testutil.ToFloat64(m.requestsTotal.WithLabelValues(tt.method, tt.path, status))
			assert.Greater(t, count, 0.0)

			// Verify error counter for error status codes
			if tt.statusCode >= 400 {
				errorCount := testutil.ToFloat64(m.requestErrors.WithLabelValues(tt.method, tt.path, status))
				assert.Greater(t, errorCount, 0.0)
			}
		})
	}
}

func TestUpdateRuntimeMetrics(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Update metrics
	m.UpdateRuntimeMetrics()

	// Verify goroutines metric
	goroutines := testutil.ToFloat64(m.goroutines)
	assert.Greater(t, goroutines, 0.0)
	assert.LessOrEqual(t, goroutines, float64(runtime.NumGoroutine()+10)) // Allow some tolerance

	// Verify memory metrics
	memAlloc := testutil.ToFloat64(m.memoryAlloc)
	assert.Greater(t, memAlloc, 0.0)

	memTotal := testutil.ToFloat64(m.memoryTotal)
	assert.Greater(t, memTotal, 0.0)

	memSys := testutil.ToFloat64(m.memorySystem)
	assert.Greater(t, memSys, 0.0)
}

func TestRegisterCustomCounter(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	t.Run("successful registration", func(t *testing.T) {
		err := m.RegisterCustomCounter("test_counter", "A test counter", []string{"label1", "label2"})
		assert.NoError(t, err)
		assert.Contains(t, m.customCounters, "test_counter")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		err := m.RegisterCustomCounter("test_counter", "A test counter", []string{"label1"})
		assert.Error(t, err)
	})
}

func TestRegisterCustomGauge(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	t.Run("successful registration", func(t *testing.T) {
		err := m.RegisterCustomGauge("test_gauge", "A test gauge", []string{"label1"})
		assert.NoError(t, err)
		assert.Contains(t, m.customGauges, "test_gauge")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		err := m.RegisterCustomGauge("test_gauge", "A test gauge", []string{"label1"})
		assert.Error(t, err)
	})
}

func TestRegisterCustomHistogram(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	t.Run("successful registration with buckets", func(t *testing.T) {
		buckets := []float64{0.1, 0.5, 1.0, 5.0}
		err := m.RegisterCustomHistogram("test_histogram", "A test histogram", []string{"label1"}, buckets)
		assert.NoError(t, err)
		assert.Contains(t, m.customHistograms, "test_histogram")
	})

	t.Run("successful registration without buckets", func(t *testing.T) {
		err := m.RegisterCustomHistogram("test_histogram2", "Another test histogram", []string{"label1"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, m.customHistograms, "test_histogram2")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		err := m.RegisterCustomHistogram("test_histogram", "A test histogram", []string{"label1"}, nil)
		assert.Error(t, err)
	})
}

func TestIncrementCustomCounter(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Register a custom counter
	err := m.RegisterCustomCounter("requests_by_user", "Requests by user", []string{"user", "endpoint"})
	require.NoError(t, err)

	// Increment the counter
	labels := map[string]string{"user": "alice", "endpoint": "/api/data"}
	m.IncrementCustomCounter("requests_by_user", labels)
	m.IncrementCustomCounter("requests_by_user", labels)

	// Verify the counter was incremented
	counter := m.customCounters["requests_by_user"]
	assert.NotNil(t, counter)

	count := testutil.ToFloat64(counter.With(prometheus.Labels(labels)))
	assert.Equal(t, 2.0, count)
}

func TestSetCustomGauge(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Register a custom gauge
	err := m.RegisterCustomGauge("queue_size", "Size of processing queue", []string{"queue_name"})
	require.NoError(t, err)

	// Set the gauge value
	labels := map[string]string{"queue_name": "email"}
	m.SetCustomGauge("queue_size", 42.0, labels)

	// Verify the gauge value
	gauge := m.customGauges["queue_size"]
	assert.NotNil(t, gauge)

	value := testutil.ToFloat64(gauge.With(prometheus.Labels(labels)))
	assert.Equal(t, 42.0, value)

	// Update the gauge value
	m.SetCustomGauge("queue_size", 100.0, labels)
	value = testutil.ToFloat64(gauge.With(prometheus.Labels(labels)))
	assert.Equal(t, 100.0, value)
}

func TestObserveCustomHistogram(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Register a custom histogram
	buckets := []float64{0.1, 0.5, 1.0, 5.0, 10.0}
	err := m.RegisterCustomHistogram("processing_time", "Processing time in seconds", []string{"operation"}, buckets)
	require.NoError(t, err)

	// Observe values
	labels := map[string]string{"operation": "encode"}
	m.ObserveCustomHistogram("processing_time", 0.3, labels)
	m.ObserveCustomHistogram("processing_time", 0.7, labels)
	m.ObserveCustomHistogram("processing_time", 1.5, labels)

	// Verify observations were recorded by checking the metrics output
	handler := m.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "processing_time")
	assert.Contains(t, body, `operation="encode"`)
	// The histogram count should be 3
	assert.Contains(t, body, "processing_time_count")

}

func TestHandler(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Record some metrics
	m.RecordRequest("GET", "/api/test", 200, 50*time.Millisecond)
	m.UpdateRuntimeMetrics()

	// Get the handler
	handler := m.Handler()
	assert.NotNil(t, handler)

	// Create a test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Serve the metrics
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "glyphlang_http_requests_total")
	assert.Contains(t, body, "glyphlang_http_request_duration_seconds")
	assert.Contains(t, body, "glyphlang_runtime_goroutines")
	assert.Contains(t, body, "glyphlang_runtime_memory_alloc_bytes")
}

func TestGetRegistry(t *testing.T) {
	m := NewMetrics(DefaultConfig())
	registry := m.GetRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, m.registry, registry)
}

func TestMetricsMiddleware(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	tests := []struct {
		name           string
		method         string
		path           string
		handler        server.RouteHandler
		expectedStatus int
		shouldError    bool
	}{
		{
			name:   "successful request",
			method: "GET",
			path:   "/api/users",
			handler: func(ctx *server.Context) error {
				ctx.StatusCode = 200
				return nil
			},
			expectedStatus: 200,
			shouldError:    false,
		},
		{
			name:   "request with error",
			method: "POST",
			path:   "/api/users",
			handler: func(ctx *server.Context) error {
				ctx.StatusCode = 500
				return assert.AnError
			},
			expectedStatus: 500,
			shouldError:    true,
		},
		{
			name:   "request with default status",
			method: "GET",
			path:   "/api/health",
			handler: func(ctx *server.Context) error {
				// Don't set status code
				return nil
			},
			expectedStatus: 200,
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := MetricsMiddleware(m)

			// Wrap handler with middleware
			wrappedHandler := middleware(tt.handler)

			// Create test context
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			ctx := &server.Context{
				Request:        req,
				ResponseWriter: w,
				PathParams:     make(map[string]string),
				QueryParams:    make(map[string][]string),
				Body:           make(map[string]interface{}),
			}

			// Execute handler
			err := wrappedHandler(ctx)

			// Verify error handling
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Give a small delay for metrics to be recorded
			time.Sleep(10 * time.Millisecond)

			// Verify metrics were recorded by checking the handler output
			handler := m.Handler()
			metricsReq := httptest.NewRequest("GET", "/metrics", nil)
			metricsW := httptest.NewRecorder()
			handler.ServeHTTP(metricsW, metricsReq)

			body := metricsW.Body.String()
			assert.Contains(t, body, "glyphlang_http_requests_total")
			assert.Contains(t, body, tt.path)
		})
	}
}

func TestMetricsWithCustomConfig(t *testing.T) {
	config := Config{
		Namespace:       "custom",
		Subsystem:       "api",
		DurationBuckets: []float64{0.01, 0.1, 1.0},
	}

	m := NewMetrics(config)
	assert.NotNil(t, m)

	// Record a request
	m.RecordRequest("GET", "/test", 200, 50*time.Millisecond)

	// Get metrics
	handler := m.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "custom_api_requests_total")
	assert.Contains(t, body, "custom_api_request_duration_seconds")
}

func TestConcurrentMetricsRecording(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Record metrics concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				m.RecordRequest("GET", "/test", 200, time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify metrics were recorded
	handler := m.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "glyphlang_http_requests_total")

	// Count should be 1000 (10 goroutines * 100 requests)
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.Contains(line, `glyphlang_http_requests_total{method="GET",path="/test",status="200"}`) {
			assert.Contains(t, line, "1000")
			break
		}
	}
}

func TestMemoryMetricsAccuracy(t *testing.T) {
	m := NewMetrics(DefaultConfig())

	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Update metrics
	m.UpdateRuntimeMetrics()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Verify metrics are close to actual values (allow 10% tolerance due to timing)
	goroutines := testutil.ToFloat64(m.goroutines)
	actualGoroutines := float64(runtime.NumGoroutine())
	tolerance := actualGoroutines * 0.1
	assert.InDelta(t, actualGoroutines, goroutines, tolerance)

	memAlloc := testutil.ToFloat64(m.memoryAlloc)
	assert.InDelta(t, float64(memStats.Alloc), memAlloc, float64(memStats.Alloc)*0.1)
}

func BenchmarkRecordRequest(b *testing.B) {
	m := NewMetrics(DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RecordRequest("GET", "/api/test", 200, 50*time.Millisecond)
	}
}

func BenchmarkMetricsMiddleware(b *testing.B) {
	m := NewMetrics(DefaultConfig())
	middleware := MetricsMiddleware(m)

	handler := func(ctx *server.Context) error {
		ctx.StatusCode = 200
		return nil
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	ctx := &server.Context{
		Request:        req,
		ResponseWriter: w,
		PathParams:     make(map[string]string),
		QueryParams:    make(map[string][]string),
		Body:           make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrappedHandler(ctx)
	}
}
