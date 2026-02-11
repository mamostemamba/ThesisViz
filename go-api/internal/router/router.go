package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/thesisviz/go-api/internal/handler"
)

func Setup(db *gorm.DB, rdb *redis.Client) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health
	health := handler.NewHealthHandler(db, rdb)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", health.Check)
	}

	return r
}
