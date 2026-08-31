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
