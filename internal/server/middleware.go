package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/philip-h/amics/internal/errs"
)

func (app *Application) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return makeHTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Check for token cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			return err
		}

		// Validate token
		token, err := app.Auth.ValidateJwt(cookie.Value)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {

				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    "",
					HttpOnly: true,
					MaxAge:   -1,
				})
				return &errs.InvalidJwtError{
					Message: "JWT token is either expired or not active yet",
				}
			}
			return err
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return &errs.InvalidJwtError{
				Message: "Failed to extract claims from JWT",
			}
		}

		username, err := claims.GetSubject()
		if err != nil {
			return &errs.InvalidJwtError{
				Message: "Failed to extract username from JWT claims",
			}
		}

		// Set the username in request context for later
		ctx := context.WithValue(r.Context(), "username", username)
		next(w, r.WithContext(ctx))
		return nil
	})
}

