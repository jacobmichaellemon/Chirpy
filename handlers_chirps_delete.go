package main

import (
	"Chirpy/internal/auth"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		log.Printf("No auth token found: %s", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer badToken")
		return
	}

	chirpIdStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIdStr)

	if err != nil {
		log.Printf("Error parsing chirp path: %s", err)
		w.WriteHeader(500)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)

	if err != nil {
		respondWithError(w, 404, "Chirp to delete not found")
		return
	}

	if userId != chirp.UserID {
		respondWithError(w, 403, "not authorized")
		return
	}

	deleteErr := cfg.db.DeleteChirp(r.Context(), chirp.ID)

	if deleteErr != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}
