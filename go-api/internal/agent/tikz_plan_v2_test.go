package agent

import (
	"strings"
	"testing"
)

func TestRenderV2_SingleBlock(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "enc", Label: "Encoder", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{
					{ID: "a", Label: "Conv1", Color: "primary"},
					{ID: "b", Label: "Conv2", Color: "primary"},
					{ID: "c", Label: "Pool", Color: "primary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "a", To: "b"},
			{From: "b", To: "c"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Should have named matrix
	if !strings.Contains(code, `\matrix (enc)`) {
		t.Error("expected named matrix (enc)")
	}
	// Should have fit=(enc) for background box
	if !strings.Contains(code, `fit=(enc)`) {
		t.Error("expected fit=(enc) for background box")
	}
	// First block should have no positioning attribute
	if strings.Contains(code, "below=") || strings.Contains(code, "right=") {
		t.Error("first block should have no positioning")
	}
	// Should have intra-block edges with explicit anchors (row layout → east/west)
	if !strings.Contains(code, "(enc-1-1.east) -- (enc-1-2.west)") {
		t.Error("expected anchored intra-block edge (enc-1-1.east) -- (enc-1-2.west)")
	}
}

func TestRenderV2_TwoBlocksVertical(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "a", Label: "Top", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{
					{ID: "n1", Label: "Node 1", Color: "primary"},
				},
			},
			{
				ID: "b", Label: "Bottom", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "a"},
				Nodes: []PlanNode{
					{ID: "n2", Label: "Node 2", Color: "secondary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "n1", To: "n2"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Block b should have "below=...of a"
	if !strings.Contains(code, "below=") || !strings.Contains(code, "of a") {
		t.Error("expected block b positioned below a, got:\n" + code)
	}
}

func TestRenderV2_SideBySide(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "enc", Label: "Encoder", Color: "primary",
				Layout: "column",
				Nodes: []PlanNode{
					{ID: "e1", Label: "Self-Attn", Color: "primary"},
					{ID: "e2", Label: "FFN", Color: "primary"},
				},
			},
			{
				ID: "dec", Label: "Decoder", Color: "secondary",
				Layout: "column",
				Position: &BlockPosition{Right: "enc"},
				Nodes: []PlanNode{
					{ID: "d1", Label: "Cross-Attn", Color: "secondary"},
					{ID: "d2", Label: "FFN", Color: "secondary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "e1", To: "e2"},
			{From: "e2", To: "d1"},
			{From: "d1", To: "d2"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Block dec should have "right=...of enc"
	if !strings.Contains(code, "right=") || !strings.Contains(code, "of enc") {
		t.Error("expected dec positioned right of enc, got:\n" + code)
	}
	// Cross-block edge should use .east/.west anchors
	if !strings.Contains(code, ".east") || !strings.Contains(code, ".west") {
		t.Error("expected .east/.west anchors for horizontal cross-block edge")
	}
}

func TestRenderV2_ColumnLayout(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "pipe", Label: "Pipeline", Color: "primary",
				Layout: "column",
				Nodes: []PlanNode{
					{ID: "s1", Label: "Step 1", Color: "primary"},
					{ID: "s2", Label: "Step 2", Color: "primary"},
					{ID: "s3", Label: "Step 3", Color: "primary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "s1", To: "s2"},
			{From: "s2", To: "s3"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Column layout: each node on its own row, col 1 → south/north anchors
	if !strings.Contains(code, "(pipe-1-1.south) -- (pipe-2-1.north)") {
		t.Error("expected anchored column edge (pipe-1-1.south) -- (pipe-2-1.north)")
	}
	if !strings.Contains(code, "(pipe-2-1.south) -- (pipe-3-1.north)") {
		t.Error("expected anchored column edge (pipe-2-1.south) -- (pipe-3-1.north)")
	}
}

func TestRenderV2_GridLayout(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "grid", Label: "Grid Block", Color: "tertiary",
				Layout: "grid",
				Nodes: []PlanNode{
					{ID: "g1", Label: "A", Color: "tertiary"},
					{ID: "g2", Label: "B", Color: "tertiary"},
					{ID: "g3", Label: "C", Color: "tertiary"},
					{ID: "g4", Label: "D", Color: "tertiary"},
					{ID: "g5", Label: "E", Color: "tertiary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "g1", To: "g2"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// 5 nodes, grid → ceil(sqrt(5))=3 cols → 2 rows: [A,B,C] [D,E,pad]
	// Should have invisible placeholder for incomplete last row
	if !strings.Contains(code, "draw=none, fill=none") {
		t.Error("expected invisible padding cell for incomplete grid row")
	}
	// g1 → (grid-1-1), g2 → (grid-1-2): same row → east/west anchors
	if !strings.Contains(code, "(grid-1-1.east) -- (grid-1-2.west)") {
		t.Error("expected anchored grid edge (grid-1-1.east) -- (grid-1-2.west)")
	}
}

func TestRenderV2_CrossBlockEdge(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "top", Label: "Top", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{
					{ID: "t1", Label: "T1", Color: "primary"},
				},
			},
			{
				ID: "bot", Label: "Bottom", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "top"},
				Nodes: []PlanNode{
					{ID: "b1", Label: "B1", Color: "secondary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "t1", To: "b1"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Cross-block vertical edge: .south with buffer offset -| .north
	if !strings.Contains(code, "(top-1-1.south) -- ++(0,-0.5cm) -| (bot-1-1.north)") {
		t.Error("expected vertical cross-block edge with buffer offset, got:\n" + code)
	}
}

func TestRenderV2_IntraBlockEdge(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "blk", Label: "Block", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{
					{ID: "x1", Label: "X", Color: "primary"},
					{ID: "x2", Label: "Y", Color: "primary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "x1", To: "x2"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Intra-block: row layout → east/west anchors
	if !strings.Contains(code, "(blk-1-1.east) -- (blk-1-2.west)") {
		t.Error("expected anchored intra-block edge (blk-1-1.east) -- (blk-1-2.west)")
	}
}

func TestRenderV2_NoDeadSpace(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "wide", Label: "Wide", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{
					{ID: "w1", Label: "A", Color: "primary"},
					{ID: "w2", Label: "B", Color: "primary"},
					{ID: "w3", Label: "C", Color: "primary"},
				},
			},
			{
				ID: "narrow", Label: "Narrow", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "wide"},
				Nodes: []PlanNode{
					{ID: "n1", Label: "X", Color: "secondary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "w1", To: "n1"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// The narrow block (1 node, row layout) should NOT have draw=none placeholders
	// Only grid layouts with incomplete last rows should have placeholders
	// Count occurrences of draw=none — should be zero since neither block is grid
	if strings.Contains(code, "draw=none") {
		t.Error("non-grid blocks should not have invisible padding cells")
	}

	// Each block gets its own fit — no forced uniform width
	if !strings.Contains(code, "fit=(wide)") {
		t.Error("expected fit=(wide) for wide block")
	}
	if !strings.Contains(code, "fit=(narrow)") {
		t.Error("expected fit=(narrow) for narrow block")
	}
}

func TestRenderV2_Annotations(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "blk", Label: "Block", Color: "primary",
				Layout: "column",
				Nodes: []PlanNode{
					{ID: "p1", Label: "Step 1", Color: "primary"},
					{ID: "p2", Label: "Step 2", Color: "primary"},
					{ID: "p3", Label: "Step 3", Color: "primary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "p1", To: "p2"},
			{From: "p2", To: "p3"},
		},
		Annotations: []PlanAnnotation{
			{Type: "brace", Cover: []string{"p1", "p3"}, Label: "Pipeline", Side: "right"},
		},
	}
	code := RenderTikZPlanV2(plan)

	// Annotation should reference V2 node names
	if !strings.Contains(code, "(blk-1-1.north)") {
		t.Error("expected annotation to reference blk-1-1")
	}
	if !strings.Contains(code, "(blk-3-1.south)") {
		t.Error("expected annotation to reference blk-3-1")
	}
	if !strings.Contains(code, "Pipeline") {
		t.Error("expected annotation label 'Pipeline'")
	}
}

func TestRenderV2_DiagonalEdge(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "a", Label: "A", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{{ID: "a1", Label: "A1", Color: "primary"}},
			},
			{
				ID: "b", Label: "B", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Right: "a"},
				Nodes: []PlanNode{{ID: "b1", Label: "B1", Color: "secondary"}},
			},
			{
				ID: "c", Label: "C", Color: "tertiary",
				Layout: "row",
				Position: &BlockPosition{Below: "b"},
				Nodes: []PlanNode{{ID: "c1", Label: "C1", Color: "tertiary"}},
			},
		},
		Edges: []PlanEdge{
			{From: "a1", To: "c1"}, // diagonal: a is not directly adjacent to c
		},
	}
	code := RenderTikZPlanV2(plan)

	// Diagonal edge should use the 3-segment routing with -|
	if !strings.Contains(code, "-|") {
		t.Error("expected Manhattan routing (-|) for diagonal cross-block edge, got:\n" + code)
	}
}

func TestRenderV2_BiArrowEdge(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "blk", Label: "Block", Color: "primary",
				Layout: "row",
				Nodes: []PlanNode{
					{ID: "x", Label: "X", Color: "primary"},
					{ID: "y", Label: "Y", Color: "primary"},
				},
			},
		},
		Edges: []PlanEdge{
			{From: "x", To: "y", Style: "biarrow"},
		},
	}
	code := RenderTikZPlanV2(plan)

	if !strings.Contains(code, "nice_biarrow") {
		t.Error("expected nice_biarrow style for biarrow edge")
	}
}

func TestRenderV2_SkipConnectionCurve(t *testing.T) {
	// 3-block vertical stack: A → B → C, with skip edge A→C
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "a", Label: "Block A", Color: "primary",
				Layout: "row",
				Nodes:  []PlanNode{{ID: "a1", Label: "A1", Color: "primary"}},
			},
			{
				ID: "b", Label: "Block B", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "a"},
				Nodes:    []PlanNode{{ID: "b1", Label: "B1", Color: "secondary"}},
			},
			{
				ID: "c", Label: "Block C", Color: "tertiary",
				Layout: "row",
				Position: &BlockPosition{Below: "b"},
				Nodes:    []PlanNode{{ID: "c1", Label: "C1", Color: "tertiary"}},
			},
		},
		Edges: []PlanEdge{
			{From: "a1", To: "b1"},
			{From: "b1", To: "c1"},
			{From: "a1", To: "c1"}, // skip connection: crosses block B
		},
	}
	code := RenderTikZPlanV2(plan)

	// Skip connection should produce smooth curve via west side
	if !strings.Contains(code, "to[out=180, in=180") {
		t.Error("expected skip connection curve with to[out=180, in=180], got:\n" + code)
	}
	if !strings.Contains(code, "(a-1-1.west)") {
		t.Error("expected skip curve to start from .west anchor, got:\n" + code)
	}
	if !strings.Contains(code, "(c-1-1.west)") {
		t.Error("expected skip curve to end at .west anchor, got:\n" + code)
	}
	// Adjacent edges should still use Manhattan routing
	if !strings.Contains(code, "(a-1-1.south) -- ++(0,-0.5cm) -| (b-1-1.north)") {
		t.Error("expected adjacent edge A→B to use Manhattan routing, got:\n" + code)
	}
}

func TestRenderV2_SkipConnectionLooseness(t *testing.T) {
	// 4-block vertical chain: A → B → C → D, with skip edge A→D (distance 3)
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "a", Label: "A", Color: "primary",
				Layout: "row",
				Nodes:  []PlanNode{{ID: "a1", Label: "A1", Color: "primary"}},
			},
			{
				ID: "b", Label: "B", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "a"},
				Nodes:    []PlanNode{{ID: "b1", Label: "B1", Color: "secondary"}},
			},
			{
				ID: "c", Label: "C", Color: "tertiary",
				Layout: "row",
				Position: &BlockPosition{Below: "b"},
				Nodes:    []PlanNode{{ID: "c1", Label: "C1", Color: "tertiary"}},
			},
			{
				ID: "d", Label: "D", Color: "quaternary",
				Layout: "row",
				Position: &BlockPosition{Below: "c"},
				Nodes:    []PlanNode{{ID: "d1", Label: "D1", Color: "quaternary"}},
			},
		},
		Edges: []PlanEdge{
			{From: "a1", To: "d1"}, // skip: distance 3 → looseness=1.5
		},
	}
	code := RenderTikZPlanV2(plan)

	if !strings.Contains(code, "looseness=1.5") {
		t.Error("expected looseness=1.5 for chain distance 3, got:\n" + code)
	}
}

func TestRenderV2_SkipConnectionWithLabel(t *testing.T) {
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "a", Label: "A", Color: "primary",
				Layout: "row",
				Nodes:  []PlanNode{{ID: "a1", Label: "A1", Color: "primary"}},
			},
			{
				ID: "b", Label: "B", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "a"},
				Nodes:    []PlanNode{{ID: "b1", Label: "B1", Color: "secondary"}},
			},
			{
				ID: "c", Label: "C", Color: "tertiary",
				Layout: "row",
				Position: &BlockPosition{Below: "b"},
				Nodes:    []PlanNode{{ID: "c1", Label: "C1", Color: "tertiary"}},
			},
		},
		Edges: []PlanEdge{
			{From: "a1", To: "c1", Label: "Residual"},
		},
	}
	code := RenderTikZPlanV2(plan)

	if !strings.Contains(code, "node[midway, left, fill=white") {
		t.Error("expected labeled skip curve with midway left node, got:\n" + code)
	}
	if !strings.Contains(code, "Residual") {
		t.Error("expected label 'Residual' in skip curve, got:\n" + code)
	}
}

