package sanitize

import "regexp"

var (
	defineColorRe  = regexp.MustCompile(`\\definecolor\{[^}]*\}\{[^}]*\}\{[^}]*\}\s*,?\s*\n?`)
	tikzDefColorRe = regexp.MustCompile(`\s*define\s+color\s*=\s*\{[^}]*\}\{[^}]*\}\{[^}]*\}\s*,?`)

	// Text-command sanitizers: replace risky LaTeX commands that cause
	// "raw code leakage" when used inside TikZ matrix cells.
	textbfRe    = regexp.MustCompile(`\\textbf\{([^}]*)\}`)
	textitRe    = regexp.MustCompile(`\\textit\{([^}]*)\}`)
	underlineRe = regexp.MustCompile(`\\underline\{([^}]*)\}`)
	emphRe      = regexp.MustCompile(`\\emph\{([^}]*)\}`)
	textRe      = regexp.MustCompile(`\\text\{([^}]*)\}`)
	mboxRe      = regexp.MustCompile(`\\mbox\{([^}]*)\}`)
	fontSizeRe  = regexp.MustCompile(`\\(?:footnotesize|scriptsize|tiny|small|large|Large|LARGE|huge|Huge|normalsize)\b\s*`)
	tabularRe   = regexp.MustCompile(`\\begin\{tabular\}(?:\{[^}]*\})?([\s\S]*?)\\end\{tabular\}`)
)

// stripTabularContent extracts plain text from a tabular body by removing
// column separators (&), row separators (\\), \hline, \cline, etc.
func stripTabularContent(body string) string {
	// Remove \hline, \cline{...}
	s := regexp.MustCompile(`\\(?:hline|cline\{[^}]*\})`).ReplaceAllString(body, "")
	// Replace & and \\ with spaces
	s = regexp.MustCompile(`[&]`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\\\\`).ReplaceAllString(s, " ")
	// Collapse whitespace
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s
}

// TikZ removes duplicate \definecolor, tikzpicture-option "define color" lines,
// and sanitizes risky LaTeX text commands that cause raw code leakage in
// rendered TikZ matrix cells.
func TikZ(code string) string {
	code = defineColorRe.ReplaceAllString(code, "")
	code = tikzDefColorRe.ReplaceAllString(code, "")

	// Replace \textbf{...} → {\bfseries ...}
	code = textbfRe.ReplaceAllString(code, `{\bfseries $1}`)
	// Replace \textit{...} → {\itshape ...}
	code = textitRe.ReplaceAllString(code, `{\itshape $1}`)
	// Replace \emph{...} → {\itshape ...}
	code = emphRe.ReplaceAllString(code, `{\itshape $1}`)
	// Replace \underline{...} → plain text (underline rarely needed in diagrams)
	code = underlineRe.ReplaceAllString(code, `$1`)
	// Replace \text{...} → plain content
	code = textRe.ReplaceAllString(code, `$1`)
	// Replace \mbox{...} → plain content
	code = mboxRe.ReplaceAllString(code, `$1`)
	// Remove font size commands
	code = fontSizeRe.ReplaceAllString(code, "")
	// Replace \begin{tabular}...\end{tabular} → extracted plain text
	code = tabularRe.ReplaceAllStringFunc(code, func(match string) string {
		subs := tabularRe.FindStringSubmatch(match)
		if len(subs) < 2 {
			return match
		}
		return stripTabularContent(subs[1])
	})

	return code
}
