.PHONY: build clean frontend backend test dev run

# Default target
all: build

# Build both frontend and backend
build: frontend backend

# Build frontend
frontend:
	cd frontend && npm install && npm run build

# Build backend (requires frontend to be built first)
backend:
	go build -o lychee-meta-tool .

# Clean build artifacts
clean:
	rm -f lychee-meta-tool
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# Run tests
test:
	go test ./backend/...

# Development mode - build and run
dev: build
	./lychee-meta-tool -config config.example.yaml

# Run with existing binary
run:
	./lychee-meta-tool -config config.example.yaml

# Install frontend dependencies
install-deps:
	cd frontend && npm install

# Development frontend server (separate from backend)
dev-frontend:
	cd frontend && npm run dev

# Build for release (with optimizations)
release: frontend
	CGO_ENABLED=1 go build -ldflags="-s -w" -o lychee-meta-tool .

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build both frontend and backend"
	@echo "  frontend     - Build frontend only"
	@echo "  backend      - Build backend only"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run backend tests"
	@echo "  dev          - Build and run in development mode"
	@echo "  run          - Run with existing binary"
	@echo "  release      - Build optimized release binary"
	@echo "  help         - Show this help"