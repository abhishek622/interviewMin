package fetcher

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

	title := strings.TrimSpace(doc.Find("h1.entry-title").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

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

	content := strings.Join(contentParts, "\n\n")
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

func ParseGeeksforgeeksURL(raw string) (cleanURL string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	host := strings.ToLower(u.Host)
	if !(host == "geeksforgeeks.org" || host == "www.geeksforgeeks.org") {
		return "", fmt.Errorf("url host must be geeksforgeeks.org, got %s", u.Host)
	}

	// Validate path format
	// e.g - /interview-experiences/infosys-interview-experience-for-systems-engineer-trainee/
	// Pattern: /interview-experiences/{topic-name}/
	re := regexp.MustCompile(`^/interview-experiences/([\w-]+)/?$`)
	m := re.FindStringSubmatch(u.Path)
	if len(m) < 2 {
		return "", errors.New("url path is not a valid geeksforgeeks interview experience url")
	}

	topicName := m[1]
	cleanURL = fmt.Sprintf("%s://%s/interview-experiences/%s/", u.Scheme, u.Host, topicName)

	return cleanURL, nil
}
