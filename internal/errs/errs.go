package errs

import "net/http"

// ============================================================================
// Http Error
// ============================================================================
type ServerError struct {
	Status   int
	Internal string
}

func (e *ServerError) Error() string {
	return http.StatusText(e.Status)
}

func NewServerError(status int, internal string) *ServerError {
	return &ServerError{
		Status:   status,
		Internal: internal,
	}
}

// ============================================================================
// Invalid JWT Error
// ============================================================================
type JwtError struct {
	Message string
}

func (e *JwtError) Error() string {
	return e.Message
}

func NewInvalidJwtError(message string) *JwtError {
	return &JwtError{
		Message: message,
	}
}

// ============================================================================
// Unauthorized Error
// ============================================================================
type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "Unauthorized access to teacher resource"
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{}
}
