package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
	}
	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Error deleting the db: %s", err)
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
