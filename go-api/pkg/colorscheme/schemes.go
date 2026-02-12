package colorscheme

import (
	"fmt"
	"strings"
)

// Scheme holds all color-related configuration for a single visual theme.
type Scheme struct {
	Name                 string
	TikZColors           string // legacy scheme-specific color names
	TikZPrompt           string
	MatplotlibColors     string
	MatplotlibEdgeColors string
	MermaidTheme         string
}

// ColorPair holds a light fill color and a darker line/border color.
type ColorPair struct {
	Fill string `json:"fill"`
	Line string `json:"line"`
}

// CustomColors holds 4-8 fill/line pairs extracted from a reference image.
type CustomColors struct {
	Pairs []ColorPair `json:"pairs"`
}

// stripHash removes leading '#' from a hex color string.
func stripHash(hex string) string {
	return strings.TrimPrefix(hex, "#")
}

// Semantic alias names for the first 6 slots.
var semanticNames = [6]string{"primary", "secondary", "tertiary", "quaternary", "highlight", "neutral"}

// Default neutral fallback when fewer than 6 pairs are provided.
var defaultNeutral = ColorPair{Fill: "#F5F5F5", Line: "#666666"}

// FromCustom builds a complete Scheme from user-supplied custom colors (4-8 pairs).
//
// It generates:
//  1. Indexed names: color1Fill/color1Line .. colorNFill/colorNLine
//  2. Semantic aliases: primaryFill=color1Fill .. neutralFill=color6Fill (up to 6)
//     - If fewer than 6 pairs, missing semantic slots get sensible defaults
func FromCustom(c CustomColors) Scheme {
	n := len(c.Pairs)
	if n == 0 {
		return Get("drawio")
	}

	// --- TikZ color definitions ---
	var tikzLines []string

	// 1) Indexed names
	for i, p := range c.Pairs {
		f := stripHash(p.Fill)
		l := stripHash(p.Line)
		tikzLines = append(tikzLines,
			fmt.Sprintf(`\definecolor{color%dFill}{HTML}{%s}`, i+1, f),
			fmt.Sprintf(`\definecolor{color%dLine}{HTML}{%s}`, i+1, l),
		)
	}

	// 2) Semantic aliases (map first N pairs, pad missing with defaults)
	for i, name := range semanticNames {
		var f, l string
		switch {
		case i < n:
			f = stripHash(c.Pairs[i].Fill)
			l = stripHash(c.Pairs[i].Line)
		case i == 5: // neutral slot — use dedicated default
			f = stripHash(defaultNeutral.Fill)
			l = stripHash(defaultNeutral.Line)
		default:
			// Wrap around to reuse earlier colors
			f = stripHash(c.Pairs[i%n].Fill)
			l = stripHash(c.Pairs[i%n].Line)
		}
		tikzLines = append(tikzLines,
			fmt.Sprintf(`\definecolor{%sFill}{HTML}{%s}`, name, f),
			fmt.Sprintf(`\definecolor{%sLine}{HTML}{%s}`, name, l),
		)
	}
	tikzColors := strings.Join(tikzLines, "\n")

	// --- TikZ prompt ---
	var promptLines []string
	promptLines = append(promptLines,
		"These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:")
	// Always list semantic names
	for i, name := range semanticNames {
		if i < n {
			promptLines = append(promptLines,
				fmt.Sprintf("  - %s: fill=%sFill, draw=%sLine  (also: fill=color%dFill, draw=color%dLine)",
					strings.Title(name), name, name, i+1, i+1))
		} else {
			promptLines = append(promptLines,
				fmt.Sprintf("  - %s: fill=%sFill, draw=%sLine  (fallback/default)",
					strings.Title(name), name, name))
		}
	}
	// Extra indexed colors beyond 6
	for i := 6; i < n; i++ {
		promptLines = append(promptLines,
			fmt.Sprintf("  - Extra color %d: fill=color%dFill, draw=color%dLine", i+1, i+1, i+1))
	}
	tikzPrompt := strings.Join(promptLines, "\n")

	// --- Matplotlib colors ---
	var mplFills, mplLines []string
	for _, p := range c.Pairs {
		mplFills = append(mplFills, fmt.Sprintf("'#%s'", stripHash(p.Fill)))
		mplLines = append(mplLines, fmt.Sprintf("'#%s'", stripHash(p.Line)))
	}
	matplotlibColors := "[" + strings.Join(append(mplFills, mplLines...), ", ") + "]"
	matplotlibEdge := "[" + strings.Join(mplLines, ", ") + "]"

	return Scheme{
		Name:                 "Custom",
		TikZColors:           tikzColors,
		TikZPrompt:           tikzPrompt,
		MatplotlibColors:     matplotlibColors,
		MatplotlibEdgeColors: matplotlibEdge,
		MermaidTheme:         "default",
	}
}

