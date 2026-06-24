// 文档解析管线 handler：上传(存原件+去重) → 异步解析 → 详情(字段核对) → 确认入库。
// RBAC document_management；文档按省份隔离（上传需具体活跃省）。
package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/document"
	"github.com/ptis/backend/internal/storage"
	"github.com/rs/zerolog/log"
)

type DocumentHandler struct {
	repo      *db.DocumentRepository
	worker    *document.Worker
	importers *document.Importers
	store     *storage.ObjectStore
	permSvc   *auth.PermissionService
}

func NewDocumentHandler(repo *db.DocumentRepository, worker *document.Worker,
	importers *document.Importers, store *storage.ObjectStore,
	permSvc *auth.PermissionService) *DocumentHandler {
	return &DocumentHandler{repo: repo, worker: worker, importers: importers, store: store, permSvc: permSvc}
}

// 单文件上限 32MB。
const maxDocUploadBytes = 32 << 20

// Upload POST /api/v1/documents（multipart：file）→ 存原件 + 异步解析，立即返回。
func (h *DocumentHandler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少上传文件 file"})
		return
	}
	if fh.Size > maxDocUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "文件过大（上限 32MB）"})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取上传文件失败"})
		return
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, maxDocUploadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取上传文件失败"})
		return
	}

	sum := sha256.Sum256(content)
	sha := hex.EncodeToString(sum[:])

	// 同省同内容去重：直接返回已有文档
	if exist, err := h.repo.FindBySha(c.Request.Context(), sha); err == nil && exist != nil {
		c.JSON(http.StatusOK, gin.H{"document": exist, "duplicated": true,
			"message": "相同内容的文件已存在，未重复入库"})
		return
	}

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 先建记录拿 id（含省份校验），再按 id 归档原件
	doc, err := h.repo.Create(c.Request.Context(), fh.Filename, contentType, fh.Size,
		sha, document.SourceKind(fh.Filename), "", claimsUserID(c))
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("创建文档记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 原件归档（MinIO 未配置时降级：仅解析不存原件，前端隐藏下载）
	if h.store.Enabled() {
		org := ""
		if doc.OrgID != nil {
			org = *doc.OrgID
		}
		key := document.OriginalKey(org, doc.ID, fh.Filename)
		if err := h.store.Put(c.Request.Context(), key, bytes.NewReader(content),
			int64(len(content)), contentType); err != nil {
			log.Warn().Err(err).Msg("原件写入对象存储失败（继续解析，原件未归档）")
		} else {
			if err := h.repo.SetOriginalKey(c.Request.Context(), uuid.MustParse(doc.ID), key); err == nil {
				doc.OriginalObjectKey = &key
			}
		}
	} else {
		log.Warn().Msg("MINIO 未配置，原件未归档")
	}

	h.worker.Enqueue(uuid.MustParse(doc.ID), content)

	// 可选：关联业务实体（customer_id / contract_id / intent_customer_id）
	customerID := c.Query("customer_id")
	contractID := c.Query("contract_id")
	intentCustomerID := c.Query("intent_customer_id")
	if customerID != "" || contractID != "" || intentCustomerID != "" {
		if err := h.repo.LinkEntities(c.Request.Context(), uuid.MustParse(doc.ID), customerID, contractID, intentCustomerID); err != nil {
			log.Warn().Err(err).Msg("关联实体失败（文档已创建）")
		}
	}

	c.JSON(http.StatusAccepted, gin.H{"document": doc, "message": "已接收，正在解析"})
}

