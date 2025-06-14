package errors

// ValidationError is a specialized error type for input validation failures.
// It embeds baseError to inherit all the standard error functionality, then adds
// validation-specific fields that help identify exactly what validation rules
// were violated and provide guidance on how to correct the input.
type ValidationError struct {
	// Embed the base error to inherit all standard error functionality
	// including error chaining, structured details, and error codes.
	*baseError

	// Identifies which specific field or parameter failed validation.
	// This allows clients to highlight the problematic field in user interfaces
	// or programmatically correct specific validation issues.
	field string

	// Specifies which validation rule was violated (e.g., "required", "max_length", "format").
	// This provides semantic information about what constraint was not met,
	// enabling clients to show appropriate error messages or apply corrections.
	rule string

	// Captures what value was actually provided that failed validation.
	// This context helps with debugging and allows validation error messages
	// to show users exactly what they provided that was problematic.
	provided any

	// Describes what would have been valid.
	// This provides guidance to users or calling systems about how to fix the input.
	expected any
}

// NewValidationError creates a new validation-specific error with the provided context.
// This constructor follows the established pattern for error creation, taking
// a causing error, error code, and descriptive message as the foundation.
func NewValidationError(err error, code ErrorCode, msg string) *ValidationError {
	return &ValidationError{baseError: NewBaseError(err, code, msg)}
}

// Override base error methods to return *ValidationError instead of *baseError.
// This ensures that method chaining maintains the correct error type throughout
// the validation error construction process.

// WithMessage updates the error message while maintaining the ValidationError type.
func (ve *ValidationError) WithMessage(msg string) *ValidationError {
	ve.baseError.WithMessage(msg)
	return ve
}

// WithCode sets the error code while preserving the ValidationError type.
func (ve *ValidationError) WithCode(code ErrorCode) *ValidationError {
	ve.baseError.WithCode(code)
	return ve
}

// WithDetail adds contextual information while maintaining the ValidationError type.
func (ve *ValidationError) WithDetail(key string, value any) *ValidationError {
	ve.baseError.WithDetail(key, value)
	return ve
}

// Validation-specific methods that add domain-specific context to the error.
// These methods follow the fluent interface pattern, enabling readable
// error construction through method chaining.

// WithField sets which field failed validation.
func (ve *ValidationError) WithField(field string) *ValidationError {
	ve.field = field
	return ve
}

// WithRule specifies which validation rule was violated.
func (ve *ValidationError) WithRule(rule string) *ValidationError {
	ve.rule = rule
	return ve
}

// WithProvided captures what value was provided that failed validation.
func (ve *ValidationError) WithProvided(value any) *ValidationError {
	ve.provided = value
	return ve
}

// WithExpected describes what would have been a valid value.
func (ve *ValidationError) WithExpected(value any) *ValidationError {
	ve.expected = value
	return ve
}

// Getter methods provide access to the ValidationError-specific context.
// These methods allow error handling code to make decisions based on
// the specific validation context captured when the error was created.

// Field returns the field name that failed validation.
func (ve *ValidationError) Field() string {
	return ve.field
}

// Rule returns the validation rule that was violated.
func (ve *ValidationError) Rule() string {
	return ve.rule
}

// Provided returns the value that was provided and failed validation.
func (ve *ValidationError) Provided() any {
	return ve.provided
}

// Expected returns what would have been a valid value.
func (ve *ValidationError) Expected() any {
	return ve.expected
}

// Helper functions for creating common validation errors with appropriate context.
// These convenience functions encapsulate the knowledge about what context
// should be captured for specific validation error scenarios.

// NewRequiredFieldError creates a specialized error for missing required fields.
func NewRequiredFieldError(fieldName string) *ValidationError {
	return NewValidationError(
		nil,
		ErrorCodeInvalidInput,
		"Required field is missing or empty",
	).WithField(fieldName).WithRule("required")
}

// NewFieldFormatError creates an error for fields that don't match expected format.
func NewFieldFormatError(fieldName string, provided any, expected string) *ValidationError {
	return NewValidationError(
		nil,
		ErrorCodeInvalidInput,
		"Field value does not match expected format",
	).WithField(fieldName).WithRule("format").WithProvided(provided).WithExpected(expected)
}

// NewFieldRangeError creates an error for fields that are outside acceptable ranges.
func NewFieldRangeError(fieldName string, provided any, min, max any) *ValidationError {
	return NewValidationError(
		nil,
		ErrorCodeInvalidInput,
		"Field value is outside acceptable range",
	).WithField(fieldName).
		WithRule("range").
		WithProvided(provided).
		WithDetail("minValue", min).
		WithDetail("maxValue", max)
}

// NewConfigurationValidationError creates an error for invalid configuration objects.
func NewConfigurationValidationError(field string, issue string) *ValidationError {
	return NewValidationError(
		nil,
		ErrorCodeInvalidInput,
		"Configuration validation failed",
	).WithField(field).
		WithRule("configuration_integrity").
		WithDetail("validationIssue", issue)
}
