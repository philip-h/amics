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
func loadTemplates() (map[string]*template.Template, error) {
	pages := []string{"home", "hello"}

	cache := map[string]*template.Template{}

	for _, page := range pages {
		tmpl, err := template.New("").
			Funcs(template.FuncMap{
				"unixToDate": func(unix int64) string {
					return time.Unix(unix, 0).Format("Mon Jan 2 @ 15:04")
				},
			}).
			ParseFS(
				templates.TemplateFS,
				"layouts/*.html",
				"partials/*.html",
				"pages/"+page+".html",
			)

		if err != nil {
			return nil, err
		}

		cache[page] = tmpl
	}

	return cache, nil
}

func newTestApplication(t *testing.T, cfg Config) *Application {
	t.Helper()

	mockStore := store.NewMockStore()
	mockAuth := auth.NewMockAuthenticator()
	templates, _ := loadTemplates()

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
