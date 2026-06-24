// weather_obs_repository.go —— 外部气象观测（风电场风速 / 水库水文）查询与演示数据。
// 数据源 md_weather_wind_hourly、md_weather_hydrology_daily（原属市场行情，现并入气象数据模块）。
// 二者为按站点采集的共享观测数据，不做省份(org_id)隔离。
package db

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// WeatherStation 观测站点（风电场 / 水库）。
type WeatherStation struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// WindHourlyRow 风电场逐小时观测。
type WindHourlyRow struct {
	LocationName  string    `json:"location_name"`
	ObsTime       time.Time `json:"obs_time"`
	WindSpeed100m *float64  `json:"wind_speed_100m"`
	WindDir100m   *int16    `json:"wind_dir_100m"`
	Temperature2m *float64  `json:"temperature_2m"`
	Humidity2m    *float64  `json:"humidity_2m"`
}

// HydroDailyRow 水库水文逐日观测。
type HydroDailyRow struct {
	LocationName  string    `json:"location_name"`
	ObsDate       time.Time `json:"obs_date"`
	TempMean      *float64  `json:"temp_mean"`
	HumidityMean  *float64  `json:"humidity_mean"`
	Precipitation *float64  `json:"precipitation_sum"`
	RainSum       *float64  `json:"rain_sum"`
	Et0           *float64  `json:"et0_evapotranspiration"`
	WindSpeed10m  *float64  `json:"wind_speed_10m_mean"`
}

// WindFarmStations 列出风电场站点。
func (r *WeatherRepository) WindFarmStations(ctx context.Context) ([]*WeatherStation, error) {
	return r.stations(ctx, "md_weather_wind_hourly")
}

// HydrologyStations 列出水库水文站点。
func (r *WeatherRepository) HydrologyStations(ctx context.Context) ([]*WeatherStation, error) {
	return r.stations(ctx, "md_weather_hydrology_daily")
}

// stations 取某观测表的去重站点列表（table 仅取自固定字面量，安全）。
func (r *WeatherRepository) stations(ctx context.Context, table string) ([]*WeatherStation, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT DISTINCT location_code, location_name FROM %s ORDER BY location_name`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*WeatherStation, 0)
	for rows.Next() {
		var s WeatherStation
		if err := rows.Scan(&s.Code, &s.Name); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

// WindFarmHourly 取某站点近 hours 小时风速观测（station 为空则不限站点），按时间倒序。
func (r *WeatherRepository) WindFarmHourly(ctx context.Context, station string, hours int) ([]*WindHourlyRow, error) {
	if hours <= 0 || hours > 24*60 {
		hours = 72
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	args := []any{since}
	q := `SELECT location_name, obs_time, wind_speed_100m, wind_dir_100m, temperature_2m, humidity_2m
	      FROM md_weather_wind_hourly WHERE obs_time >= $1`
	if station != "" {
		args = append(args, station)
		q += fmt.Sprintf(" AND location_code = $%d", len(args))
	}
	q += " ORDER BY obs_time DESC LIMIT 2000"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*WindHourlyRow, 0)
	for rows.Next() {
		var w WindHourlyRow
		if err := rows.Scan(&w.LocationName, &w.ObsTime, &w.WindSpeed100m,
			&w.WindDir100m, &w.Temperature2m, &w.Humidity2m); err != nil {
			return nil, err
		}
		list = append(list, &w)
	}
	return list, rows.Err()
}

// HydrologyDaily 取某站点近 days 天水文观测（station 为空则不限站点），按日期倒序。
func (r *WeatherRepository) HydrologyDaily(ctx context.Context, station string, days int) ([]*HydroDailyRow, error) {
	if days <= 0 || days > 3650 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT location_name, obs_date, temp_mean, humidity_mean, precipitation_sum,
	             rain_sum, et0_evapotranspiration, wind_speed_10m_mean
	      FROM md_weather_hydrology_daily WHERE obs_date >= $1`
	if station != "" {
		args = append(args, station)
		q += fmt.Sprintf(" AND location_code = $%d", len(args))
	}
	q += " ORDER BY obs_date DESC LIMIT 2000"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*HydroDailyRow, 0)
	for rows.Next() {
		var h HydroDailyRow
		if err := rows.Scan(&h.LocationName, &h.ObsDate, &h.TempMean, &h.HumidityMean,
			&h.Precipitation, &h.RainSum, &h.Et0, &h.WindSpeed10m); err != nil {
			return nil, err
		}
		list = append(list, &h)
	}
	return list, rows.Err()
}

