package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUrl        string `mapstructure:"DB_URL"`
	RedisUrl     string `mapstructure:"REDIS_URL"`
	MinioEndpoint  string `mapstructure:"MINIO_ENDPOINT"`
	MinioAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinioBucket    string `mapstructure:"MINIO_BUCKET"`
	MinioUseSSL    bool   `mapstructure:"MINIO_USE_SSL"`
	GeminiAPIKey   string `mapstructure:"GEMINI_API_KEY"`
	GoAPIPort      string `mapstructure:"GO_API_PORT"`
	PyRenderURL    string `mapstructure:"PY_RENDER_URL"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("GO_API_PORT", "8080")
	viper.SetDefault("PY_RENDER_URL", "http://localhost:8081")
	viper.SetDefault("DB_URL", "postgres://thesisviz:thesisviz_dev@localhost:5432/thesisviz?sslmode=disable")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379/0")
	viper.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	viper.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	viper.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	viper.SetDefault("MINIO_BUCKET", "thesisviz")
	viper.SetDefault("MINIO_USE_SSL", false)

	// .env file is optional; env vars take precedence
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
