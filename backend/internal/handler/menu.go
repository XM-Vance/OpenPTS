// 菜单页面可见性管理 handler。
// GET  /api/v1/menu/pages       — 管理员：所有页面 + 每个角色是否勾选
// GET  /api/v1/menu/visible      — 当前用户：可见页面列表
// PUT  /api/v1/menu/roles/:code  — 管理员：更新某角色的可见页面集
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
)

// ── 数据模型 ──

type MenuPage struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	Href      string `json:"href"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	GroupName string `json:"group_name"`
	IsRequired bool  `json:"is_required"`
}

type MenuPageWithRoles struct {
	MenuPage
	Roles []string `json:"roles"` // 拥有此页面可见权限的 role_code 列表
}

type MenuHandler struct {
	pool *db.Pool
}

func NewMenuHandler(pool *db.Pool) *MenuHandler {
	return &MenuHandler{pool: pool}
}

// GetAllPages GET /api/v1/menu/pages
// 返回所有页面列表，每页附带已分配的角色 codes
func (h *MenuHandler) GetAllPages(c *gin.Context) {
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT p.code, p.label, p.href, p.icon, p.sort_order, p.group_name, p.is_required,
		       COALESCE(array_agg(rp.role_code) FILTER (WHERE rp.role_code IS NOT NULL), '{}') as roles
		FROM menu_pages p
		LEFT JOIN role_menu_pages rp ON p.code = rp.page_code
		GROUP BY p.code, p.label, p.href, p.icon, p.sort_order, p.group_name, p.is_required
		ORDER BY p.sort_order
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	pages := []MenuPageWithRoles{}
	for rows.Next() {
		var p MenuPageWithRoles
		if err := rows.Scan(&p.Code, &p.Label, &p.Href, &p.Icon, &p.SortOrder, &p.GroupName, &p.IsRequired, &p.Roles); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描失败"})
			return
		}
		pages = append(pages, p)
	}
	c.JSON(http.StatusOK, gin.H{"items": pages})
}

// GetVisiblePages GET /api/v1/menu/visible?role=<role_code>
// 返回当前用户角色可见的页面列表（前端侧边栏用）
func (h *MenuHandler) GetVisiblePages(c *gin.Context) {
	roleCode := c.Query("role")
	if roleCode == "" {
		// 从 JWT claims 取 userID，再查角色
		uid := claimsUserID(c)
		if uid != nil {
			err := h.pool.QueryRow(c.Request.Context(),
				"SELECT role_code FROM user_roles WHERE user_id = $1 LIMIT 1", *uid).Scan(&roleCode)
			if err != nil {
				roleCode = ""
			}
		}
	}
	if roleCode == "" {
		c.JSON(http.StatusOK, gin.H{"items": []MenuPage{}})
		return
	}

	pages, err := h.queryVisiblePages(c.Request.Context(), roleCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": pages})
}

func (h *MenuHandler) queryVisiblePages(ctx context.Context, roleCode string) ([]MenuPage, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT p.code, p.label, p.href, p.icon, p.sort_order, p.group_name, p.is_required
		FROM menu_pages p
		JOIN role_menu_pages r ON p.code = r.page_code
		WHERE r.role_code = $1
		ORDER BY p.sort_order
	`, roleCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := []MenuPage{}
	for rows.Next() {
		var p MenuPage
		if err := rows.Scan(&p.Code, &p.Label, &p.Href, &p.Icon, &p.SortOrder, &p.GroupName, &p.IsRequired); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, nil
}

// UpdateRolePages PUT /api/v1/menu/roles/:code
// Body: {"page_codes": ["page:dashboard", "page:customers", ...]}
func (h *MenuHandler) UpdateRolePages(c *gin.Context) {
	roleCode := c.Param("code")
	if roleCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少角色代码"})
		return
	}

	var body struct {
		PageCodes []string `json:"page_codes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 确保 necessary 页面始终包含
	requiredRows, err := h.pool.Query(c.Request.Context(),
		"SELECT code FROM menu_pages WHERE is_required = TRUE")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询必要页面失败"})
		return
	}
	requiredCodes := map[string]bool{}
	for requiredRows.Next() {
		var code string
		requiredRows.Scan(&code)
		requiredCodes[code] = true
	}
	requiredRows.Close()

	// 合并必要页面
	finalCodes := map[string]bool{}
	for _, code := range body.PageCodes {
		finalCodes[code] = true
	}
	for code := range requiredCodes {
		finalCodes[code] = true
	}

	// 事务：先删旧关联再插入新关联
	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "事务启动失败"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	_, err = tx.Exec(c.Request.Context(),
		"DELETE FROM role_menu_pages WHERE role_code = $1", roleCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除旧权限失败"})
		return
	}

	for code := range finalCodes {
		_, err = tx.Exec(c.Request.Context(),
			"INSERT INTO role_menu_pages (role_code, page_code) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			roleCode, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入权限失败"})
			return
		}
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	// 返回更新后的页面数
	count := len(finalCodes)
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"role":        roleCode,
		"page_count":  count,
		"message":     "菜单权限已更新",
	})
}

// ── 防止 unused import ──
var _ = errors.Is
var _ = json.Marshal
