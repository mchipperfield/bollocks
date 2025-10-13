package firestore

import (
	"context"
	"errors"

	"github.com/mchipperfield/bollocks/api.bollocks.social/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type profile struct {
	Interests []string `firestore:"interests"`
}

func (s *Service) GetMyProfile(ctx context.Context) (*api.Profile, error) {
	userID, ok := api.ContextGetUserId(ctx)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	docRef := s.client.Collection("profiles").Doc(userID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Profile doesn't exist, return an empty one
			return &api.Profile{Interests: []string{}}, nil
		}
		return nil, err
	}

	var profile profile
	if err := docSnap.DataTo(&profile); err != nil {
		return nil, err
	}

	return &api.Profile{Interests: profile.Interests}, nil
}

func (s *Service) UpdateMyProfile(ctx context.Context, interests []string) (*api.Profile, error) {
	userID, ok := api.ContextGetUserId(ctx)
	if !ok {
		return nil, errors.New("user not found in context")
	}

	docRef := s.client.Collection("profiles").Doc(userID)
	_, err := docRef.Set(ctx, profile{Interests: interests})
	if err != nil {
		return nil, err
	}

	return &api.Profile{Interests: interests}, nil
}
