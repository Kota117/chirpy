package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden) // HTTP Status Code 403
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.Reset(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // HTTP Status Code 500
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK) // HTTP Status Code 200
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}
