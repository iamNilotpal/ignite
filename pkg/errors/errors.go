// This package addresses the fundamental challenge that generic error handling presents in complex
// systems: when an error occurs, developers and operators need much more than just "something went wrong."
// They need to understand exactly what failed, why it failed, where it failed, and most importantly,
// what they can do about it. This package transforms error handling from reactive debugging into
// proactive problem resolution.
//
// Architecture and Design Philosophy:
//
// The error system is built around a hierarchical structure that starts with a foundational baseError
// and extends into domain-specific error types. This design provides several key advantages:
// it maintains consistency across all error types while allowing specialized context for different
// domains, enables rich error chaining that preserves the complete failure context, supports
// programmatic error handling through standardized error codes, and facilitates comprehensive
// logging and monitoring through structured error details.
//
// The system recognizes that different parts of a storage application fail in fundamentally different
// ways and require different types of contextual information for effective diagnosis and recovery.
// A validation error needs to know which field failed and what rule was violated. A storage error
// needs to know which file and byte offset were involved. An index error needs to know which key
// and operation were being processed. By capturing this domain-specific context at the point of
// failure, the system enables much more intelligent error handling throughout the application stack.
//
// Error Classification and Codes:
//
// Central to this system is a comprehensive error code taxonomy that provides standardized
// categorization of failures. These codes serve multiple purposes: they enable programmatic
// error handling that doesn't rely on parsing error messages, they provide consistent
// categorization for monitoring and alerting systems, they facilitate error recovery logic
// by identifying specific failure modes, and they support internationalization by separating
// error identification from error presentation.
//
// The error codes are organized into several categories. Base codes cover fundamental failure
// types that can occur in any system: IO_ERROR for input/output failures, INVALID_INPUT for
// client-side validation problems, and INTERNAL_ERROR for unexpected system failures.
// Storage-specific codes handle the unique failure modes of persistent storage: SEGMENT_CORRUPTED
// for data integrity issues, PERMISSION_DENIED for access control problems, DISK_FULL for
// capacity issues, and various read/write failure codes for different types of I/O problems.
// Index-specific codes address the specialized needs of index operations: INDEX_KEY_NOT_FOUND
// for missing keys, INDEX_CORRUPTED for structural integrity issues, and various recovery and
// validation failure codes.
//
// Usage Patterns and Best Practices:
//
// This error handling system is designed to support several key usage patterns that improve
// both developer experience and operational visibility.
//
// For error creation, the package encourages building errors with comprehensive context at
// the point of failure. This means capturing not just what went wrong, but where it went
// wrong, what was being attempted, and what conditions led to the failure. The fluent
// interface pattern makes this context capture both readable and maintainable.
//
// For error handling, the package supports both programmatic error handling (using error
// codes and type detection) and human-readable error reporting (using structured messages
// and details). This dual approach enables both robust automated error recovery and
// effective human troubleshooting.
//
// For error propagation, the package encourages preserving error context as errors flow
// through system layers while adding layer-specific context when appropriate. This creates
// a comprehensive audit trail of what happened during a failure, making root cause analysis
// much more effective.
//
// Operational Benefits:
//
// The structured approach to error handling provides significant operational benefits.
// Monitoring and alerting systems can categorize and group errors based on error codes
// rather than parsing error messages. Log analysis becomes more effective because errors
// include structured context that can be easily indexed and searched. Error recovery
// logic becomes more sophisticated because it can make decisions based on specific error
// types and context rather than generic failure notifications.
//
// The system also improves the development experience by making errors more debuggable
// and providing clear patterns for error creation and handling. Developers can quickly
// understand what went wrong and why, rather than spending time deciphering generic
// error messages or trying to reproduce failure conditions
package errors

import (
	stdErrors "errors"
	"os"
	"syscall"
)

