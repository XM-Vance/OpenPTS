// 客户档案 CRUD handler。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/ptis/backend/internal/masking"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type CustomersHandler struct {
	repo    *db.CustomerRepository
	permSvc *auth.PermissionService
}

func NewCustomersHandler(repo *db.CustomerRepository, permSvc *auth.PermissionService) *CustomersHandler {
	return &CustomersHandler{repo: repo, permSvc: permSvc}
}

// canViewSensitive 检查当前用户是否有权查看客户敏感信息。
func (h *CustomersHandler) canViewSensitive(c *gin.Context) bool {
	if h.permSvc == nil {
		return true // 无权限服务时默认放行
	}
	uid := claimsUserID(c)
	if uid == nil {
		return false
	}
	ok, _ := h.permSvc.Has(c.Request.Context(), *uid, "customer_management:view_sensitive")
	return ok
}

// maskCustomerForUser 按权限脱敏客户敏感字段。
func (h *CustomersHandler) maskCustomerForUser(c *gin.Context, cust *db.Customer) {
	if h.canViewSensitive(c) {
		return // 有权限，不脱敏
	}
	cust.UserName = masking.MaskName(cust.UserName)
	if cust.ShortName != nil {
		s := masking.MaskName(*cust.ShortName)
		cust.ShortName = &s
	}
	if cust.Manager != nil {
		s := masking.MaskManager(*cust.Manager)
		cust.Manager = &s
	}
	if cust.Location != nil {
		s := masking.MaskLocation(*cust.Location)
		cust.Location = &s
	}
}

// List GET /api/v1/customers?keyword=&tag=&manager=&limit=&offset=
func (h *CustomersHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.repo.List(c.Request.Context(), db.CustomerListFilter{
		Keyword: c.Query("keyword"),
		Tag:     c.Query("tag"),
		Manager: c.Query("manager"),
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	// 按权限自动脱敏（无 view_sensitive 权限的用户看到脱敏数据）
	for _, cu := range items {
		h.maskCustomerForUser(c, cu)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "masked": !h.canViewSensitive(c)})
}

// Get GET /api/v1/customers/:id
func (h *CustomersHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	cust, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
		return
	}
	h.maskCustomerForUser(c, cust)
	c.JSON(http.StatusOK, cust)
}

type CustomerRequest struct {
	UserName  string          `json:"user_name" binding:"required,min=1,max=255"`
	ShortName string          `json:"short_name"`
	Location  string          `json:"location"`
	Source    string          `json:"source"`
	Manager   string          `json:"manager"`
	Tags      []string        `json:"tags"`
	Accounts  json.RawMessage `json:"accounts"`
	IsDemo    bool            `json:"is_demo"`
}

func (req CustomerRequest) toInput() db.CustomerInput {
	return db.CustomerInput{
		UserName:  req.UserName,
		ShortName: req.ShortName,
		Location:  req.Location,
		Source:    req.Source,
		Manager:   req.Manager,
		Tags:      req.Tags,
		Accounts:  req.Accounts,
		IsDemo:    req.IsDemo,
	}
}

// Create POST /api/v1/customers
func (h *CustomersHandler) Create(c *gin.Context) {
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var createdBy *uuid.UUID
	if claimsAny, ok := c.Get(auth.ClaimsContextKey); ok {
		uid := claimsAny.(*auth.Claims).UserID
		createdBy = &uid
	}
	cust, err := h.repo.Create(c.Request.Context(), req.toInput(), createdBy)
	if err != nil {
		if errors.Is(err, db.ErrOrgRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, cust)
}

// Update PUT /api/v1/customers/:id
func (h *CustomersHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	cust, err := h.repo.Update(c.Request.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, db.ErrCustomerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, cust)
}

// Delete DELETE /api/v1/customers/:id
func (h *CustomersHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusConflict, gin.H{"error": "该客户存在关联数据（如合同/负荷），无法删除"})
			return
		}
		if errors.Is(err, db.ErrCustomerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// View360 GET /api/v1/customers/:id/360 — 客户360视图，聚合所有关联数据。
func (h *CustomersHandler) View360(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式错误"})
		return
	}
	cust, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客户不存在"})
		return
	}

	// 用 pool 聚合查询关联数据。
	// 多租户：与 repo 层一致，活跃省为具体省时按 org_id 过滤，防止跨租户读取
	// （客户本身已由 GetByID 按 org 作用域过滤，此处为关联数据的纵深防御）。
	pool := h.repo.Pool()
	org, scoped := db.OrgFilter(c.Request.Context())

	type linkedRow struct {
		Query string
		Args  []any
	}
	// orgScope 为具体省时追加 "AND org_id = $N::uuid"，参数序号随 args 增长。
	orgClause := func(n int) string {
		if !scoped {
			return ""
		}
		return fmt.Sprintf(" AND org_id = $%d::uuid", n)
	}
	base := []any{id}
	if scoped {
		base = append(base, org) // 第 2 个参数固定为 org_id，各查询占位符序号均为 2
	}
	queries := map[string]linkedRow{
		"contracts":       {`SELECT id::text, customer_name, status, contract_type, start_date, end_date FROM retail_contracts WHERE customer_id=$1` + orgClause(2) + ` ORDER BY created_at DESC LIMIT 50`, base},
		"documents":       {`SELECT id::text, filename, doc_type, status, created_at FROM documents WHERE customer_id=$1` + orgClause(2) + ` ORDER BY created_at DESC LIMIT 50`, base},
		"settlements":     {`SELECT id::text, period, total_amount, status FROM monthly_settlements WHERE customer_id=$1` + orgClause(2) + ` ORDER BY period DESC LIMIT 50`, base},
		"alerts":          {`SELECT id::text, alert_type, severity, message, created_at FROM customer_alerts WHERE customer_id=$1` + orgClause(2) + ` ORDER BY created_at DESC LIMIT 50`, base},
		"stations":        {`SELECT id::text, station_name, capacity_kw FROM solar_stations WHERE customer_id=$1` + orgClause(2) + ` ORDER BY station_name`, base},
		"load_profile":    {`SELECT id::text, month, max_demand, avg_demand, load_factor FROM customer_load_profiles WHERE customer_id=$1` + orgClause(2) + ` ORDER BY month DESC LIMIT 12`, base},
		"profit":          {`SELECT id::text, month, revenue, cost, profit, profit_margin FROM customer_profits WHERE customer_id=$1` + orgClause(2) + ` ORDER BY month DESC LIMIT 12`, base},
		"characteristics": {`SELECT id::text, month, industry_type, voltage_level, contract_capacity, max_demand FROM load_characteristics WHERE customer_id=$1` + orgClause(2) + ` ORDER BY month DESC LIMIT 12`, base},
	}

	result := gin.H{"customer": cust}
	for key, q := range queries {
		rows, err := pool.Query(c.Request.Context(), q.Query, q.Args...)
		if err != nil {
			result[key] = []any{}
			continue
		}
		cols := rows.FieldDescriptions()
		var items []map[string]any
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			row := make(map[string]any)
			for i, col := range cols {
				row[col.Name] = vals[i]
			}
			items = append(items, row)
		}
		rows.Close()
		if items == nil {
			items = []map[string]any{}
		}
		result[key] = items
	}

	c.JSON(http.StatusOK, result)
}
