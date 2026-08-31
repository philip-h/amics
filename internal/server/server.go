package server

import (
	"log/slog"
	"net/http"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/storage"
)

type Config struct {
	Port string
	Db   *db.DbConfig
}

type Logger struct {
	L      *slog.Logger
	LogLvl *slog.LevelVar
}

func NewServer(
	logger *Logger,
	store *storage.Storage,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		logger,
		store,
	)

	var handler http.Handler = http.NewCrossOriginProtection().Handler(mux)
	handler = panicRecovery(logger, handler)
	handler = checkAuthMiddlewear(logger, store, handler)
	return handler
}
