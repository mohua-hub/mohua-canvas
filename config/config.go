package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	DatabaseDSN string `env:"DATABASE_DSN" envDefault:"data/infinite-canvas.db"`
	AILogDir    string `env:"AI_LOG_DIR" envDefault:"data/logs/ai-calls"`
}

var Cfg Config

func Load() error {
	_ = godotenv.Load()
	return env.Parse(&Cfg)
}

