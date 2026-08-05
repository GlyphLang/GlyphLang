package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/server"
)

// Routes declare `+ auth(jwt)` and `+ ratelimit(100/min)`, the parser records
// them on ast.Route, and pkg/server already applies Route.Middlewares - but
// nothing ever built the middlewares, so both directives were silently inert
// and a route declaring auth served anyone.
//
// Credentials come from the environment, like the provider configuration, so a
// .glyph file carries none.
const (
	envJWTSecret = "GLYPH_JWT_SECRET"
	envAPIKeys   = "GLYPH_API_KEYS"
)

// authEnvHint names the variable a given auth type needs.
func authEnvHint(authType string) string {
	if strings.EqualFold(authType, "apikey") {
		return envAPIKeys + " (comma-separated keys)"
	}
	return envJWTSecret
}

// routeMiddlewares builds the middleware chain for one route from its declared
// directives. Order matters: rate limiting runs first so an unauthenticated
// flood is cheap to reject.
func routeMiddlewares(route *ast.Route) []server.Middleware {
	var middlewares []server.Middleware

	if limit := rateLimitMiddleware(route.RateLimit); limit != nil {
		middlewares = append(middlewares, limit)
	}
	if auth := authMiddleware(route.Auth); auth != nil {
		middlewares = append(middlewares, auth)
	}
	return middlewares
}

// rateLimitMiddleware converts a declared limit into a per-minute budget.
func rateLimitMiddleware(limit *ast.RateLimit) server.Middleware {
	if limit == nil || limit.Requests == 0 {
		return nil
	}

	perMinute := int(limit.Requests)
	switch strings.ToLower(strings.TrimSpace(limit.Window)) {
	case "sec", "second", "s":
		perMinute = int(limit.Requests) * 60
	case "hour", "hr", "h":
		// Round up so a declared budget is never reduced to zero.
		perMinute = (int(limit.Requests) + 59) / 60
	case "day", "d":
		perMinute = (int(limit.Requests) + 1439) / 1440
	}
	if perMinute < 1 {
		perMinute = 1
	}

	return server.RateLimitMiddleware(server.RateLimiterConfig{
		RequestsPerMinute: perMinute,
		BurstSize:         perMinute,
	})
}

// authMiddleware builds credential checking for a declared auth directive.
//
// It fails closed: when a route declares auth and no credential source is
// configured, every request is rejected rather than served. An unconfigured
// deployment that silently served protected routes is the bug this fixes, so
// the safe direction is to deny and warn loudly at startup.
func authMiddleware(auth *ast.AuthConfig) server.Middleware {
	if auth == nil {
		return nil
	}

	if strings.EqualFold(auth.AuthType, "apikey") {
		keys := parseAPIKeys(os.Getenv(envAPIKeys))
		if len(keys) == 0 {
			return denyAllMiddleware("apikey")
		}
		return apiKeyMiddleware(keys)
	}

	// Everything else is treated as bearer-token auth (jwt is the only other
	// form the notation defines today).
	secret := strings.TrimSpace(os.Getenv(envJWTSecret))
	if secret == "" {
		return denyAllMiddleware(auth.AuthType)
	}
	return server.BasicAuthMiddleware(map[string]bool{secret: true})
}

// parseAPIKeys splits the configured key list, ignoring blanks.
func parseAPIKeys(raw string) map[string]bool {
	keys := make(map[string]bool)
	for _, key := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			keys[trimmed] = true
		}
	}
	return keys
}

// apiKeyMiddleware accepts a key from X-API-Key or a Bearer authorization
// header, so a caller can use whichever its client library supports.
func apiKeyMiddleware(validKeys map[string]bool) server.Middleware {
	return func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) error {
			key := strings.TrimSpace(ctx.Request.Header.Get("X-API-Key"))
			if key == "" {
				const bearer = "Bearer "
				authHeader := ctx.Request.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, bearer) {
					key = strings.TrimSpace(strings.TrimPrefix(authHeader, bearer))
				}
			}

			if key == "" {
				return server.SendError(ctx, 401, "unauthorized: missing API key")
			}
			if !validKeys[key] {
				return server.SendError(ctx, 401, "unauthorized: invalid API key")
			}
			return next(ctx)
		}
	}
}

// denyAllMiddleware rejects every request to a route whose auth is declared but
// unconfigured. The response says nothing about the server's configuration; the
// startup warning tells the operator what to set.
func denyAllMiddleware(authType string) server.Middleware {
	return func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) error {
			return server.SendError(ctx, 401, "unauthorized")
		}
	}
}

// warnUnconfiguredAuth reports, once per auth type, that routes declare auth
// with nothing configured to check it - the case where every request to those
// routes is denied.
func warnUnconfiguredAuth(module *ast.Module) {
	warned := make(map[string]bool)

	for _, item := range module.Items {
		route, ok := item.(*ast.Route)
		if !ok || route.Auth == nil {
			continue
		}

		authType := strings.ToLower(route.Auth.AuthType)
		if warned[authType] {
			continue
		}

		configured := os.Getenv(envJWTSecret) != ""
		if strings.EqualFold(authType, "apikey") {
			configured = len(parseAPIKeys(os.Getenv(envAPIKeys))) > 0
		}
		if configured {
			continue
		}

		warned[authType] = true
		printWarning(fmt.Sprintf(
			"routes declare auth(%s) but %s is not set: those routes will reject every request",
			authType, authEnvHint(authType)))
	}
}
