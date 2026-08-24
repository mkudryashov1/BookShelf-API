package main

import (
	"bookshelf-api/internal/config"
	"bookshelf-api/internal/handlers"
	"bookshelf-api/internal/repository"
	"bookshelf-api/internal/routes"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file found")
	}

	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	for {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

		err := db.PingContext(pingCtx)
		cancel()

		if err == nil {
			log.Println("database is ready")
			break
		}

		if ctx.Err() != nil {
			log.Println("shutdown requested while waiting for database")
			return
		}

		log.Println("waiting for database...")
		time.Sleep(500 * time.Millisecond)
	}

	bookRepo := repository.NewPostgresBookRepository(db)
	bookHandler := handlers.NewBookHandler(bookRepo)

	router := chi.NewRouter()
	routes.Register(router, bookHandler)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Println("starting server on", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}

	case <-ctx.Done():
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		return
	}

	log.Println("server stopped")
}
