package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

var swears = []string{"kerfuffle", "sharbert", "fornax"}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(payload)
}

func filterProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		badword := slices.Contains(swears, strings.ToLower(word))
		if badword {
			words[i] = "****"
		}
	}
	cleaned_text := strings.Join(words, " ")
	return cleaned_text
}
