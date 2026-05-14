package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
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
		s := r.URL.Query().Get("author_id")
		// s is a string that contains the value of the author_id query parameter
		// if it exists, or an empty string if it doesn't
		if s != "" {
			if s == chirp.UserID.String() {
				newChirp := Chirp{
					ID:        chirp.ID,
					CreatedAt: chirp.CreatedAt,
					UpdatedAt: chirp.UpdatedAt,
					Body:      chirp.Body,
					User_ID:   chirp.UserID,
				}
				respBody = append(respBody, newChirp)
			}
		} else {
			newChirp := Chirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				User_ID:   chirp.UserID,
			}
			respBody = append(respBody, newChirp)
		}
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

	respondWithJSON(w, 200, dat)

}
