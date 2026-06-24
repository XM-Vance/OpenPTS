// 合同电价（日维度）仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── D6 合同电价日维度 ───────────────

type ContractPriceDaily struct {
	ID               string    `json:"id"`
	ContractID       string    `json:"contract_id"`
	PriceDate        time.Time `json:"price_date"`
	UnitPrice        float64   `json:"unit_price"`
	DailyEnergy      float64   `json:"daily_energy"`
	DailyAmount      float64   `json:"daily_amount"`
	CumulativeEnergy float64   `json:"cumulative_energy"`
	CumulativeAmount float64   `json:"cumulative_amount"`
	CreatedAt        time.Time `json:"created_at"`
}

type ContractPriceRepository struct{ pool *Pool }

func NewContractPriceRepository(pool *Pool) *ContractPriceRepository {
	return &ContractPriceRepository{pool: pool}
}

func (r *ContractPriceRepository) List(ctx context.Context, contractID string, days int) ([]*ContractPriceDaily, error) {
	if days <= 0 || days > 180 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, contract_id::text, price_date, unit_price, daily_energy, daily_amount,
			cumulative_energy, cumulative_amount, created_at
		  FROM contract_price_daily WHERE price_date >= $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	if contractID != "" {
		args = append(args, contractID)
		q += fmt.Sprintf(" AND contract_id = $%d", len(args))
	}
	q += " ORDER BY price_date DESC, contract_id ASC LIMIT 500"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ContractPriceDaily, 0)
	for rows.Next() {
		var p ContractPriceDaily
		if err := rows.Scan(&p.ID, &p.ContractID, &p.PriceDate, &p.UnitPrice,
			&p.DailyEnergy, &p.DailyAmount, &p.CumulativeEnergy, &p.CumulativeAmount,
			&p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// GenerateDemo 为所有 active 合同生成最近 30 天的电价记录。
func (r *ContractPriceRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, purchasing_energy_mwh FROM retail_contracts WHERE status = 'active'`)
	if err != nil {
		return 0, err
	}
	contracts := make([]struct {
		id     string
		energy float64
	}, 0)
	for rows.Next() {
		var id string
		var e float64
		if err := rows.Scan(&id, &e); err != nil {
			rows.Close()
			return 0, err
		}
		contracts = append(contracts, struct {
			id     string
			energy float64
		}{id, e})
	}
	rows.Close()

	cnt := 0
	for _, c := range contracts {
		var cumEnergy, cumAmount float64
		// 估算每日基准电量（合同总量 / 365）
		dailyBase := c.energy / 365
		for i := 29; i >= 0; i-- {
			d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
			price := 380 + rand.Float64()*60 // 380-440 元/MWh
			energy := dailyBase * (0.85 + rand.Float64()*0.3)
			amount := price * energy
			cumEnergy += energy
			cumAmount += amount
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO contract_price_daily
				   (contract_id, price_date, unit_price, daily_energy, daily_amount,
				    cumulative_energy, cumulative_amount, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8::uuid)
				 ON CONFLICT (org_id, contract_id, price_date) DO UPDATE SET
				   unit_price = EXCLUDED.unit_price, daily_energy = EXCLUDED.daily_energy,
				   daily_amount = EXCLUDED.daily_amount,
				   cumulative_energy = EXCLUDED.cumulative_energy,
				   cumulative_amount = EXCLUDED.cumulative_amount`,
				c.id, d, price, energy, amount, cumEnergy, cumAmount, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
