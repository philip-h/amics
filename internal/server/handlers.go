package server

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/philip-h/amics/internal/errs"
	"github.com/philip-h/amics/internal/store"
	"github.com/yuin/goldmark"
)

// ============================================================================
// Helpers
// ============================================================================
func (app *Application) renderPage(w http.ResponseWriter, name string, data any) error {
	tmpl, ok := app.Templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}

func (app *Application) renderPartial(w http.ResponseWriter, name string, data any) error {
	tmpl, ok := app.Templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

func (app *Application) requestLogger(r *http.Request, fn string) *slog.Logger {
	return app.Logger.WithGroup("where").With(
		slog.String("function", fn),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("host", r.Host),
		slog.String("remote_addr", r.RemoteAddr),
	)
}

type NavLink struct {
	Text string
	Href string
}

// ============================================================================
// Homepage
// ============================================================================
func (app *Application) handleIndex(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		err := app.renderPage(w, "error_page", map[string]any{"Code": http.StatusNotFound, "Text": http.StatusText(http.StatusNotFound)})
		return err
	}
	return app.renderPage(w, "home", map[string]string{"Active": "home"})
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

func (app *Application) handleLoginValidation(w http.ResponseWriter, r *http.Request) error {
	body := &LoginReq{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}
	errors := app.validateLoginReq(body)

	return app.renderPartial(w, "login_form_errors", map[string]any{"Errors": errors, "Body": body})
}

func (app *Application) handleLoginGet(w http.ResponseWriter, r *http.Request) error {
	// Check to see if the user is already logged in, if so redirect to home page
	_, err := r.Cookie("token")
	if err == nil {
		// This redirect will check the validity of the token
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return nil
	}

	body := &LoginReq{}
	errors := make(map[string]string)
	return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
}

func (app *Application) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleLoginPost")
	// Read the request body from form values
	body := &LoginReq{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	errors := app.validateLoginReq(body)
	if len(errors) != 0 {
		return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
	}
	// Look for student in the database
	student, err := app.Store.Students.GetByUsername(body.Username)
	if err != nil {
		log.Error("Could not get student by username", slog.String("msg", err.Error()))
		errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
		body.Password = ""
		return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
	}
	type UserPass struct {
		Id       int
		Username string
		Password string
		Role     string
	}

	var user *UserPass
	if student != nil {
		user = &UserPass{
			Id:       student.Id,
			Username: student.Username,
			Password: student.Password,
			Role:     "student",
		}
	} else {
		// Check if the user is a teacher
		teacher, err := app.Store.Teachers.GetByUsername(body.Username)
		if err != nil {
			log.Error("Could not get teacher by username", slog.String("msg", err.Error()))
			errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			body.Password = ""
			return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
		}
		if teacher == nil {
			errors["server"] = "Hmm, I could not find your account."
			body.Password = ""
			return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
		}
		user = &UserPass{
			Id:       teacher.Id,
			Username: teacher.Username,
			Password: teacher.Password,
			Role:     "teacher",
		}
	}

	if ok := app.Store.Students.CompareHashAndPassword(user.Password, body.Password); !ok {
		errors["server"] = "Hmm, I could not find your account."
		body.Password = ""
		return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
	}
	// Create JWT token and set it as a cookie
	token, err := app.Auth.CreateJwt(strconv.Itoa(user.Id), user.Role, time.Now().Add(90*time.Minute))
	if err != nil {
		log.Error("Could not create jwt", slog.String("msg", err.Error()))
		errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
		body.Password = ""
		return app.renderPage(w, "login", map[string]any{"Body": body, "Errors": errors})
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

type RegisterReq struct {
	StudentNumber string
	Username      string
	Password      string
	JoinCode      string
}

func (app *Application) handleRegisterValidation(w http.ResponseWriter, r *http.Request) error {

	body := &RegisterReq{
		StudentNumber: r.FormValue("student-number"),
		Username:      r.FormValue("username"),
		Password:      r.FormValue("password"),
		JoinCode:      r.FormValue("join-code"),
	}
	errors := app.validateRegisterReq(body)

	return app.renderPartial(w, "register_form_errors", map[string]any{"Body": body, "Errors": errors})
}

func (app *Application) handleRegisterGet(w http.ResponseWriter, r *http.Request) error {
	// Check to see if the user is already logged in, if so redirect to home page
	_, err := r.Cookie("token")
	if err == nil {
		// This redirect will check the validity of the token
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return nil
	}
	// If the route has a query parameter for the join code, pass it into the template
	joinCode := r.URL.Query().Get("joincode")
	body := &RegisterReq{}
	errors := make(map[string]string)
	return app.renderPage(w, "register", map[string]any{"Body": body, "JoinCode": joinCode, "Errors": errors})
}

func (app *Application) handleRegisterPost(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleRegisterPost")
	// Read the request body from form values
	body := &RegisterReq{
		StudentNumber: r.FormValue("student-number"),
		Username:      r.FormValue("username"),
		Password:      r.FormValue("password"),
		JoinCode:      r.FormValue("join-code"),
	}

	errors := app.validateRegisterReq(body)
	if len(errors) != 0 {
		return app.renderPage(w, "register", map[string]any{"Body": body, "Errors": errors})
	}
	course, err := app.Store.Courses.GetByJoinCode(body.JoinCode)
	if err != nil {
		return err
	}

	// Create a user
	user := &store.Student{
		StudentNumber: body.StudentNumber,
		Username:      body.Username,
		Password:      body.Password,
		CourseId:      course.Id,
	}

	err = app.Store.Students.Create(user)
	if err != nil {
		log.Error("Could not create student"+body.Username, slog.String("msg", err.Error()))
		errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
		return app.renderPage(w, "register", map[string]any{"Body": body, "Errors": errors})
	}
	// Create jwt token and set it as a cookie
	token, err := app.Auth.CreateJwt(strconv.Itoa(user.Id), "student", time.Now().Add(90*time.Minute))
	if err != nil {
		log.Error("COuld not create jwt", slog.String("msg", err.Error()))
		errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
		return app.renderPage(w, "register", map[string]any{"Body": body, "Errors": errors})
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
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// ============================================================================
// App Route Handlers
// ============================================================================
func (app *Application) handleDashboard(w http.ResponseWriter, r *http.Request) error {
	userIdStr := r.Context().Value("userId").(string)
	is_teacher := r.Context().Value("is-teacher").(bool)
	if is_teacher {
		http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	}

	studentId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + userIdStr,
		}
	}

	assignments, err := app.Store.Assignments.GetWithGradeByStudentId(studentId)
	if err != nil {
		return err
	}

	// Organize assignment by units for nicer rendering
	byUnit := make(map[string][]*store.AssignmentWithGrade)
	// And also calculate student average
	var studentAvgNum float64
	var studentAvgDenom float64

	for _, ass := range assignments {
		if ass.Visible {
			byUnit[ass.UnitName] = append(byUnit[ass.UnitName], ass)
			if ass.Grade.Valid {
				studentAvgNum += float64(ass.Grade.Int64)
			}
			studentAvgDenom += float64(ass.Points)
		}
	}
	studentAverage := math.Round((studentAvgNum / studentAvgDenom) * 100)

	return app.renderPage(w,
		"app",
		map[string]any{
			"Assignments":    byUnit,
			"StudentAverage": studentAverage,
			"NavLinks": []NavLink{
				{Text: "Dashboard", Href: ""},
			},
		})
}

func (app *Application) handleAssignmentDetail(w http.ResponseWriter, r *http.Request) error {
	userIdStr := r.Context().Value("userId").(string)
	studentId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + userIdStr,
		}
	}
	assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("assignmentId"),
		}
	}

	aws, err := app.Store.Assignments.GetWithSubmissionByAssignmentAndStudentIds(assignmentId, studentId)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	var htmlDescription string
	err = goldmark.Convert([]byte(aws.Description), &buf)
	if err != nil {
		htmlDescription = aws.Description
	} else {
		htmlDescription = buf.String()
	}

	var htmlComments strings.Builder
	if aws.Submission != nil && aws.Submission.Comments.Valid {

		lines := strings.SplitSeq(strings.ReplaceAll(aws.Submission.Comments.String, "\r\n", "\n"), "\n")

		for line := range lines {
			if strings.HasPrefix(line, "E") || strings.HasPrefix(line, ">") || strings.HasPrefix(line, "✘") {
				htmlComments.WriteString("<span style='color: rgb(136, 56.5, 53)'>" + line + "</span>")
			} else if strings.HasPrefix(line, "✔") {
				htmlComments.WriteString("<span style='color: rgb(28.5, 105.5, 84)'>" + line + "</span>")
			} else {
				htmlComments.WriteString(line)
			}
			htmlComments.WriteString("\n")
		}
	}

	return app.renderPage(w,
		"assignment",
		map[string]any{
			"Assignment":  aws.Assignment,
			"Submission":  aws.Submission,
			"Comments":    template.HTML(htmlComments.String()),
			"Description": template.HTML(htmlDescription),
      "DisableSubmit": aws.Submission.Status == "grading",
			"NavLinks": []NavLink{
				{Text: "Dashboard", Href: "/app"},
				{Text: aws.Assignment.Name, Href: ""},
			},
		})
}

