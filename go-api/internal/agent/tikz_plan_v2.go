package agent

import (
	"fmt"
	"math"
	"strings"
)

// TikZPlanV2 is the V2 JSON layout specification with nested blocks.
type TikZPlanV2 struct {
	Blocks      []PlanBlock      `json:"blocks"`
	Edges       []PlanEdge       `json:"edges"`
	Annotations []PlanAnnotation `json:"annotations,omitempty"`
}

// PlanBlock represents a logical group of nodes with its own internal layout.
type PlanBlock struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Color    string         `json:"color"`
	Position *BlockPosition `json:"position,omitempty"` // nil = first block (origin)
	Nodes    []PlanNode     `json:"nodes"`
	Layout   string         `json:"layout"` // "row" | "column" | "grid"
}

// BlockPosition describes how a block is positioned relative to another block.
type BlockPosition struct {
	Below string `json:"below,omitempty"`
	Right string `json:"right,omitempty"`
	Above string `json:"above,omitempty"`
	Left  string `json:"left,omitempty"`
}

// nodeRef records a node's block membership and TikZ reference name.
type nodeRef struct {
	BlockID  string
	TikZName string // e.g. "encoder-1-2"
}

// buildNodeRefMapV2 creates a mapping from node ID to its TikZ reference.
func buildNodeRefMapV2(plan TikZPlanV2) map[string]nodeRef {
	m := make(map[string]nodeRef)
	for _, block := range plan.Blocks {
		cols := gridCols(block.Layout, len(block.Nodes))
		for j, node := range block.Nodes {
			var row, col int
			switch block.Layout {
			case "row":
				row, col = 1, j+1
			case "column":
				row, col = j+1, 1
			default: // grid
				row = j/cols + 1
				col = j%cols + 1
			}
			ref := fmt.Sprintf("%s-%d-%d", block.ID, row, col)
			m[node.ID] = nodeRef{BlockID: block.ID, TikZName: ref}
		}
	}
	return m
}

