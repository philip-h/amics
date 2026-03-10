package server

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/philip-h/amics/internal/errs"
)

func (app *Application) withAuth(role string, next http.HandlerFunc) http.HandlerFunc {
	return makeHTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Check for token cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			if role == "teacher" {
				return &errs.UnauthorizedError{}
			}
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
				
				if role == "teacher" {
					return &errs.UnauthorizedError{}
				}
				return &errs.JwtError{
					Message: "JWT token is either expired or not active yet",
				}
			}
			return err
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			if role == "teacher" {
				return &errs.UnauthorizedError{}
			}
			return &errs.JwtError{
				Message: "Failed to extract claims from JWT",
			}
		}

		userId, err := claims.GetSubject()
		if err != nil {
			if role == "teacher" {
				return &errs.UnauthorizedError{}
			}
			return &errs.JwtError{
				Message: "Failed to extract username from JWT claims",
			}
		}

		aud, err := claims.GetAudience()
		if err != nil {
			if role == "teacher" {
				return &errs.UnauthorizedError{}
			}
			return &errs.JwtError{
				Message: "Failed to extract role from JWT claims",
			}
		}

		// A teacher can access both teacher and student routes, but a student can only access student routes
		allowedRoles := map[string][]string{"teacher": {"teacher", "student"}, "student": {"student"}}
		if allowed, ok := allowedRoles[aud[0]]; !ok || !slices.Contains(allowed, role) {
			if role == "teacher" {
				return &errs.UnauthorizedError{}
			}
			return &errs.JwtError{
				Message: "Role does not match expected role",
			}
		}

		// Set the user id in request context for later
		ctx := context.WithValue(r.Context(), "userId", userId)
    ctx = context.WithValue(ctx, "is-teacher", aud[0] == "teacher")
		next(w, r.WithContext(ctx))
		return nil
	})
}
