package errors

// IndexError provides specialized error handling for index-related operations.
// This structure extends the base error system with index-specific context
// while properly supporting method chaining through all base error methods.
type IndexError struct {
	// Embed the base error to inherit all standard error functionality
	// including error chaining, structured details, and error codes.
	*baseError

	// Identifies which key was being processed when the error occurred.
	// This is particularly valuable for debugging because it tells you exactly
	// which piece of data was involved in the failed operation.
	key string

	// Indicates which segment was involved in the error, if applicable.
	// This helps correlate index errors with specific segment files and can
	// guide recovery operations or compaction decisions.
	segmentID uint16

	// Describes what index operation was being performed when the
	// error occurred (e.g., "Get", "Put", "Delete", "Recovery"). This context
	// helps understand the system state and user actions that led to the error.
	operation string

	// Captures the size of the index at the time of the error.
	// This information helps diagnose capacity-related issues and provides
	// context about the scale of the system when problems occur.
	indexSize int

	// Estimates how much memory the index was consuming when
	// the error occurred. This helps diagnose memory-related issues and
	// provides context for capacity planning decisions.
	memoryUsage int64
}

// NewIndexError creates a new index-specific error with the provided context.
// This constructor follows the same pattern as other error types in the system,
// taking a causing error, error code, and descriptive message.
func NewIndexError(err error, code ErrorCode, msg string) *IndexError {
	return &IndexError{
		baseError: NewBaseError(err, code, msg),
	}
}

// Override base error methods to return *IndexError instead of *baseError.

// WithMessage updates the error message while maintaining the IndexError type.
func (ie *IndexError) WithMessage(msg string) *IndexError {
	ie.baseError.WithMessage(msg)
	return ie
}

// WithCode sets the error code while preserving the IndexError type.
func (ie *IndexError) WithCode(code ErrorCode) *IndexError {
	ie.baseError.WithCode(code)
	return ie
}

// WithDetail adds contextual information while maintaining the IndexError type.
func (ie *IndexError) WithDetail(key string, value any) *IndexError {
	ie.baseError.WithDetail(key, value)
	return ie
}

// Index-specific methods that add domain-specific context to the error.
// These methods enable comprehensive error reporting for index operations
// while maintaining the fluent interface pattern for readable error construction.

// WithKey records which key was being processed when the error occurred.
// This information proves invaluable for debugging because it enables
// reproduction of the error by attempting the same operation on the same key.
func (ie *IndexError) WithKey(key string) *IndexError {
	ie.key = key
	return ie
}

// WithSegmentID captures which segment was involved in the error.
// This information provides a direct link between index errors and
// the underlying storage system, facilitating cross-layer debugging.
func (ie *IndexError) WithSegmentID(segmentID uint16) *IndexError {
	ie.segmentID = segmentID
	return ie
}

// WithOperation records what index operation was being performed.
// This context helps understand the system state and operation sequence
// that led to the error condition, enabling more effective debugging.
func (ie *IndexError) WithOperation(operation string) *IndexError {
	ie.operation = operation
	return ie
}

// WithIndexSize captures the size of the index when the error occurred.
// This information helps diagnose capacity-related issues and provides
// context about system scale when problems arise.
func (ie *IndexError) WithIndexSize(size int) *IndexError {
	ie.indexSize = size
	return ie
}

// WithMemoryUsage records the estimated memory consumption of the index.
// This provides crucial context for diagnosing memory-related issues and
// understanding resource utilization when errors occur.
func (ie *IndexError) WithMemoryUsage(usage int64) *IndexError {
	ie.memoryUsage = usage
	return ie
}

// Getter methods provide access to the IndexError-specific context.
// These methods enable error handling code to make informed decisions
// based on the specific context captured during error creation.

// Key returns the key that was being processed when the error occurred.
func (ie *IndexError) Key() string {
	return ie.key
}

// SegmentID returns the segment identifier associated with the error.
func (ie *IndexError) SegmentID() uint16 {
	return ie.segmentID
}

// Operation returns the name of the operation that was being performed.
func (ie *IndexError) Operation() string {
	return ie.operation
}

// IndexSize returns the size of the index when the error occurred.
func (ie *IndexError) IndexSize() int {
	return ie.indexSize
}

// MemoryUsage returns the estimated memory consumption when the error occurred.
func (ie *IndexError) MemoryUsage() int64 {
	return ie.memoryUsage
}

// Helper functions for creating common index errors with appropriate context.
// These convenience functions encapsulate best practices for index error
// creation while reducing the cognitive burden on developers using the system.

// NewKeyNotFoundError creates a specialized error for missing keys.
// This constructor demonstrates how the fixed method chaining enables
// seamless mixing of base methods and index-specific methods.
func NewKeyNotFoundError(key string) *IndexError {
	return NewIndexError(nil, ErrorCodeIndexKeyNotFound, "key not found in index").
		WithKey(key).
		WithOperation("Get").
		WithDetail("lookup_time", "immediate"). // Base method works seamlessly
		WithDetail("cache_checked", true)
}

// NewSegmentIDError creates an error for invalid segment ID conditions.
// This constructor demonstrates building comprehensive error context
// using both domain-specific and general contextual information.
func NewSegmentIDError(segmentID uint16, key string) *IndexError {
	return NewIndexError(nil, ErrorCodeIndexInvalidSegmentID, "segment ID not found").
		WithSegmentID(segmentID).
		WithKey(key).
		WithOperation("Get").
		WithDetail("segment_file_exists", false).
		WithDetail("index_consistency_check", "failed")
}

// NewTimestampExtractionError creates an error for filename parsing failures.
// This constructor shows how to properly chain complex error context
// while maintaining type safety throughout the construction process.
func NewTimestampExtractionError(filename string, cause error) *IndexError {
	return NewIndexError(cause, ErrorCodeIndexTimestampExtraction, "failed to extract timestamp from filename").
		WithOperation("TimestampExtraction").
		WithDetail("filename", filename).
		WithDetail("expected_format", "prefix_NNNNN_timestamp.seg").
		WithDetail("parsing_stage", "timestamp_component")
}

// NewIndexCorruptionError creates an error for index corruption scenarios.
// This specialized constructor provides comprehensive context for
// serious index integrity issues that require immediate attention.
func NewIndexCorruptionError(operation string, indexSize int, cause error) *IndexError {
	return NewIndexError(cause, ErrorCodeIndexCorrupted, "index data structure corrupted").
		WithOperation(operation).
		WithIndexSize(indexSize).
		WithDetail("corruption_detected", true).
		WithDetail("recovery_required", true).
		WithDetail("backup_recommended", true)
}
