// 配置加载：从环境变量 + .env 读取并校验。
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Environment         string // dev / prod
	LogLevel            string
	DatabaseURL         string
	DatabaseReplicaURL  string // 只读副本（可选）
	DoclingServiceURL   string // 文档解析服务（docling）
	JWTSecret           string
	JWTTTL              time.Duration
}

// Load 解析环境变量。.env 不存在时不报错（容器场景由编排注入）。
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", getEnv("ENV", "development")),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DatabaseReplicaURL: getEnv("DATABASE_REPLICA_URL", ""),
		DoclingServiceURL: getEnv("DOCLING_SERVICE_URL", "http://localhost:8300"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
	}

	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "8"))
	if err != nil {
		return nil, fmt.Errorf("JWT_TTL_HOURS 非整数: %w", err)
	}
	cfg.JWTTTL = time.Duration(ttlHours) * time.Hour

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("缺少 DATABASE_URL")
	}
	if c.JWTSecret == "" || c.JWTSecret == "change_me_in_production" {
		if c.IsProd() {
			return errors.New("生产环境必须设置 JWT_SECRET")
		}
		// dev 模式给一个不安全但能跑的默认值
		c.JWTSecret = "dev-only-insecure-secret-do-not-use-in-prod"
	}
	return nil
}

func (c *Config) IsProd() bool {
	return c.Environment == "prod" || c.Environment == "production"
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
