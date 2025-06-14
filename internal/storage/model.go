package storage

import (
	"os"
	"sync/atomic"

	"github.com/iamNilotpal/ignite/pkg/options"
	"go.uber.org/zap"
)

// Storage represents the core file-based storage component responsible for managing segment files
// and handling data persistence operations. It maintains the currently active segment file and
// provides the foundation for append-only data storage with automatic segment rotation.
//
// The Storage struct encapsulates all the state needed to manage segment files effectively:
// the current active file handle, configuration options that control behavior, a logger for
// observability, and size tracking for determining when segment rotation is needed.
type Storage struct {
	size            int64              // Current size of the active segment file in bytes.
	activeSegmentId uint64             // Unique identifier for the currently active segment file being written to.
	closed          atomic.Bool        // Flag indicating whether the storage has been closed.
	activeSegment   *os.File           // The currently active segment file where new data is written.
	options         *options.Options   // Configuration parameters controlling storage behavior.
	log             *zap.SugaredLogger // Structured logger for operational visibility and debugging.
}

// Config encapsulates all the configuration parameters required to initialize a Storage instance.
type Config struct {
	Options *options.Options
	Logger  *zap.SugaredLogger
}
