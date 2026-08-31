package server

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
	"github.com/philip-h/amics/view/student"
	"github.com/yuin/goldmark"
)

func setLocation(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Location", url)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func handleStudentCoursesGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		personId := r.Context().Value(personKey).(int)
		personIdHex := strconv.FormatInt(int64(personId), 16)
		role := r.Context().Value(roleKey).(string)

		var courses []*storage.Course
		var err error
		if role == "teacher" {
			courses, err = store.Courses.GetByTeacherId(personId)
			if err != nil {
				return fmt.Errorf("[student.courses.get] Could not get courses by teacher id: %w", err)
			}
		} else {
			courses, err = store.Courses.GetByStudentId(personId)
			if err != nil {
				return fmt.Errorf("[student.courses.get] Could not get courses by student id: %w", err)
			}
		}

		if len(courses) == 1 {
			setLocation(w, r, "/c/"+strconv.Itoa(courses[0].Id))
			return nil
		}

		return student.Courses(courses, personIdHex, role == "teacher").Render(r.Context(), w)
	})
}
func handleStudentDashboardGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		personId := r.Context().Value(personKey).(int)
		personIdHex := strconv.FormatInt(int64(personId), 16)
		role := r.Context().Value(roleKey).(string)

		courseId := r.PathValue("courseId")
		courseIdInt, err := strconv.Atoi(courseId)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[student.dashboard.get] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}

		assignments, err := store.Assignments.GetWithGrade(personId, courseIdInt)
		if err != nil {
			return fmt.Errorf("[student.dashboard.get] Could not get assignments: %w", err)
		}

		// Organize assignment by units for nicer rendering
		byUnit := make(map[string][]*storage.AssignmentWithGrade)
		// maps are not ordered, so this keys slice will preserve the order of the sql query
		keys := []string{}
		// And also calculate student average
		var studentAvgNum float64
		var studentAvgDenom float64

		for _, ass := range assignments {
			if ass.Visible {
				byUnit[ass.UnitName] = append(byUnit[ass.UnitName], ass)
				if !slices.Contains(keys, ass.UnitName) {
					keys = append(keys, ass.UnitName)
				}
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

		return student.Dashboard(
			byUnit,
			keys,
			courseId,
			studentAverage,
			personIdHex,
			role == "teacher",
		).Render(r.Context(), w)
	})
}

func handleStudentAssignmentGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId := r.Context().Value(personKey).(int)
		studentIdHex := strconv.FormatInt(int64(studentId), 16)
		courseId := r.PathValue("courseId")

		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[student.assignment.get] Could not convert assignment ID to int: %w", err)
		}

		aws, err := store.Assignments.GetWithSubmissionByAssignmentAndStudentIds(assignmentId, studentId)
		if err != nil {
			return fmt.Errorf("[student.assignment.get] Could not get assignment: %w", err)
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
					htmlComments.WriteString("<span class='text-red-700'>")
					htmlComments.WriteString(line)
					htmlComments.WriteString("</span>")
				} else if strings.HasPrefix(line, "✔") {
					htmlComments.WriteString("<span class='text-green-700'>")
					htmlComments.WriteString(line)
					htmlComments.WriteString("</span>")
				} else {
					htmlComments.WriteString(line)
				}
				htmlComments.WriteString("\n")
			}
		}

		return student.Assignment(
			&aws.Assignment,
			aws.Submission,
			htmlComments.String(),
			htmlDescription,
			studentIdHex,
			courseId,
		).Render(r.Context(), w)
	})
}

func handleStudentAssignmentPost(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId := r.Context().Value(personKey).(int)

		courseId, err := intPathValue(r, "courseId")
		if err != nil {
			return fmt.Errorf("[student.assignment.post] Could not convert course ID to int: %w", err)
		}
		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[student.assignment.post] Could not convert assignment ID to int: %w", err)
		}

		// Parse up to 10 MB in memory; rest stored in temp files
		err = r.ParseMultipartForm(10 << 20)
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[student.assignment.post] Could not parse multipart form: %w", err), http.StatusBadRequest)
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[student.assignment.post] Could not get file from form: %w", err), http.StatusBadRequest)
		}
		defer file.Close()

		// Read the file content into a byte slice
		fileContent := make([]byte, handler.Size)
		nbytes, err := file.Read(fileContent)
		if nbytes == 0 {
			return httpe.ServerError(errors.New("[student.assignment.post] Content of "+handler.Filename+" is empty"), http.StatusUnprocessableEntity)
		}
		if err != nil {
			return fmt.Errorf("[student.assignment.post] Could not read file content: %w", err)
		}

		fileContentStr := strings.TrimSpace(string(fileContent))
		if len(fileContentStr) == 0 {
			return httpe.ServerError(errors.New("[student.assignment.post] Content of "+handler.Filename+" is empty"), http.StatusUnprocessableEntity)
		}
		// Status is set to 'grading' by deault
		// all 'grading' statuses will be picked up byeeorker started in main function
		err = store.Submissions.Create(assignmentId, studentId, fileContentStr)
		if err != nil {
			return fmt.Errorf("[student.assignment.post] Could not create submission: %w", err)
		}

		setLocation(w, r, "/c/"+strconv.Itoa(courseId)+"/a/"+strconv.Itoa(assignmentId)+"/details")
		return nil
	})
}

func handleStudentAssignmentPoll(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId := r.Context().Value(personKey).(int)
		courseId := r.PathValue("courseId")
		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[student.assignment.poll] Could not convert assignment ID to int: %w", err)
		}

		aws, err := store.Assignments.GetWithSubmissionByAssignmentAndStudentIds(assignmentId, studentId)
		if err != nil {
			return fmt.Errorf("[student.assignment.poll] Could not get assignment with submission: %w", err)
		}
		var htmlComments strings.Builder
		if aws.Submission != nil && aws.Submission.Comments.Valid {

			lines := strings.SplitSeq(strings.ReplaceAll(aws.Submission.Comments.String, "\r\n", "\n"), "\n")

			for line := range lines {
				if strings.HasPrefix(line, "E") || strings.HasPrefix(line, ">") {
					htmlComments.WriteString("<span class='text-red-700'")
					htmlComments.WriteString(">")
					htmlComments.WriteString(line)
					htmlComments.WriteString("</span>")
				} else {
					htmlComments.WriteString(line)
				}
				htmlComments.WriteString("\n")
			}
		}

		return templ.RenderFragments(r.Context(), w,
			student.Assignment(
				&aws.Assignment,
				aws.Submission,
				htmlComments.String(),
				"",
				"",
				courseId,
			),
			"submission-overview",
		)
	})
}

func handleViewSubmissionCode(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		studentId := r.Context().Value(personKey).(int)

		assignmentId, err := intPathValue(r, "assignmentId")
		if err != nil {
			return fmt.Errorf("[student.viewcode] Could not convert assignment ID to int: %w", err)
		}

		submission, err := store.Submissions.GetByAssignmentAndStudentIds(assignmentId, studentId)
		if err != nil {
			return fmt.Errorf("[student.viewcode] Could not get submission: %w", err)
		}

		if submission == nil {
			return httpe.ServerError(errors.New("[student.viewcode] Submission not found"), http.StatusNotFound)
		}

		// Write code to response
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(submission.Code))
		if err != nil {
			return fmt.Errorf("[student.viewcode] Could not write submission code to response: %w", err)
		}

		return nil
	})
}
