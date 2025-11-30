package fetcher

import (
	"encoding/json"
	"net/http"
)

type Listing struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string `json:"title"`
				Permalink string `json:"permalink"`
				Selftext  string `json:"selftext"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RedditPostResponse struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func GetRedditPost(postURL, userAgent string) (RedditPostResponse, error) {
	if postURL[len(postURL)-1] != '/' {
		postURL += "/"
	}
	jsonURL := postURL + ".json"

	req, _ := http.NewRequest("GET", jsonURL, nil)
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return RedditPostResponse{}, err
	}
	defer resp.Body.Close()

	var data []Listing
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return RedditPostResponse{}, err
	}

	post := data[0].Data.Children[0].Data

	title := post.Title
	url := "https://reddit.com" + post.Permalink
	content := post.Selftext

	return RedditPostResponse{Title: title, URL: url, Content: content}, nil
}
