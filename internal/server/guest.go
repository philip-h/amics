package server

import (
	"html/template"
	"net/http"

	"github.com/philip-h/amics/internal/httpe"
)

func handleGuestIndexGet(logger *Logger, tmpl *template.Template) http.Handler {
	return httpe.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) error {
			if r.URL.Path != "/" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return nil
			}
			return tmpl.ExecuteTemplate(w, "base", nil)
		},
	)
}
