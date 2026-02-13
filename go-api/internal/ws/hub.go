package ws

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub manages WebSocket connections keyed by task ID.
// Messages sent before the client connects are buffered and replayed on connect.
type Hub struct {
	mu      sync.Mutex
	conns   map[string]*websocket.Conn
	buffers map[string][][]byte // buffered JSON messages per task
	cancels map[string]context.CancelFunc
	logger  *zap.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		conns:   make(map[string]*websocket.Conn),
		buffers: make(map[string][][]byte),
		cancels: make(map[string]context.CancelFunc),
		logger:  logger,
	}
}

// RegisterCancel stores a cancel function for a task so it can be stopped later.
func (h *Hub) RegisterCancel(taskID string, cancel context.CancelFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cancels[taskID] = cancel
}

// CancelTask invokes the cancel function for a task and cleans up.
// Returns true if the task was found and cancelled.
func (h *Hub) CancelTask(taskID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	cancel, ok := h.cancels[taskID]
	if !ok {
		return false
	}
	cancel()
	delete(h.cancels, taskID)
	return true
}

// Register adds a WebSocket connection for a task and replays any buffered messages.
func (h *Hub) Register(taskID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close existing connection for this task if any
	if old, ok := h.conns[taskID]; ok {
		old.Close()
	}
	h.conns[taskID] = conn

	// Replay buffered messages
	if buf, ok := h.buffers[taskID]; ok {
		for _, data := range buf {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				h.logger.Warn("failed to replay buffered message", zap.String("task_id", taskID), zap.Error(err))
				break
			}
		}
		delete(h.buffers, taskID)
	}
}

// Unregister removes and closes the connection for a task.
func (h *Hub) Unregister(taskID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn, ok := h.conns[taskID]; ok {
		conn.Close()
		delete(h.conns, taskID)
	}
	delete(h.buffers, taskID)
	delete(h.cancels, taskID)
}

// Send serializes msg as JSON and writes to the task's WebSocket connection.
// If the client hasn't connected yet, the message is buffered for replay.
func (h *Hub) Send(taskID string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	conn, ok := h.conns[taskID]
	if !ok {
		// Buffer the message — client hasn't connected yet
		h.buffers[taskID] = append(h.buffers[taskID], data)
		return nil
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}
