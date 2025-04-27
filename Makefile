# --- Application Variables ---

# Binary name for your application
BINARY_NAME := ignite

# Main package directory (where your main.go is located)
# . means the current directory where the Makefile is located
MAIN_PACKAGE ?= ./cmd/ignited/main.go

# Go build flags (e.g., -v for verbose output, -ldflags for linker flags)
# -ldflags="-s -w" removes debug information and symbol table, reducing binary size
BUILD_FLAGS := -v -ldflags="-s -w"

# Extra build flags that can be passed via command line (e.g., make build EXTRA_BUILD_FLAGS="-race")
EXTRA_BUILD_FLAGS ?=

# Go test flags (e.g., -v for verbose output, -race to enable data race detection)
TEST_FLAGS := -v -race

# Extra test flags that can be passed via command line (e.g., make test EXTRA_TEST_FLAGS="-count=10")
EXTRA_TEST_FLAGS ?=

# Arguments to pass to the application when running (e.g., make run RUN_ARGS="--port 8080")
RUN_ARGS ?=

# Output directory for cross-compiled binaries
CROSS_BUILD_DIR := ./dist

# Go install path for the built binary
# GOBIN is the directory where go install will put the compiled binaries
# If not set, it defaults to $GOPATH/bin or $HOME/go/bin
GOBIN := $(shell go env GOBIN)
# Use default if GOBIN is not set
ifndef GOBIN
  GOBIN = $(shell go env GOPATH)/bin
  ifndef GOBIN
    GOBIN = $(HOME)/go/bin
  endif
endif

# ANSI Color Codes
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
RESET := \033[0m

# Emojis
BUILD_EMOJI := ⚙️
RUN_EMOJI := 🚀
TEST_EMOJI := 🔍
MODULE_EMOJI := 📥
FORMAT_EMOJI := 📝
CLEAN_EMOJI := 🧹
SUCCESS_EMOJI := ✅
HELP_EMOJI := ℹ️
STAR_EMOJI := ✳️

# --- Targets ---

# Default target: builds and runs the application for the host OS/ARCH
all: build run

# Build the application for the host OS/ARCH
build: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)...$(RESET)"
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete.$(RESET)"

# Run the application (uses the binary built for the host OS/ARCH)
run: build
	@echo "$(CYAN)$(RUN_EMOJI) Running $(BINARY_NAME) $(RUN_ARGS)...$(RESET)"
	@$(CROSS_BUILD_DIR)/$(BINARY_NAME) $(RUN_ARGS)

# Install the application binary to GOBIN for the host OS/ARCH
# Note: go install also builds the binary if needed
install: tidy
	@echo "$(CYAN)$(MODULE_EMOJI) Installing $(BINARY_NAME) to $(GOBIN) for $(shell go env GOOS)/$(shell go env GOARCH)...$(RESET)"
	@go install $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) $(BINARY_NAME) installed to $(GOBIN).$(RESET)"

# Generic target to trigger cross-compilation for a specific OS and Architecture
# Usage: make build-cross OS=<os> ARCH=<arch> [EXTRA_BUILD_FLAGS='...']
# This target sets the OS/ARCH environment variables and runs the build command.
build-cross: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for $(OS)/$(ARCH)...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/$(OS)/$(ARCH)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/$(OS)/$(ARCH)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for $(OS)/$(ARCH).$(RESET)"


# Example targets for common Unix-based OS and architectures (x86_64)
# These targets directly invoke the build logic with environment variables.
build-linux-amd64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for linux/amd64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/linux/amd64
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/linux/amd64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for linux/amd64.$(RESET)"

build-darwin-amd64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for darwin/amd64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/darwin/amd64
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/darwin/amd64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for darwin/amd64.$(RESET)"

build-freebsd-amd64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for freebsd/amd64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/freebsd/amd64
	@GOOS=freebsd GOARCH=amd64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/freebsd/amd64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for freebsd/amd64.$(RESET)"

build-openbsd-amd64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for openbsd/amd64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/openbsd/amd64
	@GOOS=openbsd GOARCH=amd64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/openbsd/amd64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for openbsd/amd64.$(RESET)"

# Example targets for common Unix-based OS and architectures (ARM64)
# These targets directly invoke the build logic with environment variables.
build-linux-arm64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for linux/arm64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/linux/arm64
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/linux/arm64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for linux/arm64.$(RESET)"

build-darwin-arm64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for darwin/arm64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/darwin/arm64
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/darwin/arm64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for darwin/arm64.$(RESET)"

build-freebsd-arm64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for freebsd/arm64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/freebsd/arm64
	@GOOS=freebsd GOARCH=arm64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/freebsd/arm64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for freebsd/arm64.$(RESET)"