// List GET /api/v1/documents?status=&doc_type=&limit=
func (h *DocumentHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	// 文档隔离策略（按角色分）：
	//   - 拥有 document_management:read_all 权限（admin/analyst/super_admin）→ 看同省所有文档
	//   - 没有（viewer）→ 仅看自己上传的
	//   - scope=mine 可强制只看自己的（任何角色）
	var uploadedBy *uuid.UUID
	cl := authClaims(c)
	if cl != nil && h.permSvc != nil {
		canReadAll, _ := h.permSvc.Has(c.Request.Context(), cl.UserID, "document_management:read_all")
		if !canReadAll {
			uid := cl.UserID
			uploadedBy = &uid
		}
	}
	if c.Query("scope") == "mine" {
		if uid := claimsUserID(c); uid != nil {
			uploadedBy = uid
		}
	}
	list, err := h.repo.List(c.Request.Context(), c.Query("status"), c.Query("doc_type"), limit, uploadedBy)
	if err != nil {
		log.Error().Err(err).Msg("查询文档列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *DocumentHandler) docByParam(c *gin.Context) (*db.Document, uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档 id 非法"})
		return nil, uuid.Nil, false
	}
	doc, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("查询文档失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return nil, uuid.Nil, false
	}
	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return nil, uuid.Nil, false
	}
	return doc, id, true
}

// Get GET /api/v1/documents/:id → 详情 + 提取字段 + 入库历史 + 推荐目标。
func (h *DocumentHandler) Get(c *gin.Context) {
	doc, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	exts, err := h.repo.ListExtractions(c.Request.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("查询提取字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	applies, _ := h.repo.ListApplies(c.Request.Context(), id)
	suggest := ""
	if doc.DocType != nil {
		suggest = document.SuggestTarget(*doc.DocType)
	}
	c.JSON(http.StatusOK, gin.H{
		"document": doc, "extractions": exts, "applies": applies,
		"suggest_target":  suggest,
		"storage_enabled": h.store.Enabled(),
	})
}

// serveFile 设置下载头并回文件字节（文件名兼容中文：RFC 5987）。
func serveFile(c *gin.Context, filename, contentType string, data []byte) {
	c.Header("Content-Disposition",
		fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename)))
	c.Data(http.StatusOK, contentType, data)
}

// Original GET /api/v1/documents/:id/original → 经网关流式下载原件。
// 不用 presigned URL：MinIO 通常在内网（docker 网络/未暴露主机），浏览器直连会失败。
func (h *DocumentHandler) Original(c *gin.Context) {
	doc, _, ok := h.docByParam(c)
	if !ok {
		return
	}
	if !h.store.Enabled() || doc.OriginalObjectKey == nil || *doc.OriginalObjectKey == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "原件未归档（历史数据或对象存储未配置）"})
		return
	}
	rc, err := h.store.Get(c.Request.Context(), *doc.OriginalObjectKey)
	if err != nil {
		log.Error().Err(err).Msg("取回原件失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		log.Error().Err(err).Msg("读取原件失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	serveFile(c, doc.Filename, doc.ContentType, data)
}

// Parsed GET /api/v1/documents/:id/parsed → 下载解析件(.md)。
// 对象存储缺失时回退库内全文（解析件下载不依赖 MinIO）。
func (h *DocumentHandler) Parsed(c *gin.Context) {
	doc, _, ok := h.docByParam(c)
	if !ok {
		return
	}
	filename := doc.Filename + ".md"
	if h.store.Enabled() && doc.ParsedObjectKey != nil && *doc.ParsedObjectKey != "" {
		if rc, err := h.store.Get(c.Request.Context(), *doc.ParsedObjectKey); err == nil {
			defer rc.Close()
			if data, err := io.ReadAll(rc); err == nil {
				serveFile(c, filename, "text/markdown; charset=utf-8", data)
				return
			}
		}
		log.Warn().Msg("解析件对象读取失败，回退库内全文")
	}
	if doc.TextContent == nil || *doc.TextContent == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "解析件未生成"})
		return
	}
	serveFile(c, filename, "text/markdown; charset=utf-8", []byte(*doc.TextContent))
}

