package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/thesisviz/go-api/internal/service"
)

type RenderHandler struct {
	svc *service.RenderService
}

func NewRenderHandler(svc *service.RenderService) *RenderHandler {
	return &RenderHandler{svc: svc}
}

type renderRequest struct {
	Code         string `json:"code" binding:"required"`
	Format       string `json:"format" binding:"required"`
	Language     string `json:"language,omitempty"`
	ColorScheme  string `json:"color_scheme,omitempty"`
	GenerationID string `json:"generation_id,omitempty"`
	DPI          int    `json:"dpi,omitempty"`
	Timeout      int    `json:"timeout,omitempty"`
}

func (h *RenderHandler) Render(c *gin.Context) {
	var req renderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svcReq := service.RenderCodeRequest{
		Code:         req.Code,
		Format:       req.Format,
		Language:     req.Language,
		ColorScheme:  req.ColorScheme,
		GenerationID: req.GenerationID,
		DPI:          req.DPI,
		Timeout:      req.Timeout,
	}

	resp, err := h.svc.RenderCode(c.Request.Context(), svcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}

	if resp.Status == "error" {
		c.JSON(http.StatusUnprocessableEntity, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}
