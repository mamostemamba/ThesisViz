package colorscheme

// Scheme holds all color-related configuration for a single visual theme.
type Scheme struct {
	Name                string
	TikZColors          string
	TikZPrompt          string
	MatplotlibColors    string
	MatplotlibEdgeColors string
	MermaidTheme        string
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
  - Blue nodes: fill=drawBlueFill, draw=drawBlueLine
  - Green nodes: fill=drawGreenFill, draw=drawGreenLine
  - Orange nodes: fill=drawOrangeFill, draw=drawOrangeLine
  - Purple nodes: fill=drawPurpleFill, draw=drawPurpleLine
  - Red nodes: fill=drawRedFill, draw=drawRedLine
  - Grey nodes: fill=drawGreyFill, draw=drawGreyLine
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
  - accent1 = steel blue (primary nodes): fill=accent1!20, draw=accent1
  - accent2 = sea green (secondary): fill=accent2!20, draw=accent2
  - accent3 = orange (highlights): fill=accent3!20, draw=accent3
  - accent4 = crimson (alerts): fill=accent4!20, draw=accent4
  - bgcolor = light background`,
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
  - accent1 = red: fill=accent1!20, draw=accent1
  - accent2 = cyan: fill=accent2!20, draw=accent2
  - accent3 = teal: fill=accent3!20, draw=accent3
  - accent4 = dark blue: fill=accent4!20, draw=accent4
  - accent5 = salmon: fill=accent5!20, draw=accent5`,
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
  - accent1 = blue: fill=accent1!20, draw=accent1
  - accent2 = vermillion: fill=accent2!20, draw=accent2
  - accent3 = green: fill=accent3!20, draw=accent3
  - accent4 = pink: fill=accent4!20, draw=accent4
  - accent5 = yellow: fill=accent5!20, draw=accent5`,
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

// Names returns all available scheme names.
func Names() []string {
	out := make([]string, 0, len(schemes))
	for k := range schemes {
		out = append(out, k)
	}
	return out
}
