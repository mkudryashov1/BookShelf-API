package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPPort string
	DBDSN    string
}

func Load() *Config {
	port := os.Getenv("HTTP_PORT")
	dbDsn := os.Getenv("DB_DSN")
	if port == "" {
		port = "8080"
	}

	if dbDsn == "" {
		log.Fatal("DB_DSN is required")
	}

	return &Config{
		HTTPPort: port,
	}

}
