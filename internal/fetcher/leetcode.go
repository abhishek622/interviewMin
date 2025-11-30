package fetcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// --- Public response (only the 3 fields you asked for) ---
type LeetcodePostResponse struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

type discussionResponse struct {
	Data struct {
		UGCArticle struct {
			Title     string `json:"title"`
			Permalink string `json:"slug"` // slug is not a full permalink; we will form canonical url from returned data if available
			Content   string `json:"content"`
			// Note: API returns `ugcArticleDiscussionArticle`, map appropriately in the outer field tag below
		} `json:"ugcArticleDiscussionArticle"`
	} `json:"data"`
}

func ParseLeetcodeDiscussURL(raw string) (topicID string, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	// Validate host
	host := strings.ToLower(u.Host)
	if !(host == "leetcode.com" || host == "www.leetcode.com") {
		return "", fmt.Errorf("url host must be leetcode.com, got %s", u.Host)
	}

	// Path should start with /discuss/post/<topicId>
	// Accept optional trailing segments (slug, etc)
	re := regexp.MustCompile(`^/discuss/post/(\d+)(/.*)?$`)
	m := re.FindStringSubmatch(u.Path)
	if len(m) < 2 {
		return "", errors.New("url path is not a discuss post (expected /discuss/post/<id>/...)")
	}
	return m[1], nil
}

func GetLeetcodePost(topicID, userAgent, csrfToken string) (*LeetcodePostResponse, error) {
	graphqlURL := "https://leetcode.com/graphql/"

	graphqlBody := GraphQLRequest{
		Query: `
    query discussPostDetail($topicId: ID!) {
  ugcArticleDiscussionArticle(topicId: $topicId) {
    title
    slug
    summary
    content
    articleType
  }
}
    `,
		Variables:     map[string]interface{}{"topicId": topicID},
		OperationName: "discussPostDetail",
	}

	jsonData, err := json.Marshal(graphqlBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal graphql body: %w", err)
	}

	req, err := http.NewRequest("POST", graphqlURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", "csrftoken="+csrfToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("unexpected status %d from leetcode: %s", resp.StatusCode, string(body))
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp discussionResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode leetcode response: %w / raw: %s", err, string(respBytes))
	}

	article := apiResp.Data.UGCArticle
	canonicalURL := fmt.Sprintf("https://leetcode.com/discuss/post/%s/", topicID)
	if article.Permalink != "" {
		canonicalURL = fmt.Sprintf("https://leetcode.com/discuss/post/%s/%s", topicID, strings.Trim(article.Permalink, "/"))
	}

	return &LeetcodePostResponse{
		Title:   strings.TrimSpace(article.Title),
		URL:     canonicalURL,
		Content: strings.TrimSpace(article.Content),
	}, nil
}
