package server

import (
	"encoding/csv"
	"errors"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
)

func handleTeacherDashboardGet(store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		teacherId := r.Context().Value(personKey).(int)

		courses, err := store.Courses.GetByTeacherId(teacherId)
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(w, "base", map[string]any{
			"Courses": courses,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashboard", Href: ""},
			},
		})
	})
}

func handleTeacherCourseGet(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseIdStr := r.PathValue("courseId")
		problems := make(map[string]string)

		if courseIdStr == "new" {
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Course":   nil,
				"Problems": problems,
				"NavLinks": []NavLink{
					{Text: "Teacher Dashbord", Href: "/teacher"},
					{Text: "New Course", Href: ""},
				},
			})
		}

		courseId, err := strconv.Atoi(courseIdStr)
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}
		course, err := store.Courses.GetById(courseId)
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(w, "base", map[string]any{
			"Course":   course,
			"Problems": problems,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashbord", Href: "/teacher"},
				{Text: course.Name, Href: ""},
			},
		})
	})
}

func handleTeacherCoursePost(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {

	type request struct {
		Id       string
		JoinCode string
		Year     string
		Semester string
		Name     string
	}

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		teacherId := r.Context().Value(personKey).(int)

		// Read the request body from form values
		body := &request{
			JoinCode: r.PostFormValue("join-code"),
			Year:     r.PostFormValue("year"),
			Semester: r.PostFormValue("semester"),
			Name:     r.PostFormValue("name"),
		}

		problems := validateCourseReq(
			"new",
			body.JoinCode,
			body.Year,
			body.Semester,
			body.Name,
		)

		if len(problems) > 0 {
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Course":   nil,
				"Body":     body,
				"Problems": problems,
			})
		}

		// Already validated... I know - parse don't validate. Sue me <3
		semInt, _ := strconv.Atoi(body.Semester)
		yearInt, _ := strconv.Atoi(body.Year)

		course := &storage.Course{
			JoinCode: body.JoinCode,
			Year:     yearInt,
			Semester: semInt,
			Name:     body.Name,
		}

		err := store.Courses.Create(course, teacherId)
		if err != nil {
			logger.L.Error("Could not create course", slog.String("err", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Course":   course,
				"Problems": problems,
			})
		}

		w.Header().Set("HX-Redirect", "/teacher")
		w.Header().Set("HX-Push-Url", "/teacher")
		return nil
	})
}

func handleTeacherCoursePut(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {

	type request struct {
		Id       string
		JoinCode string
		Year     string
		Semester string
		Name     string
	}

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := strconv.Atoi(r.PathValue("courseId"))

		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		// Read the request body from form values
		body := &request{
			Id:       r.PostFormValue("id"),
			JoinCode: r.PostFormValue("join-code"),
			Year:     r.PostFormValue("year"),
			Semester: r.PostFormValue("semester"),
			Name:     r.PostFormValue("name"),
		}

		problems := validateCourseReq(
			body.Id,
			body.JoinCode,
			body.Year,
			body.Semester,
			body.Name,
		)

		if len(problems) > 0 {
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Course": body, "Body": body, "Problems": problems})
		}

		// Already validated... I know - parse don't validate. Sue me <3
		semInt, _ := strconv.Atoi(body.Semester)
		yearInt, _ := strconv.Atoi(body.Year)

		course := &storage.Course{
			Id:       courseId,
			JoinCode: body.JoinCode,
			Year:     yearInt,
			Semester: semInt,
			Name:     body.Name,
		}

		err = store.Courses.Update(course)
		if err != nil {
			logger.L.Error("Could not update course", slog.String("err", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "base", map[string]any{"Course": course, "Body": body, "Problems": problems})
		}

		w.Header().Set("HX-Redirect", "/teacher")
		w.Header().Set("HX-Push-Url", "/teacher")
		return nil
	})
}

func handleTeacherAssignmentsGet(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		assignments, err := store.Assignments.GetByCourseId(courseId)
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(w, "base", map[string]any{
			"Assignments": assignments,
			"CourseId":    courseId,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashboard", Href: "/teacher"},
				{Text: "Manage Assignments", Href: ""},
			},
		})
	})
}

