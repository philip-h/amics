package server

import "net/http"

// ============================================================================
// Json Error
// ============================================================================
type JsonError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Internal string `json:"-"`
}

func (e *JsonError) Error() string {
	return e.Message
}

func NewJsonError(status int, message string, internal string) *JsonError {
	return &JsonError{
		Status:  status,
		Message: message,
		Internal: internal,
	}
}

// ============================================================================
// Http Error
// ============================================================================
type ServerError struct {
	Status int
	Internal string
}

func (e *ServerError) Error() string {
	return http.StatusText(e.Status)
}

func NewServerError(status int, internal string) *ServerError {
	return &ServerError{
		Status: status,
		Internal: internal,
	}
}	