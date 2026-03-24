package repository

import (
	"bookshelf-api/internal/models"
	"context"
	"database/sql"
)

type BookRepository interface {
	Create(ctx context.Context, book *models.Book) error
	GetByID(ctx context.Context, id uint) (*models.Book, error)
	List(ctx context.Context, limit, offset int) ([]models.Book, error)
	Update(ctx context.Context, book *models.Book) error
	Delete(ctx context.Context, id uint) error
	MarkOutOfStock(ctx context.Context, id uint) error
	GetTopRated(ctx context.Context, limit int) ([]models.Book, error)
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

func (r *PostgresBookRepository) List(ctx context.Context, limit, offset int) ([]models.Book, error) {
	query := `
		SELECT id, title, author, year, isbn, rating, out_of_stock, created_at, updated_at
		FROM books
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book

	for rows.Next() {
		var b models.Book
		err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.Author,
			&b.Year,
			&b.ISBN,
			&b.Rating,
			&b.OutOfStock,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	return books, nil
}

func (r *PostgresBookRepository) Update(ctx context.Context, book *models.Book) error {
	query := `
		UPDATE books
		SET title = $1, author = $2, year = $3, isbn = $4, rating = $5, out_of_stock = $6, updated_at = NOW()
		WHERE id = $7
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		book.Title,
		book.Author,
		book.Year,
		book.ISBN,
		book.Rating,
		book.OutOfStock,
		book.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresBookRepository) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM books WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresBookRepository) MarkOutOfStock(ctx context.Context, id uint) error {
	query := `
		UPDATE books
		SET out_of_stock = true, updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresBookRepository) GetTopRated(ctx context.Context, limit int) ([]models.Book, error) {
	query := `
		SELECT id, title, author, year, isbn, rating, out_of_stock, created_at, updated_at
		FROM books
		WHERE rating IS NOT NULL
		ORDER BY rating DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []models.Book{}

	for rows.Next() {
		var b models.Book
		err := rows.Scan(
			&b.ID,
			&b.Title,
			&b.Author,
			&b.Year,
			&b.ISBN,
			&b.Rating,
			&b.OutOfStock,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	return books, nil
}
