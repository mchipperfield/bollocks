package api

import (
	"encoding/json"
	"net/http"
)

func NewHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", Health)
	return mux
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/health+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "pass",
		"serviceId":   "https://api.bollocks.social",
		"description": "health check endpoint for the bollocks.social API",
	})
}
