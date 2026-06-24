// 零售管理 handler：定价模型（只读）、零售套餐 CRUD、零售合同 CRUD。
package handler

import (
	"errors"
	"net/http"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type RetailHandler struct {
	repo *db.RetailRepository
}

func NewRetailHandler(repo *db.RetailRepository) *RetailHandler {
	return &RetailHandler{repo: repo}
}

func currentUserID(c *gin.Context) *uuid.UUID {
	if claimsAny, ok := c.Get(auth.ClaimsContextKey); ok {
		uid := claimsAny.(*auth.Claims).UserID
		return &uid
	}
	return nil
}

// ─── 定价模型 ───

// ListPricingModels GET /api/v1/retail/pricing-models
func (h *RetailHandler) ListPricingModels(c *gin.Context) {
	list, err := h.repo.ListPricingModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// ─── 零售套餐 ───

type packageRequest struct {
	PackageName  string `json:"package_name" binding:"required,min=1,max=255"`
	PackageType  string `json:"package_type" binding:"required,min=1,max=64"`
	ModelCode    string `json:"model_code"`
	IsGreenPower bool   `json:"is_green_power"`
	Status       string `json:"status"`
	Description  string `json:"description"`
}

func (req packageRequest) toInput() db.PackageInput {
	return db.PackageInput{
		PackageName:  req.PackageName,
		PackageType:  req.PackageType,
		ModelCode:    req.ModelCode,
		IsGreenPower: req.IsGreenPower,
		Status:       req.Status,
		Description:  req.Description,
	}
}

// ListPackages GET /api/v1/retail/packages?keyword=&status=
func (h *RetailHandler) ListPackages(c *gin.Context) {
	list, err := h.repo.ListPackages(c.Request.Context(), c.Query("keyword"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// CreatePackage POST /api/v1/retail/packages
func (h *RetailHandler) CreatePackage(c *gin.Context) {
	var req packageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	p, err := h.repo.CreatePackage(c.Request.Context(), req.toInput(), currentUserID(c))
	if err != nil {
		if errors.Is(err, db.ErrOrgRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// UpdatePackage PUT /api/v1/retail/packages/:id
func (h *RetailHandler) UpdatePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req packageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	p, err := h.repo.UpdatePackage(c.Request.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, db.ErrPackageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "套餐不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// DeletePackage DELETE /api/v1/retail/packages/:id
func (h *RetailHandler) DeletePackage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.DeletePackage(c.Request.Context(), id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusConflict, gin.H{"error": "该套餐已被合同引用，无法删除"})
			return
		}
		if errors.Is(err, db.ErrPackageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "套餐不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ─── 零售合同 ───

type contractRequest struct {
	CustomerID          string   `json:"customer_id" binding:"required,uuid"`
	PackageID           string   `json:"package_id" binding:"required,uuid"`
	PurchasingEnergyMWH *float64 `json:"purchasing_energy_mwh"`
	GreenPowerRatio     *float64 `json:"green_power_ratio"`
	PurchaseStartMonth  string   `json:"purchase_start_month" binding:"required,len=7"`
	PurchaseEndMonth    string   `json:"purchase_end_month" binding:"required,len=7"`
	Status              string   `json:"status"`
}

func (req contractRequest) toInput() (db.ContractInput, error) {
	cid, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return db.ContractInput{}, errors.New("customer_id 非法")
	}
	pid, err := uuid.Parse(req.PackageID)
	if err != nil {
		return db.ContractInput{}, errors.New("package_id 非法")
	}
	var energy float64
	if req.PurchasingEnergyMWH != nil {
		energy = *req.PurchasingEnergyMWH
	}
	return db.ContractInput{
		CustomerID:          cid,
		PackageID:           pid,
		PurchasingEnergyMWH: energy,
		GreenPowerRatio:     req.GreenPowerRatio,
		PurchaseStartMonth:  req.PurchaseStartMonth,
		PurchaseEndMonth:    req.PurchaseEndMonth,
		Status:              req.Status,
	}, nil
}

// ListContracts GET /api/v1/retail/contracts?keyword=&status=
func (h *RetailHandler) ListContracts(c *gin.Context) {
	list, err := h.repo.ListContracts(c.Request.Context(), c.Query("keyword"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// GetContract GET /api/v1/retail/contracts/:id — 单个合同详情（P1-7：补齐缺失的读路由）。
func (h *RetailHandler) GetContract(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	contract, err := h.repo.GetContract(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrContractNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, contract)
}

// CreateContract POST /api/v1/retail/contracts
func (h *RetailHandler) CreateContract(c *gin.Context) {
	var req contractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	in, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contract, err := h.repo.CreateContract(c.Request.Context(), in, currentUserID(c))
	if err != nil {
		if errors.Is(err, db.ErrPackageNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "所选套餐不存在"})
			return
		}
		if errors.Is(err, db.ErrOrgRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "所选客户不存在"})
			return
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "起始月份不能晚于结束月份"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, contract)
}

// UpdateContract PUT /api/v1/retail/contracts/:id
func (h *RetailHandler) UpdateContract(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req contractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	in, err := req.toInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contract, err := h.repo.UpdateContract(c.Request.Context(), id, in)
	if err != nil {
		if errors.Is(err, db.ErrContractNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "所选客户或套餐不存在"})
			return
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "起始月份不能晚于结束月份"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, contract)
}

// DeleteContract DELETE /api/v1/retail/contracts/:id
func (h *RetailHandler) DeleteContract(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.DeleteContract(c.Request.Context(), id); err != nil {
		if errors.Is(err, db.ErrContractNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "合同不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
