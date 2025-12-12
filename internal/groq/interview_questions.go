package groq

import (
	"context"
	"encoding/json"
	"fmt"
)

type ExtractedQuestions struct {
	Question string `json:"question"`
	Type     string `json:"type"`
}

func (c *Client) InterviewQuestions(ctx context.Context, content string) (*[]ExtractedQuestions, error) {
	systemMsg := `You are a precise question extractor. Your task is to read the provided interview experience and output ONLY a valid JSON array of questions, with no additional text, markdown, or explanation.

Each item must be an object with:
- "question": the exact question text as it appears or is clearly implied
- "type": one of "dsa", "system_design", "behavioral", or "other"

Rules:
- Include ONLY actual interview questions.
- NEVER invent, paraphrase, or hallucinate.
- If no questions are found, return an empty array: []
- Classify strictly:
    • "dsa" for data structures, algorithms, coding problems
    • "system_design" for system design or architecture questions
    • "behavioral" for HR, teamwork, leadership, or soft-skill questions
    • "other" for anything else (e.g., domain knowledge, puzzles)
- Output must be valid JSON. No prefix, suffix, or backticks.
`

	userPrompt := fmt.Sprintf("Interview experience:\n%s", content)
	if len(userPrompt) > 10000 {
		userPrompt = userPrompt[:10000]
	}

	chatReq := ChatRequest{
		Messages: []map[string]string{
			{"role": "system", "content": systemMsg},
			{"role": "user", "content": userPrompt},
		},
		MaxTokens:   2000,
		Temperature: 0.0,
	}

	respStr, err := c.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	var extracted []ExtractedQuestions
	if err := json.Unmarshal([]byte(respStr), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON array of questions: %w; raw response: %q", err, respStr)
	}

	return &extracted, nil
}
