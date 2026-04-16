package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"log"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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

var cfg apiConfig

func main () {
	app := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(app))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
	
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
        	Body string `json:"body"`
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

		type returnVals struct {
			Valid bool `json:"valid"`
    	}
   		respBody := returnVals{
			Valid: true,
    	}
    	dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

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
		cfg.fileserverHits.Store(0)
		w.WriteHeader(http.StatusOK)
	})

	server := http.Server{
		Handler:	mux,
		Addr:		":8080",
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
	return
}

func respondWithJSON(w http.ResponseWriter, code int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(payload)
}
