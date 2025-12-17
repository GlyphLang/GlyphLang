package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

// SimulatedDatabase simulates a database with health checking
type SimulatedDatabase struct {
	connected bool
	latency   time.Duration
}

func NewSimulatedDatabase() *SimulatedDatabase {
	return &SimulatedDatabase{
		connected: true,
		latency:   20 * time.Millisecond,
	}
}

func (db *SimulatedDatabase) Ping(ctx context.Context) error {
	if !db.connected {
		return fmt.Errorf("database not connected")
	}

	// Simulate network latency
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(db.latency):
		return nil
	}
}

// SimulatedCache simulates a cache service with health checking
type SimulatedCache struct {
	healthy bool
}

func NewSimulatedCache() *SimulatedCache {
	return &SimulatedCache{healthy: true}
}

func (c *SimulatedCache) Ping(ctx context.Context) error {
	if !c.healthy {
		return fmt.Errorf("cache service unavailable")
	}
	return nil
}

func main() {
	fmt.Println("GLYPHLANG Health Check Demo")
	fmt.Println("===========================")

	// Create simulated dependencies
	db := NewSimulatedDatabase()
	cache := NewSimulatedCache()

	// Create health manager with custom timeout
	healthManager := server.NewHealthManager(
		server.WithHealthCheckTimeout(5 * time.Second),
	)

	// Register database health checker
	dbChecker := server.NewDatabaseHealthChecker("database", db.Ping)
	healthManager.RegisterChecker(dbChecker)

	// Register cache health checker
	cacheChecker := server.NewHealthCheckFunc("cache", func(ctx context.Context) *server.CheckResult {
		start := time.Now()
		err := cache.Ping(ctx)
		latency := time.Since(start).Milliseconds()

		if err != nil {
			return &server.CheckResult{
				Status:    server.StatusUnhealthy,
				LatencyMs: latency,
				Error:     err.Error(),
			}
		}

		return &server.CheckResult{
			Status:    server.StatusHealthy,
			LatencyMs: latency,
		}
	})
	healthManager.RegisterChecker(cacheChecker)

	// Register a custom application health checker
	appChecker := server.NewHealthCheckFunc("application", func(ctx context.Context) *server.CheckResult {
		// In a real application, check various app-specific conditions
		// For example: memory usage, goroutine count, etc.

		memoryUsage := 45.2 // Simulated memory usage percentage
		status := server.StatusHealthy
		message := ""

		if memoryUsage > 80 {
			status = server.StatusDegraded
			message = fmt.Sprintf("high memory usage: %.1f%%", memoryUsage)
		} else if memoryUsage > 95 {
			status = server.StatusUnhealthy
			message = fmt.Sprintf("critical memory usage: %.1f%%", memoryUsage)
		}

		return &server.CheckResult{
			Status:  status,
			Message: message,
		}
	})
	healthManager.RegisterChecker(appChecker)

	// Create GLYPHLANG server
	srv := server.NewServer(
		server.WithAddr(":8080"),
	)

	// Register health check routes
	if err := srv.RegisterHealthRoutes(healthManager); err != nil {
		log.Fatalf("Failed to register health routes: %v", err)
	}

	// Register a sample API route to demonstrate a working application
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/hello",
		Handler: func(ctx *server.Context) error {
			return server.SendJSON(ctx, http.StatusOK, map[string]interface{}{
				"message": "Hello from GLYPHLANG!",
				"time":    time.Now().Format(time.RFC3339),
			})
		},
	})

	// Register a route to simulate dependency failures (for demo purposes)
	srv.RegisterRoute(&server.Route{
		Method: server.POST,
		Path:   "/admin/simulate/:service/:status",
		Handler: func(ctx *server.Context) error {
			service := ctx.PathParams["service"]
			status := ctx.PathParams["status"]

			switch service {
			case "database":
				if status == "down" {
					db.connected = false
				} else if status == "up" {
					db.connected = true
				} else if status == "slow" {
					db.latency = 200 * time.Millisecond
				} else if status == "fast" {
					db.latency = 20 * time.Millisecond
				}
			case "cache":
				if status == "down" {
					cache.healthy = false
				} else if status == "up" {
					cache.healthy = true
				}
			default:
				return server.SendError(ctx, http.StatusBadRequest, "unknown service")
			}

			return server.SendJSON(ctx, http.StatusOK, map[string]interface{}{
				"service": service,
				"status":  status,
				"message": fmt.Sprintf("Simulated %s status changed to %s", service, status),
			})
		},
	})

	// Print available endpoints
	fmt.Println("\nAvailable Endpoints:")
	fmt.Println("  Health Checks:")
	fmt.Println("    GET  http://localhost:8080/health/live   - Liveness probe")
	fmt.Println("    GET  http://localhost:8080/health/ready  - Readiness probe")
	fmt.Println("    GET  http://localhost:8080/health        - Detailed health status")
	fmt.Println("\n  Application:")
	fmt.Println("    GET  http://localhost:8080/api/hello     - Sample API endpoint")
	fmt.Println("\n  Admin (Demo):")
	fmt.Println("    POST http://localhost:8080/admin/simulate/database/down  - Simulate DB failure")
	fmt.Println("    POST http://localhost:8080/admin/simulate/database/up    - Restore DB")
	fmt.Println("    POST http://localhost:8080/admin/simulate/database/slow  - Simulate slow DB")
	fmt.Println("    POST http://localhost:8080/admin/simulate/cache/down     - Simulate cache failure")
	fmt.Println("    POST http://localhost:8080/admin/simulate/cache/up       - Restore cache")

	// Start a goroutine to periodically check and display health status
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			displayHealthStatus()
		}
	}()

	// Start the server in a goroutine
	go func() {
		fmt.Printf("\nStarting server on :8080...\n\n")
		if err := srv.Start(""); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for initial startup
	time.Sleep(500 * time.Millisecond)

	// Display initial health status
	displayHealthStatus()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n\nShutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	fmt.Println("Server stopped")
}

