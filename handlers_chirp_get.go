package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {

	chirpIdStr := r.PathValue("id")
	chirpId, err := uuid.Parse(chirpIdStr)

	if err != nil {
		log.Printf("User_ID not found: %s", err)
		w.WriteHeader(404)
	}

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

}
