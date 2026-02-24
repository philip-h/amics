package server

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/philip-h/amics/internal/auth"
	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/errs"
	"github.com/philip-h/amics/internal/store"
)

type Application struct {
	Config    Config
	Store     store.Storage
	Auth      auth.Authenticator
	Templates *template.Template
}

type Config struct {
	Port string
	Db   *db.DbConfig
}

func (app *Application) Mount() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Homepage
	mux.HandleFunc("GET /", makeHTTPHandlerFunc(app.handleIndex))

	// Auth handlers
	mux.HandleFunc("GET /login", makeHTTPHandlerFunc(app.handleLoginGet))
	mux.HandleFunc("POST /login", makeHTTPHandlerFunc(app.handleLoginPost))
	mux.HandleFunc("GET /register", makeHTTPHandlerFunc(app.handleRegisterGet))
	mux.HandleFunc("POST /register", makeHTTPHandlerFunc(app.handleRegisterPost))
	mux.HandleFunc("POST /logout", makeHTTPHandlerFunc(app.handleLogout))

	// App Routes
	mux.HandleFunc("GET /app", app.withAuth("student", makeHTTPHandlerFunc(app.handleDashboard)))
	mux.HandleFunc("GET /app/assignments/{assignmentId}", app.withAuth("student", makeHTTPHandlerFunc(app.handleAssignmentDetail)))
	mux.HandleFunc("POST /app/assignments/{assignmentId}/submit", app.withAuth("student", makeHTTPHandlerFunc(app.handleAssignmentSubmit)))

	// Admin Routes
	mux.HandleFunc("GET /teacher", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleTeacherDashboard)))
	mux.HandleFunc("GET /teacher/courses/{courseId}/assignments", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleTeacherAssignments)))
	mux.HandleFunc("GET /teacher/assignments/{assignmentId}", app.withAuth("teacher", makeHTTPHandlerFunc((app.handleTeacherAssignmentDetail))))

	return mux
}

func (app *Application) Run(mux *http.ServeMux) error {
	server := &http.Server{
		Addr:    app.Config.Port,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func writeJSON(w http.ResponseWriter, status int, resource any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&resource)
}

func makeHTTPHandlerFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			je := &errs.JsonError{}
			se := &errs.ServerError{}
			jwte := &errs.JwtError{}
			if errors.As(err, &je) {
				log.Printf("%s: %s", r.URL.Path, je.Internal)
				writeJSON(w, je.Status, je)
			} else if errors.As(err, &se) {
				log.Printf("%s: %s", r.URL.Path, se.Internal)
				http.Error(w, se.Error(), se.Status)
			} else if errors.Is(err, http.ErrNoCookie) {
				log.Printf("%s: No auth token provided", r.URL.Path)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.As(err, &jwte) {
				log.Printf("%s: %s", r.URL.Path, jwte.Message)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.Is(err, &errs.UnauthorizedError{}) {
				log.Printf("%s: Unauthorized access attempt", r.URL.Path)
				http.NotFound(w, r)
			} else {
				log.Printf("[?] %s: %s!!", r.URL.Path, err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	}
}
