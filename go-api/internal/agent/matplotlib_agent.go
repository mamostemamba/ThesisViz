package agent

import (
	"context"
	"fmt"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
	"github.com/thesisviz/go-api/pkg/colorscheme"
)

type MatplotlibAgent struct {
	llm *llm.GeminiClient
}

func NewMatplotlibAgent(llm *llm.GeminiClient) *MatplotlibAgent {
	return &MatplotlibAgent{llm: llm}
}

func (a *MatplotlibAgent) Format() string { return "matplotlib" }

func (a *MatplotlibAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	scheme := colorscheme.Get(opts.ColorScheme)
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.Matplotlib(opts.Language, scheme.TikZPrompt, scheme.MatplotlibColors, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, 0.4)
	if err != nil {
		return "", fmt.Errorf("matplotlib generate: %w", err)
	}
	return ParseMatplotlib(raw)
}

func (a *MatplotlibAgent) Refine(ctx context.Context, code, modification string, opts AgentOpts) (string, error) {
	scheme := colorscheme.Get(opts.ColorScheme)
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.Matplotlib(opts.Language, scheme.TikZPrompt, scheme.MatplotlibColors, identity)

	userMsg := fmt.Sprintf("Here is the current Python/Matplotlib code:\n\n%s\n\nPlease fix these issues:\n%s\n\nOutput ONLY the complete fixed Python code.", code, modification)

	raw, err := a.llm.Generate(ctx, sysPrompt, userMsg, 0.4)
	if err != nil {
		return "", fmt.Errorf("matplotlib refine: %w", err)
	}
	return ParseMatplotlib(raw)
}

func (a *MatplotlibAgent) RefineWithImage(ctx context.Context, code, modification string, img []byte, opts AgentOpts) (string, error) {
	scheme := colorscheme.Get(opts.ColorScheme)
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.Matplotlib(opts.Language, scheme.TikZPrompt, scheme.MatplotlibColors, identity)

	userMsg := fmt.Sprintf("Here is the current Python/Matplotlib code:\n\n%s\n\nThe rendered result is shown in the attached image. Please fix these issues:\n%s\n\nOutput ONLY the complete fixed Python code.", code, modification)

	raw, err := a.llm.GenerateWithImage(ctx, sysPrompt, userMsg, img, 0.4)
	if err != nil {
		return "", fmt.Errorf("matplotlib refine with image: %w", err)
	}
	return ParseMatplotlib(raw)
}
