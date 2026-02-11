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
\usepackage[utf8]{inputenc}
\usepackage{tikz}
%s
\usetikzlibrary{arrows.meta,shapes.geometric,positioning,calc,fit,backgrounds,shadows}
%s
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

	cleanCode := sanitize.TikZ(code)
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
	cleanCode := sanitize.TikZ(tikzCode)
	ctexLine := ""
	if strings.EqualFold(language, "zh") {
		ctexLine = `\usepackage{ctex}`
	}
	if colorDefs == "" {
		colorDefs = defaultColors
	}
	return fmt.Sprintf(texTemplate, ctexLine, colorDefs, cleanCode)
}
