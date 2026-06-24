// 健康检查与连通性测试 handler。
package handler

import (
	"net/http"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
)

// Health 返回服务运行状态 + 数据库连通性。
func Health(pool *db.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbStatus := "ok"
		status := http.StatusOK
		if err := pool.HealthCheck(c.Request.Context()); err != nil {
			dbStatus = "down: " + err.Error()
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{
			"status":   "ok",
			"service":  "ptis-backend",
			"time":     time.Now().Format(time.RFC3339),
			"database": dbStatus,
		})
	}
}

// Ping 用于联通性测试。
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
