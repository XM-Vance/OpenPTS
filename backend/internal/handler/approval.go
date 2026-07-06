// 通用审批流 handler：发起 / 列表 / 审批 / 撤回。
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ptis/backend/internal/approval"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ApprovalHandler struct {
	repo     *db.ApprovalRepository
	registry *approval.Registry
	hub      *SSEHub
}

func NewApprovalHandler(repo *db.ApprovalRepository, reg *approval.Registry, hub *SSEHub) *ApprovalHandler {
	return &ApprovalHandler{repo: repo, registry: reg, hub: hub}
}

// notify 推送审批事件。按当前活跃省（orgID）定向广播——A 省审批不推给 B 省。
// 总部（isHQ）订阅者仍会收到（全局视角）。orgID 由调用方从请求上下文取。
func (h *ApprovalHandler) notify(a *db.Approval, action string, orgID string) {
	if h.hub == nil {
		return
	}
	h.hub.PublishToOrg(orgID, SSEEvent{
		Type: "approval",
		Data: map[string]any{
			"action":      action,            // submitted / approved / rejected / withdrawn
			"id":          a.ID,
			"resource":    a.Resource,
			"resource_id": a.ResourceID,
			"title":       a.Title,
			"status":      a.Status,
			"submitted_by": a.SubmittedBy,
			"reviewed_by": a.ReviewedBy,
			"message":     buildApprovalMsg(a, action),
		},
	})
}

func buildApprovalMsg(a *db.Approval, action string) string {
	displayName := a.SubmittedBy
	if a.SubmittedByName != nil && *a.SubmittedByName != "" {
		displayName = *a.SubmittedByName
	}
	switch action {
	case "submitted":
		return "新审批：" + a.Title + "（提交人 " + displayName + "）"
	case "approved":
		return "已通过：" + a.Title
	case "rejected":
		return "已驳回：" + a.Title
	case "withdrawn":
		return "已撤回：" + a.Title
	}
	return a.Title
}

type submitReq struct {
	Resource   string          `json:"resource" binding:"required"`
	ResourceID string          `json:"resource_id" binding:"required"`
	Title      string          `json:"title" binding:"required"`
	Payload    json.RawMessage `json:"payload"`
}

// Submit POST /api/v1/approvals
func (h *ApprovalHandler) Submit(c *gin.Context) {
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	uid := claimsUserID(c)
	var submittedBy string
	if uid != nil {
		submittedBy = uid.String()
	}
	a, err := h.repo.Create(c.Request.Context(), db.ApprovalInput{
		Resource:    req.Resource,
		ResourceID:  req.ResourceID,
		Title:       req.Title,
		Payload:     req.Payload,
		SubmittedBy: submittedBy,
	})
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	orgNotify, _ := db.OrgFilter(c.Request.Context()) // 总部（全部省）时为空 → notify 回退全量广播
	h.notify(a, "submitted", orgNotify)
	c.JSON(http.StatusCreated, a)
}

// List GET /api/v1/approvals?status=pending&resource=retail_contracts&mine=true&limit=50&offset=0
// 响应含 total(同条件总行数),供前端分页。
func (h *ApprovalHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	f := db.ApprovalFilter{
		Status:   c.Query("status"),
		Resource: c.Query("resource"),
		Limit:    limit,
		Offset:   offset,
	}
	if c.Query("mine") == "true" {
		if uid := claimsUserID(c); uid != nil {
			f.Submitter = uid.String()
		}
	}
	list, total, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": total})
}

// ListTemplates GET /api/v1/approvals/templates?resource=
func (h *ApprovalHandler) ListTemplates(c *gin.Context) {
	list, err := h.repo.ListTemplates(c.Request.Context(), c.Query("resource"))
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// ByResource GET /api/v1/approvals/by-resource?resource=&resource_id=
func (h *ApprovalHandler) ByResource(c *gin.Context) {
	resource := c.Query("resource")
	resourceID := c.Query("resource_id")
	if resource == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 resource 或 resource_id"})
		return
	}
	list, err := h.repo.ByResource(c.Request.Context(), resource, resourceID)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Get GET /api/v1/approvals/:id
func (h *ApprovalHandler) Get(c *gin.Context) {
	a, err := h.repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

type transitionReq struct {
	Note string `json:"note"`
}

// Approve POST /api/v1/approvals/:id/approve
// 单事务原子化（P0-B2）：Get + Transition + applier.Apply 在同一事务内完成，
// 任一步失败则整笔回滚，杜绝「已 approved 但未落库」/「部分写入后状态错乱」。
// Transition 带 status 乐观锁（P0-B8）：并发审批只有一方成功。
func (h *ApprovalHandler) Approve(c *gin.Context) {
	var req transitionReq
	_ = c.ShouldBindJSON(&req)
	reviewer := ""
	if uid := claimsUserID(c); uid != nil {
		reviewer = uid.String()
	}
	id := c.Param("id")
	ctx := c.Request.Context()

	tx, err := h.repo.BeginTx(ctx)
	if err != nil {
		log.Error().Err(err).Msg("开启审批事务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// 1) 先取当前审批以拿到 resource / payload（事务内）
	cur, err := h.repo.GetTx(ctx, tx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 2) 状态机迁移到 approved（事务内 + 乐观锁）
	a, err := h.repo.TransitionTx(ctx, tx, id, "approved", reviewer, req.Note)
	if err != nil {
		if err == db.ErrInvalidApprovalTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 3) 调用 applier 自动落库（同一事务内）；失败则整笔回滚（无需手工补偿）
	if err := h.registry.Apply(ctx, tx, cur.Resource, cur.ResourceID, cur.Payload); err != nil {
		log.Error().Err(err).Msg("审批自动落库失败，事务回滚")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("提交审批事务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	committed = true

	orgNotifyApprove, _ := db.OrgFilter(ctx)
	h.notify(a, "approved", orgNotifyApprove)
	c.JSON(http.StatusOK, a)
}

// Reject POST /api/v1/approvals/:id/reject
func (h *ApprovalHandler) Reject(c *gin.Context) {
	h.transition(c, "rejected")
}

// Withdraw POST /api/v1/approvals/:id/withdraw
func (h *ApprovalHandler) Withdraw(c *gin.Context) {
	h.transition(c, "withdrawn")
}

func (h *ApprovalHandler) transition(c *gin.Context, target string) {
	var req transitionReq
	_ = c.ShouldBindJSON(&req)
	reviewer := ""
	if uid := claimsUserID(c); uid != nil {
		reviewer = uid.String()
	}

	a, err := h.repo.Transition(c.Request.Context(), c.Param("id"), target, reviewer, req.Note)
	if err != nil {
		if err == db.ErrApprovalNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == db.ErrInvalidApprovalTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	orgNotifyTrans, _ := db.OrgFilter(c.Request.Context())
	h.notify(a, target, orgNotifyTrans)
	c.JSON(http.StatusOK, a)
}

func claimsUsername(c *gin.Context) string {
	v, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		return ""
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		return ""
	}
	return cl.Username
}
