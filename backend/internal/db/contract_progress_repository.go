// 签约进度跟踪。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 签约进度跟踪 ───────────────

type ContractProgress struct {
	ID             string          `json:"id"`
	ContractID     string          `json:"contract_id"`
	CustomerName   string          `json:"customer_name,omitempty"`
	OperatingMonth string          `json:"operating_month"`
	PlannedEnergy  float64         `json:"planned_energy_mwh"`
	ActualEnergy   float64         `json:"actual_energy_mwh"`
	CompletionRate float64         `json:"completion_rate"`
	Status         string          `json:"status"`
	Milestones     json.RawMessage `json:"milestones,omitempty"`
	Note           *string         `json:"note,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type ContractProgressRepository struct{ pool *Pool }

func NewContractProgressRepository(pool *Pool) *ContractProgressRepository {
	return &ContractProgressRepository{pool: pool}
}

func (r *ContractProgressRepository) List(ctx context.Context, month, status string, limit int) ([]*ContractProgress, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{}
	conds := []string{}
	idx := 1
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		conds = append(conds, fmt.Sprintf("p.org_id = $%d::uuid", idx))
		idx++
	}
	if month != "" {
		args = append(args, month)
		conds = append(conds, "p.operating_month = $"+itoaNew(idx))
		idx++
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, "p.status = $"+itoaNew(idx))
		idx++
	}
	q := `SELECT p.id, p.contract_id::text, c.user_name, p.operating_month,
			p.planned_energy_mwh, p.actual_energy_mwh, p.completion_rate,
			p.status, p.milestones, p.note, p.created_at, p.updated_at
		  FROM contract_progress p
		  JOIN retail_contracts rc ON rc.id = p.contract_id
		  JOIN customers c ON c.id = rc.customer_id`
	for i, cond := range conds {
		if i == 0 {
			q += " WHERE " + cond
		} else {
			q += " AND " + cond
		}
	}
	q += " ORDER BY p.operating_month DESC, c.user_name ASC LIMIT " + itoaNew(limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ContractProgress, 0)
	for rows.Next() {
		var p ContractProgress
		if err := rows.Scan(&p.ID, &p.ContractID, &p.CustomerName, &p.OperatingMonth,
			&p.PlannedEnergy, &p.ActualEnergy, &p.CompletionRate,
			&p.Status, &p.Milestones, &p.Note, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

type ContractProgressInput struct {
	ContractID     string
	OperatingMonth string
	PlannedEnergy  float64
	ActualEnergy   float64
	Status         string
	Note           string
}

func (r *ContractProgressRepository) Create(ctx context.Context, in ContractProgressInput) (string, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return "", ErrOrgRequired
	}
	completion := 0.0
	if in.PlannedEnergy > 0 {
		completion = in.ActualEnergy / in.PlannedEnergy * 100
	}
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO contract_progress
		   (contract_id, operating_month, planned_energy_mwh, actual_energy_mwh,
		    completion_rate, status, note, org_id)
		 VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8::uuid)
		 RETURNING id`,
		in.ContractID, in.OperatingMonth, in.PlannedEnergy, in.ActualEnergy,
		completion, in.Status, in.Note, org).Scan(&id)
	return id, err
}

func (r *ContractProgressRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	// 查询同 org 下的活跃合同
	rows, err := r.pool.Query(ctx,
		`SELECT id, purchasing_energy_mwh FROM retail_contracts WHERE status = 'active' AND org_id = $1::uuid`, orgID)
	if err != nil {
		return 0, err
	}
	contracts := []struct {
		id     string
		energy float64
	}{}
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
	statuses := []string{"on_track", "on_track", "on_track", "ahead", "behind", "completed"}
	cnt := 0
	for _, c := range contracts {
		for i := 0; i < 6; i++ {
			ym := time.Now().AddDate(0, -i, 0).Format("2006-01")
			planned := c.energy / 12
			actual := planned * (0.75 + rand.Float64()*0.4)
			rate := actual / planned * 100
			status := statuses[rand.Intn(len(statuses))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO contract_progress
				   (contract_id, operating_month, planned_energy_mwh, actual_energy_mwh,
				    completion_rate, status, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7::uuid)
				 ON CONFLICT (org_id, contract_id, operating_month) DO UPDATE SET
				   planned_energy_mwh = EXCLUDED.planned_energy_mwh,
				   actual_energy_mwh = EXCLUDED.actual_energy_mwh,
				   completion_rate = EXCLUDED.completion_rate,
				   status = EXCLUDED.status`,
				c.id, ym, planned, actual, rate, status, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
