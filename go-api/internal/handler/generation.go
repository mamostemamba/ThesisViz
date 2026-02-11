package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/service"
	"github.com/thesisviz/go-api/internal/storage"
)

type GenerationHandler struct {
	svc     *service.GenerationService
	storage *storage.MinIOStorage
}

func NewGenerationHandler(svc *service.GenerationService, store *storage.MinIOStorage) *GenerationHandler {
	return &GenerationHandler{svc: svc, storage: store}
}

type createGenerationRequest struct {
	Format   string  `json:"format" binding:"required"`
	Prompt   string  `json:"prompt" binding:"required"`
	ParentID *string `json:"parent_id,omitempty"`
	Code     *string `json:"code,omitempty"`
}

func (h *GenerationHandler) Create(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req createGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g := model.Generation{
		ProjectID: projectID,
		Format:    req.Format,
		Prompt:    req.Prompt,
		Code:      req.Code,
	}
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_id"})
			return
		}
		g.ParentID = &pid
	}

	if err := h.svc.Create(&g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, g)
}

func (h *GenerationHandler) ListByProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListByProject(projectID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *GenerationHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid generation id"})
		return
	}

	g, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generation not found"})
		return
	}

	// Generate presigned URL if image_key exists
	resp := gin.H{
		"id":            g.ID,
		"project_id":    g.ProjectID,
		"parent_id":     g.ParentID,
		"format":        g.Format,
		"prompt":        g.Prompt,
		"status":        g.Status,
		"code":          g.Code,
		"image_key":     g.ImageKey,
		"explanation":   g.Explanation,
		"review_issues": g.ReviewIssues,
		"created_at":    g.CreatedAt,
	}

	if g.ImageKey != nil && h.storage != nil {
		url, err := h.storage.PresignedURL(c.Request.Context(), *g.ImageKey)
		if err == nil {
			resp["image_url"] = url
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GenerationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid generation id"})
		return
	}

	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
