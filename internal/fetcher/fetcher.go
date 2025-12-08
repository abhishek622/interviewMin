package fetcher

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/abhishek622/interviewMin/pkg/model"
)

type FetchResult struct {
	Title   string
	URL     string
	Content string
}

func Fetch(rawURL string, source model.Source, userAgent string) (*FetchResult, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	host := strings.ToLower(u.Host)

	if strings.Contains(host, "leetcode.com") && source == model.SourceLeetcode {
		topicID, err := ParseLeetcodeDiscussURL(rawURL)
		if err != nil {
			return nil, err
		}
		// Passing a dummy CSRF token.
		res, err := GetLeetcodePost(topicID, userAgent, "jcGYFOTSHkll4nJtvZIa2Wg0YGiHfT")
		if err != nil {
			return nil, err
		}
		return &FetchResult{Title: res.Title, URL: res.URL, Content: res.Content}, nil
	} else if strings.Contains(host, "reddit.com") && source == model.SourceReddit {
		res, err := GetRedditPost(rawURL, userAgent)
		if err != nil {
			return nil, err
		}
		return &FetchResult{Title: res.Title, URL: res.URL, Content: res.Content}, nil
	} else if strings.Contains(host, "geeksforgeeks.org") && source == model.SourceGFG {
		cleanURL, err := ParseGeeksforgeeksURL(rawURL)
		if err != nil {
			return nil, err
		}
		res, err := GetGfGPost(cleanURL, userAgent)
		if err != nil {
			return nil, err
		}
		return &FetchResult{Title: res.Title, URL: res.URL, Content: res.Content}, nil
	}

	return nil, fmt.Errorf("unsupported domain: %s", host)
}
