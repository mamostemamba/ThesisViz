package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/thesisviz/go-api/internal/agent"
	"github.com/thesisviz/go-api/internal/config"
	"github.com/thesisviz/go-api/internal/handler"
	"github.com/thesisviz/go-api/internal/llm"
	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/renderer"
	"github.com/thesisviz/go-api/internal/repo"
	"github.com/thesisviz/go-api/internal/router"
	"github.com/thesisviz/go-api/internal/service"
	"github.com/thesisviz/go-api/internal/storage"
	"github.com/thesisviz/go-api/internal/ws"
)

func main() {
	// Logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	// Config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Database
	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	logger.Info("connected to database")

	// AutoMigrate
	if err := db.AutoMigrate(&model.Project{}, &model.Generation{}); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}
	logger.Info("database migration completed")

	// Redis
	opt, err := redis.ParseURL(cfg.RedisUrl)
	if err != nil {
		logger.Warn("failed to parse redis URL, running without redis", zap.Error(err))
	}
	var rdb *redis.Client
	if opt != nil {
		rdb = redis.NewClient(opt)
	}

	// MinIO
	store, err := storage.NewMinIOStorage(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioBucket,
		cfg.MinioUseSSL,
	)
	if err != nil {
		logger.Fatal("failed to init minio", zap.Error(err))
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		logger.Warn("failed to ensure minio bucket (will retry on first upload)", zap.Error(err))
	} else {
		logger.Info("minio bucket ready", zap.String("bucket", cfg.MinioBucket))
	}

	// Repos
	projectRepo := repo.NewProjectRepo(db)
	generationRepo := repo.NewGenerationRepo(db)

	// Services
	projectSvc := service.NewProjectService(projectRepo)
	generationSvc := service.NewGenerationService(generationRepo)

	// Renderers
	tikzRenderer := renderer.NewTikZRenderer()
	matplotlibRenderer := renderer.NewMatplotlibRenderer(cfg.PyRenderURL)

	renderSvc := service.NewRenderService(tikzRenderer, matplotlibRenderer, store, generationSvc)

	// Handlers
	projectHandler := handler.NewProjectHandler(projectSvc)
	generationHandler := handler.NewGenerationHandler(generationSvc, store)
	renderHandler := handler.NewRenderHandler(renderSvc)

	// LLM + Agents (optional — only if GEMINI_API_KEY is set)
	var generateHandler *handler.GenerateHandler
	var colorExtractHandler *handler.ColorExtractHandler
	var wsHandler *handler.WSHandler

	if cfg.GeminiAPIKey != "" {
		geminiClient, err := llm.NewGeminiClient(context.Background(), cfg.GeminiAPIKey, cfg.GeminiModel)
		if err != nil {
			logger.Fatal("failed to create gemini client", zap.Error(err))
		}
		logger.Info("gemini client ready", zap.String("model", cfg.GeminiModel))

		agents := map[string]agent.Agent{
			"tikz":       agent.NewTikZAgent(geminiClient),
			"matplotlib": agent.NewMatplotlibAgent(geminiClient),
			"mermaid":    agent.NewMermaidAgent(geminiClient),
		}
		routerAgent := agent.NewRouterAgent(geminiClient, logger)

		agentSvc := service.NewAgentService(geminiClient, renderSvc, generationSvc, store, agents, routerAgent, logger)

		wsHub := ws.NewHub(logger)
		generateHandler = handler.NewGenerateHandler(agentSvc, generationSvc, store, wsHub, logger)
		colorExtractHandler = handler.NewColorExtractHandler(geminiClient, logger)
		wsHandler = handler.NewWSHandler(wsHub, logger)
	} else {
		logger.Warn("GEMINI_API_KEY not set — AI generation endpoints disabled")
	}

	// Router
	r := router.Setup(router.Deps{
		DB:                db,
		Redis:             rdb,
		Storage:           store,
		ProjectHandler:      projectHandler,
		GenerationHandler:   generationHandler,
		RenderHandler:       renderHandler,
		GenerateHandler:     generateHandler,
		ColorExtractHandler: colorExtractHandler,
		WSHandler:           wsHandler,
	})

	// Start
	addr := ":" + cfg.GoAPIPort
	logger.Info("starting server", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}
