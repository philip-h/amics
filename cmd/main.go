package main

import (
	"log"

	"github.com/philip-h/amics/internal/db"
	"github.com/philip-h/amics/internal/server"
	"github.com/philip-h/amics/internal/store"
)

func main() {
	dbConfig := &db.DbConfig{
		DbName:     "amics",
	}
	cfg := server.Config{
		Port: ":8080",
		Db:   dbConfig,
	}

	db, err := db.New(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	store := store.New(db)

	app := &server.Application{
		Config: cfg,
		Store:  store,
	}

	mux := app.Mount()
	log.Printf("Starting server on port %s", cfg.Port)
	if err := app.Run(mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}