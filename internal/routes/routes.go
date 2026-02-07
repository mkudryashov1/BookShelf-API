package routes

import (
	"bookshelf-api/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func Register(router chi.Router) {
	router.Get("/health", handlers.Health)
}
