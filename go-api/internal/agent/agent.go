package agent

import (
	"context"

	"github.com/thesisviz/go-api/pkg/colorscheme"
)

// defaultTemperature is the LLM temperature used for code generation.
const defaultTemperature float32 = 0.4

// Agent is the interface for code generation agents (TikZ, Matplotlib, Mermaid).
type Agent interface {
	// Generate creates code from a natural language prompt.
	Generate(ctx context.Context, prompt string, opts AgentOpts) (string, error)
	// Refine modifies existing code based on a modification instruction (e.g. compilation error fix).
	Refine(ctx context.Context, code, modification string, opts AgentOpts) (string, error)
	// RefineWithImage modifies code with a modification instruction and the current rendered image.
	RefineWithImage(ctx context.Context, code, modification string, img []byte, opts AgentOpts) (string, error)
	// Format returns the agent's output format name.
	Format() string
}

// AgentOpts carries context for code generation.
type AgentOpts struct {
	Language       string
	ColorScheme    string
	CustomColors   *colorscheme.CustomColors
	ThesisTitle    string
	ThesisAbstract string
	Model          string
}

// ResolveScheme returns the appropriate Scheme — from CustomColors if set, otherwise from the preset.
func (o AgentOpts) ResolveScheme() colorscheme.Scheme {
	if o.CustomColors != nil {
		return colorscheme.FromCustom(*o.CustomColors)
	}
	return colorscheme.Get(o.ColorScheme)
}

// buildIdentity constructs a thesis identity string from title and abstract.
func buildIdentity(title, abstract string) string {
	if title == "" && abstract == "" {
		return ""
	}
	identity := ""
	if title != "" {
		identity += "Thesis: " + title
	}
	if abstract != "" {
		if identity != "" {
			identity += ". "
		}
		identity += "Abstract: " + abstract
	}
	return identity
}
