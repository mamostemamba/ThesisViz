package agent

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
)

type TikZAgent struct {
	llm    *llm.GeminiClient
	logger *zap.Logger
}

func NewTikZAgent(llm *llm.GeminiClient, logger *zap.Logger) *TikZAgent {
	return &TikZAgent{llm: llm, logger: logger}
}

func (a *TikZAgent) Format() string { return "tikz" }

// Generate implements the two-step chain:
//
//	Phase 1 (Layout Planning): LLM outputs a structured JSON describing rows, columns, nodes, edges.
//	Phase 2 (Code Generation): LLM maps the JSON to TikZ \matrix code.
func (a *TikZAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	// --- Phase 1: Layout Planning ---
	layoutJSON, planErr := a.planLayout(ctx, userPrompt, opts)
	if planErr != nil {
		a.logger.Warn("layout planning failed, falling back to direct generation", zap.Error(planErr))
		return a.directGenerate(ctx, userPrompt, opts)
	}

	a.logger.Info("layout plan generated", zap.Int("json_len", len(layoutJSON)))

	// --- Phase 2: Code Generation from layout ---
	code, err := a.generateFromLayout(ctx, userPrompt, layoutJSON, opts)
	if err != nil {
		a.logger.Warn("generation from layout failed, falling back to direct generation", zap.Error(err))
		return a.directGenerate(ctx, userPrompt, opts)
	}

	return code, nil
}

// planLayout calls the LLM to produce a layout JSON.
func (a *TikZAgent) planLayout(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	sysPrompt := prompt.TikZLayoutPlan(opts.Language)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, 0.3, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz plan layout: %w", err)
	}

	// Extract JSON from the response (LLM may wrap it in markdown fences)
	jsonStr, err := ParseJSON(raw)
	if err != nil {
		// If ParseJSON fails, try using raw response as-is (it might already be clean JSON)
		jsonStr = strings.TrimSpace(raw)
	}

	if len(jsonStr) < 10 {
		return "", fmt.Errorf("tikz plan layout: response too short (%d chars)", len(jsonStr))
	}

	return jsonStr, nil
}

// generateFromLayout feeds the layout JSON to the TikZ code generation prompt.
func (a *TikZAgent) generateFromLayout(ctx context.Context, userPrompt, layoutJSON string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.TikZFromLayout(opts.Language, scheme.TikZPrompt, identity)

	userMsg := fmt.Sprintf("Original drawing request:\n%s\n\n=== LAYOUT JSON (follow this structure strictly) ===\n%s", userPrompt, layoutJSON)

	raw, err := a.llm.Generate(ctx, sysPrompt, userMsg, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz generate from layout: %w", err)
	}

	return ParseTikZ(raw)
}

// directGenerate is the single-step fallback (original behavior).
func (a *TikZAgent) directGenerate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
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

	userMsg := fmt.Sprintf("Here is the current TikZ code:\n\n%s\n\nPlease fix these issues:\n%s", code, modification)

	raw, err := a.llm.Generate(ctx, sysPrompt, userMsg, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz refine: %w", err)
	}
	return ParseTikZ(extractFixedCode(raw))
}

func (a *TikZAgent) RefineWithImage(ctx context.Context, code, modification string, img []byte, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := buildIdentity(opts.ThesisTitle, opts.ThesisAbstract)
	sysPrompt := prompt.TikZ(opts.Language, scheme.TikZPrompt, identity)

	userMsg := fmt.Sprintf("Here is the current TikZ code:\n\n%s\n\nThe rendered result is shown in the attached image. Please fix these issues:\n%s", code, modification)

	raw, err := a.llm.GenerateWithImage(ctx, sysPrompt, userMsg, img, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz refine with image: %w", err)
	}
	return ParseTikZ(extractFixedCode(raw))
}

// extractFixedCode pulls the code section after "=== FIXED CODE ===" marker.
// If the marker is absent, returns the full text unchanged (backward compatible).
func extractFixedCode(raw string) string {
	const marker = "=== FIXED CODE ==="
	if idx := strings.Index(raw, marker); idx >= 0 {
		return raw[idx+len(marker):]
	}
	return raw
}
