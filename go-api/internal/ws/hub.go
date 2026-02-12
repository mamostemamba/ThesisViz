package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub manages WebSocket connections keyed by task ID.
type Hub struct {
	mu     sync.RWMutex
	conns  map[string]*websocket.Conn
	logger *zap.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		conns:  make(map[string]*websocket.Conn),
		logger: logger,
	}
}

// Register adds a WebSocket connection for a task.
func (h *Hub) Register(taskID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Close existing connection for this task if any
	if old, ok := h.conns[taskID]; ok {
		old.Close()
	}
	h.conns[taskID] = conn
}

// Unregister removes and closes the connection for a task.
func (h *Hub) Unregister(taskID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, ok := h.conns[taskID]; ok {
		conn.Close()
		delete(h.conns, taskID)
	}
}

// Send serializes msg as JSON and writes to the task's WebSocket connection.
func (h *Hub) Send(taskID string, msg interface{}) error {
	h.mu.RLock()
	conn, ok := h.conns[taskID]
	h.mu.RUnlock()

	if !ok {
		// No active connection — not an error, client may not have connected yet
		h.logger.Debug("no ws connection for task", zap.String("task_id", taskID))
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
