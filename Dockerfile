# Multi-stage Dockerfile for Glyph
# Stage 1: Build Go CLI
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source code
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/

# Build Go CLI (static binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o glyph ./cmd/glyph

# Stage 2: Runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 glyph && \
    adduser -D -u 1000 -G glyph glyph

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/glyph /usr/local/bin/glyph

# Copy examples (optional)
COPY examples/ ./examples/

# Set ownership
RUN chown -R glyph:glyph /app

# Switch to non-root user
USER glyph

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/glyph"]
CMD ["run", "examples/rest-api/main.glyph", "--port", "8080"]