func (app *Application) handleAssignmentSubmit(w http.ResponseWriter, r *http.Request) error {
	userIdStr := r.Context().Value("userId").(string)
	studentId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + userIdStr,
		}
	}

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

	// Status is set to pending by deault
	// all pending statuses will be picked up by worker started in main function
	err = app.Store.Submissions.Create(assignmentId, studentId, string(fileContent))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusInternalServerError,
			Internal: err.Error(),
		}
	}
	http.Redirect(w, r, "/app/assignments/"+strconv.Itoa(assignmentId), http.StatusSeeOther)
	return nil
}

func (app *Application) handleSubmitPoll(w http.ResponseWriter, r *http.Request) error {
	userIdStr := r.Context().Value("userId").(string)
	studentId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + userIdStr,
		}
	}

	assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("assignmentId"),
		}
	}

	aws, err := app.Store.Assignments.GetWithSubmissionByAssignmentAndStudentIds(assignmentId, studentId)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Could not get submission with student_id of " + userIdStr + " and assignment_id of " + r.PathValue("assignmentId"),
		}
	}
	var htmlComments strings.Builder
	if aws.Submission != nil && aws.Submission.Comments.Valid {

		lines := strings.SplitSeq(strings.ReplaceAll(aws.Submission.Comments.String, "\r\n", "\n"), "\n")

		for line := range lines {
			if strings.HasPrefix(line, "E") || strings.HasPrefix(line, ">") {
				htmlComments.WriteString("<span style='color: rgb(136, 56.5, 53)'>" + line + "</span>")
			} else {
				htmlComments.WriteString(line)
			}
			htmlComments.WriteString("\n")
		}
	}

	return app.renderPartial(w, "submission_overview", map[string]any{
		"Assignment": aws.Assignment,
		"Submission": aws.Submission,
		"Comments":   template.HTML(htmlComments.String()),
	})

}

