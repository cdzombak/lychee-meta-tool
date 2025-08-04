SHELL:=/usr/bin/env bash

BIN_NAME:=lychee-meta-tool
BIN_VERSION:=$(shell ./.version.sh)

default: help
.PHONY: help
help: ## Print help
	@grep -E '^[a-zA-Z_-/]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: clean build-linux-amd64 build-linux-arm64 build-linux-386 build-linux-armv7 build-linux-armv6 build-darwin-amd64 build-darwin-arm64 ## Build for macOS (amd64, arm64) and Linux (amd64, 386, arm64, armv7, armv6)

.PHONY: clean
clean: ## Remove build products (./out)
	rm -rf ./out
	rm -rf frontend/dist
	rm -rf frontend/node_modules

.PHONY: frontend
frontend: ## Build frontend
	cd frontend && npm install && npm run build

.PHONY: build
build: frontend ## Build for the current platform & architecture to ./out
	mkdir -p out
	env CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME} .

.PHONY: build-linux-amd64
build-linux-amd64: frontend ## Build for Linux/amd64 to ./out
	mkdir -p out
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-linux-amd64 .

.PHONY: build-linux-arm64
build-linux-arm64: frontend ## Build for Linux/arm64 to ./out
	mkdir -p out
	env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-linux-arm64 .

.PHONY: build-linux-386
build-linux-386: frontend ## Build for Linux/386 to ./out
	mkdir -p out
	env GOOS=linux GOARCH=386 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-linux-386 .

.PHONY: build-linux-armv7
build-linux-armv7: frontend ## Build for Linux/armv7 to ./out
	mkdir -p out
	env GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-linux-armv7 .

.PHONY: build-linux-armv6
build-linux-armv6: frontend ## Build for Linux/armv6 to ./out
	mkdir -p out
	env GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-linux-armv6 .

.PHONY: build-darwin-amd64
build-darwin-amd64: frontend ## Build for macOS/amd64 to ./out
	mkdir -p out
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-darwin-amd64 .

.PHONY: build-darwin-arm64
build-darwin-arm64: frontend ## Build for macOS/arm64 to ./out
	mkdir -p out
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-darwin-arm64 .

.PHONY: package
package: all ## Build all binaries + .deb packages to ./out (requires fpm: https://fpm.readthedocs.io)
	fpm -t deb -v ${BIN_VERSION} -p ./out/${BIN_NAME}-${BIN_VERSION}-amd64.deb -a amd64 ./out/${BIN_NAME}-${BIN_VERSION}-linux-amd64=/usr/bin/${BIN_NAME}
	fpm -t deb -v ${BIN_VERSION} -p ./out/${BIN_NAME}-${BIN_VERSION}-arm64.deb -a arm64 ./out/${BIN_NAME}-${BIN_VERSION}-linux-arm64=/usr/bin/${BIN_NAME}
	fpm -t deb -v ${BIN_VERSION} -p ./out/${BIN_NAME}-${BIN_VERSION}-386.deb -a i386 ./out/${BIN_NAME}-${BIN_VERSION}-linux-386=/usr/bin/${BIN_NAME}
	fpm -t deb -v ${BIN_VERSION} -p ./out/${BIN_NAME}-${BIN_VERSION}-armv7.deb -a armhf ./out/${BIN_NAME}-${BIN_VERSION}-linux-armv7=/usr/bin/${BIN_NAME}
	fpm -t deb -v ${BIN_VERSION} -p ./out/${BIN_NAME}-${BIN_VERSION}-armv6.deb -a armel ./out/${BIN_NAME}-${BIN_VERSION}-linux-armv6=/usr/bin/${BIN_NAME}

.PHONY: test
test: ## Run backend tests
	go test ./backend/...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: dev
dev: build ## Build and run in development mode
	./out/${BIN_NAME} -config config.example.yaml

.PHONY: dev-frontend
dev-frontend: ## Run frontend development server
	cd frontend && npm run dev

.PHONY: install-deps
install-deps: ## Install frontend dependencies
	cd frontend && npm install