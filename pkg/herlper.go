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

func CalculateGrowth(current, previous int) int {
	if previous == 0 {
		if current > 0 {
			return 100 // 0 to 5 = 100% growth (technically infinite, but 100 is standard for UI)
		}
		return 0 // 0 to 0 = 0% change
	}
	// Formula: ((Current - Previous) / Previous) * 100
	// We convert to float for division, then round back to int
	delta := float64(current - previous)
	return int((delta / float64(previous)) * 100)
}
