package server

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/philip-h/amics/internal/storage"
)

// ============================================================================
// Login
// ============================================================================
func handleAuthLoginGet(tmpl *template.Template) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			// Check to see if the user is already logged in, if so redirect to home page
			_, err := r.Cookie("session_id")
			if err == nil {
				http.Redirect(w, r, "/app", http.StatusSeeOther)
				return
			}

			body := make(map[string]string)
			problems := make(map[string]string)
			tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		},
	)
}

func handleAuthLoginPost(
	logger *Logger,
	store *storage.Storage,
	tmpl *template.Template,
) http.Handler {
	type request struct {
		Username   string
		Password   string
		RememberMe bool
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Read the request body from form values
			body := &request{
				Username:   r.PostFormValue("username"),
				Password:   r.PostFormValue("password"),
				RememberMe: false,
			}

			problems := validateLoginReq(body.Username, body.Password)
			if len(problems) != 0 {
				tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
				return
			}

			// Look for person in the database
			person, err := store.People.GetByUsername(body.Username)
			if err != nil {
				logger.L.Error("Could not get person by username", slog.String("msg", err.Error()))
				problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
				body.Password = ""
				tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
				return
			}

			if person == nil {
				problems["server"] = "Hmm, I could not find your account."
				logger.L.Info("Cannot find user", "problems", problems)
				body.Password = ""
				tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
				return
			}

			if ok := store.People.CompareHashAndPassword(person.Password, body.Password); !ok {
				problems["server"] = "Hmm, I could not find your account."
				logger.L.Info("Incorrect password", "problems", problems)
				body.Password = ""
				tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
				return
			}

			// Create session
			sessionId, err := store.Sessions.Create(person.Id, body.RememberMe)
			if err != nil {
				logger.L.Error("Could not create session", slog.String("msg", err.Error()))
				problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
				body.Password = ""
				tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
				return
			}

			var maxAge int
			if body.RememberMe {
				// 7 days in seconds
				maxAge = 7 * 24 * 60 * 60
			} else {
				// expires when browser is closed!
				maxAge = 0
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionId,
				HttpOnly: true,
				MaxAge:   maxAge,
				SameSite: http.SameSiteLaxMode,
			})

			// Redirect to home page and send back the cookie
			http.Redirect(w, r, "/app", http.StatusSeeOther)
		},
	)

}

func handleAuthLoginValidation(tmpl *template.Template) http.Handler {
	type request struct {
		Username string
		Password string
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			body := &request{
				Username: r.PostFormValue("username"),
				Password: r.PostFormValue("password"),
			}
			problems := validateLoginReq(body.Username, body.Password)

			tmpl.ExecuteTemplate(w, "login_form_errors", map[string]any{"Problems": problems, "Body": body})
		},
	)
}

// ============================================================================
// Register
// ============================================================================

func handleAuthRegisterGet(tmpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check to see if the user is already logged in, if so redirect to home page
		_, err := r.Cookie("session_id")

		if err != http.ErrNoCookie {
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			return
		}
		// If the route has a query parameter for the join code, pass it into the template
		joinCode := r.URL.Query().Get("joincode")
		body := make(map[string]string)
		problems := make(map[string]string)
		tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "JoinCode": joinCode, "Problems": problems})
	})
}

func handleAuthRegisterPost(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {

	type request struct {
		StudentNumber string
		FirstName     string
		Username      string
		Password      string
		JoinCode      string
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Read the request body from form values
		body := &request{
			StudentNumber: r.FormValue("student-number"),
			FirstName:     r.FormValue("first-name"),
			Username:      r.FormValue("username"),
			Password:      r.FormValue("password"),
			JoinCode:      r.FormValue("join-code"),
		}

		problems := validateRegisterReq(
			store,
			body.StudentNumber,
			body.FirstName,
			body.Username,
			body.Password,
			body.JoinCode,
		)

		if len(problems) != 0 {
			tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
			return
		}
		course, err := store.Courses.GetByJoinCode(body.JoinCode)
		if err != nil {
			logger.L.Error("Could not get course by join code", slog.String("err", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
			return
		}

		// Create a user
		intStudentNumber, err := strconv.Atoi(body.StudentNumber)
		if err != nil {
			problems["student_number"] = "Student number must be only numeric values"
			body.Password = ""
			body.StudentNumber = ""
			tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
			return
		}
		person := &storage.Person{
			Id:       intStudentNumber,
			Username: body.Username,
			Password: body.Password,
		}

		err = store.People.Create(person, course.Id)
		if err != nil {
			logger.L.Error("Could not create student"+body.Username, slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			tmpl.ExecuteTemplate(w, "register", map[string]any{"Body": body, "Problems": problems})
			return
		}

		// Create session
		sessionId, err := store.Sessions.Create(person.Id, false)
		if err != nil {
			logger.L.Error("Could not create session", slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionId,
			HttpOnly: true,
			MaxAge:   0,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to home page and send back the cookie
		http.Redirect(w, r, "/app", http.StatusSeeOther)
	})
}

func handleAuthRegisterValidation(store *storage.Storage, tmpl *template.Template) http.Handler {
	type request struct {
		StudentNumber string
		FirstName     string
		Username      string
		Password      string
		JoinCode      string
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			body := &request{
				StudentNumber: r.PostFormValue("student-number"),
				FirstName:     r.PostFormValue("first-name"),
				Username:      r.PostFormValue("username"),
				Password:      r.PostFormValue("password"),
				JoinCode:      r.PostFormValue("join-code"),
			}
			problems := validateRegisterReq(
				store,
				body.StudentNumber,
				body.FirstName,
				body.Username,
				body.Password,
				body.JoinCode,
			)

			tmpl.ExecuteTemplate(w, "register_form_errors", map[string]any{"Problems": problems, "Body": body})
		},
	)
}

// ============================================================================
// Logout
// ============================================================================
func handleAuthLogout(logger *Logger, store *storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session_id")
		if err == http.ErrNoCookie {
			return
		}

		err = store.Sessions.Delete(cookie.Value)
		if err != nil {
			logger.L.Error("Could not delete session", "err", err)
			// no return - I still want to clear the session cookie even if I cannot delete the session
		}

		// Clear the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			HttpOnly: true,
			MaxAge:   -1,
		})
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
	})
}
