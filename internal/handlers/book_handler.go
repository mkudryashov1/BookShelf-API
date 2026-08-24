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
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if book.Title == "" || book.Author == "" {
		writeError(w, http.StatusBadRequest, "title and author are required")
		return
	}

	if book.Year < 0 {
		writeError(w, http.StatusBadRequest, "year must be greater than or equal to 0")
		return
	}

	if book.Rating != nil && (*book.Rating < 1 || *book.Rating > 5) {
		writeError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	if err := h.repo.Create(r.Context(), &book); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create book")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(book); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	book, err := h.repo.GetByID(r.Context(), uint(id))
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get book")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(book); err != nil {
		return
	}
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
			writeError(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}

		page = parsed
	}

	if page > maxPage {
		writeError(w, http.StatusBadRequest, "page is too large")
		return
	}

	limit := defaultLimit
	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}

		if parsed > maxLimit {
			writeError(w, http.StatusBadRequest, "limit must be less than or equal to 100")
			return
		}

		limit = parsed
	}

	offset := (page - 1) * limit

	books, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list books")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		return
	}
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var book models.Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if book.Title == "" || book.Author == "" {
		writeError(w, http.StatusBadRequest, "title and author are required")
		return
	}

	if book.Year < 0 {
		writeError(w, http.StatusBadRequest, "year must be greater than or equal to 0")
		return
	}

	if book.Rating != nil && (*book.Rating < 1 || *book.Rating > 5) {
		writeError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	book.ID = uint(id)

	if err := h.repo.Update(r.Context(), &book); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to update book")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(book); err != nil {
		return
	}
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	err = h.repo.Delete(r.Context(), uint(id))
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to delete book")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) MarkOutOfStock(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	err = h.repo.MarkOutOfStock(r.Context(), uint(id))
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to update book")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BookHandler) RecommendBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.repo.GetTopRated(r.Context(), 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get recommendations")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		return
	}
}
