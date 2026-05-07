package main

import (
	"Chirpy/internal/auth"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refresh, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Refresh token not found: %s", err)
		w.WriteHeader(401)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refresh)
	if err != nil {
		log.Printf("No user found with refresh token: %s", err)
		w.WriteHeader(401)
		return
	}
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.MakeJWT(user, cfg.secret, (1 * time.Hour))
	if err != nil {
		log.Printf("Error creating access token: %s", err)
		w.WriteHeader(401)
		return
	}

	tokenResponse := response{
		Token: token,
	}

	dat, err := json.Marshal(tokenResponse)

	respondWithJSON(w, 200, dat)
}
