// 储能仓储：站点维表 + 日运营记录。
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StorageStation struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CapacityMWh float64   `json:"capacity_mwh"`
	MaxPowerMW  float64   `json:"max_power_mw"`
	Location    *string   `json:"location,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StorageOperation struct {
	ID            uuid.UUID `json:"id"`
	StationID     uuid.UUID `json:"station_id"`
	OperationDate time.Time `json:"operation_date"`
	ChargeMWh     float64   `json:"charge_mwh"`
	DischargeMWh  float64   `json:"discharge_mwh"`
	Revenue       *float64  `json:"revenue,omitempty"`
	AvgSOC        *float64  `json:"avg_soc,omitempty"`
	Cycles        *float64  `json:"cycles,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StorageRepository struct {
	pool *Pool
}

func NewStorageRepository(pool *Pool) *StorageRepository {
	return &StorageRepository{pool: pool}
}

const stationColumns = "id, name, capacity_mwh, max_power_mw, location, status, created_at, updated_at"

func (r *StorageRepository) ListStations(ctx context.Context) ([]*StorageStation, error) {
	q := `SELECT ` + stationColumns + ` FROM storage_stations`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*StorageStation, 0, 8)
	for rows.Next() {
		var s StorageStation
		if err := rows.Scan(&s.ID, &s.Name, &s.CapacityMWh, &s.MaxPowerMW,
			&s.Location, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

// UpsertStation 以 org_id + name 作为业务唯一键。
func (r *StorageRepository) UpsertStation(
	ctx context.Context, name string, capacityMWh, maxPowerMW float64, location, status string,
) (*StorageStation, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	const q = `
		INSERT INTO storage_stations (name, capacity_mwh, max_power_mw, location, status, org_id)
		VALUES ($1, $2, $3, $4, $5, $6::uuid)
		ON CONFLICT (org_id, name) DO UPDATE SET
			capacity_mwh = EXCLUDED.capacity_mwh,
			max_power_mw = EXCLUDED.max_power_mw,
			location     = EXCLUDED.location,
			status       = EXCLUDED.status
		RETURNING ` + stationColumns
	var s StorageStation
	err := r.pool.QueryRow(ctx, q, name, capacityMWh, maxPowerMW, nullStr(location), status, org).
		Scan(&s.ID, &s.Name, &s.CapacityMWh, &s.MaxPowerMW,
			&s.Location, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StorageRepository) ListOperations(
	ctx context.Context, stationID uuid.UUID, limit int,
) ([]*StorageOperation, error) {
	q := `
		SELECT id, station_id, operation_date, charge_mwh, discharge_mwh,
			revenue, avg_soc, cycles, created_at, updated_at
		FROM storage_daily_operation
		WHERE station_id = $1`
	args := []any{stationID}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += `
		ORDER BY operation_date DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*StorageOperation, 0, limit)
	for rows.Next() {
		var o StorageOperation
		if err := rows.Scan(&o.ID, &o.StationID, &o.OperationDate,
			&o.ChargeMWh, &o.DischargeMWh, &o.Revenue, &o.AvgSOC, &o.Cycles,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &o)
	}
	return list, rows.Err()
}

func (r *StorageRepository) UpsertOperation(
	ctx context.Context, stationID uuid.UUID, d time.Time,
	chargeMWh, dischargeMWh, revenue, avgSOC, cycles float64,
) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	const q = `
		INSERT INTO storage_daily_operation
			(station_id, operation_date, charge_mwh, discharge_mwh, revenue, avg_soc, cycles, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::uuid)
		ON CONFLICT (org_id, station_id, operation_date) DO UPDATE SET
			charge_mwh    = EXCLUDED.charge_mwh,
			discharge_mwh = EXCLUDED.discharge_mwh,
			revenue       = EXCLUDED.revenue,
			avg_soc       = EXCLUDED.avg_soc,
			cycles        = EXCLUDED.cycles`
	_, err := r.pool.Exec(ctx, q, stationID, d, chargeMWh, dischargeMWh, revenue, avgSOC, cycles, org)
	return err
}
