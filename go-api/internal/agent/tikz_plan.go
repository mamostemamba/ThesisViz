package agent

import (
	"fmt"
	"strings"
)

// TikZPlan is the JSON layout specification output by the planner LLM.
type TikZPlan struct {
	Layers      []PlanLayer      `json:"layers"`
	Edges       []PlanEdge       `json:"edges"`
	Annotations []PlanAnnotation `json:"annotations,omitempty"`
}

// PlanLayer represents one horizontal row in the diagram.
type PlanLayer struct {
	Name  string     `json:"name"`
	Nodes []PlanNode `json:"nodes"`
}

// PlanNode is a single element in the diagram.
type PlanNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"` // primary, secondary, tertiary, quaternary, highlight, neutral
}

// PlanEdge is a connection between two nodes.
type PlanEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
	Style string `json:"style,omitempty"` // "arrow" (default) or "biarrow"
}

// PlanAnnotation is a decorative annotation (brace, etc.).
type PlanAnnotation struct {
	Type  string   `json:"type"`            // "brace" or "brace_mirror"
	Cover []string `json:"cover,omitempty"` // node IDs the annotation spans
	Label string   `json:"label"`
	Side  string   `json:"side,omitempty"` // "left" or "right"
}

// nodePos records a node's position in the TikZ matrix (1-indexed).
type nodePos struct {
	Row int
	Col int
}

// validColors is the set of allowed color category names.
var validColors = map[string]bool{
	"primary": true, "secondary": true, "tertiary": true,
	"quaternary": true, "highlight": true, "neutral": true,
}

func sanitizeColor(c string) string {
	if validColors[c] {
		return c
	}
	return "primary"
}

