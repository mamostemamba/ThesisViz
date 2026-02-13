package prompt

import (
	"strings"
	"testing"
)

func TestTikZLayoutPlan_Chinese(t *testing.T) {
	p := TikZLayoutPlan("zh")
	if !strings.Contains(p, "简体中文") {
		t.Error("Chinese layout plan should contain '简体中文'")
	}
}

func TestTikZLayoutPlan_English(t *testing.T) {
	p := TikZLayoutPlan("en")
	if !strings.Contains(p, "English") {
		t.Error("English layout plan should contain 'English'")
	}
}

func TestTikZFromLayout_ContainsGuidance(t *testing.T) {
	p := TikZFromLayout("en", "use warm colors", "")
	if !strings.Contains(p, "LAYOUT JSON GUIDANCE") {
		t.Error("TikZFromLayout should contain 'LAYOUT JSON GUIDANCE'")
	}
}

func TestTikZ_ContainsColorRules(t *testing.T) {
	p := TikZ("en", "use warm colors", "")
	if !strings.Contains(p, "COLOR RULES (mandatory)") {
		t.Error("TikZ prompt should contain mandatory color rules")
	}
}
