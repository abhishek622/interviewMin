package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	apiKey string
	base   string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		base:   "https://api.openai.com/v1",
		http:   &http.Client{},
	}
}

type ChatRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float32             `json:"temperature,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (string, error) {
	url := c.base + "/chat/completions"
	b, _ := json.Marshal(req)
	r, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	r.Header.Set("Authorization", "Bearer "+c.apiKey)
	r.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var ch ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return "", err
	}
	if len(ch.Choices) == 0 {
		return "", fmt.Errorf("no choices")
	}
	return ch.Choices[0].Message.Content, nil
}
