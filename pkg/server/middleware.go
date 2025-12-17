package server

import (
	"fmt"
	"log"
	"time"
)

// LoggingMiddleware logs request details
func LoggingMiddleware() Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			start := time.Now()

			// Log request
			log.Printf("[REQUEST] %s %s", ctx.Request.Method, ctx.Request.URL.Path)

			// Call next handler
			err := next(ctx)

			// Log response
			duration := time.Since(start)
			status := ctx.StatusCode
			if err != nil {
				status = 500
			}

			log.Printf("[RESPONSE] %s %s - %d (%v)",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				status,
				duration,
			)

			return err
		}
	}
}

// RecoveryMiddleware recovers from panics and returns 500 error
func RecoveryMiddleware() Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] %s %s: %v", ctx.Request.Method, ctx.Request.URL.Path, r)
					// Create a proper InternalError with panic details
					panicErr := &InternalError{
						BaseError: &BaseError{
							Code:   500,
							Type:   "InternalError",
							Msg:    "internal server error",
							Detail: fmt.Sprintf("panic recovered: %v", r),
						},
					}
					// Send the error response
					SendHTTPError(ctx, panicErr)
					// Return the error to indicate a panic was recovered
					err = panicErr
				}
			}()

			return next(ctx)
		}
	}
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			origin := ctx.Request.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if origin != "" {
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", origin)
				} else if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" {
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
				}

				ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				ctx.ResponseWriter.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight requests
			if ctx.Request.Method == "OPTIONS" {
				ctx.StatusCode = 204
				ctx.ResponseWriter.WriteHeader(204)
				return nil
			}

			return next(ctx)
		}
	}
}

// HeaderMiddleware adds custom headers to all responses
func HeaderMiddleware(headers map[string]string) Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			for key, value := range headers {
				ctx.ResponseWriter.Header().Set(key, value)
			}
			return next(ctx)
		}
	}
}

// ChainMiddlewares combines multiple middlewares into one
func ChainMiddlewares(middlewares ...Middleware) Middleware {
	return func(next RouteHandler) RouteHandler {
		// Apply middlewares in reverse order
		handler := next
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// AuthMiddleware is a placeholder for authentication middleware
// In production, this would validate JWT tokens, API keys, or session tokens
func AuthMiddleware(validateFunc func(*Context) (bool, error)) Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			// Extract token from Authorization header
			token := ctx.Request.Header.Get("Authorization")
			if token == "" {
				return SendError(ctx, 401, "unauthorized: missing authorization header")
			}

			// Validate token using the provided function
			if validateFunc != nil {
				valid, err := validateFunc(ctx)
				if err != nil {
					log.Printf("[AUTH] Validation error: %v", err)
					return SendError(ctx, 401, "unauthorized: invalid token")
				}
				if !valid {
					return SendError(ctx, 401, "unauthorized: authentication failed")
				}
			}

			return next(ctx)
		}
	}
}

// BasicAuthMiddleware provides simple token-based authentication
func BasicAuthMiddleware(validTokens map[string]bool) Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			token := ctx.Request.Header.Get("Authorization")
			if token == "" {
				return SendError(ctx, 401, "unauthorized: missing authorization header")
			}

			// Remove "Bearer " prefix if present
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			// Check if token is valid
			if validTokens != nil && !validTokens[token] {
				return SendError(ctx, 401, "unauthorized: invalid token")
			}

			return next(ctx)
		}
	}
}

// RateLimitMiddleware is a placeholder for rate limiting middleware
// In production, this would use a proper rate limiter (e.g., token bucket, Redis)
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// RateLimitMiddleware implements simple in-memory rate limiting
func RateLimitMiddleware(config RateLimiterConfig) Middleware {
	// Simple in-memory store (not production-ready, use Redis in production)
	type clientLimit struct {
		tokens       int
		lastRefill   time.Time
		requestCount int
	}

	limits := make(map[string]*clientLimit)

	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			// Get client identifier (IP address)
			clientIP := ctx.Request.RemoteAddr

			// Get or create limit for this client
			limit, exists := limits[clientIP]
			if !exists {
				limit = &clientLimit{
					tokens:     config.BurstSize,
					lastRefill: time.Now(),
				}
				limits[clientIP] = limit
			}

			// Refill tokens based on time passed
			now := time.Now()
			elapsed := now.Sub(limit.lastRefill)
			tokensToAdd := int(elapsed.Minutes() * float64(config.RequestsPerMinute))

			if tokensToAdd > 0 {
				limit.tokens += tokensToAdd
				if limit.tokens > config.BurstSize {
					limit.tokens = config.BurstSize
				}
				limit.lastRefill = now
			}

			// Check if request is allowed
			if limit.tokens <= 0 {
				log.Printf("[RATE_LIMIT] Rate limit exceeded for %s", clientIP)
				return SendError(ctx, 429, "rate limit exceeded")
			}

			// Consume a token
			limit.tokens--
			limit.requestCount++

			return next(ctx)
		}
	}
}

// TimeoutMiddleware adds a timeout to request processing
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			done := make(chan error, 1)

			go func() {
				done <- next(ctx)
			}()

			select {
			case err := <-done:
				return err
			case <-time.After(timeout):
				log.Printf("[TIMEOUT] Request timeout after %v: %s %s",
					timeout, ctx.Request.Method, ctx.Request.URL.Path)
				return SendError(ctx, 504, "request timeout")
			}
		}
	}
}

// TracingMiddleware creates a middleware that adds OpenTelemetry distributed tracing
// This middleware integrates with the pkg/tracing package
// It should be added early in the middleware chain to trace the entire request lifecycle
//
// Usage:
//   import "github.com/glyphlang/glyph/pkg/tracing"
//
//   config := tracing.DefaultMiddlewareConfig()
//   server := NewServer(
//       WithMiddleware(TracingMiddleware(config)),
//   )
//
// Note: This requires the tracing package to be initialized first:
//   tp, err := tracing.InitTracing(tracing.DefaultConfig())
//   defer tp.Shutdown(context.Background())
func TracingMiddleware(config interface{}) Middleware {
	// The actual implementation is in pkg/tracing/integration.go
	// This is just a placeholder that can be replaced when tracing is properly initialized
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			// If tracing is not initialized, just pass through
			return next(ctx)
		}
	}
}
