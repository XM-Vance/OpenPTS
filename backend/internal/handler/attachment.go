// 附件 handler：上传到 MinIO + 元信息入库 + 生成下载 presigned URL。
package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type AttachmentHandler struct {
	repo  *db.AttachmentRepository
	store *storage.ObjectStore
}

func NewAttachmentHandler(repo *db.AttachmentRepository, store *storage.ObjectStore) *AttachmentHandler {
	return &AttachmentHandler{repo: repo, store: store}
}

// Upload POST /api/v1/attachments?resource=customers&resource_id=<uuid>
func (h *AttachmentHandler) Upload(c *gin.Context) {
	if !h.store.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "对象存储未配置（MINIO_ENDPOINT 缺失）"})
		return
	}
	resource := c.Query("resource")
	resourceID := c.Query("resource_id")
	if resource == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 resource 或 resource_id"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段"})
		return
	}
	if fh.Size > 50<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大于 50MB"})
		return
	}
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	defer f.Close()

	// 生成 object key：resource/resource_id/<uuid>.<ext>
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	objectKey := fmt.Sprintf("%s/%s/%s%s", resource, resourceID, uuid.New().String(), ext)

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := h.store.Put(c.Request.Context(), objectKey, f, fh.Size, contentType); err != nil {
		log.Error().Err(err).Msg("上传失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	uploadedBy := ""
	if v, ok := c.Get(auth.ClaimsContextKey); ok {
		if cl, ok := v.(*auth.Claims); ok {
			uploadedBy = cl.Username
		}
	}

	att, err := h.repo.Create(c.Request.Context(), db.AttachmentInput{
		Resource:    resource,
		ResourceID:  resourceID,
		Filename:    fh.Filename,
		ObjectKey:   objectKey,
		ContentType: contentType,
		Size:        fh.Size,
		UploadedBy:  uploadedBy,
		Note:        c.PostForm("note"),
	})
	if err != nil {
		// 入库失败回滚对象
		_ = h.store.Remove(c.Request.Context(), objectKey)
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, att)
}

// List GET /api/v1/attachments?resource=customers&resource_id=<uuid>
func (h *AttachmentHandler) List(c *gin.Context) {
	resource := c.Query("resource")
	resourceID := c.Query("resource_id")
	if resource == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 resource 或 resource_id"})
		return
	}
	list, err := h.repo.ListByResource(c.Request.Context(), resource, resourceID)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// DownloadURL GET /api/v1/attachments/:id/url
// 返回临时签名 URL（10 分钟），前端浏览器直接 GET 这个 URL 下载。
func (h *AttachmentHandler) DownloadURL(c *gin.Context) {
	if !h.store.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "对象存储未配置"})
		return
	}
	id := c.Param("id")
	att, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "附件不存在"})
		return
	}
	url, err := h.store.PresignedGetURL(c.Request.Context(), att.ObjectKey, 10*time.Minute)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url, "filename": att.Filename})
}

// Delete DELETE /api/v1/attachments/:id
func (h *AttachmentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	objectKey, err := h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "附件不存在"})
		return
	}
	if h.store.Enabled() {
		_ = h.store.Remove(c.Request.Context(), objectKey)
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
