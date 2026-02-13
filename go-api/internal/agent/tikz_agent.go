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
//   Phase 2 (Coder):   Deterministic Go code renders JSON → TikZ code.
// Falls back to single-phase direct generation if the planner fails.
func (a *TikZAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	// Phase 1: Plan layout
	if opts.ProgressFn != nil {
		opts.ProgressFn("planning", "规划图表布局...", 5)
	}

	code, err := a.planLayout(ctx, userPrompt, opts)
	if err != nil {
		a.logger.Warn("planner failed, falling back to direct generation", zap.Error(err))
		if opts.ProgressFn != nil {
			opts.ProgressFn("generating", "直接生成代码...", 10)
		}
		return a.generateDirect(ctx, userPrompt, opts)
	}

	// Phase 2: Render plan → TikZ code (deterministic, no LLM call)
	if opts.ProgressFn != nil {
		opts.ProgressFn("generating", "根据布局生成代码...", 15)
	}

	return code, nil
}

// planLayout calls the LLM with the planner prompt and parses the JSON result.
// It auto-detects V2 (blocks) vs V1 (layers) schema and renders accordingly.
func (a *TikZAgent) planLayout(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	identity := resolveIdentity(opts)
	sysPrompt := prompt.TikZPlanner(opts.Language, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz plan llm: %w", err)
	}

	a.logger.Debug("planner raw response", zap.String("raw", truncate(raw, 2000)))

	jsonStr, err := ParseJSON(raw)
	if err != nil {
		return "", fmt.Errorf("tikz plan parse json: %w (raw: %.500s)", err, raw)
	}

	// Probe layout_mode first — freeflow bypasses the matrix renderer entirely
	var modeProbe struct {
		LayoutMode string `json:"layout_mode"`
	}
	_ = json.Unmarshal([]byte(jsonStr), &modeProbe)
	if modeProbe.LayoutMode == "freeflow" {
		a.logger.Info("planner classified diagram as freeflow")
		if opts.ProgressFn != nil {
			opts.ProgressFn("generating", "生成自由布局代码...", 10)
		}
		return a.generateFreeFlow(ctx, userPrompt, opts)
	}

	// Auto-detect V2 (blocks) vs V1 (layers) schema
	var probe struct {
		Blocks json.RawMessage `json:"blocks"`
		Layers json.RawMessage `json:"layers"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &probe); err != nil {
		return "", fmt.Errorf("tikz plan probe: %w", err)
	}

	// Try V2 first
	if len(probe.Blocks) > 0 {
		var planV2 TikZPlanV2
		if err := json.Unmarshal([]byte(jsonStr), &planV2); err == nil && len(planV2.Blocks) > 0 {
			a.logger.Info("planner succeeded (V2 blocks)",
				zap.Int("blocks", len(planV2.Blocks)),
				zap.Int("edges", len(planV2.Edges)),
			)
			return RenderTikZPlanV2(planV2), nil
		}
	}

	// Fallback to V1
	if len(probe.Layers) > 0 {
		var plan TikZPlan
		if err := json.Unmarshal([]byte(jsonStr), &plan); err == nil && len(plan.Layers) > 0 {
			a.logger.Info("planner succeeded (V1 layers)",
				zap.Int("layers", len(plan.Layers)),
				zap.Int("edges", len(plan.Edges)),
			)
			return RenderTikZPlan(plan), nil
		}
	}

	return "", fmt.Errorf("could not parse planner output as V2 or V1")
}

// generateFreeFlow uses a single-phase LLM call with the free-flow prompt
// for diagrams where rigid matrix alignment is harmful (sequence, swimlane, etc.).
func (a *TikZAgent) generateFreeFlow(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := resolveIdentity(opts)
	sysPrompt := prompt.TikZFreeFlow(opts.Language, scheme.TikZPrompt, identity)

	raw, err := a.llm.Generate(ctx, sysPrompt, userPrompt, defaultTemperature, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz freeflow generate: %w", err)
	}
	return ParseTikZ(raw)
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
