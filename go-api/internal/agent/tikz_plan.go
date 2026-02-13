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
	Type  string `json:"type,omitempty"`  // "main_flow" (default) or "skip"
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

// maxColumns returns the maximum number of nodes across all layers.
func maxColumns(plan TikZPlan) int {
	max := 0
	for _, layer := range plan.Layers {
		if len(layer.Nodes) > max {
			max = len(layer.Nodes)
		}
	}
	return max
}

// RenderTikZPlan converts a TikZPlan into TikZ code using strict \matrix layout.
func RenderTikZPlan(plan TikZPlan) string {
	if len(plan.Layers) == 0 {
		return "\\begin{tikzpicture}\n\\node {Empty diagram};\n\\end{tikzpicture}"
	}

	maxCols := maxColumns(plan)
	nodeMap := buildPlanNodeMap(plan)
	rowSep, colSep := adaptiveSpacing(plan)

	var b strings.Builder

	// --- tikzpicture begin ---
	b.WriteString("\\begin{tikzpicture}[\n  node distance=0.5cm,\n]\n\n")

	// --- Matrix ---
	b.WriteString(fmt.Sprintf("\\matrix (m) [\n  matrix of nodes,\n  row sep=%s,\n  column sep=%s,\n  nodes={matrix_node},\n] {\n", rowSep, colSep))

	for i, layer := range plan.Layers {
		writeMatrixRow(&b, layer, maxCols)
		b.WriteString(" \\\\\n")
		_ = i
	}
	b.WriteString("};\n\n")

	// --- Layer boxes (background) — uniform width across all layers ---
	b.WriteString("\\begin{pgfonlayer}{background}\n")
	for i, layer := range plan.Layers {
		if len(layer.Nodes) == 0 {
			continue
		}
		row := i + 1
		color := sanitizeColor(layer.Nodes[0].Color)
		b.WriteString(fmt.Sprintf(
			"  \\node[layer_box=%sLine, fit=(m-%d-%d)(m-%d-%d), label=above left:{\\sffamily\\normalsize\\bfseries %s}] {};\n",
			color, row, 1, row, maxCols, escapeLaTeX(layer.Name),
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
		b.WriteString(renderPlanEdge(fromPos, toPos, edge.Label, edge.Style, maxCols))
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
// Pads shorter rows with invisible empty cells so all rows have maxCols columns.
func writeMatrixRow(b *strings.Builder, layer PlanLayer, maxCols int) {
	b.WriteString("  %% " + layer.Name + "\n")
	for j := 0; j < maxCols; j++ {
		if j > 0 {
			b.WriteString(" &\n")
		}
		if j < len(layer.Nodes) {
			node := layer.Nodes[j]
			color := sanitizeColor(node.Color)
			b.WriteString(fmt.Sprintf("  |[fill=%sFill, draw=%sLine]| %s", color, color, escapeLaTeX(node.Label)))
		} else {
			// Invisible placeholder cell — maintains column alignment
			b.WriteString("  |[draw=none, fill=none, drop shadow={opacity=0}]|")
		}
	}
}

// renderPlanEdge generates a \draw command for one edge with manhattan routing.
// For skip connections (spanning 2+ rows across columns), routes around the outside
// of the diagram to avoid crossing intermediate nodes.
func renderPlanEdge(from, to nodePos, label, style string, maxCols int) string {
	arrowStyle := "nice_arrow"
	if style == "biarrow" {
		arrowStyle = "nice_biarrow"
	}

	fromRef := fmt.Sprintf("m-%d-%d", from.Row, from.Col)
	toRef := fmt.Sprintf("m-%d-%d", to.Row, to.Col)

	var path string
	rowDiff := abs(from.Row - to.Row)

	switch {
	case from.Row == to.Row:
		// Same row → straight horizontal with explicit anchors
		if from.Col < to.Col {
			path = fmt.Sprintf("(%s.east) -- (%s.west)", fromRef, toRef)
		} else {
			path = fmt.Sprintf("(%s.west) -- (%s.east)", fromRef, toRef)
		}

	case from.Col == to.Col:
		// Same column → straight vertical with explicit anchors
		if from.Row < to.Row {
			path = fmt.Sprintf("(%s.south) -- (%s.north)", fromRef, toRef)
		} else {
			path = fmt.Sprintf("(%s.north) -- (%s.south)", fromRef, toRef)
		}

	case rowDiff == 1:
		// Adjacent rows, different column → buffer offset + rectangular Manhattan path
		if from.Row < to.Row {
			path = fmt.Sprintf("(%s.south) -- ++(0,-0.6cm) -| (%s.north)", fromRef, toRef)
		} else {
			path = fmt.Sprintf("(%s.north) -- ++(0,0.6cm) -| (%s.south)", fromRef, toRef)
		}

	default:
		// Skip connection (2+ rows apart, different column)
		// Route around the outside to avoid crossing intermediate nodes.
		// Choose side: route via the closer edge (left or right).
		avgCol := float64(from.Col+to.Col) / 2.0
		midCol := float64(maxCols+1) / 2.0

		if avgCol <= midCol {
			// Route via left side with buffer offset
			path = fmt.Sprintf("(%s.west) -- ++(-1.2cm,0) |- (%s.west)", fromRef, toRef)
		} else {
			// Route via right side with buffer offset
			path = fmt.Sprintf("(%s.east) -- ++(1.2cm,0) |- (%s.east)", fromRef, toRef)
		}
	}

	if label != "" {
		return fmt.Sprintf("\\draw[%s] %s node[midway, fill=white, font=\\sffamily\\footnotesize] {%s};",
			arrowStyle, path, escapeLaTeX(label))
	}
	return fmt.Sprintf("\\draw[%s] %s;", arrowStyle, path)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

	return fmt.Sprintf("\\draw[%s] (%s.north) -- (%s.south) node[midway, %s, font=\\sffamily\\small\\bfseries] {%s};",
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