// ============================================================================
// Admin Handlers
// ============================================================================

func (app *Application) handleTeacherDashboard(w http.ResponseWriter, r *http.Request) error {
	teacherIdStr := r.Context().Value("userId").(string)
	teacherId, err := strconv.Atoi(teacherIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + teacherIdStr,
		}
	}
	courses, err := app.Store.Courses.GetByTeacherId(teacherId)
	if err != nil {
		return err
	}
	return app.renderPage(w, "teacher", map[string]any{
		"Courses": courses,
		"NavLinks": []NavLink{
			{Text: "Teacher Dashboard", Href: ""},
		},
	})
}

// =====================================
// Course Handlers
// =====================================
type CreateCourseReq struct {
	JoinCode string
	Name     string
	Year     string
	Semester string
}

func (app *Application) handleCourseCreate(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleCourseCreate")

	teacherIdStr := r.Context().Value("userId").(string)
	teacherId, err := strconv.Atoi(teacherIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + teacherIdStr,
		}
	}

	body := &CreateCourseReq{
		JoinCode: r.FormValue("join-code"),
		Name:     r.FormValue("name"),
		Year:     r.FormValue("year"),
		Semester: r.FormValue("semester"),
	}
	errors := app.validateCourseCreate(body)
	log.Info("I", "errors", errors)

	if len(errors) > 0 {
		return app.renderPage(w, "manage_course", map[string]any{
			"Course": nil,
			"Errors": errors,
			"Body":   body,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashbord", Href: "/teacher"},
				{Text: "New Course", Href: ""},
			},
		})
	}

	// Validate course create validates that year and semester are numbers
	year, _ := strconv.Atoi(body.Year)
	semester, _ := strconv.Atoi(body.Semester)

	course := &store.Course{
		Name:      body.Name,
		Semester:  semester,
		Year:      year,
		JoinCode:  body.JoinCode,
		TeacherId: teacherId,
	}

	// Create a course
	err = app.Store.Courses.Create(course)
	if err != nil {
		log.Error("Could not create course", slog.String("msg", err.Error()))
		errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
		return app.renderPage(w, "manage_course", map[string]any{"Course": nil, "Errors": errors})
	}

	http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	return nil
}

