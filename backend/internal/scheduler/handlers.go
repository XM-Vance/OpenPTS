// 内置调度任务 handler。
package scheduler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// resolveScriptPath 解析外部脚本路径：默认值 + env 覆盖 + 白名单目录校验。
//
// 历史问题（docs/improvement-roadmap-2026-07.md H-NEW-2）：
//   - 头注释声称支持 env 覆盖（MARKET_DATA_SCRIPT / WEATHER_SCRIPT），代码实际没读；
//   - 无白名单 → 若 env 被注入任意路径，可执行 scripts/ 目录之外的脚本。
//
// 此函数统一两个外部 exec 点的路径解析：env 真读取 + 必须落在白名单根目录之下
// （防 `..` 越权）。允许的根由 env PTIS_SCRIPTS_DIR 配置，缺省 scripts/data-collection。
//
// 注意：本函数只校验路径合规，不解决容器内是否真有 python3 / 脚本——后者由 Dockerfile
// 打包（见 PR 同 commit 的 backend/Dockerfile 改动）。
func resolveScriptPath(envKey, defaultName string) (string, error) {
	root := os.Getenv("PTIS_SCRIPTS_DIR")
	if root == "" {
		root = "scripts/data-collection"
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("解析脚本根目录失败: %w", err)
	}

	name := defaultName
	if v := os.Getenv(envKey); v != "" {
		name = v
	}

	// 仅取文件名部分（env 注入绝对路径或含目录分隔符一律拒绝），
	// 强制脚本必须位于白名单根之下，杜绝路径穿越。
	base := filepath.Base(name)
	if base == "." || base == string(filepath.Separator) || base == name && name != defaultName && filepath.ToSlash(name) != base {
		return "", fmt.Errorf("env %s 必须是纯文件名（不含路径分隔符），当前: %q", envKey, name)
	}

	full := filepath.Join(rootAbs, base)
	// 二次校验：Clean 后仍需在 rootAbs 之下（防 base 含 `..` 的极端构造）
	rel, err := filepath.Rel(rootAbs, full)
	if err != nil || rel == "" || rel == ".." || len(rel) >= 3 && rel[:3] == "../" {
		return "", fmt.Errorf("脚本路径 %q 不在允许根目录 %q 之下", full, rootAbs)
	}
	return full, nil
}

// runPythonScript 公共执行体：解析路径 → exec python3 → 捕获输出 → 失败带结构化日志。
// 失败时返回 error，由 scheduler runOne 重试 + 打 metric + 落 job_runs。
func runPythonScript(ctx context.Context, envKey, defaultName, jobLabel string) error {
	// PTIS_PY_COLLECT=0：Docker 部署下 backend 容器无 python3，采集由 ptis-fetcher
	// sidecar 承担（docker-compose 已接线），本任务按配置跳过（记 success，不打 failed 指标）。
	if os.Getenv("PTIS_PY_COLLECT") == "0" {
		log.Info().Str("job", jobLabel).
			Msg(jobLabel + "采集由 fetcher sidecar 承担（PTIS_PY_COLLECT=0），本次跳过")
		return nil
	}
	scriptPath, err := resolveScriptPath(envKey, defaultName)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "python3", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Error().
			Err(err).
			Str("job", jobLabel).
			Str("script", scriptPath).
			Str("stdout", stdout.String()).
			Str("stderr", stderr.String()).
			Msg(jobLabel + "采集失败")
		// 区分「脚本不存在 / python3 不在 PATH」与脚本执行失败：
		// 前者属环境配置问题（容器未打包），重试无意义但需暴露。
		var pErr *exec.Error
		if errors.As(err, &pErr) {
			return fmt.Errorf("%s: 命令不可用 %q（检查镜像是否打包 python3 + 脚本）: %w", jobLabel, pErr.Name, err)
		}
		return fmt.Errorf("%s采集失败: %w", jobLabel, err)
	}
	log.Info().Str("job", jobLabel).Str("script", scriptPath).
		Str("stdout", stdout.String()).
		Msg(jobLabel + "采集完成")
	return nil
}

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

