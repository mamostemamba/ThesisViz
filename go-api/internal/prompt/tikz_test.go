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

func TestTikZPlanner_ContainsBlocksSchema(t *testing.T) {
	p := TikZPlanner("en", "")

	if !strings.Contains(p, `"blocks"`) {
		t.Error("planner prompt should contain blocks schema")
	}
	if !strings.Contains(p, `"layout"`) {
		t.Error("planner prompt should mention layout field")
	}
	if !strings.Contains(p, `"position"`) {
		t.Error("planner prompt should mention position field")
	}
	if !strings.Contains(p, "row | column | grid") {
		t.Error("planner prompt should list layout options")
	}
	// Should NOT contain old V1 "layers" schema description
	if strings.Contains(p, `"layers": [`) {
		t.Error("planner prompt should not contain old V1 layers schema")
	}
}

func TestTikZPlanner_LanguageRules(t *testing.T) {
	en := TikZPlanner("en", "")
	if !strings.Contains(en, "MUST be in English") {
		t.Error("English planner should require English labels")
	}

	zh := TikZPlanner("zh", "")
	if !strings.Contains(zh, "MUST be in Chinese") {
		t.Error("Chinese planner should require Chinese labels")
	}
}

func TestTikZ_ContainsSkipConnectionCurveRule(t *testing.T) {
	p := TikZ("en", "use warm colors", "")
	if !strings.Contains(p, "Rule B") {
		t.Error("TikZ prompt should contain Rule B for skip connection curves")
	}
	if !strings.Contains(p, "to[out=180, in=180") {
		t.Error("TikZ prompt should contain curve syntax example")
	}
	if !strings.Contains(p, "looseness=") {
		t.Error("TikZ prompt should mention looseness parameter")
	}
}

func TestTikZPlanner_ContainsEdgeTypeField(t *testing.T) {
	p := TikZPlanner("en", "")
	if !strings.Contains(p, `"type": "main_flow | skip"`) {
		t.Error("planner prompt should contain edge type field in schema")
	}
	if !strings.Contains(p, "skip connection") {
		t.Error("planner prompt should explain skip connections")
	}
}

func TestTikZ_ContainsNestedMatrixTemplate(t *testing.T) {
	p := TikZ("en", "use warm colors", "")
	if !strings.Contains(p, "TEMPLATE B") {
		t.Error("TikZ prompt should contain Template B for nested matrices")
	}
	if !strings.Contains(p, "NAMED MATRIX RULES") {
		t.Error("TikZ prompt should contain named matrix rules")
	}
	if !strings.Contains(p, "fit=(blockname)") {
		t.Error("TikZ prompt should mention fit=(blockname)")
	}
}

func TestTikZPlanner_ContainsLayoutMode(t *testing.T) {
	p := TikZPlanner("en", "")
	if !strings.Contains(p, `"layout_mode"`) {
		t.Error("planner prompt should contain layout_mode field")
	}
	if !strings.Contains(p, `"freeflow"`) {
		t.Error("planner prompt should mention freeflow option")
	}
	if !strings.Contains(p, `"matrix"`) {
		t.Error("planner prompt should mention matrix option")
	}
	if !strings.Contains(p, "LAYOUT MODE CLASSIFICATION") {
		t.Error("planner prompt should contain layout mode classification section")
	}
	// Example 3 should be a freeflow example
	if !strings.Contains(p, "EXAMPLE 3 (freeflow)") {
		t.Error("planner prompt should contain freeflow example")
	}
}

func TestTikZFreeFlow_ContainsPositioningRules(t *testing.T) {
	p := TikZFreeFlow("en", "use warm colors", "")

	// Must mention positioning library style
	if !strings.Contains(p, "below=") && !strings.Contains(p, "right=") {
		t.Error("freeflow prompt should contain positioning rules (below=, right=)")
	}

	// Must forbid \matrix
	if !strings.Contains(p, `\matrix`) {
		t.Error("freeflow prompt should mention \\matrix as forbidden")
	}

	// Must allow minimum height variation
	if !strings.Contains(p, "minimum height") {
		t.Error("freeflow prompt should mention minimum height for variable node sizes")
	}

	// Must contain swimlane pattern
	if !strings.Contains(p, "SWIMLANE PATTERN") {
		t.Error("freeflow prompt should contain swimlane pattern")
	}

	// Must contain sequence diagram pattern
	if !strings.Contains(p, "SEQUENCE DIAGRAM PATTERN") {
		t.Error("freeflow prompt should contain sequence diagram pattern")
	}

	// Must use pre-defined color names
	if !strings.Contains(p, "primaryFill") {
		t.Error("freeflow prompt should use pre-defined color names")
	}

	// Must have nice_arrow style
	if !strings.Contains(p, "nice_arrow") {
		t.Error("freeflow prompt should reference nice_arrow style")
	}
}

func TestTikZFreeFlow_LanguageSupport(t *testing.T) {
	en := TikZFreeFlow("en", "use warm colors", "")
	if !strings.Contains(en, "MUST be in English") {
		t.Error("English freeflow prompt should require English labels")
	}

	zh := TikZFreeFlow("zh", "use warm colors", "")
	if !strings.Contains(zh, "MUST be in Chinese") {
		t.Error("Chinese freeflow prompt should require Chinese labels")
	}
}

func TestTikZFreeFlow_Identity(t *testing.T) {
	withID := TikZFreeFlow("en", "", "computer networking")
	if !strings.Contains(withID, "computer networking") {
		t.Error("freeflow prompt should include identity when provided")
	}

	withoutID := TikZFreeFlow("en", "", "")
	if strings.Contains(withoutID, "expert in:") {
		t.Error("freeflow prompt should not include identity block when empty")
	}
}
