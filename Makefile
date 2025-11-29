.PHONY: build install install-local clean test lint schema

BINARY_NAME=gh-rset
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
GH_EXTENSION_DIR=$(HOME)/.local/share/gh/extensions/gh-rset

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
	golangci-lint run

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
