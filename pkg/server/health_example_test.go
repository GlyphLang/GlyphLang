package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

// ExampleHealthManager demonstrates basic health check setup
func ExampleHealthManager() {
	// Create a health manager
	hm := server.NewHealthManager(
		server.WithHealthCheckTimeout(5 * time.Second),
	)

	// Register a custom health checker
	hm.RegisterChecker(server.NewHealthCheckFunc("my-service", func(ctx context.Context) *server.CheckResult {
		// Perform your health check logic here
		return &server.CheckResult{
			Status:    server.StatusHealthy,
			LatencyMs: 5,
		}
	}))

	// Create server and register health routes
	srv := server.NewServer()
	if err := srv.RegisterHealthRoutes(hm); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Health routes registered: /health/live, /health/ready, /health")
	// Output: Health routes registered: /health/live, /health/ready, /health
}

// ExampleHealthManager_databaseChecker demonstrates database health checking
func ExampleHealthManager_databaseChecker() {
	// Create a health manager
	hm := server.NewHealthManager()

	// Simulate a database connection (in real code, use your actual DB)
	var db *sql.DB // This would be your actual database connection

	// Register a database health checker
	dbChecker := server.NewDatabaseHealthChecker("database", func(ctx context.Context) error {
		// In real code, use db.PingContext(ctx)
		if db == nil {
			return nil // For this example
		}
		return db.PingContext(ctx)
	})

	hm.RegisterChecker(dbChecker)

	fmt.Println("Database health checker registered")
	// Output: Database health checker registered
}

// ExampleHealthManager_httpChecker demonstrates external service health checking
func ExampleHealthManager_httpChecker() {
	// Create a health manager
	hm := server.NewHealthManager()

	// Register an HTTP service health checker
	apiChecker, err := server.NewHTTPHealthChecker("external-api", "https://api.example.com/health")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	hm.RegisterChecker(apiChecker)

	fmt.Println("External API health checker registered")
	// Output: External API health checker registered
}

// ExampleHealthManager_customChecker demonstrates creating a custom health checker
func ExampleHealthManager_customChecker() {
	// Create a health manager
	hm := server.NewHealthManager()

	// Create a custom health checker with specific logic
	customChecker := server.NewHealthCheckFunc("redis", func(ctx context.Context) *server.CheckResult {
		start := time.Now()

		// Perform your custom health check logic
		// For example, check Redis connection
		// err := redisClient.Ping(ctx).Err()

		latency := time.Since(start).Milliseconds()

		// Simulate a healthy check
		return &server.CheckResult{
			Status:    server.StatusHealthy,
			LatencyMs: latency,
			Message:   "Redis is operational",
		}
	})

	hm.RegisterChecker(customChecker)

	fmt.Println("Custom Redis health checker registered")
	// Output: Custom Redis health checker registered
}

// ExampleHealthManager_multipleCheckers demonstrates using multiple health checkers
func ExampleHealthManager_multipleCheckers() {
	// Create a health manager
	hm := server.NewHealthManager()

	// Register multiple checkers
	hm.RegisterChecker(server.NewStaticHealthChecker("database", server.StatusHealthy))
	hm.RegisterChecker(server.NewStaticHealthChecker("redis", server.StatusHealthy))
	hm.RegisterChecker(server.NewStaticHealthChecker("elasticsearch", server.StatusHealthy))

	fmt.Println("3 health checkers registered")
	// Output: 3 health checkers registered
}

// ExampleHealthManager_standaloneHTTP demonstrates using health handlers without the GLYPHLANG server
func ExampleHealthManager_standaloneHTTP() {
	// Create a health manager
	hm := server.NewHealthManager()

	// Register some checkers
	hm.RegisterChecker(server.NewStaticHealthChecker("db", server.StatusHealthy))

	// Use with standard http.ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", server.LivenessHTTPHandler())
	mux.HandleFunc("/health/ready", server.ReadinessHTTPHandler(hm))

	fmt.Println("Health endpoints registered with http.ServeMux")
	// Output: Health endpoints registered with http.ServeMux
}

// ExampleHealthManager_degradedStatus demonstrates handling degraded service status
func ExampleHealthManager_degradedStatus() {
	// Create a custom checker that returns degraded status
	checker := server.NewHealthCheckFunc("slow-service", func(ctx context.Context) *server.CheckResult {
		// Simulate a slow but functional service
		return &server.CheckResult{
			Status:    server.StatusDegraded,
			LatencyMs: 250,
			Message:   "Service is slow but operational",
		}
	})

	fmt.Println(checker.Name())
	result := checker.Check(context.Background())
	fmt.Println(result.Status)
	// Output:
	// slow-service
	// degraded
}

// ExampleHealthManager_timeout demonstrates health check with timeout
func ExampleHealthManager_timeout() {
	// Create a health manager with short timeout
	hm := server.NewHealthManager(
		server.WithHealthCheckTimeout(100 * time.Millisecond),
	)

	// Register a checker that respects context timeout
	hm.RegisterChecker(server.NewHealthCheckFunc("service", func(ctx context.Context) *server.CheckResult {
		select {
		case <-ctx.Done():
			return &server.CheckResult{
				Status: server.StatusUnhealthy,
				Error:  "timeout exceeded",
			}
		case <-time.After(50 * time.Millisecond):
			return &server.CheckResult{
				Status: server.StatusHealthy,
			}
		}
	}))

	fmt.Println("Health manager configured with 100ms timeout")
	// Output: Health manager configured with 100ms timeout
}
