package server

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
	"github.com/philip-h/amics/view/model"
	"github.com/philip-h/amics/view/teacher"
)

func handleTeacherDashboardGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		teacherId := r.Context().Value(personKey).(int)

		courses, err := store.Courses.GetByTeacherId(teacherId)
		if err != nil {
			return fmt.Errorf("[teacher.dashboard.get] Could not get courses: %w", err)
		}

		return teacher.Dashboard(courses).Render(r.Context(), w)
	})
}

func handleTeacherDashboardDetailsGet(store *storage.Storage) http.Handler {

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseIdStr := r.PathValue("courseId")

		courseId, err := strconv.Atoi(courseIdStr)
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		assignments, err := store.Assignments.GetByCourseId(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.dashboard.details.get] Could not get assignments: %w", err)
		}

		students, err := store.People.GetByCourseId(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.dashboard.details.get] Could not get students: %w", err)
		}

		// TODO: Think of a way to make this faster
		var submissionCounts map[int]int
		submissionCounts = make(map[int]int)
		for _, assignment := range assignments {
			submissions, err := store.Submissions.ListByAssignmentId(assignment.Id)
			if err != nil {
				return err
			}
			submissionCounts[assignment.Id] = len(submissions)
		}

		return teacher.CourseAssignmentsAndStudents(assignments, submissionCounts, students, courseId).Render(r.Context(), w)
	})
}

func handleTeacherCourseGet(store *storage.Storage) http.Handler {

	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseIdStr := r.PathValue("courseId")

		if courseIdStr == "new" {
			formVm := model.NewCourseFormGet()
			return teacher.ManageCourse(formVm).Render(r.Context(), w)
		}

		courseId, err := strconv.Atoi(courseIdStr)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.course.get] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}
		course, err := store.Courses.GetById(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.course.get] Could not get course: %w", err)
		}

		courseVm := model.NewFromCourse(course)

		return teacher.ManageCourse(courseVm).Render(r.Context(), w)
	})
}

func handleTeacherCoursePost(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		teacherId := r.Context().Value(personKey).(int)

		// Read the request body from form values
		reqBody := model.NewCourseFormPost(
			r.PostFormValue("id"),
			r.PostFormValue("course-code"),
			r.PostFormValue("section"),
			r.PostFormValue("name"),
			r.PostFormValue("year"),
			r.PostFormValue("semester"),
			r.PostFormValue("join-code"),
		)

		if len(reqBody.Validate()) != 0 {
			logger.L.Debug("Validation failed for course form", slog.Any("reqBody", reqBody.Validate()))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return teacher.ManageCourse(reqBody).Render(r.Context(), w)
		}

		// Already validated... I know - parse don't validate. Sue me <3
		semInt, _ := strconv.Atoi(reqBody.Semester)
		yearInt, _ := strconv.Atoi(reqBody.Year)
		sectionInt, _ := strconv.Atoi(reqBody.Section)

		course := &storage.Course{
			CourseCode: reqBody.CourseCode,
			Section:    sectionInt,
			Name:       reqBody.Name,
			Semester:   semInt,
			Year:       yearInt,
			JoinCode:   reqBody.JoinCode,
		}

		err := store.Courses.Create(course, teacherId)
		if err != nil {
			logger.L.Error("[teacher.course.post] Could not create course", slog.String("err", err.Error()))
			reqBody.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return teacher.ManageCourse(reqBody).Render(r.Context(), w)
		}

		w.Header().Set("HX-Redirect", "/teacher")
		w.Header().Set("HX-Push-Url", "/teacher")
		return nil
	})
}