// Unified semantic color names — every scheme defines the same 6 fill/line pairs.
// AI-generated code should use these names so switching schemes just changes colors.
var unifiedColors = map[string]string{
	"drawio": `\definecolor{primaryFill}{HTML}{DAE8FC}
\definecolor{primaryLine}{HTML}{6C8EBF}
\definecolor{secondaryFill}{HTML}{D5E8D4}
\definecolor{secondaryLine}{HTML}{82B366}
\definecolor{tertiaryFill}{HTML}{FFE6CC}
\definecolor{tertiaryLine}{HTML}{D79B00}
\definecolor{quaternaryFill}{HTML}{E1D5E7}
\definecolor{quaternaryLine}{HTML}{9673A6}
\definecolor{highlightFill}{HTML}{F8CECC}
\definecolor{highlightLine}{HTML}{B85450}
\definecolor{neutralFill}{HTML}{F5F5F5}
\definecolor{neutralLine}{HTML}{666666}`,

	"professional_blue": `\definecolor{primaryFill}{HTML}{DAE6F0}
\definecolor{primaryLine}{HTML}{4682B4}
\definecolor{secondaryFill}{HTML}{D8F0E4}
\definecolor{secondaryLine}{HTML}{3CB371}
\definecolor{tertiaryFill}{HTML}{FFEDCC}
\definecolor{tertiaryLine}{HTML}{FFA500}
\definecolor{quaternaryFill}{HTML}{F8D0D8}
\definecolor{quaternaryLine}{HTML}{DC143C}
\definecolor{highlightFill}{HTML}{E0D4F0}
\definecolor{highlightLine}{HTML}{9370DB}
\definecolor{neutralFill}{HTML}{F5F5FA}
\definecolor{neutralLine}{HTML}{888888}`,

	"bold_contrast": `\definecolor{primaryFill}{HTML}{FADBD7}
\definecolor{primaryLine}{HTML}{E64B35}
\definecolor{secondaryFill}{HTML}{DBF1F7}
\definecolor{secondaryLine}{HTML}{4DBBD5}
\definecolor{tertiaryFill}{HTML}{CCECE7}
\definecolor{tertiaryLine}{HTML}{00A087}
\definecolor{quaternaryFill}{HTML}{D8DDE7}
\definecolor{quaternaryLine}{HTML}{3C5488}
\definecolor{highlightFill}{HTML}{FDEBE5}
\definecolor{highlightLine}{HTML}{F39B7F}
\definecolor{neutralFill}{HTML}{F5F5F5}
\definecolor{neutralLine}{HTML}{888888}`,

	"minimal_mono": `\definecolor{primaryFill}{HTML}{CCE3F0}
\definecolor{primaryLine}{HTML}{0072B2}
\definecolor{secondaryFill}{HTML}{F7DFCC}
\definecolor{secondaryLine}{HTML}{D55E00}
\definecolor{tertiaryFill}{HTML}{CCECE3}
\definecolor{tertiaryLine}{HTML}{009E73}
\definecolor{quaternaryFill}{HTML}{F5E4ED}
\definecolor{quaternaryLine}{HTML}{CC79A7}
\definecolor{highlightFill}{HTML}{FCFAD9}
\definecolor{highlightLine}{HTML}{F0E442}
\definecolor{neutralFill}{HTML}{F5F5F5}
\definecolor{neutralLine}{HTML}{888888}`,

	"modern_teal": `\definecolor{primaryFill}{HTML}{D0F0F0}
\definecolor{primaryLine}{HTML}{009688}
\definecolor{secondaryFill}{HTML}{E3F2FD}
\definecolor{secondaryLine}{HTML}{1976D2}
\definecolor{tertiaryFill}{HTML}{FFF3E0}
\definecolor{tertiaryLine}{HTML}{FF9800}
\definecolor{quaternaryFill}{HTML}{F3E5F5}
\definecolor{quaternaryLine}{HTML}{7B1FA2}
\definecolor{highlightFill}{HTML}{E8F5E9}
\definecolor{highlightLine}{HTML}{388E3C}
\definecolor{neutralFill}{HTML}{FAFAFA}
\definecolor{neutralLine}{HTML}{757575}`,

	"soft_pastel": `\definecolor{primaryFill}{HTML}{E8D5D5}
\definecolor{primaryLine}{HTML}{B07D7D}
\definecolor{secondaryFill}{HTML}{D5DDE8}
\definecolor{secondaryLine}{HTML}{7D8DB0}
\definecolor{tertiaryFill}{HTML}{DDE8D5}
\definecolor{tertiaryLine}{HTML}{8DB07D}
\definecolor{quaternaryFill}{HTML}{E8E0D5}
\definecolor{quaternaryLine}{HTML}{B0A07D}
\definecolor{highlightFill}{HTML}{E0D5E8}
\definecolor{highlightLine}{HTML}{A07DB0}
\definecolor{neutralFill}{HTML}{F0EFED}
\definecolor{neutralLine}{HTML}{999999}`,

	"warm_earth": `\definecolor{primaryFill}{HTML}{FFF8E1}
\definecolor{primaryLine}{HTML}{F9A825}
\definecolor{secondaryFill}{HTML}{E8F5E9}
\definecolor{secondaryLine}{HTML}{2E7D32}
\definecolor{tertiaryFill}{HTML}{FBE9E7}
\definecolor{tertiaryLine}{HTML}{BF360C}
\definecolor{quaternaryFill}{HTML}{EFEBE9}
\definecolor{quaternaryLine}{HTML}{6D4C41}
\definecolor{highlightFill}{HTML}{FFF3E0}
\definecolor{highlightLine}{HTML}{E65100}
\definecolor{neutralFill}{HTML}{FAF8F5}
\definecolor{neutralLine}{HTML}{8D6E63}`,

	"cyber_dark": `\definecolor{primaryFill}{HTML}{1A237E}
\definecolor{primaryLine}{HTML}{448AFF}
\definecolor{secondaryFill}{HTML}{004D40}
\definecolor{secondaryLine}{HTML}{00E676}
\definecolor{tertiaryFill}{HTML}{4A148C}
\definecolor{tertiaryLine}{HTML}{E040FB}
\definecolor{quaternaryFill}{HTML}{1B2631}
\definecolor{quaternaryLine}{HTML}{00BCD4}
\definecolor{highlightFill}{HTML}{3E2723}
\definecolor{highlightLine}{HTML}{FF6D00}
\definecolor{neutralFill}{HTML}{263238}
\definecolor{neutralLine}{HTML}{90A4AE}`,
}

