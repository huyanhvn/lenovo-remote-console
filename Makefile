.PHONY: all build test clean fmt lint install run help certs

# Variables
BINARY_NAME=lenovo-console
BINARY_PATH=./cmd/lenovo-console
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"

# Default target
all: build

## help: Display this help message
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

## build: Build the binary
build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)

## install: Install the binary to GOPATH/bin
install:
	$(GO) install $(GOFLAGS) $(LDFLAGS) $(BINARY_PATH)

## test: Run all tests
test:
	$(GO) test $(GOFLAGS) -race -cover ./...

## test-verbose: Run tests with verbose output
test-verbose:
	$(GO) test $(GOFLAGS) -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: Format the code
fmt:
	$(GO) fmt ./...
	@echo "Code formatted"

## lint: Run linters
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install it with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## vet: Run go vet
vet:
	$(GO) vet ./...

## mod-tidy: Tidy and verify module dependencies
mod-tidy:
	$(GO) mod tidy
	$(GO) mod verify

## mod-download: Download module dependencies
mod-download:
	$(GO) mod download

## clean: Clean build artifacts
clean:
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME).exe
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "Cleaned build artifacts"

## certs: Generate self-signed certificates for testing
certs:
	@if [ ! -f server.crt ] || [ ! -f server.key ]; then \
		echo "Generating self-signed certificates..."; \
		openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
			-days 365 -nodes -subj "/CN=localhost"; \
		echo "Certificates generated: server.crt, server.key"; \
	else \
		echo "Certificates already exist"; \
	fi

## run: Run the application (requires BMC_IP, USERNAME, PASSWORD env vars)
run: build certs
	@if [ -z "$(BMC_IP)" ] || [ -z "$(USERNAME)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "Usage: make run BMC_IP=<ip> USERNAME=<user> PASSWORD=<pass> [BROWSER=firefox]"; \
		exit 1; \
	fi
	./$(BINARY_NAME) $(BMC_IP) $(USERNAME) $(PASSWORD) $(BROWSER)

## docker-build: Build Docker image
docker-build:
	docker build -t lenovo-remote-console:latest .

## release-dry-run: Test the release process without publishing
release-dry-run:
	@echo "Building release binaries..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(BINARY_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(BINARY_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(BINARY_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(BINARY_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(BINARY_PATH)
	@echo "Release binaries built in dist/"
	@ls -la dist/

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"
