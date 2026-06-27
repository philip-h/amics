package httpe

import (
	"errors"
	"log/slog"
	"net/http"
)

type serverError struct {
	error
	status int
}

func (se *serverError) Unwrap() error   { return se.error }
func (se *serverError) HttpStatus() int { return se.status }

func ServerError(err error, status int) *serverError {
	return &serverError{
		error:  err,
		status: status,
	}
}

func HandlerFunc(hf func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hf(w, r); err != nil {
			slog.Error("Big oops",
				slog.String("err", err.Error()),
				slog.Group("where",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("host", r.Host)),
			)
			se := &serverError{}
			if errors.As(err, &se) {
				http.Error(w, se.Error(), se.status)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
