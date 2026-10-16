package server

import (
	"net/http"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/view/guest"
	"github.com/philip-h/amics/view/model"
)

func handleGuestIndexGet() http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path != "/" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return nil
		}

		// Check to see if the user is already logged in, if so redirect to home page
		personId := r.Context().Value(personKey)
		if personId != nil {
			redirect(w, r, "/h")
			return nil
		}

		return guest.Home(model.NewLoginFormGet()).Render(r.Context(), w)
	})
}
