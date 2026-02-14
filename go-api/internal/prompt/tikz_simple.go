package prompt

import "fmt"

// TikZSimple returns a minimal system prompt for TikZ generation.
// Reference tex is NOT included here — it is passed separately in the user message
// by the agent, loaded at runtime from a git-ignored directory.
func TikZSimple(language, colorPrompt, identity string) string {
	langConstraint := "Use English for all labels and text."
	if language == "zh" {
		langConstraint = `使用简体中文。\usepackage{ctex} is already loaded in the preamble.`
	}

	identityBlock := ""
	if identity != "" {
		identityBlock = fmt.Sprintf("\n<identity>\n  You are: %s\n</identity>\n", identity)
	}

	return fmt.Sprintf(`You are an expert LaTeX/TikZ illustrator for academic papers.%s
<constraints>
- %s
- Use LaTeX/TikZ code.
- Output ONLY the tikzpicture code (from \begin{tikzpicture} to \end{tikzpicture}).
- Do NOT include \documentclass, \usepackage, \begin{document}, or any preamble.
- Do NOT wrap the output in markdown code fences.
- Do NOT use \definecolor — all colors are pre-defined in the preamble.
- %s
If the user provides <ref> examples, study them to understand the desired style and quality.
</constraints>

<task>
  Draw a diagram based on the user's description.
</task>
`, identityBlock, langConstraint, colorPrompt)
}
