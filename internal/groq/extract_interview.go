package groq

import (
	"context"
	"encoding/json"
	"fmt"
)

type ExtractedData struct {
	Title     string `json:"title"`
	Company   string `json:"company"`
	Position  string `json:"position"`
	Location  string `json:"location"`
	NoOfRound int    `json:"no_of_round"`
	Questions []struct {
		Question string `json:"question"`
		Type     string `json:"type"`
	} `json:"questions"`
	FullExperience string `json:"full_experience"`
}

func (c *Client) ExtractInterview(ctx context.Context, content string) (*ExtractedData, error) {
	systemMsg := `
You are an advanced Interview Experience Extraction Engine.

Your ONLY job is to read raw interview text and convert it into a STRICTLY FORMATTED JSON object.

### IMPORTANT RULES
1. Output **only valid JSON**. No explanation, no markdown, no backticks.
2. If a field is **not present** or cannot be inferred with high confidence:
   - Use "" for strings
   - Use 0 for numbers
   - Use null for objects
3. Never invent or hallucinate company names, locations, or questions.
4. If the text contains multiple companies or rounds, pick the **main one**, usually mentioned first.

---

### Extraction Goals
From the provided interview text, extract:

- **title**: short title for the experience (ex: "Amazon SDE1 Interview Experience")
- **company**: company name (exact as mentioned)
- **position**: role or profile (SDE, Backend Engineer, Data Engineer, etc.)
- **location**: city or country if mentioned
- **no_of_round**: number of interview rounds (int). If unclear, 0.
- **questions**: a list of extracted questions. These may be coding, DSA, behavioral, or system design.

---

### Question Extraction Rules
When extracting questions:

1. Identify ANY sentences that represent a question asked in interview rounds.
2. If no questions exist → return an empty array: "questions": []
3. Each question must follow:

{
  "question": "string",
  "type": "dsa" | "system_design" | "behavioral" | "other"
}

Rules:
- Classify as **dsa** if it includes algorithms, data structures, LeetCode problems.
- Classify as **system_design** for design/architecture topics.
- Classify as **behavioral** for HR, leadership, soft skills.
- Else, classify as **other**.

---

### Final Expected JSON Schema
{
  "title": "string",
  "company": "string",
  "position": "string",
  "location": "string",
  "no_of_round": 0,
  "questions": [
    {
      "question": "string",
      "type": "dsa | system_design | behavioral | other"
    }
  ],
}

---

Think step-by-step, but RETURN ONLY THE FINAL JSON.
`

	userPrompt := fmt.Sprintf(`
Read the following raw interview text and extract the structured data exactly
as per the schema described. Remember to follow all rules strictly.

TEXT START:
%s
TEXT END
`, content)
	if len(userPrompt) > 10000 {
		userPrompt = userPrompt[:10000]
	}

	chatReq := ChatRequest{
		Messages:    []map[string]string{{"role": "system", "content": systemMsg}, {"role": "user", "content": userPrompt}},
		MaxTokens:   2000,
		Temperature: 0.0,
	}

	respStr, err := c.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	var extracted ExtractedData
	if err := json.Unmarshal([]byte(respStr), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse ai response: %w", err)
	}

	return &extracted, nil
}
