package fetcher

import (
	"fmt"
	"net/url"
	"strings"
)

type FetchResult struct {
	Title   string
	URL     string
	Content string
}

func Fetch(rawURL, userAgent string) (*FetchResult, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	host := strings.ToLower(u.Host)

	if strings.Contains(host, "leetcode.com") {
		topicID, err := ParseLeetcodeDiscussURL(rawURL)
		if err != nil {
			return nil, err
		}
		// Passing a dummy CSRF token. If LeetCode API enforces it strictly for public posts, this might fail.
		// However, typically public GraphQL queries might work or we might need a better strategy.
		res, err := GetLeetcodePost(topicID, userAgent, "jcGYFOTSHkll4nJtvZIa2Wg0YGiHfT")
		if err != nil {
			return nil, err
		}
		return &FetchResult{Title: res.Title, URL: res.URL, Content: res.Content}, nil
	} else if strings.Contains(host, "reddit.com") {
		res, err := GetRedditPost(rawURL, userAgent)
		if err != nil {
			return nil, err
		}
		return &FetchResult{Title: res.Title, URL: res.URL, Content: res.Content}, nil
	} else if strings.Contains(host, "geeksforgeeks.org") {
		res, err := GetGfGPost(rawURL, userAgent)
		if err != nil {
			return nil, err
		}
		return &FetchResult{Title: res.Title, URL: res.URL, Content: res.Content}, nil
	}

	return nil, fmt.Errorf("unsupported domain: %s", host)
}
