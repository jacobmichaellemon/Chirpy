package main

import (
	"Chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq" //imported but not used, just need the side effect of the package
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_ID   uuid.UUID `json:"user_id"`
}

var cfg apiConfig

var swears = []string{"kerfuffle", "sharbert", "fornax"}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Enviornement variables failed to load")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	if platform != "" {
		cfg.platform = platform
	}
	if platform != "" {
		cfg.secret = secret
	}
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(payload)
}

func filterProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		badword := slices.Contains(swears, strings.ToLower(word))
		if badword {
			words[i] = "****"
		}
	}
	cleaned_text := strings.Join(words, " ")
	return cleaned_text
}
