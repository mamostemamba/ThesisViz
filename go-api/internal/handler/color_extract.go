package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/prompt"
	"github.com/thesisviz/go-api/pkg/colorscheme"
)

const maxImageSize = 5 << 20 // 5 MB

type ColorExtractHandler struct {
	llm    *llm.GeminiClient
	logger *zap.Logger
}

func NewColorExtractHandler(llmClient *llm.GeminiClient, logger *zap.Logger) *ColorExtractHandler {
	return &ColorExtractHandler{llm: llmClient, logger: logger}
}

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func validHex(s string) bool {
	return hexColorRe.MatchString(s)
}

// Extract handles POST /api/v1/colors/extract
func (h *ColorExtractHandler) Extract(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image file"})
		return
	}
	defer file.Close()

	if header.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("image too large (max %d MB)", maxImageSize>>20)})
		return
	}

	imgBytes, err := io.ReadAll(io.LimitReader(file, maxImageSize+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read image"})
		return
	}
	if int64(len(imgBytes)) > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("image too large (max %d MB)", maxImageSize>>20)})
		return
	}

	h.logger.Info("extracting colors from image", zap.Int("bytes", len(imgBytes)))

	raw, err := h.llm.GenerateWithImage(
		c.Request.Context(),
		prompt.ColorExtractSystem(),
		prompt.ColorExtractUser(),
		imgBytes,
		0.3,
	)
	if err != nil {
		h.logger.Error("gemini color extract failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze image colors"})
		return
	}

	// Strip markdown code fences if present
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		if len(lines) > 2 {
			lines = lines[1 : len(lines)-1]
		}
		raw = strings.Join(lines, "\n")
	}
	raw = strings.TrimSpace(raw)

	// Parse JSON array of {fill, line} pairs
	var pairs []colorscheme.ColorPair
	if err := json.Unmarshal([]byte(raw), &pairs); err != nil {
		h.logger.Error("failed to parse color extract response", zap.String("raw", raw), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse extracted colors"})
		return
	}

	// Validate count
	if len(pairs) < 4 || len(pairs) > 8 {
		h.logger.Error("unexpected color count", zap.Int("count", len(pairs)), zap.String("raw", raw))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("expected 4-8 color pairs, got %d", len(pairs))})
		return
	}

	// Validate all hex colors
	for i, p := range pairs {
		if !validHex(p.Fill) || !validHex(p.Line) {
			h.logger.Error("invalid hex in extracted colors", zap.Int("index", i), zap.String("raw", raw))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "extracted colors contain invalid hex values"})
			return
		}
	}

	colors := colorscheme.CustomColors{Pairs: pairs}
	c.JSON(http.StatusOK, gin.H{"colors": colors})
}
