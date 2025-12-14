package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client is the Groq API client
type Client struct {
	apiKey  string
	model   string
	base    string
	http    *http.Client
	logger  *zap.Logger
	timeout time.Duration
}

// NewClient creates a new Groq API client
func NewClient(apiKey, model string, timeout time.Duration, logger *zap.Logger) *Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		apiKey:  apiKey,
		model:   model,
		base:    "https://api.groq.com/openai/v1",
		http:    &http.Client{Timeout: timeout},
		logger:  logger,
		timeout: timeout,
	}
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float32             `json:"temperature,omitempty"`
}

// ChatResponse represents a chat completion response
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

// Chat sends a chat completion request to the Groq API
func (c *Client) Chat(ctx context.Context, req ChatRequest) (string, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	url := c.base + "/chat/completions"
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.logger.Error("groq: request failed",
			zap.String("model", req.Model),
			zap.Error(err),
		)
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("groq: API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("model", req.Model),
		)
		return "", fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != nil {
		c.logger.Error("groq: API returned error",
			zap.String("error_type", chatResp.Error.Type),
			zap.String("error_message", chatResp.Error.Message),
		)
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	c.logger.Debug("groq: chat completed",
		zap.String("model", req.Model),
		zap.Int("response_length", len(chatResp.Choices[0].Message.Content)),
	)

	return chatResp.Choices[0].Message.Content, nil
}
