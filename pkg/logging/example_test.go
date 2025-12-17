package logging_test

import (
	"fmt"
	"io"
	"os"

	"github.com/glyphlang/glyph/pkg/logging"
	"github.com/glyphlang/glyph/pkg/server"
)

func ExampleLogger_basic() {
	// Create a basic logger
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel: logging.INFO,
		Format:   logging.TextFormat,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Simple logging
	logger.Info("Application started")
	logger.Warn("This is a warning")
	logger.Error("This is an error")
}

func ExampleLogger_withFields() {
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel: logging.DEBUG,
		Format:   logging.JSONFormat,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Logging with structured fields
	logger.InfoWithFields("User logged in", map[string]interface{}{
		"user_id":  12345,
		"username": "john_doe",
		"ip":       "192.168.1.1",
	})

	logger.ErrorWithFields("Database query failed", map[string]interface{}{
		"query":    "SELECT * FROM users",
		"duration": "5.2s",
		"error":    "connection timeout",
	})
}

func ExampleContextLogger() {
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel: logging.DEBUG,
		Format:   logging.JSONFormat,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Create a context logger with request ID
	requestID := logging.NewRequestID()
	ctxLogger := logger.WithRequestID(requestID)

	// All logs from this logger will include the request ID
	ctxLogger.Info("Processing request")
	ctxLogger.Info("Request completed")

	// Chain additional context
	userLogger := ctxLogger.
		WithField("user_id", 123).
		WithField("session", "abc123")

	userLogger.Info("User action performed")
}

func ExampleLogger_fileLogging() {
	// Create a logger with file rotation
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel:    logging.INFO,
		Format:      logging.JSONFormat,
		FilePath:    "/var/log/myapp/app.log",
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		MaxBackups:  5,                 // Keep 5 old files
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Info("This will be logged to file with rotation")
}

func ExampleLogger_multipleOutputs() {
	// Create a logger that writes to both stdout and a file
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel: logging.INFO,
		Format:   logging.TextFormat,
		Outputs:  []io.Writer{os.Stdout, file},
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Info("This goes to both stdout and file")
}

func ExampleStructuredLoggingMiddleware() {
	// Create logger
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel:      logging.INFO,
		Format:        logging.JSONFormat,
		IncludeCaller: true,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Create router
	router := server.NewRouter()

	// Create middlewares
	loggingMW := logging.StructuredLoggingMiddleware(logger)
	recoveryMW := logging.StructuredRecoveryMiddleware(logger)

	// Add route with middlewares
	handler := func(ctx *server.Context) error {
		// Get logger for this request
		reqLogger := logging.GetRequestLogger(logger, ctx.Request)

		userID := ctx.PathParams["id"]
		reqLogger.InfoWithFields("Fetching user", map[string]interface{}{
			"user_id": userID,
		})

		// Simulate user lookup
		user := map[string]interface{}{
			"id":   userID,
			"name": "John Doe",
		}

		return server.SendJSON(ctx, 200, user)
	}

	// Register route with middlewares
	router.RegisterRoute(&server.Route{
		Method:      server.GET,
		Path:        "/users/:id",
		Handler:     handler,
		Middlewares: []server.Middleware{loggingMW, recoveryMW},
	})

	// Start server
	fmt.Println("Server starting on :8080")
	// Example: start HTTP server with router
}

func ExampleStructuredLoggingMiddlewareWithBodyLogging() {
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel: logging.DEBUG,
		Format:   logging.JSONFormat,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	router := server.NewRouter()

	// Create middleware with body logging (for debugging only)
	// WARNING: This logs request and response bodies - use only in development
	bodyLoggingMW := logging.StructuredLoggingMiddlewareWithBodyLogging(logger, 10240)

	handler := func(ctx *server.Context) error {
		// Handler implementation
		return server.SendJSON(ctx, 201, map[string]interface{}{
			"id":      "123",
			"message": "User created",
		})
	}

	// Register route with middleware
	router.RegisterRoute(&server.Route{
		Method:      server.POST,
		Path:        "/api/users",
		Handler:     handler,
		Middlewares: []server.Middleware{bodyLoggingMW},
	})
}

func ExampleLogger_withDefaultLogger() {
	// Initialize the default logger
	err := logging.InitDefaultLogger(logging.LoggerConfig{
		MinLevel:      logging.INFO,
		Format:        logging.JSONFormat,
		IncludeCaller: true,
	})
	if err != nil {
		panic(err)
	}

	// Use convenience functions anywhere in your code
	logging.Info("Application started")
	logging.Warn("Low disk space")

	// Create context loggers
	reqLogger := logging.WithRequestID("req-123")
	reqLogger.Info("Processing request")

	// Add fields
	userLogger := logging.WithFields(map[string]interface{}{
		"user": "john",
		"role": "admin",
	})
	userLogger.Info("User action")
}

func ExampleLogger_differentLevels() {
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel: logging.DEBUG,
		Format:   logging.TextFormat,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// Different log levels
	logger.Debug("Detailed debug information")
	logger.Info("Informational message")
	logger.Warn("Warning: something might be wrong")
	logger.Error("Error occurred but application continues")
	// logger.Fatal("Critical error - application will exit")
}

func ExampleLogger_productionConfig() {
	// Production-ready configuration
	logger, err := logging.NewLogger(logging.LoggerConfig{
		MinLevel:          logging.INFO,
		Format:            logging.JSONFormat,
		IncludeCaller:     true,
		IncludeStackTrace: true,
		BufferSize:        5000,
		FilePath:          "/var/log/myapp/app.log",
		MaxFileSize:       50 * 1024 * 1024, // 50MB
		MaxBackups:        10,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Info("Production application started")
}
