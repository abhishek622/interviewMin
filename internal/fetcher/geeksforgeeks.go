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

	if lastUpdated == "" {
		dateText := doc.Find(".date, .updated, .post-date").First().Text()
		lastUpdated = strings.TrimSpace(dateText)
	}

	// Extract main content
	var contentBuilder strings.Builder
	articleContent := doc.Find("div.article--viewer_content, div.entry-content, article .content").First()

	if articleContent.Length() > 0 {
		// Remove unwanted elements
		articleContent.Find("script, style, nav, header, footer, .ad, .advertisement").Remove()

		// Process each content element
		articleContent.Find("h1, h2, h3, h4, h5, h6, p, ul, ol, pre, a").Each(func(i int, s *goquery.Selection) {
			processElement(s, &contentBuilder)
		})
	}

	content := cleanFinalContent(contentBuilder.String())

	return GfGPostResponse{
		Title:       title,
		URL:         pageURL,
		Content:     content,
		LastUpdated: lastUpdated,
	}, nil
}

// processElement handles different HTML elements and formats them appropriately
func processElement(s *goquery.Selection, builder *strings.Builder) {
	nodeName := goquery.NodeName(s)

	switch nodeName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		text := cleanInlineText(s.Text())
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(text)
			builder.WriteString("\n")
		}

	case "p":
		text := processInlineElements(s)
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString(text)
		}

	case "ul", "ol":
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") {
			builder.WriteString("\n")
		}
		s.Find("li").Each(func(i int, li *goquery.Selection) {
			text := processInlineElements(li)
			if text != "" {
				builder.WriteString(" - ")
				builder.WriteString(text)
				builder.WriteString("\n")
			}
		})

	case "pre":
		code := s.Find("code").Text()
		if code == "" {
			code = s.Text()
		}
		code = strings.TrimSpace(code)
		if code != "" {
			if builder.Len() > 0 {
				builder.WriteString("\n\n")
			}
			builder.WriteString("```\n")
			builder.WriteString(code)
			builder.WriteString("\n```")
		}
	}
}

// processInlineElements handles text content with inline elements like links
func processInlineElements(s *goquery.Selection) string {
	var result strings.Builder

	s.Contents().Each(func(i int, node *goquery.Selection) {
		if goquery.NodeName(node) == "a" {
			linkText := cleanInlineText(node.Text())
			href, exists := node.Attr("href")
			if linkText != "" && exists && href != "" {
				result.WriteString(`"`)
				result.WriteString(linkText)
				result.WriteString(`"[`)
				result.WriteString(href)
				result.WriteString(`]`)
			} else if linkText != "" {
				result.WriteString(linkText)
			}
		} else {
			text := node.Text()
			if text != "" {
				result.WriteString(cleanInlineText(text))
			}
		}
	})

	return strings.TrimSpace(result.String())
}

// cleanInlineText cleans text while preserving single spaces
func cleanInlineText(text string) string {
	// Replace tabs and multiple spaces with single space
	re := regexp.MustCompile(`[\t ]+`)
	text = re.ReplaceAllString(text, " ")

	// Replace multiple newlines with single space
	re = regexp.MustCompile(`\n+`)
	text = re.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// cleanFinalContent performs final cleanup on the entire content
func cleanFinalContent(content string) string {
	// Remove excessive blank lines (more than 2 consecutive newlines)
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	// Remove trailing spaces from each line
	lines := strings.Split(content, "\n")
	var cleanedLines []string
	for _, line := range lines {
		cleanedLines = append(cleanedLines, strings.TrimRight(line, " \t"))
	}

	return strings.TrimSpace(strings.Join(cleanedLines, "\n"))
}

func ParseGeeksforgeeksURL(raw string) (cleanURL string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	// Validate scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("url scheme must be http or https, got %s", u.Scheme)
	}

	// Validate host
	host := strings.ToLower(u.Host)
	if !(host == "geeksforgeeks.org" || host == "www.geeksforgeeks.org") {
		return "", fmt.Errorf("url host must be geeksforgeeks.org, got %s", u.Host)
	}

	// Validate path format
	re := regexp.MustCompile(`^/interview-experiences/([\w-]+)/?$`)
	m := re.FindStringSubmatch(u.Path)
	if len(m) < 2 {
		return "", errors.New("url path is not a valid geeksforgeeks interview experience url")
	}

	topicName := m[1]

	// Construct clean URL without query parameters
	cleanURL = fmt.Sprintf("%s://%s/interview-experiences/%s/", u.Scheme, u.Host, topicName)

	return cleanURL, nil
}
