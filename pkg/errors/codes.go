package errors

// ErrorCode represents a typed application error code.
type ErrorCode string

const (
	// Generic
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists  ErrorCode = "ALREADY_EXISTS"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
	ErrCodeConflict       ErrorCode = "CONFLICT"

	// Auth
	ErrCodeInvalidToken   ErrorCode = "INVALID_TOKEN"
	ErrCodeExpiredToken   ErrorCode = "EXPIRED_TOKEN"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// User
	ErrCodeUserNotFound   ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserDisabled   ErrorCode = "USER_DISABLED"
	ErrCodeEmailTaken     ErrorCode = "EMAIL_TAKEN"

	// Database
	ErrCodeDBQuery        ErrorCode = "DB_QUERY_ERROR"
	ErrCodeDBConnection   ErrorCode = "DB_CONNECTION_ERROR"

	// Kafka
	ErrCodeKafkaPublish   ErrorCode = "KAFKA_PUBLISH_ERROR"
	ErrCodeKafkaConsume   ErrorCode = "KAFKA_CONSUME_ERROR"
)

// HTTPStatus returns the HTTP status code associated with an error code.
func (c ErrorCode) HTTPStatus() int {
	switch c {
	case ErrCodeNotFound, ErrCodeUserNotFound:
		return 404
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeExpiredToken, ErrCodeInvalidCredentials:
		return 401
	case ErrCodeForbidden:
		return 403
	case ErrCodeValidation, ErrCodeBadRequest:
		return 400
	case ErrCodeAlreadyExists, ErrCodeEmailTaken, ErrCodeConflict:
		return 409
	case ErrCodeTooManyRequests:
		return 429
	default:
		return 500
	}
}
