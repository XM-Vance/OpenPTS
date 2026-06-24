// 预测准确率。
// 2026-06 自 p1_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ─────────────── V3 预测准确率 ───────────────

type ForecastAccuracy struct {
	ID             string    `json:"id"`
	ForecastTarget string    `json:"forecast_target"`
	ForecastDate   time.Time `json:"forecast_date"`
	PredictedValue float64   `json:"predicted_value"`
	ActualValue    *float64  `json:"actual_value,omitempty"`
	MAPE           *float64  `json:"mape,omitempty"`
	RMSE           *float64  `json:"rmse,omitempty"`
	ModelVersion   *string   `json:"model_version,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AccuracyRepository struct{ pool *Pool }

func NewAccuracyRepository(pool *Pool) *AccuracyRepository { return &AccuracyRepository{pool: pool} }

func (r *AccuracyRepository) List(ctx context.Context, target string, days int) ([]*ForecastAccuracy, error) {
	if days <= 0 || days > 180 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	n := 2
	q := `SELECT id, forecast_target, forecast_date, predicted_value, actual_value,
	             mape, rmse, model_version, created_at
	      FROM forecast_accuracy WHERE forecast_date >= $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	if target != "" {
		args = append(args, target)
		q += fmt.Sprintf(" AND forecast_target = $%d", n)
		n++
	}
	q += " ORDER BY forecast_date DESC, forecast_target"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ForecastAccuracy, 0)
	for rows.Next() {
		var f ForecastAccuracy
		if err := rows.Scan(&f.ID, &f.ForecastTarget, &f.ForecastDate, &f.PredictedValue,
			&f.ActualValue, &f.MAPE, &f.RMSE, &f.ModelVersion, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &f)
	}
	return list, rows.Err()
}

// Summary 各预测目标的近 N 日平均 MAPE。
type AccuracySummary struct {
	Target  string  `json:"target"`
	AvgMAPE float64 `json:"avg_mape"`
	AvgRMSE float64 `json:"avg_rmse"`
	Count   int     `json:"count"`
}

func (r *AccuracyRepository) Summary(ctx context.Context, days int) ([]*AccuracySummary, error) {
	if days <= 0 || days > 180 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	n := 2
	q := `
		SELECT forecast_target, AVG(mape)::float8, AVG(rmse)::float8, COUNT(*)
		FROM forecast_accuracy
		WHERE forecast_date >= $1 AND mape IS NOT NULL`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	q += ` GROUP BY forecast_target
		ORDER BY forecast_target`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*AccuracySummary, 0)
	for rows.Next() {
		var s AccuracySummary
		if err := rows.Scan(&s.Target, &s.AvgMAPE, &s.AvgRMSE, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, rows.Err()
}

func (r *AccuracyRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	cnt := 0
	targets := []struct {
		name, version string
		baseValue     float64
	}{
		{"load", "lgbm-v3.2", 380000},
		{"price", "arima-v2.1", 420},
		{"freq", "lstm-v1.5", 1200},
	}
	for _, t := range targets {
		for i := 0; i < 30; i++ {
			d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
			predicted := t.baseValue * (0.95 + rand.Float64()*0.1)
			actual := predicted * (0.95 + rand.Float64()*0.1)
			mape := math.Abs(actual-predicted) / actual * 100
			rmse := math.Abs(actual - predicted)
			if _, err := r.pool.Exec(ctx, `
				INSERT INTO forecast_accuracy
				(forecast_target, forecast_date, predicted_value, actual_value, mape, rmse, model_version, org_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8::uuid)
				ON CONFLICT (org_id, forecast_target, forecast_date) DO UPDATE SET
				  predicted_value = EXCLUDED.predicted_value,
				  actual_value = EXCLUDED.actual_value,
				  mape = EXCLUDED.mape, rmse = EXCLUDED.rmse,
				  model_version = EXCLUDED.model_version`,
				t.name, d, predicted, actual, mape, rmse, t.version, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
