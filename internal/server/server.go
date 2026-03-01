package server

import (
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
	mux.HandleFunc("POST /teacher/courses/new", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleCourseCreate)))
	mux.HandleFunc("POST /teacher/courses/{courseId}", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleCourseUpdate)))

	mux.HandleFunc("GET /teacher/courses/{courseId}", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleTeacherCourses)))

	mux.HandleFunc("GET /teacher/courses/{courseId}/assignments", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleTeacherAssignments)))
	mux.HandleFunc("GET /teacher/courses/{courseId}/assignments/{assignmentId}", app.withAuth("teacher", makeHTTPHandlerFunc((app.handleTeacherAssignmentDetail))))
	mux.HandleFunc("POST /teacher/courses/{courseId}/assignments/new", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleTeacherAssignmentCreate)))
	mux.HandleFunc("POST /teacher/courses/{courseId}/assignments/{assignmentId}", app.withAuth("teacher", makeHTTPHandlerFunc(app.handleTeacherAssignmentUpdate)))

	return mux
}

func (app *Application) Run(mux *http.ServeMux) error {
	server := &http.Server{
		Addr:    app.Config.Port,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func makeHTTPHandlerFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			se := &errs.ServerError{}
			jwte := &errs.JwtError{}
			if errors.As(err, &se) {
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
