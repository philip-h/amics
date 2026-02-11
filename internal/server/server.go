package server

import "net/http"

type Application struct {
	Config Config
}

type Config struct {
	Port string
	// TODO: DB Config
}

func (app *Application) Mount() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.handleIndex)
	return mux
}

func (app *Application) Run(mux *http.ServeMux) error {
	server := &http.Server{
		Addr:    app.Config.Port,
		Handler: mux,
	}

	return server.ListenAndServe()
}