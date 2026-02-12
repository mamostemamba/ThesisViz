package agent

import (
	"context"
	"fmt"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
	"github.com/thesisviz/go-api/pkg/colorscheme"
)

type MermaidAgent struct {
	llm *llm.GeminiClient
}

func NewMermaidAgent(llm *llm.GeminiClient) *MermaidAgent {
	return &MermaidAgent{llm: llm}
}

func (a *MermaidAgent) Format() string { return "mermaid" }

func (a *MermaidAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	scheme := colorscheme.Get(opts.ColorScheme)
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.Mermaid(opts.Language, scheme.TikZPrompt, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, 0.4)
	if err != nil {
		return "", fmt.Errorf("mermaid generate: %w", err)
	}
	return ParseMermaid(raw)
}

func (a *MermaidAgent) Refine(ctx context.Context, code, modification string, opts AgentOpts) (string, error) {
	scheme := colorscheme.Get(opts.ColorScheme)
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.Mermaid(opts.Language, scheme.TikZPrompt, identity)

	userMsg := fmt.Sprintf("Here is the current Mermaid code:\n\n%s\n\nPlease fix these issues:\n%s\n\nOutput ONLY the complete fixed Mermaid code.", code, modification)

	raw, err := a.llm.Generate(ctx, sysPrompt, userMsg, 0.4)
	if err != nil {
		return "", fmt.Errorf("mermaid refine: %w", err)
	}
	return ParseMermaid(raw)
}

func (a *MermaidAgent) RefineWithImage(ctx context.Context, code, modification string, img []byte, opts AgentOpts) (string, error) {
	// Mermaid is rendered in the browser, so RefineWithImage just delegates to Refine.
	return a.Refine(ctx, code, modification, opts)
}
