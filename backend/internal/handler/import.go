// 数据导入 handler：接收 multipart/form-data，解析 CSV/XLSX，dry-run 预览或正式入库。
//
// 端点：
//   POST /api/v1/import/customers?dry_run=true|false   multipart file=...
//
// CSV/XLSX 格式（第一行表头）：
//   客户名称, 简称, 所在地, 客户经理, 标签（逗号或顿号分隔）, 演示（是/否）
package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"github.com/rs/zerolog/log"
)

type ImportHandler struct {
	customerRepo *db.CustomerRepository
}

func NewImportHandler(cust *db.CustomerRepository) *ImportHandler {
	return &ImportHandler{customerRepo: cust}
}

type importPreviewRow struct {
	LineNo  int               `json:"line_no"`
	OK      bool              `json:"ok"`
	Errors  []string          `json:"errors,omitempty"`
	Payload db.CustomerInput  `json:"payload"`
}

type importResult struct {
	Total     int                `json:"total"`
	Valid     int                `json:"valid"`
	Invalid   int                `json:"invalid"`
	Inserted  int                `json:"inserted,omitempty"`
	DryRun    bool               `json:"dry_run"`
	Rows      []importPreviewRow `json:"rows"`
}

// Customers POST /api/v1/import/customers?dry_run=true|false
func (h *ImportHandler) Customers(c *gin.Context) {
	dryRun := c.Query("dry_run") != "false" // 默认 true（先预览再正式提交）

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段"})
		return
	}
	if fileHeader.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大于 5MB"})
		return
	}

	rows, err := parseTabular(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败: " + err.Error()})
		return
	}
	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无数据行（仅表头）"})
		return
	}

	// 表头映射（按列名）
	header := rows[0]
	idx := map[string]int{}
	for i, col := range header {
		idx[strings.TrimSpace(col)] = i
	}
	required := []string{"客户名称"}
	for _, k := range required {
		if _, ok := idx[k]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必需列: " + k})
			return
		}
	}

	result := importResult{DryRun: dryRun}
	previews := make([]importPreviewRow, 0, len(rows)-1)
	valid := make([]db.CustomerInput, 0)

	for ln, row := range rows[1:] {
		preview := importPreviewRow{LineNo: ln + 2} // 行号从 2 开始（1 是表头）
		name := getCol(row, idx, "客户名称")
		if name == "" {
			preview.Errors = append(preview.Errors, "客户名称为空")
		}
		tags := splitTags(getCol(row, idx, "标签"))
		isDemo := getCol(row, idx, "演示") == "是"

		in := db.CustomerInput{
			UserName:  name,
			ShortName: getCol(row, idx, "简称"),
			Location:  getCol(row, idx, "所在地"),
			Manager:   getCol(row, idx, "客户经理"),
			Tags:      tags,
			IsDemo:    isDemo,
			Accounts:  json.RawMessage("[]"),
		}
		preview.Payload = in
		if len(preview.Errors) == 0 {
			preview.OK = true
			valid = append(valid, in)
		}
		previews = append(previews, preview)
	}

	result.Total = len(previews)
	result.Valid = len(valid)
	result.Invalid = result.Total - result.Valid
	result.Rows = previews

	// 正式提交：批量 INSERT
	if !dryRun && result.Valid > 0 {
		uid := claimsUserID(c)
		n, err := h.bulkInsert(c.Request.Context(), valid, uid)
		if err != nil {
			log.Error().Err(err).Msg("操作失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		result.Inserted = n
	}

	c.JSON(http.StatusOK, result)
}

func (h *ImportHandler) bulkInsert(ctx context.Context, list []db.CustomerInput, uid *uuid.UUID) (int, error) {
	cnt := 0
	for _, in := range list {
		if _, err := h.customerRepo.Create(ctx, in, uid); err != nil {
			// 单条失败不中断（如重名），继续；前端通过日志查
			continue
		}
		cnt++
	}
	return cnt, nil
}

// ─── 工具 ───

func claimsUserID(c *gin.Context) *uuid.UUID {
	cl := authClaims(c)
	if cl == nil {
		return nil
	}
	return &cl.UserID
}

// authClaims 从 gin context 取出 JWT Claims（可能为 nil）。
func authClaims(c *gin.Context) *auth.Claims {
	v, ok := c.Get(auth.ClaimsContextKey)
	if !ok {
		return nil
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		return nil
	}
	return cl
}

func getCol(row []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	for _, t := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '，' || r == '、' || r == ';' || r == '；'
	}) {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// parseTabular 根据 Content-Type / 后缀解析 CSV 或 XLSX。
func parseTabular(fh *multipart.FileHeader) ([][]string, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	name := strings.ToLower(fh.Filename)
	switch {
	case strings.HasSuffix(name, ".xlsx") || strings.HasSuffix(name, ".xlsm"):
		x, err := excelize.OpenReader(f)
		if err != nil {
			return nil, fmt.Errorf("XLSX 解析失败: %w", err)
		}
		defer x.Close()
		sheets := x.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("XLSX 无 sheet")
		}
		return x.GetRows(sheets[0])
	default:
		// CSV / TXT；自动跳过 UTF-8 BOM
		buf, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		// 跳过 UTF-8 BOM (EF BB BF)
		if len(buf) >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
			buf = buf[3:]
		}
		text := string(buf)
		r := csv.NewReader(strings.NewReader(text))
		r.FieldsPerRecord = -1 // 容忍列数不一致
		return r.ReadAll()
	}
}
