package main

import (
	"bookshelf-api/internal/config"
	"bookshelf-api/internal/routes"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file found")
	}

	router := chi.NewRouter()

	routes.Register(router)

	cfg := config.Load()
	addr := ":" + cfg.HTTPPort

	http.ListenAndServe(addr, router)
}
