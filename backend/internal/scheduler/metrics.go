// Scheduler Prometheus 指标：任务执行次数（按 job/status）+ 耗时直方图。
// 在 scheduler.go runOne 的成功/失败分支打点；端点 /metrics 暴露（无需鉴权）。
package scheduler

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// jobRunsTotal 维度：job（任务名）+ status（success/failed）。
	// 历史：scheduler 任务失败曾「只记日志、不告警」（见 docs/improvement-roadmap-2026-07.md H-NEW-2），
	// 加 metric 后 Grafana 可对 rate(ptis_scheduler_job_runs_total{status="failed"}[5m]) > 0 告警。
	jobRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ptis_scheduler_job_runs_total",
			Help: "调度任务执行次数（按 job 名 + 终态 success/failed）",
		},
		[]string{"job", "status"},
	)

	// jobDurationSeconds 维度：job。bucket 覆盖秒级到 5 分钟（context 超时上限 5min）。
	jobDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ptis_scheduler_job_duration_seconds",
			Help:    "调度任务执行耗时直方图（秒）",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"job"},
	)
)

func init() {
	prometheus.MustRegister(jobRunsTotal, jobDurationSeconds)
}