func handleTeacherAssignmentGet(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		assignmentIdStr := r.PathValue("assignmentId")

		if assignmentIdStr == "new" {
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"CourseId": courseId,
				"NavLinks": []NavLink{
					{Text: "Teacher Dashboard", Href: "/teacher"},
					{Text: "Manage Assignments", Href: "/teacher/courses/" + r.PathValue("courseId") + "/assignments"},
					{Text: "New Assignment", Href: ""},
				},
			})
		}

		assignmentId, err := strconv.Atoi(assignmentIdStr)
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		assignment, err := store.Assignments.GetById(assignmentId)
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(w, "base",
			map[string]any{
				"Assignment": assignment,
				"CourseId":   courseId,
				"NavLinks": []NavLink{
					{Text: "Teacher Dashboard", Href: "/teacher"},
					{Text: "Manage Assignments", Href: "/teacher/courses/" + r.PathValue("courseId") + "/assignments"},
					{Text: assignment.Name, Href: ""},
				},
			})
	})
}

func handleTeacherAssignmentPost(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	type request struct {
		UnitName         string
		Name             string
		Description      string
		RequiredFilename string
		PytestCode       string
		Points           string
		DueDate          string
		Visible          string
	}

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		body := &request{
			UnitName:         r.PostFormValue("unit-name"),
			Name:             r.PostFormValue("name"),
			Description:      r.PostFormValue("description"),
			RequiredFilename: r.PostFormValue("required-filename"),
			PytestCode:       r.PostFormValue("pyfile-content"),
			Points:           r.PostFormValue("points"),
			DueDate:          r.PostFormValue("due-date"),
			Visible:          r.PostFormValue("visible"),
		}

		problems := validateAssignmentReq(body.UnitName,
			body.Name,
			body.Description,
			body.RequiredFilename,
			body.PytestCode,
			body.Points,
			body.DueDate,
		)

		if len(problems) > 0 {
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Body":     body,
				"CourseId": courseId,
				"Problems": problems,
			})
		}

		// already validated
		pointsInt, _ := strconv.Atoi(body.Points)

		loc, err := time.LoadLocation("America/Toronto")
		if err != nil {
			return err
		}

		dueDateTime, err := time.ParseInLocation(
			"2006-01-02T15:04",
			body.DueDate,
			loc,
		)
		if err != nil {
			logger.L.Error("Could not convert clientside datetime-local to time.Time", "err", err)
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Body":     body,
				"CourseId": courseId,
				"Problems": problems,
			})
		}

		assignment := &storage.Assignment{
			UnitName:         body.UnitName,
			Name:             body.Name,
			Description:      body.Description,
			RequiredFilename: body.RequiredFilename,
			PytestCode:       body.PytestCode,
			Points:           pointsInt,
			DueDate:          dueDateTime,
			Visible:          body.Visible == "on",
			CourseId:         courseId,
		}

		err = store.Assignments.Create(assignment)
		if err != nil {
			logger.L.Error("Could not create assignment", slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Body":     body,
				"CourseId": courseId,
				"Problems": problems,
			})
		}

		redirectUrl := "/teacher/courses/" + strconv.Itoa(courseId) + "/assignments"
		w.Header().Set("HX-Redirect", redirectUrl)
		w.Header().Set("HX-Push-Url", redirectUrl)
		return nil
	})
}

func handleTeacherAssignmentPut(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	type request struct {
		Id               int
		UnitName         string
		Name             string
		Description      string
		RequiredFilename string
		PytestCode       string
		Points           string
		DueDate          string
		Visible          string
	}

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		idInt, err := strconv.Atoi(r.PostFormValue("id"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)

		}
		body := &request{
			Id:               idInt,
			UnitName:         r.PostFormValue("unit-name"),
			Name:             r.PostFormValue("name"),
			Description:      r.PostFormValue("description"),
			RequiredFilename: r.PostFormValue("required-filename"),
			PytestCode:       r.PostFormValue("pyfile-content"),
			Points:           r.PostFormValue("points"),
			DueDate:          r.PostFormValue("due-date"),
			Visible:          r.PostFormValue("visible"),
		}

		problems := validateAssignmentReq(
			body.UnitName,
			body.Name,
			body.Description,
			body.RequiredFilename,
			body.PytestCode,
			body.Points,
			body.DueDate,
		)

		if len(problems) > 0 {
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Assignment": body,
				"Body":       body,
				"CourseId":   courseId,
				"Problems":   problems,
			})
		}

		// already validated
		pointsInt, _ := strconv.Atoi(body.Points)

		loc, err := time.LoadLocation("America/Toronto")
		if err != nil {
			return err
		}
		dueDateTime, err := time.ParseInLocation(
			"2006-01-02T15:04",
			body.DueDate,
			loc,
		)
		if err != nil {
			logger.L.Error("Could not convert clientside datetime-local to time.Time", "err", err)
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Assignment": body,
				"Body":       body,
				"CourseId":   courseId,
				"Problems":   problems,
			})
		}

		assignment := &storage.Assignment{
			Id:               body.Id,
			UnitName:         body.UnitName,
			Name:             body.Name,
			Description:      body.Description,
			RequiredFilename: body.RequiredFilename,
			PytestCode:       body.PytestCode,
			Points:           pointsInt,
			DueDate:          dueDateTime,
			Visible:          body.Visible == "on",
			CourseId:         courseId,
		}

		err = store.Assignments.Update(assignment)
		if err != nil {
			logger.L.Error("Could not update assignment", slog.String("msg", err.Error()))
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return tmpl.ExecuteTemplate(w, "base", map[string]any{
				"Assignment": body,
				"Body":       body,
				"CourseId":   courseId,
				"Problems":   problems,
			})
		}

		redirectUrl := "/teacher/courses/" + strconv.Itoa(courseId) + "/assignments"
		w.Header().Set("HX-Redirect", redirectUrl)
		w.Header().Set("HX-Push-Url", redirectUrl)
		return nil
	})
}

