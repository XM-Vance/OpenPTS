// 光伏仓储：站点维表 + 发电预测 + 收益结算。
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ── 站点 ──

type SolarStation struct {
	ID            uuid.UUID  `json:"id"`
	StationName   string     `json:"station_name"`
	Location      string     `json:"location"`
	CapacityKW    float64    `json:"capacity_kw"`
	Status        string     `json:"status"`
	InstalledDate *time.Time `json:"installed_date,omitempty"`
	Latitude      *float64   `json:"latitude,omitempty"`
	Longitude     *float64   `json:"longitude,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ── 发电预测 ──

type SolarGenerationForecast struct {
	ID              uuid.UUID  `json:"id"`
	StationID       uuid.UUID  `json:"station_id"`
	ForecastDate    time.Time  `json:"forecast_date"`
	Period          int        `json:"period"`
	ForecastPowerKW float64    `json:"forecast_power_kw"`
	ActualPowerKW   *float64   `json:"actual_power_kw,omitempty"`
	DeviationRate   *float64   `json:"deviation_rate,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ── 收益结算 ──

type SolarRevenueSettlement struct {
	ID              uuid.UUID `json:"id"`
	StationID       uuid.UUID `json:"station_id"`
	SettlementMonth string    `json:"settlement_month"`
	EnergyKWh       float64   `json:"energy_kwh"`
	Revenue         float64   `json:"revenue"`
	AvgPrice        float64   `json:"avg_price"`
	Subsidy         float64   `json:"subsidy"`
	NetIncome       float64   `json:"net_income"`
	CreatedAt       time.Time `json:"created_at"`
}

// ── Repository ──

type SolarRepository struct {
	pool *Pool
}

func NewSolarRepository(pool *Pool) *SolarRepository {
	return &SolarRepository{pool: pool}
}

// ── 站点 CRUD ──

const solarStationCols = "id, station_name, location, capacity_kw, status, installed_date, latitude, longitude, created_at"

func (r *SolarRepository) ListStations(ctx context.Context) ([]*SolarStation, error) {
	q := `SELECT ` + solarStationCols + ` FROM solar_stations`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY station_name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*SolarStation, 0, 8)
	for rows.Next() {
		var s SolarStation
		if err := rows.Scan(&s.ID, &s.StationName, &s.Location, &s.CapacityKW,
			&s.Status, &s.InstalledDate, &s.Latitude, &s.Longitude, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *SolarRepository) GetStation(ctx context.Context, id uuid.UUID) (*SolarStation, error) {
	q := `SELECT ` + solarStationCols + ` FROM solar_stations WHERE id = $1`
	args := []any{id}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	var s SolarStation
	err := r.pool.QueryRow(ctx, q, args...).Scan(&s.ID, &s.StationName, &s.Location, &s.CapacityKW,
		&s.Status, &s.InstalledDate, &s.Latitude, &s.Longitude, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SolarRepository) CreateStation(ctx context.Context, name, location string, capacityKW float64, status string, installedDate *time.Time, lat, lng *float64) (*SolarStation, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	const q = `
		INSERT INTO solar_stations (station_name, location, capacity_kw, status, installed_date, latitude, longitude, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::uuid)
		RETURNING ` + solarStationCols
	var s SolarStation
	err := r.pool.QueryRow(ctx, q, name, location, capacityKW, status, installedDate, lat, lng, org).
		Scan(&s.ID, &s.StationName, &s.Location, &s.CapacityKW,
			&s.Status, &s.InstalledDate, &s.Latitude, &s.Longitude, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SolarRepository) UpdateStation(ctx context.Context, id uuid.UUID, name, location string, capacityKW float64, status string, installedDate *time.Time, lat, lng *float64) (*SolarStation, error) {
	const q = `
		UPDATE solar_stations SET station_name=$2, location=$3, capacity_kw=$4, status=$5, installed_date=$6, latitude=$7, longitude=$8
		WHERE id=$1
		RETURNING ` + solarStationCols
	var s SolarStation
	err := r.pool.QueryRow(ctx, q, id, name, location, capacityKW, status, installedDate, lat, lng).
		Scan(&s.ID, &s.StationName, &s.Location, &s.CapacityKW,
			&s.Status, &s.InstalledDate, &s.Latitude, &s.Longitude, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SolarRepository) DeleteStation(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM solar_stations WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

// ── 发电预测 ──

func (r *SolarRepository) ListForecast(ctx context.Context, stationID *uuid.UUID, limit int) ([]*SolarGenerationForecast, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var q string
	var args []any
	argN := 0
	if stationID != nil {
		argN++
		q = `SELECT id, station_id, forecast_date, period, forecast_power_kw, actual_power_kw, deviation_rate, created_at
			FROM solar_generation_forecast
			WHERE station_id = $` + fmt.Sprintf("%d", argN)
		args = append(args, *stationID)
	} else {
		q = `SELECT id, station_id, forecast_date, period, forecast_power_kw, actual_power_kw, deviation_rate, created_at
			FROM solar_generation_forecast`
	}
	org, scoped := OrgFilter(ctx)
	if scoped {
		argN++
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", argN)
	}
	argN++
	q += ` ORDER BY forecast_date DESC, period ASC
		LIMIT $` + fmt.Sprintf("%d", argN)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*SolarGenerationForecast, 0, limit)
	for rows.Next() {
		var f SolarGenerationForecast
		if err := rows.Scan(&f.ID, &f.StationID, &f.ForecastDate, &f.Period,
			&f.ForecastPowerKW, &f.ActualPowerKW, &f.DeviationRate, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &f)
	}
	return list, rows.Err()
}

func (r *SolarRepository) UpsertForecast(ctx context.Context, stationID uuid.UUID, forecastDate time.Time, period int, forecastKW, actualKW, deviation *float64) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	const q = `
		INSERT INTO solar_generation_forecast (station_id, forecast_date, period, forecast_power_kw, actual_power_kw, deviation_rate, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7::uuid)
		ON CONFLICT (station_id, forecast_date, period) DO UPDATE SET
			forecast_power_kw = EXCLUDED.forecast_power_kw,
			actual_power_kw   = EXCLUDED.actual_power_kw,
			deviation_rate    = EXCLUDED.deviation_rate`
	_, err := r.pool.Exec(ctx, q, stationID, forecastDate, period, forecastKW, actualKW, deviation, org)
	return err
}

// ── 收益结算 ──

func (r *SolarRepository) ListRevenue(ctx context.Context, stationID *uuid.UUID, limit int) ([]*SolarRevenueSettlement, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var q string
	var args []any
	argN := 0
	if stationID != nil {
		argN++
		q = `SELECT id, station_id, settlement_month, energy_kwh, revenue, avg_price, subsidy, net_income, created_at
			FROM solar_revenue_settlement
			WHERE station_id = $` + fmt.Sprintf("%d", argN)
		args = append(args, *stationID)
	} else {
		q = `SELECT id, station_id, settlement_month, energy_kwh, revenue, avg_price, subsidy, net_income, created_at
			FROM solar_revenue_settlement`
	}
	org, scoped := OrgFilter(ctx)
	if scoped {
		argN++
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", argN)
	}
	argN++
	q += ` ORDER BY settlement_month DESC
		LIMIT $` + fmt.Sprintf("%d", argN)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*SolarRevenueSettlement, 0, limit)
	for rows.Next() {
		var s SolarRevenueSettlement
		if err := rows.Scan(&s.ID, &s.StationID, &s.SettlementMonth, &s.EnergyKWh,
			&s.Revenue, &s.AvgPrice, &s.Subsidy, &s.NetIncome, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

func (r *SolarRepository) UpsertRevenue(ctx context.Context, stationID uuid.UUID, month string, energyKWh, revenue, avgPrice, subsidy, netIncome float64) error {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return ErrOrgRequired
	}
	const q = `
		INSERT INTO solar_revenue_settlement (station_id, settlement_month, energy_kwh, revenue, avg_price, subsidy, net_income, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::uuid)
		ON CONFLICT (station_id, settlement_month) DO UPDATE SET
			energy_kwh = EXCLUDED.energy_kwh,
			revenue    = EXCLUDED.revenue,
			avg_price  = EXCLUDED.avg_price,
			subsidy    = EXCLUDED.subsidy,
			net_income = EXCLUDED.net_income`
	_, err := r.pool.Exec(ctx, q, stationID, month, energyKWh, revenue, avgPrice, subsidy, netIncome, org)
	return err
}
