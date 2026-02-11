package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/thesisviz/go-api/internal/renderer"
	"github.com/thesisviz/go-api/internal/storage"
	"github.com/thesisviz/go-api/pkg/colorscheme"
)

type RenderService struct {
	tikz       renderer.Renderer
	matplotlib renderer.Renderer
	storage    *storage.MinIOStorage
	genSvc     *GenerationService
}

func NewRenderService(
	tikz renderer.Renderer,
	matplotlib renderer.Renderer,
	store *storage.MinIOStorage,
	genSvc *GenerationService,
) *RenderService {
	return &RenderService{
		tikz:       tikz,
		matplotlib: matplotlib,
		storage:    store,
		genSvc:     genSvc,
	}
}

type RenderCodeRequest struct {
	Code         string `json:"code"`
	Format       string `json:"format"`       // "tikz" or "matplotlib"
	Language     string `json:"language"`      // "en" or "zh"
	ColorScheme  string `json:"color_scheme"`  // e.g. "drawio", "academic_blue"
	GenerationID string `json:"generation_id"` // optional: update existing generation
	DPI          int    `json:"dpi"`
	Timeout      int    `json:"timeout"`
}

type RenderCodeResponse struct {
	Status   string `json:"status"`
	ImageURL string `json:"image_url,omitempty"`
	ImageKey string `json:"image_key,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (s *RenderService) RenderCode(ctx context.Context, req RenderCodeRequest) (*RenderCodeResponse, error) {
	// Pick renderer
	var r renderer.Renderer
	switch req.Format {
	case "tikz":
		r = s.tikz
	case "matplotlib":
		r = s.matplotlib
	default:
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	// Build render options
	opts := renderer.RenderOpts{
		Language: req.Language,
		DPI:      req.DPI,
		Timeout:  req.Timeout,
	}

	// Apply color scheme colors for TikZ — always include all drawio colors
	// plus the selected scheme's accent colors so any code compiles
	if req.Format == "tikz" {
		opts.Colors = colorscheme.AllTikZColors(req.ColorScheme)
	}

	// Render
	result, err := r.Render(ctx, req.Code, opts)
	if err != nil {
		return &RenderCodeResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	// Upload to MinIO
	imageKey := fmt.Sprintf("generations/%s/render.png", uuid.New().String())
	if err := s.storage.Upload(ctx, imageKey, result.ImageBytes, "image/png"); err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	// Generate presigned URL
	imageURL, err := s.storage.PresignedURL(ctx, imageKey)
	if err != nil {
		return nil, fmt.Errorf("generate presigned url: %w", err)
	}

	// Optionally update generation record
	if req.GenerationID != "" && s.genSvc != nil {
		genID, err := uuid.Parse(req.GenerationID)
		if err == nil {
			g, err := s.genSvc.GetByID(genID)
			if err == nil {
				g.ImageKey = &imageKey
				g.Status = "success"
				g.Code = &req.Code
				_ = s.genSvc.Update(g)
			}
		}
	}

	return &RenderCodeResponse{
		Status:   "ok",
		ImageURL: imageURL,
		ImageKey: imageKey,
	}, nil
}
