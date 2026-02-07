package main

import (
	"bookshelf-api/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	router.Get("/health", handlers.Health)

	http.ListenAndServe(":8080", router)
}
