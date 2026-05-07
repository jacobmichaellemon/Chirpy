package main

import (
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email              string `json:"email"`
		Password           string `json:"password"`
		Expires_In_Seconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	refreshparams := database.CreateRefreshTokenParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if params.Expires_In_Seconds == 0 || params.Expires_In_Seconds > 3600 {
		params.Expires_In_Seconds = 3600
	}

	user, err := cfg.db.UserLookupEmail(r.Context(), params.Email)

	if err != nil {
		log.Printf("Error sending query to db: %s", err)
		w.WriteHeader(500)
		return
	}

	valid_password, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if valid_password == false || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, (time.Duration(params.Expires_In_Seconds) * time.Second))

	if err != nil {
		log.Println("Issue making JWT token")
	}

	refreshtoken := auth.MakeRefreshToken()
	refreshparams.Token = refreshtoken
	refreshparams.UserID = user.ID
	cfg.db.CreateRefreshToken(r.Context(), refreshparams)

	respBody := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshparams.Token,
	}

	dat, err := json.Marshal(respBody)

	respondWithJSON(w, 200, dat)

}
