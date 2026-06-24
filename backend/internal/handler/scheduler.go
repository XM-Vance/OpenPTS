// 调度任务 handler：查看任务、查看执行记录、启停、立即触发。
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type SchedulerHandler struct {
	repo *db.SchedulerRepository
	sch  *scheduler.Scheduler
}

func NewSchedulerHandler(repo *db.SchedulerRepository, sch *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{repo: repo, sch: sch}
}

type jobWithNext struct {
	*db.ScheduledJob
	NextRunAt *time.Time `json:"next_run_at,omitempty"`
}

// ListJobs GET /api/v1/scheduler/jobs
func (h *SchedulerHandler) ListJobs(c *gin.Context) {
	jobs, err := h.repo.ListJobs(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("查询失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	out := make([]jobWithNext, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, jobWithNext{ScheduledJob: j, NextRunAt: h.sch.NextRun(j.ID)})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

// ListRuns GET /api/v1/scheduler/runs?limit=50
func (h *SchedulerHandler) ListRuns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	runs, err := h.repo.ListRuns(c.Request.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("查询失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": runs})
}

// Trigger POST /api/v1/scheduler/jobs/:id/trigger
func (h *SchedulerHandler) Trigger(c *gin.Context) {
	id := c.Param("id")
	if err := h.sch.TriggerByID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已触发，请稍后查看执行记录"})
}

type setEnabledReq struct {
	Enabled bool `json:"enabled"`
}

// SetEnabled PUT /api/v1/scheduler/jobs/:id/enabled
func (h *SchedulerHandler) SetEnabled(c *gin.Context) {
	id := c.Param("id")
	var req setEnabledReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.sch.SetEnabled(c.Request.Context(), id, req.Enabled); err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
