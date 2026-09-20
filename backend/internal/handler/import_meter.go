// 表计数据导入 handler（WP0.3）：营销系统/交易中心导出（CSV/XLSX）→ raw_meter_data，
// 再按 (customer, date) 聚合到 user_load_data + unified_load_curve（多表计 Σ×倍率）。
// 接通「计量数据底座」死表链：raw_meter_data 建表（0005）以来零写入。
//
// 端点：
//	POST /import/meter?dry_run=true|false          multipart file=...（表计日冻结/曲线）
//	POST /load/aggregate-meter?start=&end=         手动触发聚合作业（导入后自动跑当日）
//
// 文件版式（首行表头，列名中文/英文皆可）：
//	客户（名称或 UUID）, 电表编号, 日期, 倍率(可选，默认 1), 曲线...
// 曲线三种形态（自动识别）：
//	1) 第 5 列为逗号串 "v1,v2,...,v96"（或 48 点，自动 ×2 摊到 96）
//	2) 第 5 列起展开 96（或 48）个数值列
//	3) 第 5 列单个数值 = 日总电量 kWh → 均匀铺 96 点（响应中标注 estimated）
package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type MeterImportHandler struct {
	meterRepo *db.MeterRepository
	custRepo  *db.CustomerRepository
}

func NewMeterImportHandler(meterRepo *db.MeterRepository, custRepo *db.CustomerRepository) *MeterImportHandler {
	return &MeterImportHandler{meterRepo: meterRepo, custRepo: custRepo}
}

// meterImportRow 单行导入预览。
type meterImportRow struct {
	LineNo     int      `json:"line_no"`
	Customer   string   `json:"customer"`
	MeterID    string   `json:"meter_id"`
	Date       string   `json:"date"`
	Multiplier float64  `json:"multiplier"`
	Points     int      `json:"points"` // 96（曲线）或 0（日电量铺平）
	TotalKWh   float64  `json:"total_kwh"`
	Estimated  bool     `json:"estimated"` // 日电量均匀铺 96 点
	Errors     []string `json:"errors,omitempty"`
}

type meterImportResult struct {
	DryRun      bool             `json:"dry_run"`
	Total       int              `json:"total"`
	Valid       int              `json:"valid"`
	Invalid     int              `json:"invalid"`
	Inserted    int              `json:"inserted,omitempty"`
	Aggregated  int              `json:"aggregated,omitempty"` // 导入后自动聚合的 (customer,date) 数
	SkippedRaw  int              `json:"skipped_raw,omitempty"`
	Rows        []meterImportRow `json:"rows"`
}

// normalizeCurve96 48 点 → 96（每点重复）；单值 → 均匀 96 点。返回 (curve, estimated)。
func normalizeCurve96(vals []float64) ([]float64, bool) {
	switch len(vals) {
	case 96:
		return vals, false
	case 48:
		out := make([]float64, 0, 96)
		for _, v := range vals {
			out = append(out, v, v)
		}
		return out, false
	case 1:
		v := vals[0] / 24 // 日总电量 kWh ÷ 24h = 平均 kW（曲线语义 kW/15min，聚合 Σcurve/4 还原日电量）
		out := make([]float64, 96)
		for i := range out {
			out[i] = math.Round(v*10000) / 10000
		}
		return out, true
	default:
		return nil, false
	}
}

