package agent

import "context"

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
	ThesisTitle    string
	ThesisAbstract string
}
