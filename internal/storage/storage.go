// Package storage provides a file-based storage mechanism for segments of data.
// It handles the creation and management of segment files, ensuring data can be
// appended and retrieved efficiently, and that new segments are correctly
// identified and created based on existing data.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iamNilotpal/ignite/pkg/filesys"
	"github.com/iamNilotpal/ignite/pkg/options"
	"go.uber.org/zap"
)

// Storage represents the file-based storage component.
type Storage struct {
	file    *os.File           // The currently active segment file for writing.
	options *options.Options   // Configuration options for the storage.
	log     *zap.SugaredLogger // Logger instance for logging messages.
}

// Config holds the configuration parameters required to initialize a Storage instance.
type Config struct {
	Options *options.Options   // Application configuration options.
	Logger  *zap.SugaredLogger // Logger instance.
}

// New creates and initializes a new Storage instance.
// It sets up the data directory, finds the latest segment, and opens the
// appropriate file for appending new data.
// It returns a pointer to the initialized Storage and an error if any occurs during setup.
func New(context context.Context, config *Config) (*Storage, error) {
	// Construct the full path for the segment directory (e.g., data/segments).
	path := filepath.Join(config.Options.DataDir, config.Options.SegmentOptions.Directory)

	// Create the segment directory if it doesn't exist.
	// The permissions are set to 0755 (rwxr-xr-x), and 'true' forces creation if it exists.
	if err := filesys.CreateDir(path, 0755, true); err != nil {
		return nil, err
	}

	// Initialize the Storage struct with provided logger and options.
	storage := &Storage{
		log:     config.Logger,
		options: config.Options,
	}

	return storage, nil
}

func (s *Storage) Close() error {
	return s.file.Close()
}

// Creates a unique filename for a segment based on its ID and the current timestamp.
// The format is "prefix_ID_timestamp.seg" (e.g., "segment_00001_1678881234567890.seg").
func (s *Storage) generateName(id uint64) string {
	return fmt.Sprintf("%s_%05d_%d.seg", s.options.SegmentOptions.Prefix, id, time.Now().UnixNano())
}