// parseMeterFile 解析表格行（纯函数）。customerIDs 由调用方传入「名称→UUID」解析结果缓存。
func parseMeterFile(rows [][]string, resolveCustomer func(name string) (string, bool)) (valid []*db.RawMeterRow, previews []meterImportRow, invalid int) {
	if len(rows) < 2 {
		return nil, []meterImportRow{{LineNo: 1, Errors: []string{"无数据行（至少表头 + 1 行）"}}}, 1
	}
	nCols := len(rows[1])
	for ln, row := range rows[1:] {
		p := meterImportRow{LineNo: ln + 2}
		if len(row) < 4 || strings.TrimSpace(row[0]) == "" {
			p.Errors = append(p.Errors, "列数不足或客户列为空")
			previews = append(previews, p)
			invalid++
			continue
		}
		p.Customer = strings.TrimSpace(row[0])
		p.MeterID = strings.TrimSpace(row[1])
		dateStr := strings.TrimSpace(row[2])
		multStr := strings.TrimSpace(row[3])
		if p.MeterID == "" {
			p.Errors = append(p.Errors, "电表编号为空")
		}

		d, derr := time.Parse("2006-01-02", dateStr)
		if derr != nil {
			p.Errors = append(p.Errors, "日期格式错误（应 YYYY-MM-DD）")
		} else {
			p.Date = dateStr
			if d.Year() < 2015 {
				p.Errors = append(p.Errors, "日期过早（< 2015）")
			}
		}

		mult := 1.0
		if multStr != "" {
			v, err := strconv.ParseFloat(multStr, 64)
			if err != nil || v <= 0 {
				p.Errors = append(p.Errors, "倍率应为正数")
			} else {
				mult = v
			}
		}
		p.Multiplier = mult

		// 曲线：第 5 列起
		var vals []float64
		if len(row) >= 5 {
			curveCell := strings.TrimSpace(row[4])
			if strings.Contains(curveCell, ",") {
				// 紧凑串
				for _, part := range strings.Split(curveCell, ",") {
					if v, err := strconv.ParseFloat(strings.TrimSpace(part), 64); err == nil {
						vals = append(vals, v)
					} else {
						p.Errors = append(p.Errors, "曲线含非数字: "+strings.TrimSpace(part))
						break
					}
				}
			} else {
				// 展开列（含单值日电量）
				for i := 4; i < len(row) && i < nCols; i++ {
					cell := strings.TrimSpace(row[i])
					if cell == "" {
						continue
					}
					if v, err := strconv.ParseFloat(cell, 64); err == nil {
						vals = append(vals, v)
					} else {
						p.Errors = append(p.Errors, fmt.Sprintf("第 %d 列非数字", i+1))
						break
					}
				}
			}
		}
		curve, estimated := normalizeCurve96(vals)
		if curve == nil {
			p.Errors = append(p.Errors, "曲线应为 96/48 点或 1 个日电量值")
		} else {
			p.Points = 96
			p.Estimated = estimated
			if estimated {
				p.TotalKWh = math.Round(vals[0]*100) / 100
			} else {
				total := 0.0
				for _, v := range curve {
					if v < 0 {
						p.Errors = append(p.Errors, "曲线含负值")
						break
					}
					total += v
				}
				p.TotalKWh = math.Round(total/4*100) / 100
			}
		}

		// 客户解析：UUID 直用，否则按名称查
		custID := ""
		if _, err := uuid.Parse(p.Customer); err == nil {
			custID = p.Customer
		} else if id, ok := resolveCustomer(p.Customer); ok {
			custID = id
		} else {
			p.Errors = append(p.Errors, "客户不存在: "+p.Customer)
		}

		if len(p.Errors) > 0 {
			previews = append(previews, p)
			invalid++
			continue
		}
		valid = append(valid, &db.RawMeterRow{
			MeterID:    p.MeterID,
			CustomerID: custID,
			Date:       d,
			Multiplier: mult,
			Curve96:    curve,
		})
		p.Errors = nil
		previews = append(previews, p)
	}
	return valid, previews, invalid
}

// Import POST /api/v1/import/meter?dry_run=true|false
func (h *MeterImportHandler) Import(c *gin.Context) {
	dryRun := c.Query("dry_run") != "false" // 默认预览
	if _, err := db.MustScoped(c.Request.Context()); err != nil {
		respondOrgRequired(c, err)
		return
	}
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

	ctx := c.Request.Context()
	cache := map[string]string{}
	resolve := func(name string) (string, bool) {
		if id, ok := cache[name]; ok {
			return id, true
		}
		list, _, err := h.custRepo.List(ctx, db.CustomerListFilter{Keyword: name, Limit: 1})
		if err != nil || len(list) == 0 {
			return "", false
		}
		cache[name] = list[0].ID.String()
		return cache[name], true
	}
	valid, previews, invalid := parseMeterFile(rows, resolve)

	res := meterImportResult{
		DryRun: dryRun, Total: len(previews), Valid: len(valid), Invalid: invalid,
		Rows: previews,
	}
	if !dryRun && len(valid) > 0 {
		inserted, err := h.meterRepo.BulkUpsertRaw(ctx, valid)
		if err != nil {
			if respondOrgRequired(c, err) {
				return
			}
			log.Error().Err(err).Msg("表计数据写入失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		res.Inserted = inserted
		// 自动聚合本次导入覆盖的日期区间
		start, end := valid[0].Date, valid[0].Date
		for _, r := range valid {
			if r.Date.Before(start) {
				start = r.Date
			}
			if r.Date.After(end) {
				end = r.Date
			}
		}
		written, _, err := h.meterRepo.AggregateMeters(ctx, start, end)
		if err != nil {
			log.Warn().Err(err).Msg("表计聚合失败（原始数据已入库，可手动重跑 /load/aggregate-meter）")
		}
		res.Aggregated = written
	}
	c.JSON(http.StatusOK, res)
}

// Aggregate POST /api/v1/load/aggregate-meter?start=YYYY-MM-DD&end=YYYY-MM-DD
// 手动触发聚合作业（缺省近 7 天）。幂等：重跑按 (meter,date) 覆盖重聚合。
func (h *MeterImportHandler) Aggregate(c *gin.Context) {
	end := db.StartOfDay(time.Now())
	start := end.AddDate(0, 0, -6)
	if s := c.Query("start"); s != "" {
		v, err := time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start 格式应为 YYYY-MM-DD"})
			return
		}
		start = v
	}
	if e := c.Query("end"); e != "" {
		v, err := time.Parse("2006-01-02", e)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end 格式应为 YYYY-MM-DD"})
			return
		}
		end = v
	}
	if start.After(end) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start 不能晚于 end"})
		return
	}
	written, skipped, err := h.meterRepo.AggregateMeters(c.Request.Context(), start, end)
	if err != nil {
		if respondOrgRequired(c, err) {
			return
		}
		log.Error().Err(err).Msg("表计聚合失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"start": start.Format("2006-01-02"), "end": end.Format("2006-01-02"),
		"aggregated": written, "skipped": skipped,
		"message": fmt.Sprintf("已聚合 %d 个客户日（跳过 %d 条无归属/损坏表计行）", written, skipped),
	})
}
