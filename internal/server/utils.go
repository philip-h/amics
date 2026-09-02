package server

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/httpe"
)

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

type Flash struct {
	Message string
	IsError bool
}

func setFlash(w http.ResponseWriter, message *Flash) error {
	encoded, err := encode(message)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     "amics-flash",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	return nil
}

func getFlash(w http.ResponseWriter, r *http.Request) (*Flash, error) {
	cookie, err := r.Cookie("amics-flash")
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return nil, nil
		default:
			return nil, err
		}
	}

	value, err := decode(cookie.Value)
	if err != nil {
		return nil, err
	}

	// Delete the flash cookie after reading it; Path must match the cookie set in setFlash
	http.SetCookie(w, &http.Cookie{
		Name:     "amics-flash",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return value, nil
}

func encode(src *Flash) (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(src); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

func decode(src string) (*Flash, error) {
	flash := &Flash{}
	data, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(flash); err != nil {
		return nil, err
	}

	return flash, nil
}
