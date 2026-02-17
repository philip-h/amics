package server

import (
	"net/http"
	"time"

	"github.com/philip-h/amics/internal/errs"
	"github.com/philip-h/amics/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func (app *Application) renderTemplate(w http.ResponseWriter, name string, data any) error {
	return app.Templates.ExecuteTemplate(w, name, data)
}

// ============================================================================
// Homepage
// ============================================================================
func (app *Application) handleIndex(w http.ResponseWriter, r *http.Request) error {
	return app.renderTemplate(w, "home", map[string]string{"Active": "home"})
}

// ============================================================================
// Auth Handlers
// ============================================================================

// =====================================
// Login Handlers
// =====================================

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *Application) handleLoginGet(w http.ResponseWriter, r *http.Request) error {
	// Check to see if the user is already logged in, if so redirect to home page
	_, err := r.Cookie("token")
	if err == nil {
		// This redirect will check the validity of the token
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return nil
	}

	return app.renderTemplate(w, "login", map[string]string{"Active": "home"})
}

func (app *Application) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
	// Read the request body from form values
	body := &LoginReq{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	// If either username or password is empty, return an error
	if body.Username == "" || body.Password == "" {
		return &errs.JsonError{
			Status:   http.StatusBadRequest,
			Message:  "Please make sure to fill out both the username and password fields.",
			Internal: "Missing username or password in login request",
		}
	}

	// Check user against db
	user, err := app.Store.Users.GetByUsername(body.Username)
	if err != nil {
		return &errs.JsonError{
			Status:   http.StatusInternalServerError,
			Message:  "Sorry, something went seriously wrong on our end. Please try again in a sec.",
			Internal: err.Error(),
		}
	}
	if user == nil {
		return &errs.JsonError{
			Status:   http.StatusUnauthorized,
			Message:  "Hmm, I could not find your account.",
			Internal: "No user found with username: " + body.Username,
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return &errs.JsonError{
			Status:   http.StatusUnauthorized,
			Message:  "Hmm, I could not find your account.",
			Internal: "Password mismatch for user: " + body.Username,
		}
	}

	// Create JWT token and set it as a cookie
	token, err := app.Auth.CreateJwt(user.Username, "student", time.Now().Add(90*time.Minute).Unix())
	if err != nil {
		return &errs.JsonError{
			Status:   http.StatusInternalServerError,
			Message:  "Sorry, something went seriously wrong on our end. Please try again in a sec.",
			Internal: err.Error(),
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	})

	// Redirect to home page and send back the cookie
	http.Redirect(w, r, "/app", http.StatusSeeOther)
	return nil
}

// =====================================
// Register Handlers
// =====================================
func (app *Application) handleRegisterGet(w http.ResponseWriter, r *http.Request) error {
	// Check to see if the user is already logged in, if so redirect to home page
	_, err := r.Cookie("token")
	if err == nil {
		// This redirect will check the validity of the token
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return nil
	}
	return app.renderTemplate(w, "register", map[string]string{"Active": "register"})
}

type RegisterReq struct {
	StudentNumber string `json:"student_number"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

func (app *Application) handleRegisterPost(w http.ResponseWriter, r *http.Request) error {
	// Read the request body from form values
	body := &RegisterReq{
		StudentNumber: r.FormValue("student-number"),
		Username:      r.FormValue("username"),
		Password:      r.FormValue("password"),
	}
	// If either username or password is empty, return an error
	if body.StudentNumber == "" || body.Username == "" || body.Password == "" {
		return &errs.JsonError{
			Status:   http.StatusBadRequest,
			Message:  "Please make sure to fill out all required fields.",
			Internal: "Missing student number, username, or password in registration request",
		}
	}

	// Create a user
	user := &store.User{
		StudentNumber: body.StudentNumber,
		Username:      body.Username,
		Password:      body.Password,
	}
	err := app.Store.Users.Create(user)
	if err != nil {
		return err
	}
	// Create jwt token and set it as a cookie
	token, err := app.Auth.CreateJwt(user.Username, "student", time.Now().Add(90*time.Minute).Unix())
	if err != nil {
		return &errs.JsonError{
			Status:   http.StatusInternalServerError,
			Message:  "Sorry, something went seriously wrong on our end. Please try again in a sec.",
			Internal: err.Error(),
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/app", http.StatusSeeOther)
	return nil
}

// =====================================
// Logout Handler
// =====================================
func (app *Application) handleLogout(w http.ResponseWriter, r *http.Request) error {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// ============================================================================
// Dashboard Handler
// ============================================================================
func (app *Application) handleDashboard(w http.ResponseWriter, r *http.Request) error {
	username := r.Context().Value("username").(string)

	return app.renderTemplate(w, "dashboard", map[string]string{"Active": "app", "Username": username})
}

