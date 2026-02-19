package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/errs"
	"github.com/philip-h/amics/internal/store"
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

	return app.renderTemplate(w, "login", map[string]string{"Active": "login"})
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

	// Look for student in the database
	student, err := app.Store.Students.GetByUsername(body.Username)
	if err != nil {
		return &errs.JsonError{
			Status:   http.StatusInternalServerError,
			Message:  "Sorry, something went seriously wrong on our end. Please try again in a sec.",
			Internal: err.Error(),
		}
	}
	type UserPass struct {
		Username string
		Password string
		Role     string
	}

	var user *UserPass
	if student != nil {
		user = &UserPass{
			Username: student.Username,
			Password: student.Password,
			Role:     "student",
		}
	} else {
		// Check if the user is a teacher
		teacher, err := app.Store.Teachers.GetByUsername(body.Username)
		if err != nil {
			return &errs.JsonError{
				Status:   http.StatusInternalServerError,
				Message:  "Sorry, something went seriously wrong on our end. Please try again in a sec.",
				Internal: err.Error(),
			}
		}
		if teacher == nil {
			return &errs.JsonError{
				Status:   http.StatusUnauthorized,
				Message:  "Hmm, I could not find your account.",
				Internal: "No student or teacher found with username: " + body.Username,
			}
		}
		// For simplicity, we will treat teachers the same as students for authentication purposes. In a real implementation, you would likely want to have different handling for teachers and students after this point.
		user = &UserPass{
			Username: teacher.Username,
			Password: teacher.Password,
			Role:     "teacher",
		}
	}

	// Right now all of the dummy data in the database does not have hashed passwords, so we will skip the password check for now. In a real implementation, you would want to hash the password when creating the user and compare the hashed password here.
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	// if err != nil {
	if user.Password != body.Password {
		return &errs.JsonError{
			Status:   http.StatusUnauthorized,
			Message:  "Hmm, I could not find your account.",
			Internal: "Password mismatch for user: " + body.Username,
		}
	}

	// Create JWT token and set it as a cookie
	token, err := app.Auth.CreateJwt(user.Username, user.Role, time.Now().Add(90*time.Minute).Unix())
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
	CourseId      string `json:"course_id"`
}

func (app *Application) handleRegisterPost(w http.ResponseWriter, r *http.Request) error {
	// Read the request body from form values
	body := &RegisterReq{
		StudentNumber: r.FormValue("student-number"),
		Username:      r.FormValue("username"),
		Password:      r.FormValue("password"),
		CourseId:      r.FormValue("course"),
	}
	// If either username or password is empty, return an error
	if body.StudentNumber == "" || body.Username == "" || body.Password == "" || body.CourseId == "" {
		return &errs.JsonError{
			Status:   http.StatusBadRequest,
			Message:  "Please make sure to fill out all required fields.",
			Internal: "Missing student number, username, password, or course_id in registration request",
		}
	}

	// Create a user
	courseId, err := strconv.Atoi(body.CourseId)
	if err != nil {
		return &errs.JsonError{
			Status:   http.StatusBadRequest,
			Message:  "Invalid course selection.",
			Internal: "Failed to convert course_id to int: " + body.CourseId,
		}
	}

	user := &store.Student{
		StudentNumber: body.StudentNumber,
		Username:      body.Username,
		Password:      body.Password,
		CourseId:      courseId,
	}

	err = app.Store.Students.Create(user)
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
// App Route Handlers
// ============================================================================
func (app *Application) handleDashboard(w http.ResponseWriter, r *http.Request) error {
	username := r.Context().Value("username").(string)

	assignments, err := app.Store.Assignments.GetByUsername(username)
	if err != nil {
		return err
	}

	return app.renderTemplate(w,
		"dashboard",
		map[string]any{
			"Active":      "app",
			"Username":    username,
			"Assignments": assignments,
		})
}

func (app *Application) handleAssignmentDetail(w http.ResponseWriter, r *http.Request) error {

	username := r.Context().Value("username").(string)
	assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("assignmentId"),
		}
	}

	assignment, err := app.Store.Assignments.GetById(assignmentId, username)
	if err != nil {
		return err
	}

	return app.renderTemplate(w,
		"assignment_detail",
		map[string]any{
			"Active":     "app",
			"Assignment": assignment,
		})
}

func (app *Application) handleAssignmentSubmit(w http.ResponseWriter, r *http.Request) error {
	username := r.Context().Value("username").(string)
	assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("assignmentId"),
		}
	}

	// Limit upload size to 10MB
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: err.Error(),
		}
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: err.Error(),
		}
	}
	defer file.Close()

	// Read the file content into a byte slice
	fileContent := make([]byte, handler.Size)
	_, err = file.Read(fileContent)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusInternalServerError,
			Internal: err.Error(),
		}
	}

	pyFile := &store.PyFile{
		Name:    handler.Filename,
		Content: string(fileContent),
	}
	err = app.Store.Assignments.Submit(assignmentId, username, pyFile)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusInternalServerError,
			Internal: err.Error(),
		}
	}

	http.Redirect(w, r, "/app/assignment/"+strconv.Itoa(assignmentId), http.StatusSeeOther)
	return nil
}
