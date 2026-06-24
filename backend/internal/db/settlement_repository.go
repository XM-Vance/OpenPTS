// 结算仓储：批发日结算列表 / 详情 / 写入（演示数据用）。
// period_details 字段是 JSONB（48 时段明细），用 json.RawMessage 透传。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SettlementDaily struct {
	ID                   uuid.UUID       `json:"id"`
	OperatingDate        time.Time       `json:"operating_date"`
	Version              string          `json:"version"`
	PeriodDetails        json.RawMessage `json:"period_details,omitempty"`
	ContractFee          *float64        `json:"contract_fee,omitempty"`
	DayAheadFee          *float64        `json:"day_ahead_fee,omitempty"`
	RealTimeFee          *float64        `json:"real_time_fee,omitempty"`
	TotalEnergyFee       *float64        `json:"total_energy_fee,omitempty"`
	EnergyAvgPrice       *float64        `json:"energy_avg_price,omitempty"`
	DeviationRecoveryFee *float64        `json:"deviation_recovery_fee,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
}

var ErrSettlementNotFound = errors.New("结算记录不存在")

type SettlementRepository struct {
	pool *Pool
}

func NewSettlementRepository(pool *Pool) *SettlementRepository {
	return &SettlementRepository{pool: pool}
}

// ListRecent 列表用：不带 period_details，减少响应体。
func (r *SettlementRepository) ListRecent(ctx context.Context, limit int) ([]*SettlementDaily, error) {
	args := []any{}
	q := `SELECT id, operating_date, version,
		contract_fee, day_ahead_fee, real_time_fee,
		total_energy_fee, energy_avg_price, deviation_recovery_fee, created_at
		FROM settlement_daily WHERE 1=1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += ` ORDER BY operating_date DESC, version
		LIMIT $` + itoaNew(len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*SettlementDaily, 0, limit)
	for rows.Next() {
		var s SettlementDaily
		if err := rows.Scan(&s.ID, &s.OperatingDate, &s.Version,
			&s.ContractFee, &s.DayAheadFee, &s.RealTimeFee,
			&s.TotalEnergyFee, &s.EnergyAvgPrice, &s.DeviationRecoveryFee,
			&s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

// GetByDate 详情：含 period_details JSONB。
func (r *SettlementRepository) GetByDate(ctx context.Context, date time.Time, version string) (*SettlementDaily, error) {
	args := []any{date, version}
	q := `SELECT id, operating_date, version, period_details,
		contract_fee, day_ahead_fee, real_time_fee,
		total_energy_fee, energy_avg_price, deviation_recovery_fee, created_at
		FROM settlement_daily
		WHERE operating_date = $1 AND version = $2`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	var s SettlementDaily
	var pd []byte
	err := r.pool.QueryRow(ctx, q, args...).Scan(
		&s.ID, &s.OperatingDate, &s.Version, &pd,
		&s.ContractFee, &s.DayAheadFee, &s.RealTimeFee,
		&s.TotalEnergyFee, &s.EnergyAvgPrice, &s.DeviationRecoveryFee,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSettlementNotFound
		}
		return nil, err
	}
	s.PeriodDetails = json.RawMessage(pd)
	return &s, nil
}

// Upsert 写入或覆盖（(org_id, operating_date, version) 唯一）。
// period_details 用 json.RawMessage 传给 pgx，由其 JSONB 编解码识别。
func (r *SettlementRepository) Upsert(ctx context.Context, s *SettlementDaily) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	const q = `
		INSERT INTO settlement_daily
			(operating_date, version, period_details,
			 contract_fee, day_ahead_fee, real_time_fee,
			 total_energy_fee, energy_avg_price, deviation_recovery_fee, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::uuid)
		ON CONFLICT (org_id, operating_date, version) DO UPDATE SET
			period_details = EXCLUDED.period_details,
			contract_fee = EXCLUDED.contract_fee,
			day_ahead_fee = EXCLUDED.day_ahead_fee,
			real_time_fee = EXCLUDED.real_time_fee,
			total_energy_fee = EXCLUDED.total_energy_fee,
			energy_avg_price = EXCLUDED.energy_avg_price,
			deviation_recovery_fee = EXCLUDED.deviation_recovery_fee`
	_, err := r.pool.Exec(ctx, q,
		s.OperatingDate, s.Version, s.PeriodDetails,
		s.ContractFee, s.DayAheadFee, s.RealTimeFee,
		s.TotalEnergyFee, s.EnergyAvgPrice, s.DeviationRecoveryFee, org)
	return err
}
