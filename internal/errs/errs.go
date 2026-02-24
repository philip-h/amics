package errs

import "net/http"

// ============================================================================
// Json Error
// ============================================================================
type JsonError struct {
	Status   int    `json:"status"`
	Message  string `json:"message"`
	Internal string `json:"-"`
}

func (e *JsonError) Error() string {
	return e.Message
}

func NewJsonError(status int, message string, internal string) *JsonError {
	return &JsonError{
		Status:   status,
		Message:  message,
		Internal: internal,
	}
}

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
	return "Unauthorized"
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{}
}
