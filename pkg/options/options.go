// Package options provides data structures and functions for configuring
// the Ignite database. It defines various parameters that control Ignite's
// storage behavior, performance, and maintenance operations, such as
// directory paths, segment characteristics, and compaction intervals.
package options

import (
	"strings"
	"time"
)

// Defines configurable parameters for each segment.
// It provides fine-grained control over segment behavior, performance, and resource utilization.
type segmentOptions struct {
	// Defines the maximum size a segment can grow to before rotation.
	// When a segment reaches this size, a new segment will be created.
	// Larger segments mean fewer files but slower compaction and recovery.
	//
	//  - Default: 1GB
	//  - Maximum: 4GB
	//  - Minimum: 512MB
	Size uint64 `json:"maxSegmentSize"`

	// Specifies where segment files are stored.
	//
	// Default: "/var/lib/ignitedb/segments"
	Directory string `json:"directory"`

	// Defines the filename prefix for segment files.
	// Final filename will be: `prefix_segmentId_timestamp.seg`
	//
	// Default: "segment"
	//
	// Example: If Prefix is "mydata", a segment file might be "mydata_000001_20240525232100.seg".
	Prefix string `json:"prefix"`
}

// Defines the configuration parameters for Ignite DB.
// It provides control over storage, performance and maintenance aspects.
type Options struct {
	// Specifies the base path where files will be stored.
	//
	// Default: "/var/lib/ignitedb"
	DataDir string `json:"dataDir"`

	// Defines how often the compaction process runs to
	// merge old segments. More frequent compaction means more
	// optimal storage but higher overhead.
	//
	// Default: 5h
	CompactInterval time.Duration `json:"compactInterval"`

	// Configures segment management including size limits and naming convention.
	SegmentOptions *segmentOptions `json:"segmentOptions"`
}

// OptionFunc is a function type that modifies the Ignite system's configuration.
type OptionFunc func(*Options)

// Applies a predefined set of default configuration values to the Options struct.
func WithDefaultOptions() OptionFunc {
	return func(o *Options) {
		opts := NewDefaultOptions()
		o.DataDir = opts.DataDir
		o.SegmentOptions = opts.SegmentOptions
		o.CompactInterval = opts.CompactInterval
	}
}

// Sets the primary data directory for Ignite.
func WithDataDir(directory string) OptionFunc {
	return func(o *Options) {
		directory = strings.TrimSpace(directory)
		if directory != "" {
			o.DataDir = directory
		}
	}
}

// Sets the interval at which Ignite performs compaction operations.
func WithCompactInterval(interval time.Duration) OptionFunc {
	return func(o *Options) {
		if interval > DefaultCompactInterval {
			o.CompactInterval = interval
		}
	}
}

// Sets the directory specifically for storing segment files.
func WithSegmentDir(directory string) OptionFunc {
	return func(o *Options) {
		directory = strings.TrimSpace(directory)
		if directory != "" {
			o.SegmentOptions.Directory = directory
		}
	}
}

// Sets the file name prefix for segment files.
func WithSegmentPrefix(prefix string) OptionFunc {
	return func(o *Options) {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" {
			o.SegmentOptions.Prefix = prefix
		}
	}
}

// Sets the maximum size of individual segment files.
func WithSegmentSize(size uint64) OptionFunc {
	return func(o *Options) {
		if size > MinSegmentSize && size < MaxSegmentSize {
			o.SegmentOptions.Size = size
		}
	}
}
