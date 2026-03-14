package main

import (
	"html/template"
	"log"
	"os"
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
	// Load Env

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

	// Parse all the layouts
	templates, err := template.New("").Funcs(template.FuncMap{
		"unixToDate": func(unix int64) string {
			t := time.Unix(unix, 0)
			return t.Format("Mon Jan 2 @ 15:04")
		},
	}).ParseFS(templates.TemplateFS, "pages/*.html", "partials/*.html", "admin/*.html")
	if err != nil {
		log.Fatal("Failed to load templates: " + err.Error())
	}

	db, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := store.New(db)

	// TODO: KEY should come from config/env variable
	jwtKey := getenv("JWT_SECRET", "super-secret-key")
	jwtIss := getenv("JWT_ISS", "amics-server")
	auth := auth.NewJwtAuthenticator(jwtKey, jwtIss)

	app := &server.Application{
		Config:    cfg,
		Store:     store,
		Auth:      auth,
		Templates: templates,
	}

	worker, err := services.NewWorker(db)
	if err != nil {
		log.Fatalf("Failed to start worker. %v", err)
	}

	go worker.Start()
	defer worker.Stop()

	mux := app.Mount()
	log.Printf("Starting server on port %s", cfg.Port)
	if err := app.Run(mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