// IsValidationError checks if the given error is a ValidationError or contains one in its error chain.
//
// Example usage:
//
//	if errors.IsValidationError(err) {
//	    // Handle validation-specific error recovery
//	    // Maybe return specific HTTP 400 status codes
//	    // Or highlight specific fields in a user interface
//	}
func IsValidationError(err error) bool {
	var ve *ValidationError
	return stdErrors.As(err, &ve)
}

// IsStorageError determines if an error is related to storage operations, such as file I/O,
// disk space issues, or segment file corruption. Storage errors often require different
// handling strategies than other error types because they may indicate hardware issues,
// capacity problems, or data integrity concerns that need immediate attention.
//
// Example usage:
//
//	if errors.IsStorageError(err) {
//	    storageErr, _ := errors.AsStorageError(err)
//	    switch storageErr.Code() {
//	    case ErrorCodeDiskFull:
//	        triggerCleanupProcedures()
//	    case ErrorCodePermissionDenied:
//	        alertAdministrator(storageErr.Path())
//	    }
//	}
func IsStorageError(err error) bool {
	var se *StorageError
	return stdErrors.As(err, &se)
}

// IsIndexError identifies errors that occurred during index operations such as key lookups,
// index updates, or index recovery procedures. Index errors often provide crucial context
// about which keys were involved and what operations were being performed, which is
// essential for debugging performance issues and data consistency problems.
//
// Example usage:
//
//	if errors.IsIndexError(err) {
//	    indexErr, _ := errors.AsIndexError(err)
//	    if indexErr.Code() == ErrorCodeIndexCorrupted {
//	        scheduleIndexRebuild(indexErr.Key())
//	    }
//	}
func IsIndexError(err error) bool {
	var ie *IndexError
	return stdErrors.As(err, &ie)
}

// AsValidationError safely extracts a ValidationError from an error chain, providing access
// to validation-specific context such as which field failed, what rule was violated, and
// what values were provided versus expected. This extraction is essential for building
// meaningful error responses that help clients understand and correct their input.
//
// The extracted ValidationError provides access to specialized methods like Field(),
// Rule(), Provided(), and Expected(), which contain the detailed context needed for
// sophisticated error handling and user interface feedback.
//
// Example usage:
//
//	if validationErr, ok := errors.AsValidationError(err); ok {
//	    logData := map[string]interface{}{
//	        "field": validationErr.Field(),
//	        "rule": validationErr.Rule(),
//	        "provided": validationErr.Provided(),
//	        "expected": validationErr.Expected(),
//	    }
//	    logger.Error("Validation failed", logData)
//	}
func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	if stdErrors.As(err, &ve) {
		return ve, true
	}
	return nil, false
}

// AsStorageError extracts StorageError context from an error chain, providing access to
// storage-specific information such as segment IDs, file offsets, file names, and paths.
// This context is crucial for implementing storage error recovery procedures and for
// providing detailed information to system administrators and monitoring systems.
//
// The extracted StorageError provides access to methods like SegmentId(), Offset(),
// FileName(), and Path(), which contain the precise location information needed for
// effective storage error handling and recovery.
//
// Example usage:
//
//	if storageErr, ok := errors.AsStorageError(err); ok {
//	    errorContext := map[string]interface{}{
//	        "segmentId": storageErr.SegmentId(),
//	        "offset": storageErr.Offset(),
//	        "fileName": storageErr.FileName(),
//	        "path": storageErr.Path(),
//	        "errorCode": storageErr.Code(),
//	    }
//	    handleStorageFailure(errorContext)
//	}
func AsStorageError(err error) (*StorageError, bool) {
	var se *StorageError
	if stdErrors.As(err, &se) {
		return se, true
	}
	return nil, false
}

