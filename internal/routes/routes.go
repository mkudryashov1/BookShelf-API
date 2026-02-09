package routes

import (
	"bookshelf-api/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func Register(router chi.Router, bookHandler *handlers.BookHandler) {
	router.Get("/health", handlers.Health)

	router.Route("/books", func(r chi.Router) {
		r.Post("/", bookHandler.CreateBook)
	})
}
