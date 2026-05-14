package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpyRed(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string
		Data  struct {
			User_Id string
		}
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userID, err := uuid.Parse(params.Data.User_Id)
	if err != nil {
		log.Printf("Invalid UUID in data field: %s", err)
		w.WriteHeader(500)
		return
	}

	if params.Event == "user.upgraded" {
		err := cfg.db.UpgradeToChirpyRed(r.Context(), userID)
		if err != nil {
			log.Printf("Cannot find user to upgrade: %s", err)
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(204)
		return
	}
}
