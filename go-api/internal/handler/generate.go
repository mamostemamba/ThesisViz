package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/thesisviz/go-api/internal/service"
	"github.com/thesisviz/go-api/internal/storage"
	"github.com/thesisviz/go-api/internal/ws"
)

type GenerateHandler struct {
	agentSvc *service.AgentService
	genSvc   *service.GenerationService
	storage  *storage.MinIOStorage
	hub      *ws.Hub
}

func NewGenerateHandler(agentSvc *service.AgentService, genSvc *service.GenerationService, store *storage.MinIOStorage, hub *ws.Hub) *GenerateHandler {
	return &GenerateHandler{agentSvc: agentSvc, genSvc: genSvc, storage: store, hub: hub}
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
		req.Language = "en"
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
	ProjectID      string `json:"project_id"`
	Format         string `json:"format" binding:"required"`
	Prompt         string `json:"prompt" binding:"required"`
	Language       string `json:"language"`
	ColorScheme    string `json:"color_scheme"`
	ThesisTitle    string `json:"thesis_title"`
	ThesisAbstract string `json:"thesis_abstract"`
	Model          string `json:"model"`
}

// Create handles POST /api/v1/generate/create
func (h *GenerateHandler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Language == "" {
		req.Language = "en"
	}
	if req.ColorScheme == "" {
		req.ColorScheme = "academic_blue"
	}

	taskID := uuid.New().String()

	// Launch pipeline in background goroutine
	go func() {
		pushFn := func(msg service.ProgressMsg) {
			_ = h.hub.Send(taskID, msg)
		}

		_, err := h.agentSvc.Generate(c.Request.Context(), service.GenerateRequest{
			ProjectID:      req.ProjectID,
			Format:         req.Format,
			Prompt:         req.Prompt,
			Language:       req.Language,
			ColorScheme:    req.ColorScheme,
			ThesisTitle:    req.ThesisTitle,
			ThesisAbstract: req.ThesisAbstract,
			Model:          req.Model,
		}, pushFn)

		if err != nil {
			_ = h.hub.Send(taskID, service.ProgressMsg{
				Type:  "error",
				Phase: "done",
				Data:  service.ProgressData{Message: err.Error()},
			})
		}
	}()

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}

type refineRequestBody struct {
	GenerationID string `json:"generation_id" binding:"required"`
	Modification string `json:"modification" binding:"required"`
	Language     string `json:"language"`
	ColorScheme  string `json:"color_scheme"`
	Model        string `json:"model"`
}

// Refine handles POST /api/v1/generate/refine
func (h *GenerateHandler) Refine(c *gin.Context) {
	var req refineRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Language == "" {
		req.Language = "en"
	}
	if req.ColorScheme == "" {
		req.ColorScheme = "academic_blue"
	}

	taskID := uuid.New().String()

	go func() {
		pushFn := func(msg service.ProgressMsg) {
			_ = h.hub.Send(taskID, msg)
		}

		_, err := h.agentSvc.Refine(c.Request.Context(), service.RefineRequest{
			GenerationID: req.GenerationID,
			Modification: req.Modification,
			Language:     req.Language,
			ColorScheme:  req.ColorScheme,
			Model:        req.Model,
		}, pushFn)

		if err != nil {
			_ = h.hub.Send(taskID, service.ProgressMsg{
				Type:  "error",
				Phase: "done",
				Data:  service.ProgressData{Message: err.Error()},
			})
		}
	}()

	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
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
