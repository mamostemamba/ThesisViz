package agent

import (
	"strings"
	"testing"
)

func TestRenderTikZPlan_PadsRows(t *testing.T) {
	plan := TikZPlan{
		Layers: []PlanLayer{
			{Name: "Top", Nodes: []PlanNode{
				{ID: "a", Label: "A", Color: "primary"},
				{ID: "b", Label: "B", Color: "primary"},
				{ID: "c", Label: "C", Color: "primary"},
			}},
			{Name: "Mid", Nodes: []PlanNode{
				{ID: "x", Label: "X", Color: "secondary"},
				{ID: "y", Label: "Y", Color: "secondary"},
			}},
		},
		Edges: []PlanEdge{
			{From: "a", To: "x"},
			{From: "b", To: "y"},
		},
	}
	code := RenderTikZPlan(plan)

	// Row 2 should have an invisible padding cell (3rd column)
	if !strings.Contains(code, "draw=none, fill=none") {
		t.Error("expected invisible padding cell for shorter row")
	}

	// Layer box should span full width (m-2-1 to m-2-3)
	if !strings.Contains(code, "fit=(m-2-1)(m-2-3)") {
		t.Error("expected layer box to span maxCols=3, got:\n" + code)
	}
}

func TestRenderTikZPlan_SkipConnection(t *testing.T) {
	plan := TikZPlan{
		Layers: []PlanLayer{
			{Name: "L1", Nodes: []PlanNode{{ID: "a", Label: "A", Color: "primary"}}},
			{Name: "L2", Nodes: []PlanNode{{ID: "b", Label: "B", Color: "secondary"}}},
			{Name: "L3", Nodes: []PlanNode{{ID: "c", Label: "C", Color: "tertiary"}}},
		},
		Edges: []PlanEdge{
			{From: "a", To: "b"},
			{From: "a", To: "c"}, // skip connection: row 1 → row 3
		},
	}
	code := RenderTikZPlan(plan)

	// Skip connection (same column, 2 rows apart) should be a straight vertical line
	if !strings.Contains(code, "(m-1-1) -- (m-3-1)") {
		t.Error("expected straight vertical for same-column skip connection")
	}
}