// AsIndexError extracts IndexError context, providing access to index-specific information
// such as the key being processed, the operation being performed, segment involvement,
// and index size statistics. This context is essential for diagnosing performance issues,
// planning capacity management, and implementing index recovery procedures.
//
// The extracted IndexError provides access to methods like Key(), Operation(), SegmentID(),
// IndexSize(), and MemoryUsage(), which contain the operational context needed for
// sophisticated index error handling and performance optimization.
//
// Example usage:
//
//	if indexErr, ok := errors.AsIndexError(err); ok {
//	    performanceMetrics := map[string]interface{}{
//	        "key": indexErr.Key(),
//	        "operation": indexErr.Operation(),
//	        "segmentId": indexErr.SegmentID(),
//	        "indexSize": indexErr.IndexSize(),
//	        "memoryUsage": indexErr.MemoryUsage(),
//	    }
//	    analyzeIndexPerformance(performanceMetrics)
//	}
func AsIndexError(err error) (*IndexError, bool) {
	var ie *IndexError
	if stdErrors.As(err, &ie) {
		return ie, true
	}
	return nil, false
}

// GetErrorCode extracts the error code from any error that supports it, or returns
// ErrorCodeInternal for errors that don't have specific codes. This function provides
// a consistent way to categorize errors for monitoring and handling purposes.
//
// Example usage:
//
//	errorCode := errors.GetErrorCode(err)
//	metrics.IncrementErrorCounter(string(errorCode))
//
//	switch errorCode {
//	case errors.ErrorCodeDiskFull:
//	    triggerDiskSpaceAlert()
//	case errors.ErrorCodePermissionDenied:
//	    escalateToAdministrator()
//	}
func GetErrorCode(err error) ErrorCode {
	// Try ValidationError first.
	if ve, ok := AsValidationError(err); ok {
		return ve.Code()
	}

	// Try StorageError next.
	if se, ok := AsStorageError(err); ok {
		return se.Code()
	}

	// Try IndexError.
	if ie, ok := AsIndexError(err); ok {
		return ie.Code()
	}

	// For any other error, return a generic internal error code.
	return ErrorCodeInternal
}

// GetErrorDetails extracts structured details from any error that supports them,
// returning an empty map for errors without details. This function provides consistent
// access to additional error context regardless of the specific error type.
//
// Example usage:
//
//	details := errors.GetErrorDetails(err)
//	if len(details) > 0 {
//	    logger.WithFields(details).Error("Operation failed", "error", err.Error())
//	}
//
//	// Check for specific detail keys
//	if operation, exists := details["operation"]; exists {
//	    handleOperationSpecificError(operation.(string))
//	}
func GetErrorDetails(err error) map[string]any {
	// Try ValidationError first.
	if ve, ok := AsValidationError(err); ok {
		if details := ve.Details(); details != nil {
			return details
		}
	}

	// Try StorageError next.
	if se, ok := AsStorageError(err); ok {
		if details := se.Details(); details != nil {
			return details
		}
	}

	// Try IndexError.
	if ie, ok := AsIndexError(err); ok {
		if details := ie.Details(); details != nil {
			return details
		}
	}

	// Return empty map for errors without details.
	return make(map[string]any)
}

// Analyzes directory creation failures and returns appropriate error
// codes based on the underlying system error. This helps clients
// understand exactly what went wrong and how they might fix it.
func ClassifyDirectoryCreationError(err error, path string) error {
	// Check if this is a permission denied error.
	if os.IsPermission(err) {
		return NewStorageError(
			err, ErrorCodePermissionDenied,
			"Insufficient permissions to create segment directory",
		).WithPath(path).
			WithDetail("operation", "directory_creation").
			WithDetail("required_permission", "write").
			WithDetail("suggestion", "check directory permissions or run with elevated privileges")
	}

	// Check for disk space issues using syscall analysis.
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ENOSPC:
				{
					return NewStorageError(
						err, ErrorCodeDiskFull,
						"Insufficient disk space to create segment directory",
					).WithPath(path).
						WithDetail("operation", "directory_creation").
						WithDetail("suggestion", "free up disk space or choose a different location")
				}
			case syscall.EROFS:
				{
					return NewStorageError(
						err, ErrorCodeFilesystemReadonly,
						"Cannot create directory on read-only filesystem",
					).WithPath(path).
						WithDetail("operation", "directory_creation").
						WithDetail("suggestion", "remount filesystem with write permissions")
				}
			}
		}
	}

	// For any other I/O errors, provide the generic I/O error with context
	return NewStorageError(
		err, ErrorCodeIO, "Failed to create segment directory",
	).WithPath(path).WithDetail("operation", "directory_creation")
}

