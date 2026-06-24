// 全局搜索 handler：跨 customers / retail_contracts / documents / intent_customers 搜索。
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// SearchHandler 全局搜索 handler。
type SearchHandler struct{ pool *db.Pool }

// NewSearchHandler 创建全局搜索 handler。
func NewSearchHandler(pool *db.Pool) *SearchHandler {
	return &SearchHandler{pool: pool}
}

// searchCustomer 搜索客户结果项。
type searchCustomer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// searchContract 搜索合同结果项。
type searchContract struct {
	ID           string `json:"id"`
	CustomerName string `json:"customer_name"`
	Status       string `json:"status"`
}

// searchDocument 搜索文档结果项。
type searchDocument struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	DocType  string `json:"doc_type"`
}

// searchIntentCustomer 搜索意向客户结果项。
type searchIntentCustomer struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
	Status      string `json:"status"`
}

// Search GET /api/v1/search?q=keyword&entity_types=customer,contract,document
func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少搜索关键词 q"})
		return
	}
	keyword := "%" + q + "%"

	// 解析要搜索的实体类型，默认全部
	entities := map[string]bool{
		"customer":        true,
		"contract":        true,
		"document":        true,
		"intent_customer": true,
	}
	if et := c.Query("entity_types"); et != "" {
		entities = make(map[string]bool)
		for _, e := range strings.Split(et, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				entities[e] = true
			}
		}
	}

	ctx := c.Request.Context()
	org, scoped := db.OrgFilter(ctx)

	result := gin.H{}

	if entities["customer"] {
		customers, err := h.searchCustomers(ctx, keyword, org, scoped)
		if err != nil {
			log.Error().Err(err).Msg("搜索客户失败")
		} else {
			result["customers"] = customers
		}
	}

	if entities["contract"] {
		contracts, err := h.searchContracts(ctx, keyword, org, scoped)
		if err != nil {
			log.Error().Err(err).Msg("搜索合同失败")
		} else {
			result["contracts"] = contracts
		}
	}

	if entities["document"] {
		documents, err := h.searchDocuments(ctx, keyword, org, scoped)
		if err != nil {
			log.Error().Err(err).Msg("搜索文档失败")
		} else {
			result["documents"] = documents
		}
	}

	if entities["intent_customer"] {
		intentCustomers, err := h.searchIntentCustomers(ctx, keyword, org, scoped)
		if err != nil {
			log.Error().Err(err).Msg("搜索意向客户失败")
		} else {
			result["intent_customers"] = intentCustomers
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) searchCustomers(ctx context.Context, keyword, org string, scoped bool) ([]searchCustomer, error) {
	q := `SELECT id::text, user_name, COALESCE(location, '') AS location FROM customers WHERE user_name ILIKE $1`
	args := []any{keyword}
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " ORDER BY user_name LIMIT 20"

	rows, err := h.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []searchCustomer
	for rows.Next() {
		var c searchCustomer
		if err := rows.Scan(&c.ID, &c.Name, &c.Location); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (h *SearchHandler) searchContracts(ctx context.Context, keyword, org string, scoped bool) ([]searchContract, error) {
	q := `SELECT rc.id::text, c.user_name AS customer_name, rc.status
		FROM retail_contracts rc
		JOIN customers c ON c.id = rc.customer_id
		WHERE c.user_name ILIKE $1`
	args := []any{keyword}
	if scoped {
		q += fmt.Sprintf(" AND rc.org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " ORDER BY c.user_name LIMIT 20"

	rows, err := h.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []searchContract
	for rows.Next() {
		var c searchContract
		if err := rows.Scan(&c.ID, &c.CustomerName, &c.Status); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (h *SearchHandler) searchDocuments(ctx context.Context, keyword, org string, scoped bool) ([]searchDocument, error) {
	q := `SELECT id::text, filename, COALESCE(doc_type, '') AS doc_type
		FROM documents WHERE filename ILIKE $1`
	args := []any{keyword}
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " ORDER BY filename LIMIT 20"

	rows, err := h.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []searchDocument
	for rows.Next() {
		var d searchDocument
		if err := rows.Scan(&d.ID, &d.Filename, &d.DocType); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (h *SearchHandler) searchIntentCustomers(ctx context.Context, keyword, org string, scoped bool) ([]searchIntentCustomer, error) {
	q := `SELECT id::text, customer_name AS company_name, status
		FROM intent_customers WHERE customer_name ILIKE $1`
	args := []any{keyword}
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " ORDER BY customer_name LIMIT 20"

	rows, err := h.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []searchIntentCustomer
	for rows.Next() {
		var ic searchIntentCustomer
		if err := rows.Scan(&ic.ID, &ic.CompanyName, &ic.Status); err != nil {
			return nil, err
		}
		list = append(list, ic)
	}
	return list, rows.Err()
}