func handleTeacherCoursePut(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := intPathValue(r, "courseId")
		if err != nil {
			return fmt.Errorf("[teacher.course.put] Could not convert course ID to int: %w", err)
		}

		// Read the request body from form values
		reqBody := model.NewCourseFormPost(
			r.PostFormValue("id"),
			r.PostFormValue("course-code"),
			r.PostFormValue("section"),
			r.PostFormValue("name"),
			r.PostFormValue("year"),
			r.PostFormValue("semester"),
			r.PostFormValue("join-code"),
		)

		if len(reqBody.Validate()) != 0 {
			logger.L.Debug("[teacher.course.put] Failed validation of reqBody", slog.Any("reqBody", reqBody.Validate()))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return teacher.ManageCourse(reqBody).Render(r.Context(), w)
		}

		// Already validated... I know - parse don't validate. Sue me <3
		semInt, _ := strconv.Atoi(reqBody.Semester)
		yearInt, _ := strconv.Atoi(reqBody.Year)
		sectionInt, _ := strconv.Atoi(reqBody.Section)

		course := &storage.Course{
			Id:         courseId,
			CourseCode: reqBody.CourseCode,
			Section:    sectionInt,
			Name:       reqBody.Name,
			Semester:   semInt,
			Year:       yearInt,
			JoinCode:   reqBody.JoinCode,
		}

		err = store.Courses.Update(course)
		if err != nil {
			logger.L.Error("[teacher.course.put] Could not update course", slog.String("err", err.Error()))
			reqBody.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return teacher.ManageCourse(reqBody).Render(r.Context(), w)
		}

		w.Header().Set("HX-Redirect", "/teacher")
		w.Header().Set("HX-Push-Url", "/teacher")
		return nil
	})
}

func handleTeacherAssignmentGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := intPathValue(r, "courseId")
		if err != nil {
			return fmt.Errorf("[teacher.assignment.get] Could not convert course ID to int: %w", err)
		}

		assignmentIdStr := r.PathValue("assignmentId")
		unitNames, err := store.Assignments.GetUnitNamesByCourseId(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.assignment.get] Could not get unit names: %w", err)
		}

		if assignmentIdStr == "new" {
			formVm := model.NewAssignmentFormGet()
			formVm.CourseId = strconv.Itoa(courseId)
			return teacher.ManageAssignment(formVm, unitNames).Render(r.Context(), w)
		}

		assignmentId, err := strconv.Atoi(assignmentIdStr)
		if err != nil {
			return fmt.Errorf("[teacher.assignment.get] Could not convert assignment ID to int: %w", err)
		}

		assignment, err := store.Assignments.GetById(assignmentId)
		if err != nil {
			return fmt.Errorf("[teacher.assignment.get] Could not get assignment: %w", err)
		}

		formVm := model.NewFromAssignment(assignment)

		return teacher.ManageAssignment(formVm, unitNames).Render(r.Context(), w)
	})
}

func handleTeacherAssignmentPost(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId := r.PathValue("courseId")
		courseIdInt, err := strconv.Atoi(courseId)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.assignment.post] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}
		unitNames, err := store.Assignments.GetUnitNamesByCourseId(courseIdInt)
		if err != nil {
			return fmt.Errorf("[teacher.assignment.post] Could not get unit names: %w", err)
		}

		reqBody := model.NewAssignmentFormPost(
			r.PostFormValue("id"),
			r.PostFormValue("unit-name"),
			r.PostFormValue("name"),
			r.PostFormValue("description"),
			r.PostFormValue("required-filename"),
			r.PostFormValue("pytest-code"),
			r.PostFormValue("points"),
			r.PostFormValue("due-date"),
			r.PostFormValue("visible") == "on",
			courseId,
		)

		if len(reqBody.Validate()) > 0 {
			logger.L.Debug("[teacher.assignment.post] Error in initial validation", "problems", reqBody.Validate())
			return teacher.ManageAssignment(reqBody, unitNames).Render(r.Context(), w)
		}

		// already validated
		pointsInt, _ := strconv.Atoi(reqBody.Points)

		loc, err := time.LoadLocation("America/Toronto")
		if err != nil {
			return fmt.Errorf("[teacher.assignment.post] Could not load location: %w", err)
		}

		dueDateTime, err := time.ParseInLocation(
			"2006-01-02T15:04",
			reqBody.DueDate,
			loc,
		)
		if err != nil {
			logger.L.Error("[teacher.assignment.post] Could not convert clientside datetime-local to time.Time", "err", err)
			reqBody.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return teacher.ManageAssignment(reqBody, unitNames).Render(r.Context(), w)
		}

		assignment := &storage.Assignment{
			UnitName:         reqBody.UnitName,
			Name:             reqBody.Name,
			Description:      reqBody.Description,
			RequiredFilename: reqBody.RequiredFilename,
			PytestCode:       reqBody.PytestCode,
			Points:           pointsInt,
			DueDate:          dueDateTime,
			Visible:          reqBody.Visible,
			CourseId:         courseIdInt,
		}

		err = store.Assignments.Create(assignment)
		if err != nil {
			logger.L.Error("[teacher.assignment.post] Could not create assignment", slog.String("msg", err.Error()))
			reqBody.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return teacher.ManageAssignment(reqBody, unitNames).Render(r.Context(), w)
		}

		redirectUrl := "/teacher"
		w.Header().Set("HX-Redirect", redirectUrl)
		w.Header().Set("HX-Push-Url", redirectUrl)
		return nil
	})
}

