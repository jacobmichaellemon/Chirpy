package main

import (
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerUpdateCredentials(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("No auth token found: %s", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer badToken")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password %s", err)
		w.WriteHeader(500)
		return
	}

	updated_user, err := cfg.db.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{ID: userId, Email: params.Email, HashedPassword: hashedPassword})
	if err != nil {
		log.Printf("Error updating user: %s", err)
		w.WriteHeader(401)
		return
	}

	respBody := User{
		ID:        updated_user.ID,
		CreatedAt: updated_user.CreatedAt,
		UpdatedAt: updated_user.UpdatedAt,
		Email:     params.Email,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshaling response data: %s", err)
		w.WriteHeader(500)
		return
	}

	respondWithJSON(w, 200, dat)
}
