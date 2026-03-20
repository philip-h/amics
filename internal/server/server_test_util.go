package server

import (
	"html/template"
	"net/http"
	"testing"
	"time"

	"github.com/philip-h/amics/internal/auth"
	"github.com/philip-h/amics/internal/store"
	"github.com/philip-h/amics/templates"
)

func newTestApplication(t *testing.T, cfg Config) *Application {
	t.Helper()

	mockStore := store.NewMockStore()
	mockAuth := auth.NewMockAuthenticator()
	templates, _ := template.ParseFS(templates.TemplateFS, "pages/*.gohtml", "partials/*.gohtml", "admin/*.gohtml")

	return &Application{
		Config:    cfg,
		Store:     mockStore,
		Auth:      mockAuth,
		Templates: templates,
	}
}

func checkStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status code %d, want %d", got, want)
	}
}

func createLegitStudentCookie(t *testing.T, app *Application) *http.Cookie {
	t.Helper()
	token, _ := app.Auth.CreateJwt("42", "student", time.Now().Add(time.Minute))
	return &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	}
}

func createLegitTeacherCookie(t *testing.T, app *Application) *http.Cookie {
	t.Helper()
	token, _ := app.Auth.CreateJwt("72", "teacher", time.Now().Add(time.Minute))
	return &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	}
}

func createSillyCookie(t *testing.T) *http.Cookie {
	t.Helper()
	return &http.Cookie{
		Name:     "dummy",
		Value:    "dummy-value",
		HttpOnly: true,
	}
}