// AggregateDailyActive 汇总日活用户并落库到 dau_daily（供趋势查询）。
// 数据源：auth_sessions（昨日去重 user_id）。UPSERT 到 dau_daily。
func AggregateDailyActive(ctx context.Context, pool *db.Pool) error {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM auth_sessions
		 WHERE created_at::date = $1::date`, yesterday).Scan(&n)
	if err != nil {
		return err
	}
	// 落库（UPSERT：重算同一日则覆盖）
	if _, err := pool.Exec(ctx,
		`INSERT INTO dau_daily (date, user_count, computed_at)
		 VALUES ($1::date, $2, now())
		 ON CONFLICT (date) DO UPDATE SET user_count = EXCLUDED.user_count, computed_at = now()`,
		yesterday, n); err != nil {
		return err
	}
	log.Info().Str("date", yesterday).Int("dau", n).
		Msg("汇总日活用户（已落库 dau_daily）")
	return nil
}

// RefreshDashboardKPI 仪表盘存活探针（非 KPI 预聚合）。
// 仪表盘走即时 SQL + 内存 TTL 缓存（dashboard.go），无需预聚合到 cache 表；
// 本任务仅周期性 ping DB 确认调度链路 + 数据库连通，不产生业务副作用。
// 若数据量达到千万级慢查询阈值，再考虑建预聚合表。
func RefreshDashboardKPI(ctx context.Context, pool *db.Pool) error {
	var n int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM customers WHERE lifecycle_stage NOT IN ('intent','lead')`).Scan(&n); err != nil {
		return err
	}
	log.Info().Int("customers", n).
		Msg("仪表盘存活探针（DB 连通正常）")
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

// FetchMarketData 调用外部 AKShare 采集脚本，刷新 30 类市场行情数据（md_* 表）。
// 脚本文件名默认 fetch_market_data.py，位于 PTIS_SCRIPTS_DIR（缺省 scripts/data-collection），
// 可用环境变量 MARKET_DATA_SCRIPT 覆盖（仅文件名，路径穿越会被拒）。
// 采集失败：metric + 日志 + job_runs 三路可见，由 scheduler runOne 重试。
func FetchMarketData(ctx context.Context, pool *db.Pool) error {
	_ = pool // 脚本自行连库，保留入参以满足 JobFunc 签名。
	return runPythonScript(ctx, "MARKET_DATA_SCRIPT", "fetch_market_data.py", "市场行情")
}

// FetchWeatherData 调用外部 Open-Meteo 采集脚本，刷新气象原始观测数据
// （md_weather_wind_hourly 风电场逐时 + md_weather_hydrology_daily 水库水文逐日）。
// 时序上排在 fetch_weather_actuals(19:00) 之前，保证 ETL 上游数据已就绪。
// 路径解析同 FetchMarketData（WEATHER_SCRIPT 覆盖 + 白名单根）。
func FetchWeatherData(ctx context.Context, pool *db.Pool) error {
	_ = pool
	return runPythonScript(ctx, "WEATHER_SCRIPT", "fetch_weather_data.py", "气象数据")
}

// FetchWeatherActuals 把 md_weather_hydrology_daily（Open-Meteo 脚本已采）聚合到 weather_actuals。
// 接通"死表"weather_actuals（此前唯一写入是 demo，下游 ActualsSummary 永远空）。
// 步骤：
//  1. 兜底补建 weather_locations 站点（防 0114 迁移时 md 表为空；同名站点取均值经纬度）
//  2. ETL 最近 7 天的 md_weather_hydrology_daily → weather_actuals（UPSERT）
//
// min_temp/max_temp 由 0125 迁移新增的 temp_max/temp_min 列透传（Open-Meteo daily 本就返回极值）。
func FetchWeatherActuals(ctx context.Context, pool *db.Pool) error {
	// 1. 兜底补建站点
	if _, err := pool.Exec(ctx, `
		INSERT INTO weather_locations (name, latitude, longitude)
		SELECT h.location_name, AVG(h.lat), AVG(h.lon)
		FROM md_weather_hydrology_daily h
		WHERE h.location_name IS NOT NULL
		  AND NOT EXISTS (SELECT 1 FROM weather_locations w WHERE w.name = h.location_name)
		GROUP BY h.location_name
		ON CONFLICT (name) DO NOTHING`); err != nil {
		return err
	}

	// 2. ETL 最近 7 天 md_weather → weather_actuals
	tag, err := pool.Exec(ctx, `
		INSERT INTO weather_actuals (location_name, date, avg_temp, max_temp, min_temp, humidity, wind_speed, precipitation)
		SELECT h.location_name, h.obs_date, h.temp_mean, h.temp_max, h.temp_min,
		       h.humidity_mean, h.wind_speed_10m_mean, h.precipitation_sum
		FROM md_weather_hydrology_daily h
		JOIN weather_locations w ON w.name = h.location_name
		WHERE h.obs_date >= now() - interval '7 days'
		ON CONFLICT (location_name, date) DO UPDATE SET
		    avg_temp       = EXCLUDED.avg_temp,
		    max_temp       = EXCLUDED.max_temp,
		    min_temp       = EXCLUDED.min_temp,
		    humidity       = EXCLUDED.humidity,
		    wind_speed     = EXCLUDED.wind_speed,
		    precipitation  = EXCLUDED.precipitation`)
	if err != nil {
		return err
	}
	log.Info().Int64("rows", tag.RowsAffected()).
		Msg("weather_actuals 已刷新（从 md_weather_hydrology_daily 聚合）")
	return nil
}
