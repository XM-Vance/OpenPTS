// 负荷数据诊断仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math"
	"time"
)

// ─────────────── E3 负荷数据诊断 ───────────────

type LoadDiagnosisItem struct {
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Date         time.Time `json:"date"`
	QualityFlag  string    `json:"quality_flag"`
	NullCount    int       `json:"null_count"`
	ZeroCount    int       `json:"zero_count"`
	JumpCount    int       `json:"jump_count"`
	Issue        string    `json:"issue"`
}

type LoadDiagnosisRepository struct{ pool *Pool }

func NewLoadDiagnosisRepository(pool *Pool) *LoadDiagnosisRepository {
	return &LoadDiagnosisRepository{pool: pool}
}

// List 列出最近 N 天有异常的负荷条目（quality_flag != 'ok' 或包含异常值）。
func (r *LoadDiagnosisRepository) List(ctx context.Context, days int) ([]*LoadDiagnosisItem, error) {
	if days <= 0 || days > 60 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := r.pool.Query(ctx, `
		SELECT u.customer_id::text, c.user_name, u.date, u.quality_flag, u.curve_96
		FROM user_load_data u JOIN customers c ON c.id = u.customer_id
		WHERE u.date >= $1::date
		ORDER BY u.date DESC, c.user_name ASC
		LIMIT 200`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*LoadDiagnosisItem, 0)
	for rows.Next() {
		var it LoadDiagnosisItem
		var curve []float64
		if err := rows.Scan(&it.CustomerID, &it.CustomerName, &it.Date, &it.QualityFlag, &curve); err != nil {
			return nil, err
		}
		// 扫描 curve_96 计 null/zero/jump
		var prev float64 = -1
		for _, v := range curve {
			if math.IsNaN(v) {
				it.NullCount++
				continue
			}
			if v == 0 {
				it.ZeroCount++
			}
			if prev >= 0 && prev > 0 && math.Abs(v-prev)/prev > 0.6 {
				it.JumpCount++
			}
			prev = v
		}
		// 判定 issue
		switch {
		case it.NullCount > 0:
			it.Issue = "存在 NaN/缺失点"
		case it.ZeroCount > 20:
			it.Issue = "零点过多"
		case it.JumpCount > 5:
			it.Issue = "曲线跳变多次"
		case it.QualityFlag != "ok":
			it.Issue = "质量标记: " + it.QualityFlag
		default:
			it.Issue = "正常"
		}
		// 仅返回有问题或质量标记非 ok 的
		if it.Issue != "正常" {
			list = append(list, &it)
		}
	}
	return list, rows.Err()
}
