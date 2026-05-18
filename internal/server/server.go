package server

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/storage"
)

type Application struct {
	Config    Config
	Store     storage.Storage
	Templates map[string]*template.Template
	Logger    *slog.Logger
	LogLvl    *slog.LevelVar
}

type Config struct {
	Port string
	Db   *db.DbConfig
}

type Logger struct {
	L      *slog.Logger
	LogLvl *slog.LevelVar
}

type Validator interface {
	Valid() (problems map[string]string)
}

type NavLink struct {
	Text string
	Href string
}

func NewServer(
	logger *Logger,
	store *storage.Storage,
	tmpls map[string]*template.Template,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		logger,
		store,
		tmpls,
	)

	var handler http.Handler = mux
	handler = panicRecovery(logger, handler)
	handler = checkAuthMiddlewear(logger, store, handler)
	return handler
}
