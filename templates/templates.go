package templates

import (
	"embed"
	"html/template"
)

//go:embed pages/*.html partials/*.html layouts/*.html
var templateFS embed.FS



func LoadTemplates() (map[string]*template.Template, error) {
	pages := map[string][]string{
		// student pages
		"home":       {"layouts/base.html", "pages/home.html"},
		"login":      {"layouts/base.html", "pages/login.html", "partials/login_form_errors.html"},
		"register":   {"layouts/base.html", "pages/register.html", "partials/register_form_errors.html"},
		"app":        {"layouts/base.html", "pages/app.html", "partials/nav.html"},
		"assignment": {"layouts/base.html", "pages/assignment.html", "partials/nav.html", "partials/submission_overview.html", "partials/spinner.html"},

		// teacher pages
		"teacher":            {"layouts/base.html", "pages/teacher.html", "partials/nav.html"},
		"manage_course":      {"layouts/base.html", "pages/manage_course.html", "partials/nav.html"},
		"manage_assignments": {"layouts/base.html", "pages/manage_assignments.html", "partials/nav.html"},
		"manage_assignment":  {"layouts/base.html", "pages/manage_assignment.html", "partials/nav.html"},
		"manage_students":    {"layouts/base.html", "pages/manage_students.html", "partials/nav.html"},
		"manage_student":     {"layouts/base.html", "pages/manage_student.html", "partials/nav.html"},
		"import_grades":      {"layouts/base.html", "pages/import_grades.html", "partials/nav.html"},

		// error page
		"error_page": {"layouts/base.html", "pages/error_page.html"},

		// fragments
		"submission_overview":  {"partials/submission_overview.html", "partials/spinner.html"},
		"login_form_errors":    {"partials/login_form_errors.html"},
		"register_form_errors": {"partials/register_form_errors.html"},
	}

	cache := map[string]*template.Template{}

	for pageName, neededTemplates := range pages {
		tmpl, err := template.New(pageName).
			ParseFS(
				templateFS,
				neededTemplates...,
			)

		if err != nil {
			return nil, err
		}

		cache[pageName] = tmpl
	}
	return cache, nil
}
