// 进程内调度器：基于 robfig/cron/v3，按 scheduled_jobs.cron_expr 触发已注册的 handler。
// handler 名通过 scheduled_jobs.handler 字段绑定。
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// JobFunc 一个调度任务的执行体。返回 error 写入 job_runs.error。
type JobFunc func(ctx context.Context, pool *db.Pool) error

// EventPublisher 由外部注入（如 handler.SSEHub），用于推送任务事件。
type EventPublisher interface {
	PublishJobEvent(jobName, status string, durationMs int, errMsg string)
}

// Scheduler 持有 cron + 已注册 handler。
type Scheduler struct {
	cron      *cron.Cron
	handlers  map[string]JobFunc
	repo      *db.SchedulerRepository
	pool      *db.Pool
	mu        sync.Mutex
	entryByID map[string]cron.EntryID // jobID → cron.EntryID
	pub       EventPublisher
}

func New(repo *db.SchedulerRepository, pool *db.Pool) *Scheduler {
	// 使用 6 段制（带秒）的 cron 解析器。
	c := cron.New(cron.WithSeconds(), cron.WithLogger(cron.DiscardLogger))
	return &Scheduler{
		cron:      c,
		handlers:  map[string]JobFunc{},
		repo:      repo,
		pool:      pool,
		entryByID: map[string]cron.EntryID{},
	}
}

// SetPublisher 注入事件发布器（main.go 在 New 后调用）。
func (s *Scheduler) SetPublisher(p EventPublisher) { s.pub = p }

// Register 注册 handler。在 Start 之前调用。
func (s *Scheduler) Register(name string, fn JobFunc) {
	s.handlers[name] = fn
}

// Start 从数据库加载启用中的任务并启动 cron。
func (s *Scheduler) Start(ctx context.Context) error {
	jobs, err := s.repo.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("加载调度任务失败：%w", err)
	}
	for _, j := range jobs {
		if !j.Enabled {
			continue
		}
		if err := s.addEntry(j); err != nil {
			log.Error().Err(err).Str("job", j.Name).Msg("注册调度任务失败")
			continue
		}
	}
	s.cron.Start()
	log.Info().Int("count", len(s.entryByID)).Msg("调度器已启动")
	return nil
}

// Stop 停止 cron。返回的 context 在所有运行中任务结束后关闭。
func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}

func (s *Scheduler) addEntry(j *db.ScheduledJob) error {
	fn, ok := s.handlers[j.Handler]
	if !ok {
		return fmt.Errorf("未注册 handler: %s", j.Handler)
	}
	jobID := j.ID
	jobName := j.Name
	entryID, err := s.cron.AddFunc(j.CronExpr, func() {
		s.runOne(jobID, jobName, fn, "cron")
	})
	if err != nil {
		return fmt.Errorf("cron 表达式解析失败：%w", err)
	}
	s.mu.Lock()
	s.entryByID[jobID] = entryID
	s.mu.Unlock()
	return nil
}

func (s *Scheduler) removeEntry(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, ok := s.entryByID[jobID]; ok {
		s.cron.Remove(entryID)
		delete(s.entryByID, jobID)
	}
}

// runOne 执行一次任务，写 job_runs + 更新 scheduled_jobs.last_*。
func (s *Scheduler) runOne(jobID, jobName string, fn JobFunc, trigger string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	runID, err := s.repo.StartRun(ctx, jobID, trigger)
	if err != nil {
		log.Error().Err(err).Str("job", jobName).Msg("写 job_runs 失败")
		return
	}
	start := time.Now()
	jobErr := fn(ctx, s.pool)
	dur := int(time.Since(start) / time.Millisecond)
	status := "success"
	var errStr *string
	if jobErr != nil {
		status = "failed"
		s := jobErr.Error()
		errStr = &s
		log.Error().Err(jobErr).Str("job", jobName).Int("ms", dur).Msg("任务执行失败")
	} else {
		log.Info().Str("job", jobName).Int("ms", dur).Msg("任务执行成功")
	}
	if err := s.repo.FinishRun(ctx, runID, jobID, status, errStr, dur); err != nil {
		log.Error().Err(err).Msg("写 job_runs 完成失败")
	}
	// 广播 SSE 事件（如有）
	if s.pub != nil {
		em := ""
		if errStr != nil {
			em = *errStr
		}
		s.pub.PublishJobEvent(jobName, status, dur, em)
	}
}

// TriggerByID 手工触发一次。
func (s *Scheduler) TriggerByID(ctx context.Context, jobID string) error {
	j, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	fn, ok := s.handlers[j.Handler]
	if !ok {
		return fmt.Errorf("未注册 handler: %s", j.Handler)
	}
	go s.runOne(j.ID, j.Name, fn, "manual")
	return nil
}

// SetEnabled 启用/禁用任务，并同步 cron 注册状态。
func (s *Scheduler) SetEnabled(ctx context.Context, jobID string, enabled bool) error {
	if err := s.repo.SetEnabled(ctx, jobID, enabled); err != nil {
		return err
	}
	if enabled {
		j, err := s.repo.GetByID(ctx, jobID)
		if err != nil {
			return err
		}
		return s.addEntry(j)
	}
	s.removeEntry(jobID)
	return nil
}

// NextRun 返回下次执行时间（未注册则返回 nil）。
func (s *Scheduler) NextRun(jobID string) *time.Time {
	s.mu.Lock()
	entryID, ok := s.entryByID[jobID]
	s.mu.Unlock()
	if !ok {
		return nil
	}
	e := s.cron.Entry(entryID)
	if e.ID == 0 {
		return nil
	}
	t := e.Next
	return &t
}
