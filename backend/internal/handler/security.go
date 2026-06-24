// 安全大屏 handler：基于 audit_logs + scheduled_jobs 聚合异常指标。
// 端点：GET /api/v1/system/security/overview?hours=24
package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type SecurityHandler struct {
	pool *db.Pool
}

func NewSecurityHandler(pool *db.Pool) *SecurityHandler {
	return &SecurityHandler{pool: pool}
}

type securityOverview struct {
	WindowHours      int                  `json:"window_hours"`
	Total            int                  `json:"total"`
	Errors4xx        int                  `json:"errors_4xx"`
	Errors5xx        int                  `json:"errors_5xx"`
	DeleteOps        int                  `json:"delete_ops"`
	UniqueUsers      int                  `json:"unique_users"`
	UniqueIPs        int                  `json:"unique_ips"`
	TopFailedUsers   []map[string]any     `json:"top_failed_users"`
	TopActiveIPs     []map[string]any     `json:"top_active_ips"`
	RecentDeletes    []map[string]any     `json:"recent_deletes"`
	FailedSchedJobs  int                  `json:"failed_sched_jobs"`
	ErrorHourly      []map[string]any     `json:"error_hourly"`
}

// Overview GET /api/v1/system/security/overview?hours=24
func (h *SecurityHandler) Overview(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours <= 0 || hours > 168 {
		hours = 24
	}
	ctx := c.Request.Context()
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	out := securityOverview{WindowHours: hours}

	// 总数 + 各类失败计数
	row := h.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status_code >= 400 AND status_code < 500),
			COUNT(*) FILTER (WHERE status_code >= 500),
			COUNT(*) FILTER (WHERE method = 'DELETE'),
			COUNT(DISTINCT user_id),
			COUNT(DISTINCT ip)
		FROM audit_logs WHERE created_at >= $1`, since)
	if err := row.Scan(&out.Total, &out.Errors4xx, &out.Errors5xx, &out.DeleteOps,
		&out.UniqueUsers, &out.UniqueIPs); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// Top 失败用户
	if rows, err := h.pool.Query(ctx, `
		SELECT COALESCE(username, '(匿名)') AS user, COUNT(*) AS fails
		FROM audit_logs
		WHERE created_at >= $1 AND status_code >= 400
		GROUP BY username ORDER BY fails DESC LIMIT 5`, since); err == nil {
		for rows.Next() {
			var u string
			var n int
			if err := rows.Scan(&u, &n); err == nil {
				out.TopFailedUsers = append(out.TopFailedUsers, map[string]any{"user": u, "count": n})
			}
		}
		rows.Close()
	}

	// Top 活跃 IP
	if rows, err := h.pool.Query(ctx, `
		SELECT ip::text, COUNT(*) AS reqs
		FROM audit_logs
		WHERE created_at >= $1 AND ip IS NOT NULL
		GROUP BY ip ORDER BY reqs DESC LIMIT 5`, since); err == nil {
		for rows.Next() {
			var ip string
			var n int
			if err := rows.Scan(&ip, &n); err == nil {
				out.TopActiveIPs = append(out.TopActiveIPs, map[string]any{"ip": ip, "count": n})
			}
		}
		rows.Close()
	}

	// 最近 DELETE 操作
	if rows, err := h.pool.Query(ctx, `
		SELECT username, path, resource, status_code, created_at
		FROM audit_logs
		WHERE method = 'DELETE' AND created_at >= $1
		ORDER BY created_at DESC LIMIT 20`, since); err == nil {
		for rows.Next() {
			var (
				uname *string
				path  string
				res   *string
				code  int
				ts    time.Time
			)
			if err := rows.Scan(&uname, &path, &res, &code, &ts); err == nil {
				out.RecentDeletes = append(out.RecentDeletes, map[string]any{
					"username":    uname,
					"path":        path,
					"resource":    res,
					"status_code": code,
					"created_at":  ts,
				})
			}
		}
		rows.Close()
	}

	// 失败的调度任务数（last_status=failed）
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM scheduled_jobs WHERE last_status = 'failed'`).Scan(&out.FailedSchedJobs); err != nil {
		slog.Warn("security overview: failed to count scheduled_jobs", "error", err)
	}

	// 每小时错误数（用于趋势图）
	if rows, err := h.pool.Query(ctx, `
		SELECT date_trunc('hour', created_at) AS bucket, COUNT(*) FILTER (WHERE status_code >= 400) AS errs
		FROM audit_logs WHERE created_at >= $1
		GROUP BY bucket ORDER BY bucket ASC`, since); err == nil {
		for rows.Next() {
			var bucket time.Time
			var n int
			if err := rows.Scan(&bucket, &n); err == nil {
				out.ErrorHourly = append(out.ErrorHourly, map[string]any{"bucket": bucket, "count": n})
			}
		}
		rows.Close()
	}

	c.JSON(http.StatusOK, out)
}
