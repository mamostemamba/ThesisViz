package prompt

import (
	"strings"
	"testing"
)

func TestTikZ_ContainsColorRules(t *testing.T) {
	p := TikZ("en", "use warm colors", "")
	if !strings.Contains(p, "COLOR RULES (mandatory)") {
		t.Error("TikZ prompt should contain mandatory color rules")
	}
}
