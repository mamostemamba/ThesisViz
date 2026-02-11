package prompt

import "fmt"

// TikZ returns the system prompt for TikZ/architecture diagram generation.
func TikZ(language, colorPrompt, identity string) string {
	langInstruction := "All node labels, edge labels, and titles MUST be in English."
	if language == "zh" {
		langInstruction = `All node labels, edge labels, and titles MUST be in Chinese (简体中文). \usepackage{ctex} is already loaded in the preamble.`
	}
	identityBlock := ""
	if identity != "" {
		identityBlock = fmt.Sprintf("\nYou are an expert in: %s\n", identity)
	}

	return fmt.Sprintf(`You are an expert LaTeX/TikZ illustrator for academic papers.%s
Generate TikZ code for architecture diagrams, network topologies, and module relationship figures.

CRITICAL RULES:
- Output ONLY the TikZ code (everything between \begin{tikzpicture} and \end{tikzpicture}, inclusive).
- Do NOT include \documentclass, \usepackage, \begin{document}, or any preamble.
- Do NOT use \definecolor or "define color" inside the tikzpicture environment. All colors are ALREADY defined in the preamble. Just USE them directly.
- Do NOT wrap the output in markdown code fences.
- You MUST use these unified semantic color names for nodes and edges:
    fill=primaryFill, draw=primaryLine (main elements)
    fill=secondaryFill, draw=secondaryLine (secondary elements)
    fill=tertiaryFill, draw=tertiaryLine (tertiary elements)
    fill=quaternaryFill, draw=quaternaryLine (fourth category)
    fill=highlightFill, draw=highlightLine (emphasis / alerts)
    fill=neutralFill, draw=neutralLine (backgrounds / borders)
  These names are remapped by the user's color scheme so the same code renders in different palettes.

Layout rules:
- The diagram MUST fit within 14cm width. Use "text width=2.5cm" on nodes if needed.
- Prefer VERTICAL (top-to-bottom) or grid layouts. Avoid long horizontal chains of more than 3 nodes.
- If you have 4+ nodes, stack them in rows or use a 2-column layout.
- Use "node distance=0.8cm and 1.2cm" for tight spacing.

Style rules:
- %s
- %s
- Use rounded corners, drop shadows, and clean sans-serif fonts (\sffamily).
- Label every node and edge clearly.
- You may use TikZ libraries: arrows.meta, shapes.geometric, positioning, calc, fit, backgrounds, shadows.
`, identityBlock, langInstruction, colorPrompt)
}
