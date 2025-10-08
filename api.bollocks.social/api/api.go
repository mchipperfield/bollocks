package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mchipperfield/bollocks/api.bollocks.social/genai"
	"github.com/mchipperfield/gocore/log"
)

type Service interface {
	GetFeed(ctx context.Context) ([]Post, error)
	CreatePost(ctx context.Context, bollocks string, tags []string) (*Post, error)
	GetPosts(ctx context.Context) ([]Post, error)
	DeletePost(ctx context.Context, postID string) error
	UpdatePost(ctx context.Context, postID, bollocks string, tags []string) (*Post, error)
	ToggleLike(ctx context.Context, postID string) (*Post, error)
}

func NewHandler(logger log.Logger, s Service, ai *genai.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("GET /feed", GetFeed(logger, s))
	mux.HandleFunc("POST /posts", CreatePost(logger, s, ai))
	mux.HandleFunc("GET /posts", GetPosts(logger, s))
	mux.HandleFunc("PATCH /posts/{postId}", UpdatePost(logger, s, ai))
	mux.HandleFunc("DELETE /posts/{postId}", DeletePost(logger, s))
	mux.HandleFunc("POST /posts/{postId}/likes", LikePost(logger, s))
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
