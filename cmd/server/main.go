package main

import (
	"bookshelf-api/internal/routes"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	routes.Register(router)

	http.ListenAndServe(":8080", router)
}