func (app *Application) handleCourseUpdate(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleCourseUpdate")
	teacherIdStr := r.Context().Value("userId").(string)
	teacherId, err := strconv.Atoi(teacherIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + teacherIdStr,
		}
	}

	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert courseId ID to int: " + r.PathValue("courseId"),
		}
	}

	// Read the request body from form values
	type CourseBody struct {
		Id       string
		Year     string
		Semester string
		Name     string
	}

	body := &CourseBody{
		Id:       r.FormValue("cid"),
		Year:     r.FormValue("year"),
		Semester: r.FormValue("semester"),
		Name:     r.FormValue("name"),
	}

	if body.Id == "" || body.Year == "" || body.Semester == "" || body.Name == "" {
		return app.renderPage(w, "manage_course", map[string]any{"Course": body, "Error": "Please make sure to fill out all required fields."})
	}
	yearInt, err := strconv.Atoi(body.Year)
	if err != nil {
		return app.renderPage(w, "manage_course", map[string]any{"Course": body, "Error": "Year was not an int"})
	}
	semInt, err := strconv.Atoi(body.Semester)
	if err != nil {
		return app.renderPage(w, "manage_course", map[string]any{"Course": body, "Error": "Semester was not an int"})
	}

	course := &store.Course{
		Id:        courseId,
		Name:      body.Name,
		Semester:  semInt,
		Year:      yearInt,
		TeacherId: teacherId,
	}

	err = app.Store.Courses.Update(course)
	if err != nil {
		log.Error("Could not update course", slog.String("msg", err.Error()))
		return app.renderPage(w, "manage_course", map[string]any{"Course": course, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	return nil
}

func (app *Application) handleTeacherCourses(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleTeacherCourses")
	courseIdStr := r.PathValue("courseId")
	errors := make(map[string]string)

	if courseIdStr == "new" {
		return app.renderPage(w, "manage_course", map[string]any{
			"Course": nil,
			"Errors": errors,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashbord", Href: "/teacher"},
				{Text: "New Course", Href: ""},
			},
		})
	}

	courseId, err := strconv.Atoi(courseIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("courseId"),
		}
	}
	course, err := app.Store.Courses.GetById(courseId)
	if err != nil {
		log.Error("Could not get course by id", slog.String("msg", err.Error()))
		return err
	}

	return app.renderPage(w, "manage_course", map[string]any{
		"Course": course,
		"Errors": errors,
		"NavLinks": []NavLink{
			{Text: "Teacher Dashbord", Href: "/teacher"},
			{Text: course.Name, Href: ""},
		},
	})
}

// =====================================
// Assignment Handlers
// =====================================

