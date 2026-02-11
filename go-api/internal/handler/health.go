package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status := "ok"
	details := gin.H{}

	// Check DB
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "degraded"
		details["db"] = err.Error()
	} else if err := sqlDB.PingContext(ctx); err != nil {
		status = "degraded"
		details["db"] = err.Error()
	} else {
		details["db"] = "ok"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			status = "degraded"
			details["redis"] = err.Error()
		} else {
			details["redis"] = "ok"
		}
	} else {
		details["redis"] = "not configured"
	}

	code := http.StatusOK
	if status != "ok" {
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"status":  status,
		"details": details,
	})
}
