package main

import (
	"bookshelf-api/internal/config"
	"bookshelf-api/internal/handlers"
	"bookshelf-api/internal/health"
	"bookshelf-api/internal/middleware"
	"bookshelf-api/internal/repository"
	"bookshelf-api/internal/routes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		logger.Warn("no .env file found")
	}

	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		return
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
			logger.Info("database is ready")
			break
		}

		if ctx.Err() != nil {
			logger.Info("shutdown requested while waiting for database")
			return
		}

		logger.Info("waiting for database")
		time.Sleep(500 * time.Millisecond)
	}

	bookRepo := repository.NewPostgresBookRepository(db)
	bookHandler := handlers.NewBookHandler(bookRepo)
	healthHandler := health.NewHandler(db)

	router := chi.NewRouter()
	router.Use(middleware.Logging(logger))

	routes.Register(router, bookHandler, healthHandler)

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
		logger.Info("starting server", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
		}

	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return
	}

	logger.Info("server stopped")
}
