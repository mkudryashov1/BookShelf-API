package handlers

import (
	"bookshelf-api/internal/models"
	"bookshelf-api/internal/repository"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
		return
	}

	if book.Title == "" || book.Author == "" {
		http.Error(w, "title and author are required", http.StatusBadRequest)
		return
	}

	if book.Year < 0 {
		http.Error(w, "year must be greater than or equal to 0", http.StatusBadRequest)
		return
	}

	if book.Rating != nil && (*book.Rating < 1 || *book.Rating > 5) {
		http.Error(w, "rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), &book); err != nil {
		http.Error(w, "failed to create book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(book); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid book id", http.StatusBadRequest)
		return
	}

	book, err := h.repo.GetByID(r.Context(), uint(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	const (
		defaultPage  = 1
		defaultLimit = 10
		maxLimit     = 100
		maxPage      = 1_000_000
	)

	query := r.URL.Query()

	page := defaultPage
	if value := query.Get("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			http.Error(w, "page must be a positive integer", http.StatusBadRequest)
			return
		}

		page = parsed
	}

	if page > maxPage {
		http.Error(w, "page is too large", http.StatusBadRequest)
		return
	}

	limit := defaultLimit
	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
			return
		}

		if parsed > maxLimit {
			http.Error(w, "limit must be less than or equal to 100", http.StatusBadRequest)
			return
		}

		limit = parsed
	}

	offset := (page - 1) * limit

	books, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "failed to list books", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	book.ID = uint(id)

	err = h.repo.Update(r.Context(), &book)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update book", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(r.Context(), uint(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete book", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) MarkOutOfStock(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.repo.MarkOutOfStock(r.Context(), uint(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update book", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BookHandler) RecommendBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.repo.GetTopRated(r.Context(), 5)
	if err != nil {
		http.Error(w, "failed to get recommendations", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(books)
}
