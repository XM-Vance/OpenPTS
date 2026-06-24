// 气象仓储。
// 2026-06 自 v1clone_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── D4 气象 ───────────────

type WeatherRecord struct {
	ID          string    `json:"id"`
	ObsDate     time.Time `json:"obs_date"`
	Location    string    `json:"location"`
	TempHigh    *float64  `json:"temp_high,omitempty"`
	TempLow     *float64  `json:"temp_low,omitempty"`
	Humidity    *float64  `json:"humidity,omitempty"`
	PrecipMm    *float64  `json:"precip_mm,omitempty"`
	WindKmh     *float64  `json:"wind_kmh,omitempty"`
	LoadFactor  *float64  `json:"load_factor,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type WeatherRepository struct{ pool *Pool }

func NewWeatherRepository(pool *Pool) *WeatherRepository { return &WeatherRepository{pool: pool} }

func (r *WeatherRepository) List(ctx context.Context, location string, days int) ([]*WeatherRecord, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, obs_date, location, temp_high, temp_low, humidity, precip_mm,
			wind_kmh, load_factor, description, created_at
		  FROM weather_data WHERE obs_date >= $1`
	if location != "" {
		args = append(args, location)
		q += " AND location = $2"
	}
	q += " ORDER BY obs_date DESC, location ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*WeatherRecord, 0)
	for rows.Next() {
		var w WeatherRecord
		if err := rows.Scan(&w.ID, &w.ObsDate, &w.Location, &w.TempHigh, &w.TempLow,
			&w.Humidity, &w.PrecipMm, &w.WindKmh, &w.LoadFactor, &w.Description,
			&w.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &w)
	}
	return list, rows.Err()
}

func (r *WeatherRepository) GenerateDemo(ctx context.Context) (int, error) {
	descs := []string{"晴", "多云", "阴", "阵雨", "雷阵雨", "雾"}
	locs := []string{"广州", "深圳", "佛山", "东莞"}
	cnt := 0
	for i := 0; i < 14; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		for _, loc := range locs {
			high := 22 + rand.Float64()*12
			low := high - 6 - rand.Float64()*4
			hum := 50 + rand.Float64()*40
			prec := 0.0
			if rand.Float64() < 0.3 {
				prec = rand.Float64() * 30
			}
			wind := 5 + rand.Float64()*20
			lf := 0.85 + rand.Float64()*0.3
			desc := descs[rand.Intn(len(descs))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO weather_data
				   (obs_date, location, temp_high, temp_low, humidity, precip_mm,
				    wind_kmh, load_factor, description)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				 ON CONFLICT (obs_date, location) DO UPDATE SET
				   temp_high = EXCLUDED.temp_high, temp_low = EXCLUDED.temp_low,
				   humidity = EXCLUDED.humidity, precip_mm = EXCLUDED.precip_mm,
				   wind_kmh = EXCLUDED.wind_kmh, load_factor = EXCLUDED.load_factor,
				   description = EXCLUDED.description`,
				d, loc, high, low, hum, prec, wind, lf, desc); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
