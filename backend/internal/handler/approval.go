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

func (h *ApprovalHandler) notify(a *db.Approval, action string) {
	if h.hub == nil {
		return
	}
	h.hub.Publish(SSEEvent{
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
	h.notify(a, "submitted")
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
// 通过后调用 applier 自动落库（如有注册）；落库失败则回滚状态。
func (h *ApprovalHandler) Approve(c *gin.Context) {
	var req transitionReq
	_ = c.ShouldBindJSON(&req)
	reviewer := ""
	if uid := claimsUserID(c); uid != nil {
		reviewer = uid.String()
	}
	id := c.Param("id")

	// 1) 先取当前审批以拿到 resource / payload
	cur, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 2) 状态机迁移到 approved
	a, err := h.repo.Transition(c.Request.Context(), id, "approved", reviewer, req.Note)
	if err != nil {
		if err == db.ErrInvalidApprovalTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 3) 调用 applier 自动落库；失败时回滚状态（保留 review_note 提示）
	if err := h.registry.Apply(c.Request.Context(), cur.Resource, cur.ResourceID, cur.Payload); err != nil {
		// 回滚：手工把状态拉回 pending，附带错误信息
		_, _ = h.repo.Transition(c.Request.Context(), id, "pending", reviewer, "落库失败: "+err.Error())
		log.Error().Err(err).Msg("审批已通过但自动落库失败，已回滚为待审批")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	h.notify(a, "approved")
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
	h.notify(a, target)
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
