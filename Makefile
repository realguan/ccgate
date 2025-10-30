# Makefile for ccgate

# Binary name
BINARY=ccgate

# Directories
SRC_DIR=.
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

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

# Build for different platforms
build-all: build-linux build-mac build-windows

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-amd64

build-mac:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe

# Help
help:
	@echo "Available targets:"
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
	@echo "  build-all   - Build for all platforms"
	@echo "  build-linux - Build for Linux"
	@echo "  build-mac   - Build for macOS"
	@echo "  build-windows - Build for Windows"
	@echo "  help        - Show this help message"

.PHONY: all build install install-go clean test test-cover tidy fmt vet build-all build-linux build-mac build-windows help