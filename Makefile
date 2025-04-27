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
