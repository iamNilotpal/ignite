package errors

// StorageError is a specialized error type for storage-related operations.
// It embeds baseError to inherit all the standard error functionality, then adds
// storage-specific fields that help pinpoint exactly where problems occurred.
type StorageError struct {
	*baseError
	segmentId int    // Which segment was being accessed when the error occurred.
	offset    int    // Byte offset within the segment where the problem happened.
	fileName  string // Name of the file that caused the issue.
	path      string // Path of the file that caused the issue.
}

// NewStorageError creates a new storage-specific error.
func NewStorageError(err error, code ErrorCode, msg string) *StorageError {
	return &StorageError{baseError: NewBaseError(err, code, msg)}
}

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

// WithPath captures which path was being processed when the error occurred.
func (se *StorageError) WithPath(path string) *StorageError {
	se.path = path
	return se
}

// SegmentId returns the segment identifier where the error occurred.
func (se *StorageError) SegmentId() int {
	return se.segmentId
}

// Offset returns the byte offset within the segment where the error happened.
// Combined with SegmentId, this gives you the exact location of the problem.
func (se *StorageError) Offset() int {
	return se.offset
}

// FileName returns the name of the file that was being processed.
func (se *StorageError) FileName() string {
	return se.fileName
}

// Path returns the path of the file that was being processed.
func (se *StorageError) Path() string {
	return se.fileName
}
