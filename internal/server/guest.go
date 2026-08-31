package server

import (
	"net/http"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/view/model"
	"github.com/philip-h/amics/view/guest"
)

func handleGuestIndexGet() http.Handler {
	return httpe.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) error {
			if r.URL.Path != "/" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return nil
			}

			return guest.Home(model.NewLoginFormGet()).Render(r.Context(), w)
		},
	)
}
