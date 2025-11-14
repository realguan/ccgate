# Makefile for ccgate

# Binary name
BINARY=ccgate

# Directories
SRC_DIR=.
BUILD_DIR=build

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")

# LDFLAGS for embedding build info
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build $(LDFLAGS)
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# GitHub CLI check
GH := $(shell command -v gh 2> /dev/null)

# Default target
all: build

# Build the binary
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY)

# Install the binary to ~/bin
install: build
	mkdir -p ~/bin
	cp $(BUILD_DIR)/$(BINARY) ~/bin/$(BINARY)

# Install the binary to GOPATH/bin
install-go:
	$(GOINSTALL)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Tidy go modules
tidy:
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# Platform-specific builds
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-arm64

build-mac-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64

build-mac-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64

build-windows-amd64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe

build-windows-arm64:
	GOOS=windows GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-windows-arm64.exe

build-freebsd-amd64:
	GOOS=freebsd GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-freebsd-amd64

# Legacy aliases for backward compatibility
build-linux: build-linux-amd64
build-mac: build-mac-amd64
build-windows: build-windows-amd64

# Generic cross-platform build
build-cross:
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "Error: GOOS and GOARCH environment variables must be set"; \
		echo "Example: GOOS=linux GOARCH=arm64 make build-cross"; \
		exit 1; \
	fi
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)

# Build all supported platforms
build-all: build-linux-amd64 build-linux-arm64 build-mac-amd64 build-mac-arm64 build-windows-amd64 build-windows-arm64 build-freebsd-amd64

# Package builds into archives
package: build-all
	@echo "Creating archives..."
	@mkdir -p $(BUILD_DIR)/archives
	@cd $(BUILD_DIR) && \
	tar -czf archives/$(BINARY)-$(VERSION)-linux-amd64.tar.gz $(BINARY)-linux-amd64 && \
	tar -czf archives/$(BINARY)-$(VERSION)-linux-arm64.tar.gz $(BINARY)-linux-arm64 && \
	tar -czf archives/$(BINARY)-$(VERSION)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64 && \
	tar -czf archives/$(BINARY)-$(VERSION)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64 && \
	zip -q archives/$(BINARY)-$(VERSION)-windows-amd64.zip $(BINARY)-windows-amd64.exe && \
	zip -q archives/$(BINARY)-$(VERSION)-windows-arm64.zip $(BINARY)-windows-arm64.exe && \
	tar -czf archives/$(BINARY)-$(VERSION)-freebsd-amd64.tar.gz $(BINARY)-freebsd-amd64
	@echo "Archives created in $(BUILD_DIR)/archives/"

# GitHub Release functionality
release-check:
	@if [ -z "$(GH)" ]; then \
		echo "Error: GitHub CLI (gh) is not installed or not in PATH"; \
		echo "Please install it from: https://cli.github.com/"; \
		exit 1; \
	fi
	@if ! gh auth status >/dev/null 2>&1; then \
		echo "Error: GitHub CLI is not authenticated"; \
		echo "Please run: gh auth login"; \
		exit 1; \
	fi

release: release-check
	@echo "Creating GitHub Release $(VERSION)..."
	@gh release create $(VERSION) \
		--title "$(BINARY) $(VERSION)" \
		--notes "Release $(VERSION) of $(BINARY)" \
		--draft

release-upload: release-check
	@echo "Uploading build artifacts to release $(VERSION)..."
	@gh release upload $(VERSION) \
		$(BUILD_DIR)/$(BINARY)-* \
		$(BUILD_DIR)/archives/* \
		--clobber

release-full: package release-check
	@echo "Creating complete release $(VERSION)..."
	@gh release create $(VERSION) \
		--title "$(BINARY) $(VERSION)" \
		--notes "Release $(VERSION) of $(BINARY)" \
		$(BUILD_DIR)/$(BINARY)-* \
		$(BUILD_DIR)/archives/*

# Clean all artifacts
clean-all: clean
	@rm -rf $(BUILD_DIR)/archives
	@echo "Cleaned all build artifacts"

# Help
help:
	@echo "Available targets:"
	@echo "=== Basic Commands ==="
	@echo "  all         - Build the project (default)"
	@echo "  build       - Build the binary"
	@echo "  install     - Install the binary to ~/bin"
	@echo "  install-go  - Install the binary to GOPATH/bin"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests with coverage"
	@echo "  tidy        - Tidy go modules"
	@echo "  fmt         - Format code"
	@echo "  vet         - Vet code"
	@echo ""
	@echo "=== Multi-Platform Builds ==="
	@echo "  build-all              - Build for all supported platforms"
	@echo "  build-linux-amd64      - Build for Linux AMD64"
	@echo "  build-linux-arm64      - Build for Linux ARM64"
	@echo "  build-mac-amd64        - Build for macOS AMD64"
	@echo "  build-mac-arm64        - Build for macOS ARM64"
	@echo "  build-windows-amd64    - Build for Windows AMD64"
	@echo "  build-windows-arm64    - Build for Windows ARM64"
	@echo "  build-freebsd-amd64    - Build for FreeBSD AMD64"
	@echo "  build-cross            - Build for custom platform (set GOOS/GOARCH)"
	@echo ""
	@echo "=== Packaging & Release ==="
	@echo "  package      - Package all builds into archives"
	@echo "  release      - Create GitHub Release (requires gh CLI)"
	@echo "  release-upload - Upload artifacts to existing release"
	@echo "  release-full - Create release and upload all artifacts"
	@echo "  clean-all    - Clean all artifacts including archives"
	@echo ""
	@echo "=== Legacy Aliases (Backward Compatible) ==="
	@echo "  build-linux  - Alias for build-linux-amd64"
	@echo "  build-mac    - Alias for build-mac-amd64"
	@echo "  build-windows - Alias for build-windows-amd64"
	@echo ""
	@echo "=== Examples ==="
	@echo "  make build-all                    # Build all platforms"
	@echo "  make package VERSION=v1.2.3      # Package with specific version"
	@echo "  make release VERSION=v1.2.3      # Create GitHub release"
	@echo "  make build-cross GOOS=linux GOARCH=arm64  # Custom platform"
	@echo ""
	@echo "=== Variables ==="
	@echo "  VERSION    - Set version (default: git describe --tags)"
	@echo "  COMMIT     - Git commit (default: git rev-parse --short HEAD)"
	@echo "  BUILD_DATE - Build date (default: current date)"
	@echo ""
	@echo "  help       - Show this help message"

.PHONY: all build install install-go clean test test-cover tidy fmt vet build-all build-linux build-mac build-windows help
.PHONY: build-linux-amd64 build-linux-arm64 build-mac-amd64 build-mac-arm64
.PHONY: build-windows-amd64 build-windows-arm64 build-freebsd-amd64 build-cross
.PHONY: package release release-upload release-full clean-all release-check