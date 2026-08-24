package routes

import (
	"bookshelf-api/internal/handlers"
	"bookshelf-api/internal/health"

	"github.com/go-chi/chi/v5"
)

func Register(
	router chi.Router,
	bookHandler *handlers.BookHandler,
	healthHandler *health.Handler,
) {
	router.Get("/healthz", healthHandler.Healthz)
	router.Get("/readyz", healthHandler.Readyz)

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
