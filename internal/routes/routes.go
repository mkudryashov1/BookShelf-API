package routes

import (
	"bookshelf-api/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func Register(router chi.Router, bookHandler *handlers.BookHandler) {
	router.Get("/health", handlers.Health)

	router.Route("/books", func(r chi.Router) {
		r.Post("/", bookHandler.CreateBook)
		r.Get("/", bookHandler.ListBooks)
		r.Get("/recommend", bookHandler.RecommendBooks)
		r.Get("/{id}", bookHandler.GetBookByID)
		r.Put("/{id}", bookHandler.UpdateBook)
		r.Delete("/{id}", bookHandler.DeleteBook)
		r.Post("/{id}/mark-out-of-stock", bookHandler.MarkOutOfStock)
	})
}
