package handlers

import (
	"bookshelf-api/internal/models"
	"bookshelf-api/internal/repository"
	"encoding/json"
	"net/http"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var book models.Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
	}

	if book.Title == "" || book.Author == "" {
		http.Error(w, "title and author are required", http.StatusBadRequest)
		return
	}

	err := h.repo.Create(r.Context(), &book)
	if err != nil {
		http.Error(w, "failed to create book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}
