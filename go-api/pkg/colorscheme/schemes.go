package colorscheme

// Scheme holds all color-related configuration for a single visual theme.
type Scheme struct {
	Name                 string
	TikZColors           string // legacy scheme-specific color names
	TikZPrompt           string
	MatplotlibColors     string
	MatplotlibEdgeColors string
	MermaidTheme         string
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

	"academic_blue": `\definecolor{primaryFill}{HTML}{DAE6F0}
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

	"nature": `\definecolor{primaryFill}{HTML}{FADBD7}
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

	"ieee": `\definecolor{primaryFill}{HTML}{CCE3F0}
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
}

var schemes = map[string]Scheme{
	"drawio": {
		Name: "Draw.io Classic",
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
  - Primary (blue): fill=primaryFill, draw=primaryLine
  - Secondary (green): fill=secondaryFill, draw=secondaryLine
  - Tertiary (orange): fill=tertiaryFill, draw=tertiaryLine
  - Quaternary (purple): fill=quaternaryFill, draw=quaternaryLine
  - Highlight (red): fill=highlightFill, draw=highlightLine
  - Neutral (grey): fill=neutralFill, draw=neutralLine
Example style: base_box/.style={rectangle, rounded corners=3pt, align=center, minimum height=0.9cm, minimum width=2.8cm, drop shadow={opacity=0.15}, thick}`,
		MatplotlibColors:     "['#DAE8FC', '#D5E8D4', '#FFE6CC', '#E1D5E7', '#F8CECC', '#6C8EBF', '#82B366', '#D79B00', '#9673A6', '#B85450']",
		MatplotlibEdgeColors: "['#6C8EBF', '#82B366', '#D79B00', '#9673A6', '#B85450', '#666666']",
		MermaidTheme:         "default",
	},
	"academic_blue": {
		Name: "Academic Blue",
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
	"nature": {
		Name: "Nature Journal",
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
	"ieee": {
		Name: "IEEE Minimal",
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

// Names returns all available scheme names.
func Names() []string {
	out := make([]string, 0, len(schemes))
	for k := range schemes {
		out = append(out, k)
	}
	return out
}
