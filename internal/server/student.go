package server

import (
	"bytes"
	"errors"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
	"github.com/yuin/goldmark"
)

func setLocation(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Location", url)
	} else {
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

func handleStudentDashboardGet(store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		personId := r.Context().Value(personKey).(int)
		role := r.Context().Value(roleKey).(string)

		assignments, err := store.Assignments.GetWithGrade(personId)
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
				"IsTeacher":      role == "teacher",
				"NavLinks": []NavLink{
					{Text: "Dashboard", Href: ""},
				},
			})
	})
}

func handleStudentAssignmentGet(store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId := r.Context().Value(personKey).(int)

		assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		aws, err := store.Assignments.GetWithSubmissionByAssignmentAndStudentIds(assignmentId, studentId)
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

		return tmpl.ExecuteTemplate(w,
			"base",
			map[string]any{
				"Assignment":  &aws.Assignment,
				"Submission":  aws.Submission,
				"Comments":    template.HTML(htmlComments.String()),
				"Description": template.HTML(htmlDescription),
				"NavLinks": []NavLink{
					{Text: "Dashboard", Href: "/app"},
					{Text: aws.Assignment.Name, Href: ""},
				},
			})
	})
}

func handleStudentAssignmentPost(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId := r.Context().Value(personKey).(int)

		assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		// Parse up to 10 MB in memory; rest stored in temp files
		err = r.ParseMultipartForm(10 << 20)
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}
		defer file.Close()

		// Read the file content into a byte slice
		fileContent := make([]byte, handler.Size)
		nbytes, err := file.Read(fileContent)
		if nbytes == 0 {
			return httpe.ServerError(errors.New("Content of "+handler.Filename+" is empty"), http.StatusUnprocessableEntity)
		}
		if err != nil {
			return err
		}

		fileContentStr := strings.TrimSpace(string(fileContent))
		if len(fileContentStr) == 0 {
			return httpe.ServerError(errors.New("Content of "+handler.Filename+" is empty"), http.StatusUnprocessableEntity)
		} 
		// Status is set to 'grading' by deault
		// all 'grading' statuses will be picked up byeeorker started in main function
		err = store.Submissions.Create(assignmentId, studentId, fileContentStr)
		if err != nil {
			return err
		}

		setLocation(w, r, "/app/assignments/"+strconv.Itoa(assignmentId))
		return nil
	})
}

func handleStudentAssignmentPoll(store *storage.Storage, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {

		studentId := r.Context().Value(personKey).(int)

		assignmentId, err := strconv.Atoi(r.PathValue("assignmentId"))
		if err != nil {
			return httpe.ServerError(err, http.StatusBadRequest)
		}

		aws, err := store.Assignments.GetWithSubmissionByAssignmentAndStudentIds(assignmentId, studentId)
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

		return tmpl.ExecuteTemplate(w, "submission_overview", map[string]any{
			"Assignment": aws.Assignment,
			"Submission": aws.Submission,
			"Comments":   template.HTML(htmlComments.String()),
		})
	})
}
