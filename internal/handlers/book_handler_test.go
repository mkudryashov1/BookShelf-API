package handlers

import (
	"bookshelf-api/internal/models"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockBookRepository struct {
	createFunc func(context.Context, *models.Book) error
	listFunc   func(context.Context, int, int) ([]models.Book, error)
}

func (m *mockBookRepository) Create(ctx context.Context, book *models.Book) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, book)
	}

	return nil
}

func (m *mockBookRepository) GetByID(context.Context, uint) (*models.Book, error) {
	return nil, sql.ErrNoRows
}

func (m *mockBookRepository) List(ctx context.Context, limit, offset int) ([]models.Book, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}

	return nil, nil
}

func (m *mockBookRepository) Update(context.Context, *models.Book) error {
	return nil
}

func (m *mockBookRepository) Delete(context.Context, uint) error {
	return nil
}

func (m *mockBookRepository) MarkOutOfStock(context.Context, uint) error {
	return nil
}

func (m *mockBookRepository) GetTopRated(context.Context, int) ([]models.Book, error) {
	return nil, nil
}

func TestCreateBook(t *testing.T) {
	repo := &mockBookRepository{
		createFunc: func(ctx context.Context, book *models.Book) error {
			book.ID = 1
			return nil
		},
	}

	handler := NewBookHandler(repo)

	body := `{
		"title": "1984",
		"author": "George Orwell",
		"year": 1949,
		"rating": 5
	}`

	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.CreateBook(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"id":1`) {
		t.Fatalf("expected response to contain book id, got %s", rec.Body.String())
	}
}

func TestCreateBookInvalidRating(t *testing.T) {
	repo := &mockBookRepository{}

	handler := NewBookHandler(repo)

	body := `{
		"title": "1984",
		"author": "George Orwell",
		"year": 1949,
		"rating": 10
	}`

	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateBook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateBookInvalidJSON(t *testing.T) {
	repo := &mockBookRepository{}

	handler := NewBookHandler(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/books",
		strings.NewReader(`{"title":"broken"`),
	)

	rec := httptest.NewRecorder()

	handler.CreateBook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestListBooksPagination(t *testing.T) {
	repo := &mockBookRepository{
		listFunc: func(ctx context.Context, limit, offset int) ([]models.Book, error) {
			if limit != 5 {
				t.Errorf("expected limit 5, got %d", limit)
			}

			if offset != 5 {
				t.Errorf("expected offset 5, got %d", offset)
			}

			return []models.Book{
				{
					ID:     6,
					Title:  "The Master and Margarita",
					Author: "Mikhail Bulgakov",
					Year:   1967,
				},
			}, nil
		},
	}

	handler := NewBookHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?page=2&limit=5",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListBooks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"id":6`) {
		t.Fatalf("expected response to contain book id, got %s", rec.Body.String())
	}
}

func TestListBooksInvalidLimit(t *testing.T) {
	repo := &mockBookRepository{}

	handler := NewBookHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/books?limit=101",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ListBooks(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
