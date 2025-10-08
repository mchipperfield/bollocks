package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mchipperfield/gocore/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Post defines the structure of a post as returned by the API.
// Specifically, it does not include the author field as this should not be exposed to the client.
type Post struct {
	ID        string    `json:"id"`
	Bollocks  string    `json:"bollocks"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	Likes     int       `json:"likes"`
}

// GET /feed
func GetFeed(logger log.Logger, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := s.GetFeed(r.Context())
		if err != nil {
			logger.Log("failed to get feed", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(posts)
	}
}

// POST /posts
func CreatePost(logger log.Logger, s Service, geminiAPIKey string) http.HandlerFunc {
	type request struct {
		Bollocks string `json:"bollocks"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tags, err := generateTags(r.Context(), geminiAPIKey, req.Bollocks)
		if err != nil {
			logger.Log("failed to generate AI tags, falling back", "error", err)
			// Fallback to hashtag generation on error
			tags = generateTagsFromHashtags(req.Bollocks)
		}

		post, err := s.CreatePost(r.Context(), req.Bollocks, tags)
		if err != nil {
			logger.Log("failed to create post", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/posts/"+post.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(post)

	}
}

// GET /posts
func GetPosts(logger log.Logger, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := s.GetPosts(r.Context())
		if err != nil {
			logger.Log("failed to get posts", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(posts)
	}
}

// PATCH /posts/{postId}
func UpdatePost(logger log.Logger, s Service, geminiAPIKey string) http.HandlerFunc {
	type request struct {
		Bollocks string `json:"bollocks"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		postID := r.PathValue("postId")
		tags, err := generateTags(r.Context(), geminiAPIKey, req.Bollocks)
		if err != nil {
			logger.Log("failed to generate AI tags, falling back", "error", err)
			// Fallback to hashtag generation on error
			tags = generateTagsFromHashtags(req.Bollocks)
		}

		post, err := s.UpdatePost(r.Context(), postID, req.Bollocks, tags)
		if err != nil {
			switch {
			case status.Code(err) == codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
			case status.Code(err) == codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
			default:
				logger.Log("failed to update bollocks", "error", err, "post_id", postID)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(post)

	}
}

// DELETE /posts/{postId}
func DeletePost(logger log.Logger, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postId")
		err := s.DeletePost(r.Context(), postID)
		if err != nil {
			switch {
			case status.Code(err) == codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
			case status.Code(err) == codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
			default:
				logger.Log("failed to delete post", "error", err, "post_id", postID)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// POST /posts/{postId}/likes
func LikePost(logger log.Logger, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postId")
		post, err := s.ToggleLike(r.Context(), postID)
		if err != nil {
			switch {
			case status.Code(err) == codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
			case status.Code(err) == codes.NotFound:
				w.WriteHeader(http.StatusNotFound)
			default:
				logger.Log("failed to toggle like", "error", err, "post_id", postID)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(post)

	}
}
