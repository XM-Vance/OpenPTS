// RPA 监控仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── D5 RPA 监控 ───────────────

type RPAJob struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Schedule    *string    `json:"schedule,omitempty"`
	Enabled     bool       `json:"enabled"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	LastStatus  *string    `json:"last_status,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type RPARun struct {
	ID          string     `json:"id"`
	RpaJobID    string     `json:"rpa_job_id"`
	JobName     string     `json:"job_name,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	Status      string     `json:"status"`
	DurationSec *int       `json:"duration_sec,omitempty"`
	OutputFiles int        `json:"output_files"`
	OutputBytes int64      `json:"output_bytes"`
	Error       *string    `json:"error,omitempty"`
}

type RPARepository struct{ pool *Pool }

func NewRPARepository(pool *Pool) *RPARepository { return &RPARepository{pool: pool} }

func (r *RPARepository) ListJobs(ctx context.Context) ([]*RPAJob, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, schedule, enabled, last_run_at, last_status, created_at
		 FROM rpa_jobs ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*RPAJob, 0)
	for rows.Next() {
		var j RPAJob
		if err := rows.Scan(&j.ID, &j.Name, &j.Description, &j.Schedule,
			&j.Enabled, &j.LastRunAt, &j.LastStatus, &j.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &j)
	}
	return list, rows.Err()
}

func (r *RPARepository) ListRuns(ctx context.Context, limit int) ([]*RPARun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT r.id, r.rpa_job_id, j.name, r.started_at, r.finished_at, r.status,
			r.duration_sec, r.output_files, r.output_bytes, r.error
		 FROM rpa_runs r JOIN rpa_jobs j ON j.id = r.rpa_job_id
		 ORDER BY r.started_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*RPARun, 0, limit)
	for rows.Next() {
		var run RPARun
		if err := rows.Scan(&run.ID, &run.RpaJobID, &run.JobName, &run.StartedAt,
			&run.FinishedAt, &run.Status, &run.DurationSec, &run.OutputFiles,
			&run.OutputBytes, &run.Error); err != nil {
			return nil, err
		}
		list = append(list, &run)
	}
	return list, rows.Err()
}

func (r *RPARepository) GenerateDemo(ctx context.Context) (int, error) {
	jobs := []struct{ name, desc, sched string }{
		{"南网电费下载", "每日 02:00 抓取南方电网用户电费明细 CSV", "0 2 * * *"},
		{"现货出清结果同步", "每日 18:30 同步省间现货出清", "30 18 * * *"},
		{"调频补偿结果导入", "每日 09:00 从交易中心导入调频补偿", "0 9 * * *"},
		{"日前结算公告抓取", "每周一 10:00 抓取日前结算公告", "0 10 * * 1"},
	}
	statuses := []string{"success", "success", "success", "failed"}
	cnt := 0
	for _, jb := range jobs {
		var id string
		if err := r.pool.QueryRow(ctx,
			`INSERT INTO rpa_jobs (name, description, schedule, enabled, last_run_at, last_status)
			 VALUES ($1,$2,$3, true, now(), 'success')
			 ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description,
			   schedule = EXCLUDED.schedule, last_run_at = now(), last_status = 'success'
			 RETURNING id`, jb.name, jb.desc, jb.sched).Scan(&id); err != nil {
			return cnt, err
		}
		// 每个任务生成最近 5 次运行
		for i := 0; i < 5; i++ {
			start := time.Now().Add(-time.Duration(i*24) * time.Hour).Add(-time.Duration(rand.Intn(120)) * time.Minute)
			dur := 30 + rand.Intn(300)
			fin := start.Add(time.Duration(dur) * time.Second)
			status := statuses[rand.Intn(len(statuses))]
			var errMsg *string
			if status == "failed" {
				m := "连接超时"
				errMsg = &m
			}
			files := 0
			bytes := int64(0)
			if status == "success" {
				files = 1 + rand.Intn(8)
				bytes = int64(files) * int64(50000+rand.Intn(500000))
			}
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO rpa_runs (rpa_job_id, started_at, finished_at, status,
					duration_sec, output_files, output_bytes, error)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				id, start, fin, status, dur, files, bytes, errMsg); err != nil {
				return cnt, err
			}
		}
		cnt++
	}
	return cnt, nil
}
