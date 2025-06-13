package errors

// ErrorCode represents a standardized way to categorize different types of errors.
type ErrorCode string

const (
	ErrorCodeIO           ErrorCode = "IO_ERROR"       // File system, network, or other I/O operations failed.
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"      // Requested resource doesn't exist (like HTTP 404).
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"  // User provided data that doesn't meet requirements.
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR" // Unexpected internal system error.
)

const (
	ErrorCodeSegmentCorrupted     ErrorCode = "SEGMENT_CORRUPTED"       // Data segment has been damaged or corrupted.
	ErrorCodeHeaderReadFailure    ErrorCode = "HEADER_READ_FAILURE"     // Failed to read the header portion of a file/segment.
	ErrorCodePayloadReadFailure   ErrorCode = "PAYLOAD_READ_FAILURE"    // Failed to read the data portion of a file/segment.
	ErrorCodeSegmentLimitExceeded ErrorCode = "SEGMENT_LIMIT_EXCEEDED"  // Maximum number of segments has been reached.
	ErrorCodeRecoveryFailed       ErrorCode = "STORAGE_RECOVERY_FAILED" // Attempted recovery operation failed.
)
