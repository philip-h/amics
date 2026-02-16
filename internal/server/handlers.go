package server

import (
	"html/template"
	"log"
	"net/http"

	"github.com/philip-h/amics/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Homepage
// ============================================================================
func (app *Application) handleIndex(w http.ResponseWriter, r *http.Request) error {
	t, err := template.ParseFiles("templates/base.gohtml", "templates/home.gohtml")
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "home"})
	if err != nil {
		return err
	}
	return nil
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
	// TODO: Actually verify the token instead of just checking if it exists
	cookie, err := r.Cookie("token")
	if err == nil {
		if cookie.Value == "dummy-token" {
			log.Println("User already logged in, redirecting to home page")
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			return nil
		}
		return nil
	}

	t, err := template.ParseFiles("templates/base.gohtml", "templates/login.gohtml")
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "login"})
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
	// Read the request body from form values
	body := &LoginReq{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	// If either username or password is empty, return an error
	if body.Username == "" || body.Password == "" {
		return &JsonError{
			Status: http.StatusBadRequest,
			Message: "Please make sure to fill out both the username and password fields.",
			Internal: "Missing username or password in login request",
		}
	}

	// Check user against db
	user, err := app.Store.Users.GetByUsername(body.Username)
	if err != nil {
		return &JsonError{
			Status: http.StatusInternalServerError,
			Message: "Sorry, something went seriously wrong on our end. Please try again in a sec.",
			Internal: err.Error(),

		}
	}
	if user == nil {
		return &JsonError{
			Status: http.StatusUnauthorized,
			Message: "Hmm, I could not find your account.",
			Internal: "No user found with username: " + body.Username,
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return &JsonError{
			Status: http.StatusUnauthorized,
			Message: "Hmm, I could not find your account.",
			Internal: "Password mismatch for user: " + body.Username,
		}
	}

	// Create a session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:	"dummy-token-"+user.Username, // TODO: Generate a real token
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
	t, err := template.ParseFiles("templates/base.gohtml", "templates/register.gohtml")
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "register"})
	if err != nil {
		return err
	}
	return nil
}

type RegisterReq struct {
	StudentNumber string `json:"student_number"`
	Username string `json:"username"`
	Password string `json:"password"`
}
func (app *Application) handleRegisterPost(w http.ResponseWriter, r *http.Request) error {
	// Read the request body from form values
	body := &RegisterReq{
		StudentNumber: r.FormValue("student-number"),
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	// If either username or password is empty, return an error
	if body.StudentNumber == "" || body.Username == "" || body.Password == "" {
		return &JsonError{
			Status: http.StatusBadRequest,
			Message: "Please make sure to fill out all required fields.",
			Internal: "Missing student number, username, or password in registration request",
		}
	}

	// Create a user
	user := &store.User{
		StudentNumber: body.StudentNumber,
		Username: body.Username,
		Password: body.Password,
	}
	err := app.Store.Users.Create(user)
	if err != nil {
		return err
	}
	// Create a session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:	"dummy-token-"+user.Username, // TODO: Generate a real token
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
		Name:	 "token",
		Value:	 "",
		HttpOnly: true,
		MaxAge:  -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// ============================================================================
// Dashboard Handler
// ============================================================================
func (app *Application) handleDashboard(w http.ResponseWriter, r *http.Request) error {
	// Check to see if the user is logged in, if not redirect to login page
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil
	}

	// Extract username from token
	username := cookie.Value[len("dummy-token-"):]

	t, err := template.ParseFiles("templates/base.gohtml", "templates/dashboard.gohtml")
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "app", "Username": username})
	if err != nil {
		return err
	}
	return nil
}