package errors

// baseError is a custom error type that can hold extra information.
// This struct follows the error wrapping pattern, allowing us to chain errors
// while preserving context and adding structured information for debugging.
type baseError struct {
	cause   error          // The original error that caused this one.
	message string         // The error message that will be displayed to users.
	code    ErrorCode      // Error code for categorizing the error type programmatically.
	details map[string]any // Additional context information like request IDs, timestamps, etc.
}

// NewBaseError creates a new BaseError with the given underlying error and message.
func NewBaseError(err error, code ErrorCode, msg string) *baseError {
	return &baseError{cause: err, code: code, message: msg}
}

// WithMessage updates the error message. This allows you to customize the message
// after creation, which is useful when building errors in multiple steps.
func (be *baseError) WithMessage(msg string) *baseError {
	be.message = msg
	return be
}

// WithCode sets the error code for this error. Error codes help your application
// handle different error types programmatically instead of parsing error messages.
func (be *baseError) WithCode(code ErrorCode) *baseError {
	be.code = code
	return be
}

// WithDetail adds contextual information to help with debugging and logging.
// The details map is lazily initialized to avoid allocating memory when not needed.
// Common details might include user IDs, request IDs, file paths, or operation parameters.
func (be *baseError) WithDetail(key string, value any) *baseError {
	if be.details == nil {
		be.details = make(map[string]any)
	}
	be.details[key] = value
	return be
}

// Error returns the error message, implementing Go's built-in error interface.
// This is what gets displayed when you print the error or convert it to a string.
func (b *baseError) Error() string {
	return b.message
}

// Unwrap returns the underlying error, enabling Go's error unwrapping functionality.
// This allows functions like errors.Is() and errors.As() to work with wrapped errors,
// making it possible to check for specific error types in a chain.
func (b *baseError) Unwrap() error {
	return b.cause
}

// Code returns the error code, which allows your application to handle
// different types of errors programmatically. This is more reliable than
// parsing error messages and enables better error handling strategies.
func (b *baseError) Code() ErrorCode {
	return b.code
}

// Details returns the additional context information stored with this error.
// This returns a reference to the internal map, so be careful about modifications.
// The details are particularly useful for structured logging and debugging.
func (b *baseError) Details() map[string]any {
	return b.details
}
