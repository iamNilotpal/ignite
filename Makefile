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
# \033[<color_code>m
# 32m = Green
# 33m = Yellow
# 36m = Cyan
# 0m  = Reset color
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
	@rm -rf $(CROSS_BUILD_DIR)
	@echo "$(GREEN)$(SUCCESS_EMOJI) Clean complete.$(RESET)"