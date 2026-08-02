package main

import (
	"net/http/httptest"
	"testing"

	"github.com/glyphlang/glyph/pkg/server"
	"github.com/stretchr/testify/require"
)

// BenchmarkCompiledRouteHTTP measures the real production request path used by
// `glyph run`: handler dispatch -> router.Match -> compiled-route handler with
// a fresh VM per request -> JSON response.
func BenchmarkCompiledRouteHTTP(b *testing.B) {
	src := `@ GET /api/users/:id {
  > {id: id, name: "Test User", email: "test@example.com"}
}`
	route, bytecode := compileFirstRoute(b, src)
	router := server.NewRouter()
	require.NoError(b, registerCompiledRoute(router, route, bytecode, nil))
	handler := createHandler(router)

	req := httptest.NewRequest("GET", "/api/users/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
