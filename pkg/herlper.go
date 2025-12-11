package pkg

import (
	"regexp"
	"strings"
)

func GenerateSlug(company string) string {
	slug := strings.ToLower(company)
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "unknown-company"
	}
	return slug
}
