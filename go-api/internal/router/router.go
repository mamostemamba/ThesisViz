package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/thesisviz/go-api/internal/handler"
	"github.com/thesisviz/go-api/internal/storage"
)

type Deps struct {
	DB      *gorm.DB
	Redis   *redis.Client
	Storage *storage.MinIOStorage

	ProjectHandler    *handler.ProjectHandler
	GenerationHandler *handler.GenerationHandler
	RenderHandler     *handler.RenderHandler
}

func Setup(deps Deps) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health
	health := handler.NewHealthHandler(deps.DB, deps.Redis)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", health.Check)

		// Projects
		if deps.ProjectHandler != nil {
			v1.POST("/projects", deps.ProjectHandler.Create)
			v1.GET("/projects", deps.ProjectHandler.List)
			v1.GET("/projects/:id", deps.ProjectHandler.Get)
			v1.PUT("/projects/:id", deps.ProjectHandler.Update)
			v1.DELETE("/projects/:id", deps.ProjectHandler.Delete)
		}

		// Generations
		if deps.GenerationHandler != nil {
			v1.POST("/projects/:id/generations", deps.GenerationHandler.Create)
			v1.GET("/projects/:id/generations", deps.GenerationHandler.ListByProject)
			v1.GET("/generations/:id", deps.GenerationHandler.Get)
			v1.DELETE("/generations/:id", deps.GenerationHandler.Delete)
		}

		// Render
		if deps.RenderHandler != nil {
			v1.POST("/render", deps.RenderHandler.Render)
		}
	}

	return r
}
