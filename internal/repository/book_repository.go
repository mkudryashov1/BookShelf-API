package repository

import (
	"bookshelf-api/internal/models"
	"context"
	"database/sql"
)

type BookRepository interface {
	Create(ctx context.Context, book *models.Book) error
	GetByID(ctx context.Context, id uint) (*models.Book, error)
}
type PostgresBookRepository struct {
	db *sql.DB
}

func NewPostgresBookRepository(db *sql.DB) *PostgresBookRepository {
	return &PostgresBookRepository{db: db}
}

func (r *PostgresBookRepository) Create(ctx context.Context, book *models.Book) error {
	query := `
		INSERT INTO books (title, author, year, isbn, rating, out_of_stock)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(
		ctx,
		query,
		book.Title,
		book.Author,
		book.Year,
		book.ISBN,
		book.Rating,
		book.OutOfStock,
	).Scan(&book.ID, &book.CreatedAt, &book.UpdatedAt)
}

func (r *PostgresBookRepository) GetByID(ctx context.Context, id uint) (*models.Book, error) {
	query := `
		SELECT
			id, title, author, year, isbn, rating, out_of_stock,
			created_at, updated_at
		FROM books
		WHERE id = $1
	`

	var book models.Book

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
		&book.ISBN,
		&book.Rating,
		&book.OutOfStock,
		&book.CreatedAt,
		&book.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &book, nil
}
