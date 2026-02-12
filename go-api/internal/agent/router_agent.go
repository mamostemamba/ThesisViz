package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
)

// Recommendation is a suggested figure from the router analysis.
type Recommendation struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	DrawingPrompt string `json:"drawing_prompt"`
	Format        string `json:"format,omitempty"`
	Priority      int    `json:"priority"`
}

type RouterAgent struct {
	llm *llm.GeminiClient
}

func NewRouterAgent(llm *llm.GeminiClient) *RouterAgent {
	return &RouterAgent{llm: llm}
}

// Analyze examines thesis text and returns figure recommendations.
func (a *RouterAgent) Analyze(ctx context.Context, text string, opts AgentOpts) ([]Recommendation, error) {
	sysPrompt := prompt.Router(opts.Language, opts.ThesisTitle, opts.ThesisAbstract)

	raw, err := a.llm.Generate(ctx, sysPrompt, text, 0.4, opts.Model)
	if err != nil {
		return nil, fmt.Errorf("router analyze: %w", err)
	}

	log.Printf("[router] raw LLM response (first 2000 chars): %.2000s", raw)

	jsonStr, err := ParseJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("router parse json: %w (raw response: %.500s)", err, raw)
	}

	log.Printf("[router] extracted JSON (first 2000 chars): %.2000s", jsonStr)

	// Try array first
	var recs []Recommendation
	if err := json.Unmarshal([]byte(jsonStr), &recs); err == nil && len(recs) > 0 {
		return recs, nil
	}

	// Try object wrapper (LLM may return {"recommendations": [...]})
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &wrapper); err == nil {
		// Look for an array value inside the object
		for _, v := range wrapper {
			if json.Unmarshal(v, &recs) == nil && len(recs) > 0 {
				return recs, nil
			}
		}
		// The object might be a single recommendation (has "title" and "drawing_prompt")
		if _, hasTitle := wrapper["title"]; hasTitle {
			var single Recommendation
			if err := json.Unmarshal([]byte(jsonStr), &single); err == nil && single.Title != "" {
				log.Printf("[router] model returned single object instead of array, wrapping it")
				return []Recommendation{single}, nil
			}
		}
	}

	return nil, fmt.Errorf("router unmarshal: could not parse recommendations (json: %.500s)", jsonStr)
}
