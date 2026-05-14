package main

import (
	"Chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {

	type Chirps struct {
		Request []database.Chirp
		Error   error
		Type    string
	}

	author_id := r.URL.Query().Get("author_id")
	chirps := Chirps{Request: nil, Error: nil, Type: author_id}

	if chirps.Type == "" {
		chirps.Request, chirps.Error = cfg.db.GetChirps(r.Context())
	} else if chirps.Type != "" {
		id, err := uuid.Parse(chirps.Type)

		if err != nil {
			log.Printf("Invalid author ID: %s", chirps.Error)
			w.WriteHeader(500)
			return
		}

		chirps.Request, chirps.Error = cfg.db.GetChirpsByAuthor(r.Context(), id)
	}

	if chirps.Error != nil {
		log.Printf("Error sending query to db: %s", chirps.Error)
		w.WriteHeader(500)
		return
	}

	var respBody []Chirp

	for _, chirp := range chirps.Request {
		newChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			User_ID:   chirp.UserID,
		}
		respBody = append(respBody, newChirp)
	}

	sorting := r.URL.Query().Get("sort")

	if sorting == "" {
		sorting = "asc"
	}

	if sorting == "asc" {
		sort.Slice(respBody, func(i, j int) bool {
			return respBody[i].CreatedAt.Before(respBody[j].CreatedAt)
		})
	}
	if sorting == "desc" {
		sort.Slice(respBody, func(i, j int) bool {
			return respBody[i].CreatedAt.After(respBody[j].CreatedAt)
		})
	}

	dat, err := json.Marshal(respBody)

	if err != nil {
		log.Printf("Error marshling resonse: %s", err)
		w.WriteHeader(500)
		return
	}

	respondWithJSON(w, 200, dat)

}
