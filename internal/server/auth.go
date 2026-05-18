package server

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
)

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

// ============================================================================
// Login
// ============================================================================
func handleAuthLoginGet(tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		// Check to see if the user is already logged in, if so redirect to home page
		personId := r.Context().Value(personKey)
		if personId != nil {
			redirect(w, r, "/app")
			return nil
		}

		body := make(map[string]string)
		problems := make(map[string]string)
		return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
	})
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

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Read the request body from form values
		body := &request{
			Username:   r.PostFormValue("username"),
			Password:   r.PostFormValue("password"),
			RememberMe: false,
		}

		problems := validateLoginReq(body.Username, body.Password)
		if len(problems) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		// Look for person in the database
		person, err := store.People.GetByUsername(body.Username)
		if err != nil {
			logger.L.Error("Could not get person by username", slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		if person == nil {
			problems["server"] = "Hmm, I could not find your account."
			logger.L.Info("Cannot find user", "username", body.Username)
			body.Password = ""
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		if ok := store.People.CompareHashAndPassword(person.Password, body.Password); !ok {
			problems["server"] = "Hmm, I could not find your account."
			logger.L.Info("Incorrect password", "problems", problems)
			body.Password = ""
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		// Create session
		sessionId, err := store.Sessions.Create(person.Id, body.RememberMe)
		if err != nil {
			logger.L.Error("Could not create session", slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
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

		redirect(w, r, "/app")
		return nil
	})
}

func handleAuthLoginValidation(tmpl *template.Template) http.Handler {
	type request struct {
		Username string
		Password string
	}
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		body := &request{
			Username: r.PostFormValue("username"),
			Password: r.PostFormValue("password"),
		}
		problems := validateLoginReq(body.Username, body.Password)

		return tmpl.ExecuteTemplate(w, "login_form_errors", map[string]any{"Problems": problems, "Body": body})
	})
}

// ============================================================================
// Register
// ============================================================================

func handleAuthRegisterGet(tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Check to see if the user is already logged in, if so redirect to home page
		personId := r.Context().Value(personKey)
		if personId != nil {
			redirect(w, r, "/app")
			return nil
		}

		// If the route has a query parameter for the join code, pass it into the template
		joinCode := r.URL.Query().Get("joincode")
		body := make(map[string]string)
		problems := make(map[string]string)
		return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "JoinCode": joinCode, "Problems": problems})
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

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

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
			w.WriteHeader(http.StatusUnprocessableEntity)
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		course, err := store.Courses.GetByJoinCode(body.JoinCode)
		if err != nil {
			logger.L.Error("Could not get course by join code", slog.String("err", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		// Already validated
		intStudentNumber, _ := strconv.Atoi(body.StudentNumber)
		person := &storage.Person{
			Id:       intStudentNumber,
			Username: body.Username,
			Password: body.Password,
		}

		err = store.People.Create(person, course.Id)
		if err != nil {
			logger.L.Error("Could not create student"+body.Username, slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "register", map[string]any{"Body": body, "Problems": problems})
		}

		// Create session
		sessionId, err := store.Sessions.Create(person.Id, false)
		if err != nil {
			logger.L.Error("Could not create session", slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Body": body, "Problems": problems})
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionId,
			HttpOnly: true,
			MaxAge:   0,
			SameSite: http.SameSiteLaxMode,
		})

		redirect(w, r, "/app")
		return nil
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
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
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

		return tmpl.ExecuteTemplate(w, "register_form_errors", map[string]any{"Problems": problems, "Body": body})
	})
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
		redirect(w, r, "/login")
	})
}
