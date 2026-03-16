package main

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/philip-h/amics/internal/auth"
	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/server"
	"github.com/philip-h/amics/internal/services"
	"github.com/philip-h/amics/internal/store"
	"github.com/philip-h/amics/templates"
)

func getenv(key, preset string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	} else {
		return preset
	}
}

func main() {

	// Logging setup
	levelVar := new(slog.LevelVar)
	levelVar.Set(slog.LevelInfo)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelVar,
	}))

	// DB Setup
	dbConfig := &db.DbConfig{
		User:     getenv("DATABASE_USER", "postgres"),
		Password: getenv("DATABASE_PASSWORD", ""),
		Host:     getenv("DATABASE_HOST", "0.0.0.0"),
		DbName:   getenv("DATABASE_NAME", "amics"),
		Params:   getenv("DATABASE_PARAMS", "sslmode=disable"),
	}
	cfg := server.Config{
		Port: getenv("SERVER_PORT", ":8080"),
		Db:   dbConfig,
	}
	db, err := db.New(dbConfig)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("msg", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	store := store.New(db)

	// Templates Setup
	templates, err := template.New("").Funcs(template.FuncMap{
		"unixToDate": func(unix int64) string {
			t := time.Unix(unix, 0)
			return t.Format("Mon Jan 2 @ 15:04")
		},
	}).ParseFS(templates.TemplateFS, "pages/*.html", "partials/*.html", "admin/*.html")
	if err != nil {
		logger.Error("Failed to load templates", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	// JWT Setup
	jwtKey := getenv("JWT_SECRET", "super-secret-key")
	jwtIss := getenv("JWT_ISS", "amics-server")
	auth := auth.NewJwtAuthenticator(jwtKey, jwtIss)

	// The app!
	app := &server.Application{
		Config:    cfg,
		Store:     store,
		Auth:      auth,
		Templates: templates,
		Logger:    logger,
	}

	// Grader worker
	worker, err := services.NewWorker(db, logger)
	if err != nil {
		logger.Error("Failed to start worker", slog.String("msg", err.Error()))
		os.Exit(1)
	}
	go worker.Start()
	defer worker.Stop()

	// Let's get started!
	mux := app.Mount()

	server := &http.Server{
		Addr:    app.Config.Port,
		Handler: mux,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.Info("Starting server on port " + cfg.Port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", slog.String("msg", err.Error()))
			os.Exit(1)
		}
		logger.Info("Stopped serving new connections")
	}()

	<-quit
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown error", slog.String("msg", err.Error()))
		os.Exit(1)
	}
	close(done)

	<-done
	logger.Info("Graceful shutdown copmplete")
}
