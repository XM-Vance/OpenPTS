// 进程内调度器：基于 robfig/cron/v3，按 scheduled_jobs.cron_expr 触发已注册的 handler。
// handler 名通过 scheduled_jobs.handler 字段绑定。
//
// P0-D2 互斥：runOne 用 inflight sync.Map 保证同一 job 不并发执行。
// P0-D3 重试：按 scheduled_jobs.max_retries 失败重试（指数退避）。
// P0-D4 交易日历：trade_day_only=true 的 job 用 tradeDaySchedule 跳过周末+法定假日。
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	onFailure func(jobName, errMsg string) // 任务失败告警通道（IM 推送，未配置为 nil）
	mu        sync.Mutex
	entryByID map[string]cron.EntryID // jobID → cron.EntryID
	pub       EventPublisher
	inflight  sync.Map // jobID → struct{}：同一 job 执行互斥（P0-D2）
	holidays  *db.ForecastBaseRepository // 交易日历数据源（P0-D4）
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

// SetHolidaysRepo 注入节假日数据源，供 tradeDaySchedule 判断交易日（P0-D4）。
func (s *Scheduler) SetHolidaysRepo(r *db.ForecastBaseRepository) { s.holidays = r }

// SetFailureNotifier 注入任务失败告警通道（IM 机器人）。
func (s *Scheduler) SetFailureNotifier(fn func(jobName, errMsg string)) {
	s.onFailure = fn
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
	maxRetries := j.MaxRetries
	run := func() {
		s.runOne(jobID, jobName, fn, "cron", maxRetries)
	}
	// trade_day_only：用自定义 Schedule 跳过周末+法定假日（P0-D4）
	if j.TradeDayOnly {
		sched, err := newTradeDaySchedule(j.CronExpr, s.holidays)
		if err != nil {
			return fmt.Errorf("交易日历调度初始化失败：%w", err)
		}
		entryID := s.cron.Schedule(sched, cron.FuncJob(run))
		s.mu.Lock()
		s.entryByID[jobID] = entryID
		s.mu.Unlock()
		return nil
	}
	entryID, err := s.cron.AddFunc(j.CronExpr, run)
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
// P0-D2：inflight 互斥——同一 jobID 若已有实例在跑，直接跳过（防并发执行）。
// P0-D3：按 maxRetries 失败重试，指数退避（2s, 4s, 8s...）。
func (s *Scheduler) runOne(jobID, jobName string, fn JobFunc, trigger string, maxRetries int) {
	// 互斥：LoadOrStore 原子占位，已有实例在跑则跳过
	if _, running := s.inflight.LoadOrStore(jobID, struct{}{}); running {
		log.Warn().Str("job", jobName).Msg("任务已在执行中，跳过本次触发")
		return
	}
	defer s.inflight.Delete(jobID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	runID, err := s.repo.StartRun(ctx, jobID, trigger)
	if err != nil {
		log.Error().Err(err).Str("job", jobName).Msg("写 job_runs 失败")
		return
	}
	start := time.Now()
	// OTel span：让 scheduler job 在分布式追踪里可见（异步链路补齐，N10）。
	// 与 HTTP 入口 span（otelgin）和 algo 出站 span（otelhttp）串联成完整请求链。
	tracer := otel.Tracer("ptis-backend/scheduler")
	spanCtx, span := tracer.Start(ctx, "scheduler.job."+jobName,
		trace.WithAttributes(attribute.String("job.name", jobName), attribute.Int("job.max_retries", maxRetries)))
	defer span.End()

	// 执行 + 重试：首次 + maxRetries 次，指数退避
	var jobErr error
	attempts := maxRetries + 1
	for i := 0; i < attempts; i++ {
		jobErr = fn(spanCtx, s.pool)
		if jobErr == nil || spanCtx.Err() != nil {
			break // 成功 或 上下文已取消（如停机），不再重试
		}
		if i < attempts-1 {
			backoff := time.Duration(1<<(i+1)) * time.Second // 2s, 4s, 8s...
			log.Warn().Err(jobErr).Str("job", jobName).Int("attempt", i+1).
				Dur("backoff", backoff).Msg("任务失败，准备重试")
			select {
			case <-time.After(backoff):
			case <-spanCtx.Done():
			}
			if spanCtx.Err() != nil {
				break // 退避期间被取消，停止重试
			}
		}
	}
	// span 记录执行结果（成功/失败 + 耗时），便于 Tempo 里按 status 筛选
	if jobErr != nil {
		span.SetStatus(codes.Error, jobErr.Error())
	} else {
		span.SetAttributes(attribute.String("job.status", "success"))
	}
	span.SetAttributes(attribute.Int("job.duration_ms", int(time.Since(start)/time.Millisecond)))
	dur := int(time.Since(start) / time.Millisecond)
	status := "success"
	var errStr *string
	if jobErr != nil {
		status = "failed"
		e := jobErr.Error()
		errStr = &e
		log.Error().Err(jobErr).Str("job", jobName).Int("ms", dur).
			Int("attempts", attempts).Msg("任务执行失败（已耗尽重试")
		if s.onFailure != nil { // 失败告警推送（采集断流等需即时可见）
			go s.onFailure(jobName, e)
		}
	} else {
		log.Info().Str("job", jobName).Int("ms", dur).Msg("任务执行成功")
	}
	// Prometheus 指标：让失败可见，Grafana 可对 failed rate 告警。
	jobRunsTotal.WithLabelValues(jobName, status).Inc()
	jobDurationSeconds.WithLabelValues(jobName).Observe(time.Since(start).Seconds())
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

// TriggerByID 手工触发一次（忽略 trade_day_only，手动触发总是执行）。
func (s *Scheduler) TriggerByID(ctx context.Context, jobID string) error {
	j, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	fn, ok := s.handlers[j.Handler]
	if !ok {
		return fmt.Errorf("未注册 handler: %s", j.Handler)
	}
	go s.runOne(j.ID, j.Name, fn, "manual", j.MaxRetries)
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
