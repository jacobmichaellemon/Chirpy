package main

import (
	"Chirpy/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq" //imported but not used, just need the side effect of the package
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	dbURL          string
	polkaKey       string
}

var cfg apiConfig

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Enviornement variables failed to load")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	if dbURL == "" || platform == "" || secret == "" || polkaKey == "" {
		log.Fatal("Failed to load enviornment variables")
	}

	cfg.dbURL = dbURL
	cfg.platform = platform
	cfg.secret = secret
	cfg.polkaKey = polkaKey

	log.Printf("Connecting to: %s", dbURL)

	postgres, dberr := sql.Open("postgres", dbURL)
	if dberr != nil {
		log.Fatal("Failed to open a connection to the database")
	}
	db := database.New(postgres)
	cfg.db = db

	mux := http.NewServeMux()
	cfg.registerRoutes(mux)

	server := http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}
