// 政策文件库 handler：列出/手动新增/删除政策条目（按活跃省隔离）。
// 文档解析「确认入库 → 政策文件」自动归纳的条目也在此查看。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type PolicyHandler struct {
	repo *db.PolicyRepository
}

func NewPolicyHandler(repo *db.PolicyRepository) *PolicyHandler {
	return &PolicyHandler{repo: repo}
}

// List GET /api/v1/policies?category=&limit=
func (h *PolicyHandler) List(c *gin.Context) {
	category := c.Query("category")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	list, err := h.repo.List(c.Request.Context(), category, limit)
	if err != nil {
		log.Error().Err(err).Msg("查询政策文件失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Create POST /api/v1/policies（手动新增）
func (h *PolicyHandler) Create(c *gin.Context) {
	var req struct {
		Title         string `json:"title"`
		DocNo         string `json:"doc_no"`
		Category      string `json:"category"`
		EffectiveDate string `json:"effective_date"`
		Summary       string `json:"summary"`
		DocumentID    string `json:"document_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少政策标题 title"})
		return
	}
	createdBy := ""
	if uid := claimsUserID(c); uid != nil {
		createdBy = uid.String()
	}
	id, err := h.repo.Create(c.Request.Context(), db.PolicyInput{
		DocumentID:    req.DocumentID,
		Title:         req.Title,
		DocNo:         req.DocNo,
		Category:      req.Category,
		EffectiveDate: req.EffectiveDate,
		Summary:       req.Summary,
		Source:        "manual",
		CreatedBy:     createdBy,
	})
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("新增政策失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "已新增政策"})
}

// Delete DELETE /api/v1/policies/:id
func (h *PolicyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		log.Error().Err(err).Msg("删除政策失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
