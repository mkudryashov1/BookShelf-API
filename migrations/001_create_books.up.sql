CREATE TABLE IF NOT EXISTS books (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    year INTEGER NOT NULL,
    isbn TEXT,
    rating INTEGER,
    out_of_stock BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT books_rating_check
        CHECK (rating IS NULL OR rating BETWEEN 1 AND 5),

    CONSTRAINT books_year_check
        CHECK (year >= 0)
);

CREATE INDEX IF NOT EXISTS idx_books_author
    ON books(author);

CREATE INDEX IF NOT EXISTS idx_books_rating
    ON books(rating);