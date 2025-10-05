package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// generateTags sends content to Google's Gemini API to get tags.
func generateTags(ctx context.Context, apiKey, content string) ([]string, error) {
	if apiKey == "" {
		return nil, errors.New("no API key provided")
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

	prompt := fmt.Sprintf("Analyze the following text and generate 3-5 relevant, single-word, lowercase tags. Return the tags as a JSON array of strings. Do not include any other text or markdown in your response. Text: \"%s\"", content)

	reqBody, err := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		re := map[string]any{}
		json.NewDecoder(resp.Body).Decode(&re)
		fmt.Print(re)
		return nil, fmt.Errorf("Gemini API returned non-200 status: %s", resp.Status)

	}

	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from Gemini")
	}

	// Clean up the response text which might be wrapped in markdown
	responseText := apiResp.Candidates[0].Content.Parts[0].Text
	re := regexp.MustCompile("(?s)```json\n(.*)\n```")
	matches := re.FindStringSubmatch(responseText)
	if len(matches) > 1 {
		responseText = matches[1]
	}

	var tags []string
	if err := json.Unmarshal([]byte(responseText), &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags from Gemini content: %w", err)
	}

	return tags, nil
}

// generateTagsFromHashtags is a fallback to extract hashtags from content.
func generateTagsFromHashtags(content string) []string {
	re := regexp.MustCompile(`#(\w+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	tags := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, strings.ToLower(match[1]))
		}
	}
	slices.Sort(tags)
	return slices.Compact(tags)

}
