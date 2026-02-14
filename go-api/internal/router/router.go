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

	ProjectHandler      *handler.ProjectHandler
	GenerationHandler   *handler.GenerationHandler
	RenderHandler       *handler.RenderHandler
	GenerateHandler     *handler.GenerateHandler
	ColorExtractHandler *handler.ColorExtractHandler
	WSHandler           *handler.WSHandler
	ConfigHandler       *handler.ConfigHandler
}

func Setup(deps Deps) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
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
			v1.POST("/export/tex", deps.RenderHandler.ExportTeX)
		}

		// Generate (AI pipeline)
		v1.POST("/generate/analyze", deps.GenerateHandler.Analyze)
		v1.POST("/generate/drawing-prompt", deps.GenerateHandler.DrawingPrompt)
		v1.POST("/generate/create", deps.GenerateHandler.Create)
		v1.POST("/generate/refine", deps.GenerateHandler.Refine)
		v1.POST("/generate/cancel/:taskId", deps.GenerateHandler.Cancel)
		v1.GET("/generate/:id", deps.GenerateHandler.Get)

		// Color extraction
		v1.POST("/colors/extract", deps.ColorExtractHandler.Extract)

		// WebSocket
		v1.GET("/ws/generate/:taskId", deps.WSHandler.HandleGenerate)

		// Config
		v1.POST("/config/api-key", deps.ConfigHandler.SetAPIKey)
		v1.GET("/config/status", deps.ConfigHandler.Status)
	}

	return r
}
