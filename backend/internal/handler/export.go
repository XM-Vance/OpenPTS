// 数据导出 handler：客户档案、月度结算、合同电价 → CSV 或 XLSX。
// 用法：GET /api/v1/export/{resource}?format=csv|xlsx
package handler

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"github.com/rs/zerolog/log"
)

type ExportHandler struct {
	customerRepo      *db.CustomerRepository
	monthlyRepo       *db.MonthlySettlementRepository
	contractPriceRepo *db.ContractPriceRepository
}

func NewExportHandler(
	cust *db.CustomerRepository,
	monthly *db.MonthlySettlementRepository,
	cp *db.ContractPriceRepository,
) *ExportHandler {
	return &ExportHandler{customerRepo: cust, monthlyRepo: monthly, contractPriceRepo: cp}
}

// Resource GET /api/v1/export/:resource?format=csv|xlsx
func (h *ExportHandler) Resource(c *gin.Context) {
	resource := c.Param("resource")
	format := c.DefaultQuery("format", "csv")
	if format != "csv" && format != "xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format 必须是 csv 或 xlsx"})
		return
	}

	headers, rows, err := h.fetch(c.Request.Context(), resource)
	if err != nil {
		log.Error().Err(err).Msg("操作失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	stamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.%s", resource, stamp, format)

	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		// 写 UTF-8 BOM 让 Excel 不乱码
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
		w := csv.NewWriter(c.Writer)
		w.Write(headers)
		for _, r := range rows {
			w.Write(r)
		}
		w.Flush()
	case "xlsx":
		buf, err := writeXLSX(resource, headers, rows)
		if err != nil {
			log.Error().Err(err).Msg("操作失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		io.Copy(c.Writer, buf)
	}
}

// fetch 根据 resource 拉取相应数据并转成 [headers, rows] 形态。
func (h *ExportHandler) fetch(ctx context.Context, resource string) ([]string, [][]string, error) {
	switch resource {
	case "customers":
		list, _, err := h.customerRepo.List(ctx, db.CustomerListFilter{Limit: 1000})
		if err != nil {
			return nil, nil, err
		}
		headers := []string{"客户名称", "简称", "所在地", "客户经理", "标签", "演示"}
		rows := make([][]string, 0, len(list))
		for _, cu := range list {
			rows = append(rows, []string{
				cu.UserName,
				strDeref(cu.ShortName),
				strDeref(cu.Location),
				strDeref(cu.Manager),
				strings.Join(cu.Tags, "、"),
				boolStr(cu.IsDemo),
			})
		}
		return headers, rows, nil

	case "monthly-settlement":
		list, err := h.monthlyRepo.List(ctx, 60)
		if err != nil {
			return nil, nil, err
		}
		headers := []string{"月份", "结算电量 MWh", "电能量电费", "容量电费", "辅助服务", "政策补贴", "合计", "版本"}
		rows := make([][]string, 0, len(list))
		for _, m := range list {
			rows = append(rows, []string{
				m.OperatingMonth,
				fmtFloat(m.SettledEnergyMWh),
				fmtFloat(m.EnergyFee),
				fmtFloat(m.CapacityFee),
				fmtFloat(m.AncillaryFee),
				fmtFloat(m.PolicySubsidy),
				fmtFloat(m.TotalFee),
				m.Version,
			})
		}
		return headers, rows, nil

	case "contract-price":
		list, err := h.contractPriceRepo.List(ctx, "", 90)
		if err != nil {
			return nil, nil, err
		}
		headers := []string{"合同 ID", "日期", "单价 (元/MWh)", "日电量 MWh", "日金额", "累计电量", "累计金额"}
		rows := make([][]string, 0, len(list))
		for _, p := range list {
			rows = append(rows, []string{
				p.ContractID,
				p.PriceDate.Format("2006-01-02"),
				fmtFloat(p.UnitPrice),
				fmtFloat(p.DailyEnergy),
				fmtFloat(p.DailyAmount),
				fmtFloat(p.CumulativeEnergy),
				fmtFloat(p.CumulativeAmount),
			})
		}
		return headers, rows, nil
	}
	return nil, nil, fmt.Errorf("不支持的导出资源: %s", resource)
}

func writeXLSX(sheet string, headers []string, rows [][]string) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()
	defaultSheet := "Sheet1"
	f.SetSheetName(defaultSheet, sheet)

	// 表头加粗
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E5E7EB"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	for r, row := range rows {
		for i, v := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, r+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
	// 自适应列宽（粗略：列名长度 * 1.5 或 15 兜底）
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		width := float64(len(h))*1.5 + 4
		if width < 12 {
			width = 12
		}
		if width > 30 {
			width = 30
		}
		f.SetColWidth(sheet, col, col, width)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

// ─── 工具 ───
func strDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func boolStr(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
func fmtFloat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
