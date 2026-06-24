// 储能申报策略仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── E6 储能申报策略 ───────────────

type StorageDeclaration struct {
	ID              string    `json:"id"`
	StationID       string    `json:"station_id"`
	StationName     string    `json:"station_name,omitempty"`
	DeclaredDate    time.Time `json:"declared_date"`
	ChargeMW        []float64 `json:"charge_mw"`
	DischargeMW     []float64 `json:"discharge_mw"`
	ExpectedRevenue float64   `json:"expected_revenue"`
	StrategyNote    *string   `json:"strategy_note,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type StorageDeclarationRepository struct{ pool *Pool }

func NewStorageDeclarationRepository(pool *Pool) *StorageDeclarationRepository {
	return &StorageDeclarationRepository{pool: pool}
}

func (r *StorageDeclarationRepository) List(ctx context.Context, stationID string, days int) ([]*StorageDeclaration, error) {
	if days <= 0 || days > 60 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	argN := 1
	q := `SELECT s.id, s.station_id::text, st.name, s.declared_date,
			s.charge_mw, s.discharge_mw, s.expected_revenue, s.strategy_note, s.created_at
		  FROM storage_declaration s JOIN storage_stations st ON st.id = s.station_id
		  WHERE s.declared_date >= $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		argN++
		args = append(args, org)
		q += fmt.Sprintf(" AND s.org_id = $%d::uuid", argN)
	}
	if stationID != "" {
		argN++
		args = append(args, stationID)
		q += fmt.Sprintf(" AND s.station_id = $%d", argN)
	}
	q += " ORDER BY s.declared_date DESC, st.name ASC LIMIT 100"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*StorageDeclaration, 0)
	for rows.Next() {
		var d StorageDeclaration
		if err := rows.Scan(&d.ID, &d.StationID, &d.StationName, &d.DeclaredDate,
			&d.ChargeMW, &d.DischargeMW, &d.ExpectedRevenue, &d.StrategyNote,
			&d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *StorageDeclarationRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	rows, err := r.pool.Query(ctx, `SELECT id, max_power_mw::float8 FROM storage_stations WHERE status = 'active' AND org_id = $1::uuid`, orgID)
	if err != nil {
		return 0, err
	}
	stations := make([]struct {
		id  string
		cap float64
	}, 0)
	for rows.Next() {
		var id string
		var cap float64
		if err := rows.Scan(&id, &cap); err != nil {
			rows.Close()
			return 0, err
		}
		stations = append(stations, struct {
			id  string
			cap float64
		}{id, cap})
	}
	rows.Close()

	notes := []string{"低谷充电、峰段放电", "保守策略", "激进套利", "辅助调频备用"}
	cnt := 0
	for _, st := range stations {
		for i := 0; i < 7; i++ {
			d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
			charge := make([]float64, 96)
			discharge := make([]float64, 96)
			for p := 0; p < 96; p++ {
				hour := p / 4
				if hour >= 0 && hour < 7 {
					charge[p] = st.cap * (0.7 + rand.Float64()*0.3)
				} else if hour >= 10 && hour < 12 {
					discharge[p] = st.cap * (0.6 + rand.Float64()*0.4)
				} else if hour >= 18 && hour < 21 {
					discharge[p] = st.cap * (0.7 + rand.Float64()*0.3)
				} else if hour >= 22 || hour < 24 && hour >= 22 {
					charge[p] = st.cap * 0.3
				}
			}
			rev := 0.0
			for p := 0; p < 96; p++ {
				rev += discharge[p]*0.25*(500+rand.Float64()*200) -
					charge[p]*0.25*(200+rand.Float64()*80)
			}
			note := notes[rand.Intn(len(notes))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO storage_declaration
				   (station_id, declared_date, charge_mw, discharge_mw, expected_revenue, strategy_note, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7::uuid)
				 ON CONFLICT (org_id, station_id, declared_date) DO UPDATE SET
				   charge_mw = EXCLUDED.charge_mw,
				   discharge_mw = EXCLUDED.discharge_mw,
				   expected_revenue = EXCLUDED.expected_revenue,
				   strategy_note = EXCLUDED.strategy_note`,
				st.id, d, charge, discharge, rev, note, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
