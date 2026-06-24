// 机制电量。
// 2026-06 自 p1_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── V4 机制电量 ───────────────

type MechanismEnergy struct {
	ID             string    `json:"id"`
	OperatingMonth string    `json:"operating_month"`
	VoltageLevel   string    `json:"voltage_level"`
	PlannedMWh     float64   `json:"planned_mwh"`
	ActualMWh      float64   `json:"actual_mwh"`
	DeviationMWh   float64   `json:"deviation_mwh"`
	ContractPrice  float64   `json:"contract_price"`
	SettleAmount   float64   `json:"settle_amount"`
	CreatedAt      time.Time `json:"created_at"`
}

type MechanismEnergyRepository struct{ pool *Pool }

func NewMechanismEnergyRepository(pool *Pool) *MechanismEnergyRepository {
	return &MechanismEnergyRepository{pool: pool}
}

func (r *MechanismEnergyRepository) List(ctx context.Context, voltage string, months int) ([]*MechanismEnergy, error) {
	if months <= 0 || months > 36 {
		months = 12
	}
	where := make([]string, 0, 3)
	args := make([]any, 0, 3)
	n := 1
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", n))
		n++
	}
	if voltage != "" {
		args = append(args, voltage)
		where = append(where, fmt.Sprintf("voltage_level = $%d", n))
		n++
	}
	q := `SELECT id, operating_month, voltage_level, planned_mwh, actual_mwh, deviation_mwh,
	             contract_price, settle_amount, created_at
	      FROM mechanism_energy_plan`
	if len(where) > 0 {
		q += " WHERE " + joinWhere(where)
	}
	q += " ORDER BY operating_month DESC, voltage_level LIMIT 500"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MechanismEnergy, 0)
	for rows.Next() {
		var m MechanismEnergy
		if err := rows.Scan(&m.ID, &m.OperatingMonth, &m.VoltageLevel,
			&m.PlannedMWh, &m.ActualMWh, &m.DeviationMWh, &m.ContractPrice,
			&m.SettleAmount, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *MechanismEnergyRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	volts := []string{"380V", "10kV", "35kV", "110kV"}
	baseplan := map[string]float64{"380V": 15000, "10kV": 25000, "35kV": 18000, "110kV": 35000}
	basePrice := map[string]float64{"380V": 450, "10kV": 420, "35kV": 400, "110kV": 380}
	cnt := 0
	for i := 0; i < 12; i++ {
		t := time.Now().AddDate(0, -i, 0)
		ym := t.Format("2006-01")
		for _, v := range volts {
			planned := baseplan[v] * (0.9 + rand.Float64()*0.2)
			actual := planned * (0.93 + rand.Float64()*0.14)
			dev := actual - planned
			price := basePrice[v]
			settle := actual * price
			if _, err := r.pool.Exec(ctx, `
				INSERT INTO mechanism_energy_plan
				(operating_month, voltage_level, planned_mwh, actual_mwh, deviation_mwh,
				 contract_price, settle_amount, org_id)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8::uuid)
				ON CONFLICT (org_id, operating_month, voltage_level) DO UPDATE SET
				  planned_mwh = EXCLUDED.planned_mwh,
				  actual_mwh = EXCLUDED.actual_mwh,
				  deviation_mwh = EXCLUDED.deviation_mwh,
				  contract_price = EXCLUDED.contract_price,
				  settle_amount = EXCLUDED.settle_amount`,
				ym, v, planned, actual, dev, price, settle, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
