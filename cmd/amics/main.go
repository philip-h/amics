package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/server"
	"github.com/philip-h/amics/internal/services"
	"github.com/philip-h/amics/internal/storage"
	"github.com/philip-h/amics/templates"
)

// func getenv(key, preset string) string {
// 	value, exists := os.LookupEnv(key)
// 	if exists {
// 		return value
// 	} else {
// 		return preset
// 	}
// }


func run(
	ctx context.Context,
	getenv func(string) string,
) error {
	// Logging setup
	levelVar := new(slog.LevelVar)
	levelVar.Set(slog.LevelInfo)
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelVar,
	}))
	logger := &server.Logger{L: log, LogLvl: levelVar}

	// DB Setup
	dbConn := getenv("DB_CONN")
	if dbConn == "" {
		dbConn = "postgresql://postgres@127.0.0.1/amics_test?sslmode=disable"
	}

	dbConfig := &db.DbConfig{
		ConnStr: dbConn,
	}

	port := getenv("SERVER_PORT")
	if port == "" {
		port = ":8080"
	}

	cfg := &server.Config{
		Port: port,
		Db:   dbConfig,
	}
	db, err := db.New(dbConfig)
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()
	store := storage.New(db)

	// Templates Setup
	templates, err := templates.LoadTemplates()
	if err != nil {
		return fmt.Errorf("Failed to load templates: %w", err)
	}

	// Grader worker
	worker, err := services.NewWorker(db, logger)
	if err != nil {
		return fmt.Errorf("Failed to start worker: %w", err)
	}
	go worker.Start()
	defer worker.Stop()

	myServer := server.NewServer(
		logger,
		store,
		templates,
	)
	server := &http.Server{
		Addr:    cfg.Port,
		Handler: myServer,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.L.Info("Starting server on port " + cfg.Port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.L.Error("HTTP server error", slog.String("msg", err.Error()))
			os.Exit(1)
		}
		logger.L.Info("Stopped serving new connections")
	}()

	<-quit
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("Failed to shutdown server: %w", err)
	}
	close(done)

	<-done
	logger.L.Info("Graceful shutdown copmplete")
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
