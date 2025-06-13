// Package engine provides the core database engine implementation for the Ignite storage system.
//
// The engine serves as the central coordinator and entry point for all database operations.
// It orchestrates the interaction between three main subsystems:
//   - Index: Manages in-memory data structures for fast key lookups and range queries
//   - Storage: Handles persistent data storage, including write-ahead logs and data files
//   - Compaction: Performs background maintenance to optimize storage efficiency and performance
//
// The engine implements a thread-safe interface with proper lifecycle management,
// ensuring resources are properly initialized and cleaned up. It uses atomic operations
// for state management to provide consistent behavior across concurrent operations.
package engine

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/iamNilotpal/ignite/internal/compaction"
	"github.com/iamNilotpal/ignite/internal/index"
	"github.com/iamNilotpal/ignite/internal/storage"
	"github.com/iamNilotpal/ignite/pkg/options"
	"go.uber.org/zap"
)

var (
	// ErrEngineClosed is returned when attempting to perform operations on a closed engine.
	ErrEngineClosed = errors.New("operation failed: cannot access closed engine")
)

// Engine represents the main database engine that coordinates all subsystems.
// It acts as the primary interface for database operations and manages the lifecycle
// of all internal components. The engine is designed to be thread-safe and supports
// concurrent operations while maintaining data consistency.
type Engine struct {
	options    *options.Options       // options contains all configuration parameters for the engine and its subsystems.
	log        *zap.SugaredLogger     // log provides structured logging capabilities throughout the engine.
	closed     atomic.Bool            // closed is an atomic boolean that tracks the engine's lifecycle state.
	index      *index.Index           // index manages the in-memory data structures for fast data access.
	storage    *storage.Storage       // storage handles all persistent data operations.
	compaction *compaction.Compaction // compaction manages background processes that optimize storage efficiency.
}

// Config holds all the parameters needed to initialize a new Engine instance.
type Config struct {
	Options *options.Options
	Logger  *zap.SugaredLogger
}

// New creates and initializes a new Engine instance with the provided configuration.
// This constructor follows the dependency injection pattern, making the engine
// testable and allowing for different configurations in different environments.
//
// Returns:
//   - *Engine: A fully initialized engine ready for use
//   - error: Any error encountered during initialization, typically from storage setup
func New(ctx context.Context, config *Config) (*Engine, error) {
	// Initialize the index subsystem first since it has no external dependencies.
	index := index.New()

	// Initialize the compaction subsystem, which also has minimal dependencies.
	compaction := compaction.New()

	// Initialize the storage subsystem last since it has the most complex setup.
	storage, err := storage.New(ctx, &storage.Config{
		Logger:  config.Logger,
		Options: config.Options,
	})
	if err != nil {
		// If storage initialization fails, we cannot create a functional engine.
		// Return the error immediately since the engine would be unusable.
		return nil, err
	}

	// Create and return the engine with all subsystems properly initialized.
	// At this point, all dependencies are satisfied and the engine is ready
	// to handle database operations. The closed flag defaults to false,
	// indicating the engine is in an active, usable state.
	return &Engine{
		options:    config.Options,
		log:        config.Logger,
		index:      index,
		storage:    storage,
		compaction: compaction,
	}, nil
}

// Close gracefully shuts down the engine and releases all associated resources.
// This method ensures that all pending operations complete and that data is
// properly persisted before the engine becomes unusable.
func (e *Engine) Close() error {
	// Use atomic compare-and-swap to transition from open (false) to closed (true).
	// This operation is atomic and thread-safe, ensuring only one goroutine
	// can successfully close the engine. The operation returns true if the
	// swap was successful (engine was open) or false if it failed (already closed).
	if !e.closed.CompareAndSwap(false, true) {
		return ErrEngineClosed
	}

	// Perform the actual shutdown by closing the storage subsystem.
	return e.storage.Close()
}
