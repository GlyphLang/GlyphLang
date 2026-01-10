package server

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
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
// It logs full panic details to server logs but returns a generic error to clients
// to prevent information disclosure
func RecoveryMiddleware() Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					method := ctx.Request.Method
					path := ctx.Request.URL.Path
					// Log full panic details including stack trace to server logs
					log.Printf("[PANIC] %s %s: %v\n%s", method, path, r, debug.Stack())
					// Return generic error to client - don't expose panic details
					SendError(ctx, 500, "Internal Server Error")
					// Return error to indicate a panic was recovered
					err = &InternalError{
						BaseError: &BaseError{
							Code: 500,
							Type: "InternalError",
							Msg:  "internal server error",
						},
					}
				}
			}()

			return next(ctx)
		}
	}
}

// CORSMiddleware adds CORS headers to responses
// Security: When allowedOrigins contains "*", we set the literal "*" header
// and explicitly disable credentials to prevent security vulnerabilities
func CORSMiddleware(allowedOrigins []string) Middleware {
	// Check if wildcard is configured and log warning
	hasWildcard := false
	for _, o := range allowedOrigins {
		if o == "*" {
			hasWildcard = true
			log.Printf("[SECURITY WARNING] CORS configured with wildcard origin '*'. " +
				"This allows any origin to access the API. Credentials will be disabled.")
			break
		}
	}

	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			origin := ctx.Request.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			isWildcardMatch := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" {
					allowed = true
					isWildcardMatch = true
					break
				}
				if allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if isWildcardMatch || hasWildcard {
					// When using wildcard, set literal "*" and disable credentials
					// Never reflect the origin when wildcard is configured
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Credentials", "false")
				} else if origin != "" {
					// For specific origins, reflect the origin
					ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", origin)
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

// SecurityHeadersMiddleware adds security headers to all responses
// These headers help protect against common web vulnerabilities
func SecurityHeadersMiddleware() Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			ctx.ResponseWriter.Header().Set("X-Content-Type-Options", "nosniff")
			ctx.ResponseWriter.Header().Set("X-Frame-Options", "DENY")
			ctx.ResponseWriter.Header().Set("X-XSS-Protection", "1; mode=block")
			ctx.ResponseWriter.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
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

// authFailureTracker tracks failed authentication attempts per IP
type authFailureTracker struct {
	failures    int
	lastFailure time.Time
	lockedUntil time.Time
}

// AuthRateLimitConfig configures auth rate limiting behavior
type AuthRateLimitConfig struct {
	MaxFailures     int           // Maximum failures before lockout (default: 5)
	LockoutDuration time.Duration // Initial lockout duration (default: 1 minute)
	MaxLockout      time.Duration // Maximum lockout duration with exponential backoff (default: 15 minutes)
	ResetAfter      time.Duration // Reset failure count after this duration of no failures (default: 15 minutes)
}

// DefaultAuthRateLimitConfig returns sensible defaults for auth rate limiting
func DefaultAuthRateLimitConfig() AuthRateLimitConfig {
	return AuthRateLimitConfig{
		MaxFailures:     5,
		LockoutDuration: 1 * time.Minute,
		MaxLockout:      15 * time.Minute,
		ResetAfter:      15 * time.Minute,
	}
}

// BasicAuthMiddleware provides simple token-based authentication
// with rate limiting to prevent brute force attacks
func BasicAuthMiddleware(validTokens map[string]bool) Middleware {
	return BasicAuthMiddlewareWithConfig(validTokens, DefaultAuthRateLimitConfig())
}

// BasicAuthMiddlewareWithConfig provides token-based authentication with custom rate limit config
func BasicAuthMiddlewareWithConfig(validTokens map[string]bool, config AuthRateLimitConfig) Middleware {
	// Track failed auth attempts per IP
	var mu sync.Mutex
	failureTrackers := make(map[string]*authFailureTracker)

	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			clientIP := getClientIP(ctx.Request)

			mu.Lock()
			tracker, exists := failureTrackers[clientIP]
			if !exists {
				tracker = &authFailureTracker{}
				failureTrackers[clientIP] = tracker
			}

			now := time.Now()

			// Check if client is locked out
			if now.Before(tracker.lockedUntil) {
				mu.Unlock()
				remaining := tracker.lockedUntil.Sub(now).Round(time.Second)
				log.Printf("[AUTH] IP %s is locked out for %v due to too many failed attempts", clientIP, remaining)
				return SendError(ctx, 429, "too many failed authentication attempts, try again later")
			}

			// Reset failure count if enough time has passed since last failure
			if exists && now.Sub(tracker.lastFailure) > config.ResetAfter {
				tracker.failures = 0
			}
			mu.Unlock()

			token := ctx.Request.Header.Get("Authorization")
			if token == "" {
				recordAuthFailure(clientIP, failureTrackers, &mu, config)
				return SendError(ctx, 401, "unauthorized: missing authorization header")
			}

			// Remove "Bearer " prefix if present
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			// Check if token is valid
			if validTokens != nil && !validTokens[token] {
				recordAuthFailure(clientIP, failureTrackers, &mu, config)
				return SendError(ctx, 401, "unauthorized: invalid token")
			}

			// Success - reset failure count
			mu.Lock()
			tracker.failures = 0
			mu.Unlock()

			return next(ctx)
		}
	}
}

