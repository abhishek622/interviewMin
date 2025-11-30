package fetcher

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type GfGPostResponse struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func GetGfGPost(pageURL, userAgent string) (GfGPostResponse, error) {
	req, _ := http.NewRequest("GET", pageURL, nil)
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

	title := strings.TrimSpace(doc.Find("h1").First().Text())
	url := pageURL
	var content string

	// selectors to try in order
	selectors := []string{"article", ".entry-content", "#content", ".content", ".post-content"}
	for _, sel := range selectors {
		selection := doc.Find(sel).First()
		if selection.Length() > 0 {
			text := strings.TrimSpace(selection.Text())
			if len(text) > 50 {
				content = text
				break
			}
		}
	}

	if content == "" {
		// fallback: whole body text
		content = strings.TrimSpace(doc.Find("body").Text())
		if len(content) > 20000 {
			content = content[:20000]
		}
	}

	return GfGPostResponse{Title: title, URL: url, Content: content}, nil
}
