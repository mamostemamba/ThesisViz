package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/thesisviz/go-api/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WSHandler struct {
	hub    *ws.Hub
	logger *zap.Logger
}

func NewWSHandler(hub *ws.Hub, logger *zap.Logger) *WSHandler {
	return &WSHandler{hub: hub, logger: logger}
}

// HandleGenerate upgrades HTTP to WebSocket for generation progress streaming.
// WS /api/v1/ws/generate/:taskId
func (h *WSHandler) HandleGenerate(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing task_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}

	h.hub.Register(taskID, conn)
	h.logger.Info("ws connected", zap.String("task_id", taskID))

	// Keep connection alive by reading (handle close/ping)
	defer func() {
		h.hub.Unregister(taskID)
		h.logger.Info("ws disconnected", zap.String("task_id", taskID))
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
