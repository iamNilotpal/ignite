package errors

// StorageError is a specialized error type for storage-related operations.
// It embeds baseError to inherit all the standard error functionality, then adds
// storage-specific fields that help pinpoint exactly where problems occurred.
type StorageError struct {
	// Embed the base error to inherit all standard error functionality
	// including error chaining, structured details, and error codes.
	*baseError

	// segmentId identifies which segment was being accessed when the error occurred.
	// This field helps correlate storage errors with specific files on disk,
	// enabling targeted debugging and recovery operations.
	segmentId int

	// offset specifies the byte position within the segment where the problem happened.
	// Combined with segmentId, this provides the exact location of storage issues,
	// making it possible to pinpoint corruption or access problems precisely.
	offset int

	// fileName contains the name of the file that caused the issue.
	// This provides immediate visibility into which storage file was involved,
	// helping operators quickly identify problematic files during troubleshooting.
	fileName string

	// path contains the full filesystem path of the file that caused the issue.
	// This complements fileName by providing the complete location information
	// needed for file system operations during error recovery.
	path string
}

// NewStorageError creates a new storage-specific error with the provided context.
// This constructor follows the established pattern for error creation, taking
// a causing error, error code, and descriptive message as the foundation.
func NewStorageError(err error, code ErrorCode, msg string) *StorageError {
	return &StorageError{baseError: NewBaseError(err, code, msg)}
}

// Override base error methods to return *StorageError instead of *baseError.

// WithMessage updates the error message while maintaining the StorageError type.
func (se *StorageError) WithMessage(msg string) *StorageError {
	se.baseError.WithMessage(msg)
	return se
}

// WithCode sets the error code while preserving the StorageError type.
func (se *StorageError) WithCode(code ErrorCode) *StorageError {
	se.baseError.WithCode(code)
	return se
}

// WithDetail adds contextual information while maintaining the StorageError type.
func (se *StorageError) WithDetail(key string, value any) *StorageError {
	se.baseError.WithDetail(key, value)
	return se
}

// Storage-specific methods that add domain-specific context to the error.
// These methods follow the fluent interface pattern, enabling readable
// error construction through method chaining.

// WithSegmentID sets which storage segment was involved in the error.
func (se *StorageError) WithSegmentID(id int) *StorageError {
	se.segmentId = id
	return se
}

// WithOffset records the byte position where the error occurred.
func (se *StorageError) WithOffset(offset int) *StorageError {
	se.offset = offset
	return se
}

// WithFileName captures which file was being processed when the error occurred.
func (se *StorageError) WithFileName(fileName string) *StorageError {
	se.fileName = fileName
	return se
}

// WithPath captures which filesystem path was being processed during the error.
func (se *StorageError) WithPath(path string) *StorageError {
	se.path = path
	return se
}

// Getter methods provide access to the StorageError-specific context.
// These methods allow error handling code to make decisions based on
// the specific storage context captured when the error was created.

// SegmentId returns the segment identifier where the error occurred.
func (se *StorageError) SegmentId() int {
	return se.segmentId
}

// Offset returns the byte offset within the segment where the error happened.
// Combined with SegmentId, this gives you the exact location of the problem,
// enabling precise debugging and targeted recovery operations.
func (se *StorageError) Offset() int {
	return se.offset
}

// FileName returns the name of the file that was being processed.
func (se *StorageError) FileName() string {
	return se.fileName
}

// Path returns the full filesystem path of the file that was being processed.
func (se *StorageError) Path() string {
	return se.path
}

// Helper functions for creating common storage errors with appropriate context.
// These convenience functions encapsulate the knowledge about what context
// should be captured for specific storage error scenarios, making error
// creation more consistent and reducing the cognitive load on developers.

// NewSegmentCorruptionError creates a specialized error for corrupted segment files.
// This constructor automatically includes the segment information and sets up
// appropriate error codes and messages for corruption scenarios.
func NewSegmentCorruptionError(segmentId int, offset int, cause error) *StorageError {
	return NewStorageError(cause, ErrorCodeSegmentCorrupted, "segment file corrupted").
		WithSegmentID(segmentId).
		WithOffset(offset).
		WithDetail("corruption_type", "checksum_mismatch").
		WithDetail("recovery_required", true)
}

// NewHeaderReadError creates an error for header reading failures.
// This constructor provides a specialized error for the common case where
// segment headers cannot be read, including appropriate context for debugging.
func NewHeaderReadError(fileName string, offset int, cause error) *StorageError {
	return NewStorageError(cause, ErrorCodeHeaderReadFailure, "failed to read segment header").
		WithFileName(fileName).
		WithOffset(offset).
		WithDetail("header_size_expected", 32).
		WithDetail("operation", "header_read")
}

// NewPayloadReadError creates an error for payload reading failures.
// This specialized constructor handles the case where segment data payloads
// cannot be read successfully, providing context about the read operation.
func NewPayloadReadError(fileName string, segmentId int, offset int, expectedSize int, cause error) *StorageError {
	return NewStorageError(cause, ErrorCodePayloadReadFailure, "failed to read segment payload").
		WithFileName(fileName).
		WithSegmentID(segmentId).
		WithOffset(offset).
		WithDetail("expected_payload_size", expectedSize).
		WithDetail("operation", "payload_read")
}

// NewFileAccessError creates an error for file system access problems.
// This general-purpose constructor handles various file access issues
// while providing the specific file context needed for debugging.
func NewFileAccessError(path string, fileName string, operation string, cause error) *StorageError {
	return NewStorageError(cause, ErrorCodeIO, "file access failed").
		WithPath(path).
		WithFileName(fileName).
		WithDetail("operation", operation)
}
