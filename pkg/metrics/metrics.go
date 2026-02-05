package metrics

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics collectors
type Metrics struct {
	// Request metrics
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestErrors   *prometheus.CounterVec

	// Resource usage metrics
	goroutines   prometheus.Gauge
	memoryAlloc  prometheus.Gauge
	memoryTotal  prometheus.Gauge
	memorySystem prometheus.Gauge
	gcPauseNs    prometheus.Gauge
	numGC        prometheus.Gauge

	// Custom business metrics
	customCounters   map[string]*prometheus.CounterVec
	customGauges     map[string]*prometheus.GaugeVec
	customHistograms map[string]*prometheus.HistogramVec

	registry *prometheus.Registry
}

// Config holds configuration for metrics
type Config struct {
	Namespace string
	Subsystem string
	// Custom histogram buckets for request duration (in seconds)
	DurationBuckets []float64
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Namespace: "glyphlang",
		Subsystem: "http",
		// Default buckets: 1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
		DurationBuckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(config Config) *Metrics {
	if config.Namespace == "" {
		config = DefaultConfig()
	}
	if len(config.DurationBuckets) == 0 {
		config.DurationBuckets = DefaultConfig().DurationBuckets
	}

	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry:         registry,
		customCounters:   make(map[string]*prometheus.CounterVec),
		customGauges:     make(map[string]*prometheus.GaugeVec),
		customHistograms: make(map[string]*prometheus.HistogramVec),
	}

	// Request rate metrics
	m.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Request latency metrics
	m.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds",
			Buckets:   config.DurationBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// Request error metrics
	m.requestErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "request_errors_total",
			Help:      "Total number of HTTP request errors by status code",
		},
		[]string{"method", "path", "status"},
	)

	// Resource usage metrics
	m.goroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "runtime",
			Name:      "goroutines",
			Help:      "Number of goroutines currently running",
		},
	)

	m.memoryAlloc = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "runtime",
			Name:      "memory_alloc_bytes",
			Help:      "Number of bytes allocated and still in use",
		},
	)

	m.memoryTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "runtime",
			Name:      "memory_total_alloc_bytes",
			Help:      "Total number of bytes allocated (cumulative)",
		},
	)

	m.memorySystem = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "runtime",
			Name:      "memory_sys_bytes",
			Help:      "Number of bytes obtained from system",
		},
	)

	m.gcPauseNs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "runtime",
			Name:      "gc_pause_ns",
			Help:      "Most recent GC pause time in nanoseconds",
		},
	)

	m.numGC = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "runtime",
			Name:      "gc_runs_total",
			Help:      "Total number of GC runs",
		},
	)

	// Register all metrics
	registry.MustRegister(
		m.requestsTotal,
		m.requestDuration,
		m.requestErrors,
		m.goroutines,
		m.memoryAlloc,
		m.memoryTotal,
		m.memorySystem,
		m.gcPauseNs,
		m.numGC,
	)

	// Start background goroutine to update runtime metrics
	go m.collectRuntimeMetrics()

	return m
}

// collectRuntimeMetrics periodically collects runtime metrics
func (m *Metrics) collectRuntimeMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.UpdateRuntimeMetrics()
	}
}

// UpdateRuntimeMetrics updates runtime metrics (goroutines, memory, GC)
func (m *Metrics) UpdateRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.goroutines.Set(float64(runtime.NumGoroutine()))
	m.memoryAlloc.Set(float64(memStats.Alloc))
	m.memoryTotal.Set(float64(memStats.TotalAlloc))
	m.memorySystem.Set(float64(memStats.Sys))
	m.numGC.Set(float64(memStats.NumGC))

	if memStats.NumGC > 0 {
		m.gcPauseNs.Set(float64(memStats.PauseNs[(memStats.NumGC+255)%256]))
	}
}

// RecordRequest records metrics for an HTTP request
func (m *Metrics) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)

	// Record total requests
	m.requestsTotal.WithLabelValues(method, path, status).Inc()

	// Record request duration
	m.requestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())

	// Record errors (status codes >= 400)
	if statusCode >= 400 {
		m.requestErrors.WithLabelValues(method, path, status).Inc()
	}
}

// RegisterCustomCounter registers a custom counter metric
func (m *Metrics) RegisterCustomCounter(name, help string, labels []string) error {
	if _, exists := m.customCounters[name]; exists {
		return prometheus.AlreadyRegisteredError{}
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	if err := m.registry.Register(counter); err != nil {
		return err
	}

	m.customCounters[name] = counter
	return nil
}

// RegisterCustomGauge registers a custom gauge metric
func (m *Metrics) RegisterCustomGauge(name, help string, labels []string) error {
	if _, exists := m.customGauges[name]; exists {
		return prometheus.AlreadyRegisteredError{}
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	if err := m.registry.Register(gauge); err != nil {
		return err
	}

	m.customGauges[name] = gauge
	return nil
}

// RegisterCustomHistogram registers a custom histogram metric
func (m *Metrics) RegisterCustomHistogram(name, help string, labels []string, buckets []float64) error {
	if _, exists := m.customHistograms[name]; exists {
		return prometheus.AlreadyRegisteredError{}
	}

	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)

	if err := m.registry.Register(histogram); err != nil {
		return err
	}

	m.customHistograms[name] = histogram
	return nil
}

// IncrementCustomCounter increments a custom counter
func (m *Metrics) IncrementCustomCounter(name string, labels map[string]string) {
	if counter, exists := m.customCounters[name]; exists {
		labelValues := make([]string, 0, len(labels))
		for _, v := range labels {
			labelValues = append(labelValues, v)
		}
		counter.With(prometheus.Labels(labels)).Inc()
	}
}

// SetCustomGauge sets a custom gauge value
func (m *Metrics) SetCustomGauge(name string, value float64, labels map[string]string) {
	if gauge, exists := m.customGauges[name]; exists {
		gauge.With(prometheus.Labels(labels)).Set(value)
	}
}

// ObserveCustomHistogram observes a value in a custom histogram
func (m *Metrics) ObserveCustomHistogram(name string, value float64, labels map[string]string) {
	if histogram, exists := m.customHistograms[name]; exists {
		histogram.With(prometheus.Labels(labels)).Observe(value)
	}
}

// Handler returns an HTTP handler for the /metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// GetRegistry returns the Prometheus registry
func (m *Metrics) GetRegistry() *prometheus.Registry {
	return m.registry
}
