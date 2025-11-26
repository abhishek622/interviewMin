package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abhishek622/interviewMin/pkg/model"
)

// Fetcher handles retrieving content from various sources
type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Fetch retrieves content from the given URL based on the source type
func (f *Fetcher) Fetch(url string, source model.Source) (string, error) {
	// Basic validation
	if url == "" {
		return "", fmt.Errorf("empty url")
	}

	// For now, we use a generic HTTP GET for all supported sources.
	// In the future, we can add specific logic for LeetCode (e.g. GraphQL), Reddit (API), etc.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	// Add a user agent to avoid being blocked by some sites
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; InterviewMinBot/1.0)")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return string(bodyBytes), nil
}
