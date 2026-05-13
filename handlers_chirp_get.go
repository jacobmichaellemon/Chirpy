package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {

	chirpIdStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIdStr)

	if err != nil {
		log.Printf("Chirp not found: %s", err)
		w.WriteHeader(404)
		return
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
