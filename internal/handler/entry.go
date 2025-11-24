package handler

import (
	"encoding/json"
	"net/http"

	"github.com/abhishek622/interviewMin/internal/openai"
	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/gin-gonic/gin"
)

// ConvertInterview calls OpenAI to convert raw_text -> structured entry and saves it
// POST /api/v1/interviews/:id/convert
func (app *Application) ConvertInterview(c *gin.Context) {
	interviewID := c.Param("id")
	if interviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing interview id"})
		return
	}
	var req model.ConvertInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Check if it's a binding error (allow empty body)
		if e, ok := err.(gin.Error); ok && e.Type == gin.ErrorTypeBind {
			app.Logger.Sugar().Warnw("convert interview bind error", "err", err)
		} else {
			// Non-binding error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := c.Request.Context()
	interview, err := app.InterviewRepo.GetByID(ctx, interviewID)
	if err != nil {
		app.Logger.Sugar().Errorw("convert interview: fetch failed", "id", interviewID, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "interview not found"})
		return
	}
	if interview.UserID != userID.(string) {
		app.Logger.Sugar().Warnw("convert interview: user mismatch", "user", userID, "owner", interview.UserID)
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Build deterministic prompt - ask for strict JSON output with schema example
	systemMsg := "You are a JSON extractor. Given raw interview text, output ONLY valid JSON that matches the required schema. Do not add any other text."
	userPrompt := buildConvertPrompt(interview.RawText, req.OverrideTitle)

	chatReq := openai.ChatRequest{
		Model:       app.OpenAIModel,
		Messages:    []map[string]string{{"role": "system", "content": systemMsg}, {"role": "user", "content": userPrompt}},
		MaxTokens:   1500,
		Temperature: 0.0,
	}

	respStr, err := app.OpenAI.Chat(ctx, chatReq)
	if err != nil {
		app.Logger.Sugar().Errorw("openai chat error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "openai error"})
		return
	}

	// Parse response as JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(respStr), &parsed); err != nil {
		app.Logger.Sugar().Errorw("failed to parse openai response", "err", err, "resp", respStr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse openai response"})
		return
	}

	// Derive a title (either override or from parsed)
	title := req.OverrideTitle
	if title == "" {
		if t, ok := parsed["title"].(string); ok && t != "" {
			title = t
		} else {
			// fallback: interview title or truncated text
			if interview.Title != "" {
				title = interview.Title
			} else {
				txt := interview.RawText
				if len(txt) > 100 {
					title = txt[:100]
				} else {
					title = txt
				}
			}
		}
	}

	// Save entry
	entryID, err := app.EntryRepo.Create(ctx, interviewID, title, parsed)
	if err != nil {
		app.Logger.Sugar().Errorw("entry create failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entry_id": entryID})
}

// GetEntry returns an entry by id
// GET /api/v1/entries/:id
func (app *Application) GetEntry(c *gin.Context) {
	entryID := c.Param("id")
	if entryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}
	_, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx := c.Request.Context()
	entry, err := app.EntryRepo.GetByID(ctx, entryID)
	if err != nil {
		app.Logger.Sugar().Errorw("get entry error", "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	// We only have interview_id on entry; ideally check ownership via InterviewRepo
	// If you want strict ownership enforcement, fetch interview and compare user ids.
	c.JSON(http.StatusOK, entry)
}

// ListEntries returns paginated entries for current user
// GET /api/v1/entries
func (app *Application) ListEntries(c *gin.Context) {
	var q model.ListEntriesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		app.Logger.Sugar().Warnw("list entries bad query", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit := q.PageSize
	offset := (q.Page - 1) * q.PageSize
	ctx := c.Request.Context()
	entries, total, err := app.EntryRepo.ListByUser(ctx, userID.(string), limit, offset)
	if err != nil {
		app.Logger.Sugar().Errorw("list entries repo error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       entries,
		"total":      total,
		"page":       q.Page,
		"page_size":  q.PageSize,
		"totalPages": totalPages,
	})
}

// --- Helpers ---

func buildConvertPrompt(rawText string, overrideTitle string) string {
	// Define desired schema in the prompt. Keep this strict and minimal.
	schemaExample := `
Output JSON schema:
{
  "title": "short title",
  "problems": [
    {
      "title": "problem title",
      "difficulty": "easy|medium|hard",
      "tags": ["arrays","dp"],
      "body": "problem statement",
      "solution": {
         "explanation": "short explanation",
         "code": [
           {
             "lang":"go",
             "snippet": "code here"
           }
         ]
      }
    }
  ],
  "notes": "optional notes"
}
`
	// We append overrideTitle so the model can prefer it.
	prompt := "Convert the raw interview text into JSON. " + schemaExample + "\nRaw interview:\n" + rawText
	if overrideTitle != "" {
		prompt += "\nOverride title: " + overrideTitle
	}
	prompt += "\nReturn ONLY valid JSON. Do not wrap in markdown. Do not add commentary."
	return prompt
}
