package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/philip-h/amics/internal/httpe"
)

func authCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "amics-cookie",
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func logoutCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "amics-cookie",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func intPathValue(r *http.Request, name string) (int, error) {
	value := r.PathValue(name)
	if value == "" {
		return 0, httpe.ServerError(errors.New("path value is empty"), http.StatusBadRequest)
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, httpe.ServerError(err, http.StatusBadRequest)
	}
	return intValue, nil
}
