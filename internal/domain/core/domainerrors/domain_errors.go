package domainerrors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeBusinessRule  ErrorCode = "BUSINESS_RULE_VIOLATION"
)

type DomainError struct {
	Code       ErrorCode              // Error code for programmatic handling
	Message    string                 // Human-readable message
	Details    map[string]interface{} // Additional context
	StatusCode int                    // HTTP status code mapping
	Err        error                  // Wrapped error (optional)
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is allows error comparison
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithDetail adds additional context to the error
func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithError wraps another error
func (e *DomainError) WithError(err error) *DomainError {
	e.Err = err
	return e
}

// Constructor functions for common errors
func NewNotFoundError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

func NewAlreadyExistsError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeAlreadyExists,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewInvalidInputError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeInvalidInput,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewValidationError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewBusinessRuleError(message string) *DomainError {
	return &DomainError{
		Code:       ErrCodeBusinessRule,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Code:       ErrCodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func As(err error) (*DomainError, bool) {
	if err == nil {
		return nil, false
	}

	domainErr, ok := err.(*DomainError)
	if !ok {
		return nil, false
	}

	return domainErr, true
}
