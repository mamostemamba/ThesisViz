package agent

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseTikZ extracts \begin{tikzpicture}...\end{tikzpicture} from LLM output.
// Tries code fences first, then falls back to regex.
func ParseTikZ(raw string) (string, error) {
	// Try code fences: ```latex, ```tikz, ```tex
	if code := extractCodeBlock(raw, "latex", "tikz", "tex"); code != "" {
		if tikz := extractTikZEnv(code); tikz != "" {
			return tikz, nil
		}
		return code, nil
	}

	// Fallback: direct regex for tikzpicture environment
	if tikz := extractTikZEnv(raw); tikz != "" {
		return tikz, nil
	}

	return "", fmt.Errorf("no TikZ code found in LLM output")
}

// ParseMermaid extracts Mermaid code from LLM output.
func ParseMermaid(raw string) (string, error) {
	// Try code fence: ```mermaid
	if code := extractCodeBlock(raw, "mermaid"); code != "" {
		return strings.TrimSpace(code), nil
	}

	// Fallback: look for known Mermaid keywords at line start
	keywords := []string{"graph ", "flowchart ", "sequenceDiagram", "classDiagram",
		"stateDiagram", "erDiagram", "gantt", "pie ", "mindmap", "timeline"}
	for _, kw := range keywords {
		idx := strings.Index(raw, kw)
		if idx >= 0 {
			return strings.TrimSpace(raw[idx:]), nil
		}
	}

	return "", fmt.Errorf("no Mermaid code found in LLM output")
}

// ParseMatplotlib extracts Python code from LLM output.
func ParseMatplotlib(raw string) (string, error) {
	// Try code fences: ```python, ```py
	if code := extractCodeBlock(raw, "python", "py"); code != "" {
		return strings.TrimSpace(code), nil
	}

	// Fallback: look for import lines
	if strings.Contains(raw, "import matplotlib") || strings.Contains(raw, "import numpy") {
		return strings.TrimSpace(raw), nil
	}

	return "", fmt.Errorf("no Python/Matplotlib code found in LLM output")
}

// ParseJSON extracts a JSON object or array from LLM output.
func ParseJSON(raw string) (string, error) {
	// Try code fence: ```json
	if code := extractCodeBlock(raw, "json"); code != "" {
		return strings.TrimSpace(code), nil
	}

	// Fallback: find first { or [ and match to closing brace/bracket
	trimmed := strings.TrimSpace(raw)
	if json := extractBracketed(trimmed, '{', '}'); json != "" {
		return json, nil
	}
	if json := extractBracketed(trimmed, '[', ']'); json != "" {
		return json, nil
	}

	return "", fmt.Errorf("no JSON found in LLM output")
}

// extractCodeBlock finds content inside ```<lang> ... ``` fences.
func extractCodeBlock(raw string, langs ...string) string {
	for _, lang := range langs {
		// Pattern: ```lang\n...\n```
		prefix := "```" + lang
		idx := strings.Index(strings.ToLower(raw), strings.ToLower(prefix))
		if idx < 0 {
			continue
		}
		start := idx + len(prefix)
		// Skip to next newline after the fence
		nlIdx := strings.Index(raw[start:], "\n")
		if nlIdx < 0 {
			continue
		}
		start += nlIdx + 1

		// Find closing ```
		end := strings.Index(raw[start:], "```")
		if end < 0 {
			continue
		}
		return raw[start : start+end]
	}
	return ""
}

var tikzEnvRe = regexp.MustCompile(`(?s)(\\begin\{tikzpicture\}.*?\\end\{tikzpicture\})`)

func extractTikZEnv(s string) string {
	m := tikzEnvRe.FindString(s)
	return strings.TrimSpace(m)
}

// extractBracketed finds a balanced bracket/brace sequence.
func extractBracketed(s string, open, close byte) string {
	start := strings.IndexByte(s, open)
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
