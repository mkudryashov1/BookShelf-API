package main

import (
	"bookshelf-api/internal/config"
	"bookshelf-api/internal/routes"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	routes.Register(router)

	cfg := config.Load()
	addr := ":" + cfg.HTTPPort

	http.ListenAndServe(addr, router)
}
