package main

import (
	"log"

	"github.com/philip-h/amics/internal/server"
)

func main() {
	cfg := server.Config{
		Port: ":8080",
	}

	app := &server.Application{
		Config: cfg,
	}

	mux := app.Mount()
	log.Printf("Starting server on port %s", cfg.Port)
	if err := app.Run(mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}