var schemes = map[string]Scheme{
	"drawio": {
		Name: "Draw.io 经典",
		TikZColors: `\definecolor{drawBlueFill}{HTML}{DAE8FC}
\definecolor{drawBlueLine}{HTML}{6C8EBF}
\definecolor{drawGreenFill}{HTML}{D5E8D4}
\definecolor{drawGreenLine}{HTML}{82B366}
\definecolor{drawOrangeFill}{HTML}{FFE6CC}
\definecolor{drawOrangeLine}{HTML}{D79B00}
\definecolor{drawPurpleFill}{HTML}{E1D5E7}
\definecolor{drawPurpleLine}{HTML}{9673A6}
\definecolor{drawRedFill}{HTML}{F8CECC}
\definecolor{drawRedLine}{HTML}{B85450}
\definecolor{drawGreyFill}{HTML}{F5F5F5}
\definecolor{drawGreyLine}{HTML}{666666}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (blue): fill=primaryFill, draw=primaryLine  (also: fill=drawBlueFill, draw=drawBlueLine)
  - Secondary (green): fill=secondaryFill, draw=secondaryLine  (also: fill=drawGreenFill, draw=drawGreenLine)
  - Tertiary (orange): fill=tertiaryFill, draw=tertiaryLine  (also: fill=drawOrangeFill, draw=drawOrangeLine)
  - Quaternary (purple): fill=quaternaryFill, draw=quaternaryLine  (also: fill=drawPurpleFill, draw=drawPurpleLine)
  - Highlight (red): fill=highlightFill, draw=highlightLine  (also: fill=drawRedFill, draw=drawRedLine)
  - Neutral (grey): fill=neutralFill, draw=neutralLine  (also: fill=drawGreyFill, draw=drawGreyLine)
You can use either the semantic names (primaryFill) or the drawio names (drawBlueFill) — both are defined.`,
		MatplotlibColors:     "['#DAE8FC', '#D5E8D4', '#FFE6CC', '#E1D5E7', '#F8CECC', '#6C8EBF', '#82B366', '#D79B00', '#9673A6', '#B85450']",
		MatplotlibEdgeColors: "['#6C8EBF', '#82B366', '#D79B00', '#9673A6', '#B85450', '#666666']",
		MermaidTheme:         "default",
	},
	"professional_blue": {
		Name: "专业蓝",
		TikZColors: `\definecolor{accent1}{RGB}{70,130,180}
\definecolor{accent2}{RGB}{60,179,113}
\definecolor{accent3}{RGB}{255,165,0}
\definecolor{accent4}{RGB}{220,20,60}
\definecolor{bgcolor}{RGB}{245,245,250}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (steel blue): fill=primaryFill, draw=primaryLine
  - Secondary (sea green): fill=secondaryFill, draw=secondaryLine
  - Tertiary (orange): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (crimson): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (purple): fill=highlightFill, draw=highlightLine
  - Neutral (grey): fill=neutralFill, draw=neutralLine`,
		MatplotlibColors:     "['#4682B4', '#3CB371', '#FFA500', '#DC143C', '#9370DB', '#20B2AA']",
		MatplotlibEdgeColors: "['#4682B4', '#3CB371', '#FFA500', '#DC143C', '#9370DB', '#20B2AA']",
		MermaidTheme:         "neutral",
	},
	"bold_contrast": {
		Name: "高对比",
		TikZColors: `\definecolor{accent1}{HTML}{E64B35}
\definecolor{accent2}{HTML}{4DBBD5}
\definecolor{accent3}{HTML}{00A087}
\definecolor{accent4}{HTML}{3C5488}
\definecolor{accent5}{HTML}{F39B7F}
\definecolor{bgcolor}{HTML}{FFFFFF}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (red): fill=primaryFill, draw=primaryLine
  - Secondary (cyan): fill=secondaryFill, draw=secondaryLine
  - Tertiary (teal): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (dark blue): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (salmon): fill=highlightFill, draw=highlightLine
  - Neutral (grey): fill=neutralFill, draw=neutralLine`,
		MatplotlibColors:     "['#E64B35', '#4DBBD5', '#00A087', '#3C5488', '#F39B7F', '#8491B4', '#91D1C2', '#DC0000', '#7E6148', '#B09C85']",
		MatplotlibEdgeColors: "['#E64B35', '#4DBBD5', '#00A087', '#3C5488', '#F39B7F', '#8491B4']",
		MermaidTheme:         "neutral",
	},
	"minimal_mono": {
		Name: "极简黑白",
		TikZColors: `\definecolor{accent1}{HTML}{0072B2}
\definecolor{accent2}{HTML}{D55E00}
\definecolor{accent3}{HTML}{009E73}
\definecolor{accent4}{HTML}{CC79A7}
\definecolor{accent5}{HTML}{F0E442}
\definecolor{bgcolor}{HTML}{FFFFFF}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (blue): fill=primaryFill, draw=primaryLine
  - Secondary (vermillion): fill=secondaryFill, draw=secondaryLine
  - Tertiary (green): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (pink): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (yellow): fill=highlightFill, draw=highlightLine
  - Neutral (grey): fill=neutralFill, draw=neutralLine`,
		MatplotlibColors:     "['#0072B2', '#D55E00', '#009E73', '#CC79A7', '#F0E442', '#56B4E9', '#E69F00', '#000000']",
		MatplotlibEdgeColors: "['#0072B2', '#D55E00', '#009E73', '#CC79A7', '#56B4E9', '#E69F00']",
		MermaidTheme:         "neutral",
	},
	"modern_teal": {
		Name: "现代青",
		TikZColors: `\definecolor{accent1}{HTML}{009688}
\definecolor{accent2}{HTML}{1976D2}
\definecolor{accent3}{HTML}{FF9800}
\definecolor{accent4}{HTML}{7B1FA2}
\definecolor{accent5}{HTML}{388E3C}
\definecolor{bgcolor}{HTML}{FAFAFA}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (teal): fill=primaryFill, draw=primaryLine
  - Secondary (blue): fill=secondaryFill, draw=secondaryLine
  - Tertiary (orange): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (purple): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (green): fill=highlightFill, draw=highlightLine
  - Neutral (grey): fill=neutralFill, draw=neutralLine`,
		MatplotlibColors:     "['#009688', '#1976D2', '#FF9800', '#7B1FA2', '#388E3C', '#00BCD4', '#FF5722', '#3F51B5']",
		MatplotlibEdgeColors: "['#009688', '#1976D2', '#FF9800', '#7B1FA2', '#388E3C', '#757575']",
		MermaidTheme:         "default",
	},
	"soft_pastel": {
		Name: "柔和粉彩",
		TikZColors: `\definecolor{accent1}{HTML}{B07D7D}
\definecolor{accent2}{HTML}{7D8DB0}
\definecolor{accent3}{HTML}{8DB07D}
\definecolor{accent4}{HTML}{B0A07D}
\definecolor{accent5}{HTML}{A07DB0}
\definecolor{bgcolor}{HTML}{F0EFED}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (rose): fill=primaryFill, draw=primaryLine
  - Secondary (blue-grey): fill=secondaryFill, draw=secondaryLine
  - Tertiary (sage): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (sand): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (lavender): fill=highlightFill, draw=highlightLine
  - Neutral (warm grey): fill=neutralFill, draw=neutralLine`,
		MatplotlibColors:     "['#B07D7D', '#7D8DB0', '#8DB07D', '#B0A07D', '#A07DB0', '#C4A882', '#8DAAB0', '#B08D7D']",
		MatplotlibEdgeColors: "['#B07D7D', '#7D8DB0', '#8DB07D', '#B0A07D', '#A07DB0', '#999999']",
		MermaidTheme:         "neutral",
	},
	"warm_earth": {
		Name: "暖色大地",
		TikZColors: `\definecolor{accent1}{HTML}{F9A825}
\definecolor{accent2}{HTML}{2E7D32}
\definecolor{accent3}{HTML}{BF360C}
\definecolor{accent4}{HTML}{6D4C41}
\definecolor{accent5}{HTML}{E65100}
\definecolor{bgcolor}{HTML}{FAF8F5}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (golden): fill=primaryFill, draw=primaryLine
  - Secondary (forest): fill=secondaryFill, draw=secondaryLine
  - Tertiary (brick red): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (brown): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (deep orange): fill=highlightFill, draw=highlightLine
  - Neutral (warm grey): fill=neutralFill, draw=neutralLine`,
		MatplotlibColors:     "['#F9A825', '#2E7D32', '#BF360C', '#6D4C41', '#E65100', '#FF8F00', '#1B5E20', '#D84315']",
		MatplotlibEdgeColors: "['#F9A825', '#2E7D32', '#BF360C', '#6D4C41', '#E65100', '#8D6E63']",
		MermaidTheme:         "default",
	},
	"cyber_dark": {
		Name: "深色科技",
		TikZColors: `\definecolor{accent1}{HTML}{448AFF}
\definecolor{accent2}{HTML}{00E676}
\definecolor{accent3}{HTML}{E040FB}
\definecolor{accent4}{HTML}{00BCD4}
\definecolor{accent5}{HTML}{FF6D00}
\definecolor{bgcolor}{HTML}{1B2631}`,
		TikZPrompt: `These colors are PRE-DEFINED in the preamble — do NOT redefine them, just USE them:
  - Primary (neon blue): fill=primaryFill, draw=primaryLine
  - Secondary (neon green): fill=secondaryFill, draw=secondaryLine
  - Tertiary (neon purple): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (cyan): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (neon orange): fill=highlightFill, draw=highlightLine
  - Neutral (slate): fill=neutralFill, draw=neutralLine
Note: This is a dark theme. Use the fill colors as dark backgrounds and line colors as bright accents.`,
		MatplotlibColors:     "['#448AFF', '#00E676', '#E040FB', '#00BCD4', '#FF6D00', '#FF4081', '#FFEA00', '#76FF03']",
		MatplotlibEdgeColors: "['#448AFF', '#00E676', '#E040FB', '#00BCD4', '#FF6D00', '#90A4AE']",
		MermaidTheme:         "dark",
	},
}

