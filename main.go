package main

import (
	"Chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
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
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getServerHits() int32 {
	numHits := cfg.fileserverHits.Load()
	return numHits
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
	if platform != "" {
		cfg.platform = platform
	}
	log.Printf("Connecting to: %s", dbURL)

	postgres, dberr := sql.Open("postgres", dbURL)
	if dberr != nil {
		log.Fatal("Failed to open a connection to the database")
	}
	db := database.New(postgres)
	cfg.db = db

	app := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(app))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email string
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}

		user, err := cfg.db.CreateUser(r.Context(), params.Email)

		if err != nil {
			log.Printf("Error sending query to db: %s", err)
			w.WriteHeader(500)
			return
		}

		respBody := User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		}

		dat, err := json.Marshal(respBody)

		respondWithJSON(w, 201, dat)

	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}

		if len(params.Body) > 140 {
			respondWithError(w, 400, "Chirp is too long")
			return
		}

		params.Body = filterProfanity(params.Body)

		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   params.Body,
			UserID: params.UserID,
		})

		if err != nil {
			log.Printf("Error sending query to db: %s", err)
			w.WriteHeader(500)
			return
		}

		respBody := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			User_ID:   chirp.UserID,
		}

		dat, err := json.Marshal(respBody)

		respondWithJSON(w, 201, dat)

	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {

		chirps, err := cfg.db.GetChirps(r.Context())

		if err != nil {
			log.Printf("Error sending query to db: %s", err)
			w.WriteHeader(500)
			return
		}

		var respBody []Chirp

		for _, chirp := range chirps {
			newChirp := Chirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				User_ID:   chirp.UserID,
			}
			respBody = append(respBody, newChirp)
		}

		dat, err := json.Marshal(respBody)

		respondWithJSON(w, 200, dat)

	})

	mux.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {

		chirpIdStr := r.PathValue("id")
		chirpId, err := uuid.Parse(chirpIdStr)

		if err != nil {
			log.Printf("User_ID not found: %s", err)
			w.WriteHeader(404)
		}

		log.Printf("SEARCHING FOR %v", chirpId)
		chirp, err := cfg.db.GetChirp(r.Context(), chirpId)

		if err != nil {
			w.WriteHeader(404)
			return
		}

		respBody := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			User_ID:   chirp.UserID,
		}

		dat, err := json.Marshal(respBody)

		respondWithJSON(w, 200, dat)

	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		numVisits := cfg.getServerHits()
		metrics := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", numVisits)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metrics))
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if cfg.platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
		}
		err := cfg.db.DeleteUsers(r.Context())
		if err != nil {
			log.Printf("Error deleting the db: %s", err)
		}
		cfg.fileserverHits.Store(0)
		w.WriteHeader(http.StatusOK)
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
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
