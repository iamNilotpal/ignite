// Package storage provides a comprehensive file-based storage mechanism for managing segments of data
// in high-throughput, append-only scenarios.
//
// This package was designed to solve the fundamental challenge of efficiently storing streaming data
// that arrives continuously and needs to be persisted reliably. Think of it as a specialized foundation
// for systems like write-ahead logs, event sourcing platforms, or time-series databases where data
// flows in continuously and must be stored in an organized, retrievable manner.
//
// Core Architecture:
//
// The storage system operates on the concept of "segments" - individual files that contain chunks
// of data. When a segment reaches its configured size limit, the system automatically creates a new
// segment and continues writing to it. This segmentation strategy provides several key benefits:
// it keeps individual files at manageable sizes, enables parallel processing of historical data,
// facilitates efficient cleanup of old data, and provides natural boundaries for backup operations.
//
// The storage engine maintains exactly one active segment file at any given time. This active segment
// is where all new data gets appended. Once this segment reaches its size threshold, the system
// seamlessly transitions to a new segment, ensuring continuous write availability with minimal latency.
//
// Initialization and Recovery:
//
// When the storage system starts up, it performs an intelligent recovery process. It scans the
// configured directory to discover existing segments, identifies the most recent one, and determines
// whether to continue writing to it or create a new segment. This bootstrap process ensures that
// the system can recover gracefully from restarts and continue exactly where it left off.
//
// The recovery logic handles several important scenarios: empty directories where no segments exist
// yet, partially filled segments that still have capacity for more data, segments that have reached
// their size limit and require a new segment to be created, and corrupted or incomplete segments
// that need special handling.
package storage

import (
	"context"
	stdErrors "errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iamNilotpal/ignite/pkg/errors"
	"github.com/iamNilotpal/ignite/pkg/filesys"
	"github.com/iamNilotpal/ignite/pkg/options"
	"github.com/iamNilotpal/ignite/pkg/seginfo"
	"go.uber.org/zap"
)

var (
	ErrSegmentClosed = stdErrors.New("operation failed: cannot access closed segment")
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
	activeSegment   *os.File           // The currently active segment file where new data is written.
	options         *options.Options   // Configuration parameters controlling storage behavior.
	log             *zap.SugaredLogger // Structured logger for operational visibility and debugging.
}

// Config encapsulates all the configuration parameters required to initialize a Storage instance.
// This structure provides a clean way to pass multiple configuration values while maintaining
// flexibility for future configuration additions without breaking the API.
type Config struct {
	Options *options.Options
	Logger  *zap.SugaredLogger
}

