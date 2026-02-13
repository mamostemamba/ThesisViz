package renderer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thesisviz/go-api/pkg/sanitize"
)

const defaultColors = `\definecolor{accent1}{RGB}{70,130,180}
\definecolor{accent2}{RGB}{60,179,113}
\definecolor{accent3}{RGB}{255,165,0}
\definecolor{accent4}{RGB}{220,20,60}
\definecolor{bgcolor}{RGB}{245,245,250}`

const texTemplate = `\documentclass[border=20pt]{standalone}
\usepackage{tikz}
%s
\usetikzlibrary{arrows.meta,shapes.geometric,shapes.multipart,positioning,calc,fit,backgrounds,shadows,decorations.pathmorphing,decorations.pathreplacing,decorations.markings,patterns,matrix,chains,scopes}
%s

%% ---- Pre-defined modern styles (referenced in LLM system prompt) ----
\tikzset{
    %% Base box: rounded rectangle with light drop shadow (matches V1 academic style)
    modern_box/.style={
        rectangle,
        rounded corners=3pt,
        align=center,
        minimum height=0.9cm,
        minimum width=2.8cm,
        inner sep=5pt,
        thick,
        font=\sffamily\footnotesize,
        drop shadow={opacity=0.08},
    },
    %% Professional arrow with Stealth tip — for ALL connections
    nice_arrow/.style={
        ->,
        >={Stealth[length=5pt, width=4pt]},
        thick,
        rounded corners=3pt,
        draw=black!70,
    },
    %% Bidirectional arrow
    nice_biarrow/.style={
        <->,
        >={Stealth[length=5pt, width=4pt]},
        thick,
        rounded corners=3pt,
        draw=black!70,
    },
    %% Dashed container for grouping nodes (use with fit)
    container_box/.style={
        rectangle,
        rounded corners=6pt,
        draw=neutralLine,
        dashed,
        thick,
        fill=neutralFill,
        fill opacity=0.3,
        inner sep=10pt,
        font=\sffamily\small,
    },
    %% Layer row box: solid background container for a row of matrix nodes
    layer_box/.style={
        rectangle,
        rounded corners=8pt,
        draw=#1!80!black,
        fill=#1!6,
        thick,
        inner sep=12pt,
        font=\sffamily\small\bfseries,
    },
    %% Matrix default node style
    matrix_node/.style={
        modern_box,
        minimum width=3cm,
        minimum height=1.0cm,
    },
    %% Visible brace: extra-thick curly brace for grouping (survives image compression)
    visible_brace/.style={
        decorate,
        decoration={brace, amplitude=10pt, raise=4pt},
        very thick,
        draw=black!90,
    },
    %% Visible brace (mirrored, opening to the left)
    visible_brace_mirror/.style={
        decorate,
        decoration={brace, mirror, amplitude=10pt, raise=4pt},
        very thick,
        draw=black!90,
    },
    %% Force all decoration lines to be thick enough for visual review
    every decoration/.style={very thick},
}

\begin{document}
%s
\end{document}
`

type TikZRenderer struct{}

func NewTikZRenderer() *TikZRenderer {
	return &TikZRenderer{}
}

func (r *TikZRenderer) Render(ctx context.Context, code string, opts RenderOpts) (*RenderResult, error) {
	if _, err := exec.LookPath("pdflatex"); err != nil {
		return nil, fmt.Errorf("pdflatex not found: install a TeX distribution (e.g. brew install --cask mactex)")
	}

	lang := opts.Language
	if lang == "" {
		lang = "en"
	}
	dpi := opts.DPI
	if dpi <= 0 {
		dpi = 300
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	colors := opts.Colors
	if colors == "" {
		colors = defaultColors
	}

	cleanCode := sanitize.TikZClean(code)
	ctexLine := ""
	if lang == "zh" {
		ctexLine = `\usepackage{ctex}`
	}
	texContent := fmt.Sprintf(texTemplate, ctexLine, colors, cleanCode)

	// Choose compiler
	compiler := "pdflatex"
	if lang == "zh" {
		if _, err := exec.LookPath("xelatex"); err == nil {
			compiler = "xelatex"
		}
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tikz-render-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	texPath := filepath.Join(tmpDir, "fig.tex")
	pdfPath := filepath.Join(tmpDir, "fig.pdf")

	if err := os.WriteFile(texPath, []byte(texContent), 0644); err != nil {
		return nil, fmt.Errorf("write tex: %w", err)
	}

	// Compile
	cmd := exec.CommandContext(ctx, compiler, "-interaction=nonstopmode", "-output-directory", tmpDir, texPath)
	output, err := cmd.CombinedOutput()
	if _, statErr := os.Stat(pdfPath); os.IsNotExist(statErr) {
		// Truncate output for error message
		outStr := string(output)
		if len(outStr) > 1200 {
			outStr = outStr[len(outStr)-1200:]
		}
		return nil, fmt.Errorf("%s compilation failed:\n%s", compiler, outStr)
	}

	// Convert PDF to PNG using pdftoppm (from poppler)
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm not found: install poppler (brew install poppler)")
	}

	pngPrefix := filepath.Join(tmpDir, "output")
	ppmCmd := exec.CommandContext(ctx, "pdftoppm",
		"-png",
		"-r", fmt.Sprintf("%d", dpi),
		"-singlefile",
		pdfPath,
		pngPrefix,
	)
	if ppmOut, err := ppmCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %s\n%s", err, string(ppmOut))
	}

	pngPath := pngPrefix + ".png"
	pngData, err := os.ReadFile(pngPath)
	if err != nil {
		return nil, fmt.Errorf("read png: %w", err)
	}

	return &RenderResult{
		ImageBytes: pngData,
		Format:     "png",
	}, nil
}

// BuildFullTeX returns the complete .tex source for a TikZ code snippet.
func BuildFullTeX(tikzCode, colorDefs, language string) string {
	cleanCode := sanitize.TikZClean(tikzCode)
	ctexLine := ""
	if strings.EqualFold(language, "zh") {
		ctexLine = `\usepackage{ctex}`
	}
	if colorDefs == "" {
		colorDefs = defaultColors
	}
	return fmt.Sprintf(texTemplate, ctexLine, colorDefs, cleanCode)
}