func handleTeacherAssignmentPut(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId := r.PathValue("courseId")
		courseIdInt, err := strconv.Atoi(courseId)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.assignment.put] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}

		idInt, err := strconv.Atoi(r.PostFormValue("id"))
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.assignment.put] Could not convert assignment ID to int: %w", err), http.StatusBadRequest)
		}
		unitNames, err := store.Assignments.GetUnitNamesByCourseId(courseIdInt)
		if err != nil {
			return fmt.Errorf("[teacher.assignment.put] Could not get unit names: %w", err)
		}

		reqBody := model.NewAssignmentFormPost(
			r.PostFormValue("id"),
			r.PostFormValue("unit-name"),
			r.PostFormValue("name"),
			r.PostFormValue("description"),
			r.PostFormValue("required-filename"),
			r.PostFormValue("pytest-code"),
			r.PostFormValue("points"),
			r.PostFormValue("due-date"),
			r.PostFormValue("visible") == "on",
			courseId,
		)

		if len(reqBody.Validate()) > 0 {
			logger.L.Error("[teacher.assignment.put] Error in initial validation", "problems", reqBody.Validate())
			return teacher.ManageAssignment(reqBody, unitNames).Render(r.Context(), w)
		}

		// already validated
		pointsInt, _ := strconv.Atoi(reqBody.Points)

		loc, err := time.LoadLocation("America/Toronto")
		if err != nil {
			return fmt.Errorf("[teacher.assignment.put] Could not load location: %w", err)
		}
		dueDateTime, err := time.ParseInLocation(
			"2006-01-02T15:04",
			reqBody.DueDate,
			loc,
		)
		if err != nil {
			logger.L.Error("[teacher.assignment.put] Could not convert clientside datetime-local to time.Time", "err", err)
			reqBody.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return teacher.ManageAssignment(reqBody, unitNames).Render(r.Context(), w)
		}

		assignment := &storage.Assignment{
			Id:               idInt,
			UnitName:         reqBody.UnitName,
			Name:             reqBody.Name,
			Description:      reqBody.Description,
			RequiredFilename: reqBody.RequiredFilename,
			PytestCode:       reqBody.PytestCode,
			Points:           pointsInt,
			DueDate:          dueDateTime,
			Visible:          reqBody.Visible,
			CourseId:         courseIdInt,
		}

		err = store.Assignments.Update(assignment)
		if err != nil {
			logger.L.Error("[teacher.assignment.put] Could not update assignment", slog.String("msg", err.Error()))
			reqBody.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return teacher.ManageAssignment(reqBody, unitNames).Render(r.Context(), w)
		}

		redirectUrl := "/teacher"
		w.Header().Set("HX-Redirect", redirectUrl)
		w.Header().Set("HX-Push-Url", redirectUrl)
		return nil
	})
}

func handleTeacherGradesImportGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseIdStr := r.PathValue("courseId")
		courseId, err := strconv.Atoi(courseIdStr)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.grades.import.get] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}
		assignmentIdStr := r.PathValue("assignmentId")
		assignmentId, err := strconv.Atoi(assignmentIdStr)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.grades.import.get] Could not convert assignment ID to int: %w", err), http.StatusBadRequest)
		}

		students, err := store.People.GetByCourseId(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.get] Could not get students: %w", err)
		}

		submissions, err := store.Submissions.ListByAssignmentId(assignmentId)
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.get] Could not get submissions: %w", err)
		}

		assignment, err := store.Assignments.GetById(assignmentId)
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.get] Could not get assignment: %w", err)
		}

		submissionsByStudentId := make(map[int]*storage.Submission)
		for _, submission := range submissions {
			submissionsByStudentId[submission.StudentId] = submission
		}

		return teacher.ImportGrades(students, submissionsByStudentId, assignmentIdStr, courseIdStr, assignment.Points).Render(r.Context(), w)
	})
}

func handleTeacherGradesImportTemplateGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseIdStr := r.PathValue("courseId")
		courseId, err := strconv.Atoi(courseIdStr)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.grades.import.template.get] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}
		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.template.get] Could not convert assignment ID to int: %w", err)
		}

		students, err := store.People.GetByCourseId(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.template.get] Could not get students: %w", err)
		}

		assignment, err := store.Assignments.GetById(assignmentId)
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.template.get] Could not get assignment: %w", err)
		}

		// build csv
		rawCSV := make([][]string, len(students)+1)
		rawCSV[0] = []string{"student_number", "grade", "comments"}
		for i, student := range students {
			rawCSV[i+1] = []string{strconv.Itoa(student.Id), "", ""}
		}

		// Set headers to force download
		filename := "" + strings.ToLower(strings.ReplaceAll(assignment.Name, " ", "_")) + "_grades_template.csv"
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", "text/csv")
		exportFile := csv.NewWriter(w)
		exportFile.WriteAll(rawCSV)
		return nil
	})
}

func handleTeacherGradesImportPost(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		assignmentId := r.PathValue("assignmentId")

		// I'm excpecting multiple values per key
		r.ParseForm()
		studentIds := r.PostForm["student-id"]
		grades := r.PostForm["grade"]
		comments := r.PostForm["comments"]

		submissions := make([]*storage.SubmissionImport, len(studentIds))
		for i, studentId := range studentIds {
			gradeInt, err := strconv.Atoi(grades[i])
			if err != nil {
				return httpe.ServerError(fmt.Errorf("[teacher.grades.import.post] Could not convert grade to int: %w", err), http.StatusBadRequest)
			}
			submissions[i] = &storage.SubmissionImport{
				StudentNumber: studentId,
				AssignmentId:  assignmentId,
				Grade:         sql.NullInt16{Int16: int16(gradeInt), Valid: true},
				Comments:      sql.NullString{String: comments[i], Valid: true},
			}
		}

		err := store.Submissions.UpdateAll(submissions)
		if err != nil {
			return fmt.Errorf("[teacher.grades.import.post] Could not update submissions: %w", err)
		}

		w.Header().Set("HX-Redirect", "/teacher")

		return nil
	})
}

func handleTeacherCourseGradesExport(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := strconv.Atoi(r.PathValue("courseId"))
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.course.grades.export] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}

		studentGrades, err := store.Submissions.ListByCourseId(courseId)
		if err != nil {
			return fmt.Errorf("[teacher.course.grades.export] Could not get student grades: %w", err)
		}

		// build csv
		rawCSV := make([][]string, len(studentGrades)+1)
		rawCSV[0] = []string{"Student Number", "First Name", "Assignment Name", "Grade"}
		for i, student := range studentGrades {
			grade := "NULL"
			if student.Grade.Valid {
				grade = strconv.Itoa(int(student.Grade.Int16))
			}
			rawCSV[i+1] = []string{student.StudentNumber, student.FirstName, student.AssignmentName, grade}
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

func handleTeacherCodeExport(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		_, err := intPathValue(r, "courseId")
		if err != nil {
			return fmt.Errorf("[teacher.course.code.export] Could not convert course ID to int: %w", err)
		}

		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[teacher.course.code.export] Could not convert assignment ID to int: %w", err)
		}

		submissions, err := store.Submissions.ListByAssignmentId(assignmentId)
		if err != nil {
			return fmt.Errorf("[teacher.course.code.export] Could not get submissions: %w", err)
		}
		assignment, err := store.Assignments.GetById(assignmentId)
		if err != nil {
			return fmt.Errorf("[teacher.course.code.export] Could not get assignment: %w", err)
		}

		// prep the zip archive in w
		zipWriter := zip.NewWriter(w)

		for _, submission := range submissions {
			student_number := strconv.Itoa(submission.StudentId)
			fileName := student_number + "_" + assignment.RequiredFilename
			f, err := zipWriter.Create(fileName)
			if err != nil {
				return fmt.Errorf("[teacher.course.code.export] Could not create file in zip: %w", err)
			}
			if _, err := f.Write([]byte(submission.Code)); err != nil {
				return fmt.Errorf("[teacher.course.code.export] Could not write to file in zip: %w", err)
			}
		}

		// send it to the client
		zipName := assignment.Name + "_code.zip"
		w.Header().Set("Content-Disposition", "attachment; filename="+zipName)
		w.Header().Set("Content-Type", "application/zip")
		if err := zipWriter.Close(); err != nil {
			return err
		}

		return nil
	})

}