// RenderTikZPlan converts a TikZPlan into TikZ code using strict \matrix layout.
func RenderTikZPlan(plan TikZPlan) string {
	if len(plan.Layers) == 0 {
		return "\\begin{tikzpicture}\n\\node {Empty diagram};\n\\end{tikzpicture}"
	}

	nodeMap := buildPlanNodeMap(plan)
	rowSep, colSep := adaptiveSpacing(plan)

	var b strings.Builder

	// --- tikzpicture begin ---
	b.WriteString("\\begin{tikzpicture}[\n  node distance=0.5cm,\n]\n\n")

	// --- Matrix ---
	b.WriteString(fmt.Sprintf("\\matrix (m) [\n  matrix of nodes,\n  row sep=%s,\n  column sep=%s,\n  nodes={matrix_node},\n] {\n", rowSep, colSep))

	for i, layer := range plan.Layers {
		writeMatrixRow(&b, layer)
		if i < len(plan.Layers)-1 {
			b.WriteString(" \\\\\n")
		} else {
			b.WriteString(" \\\\\n")
		}
	}
	b.WriteString("};\n\n")

	// --- Layer boxes (background) ---
	b.WriteString("\\begin{pgfonlayer}{background}\n")
	for i, layer := range plan.Layers {
		if len(layer.Nodes) == 0 {
			continue
		}
		row := i + 1
		firstCol := 1
		lastCol := len(layer.Nodes)
		color := sanitizeColor(layer.Nodes[0].Color)
		b.WriteString(fmt.Sprintf(
			"  \\node[layer_box=%sLine, fit=(m-%d-%d)(m-%d-%d), label=above left:{\\sffamily\\normalsize\\bfseries %s}] {};\n",
			color, row, firstCol, row, lastCol, escapeLaTeX(layer.Name),
		))
	}
	b.WriteString("\\end{pgfonlayer}\n\n")

	// --- Edges ---
	for _, edge := range plan.Edges {
		fromPos, okFrom := nodeMap[edge.From]
		toPos, okTo := nodeMap[edge.To]
		if !okFrom || !okTo {
			continue
		}
		b.WriteString(renderPlanEdge(fromPos, toPos, edge.Label, edge.Style))
		b.WriteString("\n")
	}

	// --- Annotations ---
	for _, ann := range plan.Annotations {
		rendered := renderPlanAnnotation(ann, nodeMap)
		if rendered != "" {
			b.WriteString("\n")
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\\end{tikzpicture}")
	return b.String()
}

// buildPlanNodeMap creates a mapping from node ID to (row, col) position.
func buildPlanNodeMap(plan TikZPlan) map[string]nodePos {
	m := make(map[string]nodePos)
	for i, layer := range plan.Layers {
		for j, node := range layer.Nodes {
			m[node.ID] = nodePos{Row: i + 1, Col: j + 1}
		}
	}
	return m
}

// adaptiveSpacing returns row/column separation based on diagram size.
func adaptiveSpacing(plan TikZPlan) (rowSep, colSep string) {
	maxCols := 0
	for _, layer := range plan.Layers {
		if len(layer.Nodes) > maxCols {
			maxCols = len(layer.Nodes)
		}
	}
	numRows := len(plan.Layers)

	switch {
	case maxCols > 5:
		colSep = "1.2cm"
	case maxCols > 3:
		colSep = "1.8cm"
	default:
		colSep = "2.2cm"
	}

	switch {
	case numRows > 5:
		rowSep = "1.0cm"
	case numRows > 3:
		rowSep = "1.5cm"
	default:
		rowSep = "1.8cm"
	}

	return rowSep, colSep
}

// writeMatrixRow writes one matrix row (one layer) to the builder.
func writeMatrixRow(b *strings.Builder, layer PlanLayer) {
	b.WriteString("  %% " + layer.Name + "\n")
	for j, node := range layer.Nodes {
		color := sanitizeColor(node.Color)
		if j > 0 {
			b.WriteString(" &\n")
		}
		b.WriteString(fmt.Sprintf("  |[fill=%sFill, draw=%sLine]| %s", color, color, escapeLaTeX(node.Label)))
	}
}

// renderPlanEdge generates a \draw command for one edge with manhattan routing.
func renderPlanEdge(from, to nodePos, label, style string) string {
	arrowStyle := "nice_arrow"
	if style == "biarrow" {
		arrowStyle = "nice_biarrow"
	}

	fromRef := fmt.Sprintf("m-%d-%d", from.Row, from.Col)
	toRef := fmt.Sprintf("m-%d-%d", to.Row, to.Col)

	var path string
	switch {
	case from.Row == to.Row:
		// Same row → straight horizontal
		path = fmt.Sprintf("(%s) -- (%s)", fromRef, toRef)
	case from.Col == to.Col:
		// Same column → straight vertical
		path = fmt.Sprintf("(%s) -- (%s)", fromRef, toRef)
	case from.Row < to.Row:
		// Going down, different column → manhattan routing
		path = fmt.Sprintf("(%s.south) |- (%s.north)", fromRef, toRef)
	default:
		// Going up, different column → manhattan routing
		path = fmt.Sprintf("(%s.north) |- (%s.south)", fromRef, toRef)
	}

	if label != "" {
		return fmt.Sprintf("\\draw[%s] %s node[midway, fill=white, font=\\sffamily\\footnotesize] {%s};",
			arrowStyle, path, escapeLaTeX(label))
	}
	return fmt.Sprintf("\\draw[%s] %s;", arrowStyle, path)
}

// renderPlanAnnotation generates a brace annotation.
func renderPlanAnnotation(ann PlanAnnotation, nodeMap map[string]nodePos) string {
	if len(ann.Cover) < 2 {
		return ""
	}

	first, okFirst := nodeMap[ann.Cover[0]]
	last, okLast := nodeMap[ann.Cover[len(ann.Cover)-1]]
	if !okFirst || !okLast {
		return ""
	}

	firstRef := fmt.Sprintf("m-%d-%d", first.Row, first.Col)
	lastRef := fmt.Sprintf("m-%d-%d", last.Row, last.Col)

	braceStyle := "visible_brace"
	anchor := "right=16pt"
	if ann.Side == "left" || ann.Type == "brace_mirror" {
		braceStyle = "visible_brace_mirror"
		anchor = "left=16pt"
	}

	return fmt.Sprintf("\\draw[%s] (%s.north) -- (%s.south) node[midway, %s] {%s};",
		braceStyle, firstRef, lastRef, anchor, escapeLaTeX(ann.Label))
}

// escapeLaTeX escapes special LaTeX characters in user-facing text.
func escapeLaTeX(s string) string {
	r := strings.NewReplacer(
		"&", "\\&",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
	)
	return r.Replace(s)
}
