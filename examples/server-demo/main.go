package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

func main() {
	// Create server with middleware
	srv := server.NewServer(
		server.WithAddr(":8080"),
		server.WithMiddleware(server.LoggingMiddleware()),
		server.WithMiddleware(server.RecoveryMiddleware()),
		server.WithInterpreter(&server.MockInterpreter{
			Response: map[string]interface{}{
				"status": "ok",
				"server": "Glyph Demo Server",
			},
		}),
	)

	// Register routes matching the GlyphLang examples

	// Simple hello world
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/hello",
		Handler: func(ctx *server.Context) error {
			return server.SendJSON(ctx, 200, map[string]interface{}{
				"message":   "Hello, World!",
				"timestamp": time.Now().Unix(),
			})
		},
	})

	// Greet with name parameter
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/greet/:name",
		Handler: func(ctx *server.Context) error {
			name := ctx.PathParams["name"]
			return server.SendJSON(ctx, 200, map[string]interface{}{
				"text":      "Hello, " + name + "!",
				"timestamp": time.Now().Unix(),
			})
		},
	})

	// Health check endpoint
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/health",
		Handler: func(ctx *server.Context) error {
			return server.SendJSON(ctx, 200, map[string]interface{}{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
			})
		},
	})

	// REST API routes for users

	// Get all users
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/users",
		Handler: func(ctx *server.Context) error {
			// Mock users
			users := []map[string]interface{}{
				{"id": 1, "name": "John Doe", "email": "john@example.com"},
				{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
			}

			// Support query params
			if page := ctx.QueryParams["page"]; page != "" {
				log.Printf("Filtering by page: %s", page)
			}

			return server.SendJSON(ctx, 200, users)
		},
	})

	// Get user by ID
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/users/:id",
		Handler: func(ctx *server.Context) error {
			id := ctx.PathParams["id"]

			// Mock user
			user := map[string]interface{}{
				"id":         id,
				"name":       "John Doe",
				"email":      "john@example.com",
				"created_at": time.Now().Unix(),
			}

			return server.SendJSON(ctx, 200, user)
		},
	})

	// Create new user
	srv.RegisterRoute(&server.Route{
		Method: server.POST,
		Path:   "/api/users",
		Handler: func(ctx *server.Context) error {
			// Get data from request body
			name, nameOk := ctx.Body["name"].(string)
			email, emailOk := ctx.Body["email"].(string)

			if !nameOk || !emailOk {
				return server.SendError(ctx, 400, "name and email are required")
			}

			// Mock creating user
			user := map[string]interface{}{
				"id":         123,
				"name":       name,
				"email":      email,
				"created_at": time.Now().Unix(),
			}

			return server.SendJSON(ctx, 201, user)
		},
	})

	// Update user
	srv.RegisterRoute(&server.Route{
		Method: server.PUT,
		Path:   "/api/users/:id",
		Handler: func(ctx *server.Context) error {
			id := ctx.PathParams["id"]

			// Mock updating user
			user := map[string]interface{}{
				"id":         id,
				"name":       ctx.Body["name"],
				"email":      ctx.Body["email"],
				"updated_at": time.Now().Unix(),
			}

			return server.SendJSON(ctx, 200, user)
		},
	})

	// Delete user
	srv.RegisterRoute(&server.Route{
		Method: server.DELETE,
		Path:   "/api/users/:id",
		Handler: func(ctx *server.Context) error {
			id := ctx.PathParams["id"]
			log.Printf("Deleting user: %s", id)

			return server.SendJSON(ctx, 200, map[string]interface{}{
				"success": true,
				"message": "User deleted",
			})
		},
	})

	// User posts - nested resource
	srv.RegisterRoute(&server.Route{
		Method: server.GET,
		Path:   "/api/users/:userId/posts/:postId",
		Handler: func(ctx *server.Context) error {
			userId := ctx.PathParams["userId"]
			postId := ctx.PathParams["postId"]

			post := map[string]interface{}{
				"id":      postId,
				"userId":  userId,
				"title":   "Sample Post",
				"content": "This is a sample post content",
			}

			return server.SendJSON(ctx, 200, post)
		},
	})

	// PATCH example
	srv.RegisterRoute(&server.Route{
		Method: server.PATCH,
		Path:   "/api/users/:id",
		Handler: func(ctx *server.Context) error {
			id := ctx.PathParams["id"]

			// Mock partial update
			result := map[string]interface{}{
				"id":      id,
				"updated": ctx.Body,
			}

			return server.SendJSON(ctx, 200, result)
		},
	})

	// Start server in goroutine
	go func() {
		log.Printf("Starting Glyph Demo Server on http://localhost:8080")
		log.Println("\nAvailable endpoints:")
		log.Println("  GET    /hello")
		log.Println("  GET    /greet/:name")
		log.Println("  GET    /health")
		log.Println("  GET    /api/users")
		log.Println("  GET    /api/users/:id")
		log.Println("  POST   /api/users")
		log.Println("  PUT    /api/users/:id")
		log.Println("  PATCH  /api/users/:id")
		log.Println("  DELETE /api/users/:id")
		log.Println("  GET    /api/users/:userId/posts/:postId")
		log.Println("\nPress Ctrl+C to stop")

		if err := srv.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("\nShutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
