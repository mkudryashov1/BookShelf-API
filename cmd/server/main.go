package main

import (
	"bookshelf-api/internal/config"
	"bookshelf-api/internal/routes"
	"context"
	"database/sql"
	"log"
	"net/http"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("failed to connect to db:", err)
	}

	router := chi.NewRouter()
	routes.Register(router)

	addr := ":" + cfg.HTTPPort
	log.Println("starting server on", addr)

	http.ListenAndServe(addr, router)
}
