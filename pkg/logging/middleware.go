package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
)

// StructuredLoggingMiddleware creates a middleware that logs HTTP requests with structured logging
func StructuredLoggingMiddleware(logger *Logger) server.Middleware {
	return func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) error {
			start := time.Now()

			// Get or generate request ID
			requestID := ctx.Request.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = NewRequestID()
				ctx.Request.Header.Set(RequestIDHeader, requestID)
			}

			// Set request ID in response header
			ctx.ResponseWriter.Header().Set(RequestIDHeader, requestID)

			// Create context logger
			ctxLogger := logger.WithRequestID(requestID).WithFields(map[string]interface{}{
				"method":     ctx.Request.Method,
				"path":       ctx.Request.URL.Path,
				"remote_ip":  ctx.Request.RemoteAddr,
				"user_agent": ctx.Request.UserAgent(),
			})

			// Log request start
			ctxLogger.InfoWithFields("request started", map[string]interface{}{
				"query": ctx.Request.URL.RawQuery,
			})

			// Capture response for logging
			responseCapture := &responseWriter{
				ResponseWriter: ctx.ResponseWriter,
				statusCode:     200, // Default status
				body:           &bytes.Buffer{},
			}
			ctx.ResponseWriter = responseCapture

			// Call next handler
			err := next(ctx)

			// Calculate duration
			duration := time.Since(start)

			// Determine status code
			statusCode := ctx.StatusCode
			if statusCode == 0 {
				statusCode = responseCapture.statusCode
			}
			if err != nil && statusCode < 400 {
				statusCode = 500
			}

			// Log response
			logFields := map[string]interface{}{
				"status":        statusCode,
				"duration_ms":   duration.Milliseconds(),
				"response_size": responseCapture.body.Len(),
			}

			// Determine log level based on status code
			if err != nil {
				logFields["error"] = err.Error()
				ctxLogger.ErrorWithFields("request failed", logFields)
			} else if statusCode >= 500 {
				ctxLogger.ErrorWithFields("request completed with server error", logFields)
			} else if statusCode >= 400 {
				ctxLogger.WarnWithFields("request completed with client error", logFields)
			} else {
				ctxLogger.InfoWithFields("request completed", logFields)
			}

			return err
		}
	}
}

// StructuredLoggingMiddlewareWithBodyLogging creates a middleware that logs request and response bodies
// WARNING: This can have significant performance impact and should only be used for debugging
func StructuredLoggingMiddlewareWithBodyLogging(logger *Logger, maxBodySize int) server.Middleware {
	if maxBodySize == 0 {
		maxBodySize = 10240 // 10KB default
	}

	return func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) error {
			start := time.Now()

			// Get or generate request ID
			requestID := ctx.Request.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = NewRequestID()
				ctx.Request.Header.Set(RequestIDHeader, requestID)
			}

			// Set request ID in response header
			ctx.ResponseWriter.Header().Set(RequestIDHeader, requestID)

			// Read and log request body
			var requestBody []byte
			if ctx.Request.Body != nil {
				requestBody, _ = io.ReadAll(io.LimitReader(ctx.Request.Body, int64(maxBodySize)))
				ctx.Request.Body.Close()
				// Restore body for handlers
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// Create context logger
			ctxLogger := logger.WithRequestID(requestID).WithFields(map[string]interface{}{
				"method":     ctx.Request.Method,
				"path":       ctx.Request.URL.Path,
				"remote_ip":  ctx.Request.RemoteAddr,
				"user_agent": ctx.Request.UserAgent(),
			})

			// Log request with body
			reqLogFields := map[string]interface{}{
				"query": ctx.Request.URL.RawQuery,
			}
			if len(requestBody) > 0 {
				// Try to parse as JSON, otherwise log as string
				var jsonBody interface{}
				if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
					reqLogFields["request_body"] = jsonBody
				} else {
					reqLogFields["request_body"] = string(requestBody)
				}
			}
			ctxLogger.InfoWithFields("request started", reqLogFields)

			// Capture response
			responseCapture := &responseWriter{
				ResponseWriter: ctx.ResponseWriter,
				statusCode:     200,
				body:           &bytes.Buffer{},
			}
			ctx.ResponseWriter = responseCapture

			// Call next handler
			err := next(ctx)

			// Calculate duration
			duration := time.Since(start)

			// Determine status code
			statusCode := ctx.StatusCode
			if statusCode == 0 {
				statusCode = responseCapture.statusCode
			}
			if err != nil && statusCode < 400 {
				statusCode = 500
			}

			// Log response with body
			respLogFields := map[string]interface{}{
				"status":        statusCode,
				"duration_ms":   duration.Milliseconds(),
				"response_size": responseCapture.body.Len(),
			}

			// Add response body if not too large
			if responseCapture.body.Len() > 0 && responseCapture.body.Len() <= maxBodySize {
				var jsonBody interface{}
				if err := json.Unmarshal(responseCapture.body.Bytes(), &jsonBody); err == nil {
					respLogFields["response_body"] = jsonBody
				} else {
					respLogFields["response_body"] = responseCapture.body.String()
				}
			}

			// Determine log level and message
			if err != nil {
				respLogFields["error"] = err.Error()
				ctxLogger.ErrorWithFields("request failed", respLogFields)
			} else if statusCode >= 500 {
				ctxLogger.ErrorWithFields("request completed with server error", respLogFields)
			} else if statusCode >= 400 {
				ctxLogger.WarnWithFields("request completed with client error", respLogFields)
			} else {
				ctxLogger.InfoWithFields("request completed", respLogFields)
			}

			return err
		}
	}
}

// StructuredRecoveryMiddleware recovers from panics and logs them with structured logging
func StructuredRecoveryMiddleware(logger *Logger) server.Middleware {
	return func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					// Get request ID if present
					requestID := ctx.Request.Header.Get(RequestIDHeader)

					// Create context logger
					ctxLogger := logger.WithRequestID(requestID).WithFields(map[string]interface{}{
						"method":     ctx.Request.Method,
						"path":       ctx.Request.URL.Path,
						"remote_ip":  ctx.Request.RemoteAddr,
						"panic":      r,
					})

					// Log panic with stack trace
					ctxLogger.Error("panic recovered")

					// Send error response and set error
					server.SendError(ctx, 500, "internal server error")
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()

			return next(ctx)
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response body
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body and writes to the underlying writer
func (rw *responseWriter) Write(b []byte) (int, error) {
	// Capture response body
	rw.body.Write(b)
	// Write to actual response
	return rw.ResponseWriter.Write(b)
}

// GetRequestLogger extracts or creates a context logger from the request
func GetRequestLogger(logger *Logger, r *http.Request) *ContextLogger {
	requestID := r.Header.Get(RequestIDHeader)
	if requestID == "" {
		requestID = NewRequestID()
	}

	return logger.WithRequestID(requestID).WithFields(map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})
}
