package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DBHost     string `env:"DB_HOST,required"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBName     string `env:"DB_NAME,required"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
	JWTSecret  string `env:"JWT_SECRET,required"`

	// Cloudflare R2 Config
	R2AccountID string `env:"R2_ACCOUNT_ID,required"`
	R2AccessKey string `env:"R2_ACCESS_KEY,required"`
	R2SecretKey string `env:"R2_SECRET_KEY,required"`
	R2Bucket    string `env:"R2_BUCKET,required"`
	R2PublicURL string `env:"R2_PUBLIC_URL,required"`

	// DeepSeek AI
	DeepSeekAPIKey string `env:"DEEPSEEK_API_KEY"`
}

func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	return &cfg, err
}

func (c *Config) DBSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		c.DBHost,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBPort,
		c.DBSSLMode,
	)
}
