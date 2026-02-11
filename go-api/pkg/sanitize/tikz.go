package sanitize

import "regexp"

var (
	defineColorRe = regexp.MustCompile(`\\definecolor\{[^}]*\}\{[^}]*\}\{[^}]*\}\s*,?\s*\n?`)
	tikzDefColorRe = regexp.MustCompile(`\s*define\s+color\s*=\s*\{[^}]*\}\{[^}]*\}\{[^}]*\}\s*,?`)
)

// TikZ removes duplicate \definecolor and tikzpicture-option "define color" lines
// that an AI model may have placed inside the tikzpicture environment.
func TikZ(code string) string {
	code = defineColorRe.ReplaceAllString(code, "")
	code = tikzDefColorRe.ReplaceAllString(code, "")
	return code
}
