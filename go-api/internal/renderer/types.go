package renderer

import "context"

type Renderer interface {
	Render(ctx context.Context, code string, opts RenderOpts) (*RenderResult, error)
}

type RenderOpts struct {
	Language string // "en" or "zh"
	DPI      int    // default 300
	Timeout  int    // seconds, default 60
	Colors   string // TikZ color definitions or matplotlib palette
	Style    string // "professional" (default) or "handdrawn"
}

type RenderResult struct {
	ImageBytes []byte
	Format     string // "png"
}
