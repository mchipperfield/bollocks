package api

import (
	"encoding/json"
	"net/http"

	"github.com/mchipperfield/gocore/log"

	"cloud.google.com/go/firestore"
)

func NewHandler(logger log.Logger, client *firestore.Client) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("GET /feed", GetFeed(logger, client))
	mux.HandleFunc("POST /posts", CreatePost())
	return mux
}

func PanicMw(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Log("recovered from panic", "error", err)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func LoggingMiddleware(logger log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Log("request received", "method", r.Method, "url", r.URL, "proto", r.Proto, "remote_addr", r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
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
