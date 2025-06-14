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
	"io"
	"os"
	"path/filepath"

	"github.com/iamNilotpal/ignite/pkg/errors"
	"github.com/iamNilotpal/ignite/pkg/filesys"
	"github.com/iamNilotpal/ignite/pkg/seginfo"
)

var (
	ErrSegmentClosed = stdErrors.New("operation failed: cannot access closed segment")
)

// New creates and initializes a new Storage instance, performing all necessary setup operations
// to prepare the storage system for data writes. This function handles the complex bootstrap
// process that ensures the storage system can continue seamlessly from any previous state.
func New(ctx context.Context, config *Config) (*Storage, error) {
	if config == nil || config.Options == nil || config.Logger == nil {
		return nil, errors.NewValidationError(
			nil, errors.ErrorCodeInvalidInput, "Storage configuration is required",
		).WithField("config").WithRule("required").WithProvided(config)
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
		return nil, errors.ClassifyDirectoryCreationError(err, segmentDirPath)
	}

	config.Logger.Infow("Segment directory created successfully", "path", segmentDirPath)

	// Initialize the Storage instance with configuration.
	storage := &Storage{
		log:     config.Logger,
		options: config.Options,
	}

	// Discover existing segments to understand the current state of the storage system
	// This is a critical step that determines whether we continue with an existing segment
	// or need to create a new one
	config.Logger.Infow(
		"Discovering existing segments",
		"dataDir", config.Options.DataDir,
		"segmentDir", config.Options.SegmentOptions.Directory,
		"prefix", config.Options.SegmentOptions.Prefix,
	)

	lastSegmentID, lastSegmentInfo, err := seginfo.GetLastSegmentInfo(
		config.Options.DataDir,
		config.Options.SegmentOptions.Directory,
		config.Options.SegmentOptions.Prefix,
	)
	if err != nil {
		return nil, errors.NewStorageError(
			err, errors.ErrorCodeIO,
			"Failed to discover existing segments during initialization",
		).WithPath(segmentDirPath).
			WithDetail("operation", "segment_discovery")
	}

	// Determine the appropriate segment to use based on discovery results.
	var targetSegmentID uint64
	var shouldCreateNewSegment bool

	if lastSegmentInfo == nil {
		// Bootstrap case: no existing segments found, start with ID 1
		storage.size = 0
		targetSegmentID = 1
		shouldCreateNewSegment = true
		config.Logger.Infow("No existing segments found, starting fresh", "newSegmentID", targetSegmentID)
	} else {
		// Existing segments found, check if we need to rotate to a new segment.
		currentSize := lastSegmentInfo.Size()
		maxSize := int64(config.Options.SegmentOptions.Size)

		if currentSize >= maxSize {
			// Current segment is full, create a new one.
			storage.size = 0
			shouldCreateNewSegment = true
			targetSegmentID = lastSegmentID + 1

			config.Logger.Infow(
				"Current segment is full, creating new segment",
				"currentSegmentID", lastSegmentID,
				"currentSize", currentSize,
				"maxSize", maxSize,
				"newSegmentID", targetSegmentID,
			)
		} else {
			// Current segment has space, continue using it.
			storage.size = currentSize
			shouldCreateNewSegment = false
			targetSegmentID = lastSegmentID

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
		return nil, err
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

// Close gracefully shuts down the storage system, ensuring all buffered data is written
// to disk and all resources are properly released.
func (s *Storage) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return ErrSegmentClosed
	}

	s.log.Infow("Closing storage system", "currentSize", s.size)

	var currentFileName string
	var currentFilePath string
	if stat, err := s.activeSegment.Stat(); err == nil {
		currentFileName = stat.Name()
		currentFilePath = filepath.Join(s.options.DataDir, s.options.SegmentOptions.Directory, currentFileName)
	}

	// First, ensure all buffered data is written to disk.
	// This is critical for data durability - without sync, recently written data
	// might still be in OS buffers and could be lost if the process crashes.
	if err := s.activeSegment.Sync(); err != nil {
		s.log.Errorw(
			"Failed to sync file before closing",
			"error", err,
			"currentSize", s.size,
			"fileName", currentFileName,
			"filePath", currentFilePath,
		)

		// Even if sync fails, we should still attempt to close the file
		// to prevent resource leaks, but we return the sync error since
		// it indicates potential data loss.
		if closeErr := s.activeSegment.Close(); closeErr != nil {
			s.log.Errorw(
				"Failed to close file after sync error",
				"syncError", err,
				"closeError", closeErr,
				"fileName", currentFileName,
				"filePath", currentFilePath,
			)
		}

		// Classify the sync error to provide specific error information
		return errors.ClassifySyncError(err, currentFileName, currentFilePath, int(s.size))
	}

	// Close the file handle and release system resources.
	if err := s.activeSegment.Close(); err != nil {
		return errors.NewStorageError(
			err, errors.ErrorCodeIO,
			"Failed to close segment file handle",
		).WithFileName(currentFileName).
			WithPath(currentFilePath).
			WithOffset(int(s.size)).
			WithDetail("operation", "file_close").
			WithDetail("currentSize", s.size)
	}

	// Clear the file reference to prevent accidental use after close.
	s.activeSegment = nil

	s.log.Infow(
		"Storage system closed successfully",
		"finalSize", s.size,
		"fileName", currentFileName,
		"filePath", currentFilePath,
	)

	return nil
}

// Handles the complex process of opening a segment file for writing.
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
		return nil, errors.ClassifyFileOpenError(err, filePath, filename)
	}

	// Position the file pointer at the end of the file.
	// This is essential even with O_APPEND to ensure we know the current position.
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		// Attempt to close the file to prevent resource leaks.
		if err := file.Close(); err != nil {
			return nil, errors.NewStorageError(
				err, errors.ErrorCodeIO,
				"Failed to close file after seek error",
			).
				WithFileName(filename).
				WithPath(filePath).
				WithDetail("seekOffset", 0).
				WithDetail("whence", io.SeekEnd)
		}

		return nil, errors.NewStorageError(
			err, errors.ErrorCodeIO,
			"Failed to seek to end of segment file",
		).WithFileName(filename).
			WithPath(filePath).
			WithDetail("seekOffset", 0).
			WithDetail("whence", io.SeekEnd).
			WithDetail("operation", "file_seek").
			WithDetail("suggestion", "file may be corrupted or filesystem may have issues")
	}

	s.log.Infow(
		"Segment file opened successfully",
		"path", filePath,
		"currentOffset", offset,
		"isNewSegment", isNewSegment,
	)

	return file, nil
}
