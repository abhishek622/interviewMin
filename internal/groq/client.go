package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	apiKey string
	model  string
	base   string
	http   *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		base:   "https://api.groq.com/openai/v1",
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
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (string, error) {
	if req.Model == "" {
		req.Model = c.model
	}

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

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println(string(bodyBytes))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("groq api error: %s", string(bodyBytes))
	}

	var ch ChatResponse
	if err := json.Unmarshal(bodyBytes, &ch); err != nil {
		return "", fmt.Errorf("decode error: %w, body: %s", err, string(bodyBytes))
	}

	if ch.Error != nil {
		return "", fmt.Errorf("api error: %s", ch.Error.Message)
	}

	if len(ch.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	return ch.Choices[0].Message.Content, nil
}
