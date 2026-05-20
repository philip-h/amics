package server

import (
	"html/template"
	"net/http"

	"github.com/philip-h/amics/internal/storage"
)

func addRoutes(
	mux *http.ServeMux,
	logger *Logger,
	store *storage.Storage,
	templates map[string]*template.Template,
) {
	// Serve static files
	mux.Handle("GET /static/{path...}", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// guest routes
	mux.Handle("GET /",
		handleGuestIndexGet(logger, templates["home"]))

	mux.Handle("GET /login",
		handleAuthLoginGet(templates["login"]))
	mux.Handle("POST /login",
		handleAuthLoginPost(logger, store, templates["login"]))
	mux.Handle("POST /login/validate",
		handleAuthLoginValidation(templates["login_form_errors"]))
	mux.Handle("POST /logout",
		handleAuthLogout(logger, store))

	mux.Handle("GET /register",
		handleAuthRegisterGet(templates["register"]))
	mux.Handle("POST /register",
		handleAuthRegisterPost(logger, store, templates["register"]))
	mux.Handle("POST /register/validate",
		handleAuthRegisterValidation(store, templates["register_form_errors"]))

	// student routes
	mux.Handle("GET /app",
		requiresStudent(
			handleStudentDashboardGet(store, templates["app"]),
		))
	mux.Handle("GET /app/assignments/{assignmentId}",
		requiresStudent(
			handleStudentAssignmentGet(store, templates["assignment"]),
		))
	mux.Handle("POST /app/assignments/{assignmentId}",
		requiresStudent(
			handleStudentAssignmentPost(store),
		))
	mux.Handle("GET /app/assignments/{assignmentId}/poll",
		requiresStudent(
			handleStudentAssignmentPoll(store, templates["submission_overview"]),
		))

	// teacher routes

	// teacher courses
	mux.Handle("GET /teacher",
		requiresTeacher(
			handleTeacherDashboardGet(store, templates["teacher"]),
		))
	mux.Handle("GET /teacher/courses/{courseId}",
		requiresTeacher(
			handleTeacherCourseGet(logger, store, templates["manage_course"]),
		))
	mux.Handle("POST /teacher/courses",
		requiresTeacher(
			handleTeacherCoursePost(logger, store, templates["manage_course"]),
		))
	mux.Handle("PUT /teacher/courses/{courseId}",
		requiresTeacher(
			handleTeacherCoursePut(logger, store, templates["manage_course"]),
		))

	// teacher assignments
	mux.Handle("GET /teacher/courses/{courseId}/assignments",
		requiresTeacher(
			handleTeacherAssignmentsGet(logger, store, templates["manage_assignments"]),
		))
	mux.Handle("GET /teacher/courses/{courseId}/assignments/{assignmentId}",
		requiresTeacher(
			handleTeacherAssignmentGet(logger, store, templates["manage_assignment"]),
		))
	mux.Handle("POST /teacher/courses/{courseId}/assignments",
		requiresTeacher(
			handleTeacherAssignmentPost(logger, store, templates["manage_assignment"]),
		))
	mux.Handle("PUT /teacher/courses/{courseId}/assignments/{assignmentId}",
		requiresTeacher(
			handleTeacherAssignmentPut(logger, store, templates["manage_assignment"]),
		))

	// teacher import / export
	mux.Handle("GET /teacher/courses/{courseId}/assignments/{assignmentId}/import",
		requiresTeacher(
			handleTeacherGradesImportGet(templates["import_grades"]),
		))
	mux.Handle("POST /teacher/courses/{courseId}/assignments/{assignmentId}/import",
		requiresTeacher(
			handleTeacherGradesImportPost(logger, store, templates["import_grades"]),
		))
	mux.Handle("GET /teacher/courses/{courseId}/export",
		requiresTeacher(
			handleTeacherCourseGradesExport(logger, store),
		))
	mux.Handle("GET /teacher/courses/{courseId}/assignments/{assignmentId}/export_code",
		requiresTeacher(
			handleTeacherCodeExport(store),
		))

	// teacher student management
	mux.Handle("GET /teacher/courses/{courseId}/students",
		requiresTeacher(
			handleTeacherStudentsGet(store, templates["manage_students"]),
		))
	mux.Handle("POST /teacher/courses/{courseId}/students/{studentId}/passwordreset",
		requiresTeacher(
			handleTeacherStudentPasswordReset(logger, store),
		))

	mux.Handle("GET /teacher/courses/{courseId}/students/{studentId}",
		requiresTeacher(
			handleTeacherStudentGet(logger, store, templates["manage_student"]),
		))
}
