// Package index provides the in-memory hash table implementation for the ignite key-value store.
// This package embodies the core Bitcask architectural principle: maintain all keys in memory
// with minimal metadata while storing actual values on disk for optimal memory utilization.
//
// The design philosophy centers on memory efficiency as the primary constraint. Every byte
// stored in the RecordPointer structure directly impacts the system's ability to handle
// large datasets. The approach here prioritizes compact data structures over convenience
// features, recognizing that memory constraints often determine system scalability limits.
//
// The index enables O(1) key lookups through an in-memory hash table while keeping
// storage overhead minimal. This allows the system to handle datasets significantly
// larger than available RAM while maintaining excellent read performance characteristics.
package index

import (
	"context"
	stdErrors "errors"

	"github.com/iamNilotpal/ignite/pkg/errors"
)

var (
	ErrIndexClosed = stdErrors.New("operation failed: cannot access closed index")
)

// New creates and initializes a new Index instance configured according to the
// provided parameters. The returned Index is immediately ready for concurrent
// use and includes optimizations like pre-allocated map capacity.
func New(ctx context.Context, config *Config) (*Index, error) {
	if config == nil || config.DataDir == "" || config.Logger == nil {
		return nil, errors.NewValidationError(
			nil, errors.ErrorCodeInvalidInput, "Index configuration is required",
		).WithField("config").WithRule("required").WithProvided(config)
	}

	return &Index{
		log:           config.Logger,
		dataDir:       config.DataDir,
		recordPointer: make(map[string]*RecordPointer, 2046),
	}, nil
}

// Close gracefully shuts down the Index, cleaning up resources and ensuring
// that the index cannot be used after closure.
func (idx *Index) Close() error {
	// Use atomic compare-and-swap to safely check and update the closed state.
	if !idx.closed.CompareAndSwap(false, true) {
		return ErrIndexClosed
	}

	idx.log.Infow("Closing index system")

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear the record pointer map to release all memory associated with
	// the index entries.
	clear(idx.recordPointer)
	idx.recordPointer = nil

	idx.log.Infow("Index system closed successfully")
	return nil
}
