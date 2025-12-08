package fetcher

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type GfGPostResponse struct {
	Title       string
	URL         string
	Content     string
	LastUpdated string
}

func GetGfGPost(pageURL, userAgent string) (GfGPostResponse, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return GfGPostResponse{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GfGPostResponse{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return GfGPostResponse{}, err
	}

	// Extract title
	title := strings.TrimSpace(doc.Find("h1.entry-title").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Extract Last Updated date
	lastUpdated := ""
	doc.Find("div.article--viewer_content time").Each(func(i int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			lastUpdated = text
		}
	})

	// If not found, try alternative selectors
	if lastUpdated == "" {
		dateText := doc.Find(".date, .updated, .post-date").First().Text()
		lastUpdated = strings.TrimSpace(dateText)
	}

	// Extract main content from article body
	var contentParts []string

	// Target the main article content area
	articleContent := doc.Find("div.article--viewer_content, div.entry-content, article .content").First()

	if articleContent.Length() > 0 {
		// Remove unwanted elements before extracting text
		articleContent.Find("script, style, nav, header, footer, .ad, .advertisement").Remove()

		// Extract paragraphs and headings
		articleContent.Find("p, h2, h3, h4, ul, ol, pre").Each(func(i int, s *goquery.Selection) {
			text := cleanText(s.Text())
			if text != "" {
				contentParts = append(contentParts, text)
			}
		})
	}

	// Join content with proper spacing
	content := strings.Join(contentParts, "\n\n")

	// Final cleanup
	content = cleanFinalContent(content)

	return GfGPostResponse{
		Title:       title,
		URL:         pageURL,
		Content:     content,
		LastUpdated: lastUpdated,
	}, nil
}

// cleanText removes excessive whitespace and newlines
func cleanText(text string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`[ \t]+`)
	text = re.ReplaceAllString(text, " ")

	// Replace multiple newlines with single newline
	re = regexp.MustCompile(`\n+`)
	text = re.ReplaceAllString(text, "\n")

	// Trim whitespace from each line
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// cleanFinalContent performs final cleanup on the entire content
func cleanFinalContent(content string) string {
	// Remove excessive blank lines (more than 2 consecutive newlines)
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}

// Example usage
func main() {
	url := "https://www.geeksforgeeks.org/interview-experiences/lti-mindtree-interview-experience-on-campus-2024/"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	result, err := GetGfGPost(url, userAgent)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Last Updated: %s\n", result.LastUpdated)
	fmt.Printf("\nContent:\n%s\n", result.Content)
}
