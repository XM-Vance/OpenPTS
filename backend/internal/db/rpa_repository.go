// RPA 监控仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移。
// 2026-08（0132 迁移）加 org_id 多租户隔离：读走 OrgFilter，写走 MustScoped。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
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
	q := `SELECT id, name, description, schedule, enabled, last_run_at, last_status, created_at
	      FROM rpa_jobs`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY name ASC"
	rows, err := r.pool.Query(ctx, q, args...)
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
	q := `SELECT r.id, r.rpa_job_id, j.name, r.started_at, r.finished_at, r.status,
		      r.duration_sec, r.output_files, r.output_bytes, r.error
	       FROM rpa_runs r JOIN rpa_jobs j ON j.id = r.rpa_job_id`
	args := []any{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE j.org_id = $%d::uuid", len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY r.started_at DESC LIMIT $%d", len(args))
	rows, err := r.pool.Query(ctx, q, args...)
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
	org, err := MustScoped(ctx)
	if err != nil {
		return 0, err
	}
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
			`INSERT INTO rpa_jobs (org_id, name, description, schedule, enabled, last_run_at, last_status)
			 VALUES ($1,$2,$3,$4, true, now(), 'success')
			 ON CONFLICT (org_id, name) DO UPDATE SET description = EXCLUDED.description,
			   schedule = EXCLUDED.schedule, last_run_at = now(), last_status = 'success'
			 RETURNING id`, org, jb.name, jb.desc, jb.sched).Scan(&id); err != nil {
			return cnt, err
		}
		// 每个任务生成最近 5 次运行
		for i := 0; i < 5; i++ {
			start := time.Now().Add(-time.Duration(i*24) * time.Hour).Add(-time.Duration(rand.IntN(120)) * time.Minute)
			dur := 30 + rand.IntN(300)
			fin := start.Add(time.Duration(dur) * time.Second)
			status := statuses[rand.IntN(len(statuses))]
			var errMsg *string
			if status == "failed" {
				m := "连接超时"
				errMsg = &m
			}
			files := 0
			bytes := int64(0)
			if status == "success" {
				files = 1 + rand.IntN(8)
				bytes = int64(files) * int64(50000+rand.IntN(500000))
			}
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO rpa_runs (org_id, rpa_job_id, started_at, finished_at, status,
					duration_sec, output_files, output_bytes, error)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
				org, id, start, fin, status, dur, files, bytes, errMsg); err != nil {
				return cnt, err
			}
		}
		cnt++
	}
	return cnt, nil
}

// ─── WP5.3 后续 / WP3.5：RPA 真实读写 ───

// RPAJobInfo job 详情（含动作类型）。
type RPAJobInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Schedule    string     `json:"schedule"`
	Action      string     `json:"action"` // declaration_export / external
	Enabled     bool       `json:"enabled"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	LastStatus  string     `json:"last_status,omitempty"`
}

// UpsertJob 注册/更新 job（name 唯一；action 记录动作类型）。
func (r *RPARepository) UpsertJob(ctx context.Context, name, description, schedule, action string) (*RPAJobInfo, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	q := `INSERT INTO rpa_jobs (org_id, name, description, schedule, action)
		VALUES ($5::uuid,$1,$2,$3,$4)
		ON CONFLICT (org_id, name) DO UPDATE SET
		  description=EXCLUDED.description, schedule=EXCLUDED.schedule, action=EXCLUDED.action
		RETURNING id::text, name, COALESCE(description,''), COALESCE(schedule,''),
		  COALESCE(action,'external'), enabled, last_run_at, COALESCE(last_status,'')`
	var j RPAJobInfo
	err = r.pool.QueryRow(ctx, q, name, description, schedule, action, org).Scan(
		&j.ID, &j.Name, &j.Description, &j.Schedule, &j.Action, &j.Enabled, &j.LastRunAt, &j.LastStatus)
	return &j, err
}

// GetJobByName 按 name 取 job。
func (r *RPARepository) GetJobByName(ctx context.Context, name string) (*RPAJobInfo, error) {
	q := `SELECT id::text, name, COALESCE(description,''), COALESCE(schedule,''),
		  COALESCE(action,'external'), enabled, last_run_at, COALESCE(last_status,'')
	  FROM rpa_jobs WHERE name = $1`
	var j RPAJobInfo
	err := r.pool.QueryRow(ctx, q, name).Scan(
		&j.ID, &j.Name, &j.Description, &j.Schedule, &j.Action, &j.Enabled, &j.LastRunAt, &j.LastStatus)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// StartRun 创建 running run 并刷新 job last_*（返回 run id）。
func (r *RPARepository) StartRun(ctx context.Context, jobID string) (string, error) {
	var runID string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO rpa_runs (rpa_job_id, status) VALUES ($1::uuid, 'running')
		RETURNING id::text`, jobID).Scan(&runID)
	if err != nil {
		return "", err
	}
	_, _ = r.pool.Exec(ctx, `
		UPDATE rpa_jobs SET last_run_at = now(), last_status = 'running' WHERE id = $1::uuid`, jobID)
	return runID, nil
}

// FinishRun 收口 run（status success/failed）并回写 job last_status。
func (r *RPARepository) FinishRun(ctx context.Context, runID, status string, files, bytes int64, errMsg string) error {
	q := `UPDATE rpa_runs SET status=$2, finished_at=now(),
		duration_sec = EXTRACT(EPOCH FROM (now() - started_at))::int,
		output_files=$3, output_bytes=$4, error=$5
	  WHERE id=$1::uuid AND status='running'
	  RETURNING rpa_job_id::text`
	var jobID string
	err := r.pool.QueryRow(ctx, q, runID, status, files, bytes, nullStr(errMsg)).Scan(&jobID)
	if err != nil {
		return err
	}
	_, _ = r.pool.Exec(ctx, `UPDATE rpa_jobs SET last_status=$2 WHERE id=$1::uuid`, jobID, status)
	return nil
}
