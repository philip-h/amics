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
func (app *Application) handleIndex(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/base.gohtml", "templates/home.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "home"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

func (app *Application) handleLoginGet(w http.ResponseWriter, r *http.Request) {
	// Check to see if the user is already logged in, if so redirect to home page
	// TODO: Actually verify the token instead of just checking if it exists
	cookie, err := r.Cookie("token")
	if err == nil {
		if cookie.Value == "dummy-token" {
			log.Println("User already logged in, redirecting to home page")
			http.Redirect(w, r, "/app", http.StatusSeeOther)
			return
		}
		return
	}

	t, err := template.ParseFiles("templates/base.gohtml", "templates/login.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "login"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Application) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	// Read the request body from form values
	user := &LoginReq{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	// If either username or password is empty, return an error
	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check user against db
	dbUser, err := app.Store.Users.GetByUsername(user.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create a session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:	"dummy-token-"+dbUser.Username, // TODO: Generate a real token
		HttpOnly: true,
	})

	// Redirect to home page and send back the cookie
	http.Redirect(w, r, "/app", http.StatusSeeOther)

}

// =====================================
// Register Handlers
// =====================================
func (app *Application) handleRegisterGet(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/base.gohtml", "templates/register.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "register"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type RegisterReq struct {
	StudentNumber string `json:"student_number"`
	Username string `json:"username"`
	Password string `json:"password"`
}
func (app *Application) handleRegisterPost(w http.ResponseWriter, r *http.Request) {
	// Read the request body from form values
	body := &RegisterReq{
		StudentNumber: r.FormValue("student-number"),
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	log.Printf("Received registration request: %+v\n", body)
	// If either username or password is empty, return an error
	if body.StudentNumber == "" || body.Username == "" || body.Password == "" {
		http.Error(w, "All fields required", http.StatusBadRequest)
		return
	}

	// Create a user
	user := &store.User{
		StudentNumber: body.StudentNumber,
		Username: body.Username,
		Password: body.Password,
	}
	err := app.Store.Users.Create(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:	"dummy-token-"+user.Username, // TODO: Generate a real token
		HttpOnly: true,
	})

	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

// =====================================
// Logout Handler
// =====================================
func (app *Application) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:	 "token",
		Value:	 "",
		HttpOnly: true,
		MaxAge:  -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ============================================================================
// Dashboard Handler
// ============================================================================
func (app *Application) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Check to see if the user is logged in, if not redirect to login page
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract username from token
	username := cookie.Value[len("dummy-token-"):]

	t, err := template.ParseFiles("templates/base.gohtml", "templates/dashboard.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "base", map[string]string{"Active": "app", "Username": username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}