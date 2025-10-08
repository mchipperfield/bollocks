package firestore

import (
	"context"
	"errors"
	"slices"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/mchipperfield/bollocks/api.bollocks.social/api"
	"google.golang.org/api/iterator"
)

// post as it is stored in firestore.
// id is not included as it is part of the document reference.
type post struct {
	Bollocks  string    `firestore:"bollocks"`
	Tags      []string  `firestore:"tags"`
	Author    string    `firestore:"author"`
	CreatedAt time.Time `firestore:"created_at"`
	Likes     []string  `firestore:"likes"`
}

type Service struct {
	client *firestore.Client
}

func NewService(client *firestore.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetFeed(ctx context.Context) ([]api.Post, error) {
	userId, _ := api.ContextGetUserId(ctx)

	query := s.client.Collection("bollocks").Where("author", "!=", userId).OrderBy("created_at", firestore.Desc)
	iter := query.Documents(ctx)
	var posts []api.Post
	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var p post
		if err := docSnap.DataTo(&p); err != nil {
			return nil, err
		}

		posts = append(posts, api.Post{
			ID:        docSnap.Ref.ID,
			Bollocks:  p.Bollocks,
			Tags:      p.Tags,
			CreatedAt: docSnap.CreateTime,
			Likes:     len(p.Likes),
		})
	}
	return posts, nil
}

func (s *Service) CreatePost(ctx context.Context, bollocks string, tags []string) (*api.Post, error) {
	userId, _ := api.ContextGetUserId(ctx)
	now := time.Now()
	docRef, _, err := s.client.Collection("bollocks").Add(ctx, map[string]any{
		"bollocks":   bollocks,
		"tags":       tags,
		"author":     userId,
		"created_at": now,
		"likes":      []string{userId},
	})
	if err != nil {
		return nil, err
	}

	return &api.Post{
		ID:        docRef.ID,
		Bollocks:  bollocks,
		Tags:      tags,
		CreatedAt: now,
		Likes:     1,
	}, nil
}

func (s *Service) GetPosts(ctx context.Context) ([]api.Post, error) {
	userId, _ := api.ContextGetUserId(ctx)

	query := s.client.Collection("bollocks").Where("author", "==", userId).OrderBy("created_at", firestore.Desc)
	iter := query.Documents(ctx)
	var posts []api.Post
	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var p post
		if err := docSnap.DataTo(&p); err != nil {
			return nil, err
		}

		posts = append(posts, api.Post{
			ID:        docSnap.Ref.ID,
			Bollocks:  p.Bollocks,
			Tags:      p.Tags,
			CreatedAt: p.CreatedAt,
			Likes:     len(p.Likes),
		})
	}
	return posts, nil
}

func (s *Service) DeletePost(ctx context.Context, postID string) error {
	docRef := s.client.Collection("bollocks").Doc(postID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return err
	}

	author, err := docSnap.DataAt("author")
	if err != nil {
		return err
	}

	userID, _ := api.ContextGetUserId(ctx)

	if author != userID {
		return errors.New("forbidden")
	}

	_, err = docRef.Delete(ctx, firestore.Exists)
	return err
}

func (s *Service) UpdatePost(ctx context.Context, postID, bollocks string, tags []string) (*api.Post, error) {
	docRef := s.client.Collection("bollocks").Doc(postID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}

	var p post
	if err := docSnap.DataTo(&p); err != nil {
		return nil, err
	}
	userID, _ := api.ContextGetUserId(ctx)
	if userID != p.Author {
		return nil, errors.New("forbidden")
	}

	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "bollocks", Value: bollocks},
		{Path: "tags", Value: tags},
	}, firestore.LastUpdateTime(docSnap.UpdateTime))
	if err != nil {
		return nil, err
	}

	return &api.Post{
		ID:        docRef.ID,
		Bollocks:  bollocks,
		Tags:      tags,
		CreatedAt: p.CreatedAt,
		Likes:     len(p.Likes),
	}, nil
}

func (s *Service) ToggleLike(ctx context.Context, postID string) (*api.Post, error) {
	docRef := s.client.Collection("bollocks").Doc(postID)
	var p post
	var isCurrentlyLiked bool
	var likesCount int
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			return err
		}
		if err := doc.DataTo(&p); err != nil {
			return err
		}
		likesCount = len(p.Likes)
		userId, _ := api.ContextGetUserId(ctx)
		isCurrentlyLiked = slices.Contains(p.Likes, userId)
		var update firestore.Update
		if isCurrentlyLiked {
			likesCount--
			update = firestore.Update{Path: "likes", Value: firestore.ArrayRemove(userId)}
		} else {
			likesCount++
			update = firestore.Update{Path: "likes", Value: firestore.ArrayUnion(userId)}
		}
		return tx.Update(docRef, []firestore.Update{update})
	})
	if err != nil {
		return nil, err
	}

	return &api.Post{
		ID:        docRef.ID,
		Bollocks:  p.Bollocks,
		Tags:      p.Tags,
		CreatedAt: p.CreatedAt,
		Likes:     likesCount,
	}, nil
}
