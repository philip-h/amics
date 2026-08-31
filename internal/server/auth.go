package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
	"github.com/philip-h/amics/view/auth"
	"github.com/philip-h/amics/view/guest"
	"github.com/philip-h/amics/view/model"
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
func handleAuthLoginGet() http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		// Check to see if the user is already logged in, if so redirect to home page
		personId := r.Context().Value(personKey)
		if personId != nil {
			redirect(w, r, "/h")
			return nil
		}

		// Copying the meta idea - the log in page _is_ the home page
		return guest.Home(model.NewLoginFormGet()).Render(r.Context(), w)
	})
}

func handleAuthLoginPost(
	logger *Logger,
	store *storage.Storage,
) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Read the request body from form values
		body := model.NewLoginFormPost(
			r.PostFormValue("username"),
			r.PostFormValue("password"),
		)

		problems := body.Validate()

		if len(problems) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			body.Password = ""
			return guest.Home(body).Render(r.Context(), w)
		}

		// Look for person in the database
		person, err := store.People.GetByUsername(body.Username)
		if err != nil {
			logger.L.Error("[auth.login.post] Could not get person by username", slog.String("msg", err.Error()))
			body.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return guest.Home(body).Render(r.Context(), w)
		}

		if person == nil {
			logger.L.Info("[auth.login.post] Cannot find user", "username", body.Username)
			body.ServerError = "Hmm, I could not find your account."
			body.Password = ""
			return guest.Home(body).Render(r.Context(), w)
		}

		if ok := store.People.CompareHashAndPassword(person.Password, body.Password); !ok {
			logger.L.Info("[auth.login.post] Incorrect password", "problems", problems)
			body.ServerError = "Hmm, I could not find your account."
			body.Password = ""
			return guest.Home(body).Render(r.Context(), w)
		}

		// Create session
		// TODO: add remember me functionality
		sessionId, err := store.Sessions.Create(person.Id, true)
		if err != nil {
			logger.L.Error("[auth.login.post] Could not create session", slog.String("msg", err.Error()))
			body.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return guest.Home(body).Render(r.Context(), w)
		}

		var maxAge int
		if true { // TODO: add remember me functionality
			// 7 days in seconds
			maxAge = 7 * 24 * 60 * 60
		} else {
			// expires when browser is closed!
			maxAge = 0
		}

		http.SetCookie(w, authCookie(sessionId, maxAge))

		redirect(w, r, "/h")
		return nil
	})
}

func handleAuthLoginValidation() http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		body := model.NewLoginFormPost(
			r.FormValue("username"),
			r.FormValue("password"),
		)

		return guest.LoginFormErrors(body).Render(r.Context(), w)
	})
}

// ============================================================================
// Register
// ============================================================================

func handleAuthRegisterGet() http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Check to see if the user is already logged in, if so redirect to home page
		personId := r.Context().Value(personKey)
		if personId != nil {
			redirect(w, r, "/h")
			return nil
		}

		// If the route has a query parameter for the join code, pass it into the template
		joinCode := r.URL.Query().Get("joincode")
		formVM := model.NewRegisterFormGet()
		formVM.JoinCode = joinCode
		return auth.Register(formVM).Render(r.Context(), w)
	})
}

func handleAuthRegisterPost(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		// Read the request body from form values
		body := model.NewRegisterFormPost(
			store,
			r.FormValue("first-name"),
			r.FormValue("student-number"),
			r.FormValue("username"),
			r.FormValue("password"),
			r.FormValue("join-code"),
		)

		if len(body.Validate()) != 0 {
			logger.L.Debug("[auth.register.post] Body validation failed", slog.Any("problems", body.Validate()))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return auth.Register(body).Render(r.Context(), w)
		}

		course, err := store.Courses.GetByJoinCode(body.JoinCode)
		if err != nil {
			logger.L.Error("[auth.register.post] Could not get course by join code", slog.String("err", err.Error()))
			body.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return auth.Register(body).Render(r.Context(), w)
		}

		// Already validated
		intStudentNumber, _ := strconv.Atoi(body.StudentNumber)
		person := &storage.Person{
			Id:        intStudentNumber,
			FirstName: body.FirstName,
			Username:  body.Username,
			Password:  body.Password,
		}

		err = store.People.Create(person, course.Id)
		if err != nil {
			logger.L.Error("[auth.register.post] Could not create student"+body.Username, slog.String("msg", err.Error()))
			body.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return auth.Register(body).Render(r.Context(), w)
		}

		// Create session
		sessionId, err := store.Sessions.Create(person.Id, false)
		if err != nil {
			logger.L.Error("[auth.register.post] Could not create session", slog.String("msg", err.Error()))
			body.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return auth.Register(body).Render(r.Context(), w)
		}

		http.SetCookie(w, authCookie(sessionId, 0))

		redirect(w, r, "/h")
		return nil
	})
}

func handleAuthRegisterValidation(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		body := model.NewRegisterFormPost(
			store,
			r.FormValue("first-name"),
			r.FormValue("student-number"),
			r.FormValue("username"),
			r.FormValue("password"),
			r.FormValue("join-code"),
		)

		return auth.RegisterFormErrors(body).Render(r.Context(), w)
	})
}

// ============================================================================
// Logout
// ============================================================================
func handleAuthLogout(logger *Logger, store *storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("amics-cookie")
		if err == http.ErrNoCookie {
			return
		}

		err = store.Sessions.Delete(cookie.Value)
		if err != nil {
			logger.L.Error("[auth.logout] Could not delete session", slog.String("msg", err.Error()))
			// no return - I still want to clear the session cookie even if I cannot delete the session
		}

		// Clear the session cookie
		http.SetCookie(w, logoutCookie())
		redirect(w, r, "/")
	})
}
