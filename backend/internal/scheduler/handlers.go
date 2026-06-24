// 内置调度任务 handler。
package scheduler

import (
	"context"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// CleanupTokens 清理过期登录会话（auth_sessions.expires_at < now()）。
func CleanupTokens(ctx context.Context, pool *db.Pool) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM auth_sessions WHERE expires_at < now()`)
	if err != nil {
		return err
	}
	log.Info().Int64("rows", tag.RowsAffected()).Msg("清理过期会话")
	return nil
}

// AggregateDailyActive 汇总日活用户（占位实现：仅打日志）。
func AggregateDailyActive(ctx context.Context, pool *db.Pool) error {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM auth_sessions
		 WHERE created_at::date = $1::date`, yesterday).Scan(&n)
	if err != nil {
		return err
	}
	log.Info().Str("date", yesterday).Int("dau", n).Msg("汇总日活用户")
	return nil
}

// RefreshDashboardKPI 刷新仪表盘 KPI 缓存（当前为占位 — 仪表盘是即时 SQL，暂无缓存）。
func RefreshDashboardKPI(ctx context.Context, pool *db.Pool) error {
	// 真实场景会在此预聚合数据到 cache 表；目前仅做存活探针。
	var n int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM customers`).Scan(&n); err != nil {
		return err
	}
	log.Info().Int("customers", n).Msg("仪表盘 KPI 探针（占位）")
	return nil
}

// ExpireContracts 把 purchase_end_month < 当月的 active 合同自动转为 expired。
// 每天凌晨跑一次足够；同时对零售套餐做同样处理（如有 end_month 字段）。
func ExpireContracts(ctx context.Context, pool *db.Pool) error {
	currentMonth := time.Now().Format("2006-01")

	tag, err := pool.Exec(ctx, `
		UPDATE retail_contracts
		SET status = 'expired', updated_at = now()
		WHERE status = 'active'
		  AND purchase_end_month IS NOT NULL
		  AND purchase_end_month < $1
	`, currentMonth)
	if err != nil {
		return err
	}
	expired := tag.RowsAffected()

	log.Info().
		Int64("expired_contracts", expired).
		Str("current_month", currentMonth).
		Msg("合同到期归档")

	return nil
}
