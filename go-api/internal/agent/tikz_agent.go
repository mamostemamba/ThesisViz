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

type TikZAgent struct {
	llm    *llm.GeminiClient
	logger *zap.Logger
}

func NewTikZAgent(llm *llm.GeminiClient, logger *zap.Logger) *TikZAgent {
	return &TikZAgent{llm: llm, logger: logger}
}

func (a *TikZAgent) Format() string { return "tikz" }

// Generate uses a two-phase pipeline:
//   Phase 1 (Planner): LLM outputs a JSON layout specification.
//   Phase 2 (Coder):   Deterministic Go code renders JSON → TikZ \matrix template.
// Falls back to single-phase direct generation if the planner fails.
func (a *TikZAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	// Phase 1: Plan layout
	if opts.ProgressFn != nil {
		opts.ProgressFn("planning", "规划图表布局...", 5)
	}

	plan, err := a.planLayout(ctx, userPrompt, opts)
	if err != nil {
		a.logger.Warn("planner failed, falling back to direct generation", zap.Error(err))
		if opts.ProgressFn != nil {
			opts.ProgressFn("generating", "直接生成代码...", 10)
		}
		return a.generateDirect(ctx, userPrompt, opts)
	}

	a.logger.Info("planner succeeded",
		zap.Int("layers", len(plan.Layers)),
		zap.Int("edges", len(plan.Edges)),
	)

	// Phase 2: Render plan → TikZ code (deterministic, no LLM call)
	if opts.ProgressFn != nil {
		opts.ProgressFn("generating", "根据布局生成代码...", 15)
	}

	code := RenderTikZPlan(*plan)
	return code, nil
}

// planLayout calls the LLM with the planner prompt and parses the JSON result.
func (a *TikZAgent) planLayout(ctx context.Context, userPrompt string, opts AgentOpts) (*TikZPlan, error) {
	identity := resolveIdentity(opts)
	sysPrompt := prompt.TikZPlanner(opts.Language, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, defaultTemperature, opts.Model)
	if err != nil {
		return nil, fmt.Errorf("tikz plan llm: %w", err)
	}

	a.logger.Debug("planner raw response", zap.String("raw", truncate(raw, 2000)))

	jsonStr, err := ParseJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("tikz plan parse json: %w (raw: %.500s)", err, raw)
	}

	var plan TikZPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("tikz plan unmarshal: %w", err)
	}

	if len(plan.Layers) == 0 {
		return nil, fmt.Errorf("tikz plan has no layers")
	}

	return &plan, nil
}

// generateDirect is the fallback single-phase generation (original approach).
func (a *TikZAgent) generateDirect(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := resolveIdentity(opts)
	sysPrompt := prompt.TikZ(opts.Language, scheme.TikZPrompt, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz generate: %w", err)
	}
	return ParseTikZ(raw)
}

func (a *TikZAgent) Refine(ctx context.Context, code, modification string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := resolveIdentity(opts)
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
	identity := resolveIdentity(opts)
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
