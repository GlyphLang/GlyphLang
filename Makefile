.PHONY: help build test bench clean docker deploy docs examples installer

# Default target
help:
	@echo "GlyphLang - AI Backend Compiler"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build Go CLI"
	@echo "  build-all     - Build for all platforms (Windows, Linux, macOS)"
	@echo "  test          - Run all tests"
	@echo "  bench         - Run benchmarks"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker        - Build Docker image"
	@echo "  deploy-k8s    - Deploy to Kubernetes"
	@echo "  examples      - Run example applications"
	@echo "  installer     - Build Windows installer (requires Inno Setup)"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linters"

# Version injection
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build targets
build:
	@echo "Building Go CLI..."
	go build -ldflags "-X main.version=$(VERSION)" -o glyph ./cmd/glyph

build-windows:
	@echo "Building Go CLI for Windows..."
	go build -ldflags "-X main.version=$(VERSION)" -o glyph.exe ./cmd/glyph

# Cross-platform build
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/glyph-windows-amd64.exe ./cmd/glyph
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/glyph-linux-amd64 ./cmd/glyph
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/glyph-darwin-amd64 ./cmd/glyph
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/glyph-darwin-arm64 ./cmd/glyph
	@echo "Built binaries in dist/"

# Windows installer (requires Inno Setup)
installer: build-all
	@echo "Building Windows installer..."
	@if command -v iscc >/dev/null 2>&1; then \
		iscc installer/glyph-setup.iss; \
	elif [ -f "/c/Program Files (x86)/Inno Setup 6/ISCC.exe" ]; then \
		"/c/Program Files (x86)/Inno Setup 6/ISCC.exe" installer/glyph-setup.iss; \
	else \
		echo "Inno Setup not found. Install from https://jrsoftware.org/isinfo.php"; \
		exit 1; \
	fi
	@echo "Installer created: dist/glyph-*-windows-setup.exe"

# Test targets
test:
	@echo "Running Go tests..."
	go test ./... -v

test-short:
	@echo "Running Go tests (short mode)..."
	go test ./... -v -short

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Benchmark targets
bench:
	@echo "Running Go benchmarks..."
	go test -bench=. ./pkg/vm/ -benchmem

bench-all:
	@echo "Running all benchmarks..."
	go test -bench=. ./pkg/... -benchmem

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f glyph glyph.exe
	rm -rf dist/
	rm -f coverage.out coverage.html

# Docker targets
docker:
	@echo "Building Docker image..."
	docker build -t glyph:latest .

docker-dev:
	@echo "Building development Docker image..."
	docker build -f Dockerfile.dev -t glyph:dev .

docker-compose-up:
	@echo "Starting Docker Compose stack..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping Docker Compose stack..."
	docker-compose down

# Kubernetes targets
deploy-k8s:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deploy/kubernetes/

# Example targets
examples: example-hello example-rest example-blog

example-hello:
	@echo "Running Hello World example..."
	go run ./cmd/glyph run examples/hello-world/main.glyph

example-rest:
	@echo "Running REST API example..."
	go run ./cmd/glyph dev examples/rest-api/main.glyph --port 8080

example-blog:
	@echo "Running Blog API example..."
	go run ./cmd/glyph dev examples/blog-api-complete/main.glyph --port 8080

# Format targets
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Lint targets
lint:
	@echo "Linting Go code..."
	golangci-lint run ./...

# Run the application
run:
	@echo "Running Glyph..."
	go run ./cmd/glyph run examples/hello-world/main.glyph

dev:
	@echo "Running Glyph in dev mode..."
	go run ./cmd/glyph dev examples/rest-api/main.glyph --port 8080