func handleTeacherGradesImportGet(tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseIdStr := r.PathValue("courseId")
		assignmentIdStr := r.PathValue("assignmentId")

		return tmpl.ExecuteTemplate(w, "base", map[string]any{
			"CourseId":     courseIdStr,
			"AssignmentId": assignmentIdStr,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashboard", Href: "/teacher"},
				{Text: "Import grades", Href: ""},
			},
		})
	})
}

func handleTeacherGradesImportPost(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseIdStr := r.PathValue("courseId")
		assignmentIdStr := r.PathValue("assignmentId")
		assignmentId, err := strconv.Atoi(assignmentIdStr)
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		// Limit upload size to 10MB
		err = r.ParseMultipartForm(10 << 20)
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}
		defer file.Close()

		csvReader := csv.NewReader(file)
		csvContent, err := csvReader.ReadAll()
		if err != nil {
			return err
		}

		err = store.Assignments.UpdateAll(assignmentId, csvContent)
		if err != nil {
			return err
		}

		w.Header().Set("HX-Redirect", "/teacher/courses/"+courseIdStr+"/assignments")
		return nil
	})
}

func handleTeacherCourseGradesExport(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		studentGrades, err := store.Submissions.GetAllByCourseId(courseId)
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
		now := time.Now().Format("01021504")
		filename := "grades_" + now + ".csv"
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", "text/csv")
		exportFile := csv.NewWriter(w)
		exportFile.WriteAll(rawCSV)
		return nil
	})
}

func handleTeacherStudentsGet(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}
		students, err := store.People.GetByCourseId(courseId)
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(w, "base", map[string]any{
			"Students": students,
			"CourseId": courseId,
			"NavLinks": []NavLink{
				{Text: "Teacher Dashboard", Href: "/teacher"},
				{Text: "Manage students", Href: ""},
			},
		})

	})
}

func handleTeacherStudentPasswordReset(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}
		// Read the request body from form values
		newPassword := r.PostFormValue("password")
		if newPassword == "" {
			return httpe.ServerError(errors.New("Password cannot be blank"), http.StatusUnprocessableEntity)
		}
		studentIdStr := r.FormValue("student-id")
		studentId, err := strconv.Atoi(studentIdStr)
		if err != nil {
			return httpe.ServerError(errors.New("Student id is not an int?"), http.StatusUnprocessableEntity)
		}

		err = store.People.ChangePassword(studentId, newPassword)
		if err != nil {
			return err

		}

		w.Header().Set("HX-Redirect", "/teacher/courses/"+strconv.Itoa(courseId)+"/students")
    return nil

	})
}

func handleTeacherStudentGet(logger *Logger, store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId := r.PathValue("courseId")
		studentId, err := strconv.Atoi(r.PathValue("studentId"))
		if err != nil {
      return httpe.ServerError(err, http.StatusBadRequest)
		}

		assignments, err := store.Assignments.GetWithGrade(studentId)
		if err != nil {
      return err
		}

		// Organize assignment by units for nicer rendering
		byUnit := make(map[string][]*storage.AssignmentWithGrade)
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

		return tmpl.ExecuteTemplate(w,
			"base",
			map[string]any{
				"Assignments":    byUnit,
				"StudentAverage": studentAverage,
				"NavLinks": []NavLink{
					{Text: "Teacher Dashboard", Href: "/teacher"},
					{Text: "Manage students", Href: "/teacher/courses/" + courseId + "/students"},
					{Text: "Manage Student", Href: ""},
				},
			})
	})
}
