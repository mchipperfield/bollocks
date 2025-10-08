package genai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Service struct {
	client *genai.Client
}

func NewService(ctx context.Context, apiKey string) (*Service, error) {
	if apiKey == "" {
		return nil, errors.New("no API key provided")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Service{
		client: client,
	}, nil
}

func (s *Service) GenerateTags(ctx context.Context, content string) ([]string, error) {
	prompt := fmt.Sprintf("Analyze the following text and generate 3-5 relevant, single-word, lowercase tags. Return the tags as a JSON array of strings. Do not include any other text or markdown in your response. Text: \"%s\"", content)

	model := s.client.GenerativeModel("gemini-2.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no content returned from Gemini")
	}

	// Clean up the response text which might be wrapped in markdown
	responseText := resp.Candidates[0].Content.Parts[0].(genai.Text)

	var tags []string
	if err := json.Unmarshal([]byte(responseText), &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags from Gemini content: %w", err)
	}

	return tags, nil
}
