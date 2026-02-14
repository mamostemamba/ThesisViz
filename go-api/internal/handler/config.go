package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/thesisviz/go-api/internal/llm"
)

type ConfigHandler struct {
	gemini *llm.GeminiClient
	model  string
}

func NewConfigHandler(gemini *llm.GeminiClient, model string) *ConfigHandler {
	return &ConfigHandler{gemini: gemini, model: model}
}

type setAPIKeyRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// SetAPIKey handles POST /api/v1/config/api-key
func (h *ConfigHandler) SetAPIKey(c *gin.Context) {
	var req setAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.gemini.SetAPIKey(c.Request.Context(), req.APIKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid API key: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Status handles GET /api/v1/config/status
func (h *ConfigHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"has_api_key": h.gemini.HasKey(),
		"model":       h.model,
	})
}
