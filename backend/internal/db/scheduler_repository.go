// 调度任务仓储：scheduled_jobs CRUD + job_runs 写入/查询。
package db

import (
	"context"
	"time"
)

type ScheduledJob struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CronExpr    string     `json:"cron_expr"`
	Handler     string     `json:"handler"`
	Enabled     bool       `json:"enabled"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	LastStatus  *string    `json:"last_status,omitempty"`
	LastError   *string    `json:"last_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type JobRun struct {
	ID         string     `json:"id"`
	JobID      string     `json:"job_id"`
	JobName    string     `json:"job_name,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Status     string     `json:"status"`
	Error      *string    `json:"error,omitempty"`
	DurationMs *int       `json:"duration_ms,omitempty"`
	Trigger    string     `json:"trigger"`
}

type SchedulerRepository struct {
	pool *Pool
}

func NewSchedulerRepository(pool *Pool) *SchedulerRepository {
	return &SchedulerRepository{pool: pool}
}

func (r *SchedulerRepository) ListJobs(ctx context.Context) ([]*ScheduledJob, error) {
	const q = `SELECT id, name, description, cron_expr, handler, enabled,
		last_run_at, last_status, last_error, created_at, updated_at
		FROM scheduled_jobs ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ScheduledJob, 0)
	for rows.Next() {
		var j ScheduledJob
		if err := rows.Scan(&j.ID, &j.Name, &j.Description, &j.CronExpr, &j.Handler,
			&j.Enabled, &j.LastRunAt, &j.LastStatus, &j.LastError,
			&j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &j)
	}
	return list, rows.Err()
}

func (r *SchedulerRepository) GetByName(ctx context.Context, name string) (*ScheduledJob, error) {
	const q = `SELECT id, name, description, cron_expr, handler, enabled,
		last_run_at, last_status, last_error, created_at, updated_at
		FROM scheduled_jobs WHERE name = $1`
	var j ScheduledJob
	if err := r.pool.QueryRow(ctx, q, name).Scan(&j.ID, &j.Name, &j.Description,
		&j.CronExpr, &j.Handler, &j.Enabled, &j.LastRunAt, &j.LastStatus,
		&j.LastError, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *SchedulerRepository) GetByID(ctx context.Context, id string) (*ScheduledJob, error) {
	const q = `SELECT id, name, description, cron_expr, handler, enabled,
		last_run_at, last_status, last_error, created_at, updated_at
		FROM scheduled_jobs WHERE id = $1`
	var j ScheduledJob
	if err := r.pool.QueryRow(ctx, q, id).Scan(&j.ID, &j.Name, &j.Description,
		&j.CronExpr, &j.Handler, &j.Enabled, &j.LastRunAt, &j.LastStatus,
		&j.LastError, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *SchedulerRepository) SetEnabled(ctx context.Context, id string, enabled bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE scheduled_jobs SET enabled = $1, updated_at = now()
		WHERE id = $2`, enabled, id)
	return err
}

// StartRun 写一条 running 记录，返回 runID。
func (r *SchedulerRepository) StartRun(ctx context.Context, jobID, trigger string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO job_runs (job_id, status, trigger) VALUES ($1, 'running', $2) RETURNING id`,
		jobID, trigger).Scan(&id)
	return id, err
}

// FinishRun 标记完成，同时更新 scheduled_jobs.last_*。
func (r *SchedulerRepository) FinishRun(ctx context.Context, runID, jobID, status string, errStr *string, durationMs int) error {
	if _, err := r.pool.Exec(ctx,
		`UPDATE job_runs SET finished_at = now(), status = $1, error = $2, duration_ms = $3
		 WHERE id = $4`, status, errStr, durationMs, runID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE scheduled_jobs SET last_run_at = now(), last_status = $1, last_error = $2, updated_at = now()
		 WHERE id = $3`, status, errStr, jobID)
	return err
}

func (r *SchedulerRepository) ListRuns(ctx context.Context, limit int) ([]*JobRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `SELECT r.id, r.job_id, j.name, r.started_at, r.finished_at,
			r.status, r.error, r.duration_ms, r.trigger
		FROM job_runs r JOIN scheduled_jobs j ON j.id = r.job_id
		ORDER BY r.started_at DESC LIMIT $1`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*JobRun, 0, limit)
	for rows.Next() {
		var jr JobRun
		if err := rows.Scan(&jr.ID, &jr.JobID, &jr.JobName, &jr.StartedAt,
			&jr.FinishedAt, &jr.Status, &jr.Error, &jr.DurationMs, &jr.Trigger); err != nil {
			return nil, err
		}
		list = append(list, &jr)
	}
	return list, rows.Err()
}
