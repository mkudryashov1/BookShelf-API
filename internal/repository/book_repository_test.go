package repository

import (
	"bookshelf-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5434/bookshelf_test?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	if _, err := db.ExecContext(context.Background(), "TRUNCATE TABLE books RESTART IDENTITY"); err != nil {
		t.Fatalf("failed to clean books table: %v", err)
	}

	return db
}

func TestPostgresBookRepository_CreateAndGetByID(t *testing.T) {
	db := testDB(t)
	repo := NewPostgresBookRepository(db)

	isbn := "9780451524935"
	rating := 5

	book := &models.Book{
		Title:  "1984",
		Author: "George Orwell",
		Year:   1949,
		ISBN:   &isbn,
		Rating: &rating,
	}

	err := repo.Create(context.Background(), book)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if book.ID == 0 {
		t.Fatal("Create() did not set book ID")
	}

	got, err := repo.GetByID(context.Background(), book.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Title != book.Title {
		t.Errorf("Title = %q, want %q", got.Title, book.Title)
	}

	if got.Author != book.Author {
		t.Errorf("Author = %q, want %q", got.Author, book.Author)
	}

	if got.Year != book.Year {
		t.Errorf("Year = %d, want %d", got.Year, book.Year)
	}

	if got.ISBN == nil || *got.ISBN != *book.ISBN {
		t.Errorf("ISBN = %v, want %v", got.ISBN, book.ISBN)
	}

	if got.Rating == nil || *got.Rating != *book.Rating {
		t.Errorf("Rating = %v, want %v", got.Rating, book.Rating)
	}
}

func TestPostgresBookRepository_Update(t *testing.T) {
	db := testDB(t)
	repo := NewPostgresBookRepository(db)

	book := &models.Book{
		Title:  "1984",
		Author: "George Orwell",
		Year:   1949,
	}

	if err := repo.Create(context.Background(), book); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	book.Title = "Animal Farm"
	book.Author = "George Orwell"
	book.Year = 1945

	if err := repo.Update(context.Background(), book); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), book.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Title != "Animal Farm" {
		t.Errorf("Title = %q, want %q", got.Title, "Animal Farm")
	}

	if got.Author != "George Orwell" {
		t.Errorf("Author = %q, want %q", got.Author, "George Orwell")
	}

	if got.Year != 1945 {
		t.Errorf("Year = %d, want %d", got.Year, 1945)
	}
}

func TestPostgresBookRepository_Delete(t *testing.T) {
	db := testDB(t)
	repo := NewPostgresBookRepository(db)

	book := &models.Book{
		Title:  "1984",
		Author: "George Orwell",
		Year:   1949,
	}

	if err := repo.Create(context.Background(), book); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(context.Background(), book.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := repo.GetByID(context.Background(), book.ID)
	if err != sql.ErrNoRows {
		t.Fatalf("GetByID() error = %v, want sql.ErrNoRows", err)
	}
}

func TestPostgresBookRepository_List(t *testing.T) {
	db := testDB(t)
	repo := NewPostgresBookRepository(db)

	for i := 1; i <= 3; i++ {
		book := &models.Book{
			Title:  fmt.Sprintf("Book %d", i),
			Author: "Author",
			Year:   2000 + i,
		}

		if err := repo.Create(context.Background(), book); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	books, err := repo.List(context.Background(), 2, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(books) != 2 {
		t.Fatalf("len(books) = %d, want 2", len(books))
	}

	books, err = repo.List(context.Background(), 2, 2)
	if err != nil {
		t.Fatalf("List() with offset error = %v", err)
	}

	if len(books) != 1 {
		t.Fatalf("len(books) = %d, want 1", len(books))
	}
}

func TestPostgresBookRepository_MarkOutOfStock(t *testing.T) {
	db := testDB(t)
	repo := NewPostgresBookRepository(db)

	book := &models.Book{
		Title:  "1984",
		Author: "George Orwell",
		Year:   1949,
	}

	if err := repo.Create(context.Background(), book); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if book.OutOfStock {
		t.Fatal("new book should not be out of stock")
	}

	if err := repo.MarkOutOfStock(context.Background(), book.ID); err != nil {
		t.Fatalf("MarkOutOfStock() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), book.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if !got.OutOfStock {
		t.Fatal("OutOfStock = false, want true")
	}
}

func TestPostgresBookRepository_GetTopRated(t *testing.T) {
	db := testDB(t)
	repo := NewPostgresBookRepository(db)

	rating5 := 5
	rating3 := 3
	rating4 := 4

	books := []*models.Book{
		{
			Title:  "Book 1",
			Author: "Author 1",
			Year:   2001,
			Rating: &rating3,
		},
		{
			Title:  "Book 2",
			Author: "Author 2",
			Year:   2002,
			Rating: &rating5,
		},
		{
			Title:  "Book 3",
			Author: "Author 3",
			Year:   2003,
			Rating: &rating4,
		},
		{
			Title:  "Book without rating",
			Author: "Author 4",
			Year:   2004,
		},
	}

	for _, book := range books {
		if err := repo.Create(context.Background(), book); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := repo.GetTopRated(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetTopRated() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}

	if got[0].Title != "Book 2" {
		t.Errorf("got[0].Title = %q, want %q", got[0].Title, "Book 2")
	}

	if got[1].Title != "Book 3" {
		t.Errorf("got[1].Title = %q, want %q", got[1].Title, "Book 3")
	}

	if got[0].Rating == nil || *got[0].Rating != 5 {
		t.Errorf("got[0].Rating = %v, want 5", got[0].Rating)
	}

	if got[1].Rating == nil || *got[1].Rating != 4 {
		t.Errorf("got[1].Rating = %v, want 4", got[1].Rating)
	}
}
