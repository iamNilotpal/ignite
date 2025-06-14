// Package seginfo provides utilities for managing sequential segment files in a file-based storage system.
//
// Filename Format: prefix_NNNNN_timestamp.seg
//
// Where:
//   - prefix: A configurable string identifying the file type (e.g., "segment", "log", "backup").
//   - NNNNN: A zero-padded 5-digit sequence number (00001, 00002, etc.).
//   - timestamp: A nanosecond-precision Unix timestamp for uniqueness and traceability.
//   - .seg: A fixed file extension (this could be made configurable in future versions).
//
// Example filenames:
//
//	segment_00001_1678881234567890.seg
//	backup_00042_1678881298765432.seg
//	log_00100_1678881356789012.seg
package seginfo

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/iamNilotpal/ignite/pkg/filesys"
)

// GetLastSegmentInfo discovers and analyzes the most recent segment file in the specified directory.
// It performs a comprehensive search of the segment directory, identifies the file with the highest
// sequence number, and returns detailed information about that file.
//
// Returns:
//   - uint64: The sequence ID of the latest segment (1 if no segments exist).
//   - os.FileInfo: File metadata for the latest segment (nil if no segments exist).
//   - error: Detailed error information if any operation fails.
func GetLastSegmentInfo(dataDir, segmentDir, prefix string) (uint64, os.FileInfo, error) {
	if dataDir == "" || segmentDir == "" || prefix == "" {
		return 0, nil, fmt.Errorf("all parameters (dataDir, segmentDir, prefix) must be non-empty")
	}

	// Discover the most recent segment file.
	lastSegmentPath, err := GetLastSegmentName(dataDir, segmentDir, prefix)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to discover latest segment: %w", err)
	}

	// Handle the bootstrap case: no existing segments found.
	if lastSegmentPath == "" {
		return 1, nil, nil
	}

	// Extract and parse the segment ID from the filename.
	segmentID, err := ParseSegmentID(lastSegmentPath, prefix)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse segment ID from %s: %w", lastSegmentPath, err)
	}

	// Retrieve file system metadata for the segment.
	fileInfo, err := GetFileInfo(lastSegmentPath)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to retrieve file info for %s: %w", lastSegmentPath, err)
	}

	return segmentID, fileInfo, nil
}

// GetLastSegmentName searches the segment directory and identifies the file with the highest sequence ID.
// This function implements a lexicographic sorting strategy that works because segment filenames
// use zero-padded IDs and monotonically increasing timestamps.
//
// Returns:
//   - string: Full path to the segment file with the highest ID (empty if none found).
//   - error: Detailed error if directory reading fails.
func GetLastSegmentName(dataDir, segmentDir, prefix string) (string, error) {
	if dataDir == "" || segmentDir == "" || prefix == "" {
		return "", fmt.Errorf("all parameters (dataDir, segmentDir, prefix) must be non-empty")
	}

	// Construct the search pattern for segment files.
	// Example: "/var/data/segments/segment_*.seg"
	searchPattern := filepath.Join(dataDir, segmentDir, prefix+"*.seg")

	// Safely read all matching files using our filesystem utility.
	matchingFiles, err := filesys.ReadDir(searchPattern)
	if err != nil {
		return "", fmt.Errorf("failed to read segment directory with pattern %s: %w", searchPattern, err)
	}

	// Handle the case where no segment files exist yet.
	if len(matchingFiles) == 0 {
		return "", nil
	}

	// Sort files lexicographically. This works correctly because:
	// 1. Segment IDs are zero-padded (00001, 00002, etc.).
	// 2. Timestamps are monotonically increasing.
	// 3. The filename format ensures proper sorting: prefix_ID_timestamp.seg.
	slices.Sort(matchingFiles)

	// Return the file with the highest ID (last in sorted order).
	return matchingFiles[len(matchingFiles)-1], nil
}

// GenerateName creates a properly formatted filename for a new segment file.
func GenerateName(id uint64, prefix string) string {
	// Return a recognizable error pattern rather than failing silently.
	if prefix == "" {
		return fmt.Sprintf("INVALID_PREFIX_%05d_%d.seg", id, time.Now().UnixNano())
	}

	// Generate timestamp with nanosecond precision for maximum uniqueness.
	timestamp := time.Now().UnixNano()

	// Format: prefix_NNNNN_timestamp.seg.
	// %05d ensures zero-padding (00001, 00002, etc.) for proper lexicographic sorting.
	return fmt.Sprintf("%s_%05d_%d.seg", prefix, id, timestamp)
}

// ParseSegmentID extracts the sequence ID from a segment filename.
func ParseSegmentID(fullPath, prefix string) (uint64, error) {
	// Extract just the filename from the full path.
	_, filename := filepath.Split(fullPath)

	// Validate that the filename starts with our expected prefix.
	if !strings.HasPrefix(filename, prefix) {
		return 0, fmt.Errorf("filename %s does not start with expected prefix %s", filename, prefix)
	}

	// Remove the prefix and file extension to get the core components.
	// Example: "segment_00001_1678881234567890.seg" -> "00001_1678881234567890"
	withoutPrefix := strings.TrimPrefix(filename, prefix)
	withoutExtension := strings.Split(withoutPrefix, ".")[0]

	// Split by underscores to separate ID and timestamp.
	// Example: "00001_1678881234567890" -> ["", "00001", "1678881234567890"]
	parts := strings.Split(withoutExtension, "_")

	// Validate that we have the expected number of parts.
	// We expect: ["", "ID", "timestamp"] (empty first element due to leading underscore).
	if len(parts) < 3 {
		return 0, fmt.Errorf("filename %s has unexpected format, expected prefix_ID_timestamp.seg", filename)
	}

	// Parse the ID component (second element after splitting).
	idStr := parts[1]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse segment ID '%s' as integer: %w", idStr, err)
	}

	return id, nil
}

// GetFileInfo safely retrieves file system metadata for a given path.
// This helper function encapsulates the file opening and stat operations,
// providing consistent error handling and resource cleanup.
func GetFileInfo(filePath string) (os.FileInfo, error) {
	// Open the file in read-only mode.
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	// Ensure the file is closed even if Stat() fails.
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	// Retrieve file metadata.
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	return stat, nil
}
