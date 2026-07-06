// 电网代理价仓储。
// 2026-06 自 v1clone_e_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

// ─────────────── E5 电网代理价 ───────────────

type GridAgencyPrice struct {
	ID             string    `json:"id"`
	OrgID          string    `json:"org_id,omitempty"`
	OperatingMonth string    `json:"operating_month"`
	VoltageLevel   string    `json:"voltage_level"`
	AvgPrice       float64   `json:"avg_price"`
	PeakPrice      float64   `json:"peak_price"`
	FlatPrice      float64   `json:"flat_price"`
	ValleyPrice    float64   `json:"valley_price"`
	CreatedAt      time.Time `json:"created_at"`
}

var ErrGridAgencyNotFound = errors.New("电网代理价记录不存在")

// GridAgencyInput 创建/更新入参。
type GridAgencyInput struct {
	OperatingMonth string
	VoltageLevel   string
	AvgPrice       float64
	PeakPrice      float64
	FlatPrice      float64
	ValleyPrice    float64
}

type GridAgencyRepository struct{ pool *Pool }

func NewGridAgencyRepository(pool *Pool) *GridAgencyRepository {
	return &GridAgencyRepository{pool: pool}
}

const gridAgencyColumns = "id, org_id, operating_month, voltage_level, avg_price, peak_price, flat_price, valley_price, created_at"

func (r *GridAgencyRepository) scan(row pgx.Row) (*GridAgencyPrice, error) {
	var g GridAgencyPrice
	err := row.Scan(&g.ID, &g.OrgID, &g.OperatingMonth, &g.VoltageLevel, &g.AvgPrice,
		&g.PeakPrice, &g.FlatPrice, &g.ValleyPrice, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGridAgencyNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *GridAgencyRepository) List(ctx context.Context, voltage string, months int) ([]*GridAgencyPrice, error) {
	if months <= 0 || months > 36 {
		months = 12
	}
	q := `SELECT ` + gridAgencyColumns + ` FROM grid_agency_price`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	if voltage != "" {
		if scoped {
			args = append(args, voltage)
			q += fmt.Sprintf(" AND voltage_level = $%d", len(args))
		} else {
			args = append(args, voltage)
			q += fmt.Sprintf(" WHERE voltage_level = $%d", len(args))
		}
	}
	q += " ORDER BY operating_month DESC, voltage_level ASC LIMIT 500"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*GridAgencyPrice, 0)
	for rows.Next() {
		g, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

func (r *GridAgencyRepository) Create(ctx context.Context, in GridAgencyInput) (*GridAgencyPrice, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	q := `INSERT INTO grid_agency_price
		(org_id, operating_month, voltage_level, avg_price, peak_price, flat_price, valley_price)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
		RETURNING ` + gridAgencyColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		org, in.OperatingMonth, in.VoltageLevel, in.AvgPrice,
		in.PeakPrice, in.FlatPrice, in.ValleyPrice))
}

func (r *GridAgencyRepository) Update(ctx context.Context, id string, in GridAgencyInput) (*GridAgencyPrice, error) {
	q := `UPDATE grid_agency_price SET
		operating_month = $2, voltage_level = $3, avg_price = $4,
		peak_price = $5, flat_price = $6, valley_price = $7
		WHERE id = $1::uuid`
	args := []any{id, in.OperatingMonth, in.VoltageLevel, in.AvgPrice,
		in.PeakPrice, in.FlatPrice, in.ValleyPrice}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += ` RETURNING ` + gridAgencyColumns
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *GridAgencyRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM grid_agency_price WHERE id = $1::uuid`
	args := []any{id}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrGridAgencyNotFound
	}
	return nil
}

func (r *GridAgencyRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	volts := []string{"380V", "10kV", "35kV", "110kV"}
	baseAvg := map[string]float64{"380V": 720, "10kV": 680, "35kV": 640, "110kV": 600}
	cnt := 0
	for i := 0; i < 12; i++ {
		ym := monthsAgoYM(i)
		for _, v := range volts {
			avg := baseAvg[v] + rand.Float64()*40 - 20
			peak := avg * 1.55
			flat := avg
			valley := avg * 0.5
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO grid_agency_price
				   (operating_month, voltage_level, avg_price, peak_price, flat_price, valley_price, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7::uuid)
				 ON CONFLICT (org_id, operating_month, voltage_level) DO UPDATE SET
				   avg_price = EXCLUDED.avg_price, peak_price = EXCLUDED.peak_price,
				   flat_price = EXCLUDED.flat_price, valley_price = EXCLUDED.valley_price`,
				ym, v, avg, peak, flat, valley, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