func (app *Application) handleTeacherAssignments(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleTeacherAssignments")

	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("courseId"),
		}
	}
	assignments, err := app.Store.Assignments.GetByCourseId(courseId)
	if err != nil {
		log.Error("Could not get assignments by course id", slog.String("msg", err.Error()))
		return err
	}
	return app.renderPage(w, "manage_assignments", map[string]any{
		"Assignments": assignments,
		"CourseId":    courseId,
		"NavLinks": []NavLink{
			{Text: "Teacher Dashboard", Href: "/teacher"},
			{Text: "Manage Assignments", Href: ""},
		},
	})
}

func (app *Application) handleTeacherAssignmentDetail(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleTeacherAssignmentDetail")
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert courseId ID to int: " + r.PathValue("courseId"),
		}
	}
	assignmentIdStr := r.PathValue("assignmentId")
	if assignmentIdStr == "new" {
		return app.renderPage(w, "manage_assignment", map[string]any{
			"Assignment": nil,
			"CourseId":   courseId,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashboard", Href: "/teacher"},
				{Text: "Manage Assignments", Href: "/teacher/courses/" + r.PathValue("courseId") + "/assignments"},
				{Text: "New Assignment", Href: ""},
			},
		})
	}

	assignmentId, err := strconv.Atoi(assignmentIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("assignmentId"),
		}
	}

	assignment, err := app.Store.Assignments.GetById(assignmentId)
	if err != nil {
		log.Error("Could not get assignment by id", slog.String("msg", err.Error()))
		return err
	}

	return app.renderPage(w, "manage_assignment",
		map[string]any{
			"Assignment": assignment,
			"CourseId":   courseId,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashboard", Href: "/teacher"},
				{Text: "Manage Assignments", Href: "/teacher/courses/" + r.PathValue("courseId") + "/assignments"},
				{Text: assignment.Name, Href: ""},
			},
		})
}

func (app *Application) handleTeacherAssignmentCreate(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleTeacherAssignmentCreate")
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert courseId ID to int: " + r.PathValue("courseId"),
		}
	}

	// Read the request body from form values
	type AssignmentBody struct {
		UnitName         string
		Name             string
		Description      string
		RequiredFilename string
		PytestCode       string
		Points           string
		DueDate          string
		Visible          string
	}

	body := &AssignmentBody{
		UnitName:         r.FormValue("unit-name"),
		Name:             r.FormValue("name"),
		Description:      r.FormValue("description"),
		RequiredFilename: r.FormValue("required-filename"),
		PytestCode:       r.FormValue("pyfile-content"),
		Points:           r.FormValue("points"),
		DueDate:          r.FormValue("due-date"),
		Visible:          r.FormValue("visible"),
	}

	if body.UnitName == "" || body.Name == "" || body.Description == "" || body.RequiredFilename == "" || body.Points == "" || body.DueDate == "" || body.PytestCode == "" {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Please make sure to fill out all required fields."})
	}

	pointsInt, err := strconv.Atoi(body.Points)
	if err != nil {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Points was not an int"})
	}

	dueDateInt, err := strconv.ParseInt(body.DueDate, 10, 64)
	if err != nil {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Due date was not an int"})
	}

	assignment := &store.Assignment{
		UnitName:         body.UnitName,
		Name:             body.Name,
		Description:      body.Description,
		RequiredFilename: body.RequiredFilename,
		PytestCode:       body.PytestCode,
		Points:           pointsInt,
		DueDate:          dueDateInt,
		Visible:          body.Visible == "on",
		CourseId:         courseId,
	}

	err = app.Store.Assignments.Create(assignment)
	if err != nil {
		log.Error("Could not create assignment", slog.String("msg", err.Error()))
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/assignments", http.StatusSeeOther)
	return nil
}

