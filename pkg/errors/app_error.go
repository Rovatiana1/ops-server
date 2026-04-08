package errors

import "fmt"

// AppError is the standard application error type.
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
	Err     error     `json:"-"` // wrapped internal error, never exposed to client
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// --- Constructors ---

func New(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func WithDetails(code ErrorCode, message string, details any) *AppError {
	return &AppError{Code: code, Message: message, Details: details}
}

// --- Helpers ---

func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// Shortcuts

func NotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

func Unauthorized(msg string) *AppError {
	return New(ErrCodeUnauthorized, msg)
}

func Forbidden(msg string) *AppError {
	return New(ErrCodeForbidden, msg)
}

func BadRequest(msg string) *AppError {
	return New(ErrCodeBadRequest, msg)
}

func Conflict(msg string) *AppError {
	return New(ErrCodeConflict, msg)
}

func Internal(err error) *AppError {
	return Wrap(ErrCodeInternal, "an internal error occurred", err)
}

func ValidationError(details any) *AppError {
	return WithDetails(ErrCodeValidation, "validation failed", details)
}
