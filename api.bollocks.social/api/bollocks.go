package api

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mchipperfield/gocore/log"
	"google.golang.org/api/iterator"
)

// Post defines the structure of a post as returned by the API.
// Specifically, it does not include the author field as this should not be exposed to the client.
type Post struct {
	ID        string    `json:"id"`
	Bollocks  string    `json:"bollocks"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// post as it is stored in firestore.
// id is not included as it is part of the document reference.
type post struct {
	Bollocks  string    `firestore:"bollocks"`
	Tags      []string  `firestore:"tags"`
	Author    string    `firestore:"author"`
	CreatedAt time.Time `firestore:"created_at"`
}

func GetFeed(logger log.Logger, client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := client.Collection("bollocks").Query

		iter := query.Documents(r.Context())
		var ret []Post
		for {
			docSnap, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var p post
			if err := docSnap.DataTo(&p); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ret = append(ret, Post{
				ID:        docSnap.Ref.ID,
				Bollocks:  p.Bollocks,
				Tags:      p.Tags,
				CreatedAt: docSnap.CreateTime,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ret)
	}
}

func CreatePost(logger log.Logger, client *firestore.Client) http.HandlerFunc {
	type request struct {
		Bollocks string   `json:"bollocks"`
		Tags     []string `json:"tags"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userId, _ := ContextGetUserId(r.Context())

		docRef, wr, err := client.Collection("bollocks").Add(r.Context(), map[string]any{
			"bollocks": req.Bollocks,
			"tags":     req.Tags,
			"author":   userId,
		})
		if err != nil {
			logger.Log("failed to create post", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/posts/"+docRef.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":         docRef.ID,
			"bollocks":   req.Bollocks,
			"tags":       req.Tags,
			"created_at": wr.UpdateTime,
		})

	}
}
