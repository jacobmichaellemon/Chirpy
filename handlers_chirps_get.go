package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {

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

}