func handleTeacherStudentPasswordReset(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		// Read the request body from form values
		newPassword := r.PostFormValue("new-password")
		if newPassword == "" {
			return httpe.ServerError(errors.New("[teacher.student.password.reset] Password cannot be blank"), http.StatusUnprocessableEntity)
		}
		studentIdStr := r.FormValue("student-id")
		studentId, err := strconv.Atoi(studentIdStr)
		if err != nil {
			return httpe.ServerError(errors.New("[teacher.student.password.reset] Could not convert student ID to int"), http.StatusUnprocessableEntity)
		}

		err = store.People.ChangePassword(studentId, newPassword)
		if err != nil {
			return fmt.Errorf("[teacher.student.password.reset] Could not change password: %w", err)

		}

		w.Header().Set("HX-Redirect", "/teacher/")
		return nil

	})
}

func handleTeacherStudentGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		courseId, err := intPathValue(r, "courseId")
		if err != nil {
			return fmt.Errorf("[teacher.student.get] Could not convert course ID to int: %w", err)
		}
		studentId, err := intPathValue(r, "studentId")
		if err != nil {
			return fmt.Errorf("[teacher.student.get] Could not convert student ID to int: %w", err)
		}

		student, err := store.People.GetById(studentId)
		if err != nil {
			return fmt.Errorf("[teacher.student.get] Could not get student: %w", err)
		}
		if student == nil {
			return httpe.ServerError(errors.New("[teacher.student.get] Student not found"), http.StatusNotFound)
		}

		assignments, err := store.Assignments.GetWithGrade(studentId, courseId)
		if err != nil {
			return fmt.Errorf("[teacher.student.get] Could not get assignments: %w", err)
		}

		// calculate student average
		var studentAvgNum float64
		var studentAvgDenom float64

		for _, ass := range assignments {
			if ass.Assignment.Visible {
				if ass.Grade.Valid {
					studentAvgNum += float64(ass.Grade.Int64)
				}
				studentAvgDenom += float64(ass.Points)
			}
		}
		var studentAverage float64
		if studentAvgDenom == 0 {
			studentAverage = 0
		} else {
			studentAverage = math.Round((studentAvgNum / studentAvgDenom) * 100)
		}

		return teacher.ManageStudent(model.NewStudentData(
			student,
			assignments,
			studentAverage,
		), strconv.Itoa(courseId)).Render(r.Context(), w)
	})
}

func handleTeacherStudentSubmissionGet(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId, err := intPathValue(r, "studentId")
		if err != nil {
			return fmt.Errorf("[teacher.student.submission.get] Could not convert student ID to int: %w", err)
		}
		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[teacher.student.submission.get] Could not convert assignment ID to int: %w", err)
		}

		submission, err := store.Submissions.GetByAssignmentAndStudentIds(assignmentId, studentId)
		if err != nil {
			return fmt.Errorf("[teacher.student.submission.get] Could not get submission: %w", err)
		}

		return teacher.StudentSubmissionDetails(submission).Render(r.Context(), w)
	})
}
