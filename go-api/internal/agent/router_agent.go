package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
)

// Recommendation is a suggested figure from the router analysis.
type Recommendation struct {
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	DrawingPrompt flexString      `json:"drawing_prompt"`
	Format        string          `json:"format,omitempty"`
	Priority      int             `json:"priority"`
}

// flexString unmarshals both a plain JSON string and a JSON object (flattened to string).
// This handles LLMs that sometimes return structured objects instead of flat strings.
type flexString string

func (f *flexString) UnmarshalJSON(data []byte) error {
	// Try plain string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = flexString(s)
		return nil
	}

	// Try object — use Decoder with Token() to preserve key insertion order.
	dec := json.NewDecoder(strings.NewReader(string(data)))
	tok, err := dec.Token()
	if err == nil {
		if delim, ok := tok.(json.Delim); ok && delim == '{' {
			var parts []string
			for dec.More() {
				keyTok, kErr := dec.Token()
				if kErr != nil {
					break
				}
				key, _ := keyTok.(string)
				var val string
				if vErr := dec.Decode(&val); vErr != nil {
					break
				}
				parts = append(parts, key+":\n"+val)
			}
			if len(parts) > 0 {
				*f = flexString(strings.Join(parts, "\n\n"))
				return nil
			}
		}
	}

	// Fallback: use raw JSON as string
	*f = flexString(string(data))
	return nil
}

func (f flexString) String() string { return string(f) }

type RouterAgent struct {
	llm    *llm.GeminiClient
	logger *zap.Logger
}

func NewRouterAgent(llm *llm.GeminiClient, logger *zap.Logger) *RouterAgent {
	return &RouterAgent{llm: llm, logger: logger}
}

// Analyze examines thesis text and returns figure recommendations.
func (a *RouterAgent) Analyze(ctx context.Context, text string, opts AgentOpts) ([]Recommendation, error) {
	sysPrompt := prompt.Router(opts.Language, opts.ThesisTitle, opts.ThesisAbstract)

	raw, err := a.llm.Generate(ctx, sysPrompt, text, defaultTemperature, opts.Model)
	if err != nil {
		return nil, fmt.Errorf("router analyze: %w", err)
	}

	a.logger.Debug("router raw LLM response", zap.String("raw", truncate(raw, 2000)))

	jsonStr, err := ParseJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("router parse json: %w (raw response: %.500s)", err, raw)
	}

	a.logger.Debug("router extracted JSON", zap.String("json", truncate(jsonStr, 2000)))

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
				a.logger.Info("router model returned single object instead of array, wrapping it")
				return []Recommendation{single}, nil
			}
		}
	}

	return nil, fmt.Errorf("router unmarshal: could not parse recommendations (json: %.500s)", jsonStr)
}

// truncate returns at most maxLen characters of s.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
