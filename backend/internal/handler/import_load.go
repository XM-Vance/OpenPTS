// 电表负荷数据导入：CSV/XLSX → user_load_data。
//
// 格式 A（紧凑）：客户名称, 日期, 96 点（逗号分隔，整列一格）
// 格式 B（展开）：客户名称, 日期, P1, P2, P3, ..., P96
//
// 自动按列数判断。日期 YYYY-MM-DD；客户按 user_name 模糊匹配第一条。
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoadImportHandler struct {
	loadRepo *db.LoadRepository
	custRepo *db.CustomerRepository
}

type validLoadRow struct {
	custID uuid.UUID
	date   time.Time
	curve  []float64
	total  float64
}

func NewLoadImportHandler(loadRepo *db.LoadRepository, custRepo *db.CustomerRepository) *LoadImportHandler {
	return &LoadImportHandler{loadRepo: loadRepo, custRepo: custRepo}
}

type loadImportRow struct {
	LineNo       int     `json:"line_no"`
	OK           bool    `json:"ok"`
	Errors       []string `json:"errors,omitempty"`
	CustomerName string  `json:"customer_name"`
	Date         string  `json:"date"`
	Curve96Len   int     `json:"curve_96_len"`
}

type loadImportResult struct {
	Total    int             `json:"total"`
	Valid    int             `json:"valid"`
	Invalid  int             `json:"invalid"`
	Inserted int             `json:"inserted,omitempty"`
	DryRun   bool            `json:"dry_run"`
	Rows     []loadImportRow `json:"rows"`
}

// Customers POST /api/v1/import/load?dry_run=true|false
func (h *LoadImportHandler) Load(c *gin.Context) {
	dryRun := c.Query("dry_run") != "false"

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段"})
		return
	}
	if fh.Size > 20<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大于 20MB"})
		return
	}
	rows, err := parseTabular(fh)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败: " + err.Error()})
		return
	}
	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无数据行"})
		return
	}

	header := rows[0]
	if len(header) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "格式错误：至少需要客户名称/日期/曲线 3 列"})
		return
	}
	// 检测格式：列数 > 3 视为展开格式
	compact := len(header) == 3

	previews := make([]loadImportRow, 0, len(rows)-1)
	validRows := make([]validLoadRow, 0)

	for ln, row := range rows[1:] {
		preview := loadImportRow{LineNo: ln + 2}
		if len(row) < 3 {
			preview.Errors = append(preview.Errors, "列数不足")
			previews = append(previews, preview)
			continue
		}
		name := strings.TrimSpace(row[0])
		dateStr := strings.TrimSpace(row[1])
		preview.CustomerName = name
		preview.Date = dateStr

		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			preview.Errors = append(preview.Errors, "日期格式错误（应 YYYY-MM-DD）")
		}

		// 客户匹配
		var custID uuid.UUID
		if name == "" {
			preview.Errors = append(preview.Errors, "客户名称为空")
		} else {
			list, _, err := h.custRepo.List(c.Request.Context(), db.CustomerListFilter{Keyword: name, Limit: 1})
			if err != nil || len(list) == 0 {
				preview.Errors = append(preview.Errors, "客户不存在: "+name)
			} else {
				custID = list[0].ID
			}
		}

		// 曲线
		var curve []float64
		if compact {
			parts := strings.Split(row[2], ",")
			curve = make([]float64, 0, 96)
			for _, p := range parts {
				v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
				if err != nil {
					preview.Errors = append(preview.Errors, "曲线包含非数字: "+strings.TrimSpace(p))
					break
				}
				curve = append(curve, v)
			}
		} else {
			curve = make([]float64, 0, 96)
			for i := 2; i < len(row) && i < 2+96; i++ {
				v, err := strconv.ParseFloat(strings.TrimSpace(row[i]), 64)
				if err != nil {
					preview.Errors = append(preview.Errors, "第 "+strconv.Itoa(i+1)+" 列非数字")
					break
				}
				curve = append(curve, v)
			}
		}
		if len(curve) != 96 {
			preview.Errors = append(preview.Errors, "曲线点数 "+strconv.Itoa(len(curve))+" ≠ 96")
		}
		preview.Curve96Len = len(curve)

		if len(preview.Errors) == 0 {
			preview.OK = true
			total := 0.0
			for _, v := range curve {
				total += v * 0.25 // 15min interval
			}
			validRows = append(validRows, validLoadRow{custID, d, curve, total})
		}
		previews = append(previews, preview)
	}

	result := loadImportResult{
		Total:   len(previews),
		Valid:   len(validRows),
		Invalid: len(previews) - len(validRows),
		DryRun:  dryRun,
		Rows:    previews,
	}

	if !dryRun && result.Valid > 0 {
		ctx := c.Request.Context()
		n := h.bulkInsert(ctx, validRows)
		result.Inserted = n
	}
	c.JSON(http.StatusOK, result)
}

func (h *LoadImportHandler) bulkInsert(ctx context.Context, rows []validLoadRow) int {
	cnt := 0
	for _, r := range rows {
		if err := h.loadRepo.UpsertCurve(ctx, r.custID, r.date, r.curve, r.total); err == nil {
			cnt++
		}
	}
	return cnt
}