// gridCols returns the number of columns for a given layout.
func gridCols(layout string, n int) int {
	switch layout {
	case "row":
		return n
	case "column":
		return 1
	default: // grid
		if n <= 4 {
			return min(4, n)
		}
		return min(4, int(math.Ceil(math.Sqrt(float64(n)))))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// findBlock returns a pointer to the block with the given ID, or nil.
func findBlock(plan TikZPlanV2, id string) *PlanBlock {
	for i := range plan.Blocks {
		if plan.Blocks[i].ID == id {
			return &plan.Blocks[i]
		}
	}
	return nil
}

// getRelativeDirection determines the spatial relationship from fromBlockID to toBlockID.
func getRelativeDirection(fromBlockID, toBlockID string, plan TikZPlanV2) string {
	toBlock := findBlock(plan, toBlockID)
	if toBlock != nil && toBlock.Position != nil {
		if toBlock.Position.Below == fromBlockID {
			return "below"
		}
		if toBlock.Position.Right == fromBlockID {
			return "right"
		}
		if toBlock.Position.Above == fromBlockID {
			return "above"
		}
		if toBlock.Position.Left == fromBlockID {
			return "left"
		}
	}
	fromBlock := findBlock(plan, fromBlockID)
	if fromBlock != nil && fromBlock.Position != nil {
		if fromBlock.Position.Below == toBlockID {
			return "above"
		}
		if fromBlock.Position.Right == toBlockID {
			return "left"
		}
		if fromBlock.Position.Above == toBlockID {
			return "below"
		}
		if fromBlock.Position.Left == toBlockID {
			return "right"
		}
	}
	return "diagonal"
}

// selectAnchors picks TikZ anchor suffixes based on block spatial relationship.
func selectAnchors(fromBlockID, toBlockID string, plan TikZPlanV2) (string, string) {
	rel := getRelativeDirection(fromBlockID, toBlockID, plan)
	switch rel {
	case "below":
		return ".south", ".north"
	case "above":
		return ".north", ".south"
	case "right":
		return ".east", ".west"
	case "left":
		return ".west", ".east"
	default:
		return ".south", ".north"
	}
}

// blockInternalSpacing returns row/column separation for a block's internal layout.
func blockInternalSpacing(block PlanBlock) (rowSep, colSep string) {
	n := len(block.Nodes)
	switch block.Layout {
	case "row":
		if n > 4 {
			colSep = "1.2cm"
		} else {
			colSep = "1.8cm"
		}
		rowSep = "1.5cm"
	case "column":
		if n > 4 {
			rowSep = "1.0cm"
		} else {
			rowSep = "1.5cm"
		}
		colSep = "2.0cm"
	default: // grid
		colSep = "1.8cm"
		rowSep = "1.2cm"
	}
	return
}

// interBlockGap returns the gap between blocks based on total block count.
func interBlockGap(totalBlocks int) string {
	switch {
	case totalBlocks > 4:
		return "2.0cm"
	case totalBlocks > 2:
		return "2.5cm"
	default:
		return "3.0cm"
	}
}

// RenderTikZPlanV2 converts a TikZPlanV2 into TikZ code using nested block matrices.
func RenderTikZPlanV2(plan TikZPlanV2) string {
	if len(plan.Blocks) == 0 {
		return "\\begin{tikzpicture}\n\\node {Empty diagram};\n\\end{tikzpicture}"
	}

	nodeMap := buildNodeRefMapV2(plan)
	gap := interBlockGap(len(plan.Blocks))

	var b strings.Builder

	// --- tikzpicture begin ---
	b.WriteString("\\begin{tikzpicture}[\n  node distance=0.5cm,\n]\n\n")

	// --- Block matrices ---
	for _, block := range plan.Blocks {
		emitBlockMatrix(&b, block, gap)
		b.WriteString("\n")
	}

	// --- Background layer boxes ---
	b.WriteString("\\begin{pgfonlayer}{background}\n")
	for _, block := range plan.Blocks {
		if len(block.Nodes) == 0 {
			continue
		}
		color := sanitizeColor(block.Color)
		b.WriteString(fmt.Sprintf(
			"  \\node[layer_box=%sLine, fit=(%s), label=above left:{\\sffamily\\normalsize\\bfseries %s}] {};\n",
			color, block.ID, escapeLaTeX(block.Label),
		))
	}
	b.WriteString("\\end{pgfonlayer}\n\n")

	// --- Edges ---
	for _, edge := range plan.Edges {
		fromRef, okFrom := nodeMap[edge.From]
		toRef, okTo := nodeMap[edge.To]
		if !okFrom || !okTo {
			continue
		}
		if fromRef.BlockID == toRef.BlockID {
			// Intra-block: direct connection
			b.WriteString(renderIntraBlockEdge(fromRef, toRef, edge.Label, edge.Style))
		} else {
			// Cross-block: Manhattan routing or skip-connection curve
			b.WriteString(renderCrossBlockEdge(fromRef, toRef, edge, plan))
		}
		b.WriteString("\n")
	}

	// --- Annotations ---
	for _, ann := range plan.Annotations {
		rendered := renderV2Annotation(ann, nodeMap)
		if rendered != "" {
			b.WriteString("\n")
			b.WriteString(rendered)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\\end{tikzpicture}")
	return b.String()
}

// emitBlockMatrix writes one block's \matrix declaration.
func emitBlockMatrix(b *strings.Builder, block PlanBlock, gap string) {
	posAttr := ""
	if block.Position != nil {
		if block.Position.Below != "" {
			posAttr = fmt.Sprintf("below=%s of %s, ", gap, block.Position.Below)
		} else if block.Position.Right != "" {
			posAttr = fmt.Sprintf("right=%s of %s, ", gap, block.Position.Right)
		} else if block.Position.Above != "" {
			posAttr = fmt.Sprintf("above=%s of %s, ", gap, block.Position.Above)
		} else if block.Position.Left != "" {
			posAttr = fmt.Sprintf("left=%s of %s, ", gap, block.Position.Left)
		}
	}

	rowSep, colSep := blockInternalSpacing(block)

	b.WriteString(fmt.Sprintf("\\matrix (%s) [\n  matrix of nodes,\n  %srow sep=%s,\n  column sep=%s,\n  nodes={matrix_node},\n] {\n",
		block.ID, posAttr, rowSep, colSep))

	switch block.Layout {
	case "row":
		b.WriteString("  %% " + block.Label + "\n")
		for j, node := range block.Nodes {
			if j > 0 {
				b.WriteString(" &\n")
			}
			writeNodeCell(b, node)
		}
		b.WriteString(" \\\\\n")

	case "column":
		b.WriteString("  %% " + block.Label + "\n")
		for _, node := range block.Nodes {
			writeNodeCell(b, node)
			b.WriteString(" \\\\\n")
		}

	default: // grid
		cols := gridCols(block.Layout, len(block.Nodes))
		b.WriteString("  %% " + block.Label + "\n")
		for j, node := range block.Nodes {
			if j > 0 && j%cols == 0 {
				b.WriteString(" \\\\\n")
			}
			if j%cols > 0 {
				b.WriteString(" &\n")
			}
			writeNodeCell(b, node)
		}
		// Pad last row if incomplete
		remainder := len(block.Nodes) % cols
		if remainder > 0 {
			for k := remainder; k < cols; k++ {
				b.WriteString(" &\n  |[draw=none, fill=none, drop shadow={opacity=0}]|")
			}
		}
		b.WriteString(" \\\\\n")
	}

	b.WriteString("};\n")
}

// writeNodeCell writes a single matrix cell for a node.
func writeNodeCell(b *strings.Builder, node PlanNode) {
	color := sanitizeColor(node.Color)
	b.WriteString(fmt.Sprintf("  |[fill=%sFill, draw=%sLine]| %s", color, color, escapeLaTeX(node.Label)))
}

// parseTikZRowCol extracts the row and column numbers from a TikZ matrix
// reference name like "encoder-1-2" → (1, 2).
func parseTikZRowCol(tikzName string) (row, col int) {
	parts := strings.Split(tikzName, "-")
	n := len(parts)
	if n >= 2 {
		fmt.Sscanf(parts[n-2], "%d", &row)
		fmt.Sscanf(parts[n-1], "%d", &col)
	}
	return
}

// renderIntraBlockEdge generates a \draw for an edge within the same block.
// Always uses explicit anchors (.north/.south/.east/.west), never center.
func renderIntraBlockEdge(from, to nodeRef, label, style string) string {
	arrowStyle := resolveArrowStyle(style)

	fromRow, fromCol := parseTikZRowCol(from.TikZName)
	toRow, toCol := parseTikZRowCol(to.TikZName)

	var path string
	switch {
	case fromRow == toRow:
		// Same row → horizontal with east/west anchors
		if fromCol < toCol {
			path = fmt.Sprintf("(%s.east) -- (%s.west)", from.TikZName, to.TikZName)
		} else {
			path = fmt.Sprintf("(%s.west) -- (%s.east)", from.TikZName, to.TikZName)
		}
	case fromCol == toCol:
		// Same column → vertical with south/north anchors
		if fromRow < toRow {
			path = fmt.Sprintf("(%s.south) -- (%s.north)", from.TikZName, to.TikZName)
		} else {
			path = fmt.Sprintf("(%s.north) -- (%s.south)", from.TikZName, to.TikZName)
		}
	default:
		// Different row and column within block → rectangular Manhattan with buffer offset
		if fromRow < toRow {
			path = fmt.Sprintf("(%s.south) -- ++(0,-0.5cm) -| (%s.north)", from.TikZName, to.TikZName)
		} else {
			path = fmt.Sprintf("(%s.north) -- ++(0,0.5cm) -| (%s.south)", from.TikZName, to.TikZName)
		}
	}

	if label != "" {
		return fmt.Sprintf("\\draw[%s] %s node[midway, fill=white, font=\\sffamily\\footnotesize] {%s};",
			arrowStyle, path, escapeLaTeX(label))
	}
	return fmt.Sprintf("\\draw[%s] %s;", arrowStyle, path)
}

// blockChainDistance walks the below/above position chain between two blocks
// to determine if they are in the same vertical column and how far apart they are.
// Returns the chain distance (2 = one intermediate block) or -1 if not in the same vertical chain.
func blockChainDistance(fromID, toID string, plan TikZPlanV2) int {
	// Try walking downward from fromID to toID via "below" chains.
	// Build a map: blockID → blockID that is directly below it.
	belowOf := make(map[string]string) // belowOf[X] = Y means Y is positioned below X
	for _, blk := range plan.Blocks {
		if blk.Position != nil && blk.Position.Below != "" {
			belowOf[blk.Position.Below] = blk.ID
		}
	}
	// Walk down from fromID
	cur := fromID
	dist := 0
	for cur != "" {
		if cur == toID {
			return dist
		}
		cur = belowOf[cur]
		dist++
		if dist > len(plan.Blocks) {
			break // safety: avoid infinite loop
		}
	}

	// Try walking upward (toID is above fromID)
	aboveOf := make(map[string]string) // aboveOf[X] = Y means Y is positioned above X
	for _, blk := range plan.Blocks {
		if blk.Position != nil && blk.Position.Above != "" {
			aboveOf[blk.Position.Above] = blk.ID
		}
	}
	cur = fromID
	dist = 0
	for cur != "" {
		if cur == toID {
			return dist
		}
		cur = aboveOf[cur]
		dist++
		if dist > len(plan.Blocks) {
			break
		}
	}

	return -1
}

// renderCrossBlockEdge generates a \draw for an edge spanning two different blocks.
func renderCrossBlockEdge(from, to nodeRef, edge PlanEdge, plan TikZPlanV2) string {
	arrowStyle := resolveArrowStyle(edge.Style)
	fromAnchor, toAnchor := selectAnchors(from.BlockID, to.BlockID, plan)

	// Check for skip connection: auto-detect via chain distance, or explicit edge.Type
	chainDist := blockChainDistance(from.BlockID, to.BlockID, plan)
	isSkip := chainDist >= 2 || edge.Type == "skip"

	rel := getRelativeDirection(from.BlockID, to.BlockID, plan)
	var path string
	switch rel {
	case "below":
		if isSkip {
			return renderSkipCurve(arrowStyle, from.TikZName, to.TikZName, edge.Label, chainDist)
		}
		// Vertical down: buffer offset clears source text, -| handles misalignment
		path = fmt.Sprintf("(%s%s) -- ++(0,-0.5cm) -| (%s%s)", from.TikZName, fromAnchor, to.TikZName, toAnchor)
	case "above":
		if isSkip {
			return renderSkipCurve(arrowStyle, from.TikZName, to.TikZName, edge.Label, chainDist)
		}
		// Vertical up: buffer offset clears source text, -| handles misalignment
		path = fmt.Sprintf("(%s%s) -- ++(0,0.5cm) -| (%s%s)", from.TikZName, fromAnchor, to.TikZName, toAnchor)
	case "right":
		// Horizontal right: buffer offset clears source text, |- handles misalignment
		path = fmt.Sprintf("(%s%s) -- ++(0.5cm,0) |- (%s%s)", from.TikZName, fromAnchor, to.TikZName, toAnchor)
	case "left":
		// Horizontal left: buffer offset clears source text, |- handles misalignment
		path = fmt.Sprintf("(%s%s) -- ++(-0.5cm,0) |- (%s%s)", from.TikZName, fromAnchor, to.TikZName, toAnchor)
	default: // diagonal — check for skip connection via vertical chain
		if chainDist >= 2 {
			return renderSkipCurve(arrowStyle, from.TikZName, to.TikZName, edge.Label, chainDist)
		}
		// True diagonal (not in same vertical chain) → Manhattan routing
		path = fmt.Sprintf("(%s%s) -- ++(0,-0.8cm) -| (%s%s)",
			from.TikZName, fromAnchor, to.TikZName, toAnchor)
	}

	if edge.Label != "" {
		return fmt.Sprintf("\\draw[%s] %s node[midway, fill=white, font=\\sffamily\\footnotesize] {%s};",
			arrowStyle, path, escapeLaTeX(edge.Label))
	}
	return fmt.Sprintf("\\draw[%s] %s;", arrowStyle, path)
}

// renderSkipCurve generates a smooth curve via the west side for skip connections.
func renderSkipCurve(arrowStyle, fromTikZ, toTikZ, label string, chainDist int) string {
	looseness := "1.2"
	if chainDist >= 3 {
		looseness = "1.5"
	}
	if label != "" {
		return fmt.Sprintf("\\draw[%s] (%s.west) to[out=180, in=180, looseness=%s] node[midway, left, fill=white, font=\\sffamily\\footnotesize] {%s} (%s.west);",
			arrowStyle, fromTikZ, looseness, escapeLaTeX(label), toTikZ)
	}
	return fmt.Sprintf("\\draw[%s] (%s.west) to[out=180, in=180, looseness=%s] (%s.west);",
		arrowStyle, fromTikZ, looseness, toTikZ)
}

// resolveArrowStyle converts edge style to TikZ arrow style name.
func resolveArrowStyle(style string) string {
	if style == "biarrow" {
		return "nice_biarrow"
	}
	return "nice_arrow"
}

// renderV2Annotation generates a brace annotation using V2 node references.
func renderV2Annotation(ann PlanAnnotation, nodeMap map[string]nodeRef) string {
	if len(ann.Cover) < 2 {
		return ""
	}

	first, okFirst := nodeMap[ann.Cover[0]]
	last, okLast := nodeMap[ann.Cover[len(ann.Cover)-1]]
	if !okFirst || !okLast {
		return ""
	}

	braceStyle := "visible_brace"
	anchor := "right=16pt"
	if ann.Side == "left" || ann.Type == "brace_mirror" {
		braceStyle = "visible_brace_mirror"
		anchor = "left=16pt"
	}

	return fmt.Sprintf("\\draw[%s] (%s.north) -- (%s.south) node[midway, %s, font=\\sffamily\\small\\bfseries] {%s};",
		braceStyle, first.TikZName, last.TikZName, anchor, escapeLaTeX(ann.Label))
}
