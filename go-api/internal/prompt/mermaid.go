package prompt

import (
	"fmt"
	"strings"
)

// Mermaid returns the system prompt for Mermaid flowchart/sequence diagram generation.
func Mermaid(language, colorPrompt, identity string) string {
	langInstruction := "All labels and text in the diagram MUST be in English."
	if language == "zh" {
		langInstruction = "All labels and text in the diagram MUST be in Chinese (简体中文)."
	}
	mermaidTheme := "neutral"
	if strings.Contains(strings.ToLower(colorPrompt), "draw.io") {
		mermaidTheme = "default"
	}
	identityBlock := ""
	if identity != "" {
		identityBlock = fmt.Sprintf("\nYou are an expert in: %s\n", identity)
	}

	return fmt.Sprintf(`You are an expert at creating flowcharts and sequence diagrams for academic papers using Mermaid.js syntax.%s

Requirements:
- Output ONLY the Mermaid diagram code (starting with graph, flowchart, sequenceDiagram, stateDiagram, etc.).
- Do NOT include markdown code fences (no `+"`"+`mermaid ... `+"`"+`).
- %s
- For flowcharts, prefer top-to-bottom (TB) layout.
- Use descriptive node IDs (e.g., A[Input Data] instead of A[A]).
- Keep the diagram readable: no more than ~15 nodes for flowcharts.
- For sequence diagrams, clearly label participants and messages.
- The output should be valid Mermaid syntax that renders without errors.
- Preferred Mermaid theme: %s.
`, identityBlock, langInstruction, mermaidTheme)
}
