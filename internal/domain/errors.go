package domain

import "fmt"

// Error codes for the notification domain.
const (
	ErrCodeInvalidChannel  = "INVALID_CHANNEL"
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeChannelDisabled = "CHANNEL_DISABLED"
	ErrCodeConfigMissing   = "CONFIG_MISSING"
	ErrCodeSendFailed      = "SEND_FAILED"
	ErrCodeNetworkError    = "NETWORK_ERROR"
	ErrCodeAuthentication  = "AUTH_ERROR"
	ErrCodeRateLimit       = "RATE_LIMIT"
)

// Error represents a domain-level error with a code and message.
type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// NewDomainError creates a new domain Error.
func NewDomainError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewDomainErrorWithCause creates a new domain Error with an underlying cause.
func NewDomainErrorWithCause(code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
