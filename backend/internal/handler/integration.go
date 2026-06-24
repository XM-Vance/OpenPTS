// 模块联动 handler：文档→合同填充、意向客户转正。
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// IntegrationHandler 跨模块联动操作。
type IntegrationHandler struct {
	pool         *db.Pool
	docRepo      *db.DocumentRepository
	retailRepo   *db.RetailRepository
	customerRepo *db.CustomerRepository
	intentRepo   *db.IntentCustomerRepository
}

func NewIntegrationHandler(
	pool *db.Pool,
	docRepo *db.DocumentRepository,
	retailRepo *db.RetailRepository,
	customerRepo *db.CustomerRepository,
	intentRepo *db.IntentCustomerRepository,
) *IntegrationHandler {
	return &IntegrationHandler{
		pool:         pool,
		docRepo:      docRepo,
		retailRepo:   retailRepo,
		customerRepo: customerRepo,
		intentRepo:   intentRepo,
	}
}

// ApplyToContract POST /api/v1/documents/:id/apply-to-contract
// 从文档提取字段自动创建零售合同（草稿状态）。
func (h *IntegrationHandler) ApplyToContract(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文档 id 非法"})
		return
	}

	// 获取文档
	doc, err := h.docRepo.Get(c.Request.Context(), docID)
	if err != nil || doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 获取提取字段
	extrs, err := h.docRepo.ListExtractions(c.Request.Context(), docID)
	if err != nil {
		log.Error().Err(err).Msg("查询提取字段失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询提取字段失败"})
		return
	}

	// 将提取字段转为 map
	extMap := make(map[string]string)
	for _, e := range extrs {
		if e.ValueText != nil && *e.ValueText != "" {
			extMap[e.FieldKey] = *e.ValueText
		}
	}

	var req struct {
		CustomerID string `json:"customer_id"`
		PackageID  string `json:"package_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没传 body，尝试从文档关联的 customer_id 获取
		if doc.CustomerID != nil && *doc.CustomerID != "" {
			req.CustomerID = *doc.CustomerID
		}
	}

	if req.CustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定关联客户"})
		return
	}

	customerUUID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id 格式错误"})
		return
	}

	// 如果没传 package_id，尝试从提取字段匹配
	if req.PackageID == "" {
		if pkgName, ok := extMap["package_name"]; ok {
			// 按名称查找套餐
			var pkgID string
			org, _ := db.OrgFilter(c.Request.Context())
			err := h.pool.QueryRow(c.Request.Context(),
				`SELECT id::text FROM retail_packages WHERE package_name ILIKE '%' || $1 || '%' AND org_id=$2::uuid LIMIT 1`,
				pkgName, org).Scan(&pkgID)
			if err == nil {
				req.PackageID = pkgID
			}
		}
	}

	if req.PackageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法确定套餐，请指定 package_id"})
		return
	}

	pkgUUID, err := uuid.Parse(req.PackageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package_id 格式错误"})
		return
	}

	// 构建合同入参
	in := db.ContractInput{
		CustomerID:          customerUUID,
		PackageID:           pkgUUID,
		PurchasingEnergyMWH: parseFloat(extMap, "energy", "purchasing_energy_mwh", "electricity"),
		PurchaseStartMonth:  extMap["start_date"],
		PurchaseEndMonth:    extMap["end_date"],
		Status:              "draft",
	}

	// 绿电比例
	if ratioStr, ok := extMap["green_power_ratio"]; ok {
		var ratio float64
		if _, err := fmt.Sscanf(ratioStr, "%f", &ratio); err == nil {
			ratio = ratio / 100 // 百分比转小数
			in.GreenPowerRatio = &ratio
		}
	}

	createdBy := claimsUserID(c)

	contract, err := h.retailRepo.CreateContract(c.Request.Context(), in, createdBy)
	if err != nil {
		log.Error().Err(err).Msg("创建合同失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建合同失败: " + err.Error()})
		return
	}

	// 回写文档的 contract_id
	if err := h.docRepo.SetContractID(c.Request.Context(), docID, contract.ID); err != nil {
		log.Warn().Err(err).Msg("回写文档 contract_id 失败")
	}

	// 记录入库操作
	detail, _ := json.Marshal(map[string]any{
		"contract_id":      contract.ID,
		"extractions_used": extMap,
	})
	_ = h.docRepo.InsertApply(c.Request.Context(), docID, "contract", 1, detail, claimsUserID(c))

	c.JSON(http.StatusCreated, gin.H{
		"contract":  contract,
		"document":  doc,
		"mapped_fields": extMap,
		"message":   "已从文档提取字段创建合同草稿",
	})
}

// ConvertIntentCustomer POST /api/v1/intent-customers/:id/convert
// 意向客户转正：创建正式客户 + 可选创建合同。
func (h *IntegrationHandler) ConvertIntentCustomer(c *gin.Context) {
	intentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}

	var req struct {
		CustomerName string `json:"customer_name"`
		PackageID    string `json:"package_id"`
		SignDate     string `json:"sign_date"`
	}
	_ = c.ShouldBindJSON(&req)

	// 获取意向客户信息
	var intentName, intentStatus string
	err = h.pool.QueryRow(c.Request.Context(),
		`SELECT customer_name, status FROM intent_customers WHERE id=$1`,
		intentID).Scan(&intentName, &intentStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "意向客户不存在"})
		return
	}
	if intentStatus == "converted" {
		c.JSON(http.StatusConflict, gin.H{"error": "该意向客户已转正"})
		return
	}

	// 创建正式客户
	name := req.CustomerName
	if name == "" {
		name = intentName
	}

	createdBy := claimsUserID(c)

	cust, err := h.customerRepo.Create(c.Request.Context(), db.CustomerInput{
		UserName: name,
		Source:   "意向客户转正",
		Tags:     []string{"意向转正"},
	}, createdBy)
	if err != nil {
		if err == db.ErrOrgRequired {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
			return
		}
		log.Error().Err(err).Msg("创建客户失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建客户失败"})
		return
	}

	// 更新意向客户状态
	_, err = h.pool.Exec(c.Request.Context(),
		`UPDATE intent_customers SET status='converted', converted_to=$2 WHERE id=$1`,
		intentID, cust.ID)
	if err != nil {
		log.Warn().Err(err).Msg("更新意向客户状态失败")
	}

	// 迁移关联文档
	_, _ = h.pool.Exec(c.Request.Context(),
		`UPDATE documents SET customer_id=$2, intent_customer_id=$1 WHERE intent_customer_id=$1`,
		intentID, cust.ID)

	result := gin.H{
		"customer":    cust,
		"intent_id":   intentID,
		"message":     "意向客户转正成功",
	}

	// 可选：创建合同
	if req.PackageID != "" {
		pkgUUID, err := uuid.Parse(req.PackageID)
		if err == nil {
			contract, err := h.retailRepo.CreateContract(c.Request.Context(), db.ContractInput{
				CustomerID: cust.ID,
				PackageID:  pkgUUID,
				Status:     "draft",
			}, createdBy)
			if err != nil {
				log.Warn().Err(err).Msg("转正时创建合同失败（客户已创建）")
				result["contract_error"] = err.Error()
			} else {
				result["contract"] = contract
				result["message"] = "意向客户转正成功，已创建合同草稿"
			}
		}
	}

	c.JSON(http.StatusCreated, result)
}

// parseFloat 从提取字段 map 中尝试多个 key 获取数值。
func parseFloat(m map[string]string, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				return f
			}
		}
	}
	return 0
}
