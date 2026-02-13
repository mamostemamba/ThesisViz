package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/service"
	"github.com/thesisviz/go-api/internal/storage"
	"github.com/thesisviz/go-api/internal/ws"
	"github.com/thesisviz/go-api/pkg/colorscheme"
)

const (
	defaultLanguage    = "zh"
	defaultColorScheme = "drawio"
	taskTimeout        = 15 * time.Minute
)

type GenerateHandler struct {
	agentSvc *service.AgentService
	genSvc   *service.GenerationService
	storage  *storage.MinIOStorage
	hub      *ws.Hub
	logger   *zap.Logger
}

func NewGenerateHandler(agentSvc *service.AgentService, genSvc *service.GenerationService, store *storage.MinIOStorage, hub *ws.Hub, logger *zap.Logger) *GenerateHandler {
	return &GenerateHandler{agentSvc: agentSvc, genSvc: genSvc, storage: store, hub: hub, logger: logger}
}

type analyzeRequest struct {
	Text           string `json:"text" binding:"required"`
	Language       string `json:"language"`
	ThesisTitle    string `json:"thesis_title"`
	ThesisAbstract string `json:"thesis_abstract"`
	Model          string `json:"model"`
}

// Analyze handles POST /api/v1/generate/analyze
func (h *GenerateHandler) Analyze(c *gin.Context) {
	var req analyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Language == "" {
		req.Language = defaultLanguage
	}

	recs, err := h.agentSvc.Analyze(c.Request.Context(), service.AnalyzeRequest{
		Text:           req.Text,
		Language:       req.Language,
		ThesisTitle:    req.ThesisTitle,
		ThesisAbstract: req.ThesisAbstract,
		Model:          req.Model,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": recs})
}

type createRequest struct {
	ProjectID      string                   `json:"project_id"`
	Format         string                   `json:"format" binding:"required"`
	Prompt         string                   `json:"prompt" binding:"required"`
	Language       string                   `json:"language"`
	ColorScheme    string                   `json:"color_scheme"`
	CustomColors   *colorscheme.CustomColors `json:"custom_colors,omitempty"`
	ThesisTitle    string                   `json:"thesis_title"`
	ThesisAbstract string                   `json:"thesis_abstract"`
	Model          string                   `json:"model"`
	Identity       string                   `json:"identity"`
}

// Create handles POST /api/v1/generate/create
func (h *GenerateHandler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Language == "" {
		req.Language = defaultLanguage
	}
	if req.ColorScheme == "" {
		req.ColorScheme = defaultColorScheme
	}

	taskID := uuid.New().String()

	// Copy request data for background goroutine — c.Request.Context() is
	// canceled once the HTTP response is sent, so we use a detached context
	// with a timeout to prevent runaway goroutines.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
		defer cancel()
		h.hub.RegisterCancel(taskID, cancel)

		pushFn := func(msg service.ProgressMsg) {
			_ = h.hub.Send(taskID, msg)
		}

		_, err := h.agentSvc.Generate(ctx, service.GenerateRequest{
			ProjectID:      req.ProjectID,
			Format:         req.Format,
			Prompt:         req.Prompt,
			Language:       req.Language,
			ColorScheme:    req.ColorScheme,
			CustomColors:   req.CustomColors,
			ThesisTitle:    req.ThesisTitle,
			ThesisAbstract: req.ThesisAbstract,
			Model:          req.Model,
			Identity:       req.Identity,
		}, pushFn)

		if err != nil {
			h.logger.Error("generation failed", zap.String("task_id", taskID), zap.Error(err))
			if ctx.Err() == nil {
				// Only send error if not cancelled — cancelled tasks send their own message
				_ = h.hub.Send(taskID, service.ProgressMsg{
					Type:  "error",
					Phase: "done",
					Data:  service.ProgressData{Message: err.Error()},
				})
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}

type refineRequestBody struct {
	GenerationID string                   `json:"generation_id" binding:"required"`
	Modification string                   `json:"modification" binding:"required"`
	Language     string                   `json:"language"`
	ColorScheme  string                   `json:"color_scheme"`
	CustomColors *colorscheme.CustomColors `json:"custom_colors,omitempty"`
	Model        string                   `json:"model"`
}

// Refine handles POST /api/v1/generate/refine
func (h *GenerateHandler) Refine(c *gin.Context) {
	var req refineRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Language == "" {
		req.Language = defaultLanguage
	}
	if req.ColorScheme == "" {
		req.ColorScheme = defaultColorScheme
	}

	taskID := uuid.New().String()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
		defer cancel()
		h.hub.RegisterCancel(taskID, cancel)

		pushFn := func(msg service.ProgressMsg) {
			_ = h.hub.Send(taskID, msg)
		}

		_, err := h.agentSvc.Refine(ctx, service.RefineRequest{
			GenerationID: req.GenerationID,
			Modification: req.Modification,
			Language:     req.Language,
			ColorScheme:  req.ColorScheme,
			CustomColors: req.CustomColors,
			Model:        req.Model,
		}, pushFn)

		if err != nil {
			h.logger.Error("refine failed", zap.String("task_id", taskID), zap.Error(err))
			if ctx.Err() == nil {
				_ = h.hub.Send(taskID, service.ProgressMsg{
					Type:  "error",
					Phase: "done",
					Data:  service.ProgressData{Message: err.Error()},
				})
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}

// Cancel handles POST /api/v1/generate/cancel/:taskId
func (h *GenerateHandler) Cancel(c *gin.Context) {
	taskID := c.Param("taskId")
	// Send cancelled message before cancelling context to avoid race
	_ = h.hub.Send(taskID, service.ProgressMsg{
		Type:  "cancelled",
		Phase: "done",
		Data:  service.ProgressData{Message: "任务已终止"},
	})
	h.hub.CancelTask(taskID)
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// Get handles GET /api/v1/generate/:id
func (h *GenerateHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	gen, err := h.genSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Add presigned URL if image exists
	resp := gin.H{
		"id":            gen.ID,
		"project_id":    gen.ProjectID,
		"parent_id":     gen.ParentID,
		"format":        gen.Format,
		"prompt":        gen.Prompt,
		"status":        gen.Status,
		"code":          gen.Code,
		"explanation":   gen.Explanation,
		"review_issues": gen.ReviewIssues,
		"created_at":    gen.CreatedAt,
	}
	if gen.ImageKey != nil && *gen.ImageKey != "" {
		url, err := h.storage.PresignedURL(c.Request.Context(), *gen.ImageKey)
		if err == nil {
			resp["image_url"] = url
		}
	}

	c.JSON(http.StatusOK, resp)
}