func TestRenderV2_SkipEdgeTypeOnAdjacentBlocks(t *testing.T) {
	// Two adjacent blocks but edge marked as "skip" → should use curve
	plan := TikZPlanV2{
		Blocks: []PlanBlock{
			{
				ID: "a", Label: "A", Color: "primary",
				Layout: "row",
				Nodes:  []PlanNode{{ID: "a1", Label: "A1", Color: "primary"}},
			},
			{
				ID: "b", Label: "B", Color: "secondary",
				Layout: "row",
				Position: &BlockPosition{Below: "a"},
				Nodes:    []PlanNode{{ID: "b1", Label: "B1", Color: "secondary"}},
			},
		},
		Edges: []PlanEdge{
			{From: "a1", To: "b1", Type: "skip"},
		},
	}
	code := RenderTikZPlanV2(plan)

	if !strings.Contains(code, "to[out=180, in=180") {
		t.Error("expected skip curve for edge with type=skip even on adjacent blocks, got:\n" + code)
	}
}

func TestRenderV2_EmptyPlan(t *testing.T) {
	plan := TikZPlanV2{Blocks: nil}
	code := RenderTikZPlanV2(plan)
	if !strings.Contains(code, "Empty diagram") {
		t.Error("expected empty diagram fallback")
	}
}
