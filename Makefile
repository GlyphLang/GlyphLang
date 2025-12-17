.PHONY: help build test bench clean docker deploy docs examples installer

# Default target
help:
	@echo "GlyphLang - AI Backend Compiler (Production Ready)"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build Rust core and Go CLI"
	@echo "  build-all     - Build for all platforms (Windows, Linux, macOS)"
	@echo "  test          - Run all tests (640+ tests)"
	@echo "  bench         - Run all benchmarks"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker        - Build Docker image"
	@echo "  deploy-k8s    - Deploy to Kubernetes"
	@echo "  examples      - Run example applications"
	@echo "  installer     - Build Windows installer (requires Inno Setup)"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linters"

# Build targets
build: build-rust build-go

build-rust:
	@echo "Building Rust core..."
	cd glyph-core && cargo build --release

build-go:
	@echo "Building Go CLI..."
	go build -o glyph.exe ./cmd/glyph

# Cross-platform build
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -o dist/glyph-windows-amd64.exe ./cmd/glyph
	GOOS=linux GOARCH=amd64 go build -o dist/glyph-linux-amd64 ./cmd/glyph
	GOOS=darwin GOARCH=amd64 go build -o dist/glyph-darwin-amd64 ./cmd/glyph
	GOOS=darwin GOARCH=arm64 go build -o dist/glyph-darwin-arm64 ./cmd/glyph
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
test: test-rust test-go

test-rust:
	@echo "Running Rust tests (84 tests)..."
	cd glyph-core && cargo test

test-go:
	@echo "Running Go tests (408+ tests)..."
	go test ./pkg/... -v

# Benchmark targets
bench: bench-rust bench-go

bench-rust:
	@echo "Running Rust benchmarks..."
	cd glyph-core && cargo bench

bench-go:
	@echo "Running Go benchmarks..."
	go test -bench=. ./pkg/vm/ -benchmem

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	cd glyph-core && cargo clean
	go clean
	rm -f aida glyph.exe

# Docker targets
docker:
	@echo "Building Docker image..."
	docker build -t glyph:latest .

docker-compose-up:
	@echo "Starting Docker Compose stack..."
	docker-compose up -d

# Kubernetes targets
deploy-k8s:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deploy/kubernetes/

# Example targets
examples: example-hello example-rest example-blog

example-hello:
	@echo "Running Hello World example..."
	./glyph.exe run examples/hello-world/main.glyph

example-rest:
	@echo "Running REST API example..."
	./glyph.exe dev examples/rest-api/main.glyph --port 8080

example-blog:
	@echo "Running Blog API example..."
	./glyph.exe dev examples/blog-api-complete/main.glyph --port 8080

# Format targets
fmt: fmt-rust fmt-go

fmt-rust:
	@echo "Formatting Rust code..."
	cd glyph-core && cargo fmt

fmt-go:
	@echo "Formatting Go code..."
	go fmt ./...

# Lint targets
lint: lint-rust lint-go

lint-rust:
	@echo "Linting Rust code..."
	cd glyph-core && cargo clippy -- -D warnings

lint-go:
	@echo "Linting Go code..."
	golangci-lint run ./...
