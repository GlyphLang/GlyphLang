package metrics

import (
	"time"

	"github.com/glyphlang/glyph/pkg/server"
)

// MetricsMiddleware creates a middleware that automatically collects HTTP metrics
func MetricsMiddleware(m *Metrics) server.Middleware {
	return func(next server.RouteHandler) server.RouteHandler {
		return func(ctx *server.Context) error {
			start := time.Now()

			// Call next handler
			err := next(ctx)

			// Calculate duration
			duration := time.Since(start)

			// Determine status code
			status := ctx.StatusCode
			if status == 0 {
				if err != nil {
					status = 500
				} else {
					status = 200
				}
			}

			// Record metrics
			method := ctx.Request.Method
			path := ctx.Request.URL.Path
			m.RecordRequest(method, path, status, duration)

			return err
		}
	}
}
