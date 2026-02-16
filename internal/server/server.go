package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/store"
)

type Application struct {
	Config Config
	Store store.Storage
}

type Config struct {
	Port string
	Db *db.DbConfig
}

func (app *Application) Mount() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
    fs := http.FileServer(http.Dir("./static"))
    mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Homepage
	mux.HandleFunc("GET /", makeHTTPHandlerFunc(app.handleIndex))

	// Auth handlers
	mux.HandleFunc("GET /login", makeHTTPHandlerFunc(app.handleLoginGet))
	mux.HandleFunc("POST /login", makeHTTPHandlerFunc(app.handleLoginPost))
	mux.HandleFunc("GET /register", makeHTTPHandlerFunc(app.handleRegisterGet))
	mux.HandleFunc("POST /register", makeHTTPHandlerFunc(app.handleRegisterPost))
	mux.HandleFunc("POST /logout", makeHTTPHandlerFunc(app.handleLogout))

	// Dashboard
	mux.HandleFunc("GET /app", makeHTTPHandlerFunc(app.handleDashboard))

	return mux
}

func (app *Application) Run(mux *http.ServeMux) error {
	server := &http.Server{
		Addr:    app.Config.Port,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func writeJSON(w http.ResponseWriter, status int, resource any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&resource)
}

func makeHTTPHandlerFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			je := &JsonError{}
			se := &ServerError{}
			if errors.As(err, &je) {
				log.Printf("%s: %s", r.URL.Path, je.Internal)
				writeJSON(w, je.Status, je)
			} else if errors.As(err, &se) {
				log.Printf("%s: %s", r.URL.Path, se.Internal)
				http.Error(w, se.Error(), se.Status)
			} else {
				log.Printf("%s: %s", r.URL.Path, err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	}
}
