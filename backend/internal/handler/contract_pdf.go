// 合同 PDF 生成 handler：从合同 ID 取数据 → 生成 PDF → 上传 MinIO → 创建附件记录。
package handler

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/ptis/backend/internal/contractpdf"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ContractPDFHandler struct {
	retail   *db.RetailRepository
	attaches *db.AttachmentRepository
	store    *storage.ObjectStore
}

func NewContractPDFHandler(r *db.RetailRepository, a *db.AttachmentRepository, s *storage.ObjectStore) *ContractPDFHandler {
	return &ContractPDFHandler{retail: r, attaches: a, store: s}
}

// Generate POST /api/v1/retail/contracts/:id/pdf
// 生成 PDF → 上传 MinIO → 关联到 attachments。
func (h *ContractPDFHandler) Generate(c *gin.Context) {
	idStr := c.Param("id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "合同 ID 格式错误"})
		return
	}
	ct, err := h.retail.GetContract(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
		return
	}

	pdfBytes, err := contractpdf.Generate(contractpdf.ContractData{
		ContractID:         ct.ID.String(),
		CustomerName:       ct.CustomerName,
		PackageName:        ct.PackageNameSnapshot,
		PurchasingEnergy:   ct.PurchasingEnergyMWH,
		GreenPowerRatio:    derefF(ct.GreenPowerRatio),
		PurchaseStartMonth: ct.PurchaseStartMonth,
		PurchaseEndMonth:   ct.PurchaseEndMonth,
		Status:             ct.Status,
	})
	if err != nil {
		log.Error().Err(err).Msg("PDF 生成失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	if h.store == nil {
		// 无 MinIO 时直接返回二进制供下载
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"contract_%s.pdf\"", ct.ID.String()))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
		return
	}

	// 上传到 MinIO + 写 attachments 记录
	filename := fmt.Sprintf("contract_%s.pdf", ct.ID.String())
	objectKey := fmt.Sprintf("contracts/%s/%s", ct.ID.String(), filename)
	if err := h.store.Put(c.Request.Context(), objectKey, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf"); err != nil {
		log.Error().Err(err).Msg("上传 MinIO 失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	username := claimsUsername(c)
	att, err := h.attaches.Create(c.Request.Context(), db.AttachmentInput{
		Resource:    "retail_contracts",
		ResourceID:  ct.ID.String(),
		Filename:    filename,
		ObjectKey:   objectKey,
		ContentType: "application/pdf",
		Size:        int64(len(pdfBytes)),
		UploadedBy:  username,
		Note:        "系统自动生成",
	})
	if err != nil {
		log.Error().Err(err).Msg("写附件记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":    "已生成合同 PDF 并加为附件",
		"attachment": att,
	})
}

func derefF(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

