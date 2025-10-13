package api

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/mchipperfield/gocore/log"
)

type Profile struct {
	Interests []string `json:"interests"`
}

func GetMyProfile(logger log.Logger, s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profile, err := s.GetMyProfile(r.Context())
		if err != nil {
			logger.Log("failed to get user profile", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(profile)
	}
}

func UpdateMyProfile(logger log.Logger, s Service) http.HandlerFunc {
	type request struct {
		Interests []string `json:"interests"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err.Error())
			return
		}

		// Basic sanitization
		var interests []string
		for _, interest := range req.Interests {
			cleanInterest := strings.ToLower(strings.TrimSpace(interest))
			if cleanInterest != "" {
				interests = append(interests, cleanInterest)
			}
		}
		interests = slices.Compact(interests)
		profile, err := s.UpdateMyProfile(r.Context(), interests)
		if err != nil {
			logger.Log("failed to update user profile", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(profile)
	}
}
