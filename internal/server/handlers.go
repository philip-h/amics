package server

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return nil
	}
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

	return app.renderTemplate(w, "login", nil)
}

func (app *Application) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
	// Read the request body from form values
	body := &LoginReq{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	if body.Username == "" || body.Password == "" {
		return app.renderTemplate(w, "login", map[string]string{"Error": "Username or password field was blank"})
	}

	// Look for student in the database
	student, err := app.Store.Students.GetByUsername(body.Username)
	if err != nil {
		log.Printf("%s: %s", "POST /login", err.Error())
		return app.renderTemplate(w, "login", map[string]string{"Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
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
			log.Printf("%s: %s", "POST /login", err.Error())
			return app.renderTemplate(w, "login", map[string]string{"Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
		}
		if teacher == nil {
			return app.renderTemplate(w, "login", map[string]string{"Error": "Hmm, I could not find your account."})
		}
		user = &UserPass{
			Id:       teacher.Id,
			Username: teacher.Username,
			Password: teacher.Password,
			Role:     "teacher",
		}
	}

	// Right now all of the dummy data in the database does not have hashed passwords, so we will skip the password check for now. In a real implementation, you would want to hash the password when creating the user and compare the hashed password here.
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	// if err != nil {
	if user.Password != body.Password {
		return app.renderTemplate(w, "login", map[string]string{"Error": "Hmm, I could not find your account."})
	}

	// Create JWT token and set it as a cookie
	token, err := app.Auth.CreateJwt(strconv.Itoa(user.Id), user.Role, time.Now().Add(90*time.Minute).Unix())
	if err != nil {
		log.Printf("%s: %s", "POST /login", err.Error())
		return app.renderTemplate(w, "login", map[string]string{"Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
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
	return app.renderTemplate(w, "register", nil)
}

type RegisterReq struct {
	StudentNumber string
	Username      string
	Password      string
	CourseId      string
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
		return app.renderTemplate(w, "register", map[string]string{"Error": "Please make sure to fill out all required fields."})
	}

	// Create a user
	courseId, err := strconv.Atoi(body.CourseId)
	if err != nil {
		return app.renderTemplate(w, "register", map[string]string{"Error": "Invalid course selection."})
	}

	user := &store.Student{
		StudentNumber: body.StudentNumber,
		Username:      body.Username,
		Password:      body.Password,
		CourseId:      courseId,
	}

	err = app.Store.Students.Create(user)
	if err != nil {
		log.Printf("%s: %s", "POST /register", err.Error())
		return app.renderTemplate(w, "register", map[string]string{"Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}
	// Create jwt token and set it as a cookie
	token, err := app.Auth.CreateJwt(user.Username, "student", time.Now().Add(90*time.Minute).Unix())
	if err != nil {
		log.Printf("%s: %s", "POST /register", err.Error())
		return app.renderTemplate(w, "register", map[string]string{"Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
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

	for _, ass := range assignments {
		byUnit[ass.UnitName] = append(byUnit[ass.UnitName], ass)
	}

	return app.renderTemplate(w,
		"student_dashboard",
		map[string]any{
			"Active":      "app",
			"Assignments": byUnit,
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

	return app.renderTemplate(w,
		"assignment_detail",
		map[string]any{
			"Assignment": aws.Assignment,
			"Submission": aws.Submission,
			"Comments":   template.HTML(htmlComments.String()),
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
	return app.renderTemplate(w, "teacher_dashboard", map[string]any{"Courses": courses})
}

func (app *Application) handleCourseCreate(w http.ResponseWriter, r *http.Request) error {
	teacherIdStr := r.Context().Value("userId").(string)
	teacherId, err := strconv.Atoi(teacherIdStr)
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert User ID to int: " + teacherIdStr,
		}
	}

	// Read the request body from form values
	bodyYearStr := r.FormValue("year")
	bodyYear, err := strconv.Atoi(bodyYearStr)
	if err != nil {
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": nil, "Error": "Year was not an int"})
	}

	bodySemStr := r.FormValue("semester")
	bodySem, err := strconv.Atoi(bodySemStr)
	if err != nil {
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": nil, "Error": "Semester was not an int"})
	}

	body := &store.Course{
		Name:      r.FormValue("name"),
		Semester:  bodySem,
		Year:      bodyYear,
		TeacherId: teacherId,
	}
	// If either username or password is empty, return an error
	if body.Name == "" || bodySemStr == "" || bodyYearStr == "" {
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": nil, "Error": "Please make sure to fill out all required fields."})
	}

	// Create a course
	err = app.Store.Courses.Create(body)
	if err != nil {
		log.Printf("%s: %s", "POST /teacher/courses", err.Error())
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": nil, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	return nil
}

func (app *Application) handleCourseUpdate(w http.ResponseWriter, r *http.Request) error {
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
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": body, "Error": "Please make sure to fill out all required fields."})
	}
	yearInt, err := strconv.Atoi(body.Year)
	if err != nil {
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": body, "Error": "Year was not an int"})
	}
	semInt, err := strconv.Atoi(body.Semester)
	if err != nil {
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": body, "Error": "Semester was not an int"})
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
		log.Printf("%s: %s", "PUT /teacher/courses", err.Error())
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": course, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	return nil
}

func (app *Application) handleTeacherCourses(w http.ResponseWriter, r *http.Request) error {

	courseIdStr := r.PathValue("courseId")

	if courseIdStr == "new" {
		return app.renderTemplate(w, "manage_course", map[string]any{"Course": nil, "Error": nil})
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
		return err
	}

	return app.renderTemplate(w, "manage_course", map[string]any{"Course": course})
}

func (app *Application) handleTeacherAssignments(w http.ResponseWriter, r *http.Request) error {

	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert assignment ID to int: " + r.PathValue("courseId"),
		}
	}
	assignments, err := app.Store.Assignments.GetByCourseId(courseId)
	if err != nil {
		return err
	}
	return app.renderTemplate(w, "manage_assignments", map[string]any{"Assignments": assignments, "CourseId": courseId})
}

func (app *Application) handleTeacherAssignmentDetail(w http.ResponseWriter, r *http.Request) error {
	courseId, err := strconv.Atoi(r.PathValue("courseId"))
	if err != nil {
		return &errs.ServerError{
			Status:   http.StatusBadRequest,
			Internal: "Failed to convert courseId ID to int: " + r.PathValue("courseId"),
		}
	}
	assignmentIdStr := r.PathValue("assignmentId")
	if assignmentIdStr == "new" {
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": nil, "CourseId": courseId})
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
		return err
	}

	return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": assignment, "CourseId": courseId})
}

func (app *Application) handleTeacherAssignmentCreate(w http.ResponseWriter, r *http.Request) error {
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
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Please make sure to fill out all required fields."})
	}

	pointsInt, err := strconv.Atoi(body.Points)
	if err != nil {
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Points was not an int"})
	}

	assignment := &store.Assignment{
		UnitName:         body.UnitName,
		Name:             body.Name,
		Description:      body.Description,
		RequiredFilename: body.RequiredFilename,
		PytestCode:       body.PytestCode,
		Points:           pointsInt,
		DueDate:          body.DueDate,
		Visible:          body.Visible == "on",
		CourseId:         courseId,
	}

	err = app.Store.Assignments.Create(assignment)
	if err != nil {
		log.Printf("%s: %s", "POST /teacher/courses/"+strconv.Itoa(courseId)+"/assignments/new", err.Error())
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/assignments", http.StatusSeeOther)
	return nil
}

func (app *Application) handleTeacherAssignmentUpdate(w http.ResponseWriter, r *http.Request) error {
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
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Please make sure to fill out all required fields."})
	}

	idInt, err := strconv.Atoi(body.Id)
	if err != nil {
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Id was not an int"})
	}
	pointsInt, err := strconv.Atoi(body.Points)
	if err != nil {
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Points was not an int"})
	}

	assignment := &store.Assignment{
		Id:               idInt,
		UnitName:         body.UnitName,
		Name:             body.Name,
		Description:      body.Description,
		RequiredFilename: body.RequiredFilename,
		PytestCode:       body.PytestCode,
		Points:           pointsInt,
		DueDate:          body.DueDate,
		Visible:          body.Visible == "on",
		CourseId:         courseId,
	}

	err = app.Store.Assignments.Update(assignment)
	if err != nil {
		log.Printf("%s: %s", "POST /teacher/courses/"+strconv.Itoa(courseId)+"/assignments/"+strconv.Itoa(assignment.Id), err.Error())
		return app.renderTemplate(w, "manage_assignment", map[string]any{"Assignment": body, "Error": "Sorry, something went seriously wrong on our end. Please try again in a sec."})
	}

	http.Redirect(w, r, "/teacher/courses/"+strconv.Itoa(courseId)+"/assignments", http.StatusSeeOther)
	return nil
}