// Get returns the color scheme by name, falling back to "drawio" if not found.
func Get(name string) Scheme {
	if s, ok := schemes[name]; ok {
		return s
	}
	return schemes["drawio"]
}

// AllTikZColors returns a combined color definition block:
// 1. Unified semantic colors (primaryFill/Line etc.) from selected scheme
// 2. Legacy drawio colors (drawBlueFill etc.) for backward compatibility
// 3. Legacy accent colors from selected scheme for backward compatibility
func AllTikZColors(selected string) string {
	// Unified colors for selected scheme
	unified := unifiedColors["drawio"]
	if u, ok := unifiedColors[selected]; ok {
		unified = u
	}

	// Legacy drawio names (always included)
	drawio := schemes["drawio"].TikZColors

	// Legacy accent names from selected scheme
	if selected == "drawio" || selected == "" {
		return unified + "\n" + drawio
	}
	return unified + "\n" + drawio + "\n" + schemes[selected].TikZColors
}

// AllTikZColorsCustom returns a combined color definition block using custom colors.
// It includes the custom unified colors + legacy drawio colors for backward compatibility.
func AllTikZColorsCustom(c CustomColors) string {
	scheme := FromCustom(c)
	drawio := schemes["drawio"].TikZColors
	return scheme.TikZColors + "\n" + drawio
}

// Names returns all available scheme names.
func Names() []string {
	out := make([]string, 0, len(schemes))
	for k := range schemes {
		out = append(out, k)
	}
	return out
}
