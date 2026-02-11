package main

import (
	"log"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/thesisviz/go-api/internal/config"
	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/router"
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

	// Router
	r := router.Setup(db, rdb)

	// Start
	addr := ":" + cfg.GoAPIPort
	logger.Info("starting server", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}