build-openbsd-arm64: tidy
	@echo "$(CYAN)$(BUILD_EMOJI) Building $(BINARY_NAME) for openbsd/arm64...$(RESET)"
	@mkdir -p $(CROSS_BUILD_DIR)/openbsd/arm64
	@GOOS=openbsd GOARCH=arm64 go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(CROSS_BUILD_DIR)/openbsd/arm64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Build complete for openbsd/arm64.$(RESET)"

# Run tests
test: tidy
	@echo "$(CYAN)$(TEST_EMOJI) Running tests...$(RESET)"
	@go test $(TEST_FLAGS) $(EXTRA_TEST_FLAGS) ./...
	@echo "$(GREEN)$(SUCCESS_EMOJI) Tests complete.$(RESET)"

# Manage dependencies: tidy go.mod and go.sum
tidy:
	@echo "$(CYAN)$(MODULE_EMOJI) Tidying Go modules...$(RESET)"
	@go mod tidy
	@echo "$(GREEN)$(SUCCESS_EMOJI) Go modules tidied.$(RESET)"

# Download dependencies
deps:
	@echo "$(CYAN)$(MODULE_EMOJI) Downloading Go modules...$(RESET)"
	@go mod download
	@echo "$(GREEN)$(SUCCESS_EMOJI) Go modules downloaded.$(RESET)"

# Format Go code
fmt:
	@echo "$(CYAN)$(FORMAT_EMOJI) Formatting Go code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)$(SUCCESS_EMOJI) Formatting complete.$(RESET)"

# Clean build artifacts and cross-compiled binaries
clean:
	@echo "$(YELLOW)$(CLEAN_EMOJI) Cleaning build artifacts...$(RESET)"
	@go clean
	@rm -rf $(CROSS_BUILD_DIR) # Remove cross-compiled binaries dir
	@echo "$(GREEN)$(SUCCESS_EMOJI) Clean complete.$(RESET)"

# Display help message
help:
	@echo "$(CYAN)$(STAR_EMOJI)Usage:$(RESET)"
	@echo "  $(GREEN)make all$(RESET)                              - Builds and runs the application for the host OS/ARCH"
	@echo "  $(GREEN)make build$(RESET)                            - Builds the application binary for the host OS/ARCH (e.g. make build EXTRA_BUILD_FLAGS='-tags netgo')"
	@echo "  $(GREEN)make run$(RESET)                              - Runs the application (host OS/ARCH binary) (e.g. make run RUN_ARGS='--port 8080')"
	@echo "  $(GREEN)make install$(RESET)                          - Installs the application binary to GOBIN (host OS/ARCH)"
	@echo ""
	@echo "$(CYAN)$(STAR_EMOJI)Cross-compilation:$(RESET)"
	@echo "  $(GREEN)make build-cross OS=<os> ARCH=<arch> [EXTRA_BUILD_FLAGS='...']$(RESET)"
	@echo "  $(GREEN)make build-linux-amd64$(RESET)                - Builds for Linux (amd64)"
	@echo "  $(GREEN)make build-darwin-amd64$(RESET)               - Builds for macOS (amd64)"
	@echo "  $(GREEN)make build-freebsd-amd64$(RESET)              - Builds for FreeBSD (amd64)"
	@echo "  $(GREEN)make build-openbsd-amd64$(RESET)              - Builds for OpenBSD (amd64)"
	@echo "  $(GREEN)make build-linux-arm64$(RESET)                - Builds for Linux (arm64)"
	@echo "  $(GREEN)make build-darwin-arm64$(RESET)               - Builds for macOS M-series (arm64)"
	@echo "  $(GREEN)make build-freebsd-arm64$(RESET)              - Builds for FreeBSD (arm64)"
	@echo "  $(GREEN)make build-openbsd-arm64$(RESET)              - Builds for OpenBSD (arm64)"
	@echo ""
	@echo "$(CYAN)$(STAR_EMOJI)Development & Maintenance:$(RESET)"
	@echo "  $(GREEN)make test$(RESET)                             - Runs tests (e.g. make test EXTRA_TEST_FLAGS='-count=10 -v')"
	@echo "  $(GREEN)make tidy$(RESET)                             - Adds missing and removes unused modules"
	@echo "  $(GREEN)make deps$(RESET)                             - Downloads required modules"
	@echo "  $(GREEN)make fmt$(RESET)                              - Formats Go code"
	@echo "  $(YELLOW)make clean$(RESET)                            - Removes build artifacts and cross-compiled binaries"
	@echo "  $(CYAN)make help$(RESET)                             - Displays this help message"

# Phony targets - prevents conflicts with files of the same name
.PHONY: all build run install test clean fmt lint help tidy deps build-cross \
	build-linux-amd64 build-darwin-amd64 build-freebsd-amd64 build-openbsd-amd64 \
	build-linux-arm64 build-darwin-arm64 build-freebsd-arm64 build-openbsd-arm64