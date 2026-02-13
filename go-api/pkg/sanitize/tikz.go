package sanitize

import (
	"regexp"
	"strings"
)

var (
	defineColorRe  = regexp.MustCompile(`\\definecolor\{[^}]*\}\{[^}]*\}\{[^}]*\}\s*,?\s*\n?`)
	tikzDefColorRe = regexp.MustCompile(`\s*define\s+color\s*=\s*\{[^}]*\}\{[^}]*\}\{[^}]*\}\s*,?`)

	// Text-command sanitizers: replace risky LaTeX commands that cause
	// "raw code leakage" when used inside TikZ matrix cells.
	textbfRe    = regexp.MustCompile(`\\textbf\{([^}]*)\}`)
	textitRe    = regexp.MustCompile(`\\textit\{([^}]*)\}`)
	textrmRe    = regexp.MustCompile(`\\textrm\{([^}]*)\}`)
	textsfRe    = regexp.MustCompile(`\\textsf\{([^}]*)\}`)
	textttRe    = regexp.MustCompile(`\\texttt\{([^}]*)\}`)
	textscRe    = regexp.MustCompile(`\\textsc\{([^}]*)\}`)
	underlineRe = regexp.MustCompile(`\\underline\{([^}]*)\}`)
	emphRe      = regexp.MustCompile(`\\emph\{([^}]*)\}`)
	textRe      = regexp.MustCompile(`\\text\{([^}]*)\}`)
	mboxRe      = regexp.MustCompile(`\\mbox\{([^}]*)\}`)
	makeboxRe   = regexp.MustCompile(`\\makebox(?:\[[^\]]*\])?\{([^}]*)\}`)
	parboxRe    = regexp.MustCompile(`\\parbox(?:\[[^\]]*\])?\{[^}]*\}\{([^}]*)\}`)
	fontSizeRe  = regexp.MustCompile(`\\(?:footnotesize|scriptsize|tiny|small|large|Large|LARGE|huge|Huge|normalsize)\b\s*`)
	tabularRe   = regexp.MustCompile(`\\begin\{tabular\}(?:\{[^}]*\})?([\s\S]*?)\\end\{tabular\}`)

	// Matrix-of-nodes detection: ensure |[style]| syntax has proper matrix option.
	matrixWithOptsRe    = regexp.MustCompile(`(\\matrix\s*(?:\([^)]*\)\s*)?\[)([^\]]*)(\])`)
	matrixWithoutOptsRe = regexp.MustCompile(`(\\matrix\s*(?:\([^)]*\)\s*)?)(\{)`)
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

// ensureMatrixOfNodes adds "matrix of nodes" option to \matrix commands
// when the code contains |[style]| syntax (which requires it).
func ensureMatrixOfNodes(code string) string {
	if !strings.Contains(code, "|[") {
		return code
	}

	// Case 1: \matrix[options] but missing "matrix of nodes"
	code = matrixWithOptsRe.ReplaceAllStringFunc(code, func(match string) string {
		subs := matrixWithOptsRe.FindStringSubmatch(match)
		if len(subs) < 4 {
			return match
		}
		if strings.Contains(subs[2], "matrix of nodes") {
			return match
		}
		return subs[1] + "matrix of nodes, " + subs[2] + subs[3]
	})

	// Case 2: \matrix{ (no options at all) → insert [matrix of nodes]
	code = matrixWithoutOptsRe.ReplaceAllStringFunc(code, func(match string) string {
		subs := matrixWithoutOptsRe.FindStringSubmatch(match)
		if len(subs) < 3 {
			return match
		}
		return subs[1] + "[matrix of nodes] " + subs[2]
	})

	return code
}

// TikZClean performs minimal, safe cleanup: removes duplicate color definitions
// and ensures matrix-of-nodes is present when |[style]| syntax is used.
// Unlike TikZ(), it does NOT rewrite \textbf, \textit, etc.
func TikZClean(code string) string {
	code = defineColorRe.ReplaceAllString(code, "")
	code = tikzDefColorRe.ReplaceAllString(code, "")
	code = ensureMatrixOfNodes(code)
	return code
}

// TikZ removes duplicate \definecolor, tikzpicture-option "define color" lines,
// and sanitizes risky LaTeX text commands that cause raw code leakage in
// rendered TikZ matrix cells.
func TikZ(code string) string {
	code = defineColorRe.ReplaceAllString(code, "")
	code = tikzDefColorRe.ReplaceAllString(code, "")

	// Replace \textXX{...} → brace-grouped equivalents or plain text
	code = textbfRe.ReplaceAllString(code, `{\bfseries $1}`)
	code = textitRe.ReplaceAllString(code, `{\itshape $1}`)
	code = textrmRe.ReplaceAllString(code, `$1`)
	code = textsfRe.ReplaceAllString(code, `{\sffamily $1}`)
	code = textttRe.ReplaceAllString(code, `{\ttfamily $1}`)
	code = textscRe.ReplaceAllString(code, `$1`)
	code = emphRe.ReplaceAllString(code, `{\itshape $1}`)
	code = underlineRe.ReplaceAllString(code, `$1`)
	code = textRe.ReplaceAllString(code, `$1`)
	code = mboxRe.ReplaceAllString(code, `$1`)
	code = makeboxRe.ReplaceAllString(code, `$1`)
	code = parboxRe.ReplaceAllString(code, `$1`)
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

	code = ensureMatrixOfNodes(code)
	return code
}