func displayHealthStatus() {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		log.Printf("Failed to fetch health status: %v", err)
		return
	}
	defer resp.Body.Close()

	var health server.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		log.Printf("Failed to decode health response: %v", err)
		return
	}

	fmt.Println("\n" + getStatusIcon(health.Status) + " Overall Status: " + string(health.Status))
	fmt.Println("Component Health:")
	for name, result := range health.Checks {
		icon := getStatusIcon(result.Status)
		fmt.Printf("  %s %-12s: %s (latency: %dms)", icon, name, result.Status, result.LatencyMs)
		if result.Message != "" {
			fmt.Printf(" - %s", result.Message)
		}
		if result.Error != "" {
			fmt.Printf(" - ERROR: %s", result.Error)
		}
		fmt.Println()
	}
	fmt.Printf("Last checked: %s\n", health.Timestamp.Format(time.RFC3339))
}

func getStatusIcon(status server.HealthStatus) string {
	switch status {
	case server.StatusHealthy:
		return "✓"
	case server.StatusDegraded:
		return "⚠"
	case server.StatusUnhealthy:
		return "✗"
	default:
		return "?"
	}
}

// Example usage scenarios:
//
// 1. Check liveness (Kubernetes liveness probe):
//    curl http://localhost:8080/health/live
//
// 2. Check readiness (Kubernetes readiness probe):
//    curl http://localhost:8080/health/ready
//
// 3. Get detailed health status:
//    curl http://localhost:8080/health
//
// 4. Simulate database failure:
//    curl -X POST http://localhost:8080/admin/simulate/database/down
//    Then check: curl http://localhost:8080/health/ready
//    (Should return 503 Service Unavailable)
//
// 5. Restore database:
//    curl -X POST http://localhost:8080/admin/simulate/database/up
//
// 6. Simulate slow database:
//    curl -X POST http://localhost:8080/admin/simulate/database/slow
//    (Database latency will be marked as degraded)