// ClassifyFileOpenError analyzes file opening failures and returns appropriate
// error codes based on the underlying system error. This provides much more
// specific information than a generic I/O error.
func ClassifyFileOpenError(err error, filePath, fileName string) error {
	// Check if this is a permission denied error.
	if os.IsPermission(err) {
		return NewStorageError(
			err, ErrorCodePermissionDenied,
			"Insufficient permissions to open segment file",
		).WithPath(filePath).
			WithFileName(fileName).
			WithDetail("operation", "file_open").
			WithDetail("required_permission", "read_write").
			WithDetail("suggestion", "check file permissions or run with elevated privileges")
	}

	// Check for disk space issues and other system-level
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ENOSPC:
				{
					return NewStorageError(
						err, ErrorCodeDiskFull,
						"Insufficient disk space to create segment file",
					).WithPath(filePath).
						WithFileName(fileName).
						WithDetail("operation", "file_open").
						WithDetail("suggestion", "free up disk space")
				}
			case syscall.EROFS:
				{
					return NewStorageError(
						err, ErrorCodeFilesystemReadonly,
						"Cannot create file on read-only filesystem",
					).WithPath(filePath).
						WithFileName(fileName).
						WithDetail("operation", "file_open").
						WithDetail("suggestion", "remount filesystem with write permissions")
				}
			}
		}
	}

	// For any other I/O errors during file opening.
	return NewStorageError(err, ErrorCodeIO, "Failed to open segment file").
		WithPath(filePath).
		WithFileName(fileName).
		WithDetail("operation", "file_open").
		WithDetail("flags", []string{"O_CREATE", "O_RDWR", "O_APPEND"})
}

// Analyzes sync operation failures and returns appropriate error codes.
// Sync failures can indicate various underlying issues from
// disk space problems to filesystem corruption.
func ClassifySyncError(err error, fileName, filePath string, offset int) error {
	// Check for specific system errors during sync operations.
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			switch errno {
			case syscall.ENOSPC:
				{
					return NewStorageError(
						err, ErrorCodeDiskFull,
						"Cannot sync file: insufficient disk space",
					).WithFileName(fileName).
						WithPath(filePath).
						WithOffset(offset).
						WithDetail("operation", "file_sync").
						WithDetail("suggestion", "free up disk space before continuing")
				}
			case syscall.EROFS:
				{
					return NewStorageError(
						err, ErrorCodeFilesystemReadonly,
						"Cannot sync file: filesystem is read-only",
					).WithFileName(fileName).
						WithPath(filePath).
						WithOffset(offset).
						WithDetail("operation", "file_sync").
						WithDetail("suggestion", "remount filesystem with write permissions")
				}
			case syscall.EIO:
				{ // I/O error during sync often indicates hardware or corruption issues.
					return NewStorageError(
						err, ErrorCodeIO,
						"I/O error during file sync - possible hardware or corruption issue",
					).WithFileName(fileName).
						WithPath(filePath).
						WithOffset(offset).
						WithDetail("operation", "file_sync").
						WithDetail("severity", "high").
						WithDetail("suggestion", "check filesystem integrity and hardware health")
				}
			}
		}
	}

	// For any other sync errors, provide generic I/O error with context
	return NewStorageError(
		err, ErrorCodeIO, "Failed to sync segment file to disk",
	).WithFileName(fileName).WithPath(filePath).WithOffset(offset).
		WithDetail("operation", "file_sync").
		WithDetail("currentSize", offset)
}
