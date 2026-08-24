package handlers

import (
	"bookshelf-api/internal/models"
	"bookshelf-api/internal/repository"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type errorTestRepository struct {
	getByIDFunc func(ctx context.Context, id uint) (*models.Book, error)
	updateFunc  func(ctx context.Context, book *models.Book) error
	deleteFunc  func(ctx context.Context, id uint) error

	updateCalled bool
	deleteCalled bool
}

var _ repository.BookRepository = (*errorTestRepository)(nil)

func (r *errorTestRepository) Create(ctx context.Context, book *models.Book) error {
	return nil
}

func (r *errorTestRepository) GetByID(ctx context.Context, id uint) (*models.Book, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, id)
	}

	return nil, nil
}

func (r *errorTestRepository) List(ctx context.Context, limit, offset int) ([]models.Book, error) {
	return nil, nil
}

func (r *errorTestRepository) Update(ctx context.Context, book *models.Book) error {
	r.updateCalled = true

	if r.updateFunc != nil {
		return r.updateFunc(ctx, book)
	}

	return nil
}

func (r *errorTestRepository) Delete(ctx context.Context, id uint) error {
	r.deleteCalled = true

	if r.deleteFunc != nil {
		return r.deleteFunc(ctx, id)
	}

	return nil
}

func (r *errorTestRepository) MarkOutOfStock(ctx context.Context, id uint) error {
	return nil
}

func (r *errorTestRepository) GetTopRated(ctx context.Context, limit int) ([]models.Book, error) {
	return nil, nil
}

func newBookRequest(method, path, body, id string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", id)

	return req.WithContext(context.WithValue(
		req.Context(),
		chi.RouteCtxKey,
		routeCtx,
	))
}

func assertJSONError(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
	expectedMessage string,
) {
	t.Helper()

	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, response.Code)
	}

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", contentType)
	}

	var errorResponse ErrorResponse

	if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errorResponse.Error != expectedMessage {
		t.Fatalf("expected error %q, got %q", expectedMessage, errorResponse.Error)
	}
}

func TestCreateBookInvalidJSONReturnsJSONError(t *testing.T) {
	repo := &errorTestRepository{}
	handler := NewBookHandler(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/books",
		strings.NewReader(`{"title":`),
	)

	rec := httptest.NewRecorder()

	handler.CreateBook(rec, req)

	assertJSONError(
		t,
		rec,
		http.StatusBadRequest,
		"invalid JSON",
	)
}

func TestGetBookByIDNotFoundReturnsJSONError(t *testing.T) {
	repo := &errorTestRepository{
		getByIDFunc: func(ctx context.Context, id uint) (*models.Book, error) {
			return nil, sql.ErrNoRows
		},
	}

	handler := NewBookHandler(repo)

	req := newBookRequest(
		http.MethodGet,
		"/books/999",
		"",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.GetBookByID(rec, req)

	assertJSONError(
		t,
		rec,
		http.StatusNotFound,
		"book not found",
	)
}

func TestUpdateBookInvalidRating(t *testing.T) {
	repo := &errorTestRepository{}

	handler := NewBookHandler(repo)

	req := newBookRequest(
		http.MethodPut,
		"/books/1",
		`{
			"title": "1984",
			"author": "George Orwell",
			"year": 1949,
			"rating": 10
		}`,
		"1",
	)

	rec := httptest.NewRecorder()

	handler.UpdateBook(rec, req)

	assertJSONError(
		t,
		rec,
		http.StatusBadRequest,
		"rating must be between 1 and 5",
	)

	if repo.updateCalled {
		t.Fatal("repository Update should not be called for invalid rating")
	}
}

func TestUpdateBookNotFound(t *testing.T) {
	repo := &errorTestRepository{
		updateFunc: func(ctx context.Context, book *models.Book) error {
			return sql.ErrNoRows
		},
	}

	handler := NewBookHandler(repo)

	req := newBookRequest(
		http.MethodPut,
		"/books/999",
		`{
			"title": "1984",
			"author": "George Orwell",
			"year": 1949,
			"rating": 5
		}`,
		"999",
	)

	rec := httptest.NewRecorder()

	handler.UpdateBook(rec, req)

	assertJSONError(
		t,
		rec,
		http.StatusNotFound,
		"book not found",
	)
}

func TestUpdateBookSuccess(t *testing.T) {
	var updatedBook models.Book

	repo := &errorTestRepository{
		updateFunc: func(ctx context.Context, book *models.Book) error {
			updatedBook = *book
			return nil
		},
	}

	handler := NewBookHandler(repo)

	req := newBookRequest(
		http.MethodPut,
		"/books/1",
		`{
			"title": "1984",
			"author": "George Orwell",
			"year": 1949,
			"rating": 5
		}`,
		"1",
	)

	rec := httptest.NewRecorder()

	handler.UpdateBook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", contentType)
	}

	if !repo.updateCalled {
		t.Fatal("repository Update should be called")
	}

	if updatedBook.ID != 1 {
		t.Fatalf("expected book ID 1, got %d", updatedBook.ID)
	}

	if updatedBook.Title != "1984" {
		t.Fatalf("expected title %q, got %q", "1984", updatedBook.Title)
	}

	if updatedBook.Rating == nil || *updatedBook.Rating != 5 {
		t.Fatalf("expected rating 5, got %v", updatedBook.Rating)
	}
}

func TestDeleteBookNotFound(t *testing.T) {
	repo := &errorTestRepository{
		deleteFunc: func(ctx context.Context, id uint) error {
			return sql.ErrNoRows
		},
	}

	handler := NewBookHandler(repo)

	req := newBookRequest(
		http.MethodDelete,
		"/books/999",
		"",
		"999",
	)

	rec := httptest.NewRecorder()

	handler.DeleteBook(rec, req)

	assertJSONError(
		t,
		rec,
		http.StatusNotFound,
		"book not found",
	)
}