// AddExtraction POST /api/v1/documents/:id/extractions → 人工补充字段。
func (h *DocumentHandler) AddExtraction(c *gin.Context) {
	_, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	var req struct {
		GroupNo    int    `json:"group_no"`
		FieldKey   string `json:"field_key"`
		FieldLabel string `json:"field_label"`
		ValueText  string `json:"value_text"`
		Unit       string `json:"unit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.FieldLabel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少字段名 field_label"})
		return
	}
	if req.FieldKey == "" {
		req.FieldKey = fmt.Sprintf("manual_%d", time.Now().UnixMilli()%1_000_000_000)
	}
	e, err := h.repo.AddExtraction(c.Request.Context(), id,
		req.GroupNo, req.FieldKey, req.FieldLabel, req.ValueText, req.Unit)
	if err != nil {
		log.Error().Err(err).Msg("新增提取字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"extraction": e})
}

// DeleteExtraction DELETE /api/v1/documents/:id/extractions/:eid → 删除字段。
func (h *DocumentHandler) DeleteExtraction(c *gin.Context) {
	_, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	eid, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "字段 id 非法"})
		return
	}
	if err := h.repo.DeleteExtraction(c.Request.Context(), id, eid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// UpdateExtraction PUT /api/v1/documents/:id/extractions/:eid {value_text} → 人工修正字段。
func (h *DocumentHandler) UpdateExtraction(c *gin.Context) {
	_, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	eid, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "字段 id 非法"})
		return
	}
	var req struct {
		ValueText string `json:"value_text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}
	if err := h.repo.UpdateExtraction(c.Request.Context(), id, eid, req.ValueText); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已更新"})
}

// Reparse POST /api/v1/documents/:id/reparse → 取回原件重新解析。
func (h *DocumentHandler) Reparse(c *gin.Context) {
	doc, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	if doc.OriginalObjectKey == nil || *doc.OriginalObjectKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原件未归档，无法重新解析"})
		return
	}
	if err := h.repo.SetStatus(c.Request.Context(), id, "uploaded", ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	h.worker.Enqueue(id, nil) // nil → worker 从对象存储取回原件
	c.JSON(http.StatusAccepted, gin.H{"message": "已加入重新解析队列"})
}

// Apply POST /api/v1/documents/:id/apply {target} → 确认入库到业务表。
func (h *DocumentHandler) Apply(c *gin.Context) {
	doc, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	if doc.Status != "parsed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档尚未解析完成，无法入库"})
		return
	}
	var req struct {
		Target string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少入库目标 target"})
		return
	}

	// 高风险目标（建客户/合同）手动入库：硬性要求对应业务写权限。
	// bot（agent 只读）即便有 document_management:write，也不能直接建主数据/合同——
	// 只能由有权限的人确认。自动路径(worker)对高风险目标本就不入库，故这里专管手动/bot 直调。
	if perm := document.ReviewTargetPermission(req.Target); perm != "" {
		uid := claimsUserID(c)
		if uid == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "高风险入库需登录用户确认"})
			return
		}
		if okPerm, _ := h.permSvc.Has(c.Request.Context(), *uid, perm); !okPerm {
			c.JSON(http.StatusForbidden, gin.H{"error": "该入库目标为高风险，需要业务写权限（人工确认）：" + perm})
			return
		}
	}

	exts, err := h.repo.ListExtractions(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 数据写进文档归属省（而非请求者当前活跃省，二者可能不同——如总部在「全部省」视图）
	applyCtx := c.Request.Context()
	if doc.OrgID != nil && *doc.OrgID != "" {
		applyCtx = db.WithOrg(applyCtx, *doc.OrgID)
	}
	applied, detail, err := h.importers.Apply(applyCtx, req.Target, id.String(), exts)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.InsertApply(c.Request.Context(), id, req.Target, applied, detail, claimsUserID(c)); err != nil {
		log.Warn().Err(err).Msg("记录入库审计失败")
	}
	c.JSON(http.StatusOK, gin.H{"applied_rows": applied, "detail": detail,
		"message": "已入库 " + strconv.Itoa(applied) + " 行"})
}

// Delete DELETE /api/v1/documents/:id → 删除记录与归档文件。
func (h *DocumentHandler) Delete(c *gin.Context) {
	_, id, ok := h.docByParam(c)
	if !ok {
		return
	}
	origKey, parsedKey, err := h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.store.Enabled() {
		for _, k := range []string{origKey, parsedKey} {
			if k != "" {
				if err := h.store.Remove(c.Request.Context(), k); err != nil {
					log.Warn().Err(err).Str("key", k).Msg("删除归档对象失败")
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
