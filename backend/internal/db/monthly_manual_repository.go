// 月度手工数据仓储。
// 2026-06 自 v1clone_f_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── F4 月度手工数据 ───────────────

type MonthlyManualItem struct {
	ID             string    `json:"id"`
	OperatingMonth string    `json:"operating_month"`
	Category       string    `json:"category"`
	ItemName       string    `json:"item_name"`
	Value          float64   `json:"value"`
	Unit           string    `json:"unit"`
	Source         *string   `json:"source,omitempty"`
	Note           *string   `json:"note,omitempty"`
	CreatedBy      *string   `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MonthlyManualRepository struct{ pool *Pool }

func NewMonthlyManualRepository(pool *Pool) *MonthlyManualRepository {
	return &MonthlyManualRepository{pool: pool}
}

func (r *MonthlyManualRepository) List(ctx context.Context, month, category string) ([]*MonthlyManualItem, error) {
	args := []any{}
	conds := []string{}
	if month != "" {
		args = append(args, month)
		conds = append(conds, "operating_month = $1")
	}
	if category != "" {
		args = append(args, category)
		idx := "$1"
		if len(args) == 2 {
			idx = "$2"
		}
		conds = append(conds, "category = "+idx)
	}
	q := `SELECT id, operating_month, category, item_name, value, unit,
			source, note, created_by, created_at, updated_at
		  FROM monthly_manual_data`
	for i, c := range conds {
		if i == 0 {
			q += " WHERE " + c
		} else {
			q += " AND " + c
		}
	}
	q += " ORDER BY operating_month DESC, category ASC, item_name ASC LIMIT 200"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MonthlyManualItem, 0)
	for rows.Next() {
		var i MonthlyManualItem
		if err := rows.Scan(&i.ID, &i.OperatingMonth, &i.Category, &i.ItemName,
			&i.Value, &i.Unit, &i.Source, &i.Note, &i.CreatedBy,
			&i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &i)
	}
	return list, rows.Err()
}

type ManualItemInput struct {
	OperatingMonth string
	Category       string
	ItemName       string
	Value          float64
	Unit           string
	Source         string
	Note           string
	CreatedBy      string
}

func (r *MonthlyManualRepository) Create(ctx context.Context, in ManualItemInput) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO monthly_manual_data
		   (operating_month, category, item_name, value, unit, source, note, created_by)
		 VALUES ($1,$2,$3,$4,$5,
			NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''))
		 RETURNING id`,
		in.OperatingMonth, in.Category, in.ItemName, in.Value, in.Unit,
		in.Source, in.Note, in.CreatedBy).Scan(&id)
	return id, err
}

func (r *MonthlyManualRepository) GenerateDemo(ctx context.Context) (int, error) {
	items := []struct {
		cat, name, unit string
	}{
		{"收入", "零售合同收入", "元"},
		{"收入", "辅助服务收益", "元"},
		{"收入", "调频收益", "元"},
		{"成本", "批发采购成本", "元"},
		{"成本", "网损费用", "元"},
		{"成本", "辅助服务分摊", "元"},
		{"偏差", "购售偏差电量", "MWh"},
		{"偏差", "偏差考核电费", "元"},
		{"其他", "客户违约金", "元"},
	}
	bases := map[string]float64{
		"零售合同收入": 80000000,
		"辅助服务收益": 1500000,
		"调频收益":   800000,
		"批发采购成本": 75000000,
		"网损费用":   2000000,
		"辅助服务分摊": 600000,
		"购售偏差电量": 1500,
		"偏差考核电费": 90000,
		"客户违约金":  120000,
	}
	cnt := 0
	for m := 0; m < 6; m++ {
		ym := time.Now().AddDate(0, -m, 0).Format("2006-01")
		for _, it := range items {
			base := bases[it.name]
			v := base * (0.85 + rand.Float64()*0.3)
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO monthly_manual_data
				   (operating_month, category, item_name, value, unit, source, created_by)
				 VALUES ($1,$2,$3,$4,$5,'系统初始化','admin')
				 ON CONFLICT DO NOTHING`,
				ym, it.cat, it.name, v, it.unit); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
