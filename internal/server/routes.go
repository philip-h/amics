package server

import (
	"net/http"

	"github.com/philip-h/amics/internal/storage"
)

func addRoutes(
	mux *http.ServeMux,
	logger *Logger,
	store *storage.Storage,
) {
	// Serve static files
	mux.Handle("GET /static/{path...}", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// guest routes
	mux.Handle("GET /",
		handleGuestIndexGet())

	mux.Handle("GET /login",
		handleAuthLoginGet())
	mux.Handle("POST /login",
		handleAuthLoginPost(logger, store))
	mux.Handle("POST /login/validate",
		handleAuthLoginValidation())
	mux.Handle("POST /logout",
		handleAuthLogout(logger, store))

	mux.Handle("GET /register",
		handleAuthRegisterGet())
	mux.Handle("POST /register",
		handleAuthRegisterPost(logger, store))
	mux.Handle("POST /register/validate",
		handleAuthRegisterValidation(store))

	// student routes
	mux.Handle("GET /h",
		requiresStudent(
			handleStudentCoursesGet(store),
		))
	mux.Handle("GET /c/{courseId}",
		requiresStudent(
			handleStudentDashboardGet(store),
		))
	mux.Handle("GET /c/{courseId}/a/{assignmentId}/details",
		requiresStudent(
			handleStudentAssignmentGet(store),
		))
	mux.Handle("POST /c/{courseId}/a/{assignmentId}",
		requiresStudent(
			handleStudentAssignmentPost(store),
		))
	mux.Handle("GET /c/{courseId}/a/{assignmentId}/poll",
		requiresStudent(
			handleStudentAssignmentPoll(store),
		))
	mux.Handle("GET /c/{courseId}/a/{assignmentId}/code",
		requiresStudent(
			handleViewSubmissionCode(store),
		))

	// teacher routes

	// teacher courses
	mux.Handle("GET /teacher",
		requiresTeacher(
			handleTeacherDashboardGet(store),
		))
	mux.Handle("GET /teacher/courses/{courseId}",
		requiresTeacher(
			handleTeacherCourseGet(store),
		))
	mux.Handle("GET /teacher/courses/{courseId}/details",
		requiresTeacher(
			handleTeacherDashboardDetailsGet(store),
		))
	mux.Handle("POST /teacher/courses",
		requiresTeacher(
			handleTeacherCoursePost(logger, store),
		))
	mux.Handle("PUT /teacher/courses/{courseId}",
		requiresTeacher(
			handleTeacherCoursePut(logger, store),
		))

	// teacher assignments
	mux.Handle("GET /teacher/courses/{courseId}/assignments/{assignmentId}",
		requiresTeacher(
			handleTeacherAssignmentGet(store),
		))
	mux.Handle("POST /teacher/courses/{courseId}/assignments",
		requiresTeacher(
			handleTeacherAssignmentPost(logger, store),
		))
	mux.Handle("PUT /teacher/courses/{courseId}/assignments/{assignmentId}",
		requiresTeacher(
			handleTeacherAssignmentPut(logger, store),
		))

	// teacher import / export
	mux.Handle("GET /teacher/courses/{courseId}/assignments/{assignmentId}/import",
		requiresTeacher(
			handleTeacherGradesImportGet(store),
		))
	mux.Handle("GET /teacher/courses/{courseId}/assignments/{assignmentId}/import/template",
		requiresTeacher(
			handleTeacherGradesImportTemplateGet(store),
		))
	mux.Handle("POST /teacher/courses/{courseId}/assignments/{assignmentId}/import",
		requiresTeacher(
			handleTeacherGradesImportPost(logger, store),
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
	mux.Handle("POST /teacher/courses/{courseId}/students/{studentId}/passwordreset",
		requiresTeacher(
			handleTeacherStudentPasswordReset(logger, store),
		))

	mux.Handle("GET /teacher/courses/{courseId}/students/{studentId}",
		requiresTeacher(
			handleTeacherStudentGet(store),
		))
	mux.Handle("GET /teacher/courses/{courseId}/students/{studentId}/submissions/{assignmentId}",
		requiresTeacher(
			handleTeacherStudentSubmissionGet(logger, store),
		))
}
