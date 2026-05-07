package main

import (
	"Chirpy/internal/auth"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refresh, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Refresh token not found: %s", err)
		w.WriteHeader(401)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refresh)
	if err != nil {
		log.Printf("Error revoking refresh token: %s", err)
		w.WriteHeader(401)
		return
	}

	w.WriteHeader(204) //success, but no body sent
}
