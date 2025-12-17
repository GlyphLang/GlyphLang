# Multi-stage Dockerfile for Glyph
# Stage 1: Build Rust compiler core
FROM rust:1.75-slim as rust-builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Copy Rust project
COPY glyph-core/ ./glyph-core/
COPY Cargo.toml Cargo.lock* ./

# Build Rust core in release mode
WORKDIR /build/glyph-core
RUN cargo build --release

# Stage 2: Build Go runtime and CLI
FROM golang:1.21-alpine as go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source code
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/

# Copy Rust artifacts from previous stage
COPY --from=rust-builder /build/glyph-core/target/release/libglyph_core.* /usr/local/lib/
COPY --from=rust-builder /build/glyph-core/target/release/libglyph_core.a /usr/local/lib/

# Build Go CLI
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o glyph ./cmd/glyph

# Stage 3: Runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    libc6-compat

# Create non-root user
RUN addgroup -g 1000 glyph && \
    adduser -D -u 1000 -G glyph glyph

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /build/glyph /usr/local/bin/glyph
COPY --from=rust-builder /build/glyph-core/target/release/libglyph_core.* /usr/local/lib/

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
CMD ["run", "examples/rest-api/main.abc", "--port", "8080"]
