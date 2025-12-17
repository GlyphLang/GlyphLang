package metrics

// This file contains integration examples showing how to use the metrics package
// with the GLYPHLANG server. These are not executable examples but serve as
// documentation for common integration patterns.

/*
Example 1: Basic Integration with Server

	package main

	import (
		"log"
		"net/http"

		"github.com/glyphlang/glyph/pkg/metrics"
		"github.com/glyphlang/glyph/pkg/server"
	)

	func main() {
		// Create metrics instance
		m := metrics.NewMetrics(metrics.DefaultConfig())

		// Create metrics middleware
		metricsMiddleware := metrics.MetricsMiddleware(m)
		loggingMiddleware := server.LoggingMiddleware()

		// Combine middlewares
		middleware := server.ChainMiddlewares(loggingMiddleware, metricsMiddleware)

		// Create a handler
		handler := func(ctx *server.Context) error {
			return server.SendJSON(ctx, 200, map[string]string{
				"message": "Hello, World!",
			})
		}

		// Apply middleware to handler
		wrappedHandler := middleware(handler)

		// Setup HTTP server
		http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
			ctx := &server.Context{
				Request:        r,
				ResponseWriter: w,
				PathParams:     make(map[string]string),
				QueryParams:    make(map[string]string),
				Body:           make(map[string]interface{}),
			}
			if err := wrappedHandler(ctx); err != nil {
				log.Printf("Error: %v", err)
			}
		})

		// Expose metrics endpoint
		http.Handle("/metrics", m.Handler())

		log.Println("Server starting on :8080")
		log.Println("Metrics available at http://localhost:8080/metrics")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

Example 2: Custom Business Metrics

	package main

	import (
		"github.com/glyphlang/glyph/pkg/metrics"
		"github.com/glyphlang/glyph/pkg/server"
	)

	func setupMetrics() *metrics.Metrics {
		m := metrics.NewMetrics(metrics.DefaultConfig())

		// Register custom counters
		m.RegisterCustomCounter(
			"user_registrations_total",
			"Total number of user registrations",
			[]string{"plan", "country"},
		)

		m.RegisterCustomCounter(
			"api_calls_by_user_total",
			"API calls grouped by user type",
			[]string{"user_type", "endpoint"},
		)

		// Register custom gauges
		m.RegisterCustomGauge(
			"active_websocket_connections",
			"Number of active WebSocket connections",
			[]string{"room"},
		)

		m.RegisterCustomGauge(
			"pending_jobs",
			"Number of pending background jobs",
			[]string{"job_type"},
		)

		// Register custom histograms
		m.RegisterCustomHistogram(
			"db_query_duration_seconds",
			"Database query duration in seconds",
			[]string{"query_type", "table"},
			[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		)

		return m
	}

	func userRegistrationHandler(m *metrics.Metrics) server.RouteHandler {
		return func(ctx *server.Context) error {
			// Your registration logic here...

			// Track the registration
			m.IncrementCustomCounter("user_registrations_total", map[string]string{
				"plan":    "premium",
				"country": "US",
			})

			return server.SendJSON(ctx, 201, map[string]string{
				"status": "registered",
			})
		}
	}

Example 3: Full Server Setup with All Middleware

	package main

	import (
		"log"
		"net/http"
		"time"

		"github.com/glyphlang/glyph/pkg/metrics"
		"github.com/glyphlang/glyph/pkg/server"
	)

	func main() {
		// Create metrics
		m := metrics.NewMetrics(metrics.Config{
			Namespace: "myapp",
			Subsystem: "api",
			DurationBuckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05,
				0.1, 0.25, 0.5, 1, 2.5, 5, 10,
			},
		})

		// Register custom metrics
		m.RegisterCustomCounter("login_attempts_total", "Login attempts", []string{"status"})
		m.RegisterCustomGauge("active_sessions", "Active user sessions", []string{})

		// Create middleware stack
		middleware := server.ChainMiddlewares(
			server.RecoveryMiddleware(),
			server.LoggingMiddleware(),
			metrics.MetricsMiddleware(m),
			server.CORSMiddleware([]string{"*"}),
			server.TimeoutMiddleware(30*time.Second),
		)

		// Define routes
		routes := map[string]server.RouteHandler{
			"/api/health": healthHandler(),
			"/api/login":  loginHandler(m),
			"/api/data":   dataHandler(m),
		}

		// Setup HTTP handlers
		for path, handler := range routes {
			wrappedHandler := middleware(handler)
			http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				ctx := &server.Context{
					Request:        r,
					ResponseWriter: w,
					PathParams:     make(map[string]string),
					QueryParams:    make(map[string]string),
					Body:           make(map[string]interface{}),
				}
				if err := wrappedHandler(ctx); err != nil {
					log.Printf("Error: %v", err)
				}
			})
		}

		// Expose metrics endpoint (without middleware)
		http.Handle("/metrics", m.Handler())

		log.Println("Server starting on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

	func healthHandler() server.RouteHandler {
		return func(ctx *server.Context) error {
			return server.SendJSON(ctx, 200, map[string]string{
				"status": "healthy",
			})
		}
	}

	func loginHandler(m *metrics.Metrics) server.RouteHandler {
		return func(ctx *server.Context) error {
			// Login logic here...

			// Track login attempt
			m.IncrementCustomCounter("login_attempts_total", map[string]string{
				"status": "success",
			})

			return server.SendJSON(ctx, 200, map[string]string{
				"token": "example-token",
			})
		}
	}

	func dataHandler(m *metrics.Metrics) server.RouteHandler {
		return func(ctx *server.Context) error {
			// Simulate database query
			start := time.Now()

			// Your data fetching logic...
			time.Sleep(10 * time.Millisecond)

			duration := time.Since(start)
			m.ObserveCustomHistogram("db_query_duration_seconds",
				duration.Seconds(),
				map[string]string{
					"query_type": "select",
					"table":      "users",
				},
			)

			return server.SendJSON(ctx, 200, map[string]interface{}{
				"data": []string{"item1", "item2"},
			})
		}
	}

Example 4: Monitoring Background Tasks

	package main

	import (
		"time"

		"github.com/glyphlang/glyph/pkg/metrics"
	)

	type BackgroundWorker struct {
		metrics *metrics.Metrics
	}

	func NewBackgroundWorker(m *metrics.Metrics) *BackgroundWorker {
		// Register metrics for background tasks
		m.RegisterCustomCounter(
			"background_jobs_processed_total",
			"Total background jobs processed",
			[]string{"job_type", "status"},
		)

		m.RegisterCustomHistogram(
			"background_job_duration_seconds",
			"Background job processing time",
			[]string{"job_type"},
			[]float64{0.1, 0.5, 1, 5, 10, 30, 60},
		)

		m.RegisterCustomGauge(
			"background_jobs_pending",
			"Number of pending background jobs",
			[]string{"job_type"},
		)

		return &BackgroundWorker{metrics: m}
	}

	func (w *BackgroundWorker) ProcessJob(jobType string, job func() error) {
		start := time.Now()

		// Update pending jobs
		w.metrics.SetCustomGauge("background_jobs_pending", 1.0, map[string]string{
			"job_type": jobType,
		})

		// Process job
		err := job()

		// Record metrics
		duration := time.Since(start)
		w.metrics.ObserveCustomHistogram("background_job_duration_seconds",
			duration.Seconds(),
			map[string]string{"job_type": jobType},
		)

		status := "success"
		if err != nil {
			status = "error"
		}

		w.metrics.IncrementCustomCounter("background_jobs_processed_total",
			map[string]string{
				"job_type": jobType,
				"status":   status,
			},
		)

		// Clear pending
		w.metrics.SetCustomGauge("background_jobs_pending", 0.0, map[string]string{
			"job_type": jobType,
		})
	}
*/
