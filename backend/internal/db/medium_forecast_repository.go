// 中期负荷预测。
// 2026-06 自 p1_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── V2 中期负荷预测 ───────────────

type MediumLoadForecast struct {
	ID             string    `json:"id"`
	ForecastMonth  string    `json:"forecast_month"`
	PredictedMWh   float64   `json:"predicted_mwh"`
	ActualMWh      *float64  `json:"actual_mwh,omitempty"`
	PeakMW         float64   `json:"peak_mw"`
	GrowthRate     *float64  `json:"growth_rate,omitempty"`
	WeatherFactor  *float64  `json:"weather_factor,omitempty"`
	EconomicFactor *float64  `json:"economic_factor,omitempty"`
	Note           *string   `json:"note,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type MediumForecastRepository struct{ pool *Pool }

func NewMediumForecastRepository(pool *Pool) *MediumForecastRepository {
	return &MediumForecastRepository{pool: pool}
}

func (r *MediumForecastRepository) List(ctx context.Context, limit int) ([]*MediumLoadForecast, error) {
	if limit <= 0 || limit > 36 {
		limit = 18
	}
	args := []any{limit}
	q := `
		SELECT id, forecast_month, predicted_mwh, actual_mwh, peak_mw, growth_rate,
		       weather_factor, economic_factor, note, created_at
		FROM medium_load_forecast`
	n := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", n)
		n++
	}
	q += " ORDER BY forecast_month DESC LIMIT $1"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MediumLoadForecast, 0, limit)
	for rows.Next() {
		var m MediumLoadForecast
		if err := rows.Scan(&m.ID, &m.ForecastMonth, &m.PredictedMWh, &m.ActualMWh,
			&m.PeakMW, &m.GrowthRate, &m.WeatherFactor, &m.EconomicFactor,
			&m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *MediumForecastRepository) GenerateDemo(ctx context.Context) (int, error) {
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
	// 过去 6 月（含实测） + 未来 12 月（仅预测）
	for i := 6; i >= -12; i-- {
		t := time.Now().AddDate(0, -i, 0)
		ym := t.Format("2006-01")
		predicted := 350000.0 + rand.Float64()*100000 // 350-450 GWh
		var actual *float64
		if i >= 0 { // 过去
			a := predicted * (0.95 + rand.Float64()*0.1)
			actual = &a
		}
		peak := 18000.0 + rand.Float64()*4000
		growth := 5 + rand.Float64()*5
		weather := 0.9 + rand.Float64()*0.2
		econ := 0.95 + rand.Float64()*0.1
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO medium_load_forecast
			(forecast_month, predicted_mwh, actual_mwh, peak_mw, growth_rate, weather_factor, economic_factor, org_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8::uuid)
			ON CONFLICT (org_id, forecast_month) DO UPDATE SET
			  predicted_mwh = EXCLUDED.predicted_mwh,
			  actual_mwh = EXCLUDED.actual_mwh,
			  peak_mw = EXCLUDED.peak_mw,
			  growth_rate = EXCLUDED.growth_rate,
			  weather_factor = EXCLUDED.weather_factor,
			  economic_factor = EXCLUDED.economic_factor`,
			ym, predicted, actual, peak, growth, weather, econ, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
