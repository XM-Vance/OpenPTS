// 月度结算仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────── D1 月度结算 ───────────────

type MonthlySettlement struct {
	ID               string          `json:"id"`
	OperatingMonth   string          `json:"operating_month"`
	SettledEnergyMWh float64         `json:"settled_energy_mwh"`
	EnergyFee        decimal.Decimal `json:"energy_fee"`     // P4: numeric(18,4)
	CapacityFee      decimal.Decimal `json:"capacity_fee"`   // P4: numeric(18,4)
	AncillaryFee     decimal.Decimal `json:"ancillary_fee"`  // P4: numeric(18,4)
	PolicySubsidy    decimal.Decimal `json:"policy_subsidy"` // P4: numeric(18,4)
	TotalFee         decimal.Decimal `json:"total_fee"`      // P4: = energy+capacity+ancillary 精确
	Version          string          `json:"version"`
	CreatedAt        time.Time       `json:"created_at"`
}

type MonthlySettlementRepository struct{ pool *Pool }

func NewMonthlySettlementRepository(pool *Pool) *MonthlySettlementRepository {
	return &MonthlySettlementRepository{pool: pool}
}

func (r *MonthlySettlementRepository) List(ctx context.Context, limit int) ([]*MonthlySettlement, error) {
	if limit <= 0 || limit > 60 {
		limit = 12
	}
	args := []any{}
	q := `SELECT id, operating_month, settled_energy_mwh, energy_fee, capacity_fee,
			ancillary_fee, policy_subsidy, total_fee, version, created_at
		 FROM batch_monthly_settlement WHERE 1=1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY operating_month DESC LIMIT $" + itoaNew(len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MonthlySettlement, 0, limit)
	for rows.Next() {
		var m MonthlySettlement
		if err := rows.Scan(&m.ID, &m.OperatingMonth, &m.SettledEnergyMWh,
			&m.EnergyFee, &m.CapacityFee, &m.AncillaryFee, &m.PolicySubsidy,
			&m.TotalFee, &m.Version, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

// Upsert 写入/更新某月结算（文档「确认入库」等真实数据来源使用）。
// 写操作要求具体活跃组织；同省同月覆盖更新。
func (r *MonthlySettlementRepository) Upsert(ctx context.Context, operatingMonth string,
	energyMWh, energyFee, capacityFee, ancillaryFee, subsidy, totalFee float64, version string) error {
	org, err := MustScoped(ctx)
	if err != nil {
		return err
	}
	if version == "" {
		version = "IMPORTED"
	}
	// P4: 金额列为 numeric，需以 decimal 写入（pgx 已注册 decimal 编解码，float 不再直接编码）。
	_, err = r.pool.Exec(ctx,
		`INSERT INTO batch_monthly_settlement
		   (operating_month, settled_energy_mwh, energy_fee, capacity_fee, ancillary_fee,
		    policy_subsidy, total_fee, version, org_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid)
		 ON CONFLICT (org_id, operating_month) DO UPDATE SET
		   settled_energy_mwh = EXCLUDED.settled_energy_mwh,
		   energy_fee = EXCLUDED.energy_fee, capacity_fee = EXCLUDED.capacity_fee,
		   ancillary_fee = EXCLUDED.ancillary_fee, policy_subsidy = EXCLUDED.policy_subsidy,
		   total_fee = EXCLUDED.total_fee, version = EXCLUDED.version`,
		operatingMonth, energyMWh,
		decimal.NewFromFloat(energyFee), decimal.NewFromFloat(capacityFee),
		decimal.NewFromFloat(ancillaryFee), decimal.NewFromFloat(subsidy),
		decimal.NewFromFloat(totalFee), version, org)
	return err
}

func (r *MonthlySettlementRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	cnt := 0
	for i := 0; i < 12; i++ {
		ym := monthsAgoYM(i)
		energy := 80000.0 + rand.Float64()*40000 // 80k~120k MWh（float）
		// P4: 各项费用 decimal（分项舍 4 位再汇总），total_fee = energy+capacity+ancillary 精确。
		energyFee := decimal.NewFromFloat(energy).Mul(decimal.NewFromFloat(320 + rand.Float64()*60)).Round(4)
		capacity := decimal.NewFromFloat(energy).Mul(decimal.NewFromInt(35)).Round(4)
		ancillary := decimal.NewFromFloat(energy).Mul(decimal.NewFromInt(8)).Round(4)
		subsidy := energyFee.Mul(decimal.NewFromFloat(0.02)).Round(4)
		total := energyFee.Add(capacity).Add(ancillary)
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO batch_monthly_settlement
			   (operating_month, settled_energy_mwh, energy_fee, capacity_fee, ancillary_fee,
			    policy_subsidy, total_fee, version, org_id)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,'PRELIMINARY',$8::uuid)
			 ON CONFLICT (org_id, operating_month) DO UPDATE SET
			   settled_energy_mwh = EXCLUDED.settled_energy_mwh,
			   energy_fee = EXCLUDED.energy_fee, capacity_fee = EXCLUDED.capacity_fee,
			   ancillary_fee = EXCLUDED.ancillary_fee, policy_subsidy = EXCLUDED.policy_subsidy,
			   total_fee = EXCLUDED.total_fee`,
			ym, energy, energyFee, capacity, ancillary, subsidy, total, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