// obsWindStations / obsHydroStations 演示站点（含经纬度）。
// 二次开发时替换为你自己的监测站点。
var obsWindStations = []struct {
	Code, Name string
	Lat, Lon   float64
}{
	{"WF_01", "示例风电场A", 31.20, 121.40},
	{"WF_02", "示例风电场B", 31.10, 121.50},
	{"WF_03", "示例风电场C", 31.30, 121.30},
}

var obsHydroStations = []struct {
	Code, Name string
	Lat, Lon   float64
}{
	{"RV_01", "示例水库", 31.25, 121.45},
	{"RV_02", "示例水文站", 31.15, 121.55},
	{"RV_03", "示例水文站B", 31.35, 121.35},
}

// GenerateObsDemo 为风电场风速（近 7 天逐时）与水库水文（近 30 天逐日）生成演示观测数据。
// 返回写入的记录条数。ON CONFLICT 幂等，可重复调用。
func (r *WeatherRepository) GenerateObsDemo(ctx context.Context) (int, error) {
	cnt := 0
	now := time.Now()

	// 风电场：近 7 天逐小时。
	for _, st := range obsWindStations {
		for i := 7 * 24; i >= 0; i-- {
			t := now.Add(-time.Duration(i) * time.Hour).Truncate(time.Hour)
			hour := float64(t.Hour())
			// 风速：白天偏小、夜间偏大的弱日变化 + 随机，范围约 3~16 m/s。
			ws := round2(8 + 4*math.Sin((hour-3)/24*2*math.Pi) + rand.Float64()*4 - 1)
			if ws < 0.5 {
				ws = 0.5
			}
			dir := int16(rand.Intn(360))
			temp := round2(18 + 8*math.Sin((hour-9)/24*2*math.Pi) + rand.Float64()*2)
			hum := round2(60 + 25*math.Sin((hour-3)/24*2*math.Pi) + rand.Float64()*5)
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO md_weather_wind_hourly
				   (location_code, location_name, lat, lon, obs_time,
				    wind_speed_100m, wind_dir_100m, temperature_2m, humidity_2m)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				 ON CONFLICT (location_code, obs_time) DO UPDATE SET
				   wind_speed_100m = EXCLUDED.wind_speed_100m, wind_dir_100m = EXCLUDED.wind_dir_100m,
				   temperature_2m = EXCLUDED.temperature_2m, humidity_2m = EXCLUDED.humidity_2m`,
				st.Code, st.Name, st.Lat, st.Lon, t, ws, dir, temp, hum); err != nil {
				return cnt, err
			}
			cnt++
		}
	}

	// 水库水文：近 30 天逐日。
	for _, st := range obsHydroStations {
		for i := 30; i >= 0; i-- {
			d := now.AddDate(0, 0, -i).Truncate(24 * time.Hour)
			tmean := round2(20 + 6*rand.Float64())
			hmean := round2(65 + 20*rand.Float64())
			precip := round2(math.Max(0, rand.Float64()*30-12)) // 多数为 0，偶有降水
			rain := round2(precip * (0.8 + rand.Float64()*0.2))
			et0 := round2(2 + rand.Float64()*3)
			ws10 := round2(2 + rand.Float64()*4)
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO md_weather_hydrology_daily
				   (location_code, location_name, lat, lon, obs_date,
				    temp_mean, humidity_mean, precipitation_sum, rain_sum, et0_evapotranspiration, wind_speed_10m_mean)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
				 ON CONFLICT (location_code, obs_date) DO UPDATE SET
				   temp_mean = EXCLUDED.temp_mean, humidity_mean = EXCLUDED.humidity_mean,
				   precipitation_sum = EXCLUDED.precipitation_sum, rain_sum = EXCLUDED.rain_sum,
				   et0_evapotranspiration = EXCLUDED.et0_evapotranspiration,
				   wind_speed_10m_mean = EXCLUDED.wind_speed_10m_mean`,
				st.Code, st.Name, st.Lat, st.Lon, d, tmean, hmean, precip, rain, et0, ws10); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
