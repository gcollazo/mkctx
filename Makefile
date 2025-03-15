# Makefile for building mkctx for multiple platforms

# Binary name
BINARY_NAME=mkctx

# Version
VERSION=1.0.0

# Build directory
BUILD_DIR=dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get

# Operating systems
OS_WINDOWS=windows
OS_LINUX=linux
OS_DARWIN=darwin

# Architecture targets
ARCH_AMD64=amd64
ARCH_386=386
ARCH_ARM=arm
ARCH_ARM64=arm64

# Ensure distribution directory exists
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Build for the current platform (development use)
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Initialize or update Go modules
.PHONY: deps
deps:
	$(GOMOD) tidy

# Build for all platforms
.PHONY: all
all: windows linux darwin

# Build for Windows (multiple architectures)
.PHONY: windows
windows: $(BUILD_DIR)
	# Windows AMD64 (64-bit)
	GOOS=$(OS_WINDOWS) GOARCH=$(ARCH_AMD64) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_windows_amd64.exe -v
	# Windows 386 (32-bit)
	GOOS=$(OS_WINDOWS) GOARCH=$(ARCH_386) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_windows_386.exe -v
	# Windows ARM64
	GOOS=$(OS_WINDOWS) GOARCH=$(ARCH_ARM64) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_windows_arm64.exe -v

# Build for Linux (multiple architectures)
.PHONY: linux
linux: $(BUILD_DIR)
	# Linux AMD64 (64-bit)
	GOOS=$(OS_LINUX) GOARCH=$(ARCH_AMD64) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64 -v
	# Linux 386 (32-bit)
	GOOS=$(OS_LINUX) GOARCH=$(ARCH_386) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_linux_386 -v
	# Linux ARM (32-bit)
	GOOS=$(OS_LINUX) GOARCH=$(ARCH_ARM) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm -v
	# Linux ARM64 (64-bit)
	GOOS=$(OS_LINUX) GOARCH=$(ARCH_ARM64) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64 -v

# Build for macOS (Intel and Apple Silicon)
.PHONY: darwin
darwin: $(BUILD_DIR)
	# macOS AMD64 (Intel)
	GOOS=$(OS_DARWIN) GOARCH=$(ARCH_AMD64) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64 -v
	# macOS ARM64 (Apple Silicon)
	GOOS=$(OS_DARWIN) GOARCH=$(ARCH_ARM64) $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64 -v

# Create zip archives for distribution
.PHONY: release
release: all
	# Windows
	cd $(BUILD_DIR) && zip $(BINARY_NAME)_$(VERSION)_windows_amd64.zip $(BINARY_NAME)_$(VERSION)_windows_amd64.exe
	cd $(BUILD_DIR) && zip $(BINARY_NAME)_$(VERSION)_windows_386.zip $(BINARY_NAME)_$(VERSION)_windows_386.exe
	cd $(BUILD_DIR) && zip $(BINARY_NAME)_$(VERSION)_windows_arm64.zip $(BINARY_NAME)_$(VERSION)_windows_arm64.exe

	# Linux
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz $(BINARY_NAME)_$(VERSION)_linux_amd64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)_$(VERSION)_linux_386.tar.gz $(BINARY_NAME)_$(VERSION)_linux_386
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)_$(VERSION)_linux_arm.tar.gz $(BINARY_NAME)_$(VERSION)_linux_arm
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz $(BINARY_NAME)_$(VERSION)_linux_arm64

	# macOS
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)_$(VERSION)_darwin_amd64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)_$(VERSION)_darwin_arm64

# Help target
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build    - Build for current OS/architecture"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make deps     - Update dependencies"
	@echo "  make all      - Build for all platforms"
	@echo "  make windows  - Build for Windows (386, amd64, arm64)"
	@echo "  make linux    - Build for Linux (386, amd64, arm, arm64)"
	@echo "  make darwin   - Build for macOS (amd64, arm64)"
	@echo "  make release  - Create distributable archives"
	@echo "  make help     - Show this help message"