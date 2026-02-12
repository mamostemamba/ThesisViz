package agent

import (
	"context"
	"fmt"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
)

type TikZAgent struct {
	llm *llm.GeminiClient
}

func NewTikZAgent(llm *llm.GeminiClient) *TikZAgent {
	return &TikZAgent{llm: llm}
}

func (a *TikZAgent) Format() string { return "tikz" }

func (a *TikZAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.TikZ(opts.Language, scheme.TikZPrompt, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz generate: %w", err)
	}
	return ParseTikZ(raw)
}

func (a *TikZAgent) Refine(ctx context.Context, code, modification string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.TikZ(opts.Language, scheme.TikZPrompt, identity)

	userMsg := fmt.Sprintf("Here is the current TikZ code:\n\n%s\n\nPlease fix these issues:\n%s\n\nOutput ONLY the complete fixed TikZ code.", code, modification)

	raw, err := a.llm.Generate(ctx, sysPrompt, userMsg, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz refine: %w", err)
	}
	return ParseTikZ(raw)
}

func (a *TikZAgent) RefineWithImage(ctx context.Context, code, modification string, img []byte, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.TikZ(opts.Language, scheme.TikZPrompt, identity)

	userMsg := fmt.Sprintf("Here is the current TikZ code:\n\n%s\n\nThe rendered result is shown in the attached image. Please fix these issues:\n%s\n\nOutput ONLY the complete fixed TikZ code.", code, modification)

	raw, err := a.llm.GenerateWithImage(ctx, sysPrompt, userMsg, img, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz refine with image: %w", err)
	}
	return ParseTikZ(raw)
}

