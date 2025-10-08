package api

import (
	"regexp"
	"slices"
	"strings"
)

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
