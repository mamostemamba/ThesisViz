package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
)

// thinkingBudget is the token budget for Gemini thinking mode.
const thinkingBudget int32 = 8192

type TikZAgent struct {
	llm    *llm.GeminiClient
	logger *zap.Logger

	refOnce sync.Once
	refTex  string // lazily loaded reference tex
}

func NewTikZAgent(llm *llm.GeminiClient, logger *zap.Logger) *TikZAgent {
	return &TikZAgent{llm: llm, logger: logger}
}

func (a *TikZAgent) Format() string { return "tikz" }

// Generate uses a simple prompt with thinking mode and reference tex.
// The LLM plans the layout internally via thinking, then generates TikZ code.
func (a *TikZAgent) Generate(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	if opts.ProgressFn != nil {
		opts.ProgressFn("generating", "生成图表代码...", 5)
	}

	code, err := a.generateSimple(ctx, userPrompt, opts)
	if err != nil {
		a.logger.Warn("simple generation failed, falling back to direct generation", zap.Error(err))
		if opts.ProgressFn != nil {
			opts.ProgressFn("generating", "直接生成代码...", 10)
		}
		return a.generateDirect(ctx, userPrompt, opts)
	}

	return code, nil
}

// generateSimple uses the simple prompt + thinking mode + reference tex.
// Reference tex is loaded from the git-ignored reference_tex/ directory
// and passed in the user message (not the system prompt).
func (a *TikZAgent) generateSimple(ctx context.Context, userPrompt string, opts AgentOpts) (string, error) {
	scheme := opts.ResolveScheme()
	identity := resolveIdentity(opts)

	sysPrompt := prompt.TikZSimple(opts.Language, scheme.TikZPrompt, identity)

	// Build user message: drawing request + reference tex (if available)
	var userMsg strings.Builder
	userMsg.WriteString(userPrompt)

	if refTex := a.loadReferenceTex(); refTex != "" {
		userMsg.WriteString("\n\n<ref>\n")
		userMsg.WriteString(refTex)
		userMsg.WriteString("\n</ref>")
	}

	raw, err := a.llm.GenerateWithThinking(ctx, sysPrompt, userMsg.String(), defaultTemperature, thinkingBudget, opts.Model)
	if err != nil {
		return "", fmt.Errorf("tikz simple generate: %w", err)
	}

	a.logger.Debug("simple generation response", zap.String("raw", truncate(raw, 2000)))

	return ParseTikZ(raw)
}

// loadReferenceTex lazily loads reference tex files from the reference_tex directory.
// Returns concatenated tikzpicture excerpts from up to 3 files.
func (a *TikZAgent) loadReferenceTex() string {
	a.refOnce.Do(func() {
		// Try common paths relative to working directory
		candidates := []string{"reference_tex", "go-api/reference_tex"}
		for _, base := range candidates {
			entries, err := os.ReadDir(base)
			if err != nil {
				continue
			}

			var refs []string
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".tex") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(base, e.Name()))
				if err != nil {
					continue
				}
				refs = append(refs, string(data))
				if len(refs) >= 3 {
					break
				}
			}
			if len(refs) > 0 {
				a.refTex = strings.Join(refs, "\n\n% --- next reference ---\n\n")
				a.logger.Info("loaded reference tex files", zap.Int("count", len(refs)), zap.String("dir", base))
				return
			}
		}
		a.logger.Warn("no reference tex files found")
	})
	return a.refTex
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