func (app *Application) handleTeacherAssignmentUpdate(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleTeacherAssignmentUpdate")
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert courseId ID to int: " + r.PathValue("courseId"),
		}
	}

	// Read the request body from form values
	type AssignmentBody struct {
		Id               string
		UnitName         string
		Name             string
		Description      string
		RequiredFilename string
		PytestCode       string
		Points           string
		DueDate          string
		Visible          string
	}

	body := &AssignmentBody{
		Id:               r.FormValue("id"),
		UnitName:         r.FormValue("unit-name"),
		Name:             r.FormValue("name"),
		Description:      r.FormValue("description"),
		RequiredFilename: r.FormValue("required-filename"),
		PytestCode:       r.FormValue("pyfile-content"),
		Points:           r.FormValue("points"),
		DueDate:          r.FormValue("due-date"),
		Visible:          r.FormValue("visible"),
	}

	if body.Id == "" || body.UnitName == "" || body.Name == "" || body.Description == "" || body.RequiredFilename == "" || body.Points == "" || body.DueDate == "" || body.PytestCode == "" {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Please make sure to fill out all required fields."})
	}

	idInt, err := strconv.Atoi(body.Id)
	if err != nil {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Id was not an int"})
	}
	pointsInt, err := strconv.Atoi(body.Points)
	if err != nil {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Points was not an int"})
	}

	dueDateInt, err := strconv.ParseInt(body.DueDate, 10, 64)
	if err != nil {
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Due date was not an int"})
	}

	assignment := &store.Assignment{
		Id:               idInt,
		UnitName:         body.UnitName,
		Name:             body.Name,
		Description:      body.Description,
		RequiredFilename: body.RequiredFilename,
		PytestCode:       body.PytestCode,
		Points:           pointsInt,
		DueDate:          dueDateInt,
		Visible:          body.Visible == "on",
		CourseId:         courseId,
	}

	err = app.Store.Assignments.Update(assignment)
	if err != nil {
		log.Error("Could not update assignment", slog.String("msg", err.Error()))
		return app.renderPage(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/assignments", http.StatusSeeOther)
	return nil
}

// =====================================
// Student Handlers
// =====================================
func (app *Application) handleStudents(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handleStudents")
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert course ID to int: " + r.PathValue("courseId"),
		}
	}
	students, err := app.Store.Students.GetByCourseId(courseId)
	if err != nil {
		log.Error("Could not get students by course id", slog.String("msg", err.Error()))
		return err
	}
	return app.renderPage(w, "manage_students", map[string]any{
		"Students": students,
		"CourseId": courseId,
		"NavLinks": []NavLink{
			{Text: "Teacher Dashboard", Href: "/teacher"},
			{Text: "Manage students", Href: ""},
		},
	})
}

func (app *Application) handlePasswordReset(w http.ResponseWriter, r *http.Request) error {
	log := app.requestLogger(r, "handlePasswordReset")
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert course ID to int: " + r.PathValue("courseId"),
		}
	}
	// Read the request body from form values
	newPassword := r.FormValue("password")
	if newPassword == "" {
		log.Warn("Incoming password was blank. Nothing happened server side")
		http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/students", http.StatusSeeOther)
	}
	studentIdStr := r.FormValue("student-id")
	studentId, err := strconv.Atoi(studentIdStr)
	if err != nil {
		log.Warn("Incoming student id was not an int. Nothing happened server side")
		http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/students", http.StatusSeeOther)
	}

	err = app.Store.Students.ChangePassword(studentId, newPassword)
	if err != nil {
		log.Error("Could not change password of student "+studentIdStr, slog.String("msg", err.Error()))
		return err
	}
	http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/students", http.StatusSeeOther)
	return nil
}

func (app *Application) handleStudentsExport(w http.ResponseWriter, r *http.Request) error {
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert course ID to int: " + r.PathValue("courseId"),
		}
	}

	studentGrades, err := app.Store.Submissions.GetAllByCourseId(courseId)
	if err != nil {
		return err
	}

	// build csv
	rawCSV := make([][]string, len(studentGrades)+1)
	rawCSV[0] = []string{"Student Number", "Assignment Name", "Grade"}
	for i, student := range studentGrades {
		grade := "NULL"
		if student.Grade.Valid {
			grade = strconv.Itoa(int(student.Grade.Int16))
		}
		rawCSV[i+1] = []string{student.StudentNumber, student.AssignmentName, grade}
	}

	// Set headers to force download
	w.Header().Set("Content-Disposition", `attachment; filename=grades.csv`)
	w.Header().Set("Content-Type", "text/csv")
	exportFile := csv.NewWriter(w)
	exportFile.WriteAll(rawCSV)

	return nil
}
