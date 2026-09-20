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
	DemoMode            bool // 演示模式：预测端点返回合成数据，便于开箱体验
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
		DemoMode:       getEnv("DEMO_MODE", "") == "true",
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
	if isKnownInsecureJWTPlaceholder(c.JWTSecret) {
		if c.IsProd() {
			return errors.New("生产环境必须设置 JWT_SECRET（不能用占位值）")
		}
		// dev 模式给一个不安全但能跑的默认值
		c.JWTSecret = "dev-only-insecure-secret-do-not-use-in-prod"
	}
	// 生产环境长度校验：HS256 安全性依赖密钥熵，<32 字符（256 bit）不符合 RFC 8729 建议。
	// 仅在显式设置（非空、非占位）且 IsProd 时强制；dev 模式仅记日志，避免破坏本地工作流。
	if c.IsProd() && c.JWTSecret != "" && !isKnownInsecureJWTPlaceholder(c.JWTSecret) && len(c.JWTSecret) < 32 {
		return fmt.Errorf("生产环境 JWT_SECRET 长度需 >= 32 字符（当前 %d），请用 `openssl rand -base64 48` 生成", len(c.JWTSecret))
	}
	return nil
}

func isKnownInsecureJWTPlaceholder(s string) bool {
	switch s {
	case "", "change_me_in_production":
		return true
	}
	// .env.example 历史占位（已被截断的不完整串）+ 明显的未替换标记
	return s == "please...hars" || s == "CHANGE_ME_RUN_openssl_rand_base64_48"
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
