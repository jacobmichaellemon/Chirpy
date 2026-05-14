package main

import "net/http"

func (cfg *apiConfig) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users", cfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateCredentials)
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)
	mux.HandleFunc("POST /api/chirps", cfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", cfg.handlerChirpsGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerChirpGet)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerChirpDelete)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerChirpyRed)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	mux.Handle("/app/", cfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	))
}
