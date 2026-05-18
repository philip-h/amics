package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/philip-h/amics/internal/storage"
)

const (
	personKey = "person_id"
	roleKey = "role"
) 

func requiresStudent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roleFromContext := r.Context().Value(roleKey)
		if roleFromContext == nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}

		role := roleFromContext.(string)
		if role != "student" && role != "teacher" {
			http.Error(w, "Not allowed", http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func requiresTeacher(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roleFromContext := r.Context().Value(roleKey)
		if roleFromContext == nil {
			http.Error(w, "Not allowed", http.StatusForbidden)
			return
		}

		role := roleFromContext.(string)
		if role != "teacher" {
			http.Error(w, "Not allowed", http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func checkAuthMiddlewear(logger *Logger, store *storage.Storage, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err == http.ErrNoCookie {
			h.ServeHTTP(w, r)
			return
		}

		session, err := store.Sessions.GetById(cookie.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				HttpOnly: true,
				MaxAge:   -1,
			})
			h.ServeHTTP(w, r)
			return
		}
		if session == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				HttpOnly: true,
				MaxAge:   -1,
			})
			h.ServeHTTP(w, r)
			return
		}

		person, err := store.People.GetById(session.PersonId)
		if err != nil {
			logger.L.Error("Could not get person by id", "err", err)
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				HttpOnly: true,
				MaxAge:   -1,
			})
			h.ServeHTTP(w, r)
			return
		}
		if person == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				HttpOnly: true,
				MaxAge:   -1,
			})
			h.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), personKey, person.Id)
		ctx = context.WithValue(ctx, roleKey, person.Role)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func panicRecovery(logger *Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					logger.L.Error(
						"panic recovered",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						"panic",
						p,
					)
					http.Error(w, "Big oopsies", http.StatusInternalServerError)
				}
			}()
			handler.ServeHTTP(w, r)
		})
}
