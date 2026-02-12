package agent

import (
	"context"
	"encoding/json"
	"fmt"

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

	raw, err := a.llm.Generate(ctx, sysPrompt, text, 0.4)
	if err != nil {
		return nil, fmt.Errorf("router analyze: %w", err)
	}

	jsonStr, err := ParseJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("router parse json: %w", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(jsonStr), &recs); err != nil {
		return nil, fmt.Errorf("router unmarshal: %w", err)
	}
	return recs, nil
}
