// weather_ext_repository.go —— 站点/实况摘要/预报摘要/可用预报日 查询，
// 对齐前端 lib/api/weather.ts 的契约（基于 weather_locations / weather_actuals / weather_forecasts 表）。
package db

import (
	"context"
	"time"
)

// WeatherLocationRow 对应前端 WeatherLocation。location_id 取站点名（actuals/forecasts 按 location_name 关联）。
type WeatherLocationRow struct {
	LocationID string  `json:"location_id"`
	Name       string  `json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Enabled    bool    `json:"enabled"`
}

// WeatherDailySummary 对应前端 DailyWeatherSummary。
type WeatherDailySummary struct {
	Date             string  `json:"date"`
	WeatherType      string  `json:"weather_type"`
	WeatherIcon      string  `json:"weather_icon"`
	MinTemp          float64 `json:"min_temp"`
	MaxTemp          float64 `json:"max_temp"`
	AvgPrecipitation float64 `json:"avg_precipitation"`
	AvgCloudCover    float64 `json:"avg_cloud_cover"`
}

func (r *WeatherRepository) ListLocations(ctx context.Context) ([]*WeatherLocationRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT name, COALESCE(latitude,0)::float8, COALESCE(longitude,0)::float8
		FROM weather_locations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*WeatherLocationRow, 0)
	for rows.Next() {
		var w WeatherLocationRow
		if err := rows.Scan(&w.Name, &w.Latitude, &w.Longitude); err != nil {
			return nil, err
		}
		w.LocationID = w.Name
		w.Enabled = true
		list = append(list, &w)
	}
	return list, rows.Err()
}

func (r *WeatherRepository) CreateLocation(ctx context.Context, name string, lat, lon float64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO weather_locations (name, latitude, longitude)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, name, lat, lon)
	return err
}

func (r *WeatherRepository) UpdateLocation(ctx context.Context, name string, lat, lon float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE weather_locations SET latitude=$2, longitude=$3, updated_at=now() WHERE name=$1`,
		name, lat, lon)
	return err
}

func (r *WeatherRepository) DeleteLocation(ctx context.Context, name string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM weather_locations WHERE name=$1`, name)
	return err
}

// ActualsSummary 取某站某日的实况日摘要；无数据返回 nil。
func (r *WeatherRepository) ActualsSummary(ctx context.Context, location, date string) (*WeatherDailySummary, error) {
	var s WeatherDailySummary
	err := r.pool.QueryRow(ctx, `
		SELECT to_char(date,'YYYY-MM-DD'),
		       COALESCE(min_temp,0)::float8, COALESCE(max_temp,0)::float8,
		       COALESCE(precipitation,0)::float8
		FROM weather_actuals WHERE location_name=$1 AND date=$2 LIMIT 1`,
		location, date).Scan(&s.Date, &s.MinTemp, &s.MaxTemp, &s.AvgPrecipitation)
	if err != nil {
		return nil, err
	}
	s.AvgCloudCover = 0 // 库内无云量字段
	return &s, nil
}

// ForecastsSummary 取某站某预报日的逐目标日预报摘要。
func (r *WeatherRepository) ForecastsSummary(ctx context.Context, location, forecastDate string) ([]*WeatherDailySummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT to_char(target_date,'YYYY-MM-DD'), COALESCE(temp_forecast,0)::float8
		FROM weather_forecasts WHERE location_name=$1 AND forecast_date=$2
		ORDER BY target_date`, location, forecastDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*WeatherDailySummary, 0)
	for rows.Next() {
		var s WeatherDailySummary
		var temp float64
		if err := rows.Scan(&s.Date, &temp); err != nil {
			return nil, err
		}
		s.MinTemp, s.MaxTemp = temp, temp // 库内仅有单一预报温度
		list = append(list, &s)
	}
	return list, rows.Err()
}

// ForecastDates 列出某站可用的预报日（可选限定目标日）。
func (r *WeatherRepository) ForecastDates(ctx context.Context, location, targetDate string) ([]string, error) {
	q := `SELECT DISTINCT forecast_date FROM weather_forecasts WHERE location_name=$1`
	args := []any{location}
	if targetDate != "" {
		q += ` AND target_date=$2`
		args = append(args, targetDate)
	}
	q += ` ORDER BY forecast_date DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dates := make([]string, 0)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		dates = append(dates, d.Format("2006-01-02"))
	}
	return dates, rows.Err()
}
