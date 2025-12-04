.PHONY: build install install-local clean test lint fmt schema \
       fetch-openapi extract-openapi generate-github-types generate

BINARY_NAME=gh-repo-settings
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
GH_EXTENSION_DIR=$(HOME)/.local/share/gh/extensions/gh-repo-settings

# Tool versions (keep in sync with CI)
GOLANGCI_LINT_VERSION=v2.0.2
GOFUMPT_VERSION=v0.7.0

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

# Install to gh extension directory for local testing
install-local:
	@mkdir -p $(GH_EXTENSION_DIR)
	go build $(LDFLAGS) -o $(GH_EXTENSION_DIR)/$(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

test:
	go test -v ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

fmt:
	go run mvdan.cc/gofumpt@$(GOFUMPT_VERSION) -w .

# Cross-compile for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

# Generate JSON Schema from Go types
schema:
	go run ./cmd/schema/main.go > schema.json

# ============================================
# GitHub OpenAPI Type Generation
# ============================================

# Download GitHub OpenAPI schema
fetch-openapi:
	@chmod +x scripts/fetch-github-openapi.sh
	@scripts/fetch-github-openapi.sh

# Extract required endpoints from OpenAPI schema
extract-openapi: fetch-openapi
	@go run scripts/extract-openapi-subset.go

# Generate Go types from OpenAPI schema
generate-github-types: extract-openapi
	@chmod +x scripts/generate-types.sh
	@scripts/generate-types.sh

# Alias for generate-github-types
generate: generate-github-types