// recordAuthFailure records a failed authentication attempt and applies lockout if needed
func recordAuthFailure(clientIP string, trackers map[string]*authFailureTracker, mu *sync.Mutex, config AuthRateLimitConfig) {
	mu.Lock()
	defer mu.Unlock()

	tracker := trackers[clientIP]
	tracker.failures++
	tracker.lastFailure = time.Now()

	if tracker.failures >= config.MaxFailures {
		// Apply exponential backoff for lockout duration
		// Each subsequent lockout doubles the duration up to MaxLockout
		lockoutMultiplier := 1 << (tracker.failures - config.MaxFailures)
		lockoutDuration := config.LockoutDuration * time.Duration(lockoutMultiplier)
		if lockoutDuration > config.MaxLockout {
			lockoutDuration = config.MaxLockout
		}
		tracker.lockedUntil = time.Now().Add(lockoutDuration)
		log.Printf("[AUTH] IP %s locked out for %v after %d failed attempts",
			clientIP, lockoutDuration, tracker.failures)
	}
}

// getClientIP extracts the real client IP address from an HTTP request
// It checks X-Forwarded-For and X-Real-IP headers (for proxy setups)
// before falling back to RemoteAddr
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs, take the first)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
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

	var mu sync.Mutex
	limits := make(map[string]*clientLimit)

	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			// Get client identifier (IP address) - supports proxied requests
			clientIP := getClientIP(ctx.Request)

			mu.Lock()

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
				mu.Unlock()
				log.Printf("[RATE_LIMIT] Rate limit exceeded for %s", clientIP)
				return SendError(ctx, 429, "rate limit exceeded")
			}

			// Consume a token
			limit.tokens--
			limit.requestCount++
			mu.Unlock()

			return next(ctx)
		}
	}
}

// RedisRateLimiter defines the interface for Redis-based rate limiting
type RedisRateLimiter interface {
	Incr(key string) (int64, error)
	Expire(key string, seconds int64) (bool, error)
}

// RateLimitMiddlewareWithRedis implements distributed rate limiting using Redis
// This is production-ready and works across multiple server instances
func RateLimitMiddlewareWithRedis(redisHandler RedisRateLimiter, config RateLimiterConfig) Middleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx *Context) error {
			clientIP := getClientIP(ctx.Request)

			// Use a sliding window with minute-based keys
			// Key format: "ratelimit:{ip}:{minute_timestamp}"
			windowKey := fmt.Sprintf("ratelimit:%s:%d", clientIP, time.Now().Unix()/60)

			// Increment the counter atomically
			count, err := redisHandler.Incr(windowKey)
			if err != nil {
				// On Redis error, fail open (allow the request) and log the error
				log.Printf("[RATE_LIMIT] Redis error: %v, allowing request for %s", err, clientIP)
				return next(ctx)
			}

			// Set expiration on first request in this window
			if count == 1 {
				// Expire after 2 minutes to ensure cleanup (1 minute window + 1 minute buffer)
				_, expErr := redisHandler.Expire(windowKey, 120)
				if expErr != nil {
					log.Printf("[RATE_LIMIT] Redis expire error: %v", expErr)
				}
			}

			// Check if rate limit exceeded
			if count > int64(config.RequestsPerMinute) {
				log.Printf("[RATE_LIMIT] Redis rate limit exceeded for %s (count: %d, limit: %d)",
					clientIP, count, config.RequestsPerMinute)
				return SendError(ctx, 429, "rate limit exceeded")
			}

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
