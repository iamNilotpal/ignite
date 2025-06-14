package errors

// ErrorCode represents a standardized way to categorize different types of errors.
type ErrorCode string

// Base error codes represent the fundamental categories of failures that can
// occur across any software system. These codes provide the foundation layer
// of error classification.
const (
	// ErrorCodeIO represents failures in input/output operations across any
	// system boundary. This includes file system operations like reading or
	// writing segment files, network operations when communicating with remote
	// systems, and device I/O when accessing storage hardware.
	ErrorCodeIO ErrorCode = "IO_ERROR"

	// ErrorCodeInvalidInput represents client-side errors where the provided
	// data doesn't meet the system's requirements or constraints. This maps
	// to HTTP 400-series errors and indicates problems with the request itself
	// rather than system failures.
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"

	// ErrorCodeInternal represents unexpected system failures that don't fit
	// into other categories. These are the equivalent of HTTP 500 errors and
	// indicate bugs, assertion failures, or other programming errors that
	// shouldn't occur during normal operation.
	ErrorCodeInternal ErrorCode = "INTERNAL_ERROR"
)

// Storage-specific error codes extend the base error taxonomy to handle the
// unique failure modes that occur in persistent storage systems. These codes
// represent problems that are specific to the storage layer of your key-value
// store, particularly focusing on segment file management and data persistence.
const (
	// ErrorCodeSegmentCorrupted indicates that a segment file's data has been
	// damaged or is in an inconsistent state.
	ErrorCodeSegmentCorrupted ErrorCode = "SEGMENT_CORRUPTED"

	// ErrorCodeHeaderReadFailure occurs when the system cannot read the header
	// portion of a segment file. Headers contain critical metadata about the
	// segment's structure, so header read failures prevent access to the
	// entire segment and all data it contains.
	ErrorCodeHeaderReadFailure ErrorCode = "HEADER_READ_FAILURE"

	// ErrorCodePayloadReadFailure indicates problems reading the actual data
	// content from segment files after successfully reading the header. This
	// represents a more localized failure compared to header problems, as the
	// segment structure is intact but specific data regions are inaccessible.
	ErrorCodePayloadReadFailure ErrorCode = "PAYLOAD_READ_FAILURE"

	// ErrorCodeRecoveryFailed indicates that the storage system's attempt to
	// recover from a previous failure was unsuccessful. This represents a
	// compound failure where both the original problem and the recovery
	// mechanism have failed, creating a more serious operational situation.
	ErrorCodeRecoveryFailed ErrorCode = "STORAGE_RECOVERY_FAILED"

	// ErrorCodePermissionDenied indicates insufficient permissions to access a resource.
	// This is distinct from generic IO errors because it has a specific resolution path:
	// the user needs to adjust file/directory permissions or run with elevated privileges.
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"

	// ErrorCodeDiskFull indicates that the storage device has run out of space.
	// This requires specific handling like cleanup operations or alerting administrators.
	ErrorCodeDiskFull ErrorCode = "DISK_FULL"

	// ErrorCodeFilesystemReadonly indicates that the filesystem is mounted read-only.
	// This requires administrative intervention to remount the filesystem with write permissions.
	ErrorCodeFilesystemReadonly ErrorCode = "FILESYSTEM_READONLY"
)
