package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

// MockInterpreter for benchmarking
type BenchmarkInterpreter struct {
	response interface{}
}

func (m *BenchmarkInterpreter) Execute(route *server.Route, ctx *server.Context) (interface{}, error) {
	return m.response, nil
}

// BenchmarkEndToEndHTTP benchmarks full HTTP request/response cycle
func BenchmarkEndToEndHTTP(b *testing.B) {
	// Create server with mock interpreter
	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: map[string]interface{}{
			"status": "ok",
			"data":   "Hello, World!",
		},
	}))

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/test",
	})

	// Create test server
	ts := httptest.NewServer(srv.GetHandler())
	defer ts.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(ts.URL + "/api/test")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// BenchmarkEndToEndHTTPWithJSON benchmarks full HTTP cycle with JSON parsing
func BenchmarkEndToEndHTTPWithJSON(b *testing.B) {
	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: map[string]interface{}{
			"id":      123,
			"name":    "Test User",
			"email":   "test@example.com",
			"active":  true,
			"created": time.Now().Format(time.RFC3339),
		},
	}))

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/users/:id",
	})

	ts := httptest.NewServer(srv.GetHandler())
	defer ts.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(ts.URL + "/api/users/123")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
	}
}

// BenchmarkHTTPHandlerDirect benchmarks handler without network overhead
func BenchmarkHTTPHandlerDirect(b *testing.B) {
	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: map[string]interface{}{
			"status": "ok",
		},
	}))

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/test",
	})

	handler := srv.GetHandler()
	req := httptest.NewRequest("GET", "/api/test", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkWithMiddlewareChain benchmarks request with full middleware chain
func BenchmarkWithMiddlewareChain(b *testing.B) {
	srv := server.NewServer(
		server.WithInterpreter(&BenchmarkInterpreter{
			response: map[string]interface{}{"status": "ok"},
		}),
		server.WithMiddleware(server.LoggingMiddleware()),
		server.WithMiddleware(server.RecoveryMiddleware()),
		server.WithMiddleware(server.SecurityHeadersMiddleware()),
		server.WithMiddleware(server.CORSMiddleware([]string{"*"})),
	)

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/test",
	})

	handler := srv.GetHandler()
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: map[string]interface{}{"status": "ok"},
	}))

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/test",
	})

	ts := httptest.NewServer(srv.GetHandler())
	defer ts.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(ts.URL + "/api/test")
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()
		}
	})
}

// LatencyPercentiles measures and reports latency percentiles
func BenchmarkLatencyPercentiles(b *testing.B) {
	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: map[string]interface{}{
			"id":    123,
			"name":  "Test User",
			"email": "test@example.com",
		},
	}))

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/users/:id",
	})

	handler := srv.GetHandler()
	req := httptest.NewRequest("GET", "/api/users/123", nil)

	// Warm up
	for i := 0; i < 1000; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Collect latencies
	latencies := make([]time.Duration, b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		start := time.Now()
		handler.ServeHTTP(w, req)
		latencies[i] = time.Since(start)
	}

	b.StopTimer()

	// Calculate percentiles
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	if len(latencies) > 0 {
		p50 := latencies[len(latencies)*50/100]
		p90 := latencies[len(latencies)*90/100]
		p99 := latencies[len(latencies)*99/100]
		max := latencies[len(latencies)-1]

		b.ReportMetric(float64(p50.Nanoseconds()), "p50-ns")
		b.ReportMetric(float64(p90.Nanoseconds()), "p90-ns")
		b.ReportMetric(float64(p99.Nanoseconds()), "p99-ns")
		b.ReportMetric(float64(max.Nanoseconds()), "max-ns")

		// Log percentiles for visibility
		b.Logf("Latency Percentiles: p50=%v, p90=%v, p99=%v, max=%v",
			p50, p90, p99, max)
	}
}

// BenchmarkRouterPerformance benchmarks route matching at scale
func BenchmarkRouterPerformance(b *testing.B) {
	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: map[string]interface{}{"status": "ok"},
	}))

	// Register many routes
	for i := 0; i < 100; i++ {
		srv.RegisterRoute(&server.Route{
			Method: server.GET,
			Path:   fmt.Sprintf("/api/v1/resource%d/:id", i),
		})
	}

	handler := srv.GetHandler()

	// Test matching the last route (worst case)
	req := httptest.NewRequest("GET", "/api/v1/resource99/123", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkJSONSerializationE2E benchmarks JSON response serialization end-to-end
func BenchmarkJSONSerializationE2E(b *testing.B) {
	// Large response object
	response := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "User 1", "email": "user1@example.com"},
			{"id": 2, "name": "User 2", "email": "user2@example.com"},
			{"id": 3, "name": "User 3", "email": "user3@example.com"},
			{"id": 4, "name": "User 4", "email": "user4@example.com"},
			{"id": 5, "name": "User 5", "email": "user5@example.com"},
		},
		"total":  5,
		"page":   1,
		"limit":  10,
		"hasMore": false,
	}

	srv := server.NewServer(server.WithInterpreter(&BenchmarkInterpreter{
		response: response,
	}))

	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/users",
	})

	handler := srv.GetHandler()
	req := httptest.NewRequest("GET", "/api/users", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkHealthCheck benchmarks health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	srv := server.NewServer()
	handler := srv.GetHandler()
	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
