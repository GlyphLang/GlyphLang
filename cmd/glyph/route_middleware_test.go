package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glyphlang/glyph/pkg/ast"
	"github.com/glyphlang/glyph/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A route declaring auth or a rate limit served every request: the directives
// were parsed onto ast.Route and never turned into middleware. These tests fix
// the behaviour in place so the wiring cannot be dropped again.

// guard builds a route's middleware chain once and returns a function that
// issues one request through it. Building the chain per request would hand
// every call a fresh rate-limit bucket and no limit would ever be reached.
func guard(t *testing.T, route *ast.Route) func(decorate func(*http.Request)) int {
	t.Helper()

	handler := server.RouteHandler(func(ctx *server.Context) error {
		return server.SendJSON(ctx, http.StatusOK, map[string]interface{}{"ok": true})
	})
	middlewares := routeMiddlewares(route)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(decorate func(*http.Request)) int {
		req := httptest.NewRequest("GET", "/guarded", nil)
		if decorate != nil {
			decorate(req)
		}
		// A fixed remote address: the identity these middlewares key on must
		// not vary per request, or no limit is ever reached.
		req.RemoteAddr = "203.0.113.7:44321"

		rec := httptest.NewRecorder()
		ctx := &server.Context{
			Request:        req,
			ResponseWriter: rec,
			PathParams:     map[string]string{},
			StatusCode:     http.StatusOK,
		}

		require.NoError(t, handler(ctx))
		return rec.Code
	}
}

func TestAPIKeyAuthEnforced(t *testing.T) {
	t.Setenv(envAPIKeys, "key-one, key-two")
	request := guard(t, &ast.Route{Auth: &ast.AuthConfig{AuthType: "apikey", Required: true}})

	assert.Equal(t, 401, request(nil), "no key must be rejected")
	assert.Equal(t, 401, request(func(r *http.Request) {
		r.Header.Set("X-API-Key", "wrong")
	}), "an unknown key must be rejected")
	assert.Equal(t, 200, request(func(r *http.Request) {
		r.Header.Set("X-API-Key", "key-two")
	}), "a configured key must pass")
	assert.Equal(t, 200, request(func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer key-one")
	}), "a configured key must pass as a bearer token too")
}

// TestUnconfiguredAuthFailsClosed is the property that matters most: a route
// that declares auth with nothing configured to check it must deny, not serve.
func TestUnconfiguredAuthFailsClosed(t *testing.T) {
	t.Setenv(envAPIKeys, "")
	t.Setenv(envJWTSecret, "")

	for _, authType := range []string{"apikey", "jwt"} {
		request := guard(t, &ast.Route{Auth: &ast.AuthConfig{AuthType: authType, Required: true}})
		assert.Equal(t, 401, request(func(r *http.Request) {
			r.Header.Set("X-API-Key", "anything")
			r.Header.Set("Authorization", "Bearer anything")
		}), "auth(%s) with no credential configured must deny", authType)
	}
}

func TestRateLimitEnforced(t *testing.T) {
	request := guard(t, &ast.Route{RateLimit: &ast.RateLimit{Requests: 2, Window: "min"}})

	assert.Equal(t, 200, request(nil), "first request within budget")
	assert.Equal(t, 200, request(nil), "second request within budget")
	assert.Equal(t, 429, request(nil), "third request exceeds 2/min")
}

// TestRateLimitWindowConversion checks the per-minute budget each declared
// window produces, including the rounding that keeps an hourly budget from
// collapsing to zero requests per minute.
func TestRateLimitWindowConversion(t *testing.T) {
	for _, tc := range []struct {
		window   string
		requests uint32
		allowed  int
	}{
		{"min", 3, 3},
		{"sec", 1, 60},
		{"hour", 120, 2},
		{"hour", 30, 1}, // rounds up rather than down to zero
	} {
		request := guard(t, &ast.Route{RateLimit: &ast.RateLimit{Requests: tc.requests, Window: tc.window}})

		for i := 0; i < tc.allowed; i++ {
			assert.Equal(t, 200, request(nil),
				"%d/%s: request %d should be within the budget of %d",
				tc.requests, tc.window, i+1, tc.allowed)
		}
		assert.Equal(t, 429, request(nil),
			"%d/%s: request %d should exceed the budget of %d",
			tc.requests, tc.window, tc.allowed+1, tc.allowed)
	}
}

func TestNoDirectivesMeansNoMiddleware(t *testing.T) {
	assert.Empty(t, routeMiddlewares(&ast.Route{}),
		"a route without directives should not be wrapped")
}
