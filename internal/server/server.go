package server

import (
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
	mux.HandleFunc("GET /", app.handleIndex)

	// Auth handlers
	mux.HandleFunc("GET /login", app.handleLoginGet)
	mux.HandleFunc("POST /login", app.handleLoginPost)
	mux.HandleFunc("GET /register", app.handleRegisterGet)
	mux.HandleFunc("POST /register", app.handleRegisterPost)
	mux.HandleFunc("POST /logout", app.handleLogout)

	// Dashboard
	mux.HandleFunc("GET /app", app.handleDashboard)

	return mux
}

func (app *Application) Run(mux *http.ServeMux) error {
	server := &http.Server{
		Addr:    app.Config.Port,
		Handler: mux,
	}

	return server.ListenAndServe()
}