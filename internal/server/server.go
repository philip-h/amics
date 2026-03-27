package server

import (
	"errors"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	Logger    *slog.Logger
	LogLvl    *slog.LevelVar
}

type Config struct {
	Port string
	Db   *db.DbConfig
}

type Logger struct {
	Log    *slog.Logger
	LovLvl *slog.LevelVar
}

func (app *Application) Mount() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		app.Logger.Error("Could not get cwd", slog.String("msg", err.Error()))
		os.Exit(1)
	}
	fs := http.FileServer(http.Dir(filepath.Join(cwd, "/static")))
	mux.Handle("GET /static/{path...}", http.StripPrefix("/static/", fs))

	// Homepage
	mux.HandleFunc("GET /", app.makeHTTPHandlerFunc(app.handleIndex))

	// Auth handlers
	mux.HandleFunc("GET /login", app.makeHTTPHandlerFunc(app.handleLoginGet))
	mux.HandleFunc("POST /login", app.makeHTTPHandlerFunc(app.handleLoginPost))
	mux.HandleFunc("GET /register", app.makeHTTPHandlerFunc(app.handleRegisterGet))
	mux.HandleFunc("POST /register", app.makeHTTPHandlerFunc(app.handleRegisterPost))
	mux.HandleFunc("POST /logout", app.makeHTTPHandlerFunc(app.handleLogout))

	// App Routes
	mux.HandleFunc("GET /app", app.withAuth("student", app.makeHTTPHandlerFunc(app.handleDashboard)))
	mux.HandleFunc("GET /app/assignments/{assignmentId}", app.withAuth("student", app.makeHTTPHandlerFunc(app.handleAssignmentDetail)))
	mux.HandleFunc("POST /app/assignments/{assignmentId}/submit", app.withAuth("student", app.makeHTTPHandlerFunc(app.handleAssignmentSubmit)))

	// Admin Routes
	mux.HandleFunc("GET /teacher", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleTeacherDashboard)))

	//   Course Routes
	mux.HandleFunc("POST /teacher/courses/new", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleCourseCreate)))
	mux.HandleFunc("POST /teacher/courses/{courseId}", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleCourseUpdate)))
	mux.HandleFunc("GET /teacher/courses/{courseId}", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleTeacherCourses)))

	//    Assignment Routes
	mux.HandleFunc("GET /teacher/courses/{courseId}/assignments", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleTeacherAssignments)))
	mux.HandleFunc("GET /teacher/courses/{courseId}/assignments/{assignmentId}", app.withAuth("teacher", app.makeHTTPHandlerFunc((app.handleTeacherAssignmentDetail))))
	mux.HandleFunc("POST /teacher/courses/{courseId}/assignments/new", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleTeacherAssignmentCreate)))
	mux.HandleFunc("POST /teacher/courses/{courseId}/assignments/{assignmentId}", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleTeacherAssignmentUpdate)))

	//    Student Routes
	mux.HandleFunc("GET /teacher/courses/{courseId}/students", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handleStudents)))
	mux.HandleFunc("POST /teacher/courses/{courseId}/students/{studentId}/passwordreset", app.withAuth("teacher", app.makeHTTPHandlerFunc(app.handlePasswordReset)))

	mux.HandleFunc("GET /ip", func(w http.ResponseWriter, r *http.Request) {
		// 1. Check X-Real-IP header (often set by nginx)
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			w.Write([]byte(xri))
			return
		}

		// 2. Check X-Forwarded-For header
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				w.Write([]byte(strings.TrimSpace(ips[0])))
				return
			}
		}

		// 3. Fallback to RemoteAddr
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Write([]byte(ip))
	})
	return mux
}

func (app *Application) makeHTTPHandlerFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := f(w, r)
		if err != nil {
			se := &errs.ServerError{}
			jwte := &errs.JwtError{}
			if errors.As(err, &se) {
				app.Logger.Error("Server Error", slog.String("msg", se.Internal))
				w.WriteHeader(se.Status)
				err := app.renderTemplate(w, "error_page", map[string]any{"Code": se.Status, "Text": http.StatusText(se.Status)})
				if err != nil {
					app.Logger.Error("Could not render error page template... aborting")
					os.Exit(1)
				}
			} else if errors.Is(err, http.ErrNoCookie) {
				app.Logger.Warn("No auth token provided trying to access " + r.URL.Path)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.As(err, &jwte) {
				app.Logger.Warn("Issue validating JWT for "+r.URL.Path, slog.String("msg", jwte.Message))
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if errors.Is(err, &errs.UnauthorizedError{}) {
				app.Logger.Warn("Unauthorized access attempt to "+r.URL.Path, slog.String("host", r.Host), slog.String("remote_addr", r.RemoteAddr))
				w.WriteHeader(http.StatusNotFound)
				err := app.renderTemplate(w, "error_page", map[string]any{"Code": http.StatusNotFound, "Text": http.StatusText(http.StatusNotFound)})
				if err != nil {
					app.Logger.Error("Could not render error page template... aborting")
					os.Exit(1)
				}
			} else {
				app.Logger.Error("Something unexpected happened", slog.String("msg", err.Error()))
				// Log user out just in case
				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    "",
					HttpOnly: true,
					MaxAge:   -1,
				})
				err := app.renderTemplate(w, "error_page", map[string]any{"Code": http.StatusInternalServerError, "Text": http.StatusText(http.StatusInternalServerError)})
				if err != nil {
					app.Logger.Error("Could not render error page template... aborting")
					os.Exit(1)
				}
			}
		}
	}
}