// New creates and initializes a new Storage instance, performing all necessary setup operations
// to prepare the storage system for data writes. This function handles the complex bootstrap
// process that ensures the storage system can continue seamlessly from any previous state.
func New(ctx context.Context, config *Config) (*Storage, error) {
	// Input validation ensures we have valid configuration before proceeding.
	if config == nil || config.Options == nil || config.Logger == nil {
		return nil, fmt.Errorf("invalid configuration")
	}

	// Log the start of initialization for operational visibility.
	config.Logger.Infow(
		"Initializing storage system",
		"dataDir", config.Options.DataDir,
		"maxSegmentSize", config.Options.SegmentOptions.Size,
		"segmentDir", config.Options.SegmentOptions.Directory,
		"segmentPrefix", config.Options.SegmentOptions.Prefix,
	)

	// Construct the full directory path where segment files will be stored.
	segmentDirPath := filepath.Join(config.Options.DataDir, config.Options.SegmentOptions.Directory)

	// Create the segment directory with appropriate permissions if it doesn't exist
	// This ensures that the storage system can operate even on a fresh installation
	if err := filesys.CreateDir(segmentDirPath, 0755, true); err != nil {
		return nil, errors.NewStorageError(
			err, errors.ErrorCodeIO, "Failed to create segment directory",
		).WithPath(segmentDirPath).WithDetail("permission", "0755").WithDetail("forceCreate", true)
	}

	config.Logger.Infow("Segment directory created successfully", "path", segmentDirPath)

	// Initialize the Storage instance with configuration.
	storage := &Storage{log: config.Logger, options: config.Options}

	// Discover existing segments to understand the current state of the storage system
	// This is a critical step that determines whether we continue with an existing segment
	// or need to create a new one
	config.Logger.Infow(
		"Discovering existing segments",
		"dataDir", config.Options.DataDir,
		"segmentDir", config.Options.SegmentOptions.Directory,
		"prefix", config.Options.SegmentOptions.Prefix,
	)

	latestSegmentID, latestSegmentInfo, err := seginfo.GetLatestSegmentInfo(
		config.Options.DataDir,
		config.Options.SegmentOptions.Directory,
		config.Options.SegmentOptions.Prefix,
	)
	if err != nil {
		return nil, errors.NewStorageError(err, errors.ErrorCodeIO, "Failed to get latest segment info")
	}

	// Determine the appropriate segment to use based on discovery results.
	var targetSegmentID uint64
	var shouldCreateNewSegment bool

	if latestSegmentInfo == nil {
		// Bootstrap case: no existing segments found, start with ID 1
		storage.size = 0
		targetSegmentID = 1
		shouldCreateNewSegment = true
		config.Logger.Infow("No existing segments found, starting fresh", "newSegmentID", targetSegmentID)
	} else {
		// Existing segments found, check if we need to rotate to a new segment.
		currentSize := latestSegmentInfo.Size()
		maxSize := int64(config.Options.SegmentOptions.Size)

		if currentSize >= maxSize {
			// Current segment is full, create a new one.
			storage.size = 0
			shouldCreateNewSegment = true
			targetSegmentID = latestSegmentID + 1

			config.Logger.Infow(
				"Current segment is full, creating new segment",
				"currentSegmentID", latestSegmentID,
				"currentSize", currentSize,
				"maxSize", maxSize,
				"newSegmentID", targetSegmentID,
			)
		} else {
			// Current segment has space, continue using it.
			storage.size = currentSize
			shouldCreateNewSegment = false
			targetSegmentID = latestSegmentID

			config.Logger.Infow(
				"Continuing with existing segment",
				"segmentID", targetSegmentID,
				"currentSize", currentSize,
				"maxSize", maxSize,
				"remainingCapacity", maxSize-currentSize,
			)
		}
	}

	// Open the target segment file for writing.
	segmentFile, err := storage.openSegmentFile(targetSegmentID, shouldCreateNewSegment)
	if err != nil {
		config.Logger.Errorw(
			"Failed to open segment file",
			"error", err,
			"segmentID", targetSegmentID,
			"isNewSegment", shouldCreateNewSegment,
		)
		return nil, fmt.Errorf("failed to open segment file for ID %d: %w", targetSegmentID, err)
	}

	// Store the file handle and complete initialization.
	storage.activeSegment = segmentFile
	storage.activeSegmentId = targetSegmentID

	config.Logger.Infow(
		"Storage system initialized successfully",
		"activeSegmentID", targetSegmentID,
		"segmentSize", storage.size,
		"isNewSegment", shouldCreateNewSegment,
	)

	return storage, nil
}

// openSegmentFile handles the complex process of opening a segment file for writing.
// This method encapsulates all the file operations needed to prepare a segment file,
// including creation, permission setting, and positioning the file pointer correctly.
//
// The function handles both new segment creation and opening existing segments for
// continued writing, ensuring that the file is always in the correct state for
// append operations.
func (s *Storage) openSegmentFile(segmentID uint64, isNewSegment bool) (*os.File, error) {
	// Generate the filename using the seginfo package's naming convention.
	filename := seginfo.GenerateName(segmentID, s.options.SegmentOptions.Prefix)
	filePath := filepath.Join(s.options.DataDir, s.options.SegmentOptions.Directory, filename)

	s.log.Infow(
		"Opening segment file",
		"segmentID", segmentID,
		"filename", filename,
		"path", filePath,
		"isNewSegment", isNewSegment,
	)

	// Open the segment file with flags appropriate for append-only operations.
	// O_CREATE: Create the file if it doesn't exist
	// O_RDWR: Open for both reading and writing (reading may be needed for verification)
	// O_APPEND: Ensure all writes go to the end of the file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, errors.NewStorageError(
			err, errors.ErrorCodeIO, "Failed to open segment file",
		).
			WithFileName(filename).
			WithPath(filePath).
			WithDetail("permission", "0644").
			WithDetail("flags", []string{"O_CREATE", "O_RDWR", "O_APPEND"})
	}

	// Position the file pointer at the end of the file.
	// This is essential even with O_APPEND to ensure we know the current position.
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		// Attempt to close the file to prevent resource leaks.
		if err := file.Close(); err != nil {
			return nil, errors.NewStorageError(err, errors.ErrorCodeIO, "Failed to close file after seek error").
				WithFileName(filename).
				WithPath(filePath).
				WithDetail("seekOffset", 0).
				WithDetail("whence", io.SeekEnd)
		}

		return nil, errors.NewStorageError(err, errors.ErrorCodeIO, "Failed to seek to end of file").
			WithFileName(filename).
			WithPath(filePath).
			WithDetail("seekOffset", 0).
			WithDetail("whence", io.SeekEnd)
	}

	s.log.Infow(
		"Segment file opened successfully",
		"path", filePath,
		"currentOffset", offset,
		"isNewSegment", isNewSegment,
	)

	return file, nil
}
