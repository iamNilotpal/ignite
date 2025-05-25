package options

import "time"

const (
	// Specifies the default base directory where IgniteDB will store its data files.
	// If no other directory is specified during initialization, this path will be used.
	DefaultDataDir = "/var/lib/ignitedb"

	// Defines the default time duration between automatic compaction operations.
	// By default, compaction will run every 5 hours.
	DefaultCompactInterval = time.Hour * 5

	// Represents the minimum allowed size for a segment file in bytes (512MB).
	MinSegmentSize uint64 = 512 * 1024 * 1024

	// Represents the maximum allowed size for a segment file in bytes (4GB).
	MaxSegmentSize uint64 = 4 * 1024 * 1024 * 1024

	// Specifies the default target size for a new segment file in bytes (1GB).
	DefaultSegmentSize uint64 = 1 * 1024 * 1024 * 1024

	// Specifies the default subdirectory within the main data directory
	// where segment files will be stored.
	DefaultSegmentDirectory = "/segments"

	// Defines the default prefix for segment file names.
	// For example, a segment file might be named "segment-00001.db".
	DefaultSegmentPrefix = "segment"
)

// Holds the default configuration settings for an IgniteDB instance.
var defaultOptions = Options{
	DataDir:         DefaultDataDir,
	CompactInterval: DefaultCompactInterval,
	SegmentOptions: &segmentOptions{
		Size:      DefaultSegmentSize,
		Prefix:    DefaultSegmentPrefix,
		Directory: DefaultSegmentDirectory,
	},
}

func NewDefaultOptions() Options {
	return defaultOptions
}
