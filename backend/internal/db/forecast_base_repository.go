// 预测基础数据（节假日+典型曲线）。
// 2026-06 自 p0_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── U3 节假日 + 典型曲线 ───────────────

type Holiday struct {
	ID          string    `json:"id"`
	HolidayDate time.Time `json:"holiday_date"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type TypicalCurve struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Season    string    `json:"season"`
	DayType   string    `json:"day_type"`
	Curve96   []float64 `json:"curve_96"`
	Note      *string   `json:"note,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type ForecastBaseRepository struct{ pool *Pool }

func NewForecastBaseRepository(pool *Pool) *ForecastBaseRepository {
	return &ForecastBaseRepository{pool: pool}
}

func (r *ForecastBaseRepository) ListHolidays(ctx context.Context, year int) ([]*Holiday, error) {
	if year == 0 {
		year = time.Now().Year()
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	args := []any{start, end}
	q := `
		SELECT id, holiday_date, name, kind, note, created_at
		FROM holidays WHERE holiday_date >= $1 AND holiday_date < $2`
	n := 3
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	q += " ORDER BY holiday_date ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Holiday, 0)
	for rows.Next() {
		var h Holiday
		if err := rows.Scan(&h.ID, &h.HolidayDate, &h.Name, &h.Kind, &h.Note, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &h)
	}
	return list, rows.Err()
}

func (r *ForecastBaseRepository) ListTypicalCurves(ctx context.Context) ([]*TypicalCurve, error) {
	args := []any{}
	n := 1
	q := `
		SELECT id, name, season, day_type, curve_96, note, enabled, created_at
		FROM typical_curves WHERE enabled = true`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	q += " ORDER BY season, day_type, name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TypicalCurve, 0)
	for rows.Next() {
		var c TypicalCurve
		if err := rows.Scan(&c.ID, &c.Name, &c.Season, &c.DayType, &c.Curve96,
			&c.Note, &c.Enabled, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (r *ForecastBaseRepository) GenerateDemo(ctx context.Context) (int, int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	// 2026 中国法定节假日（演示）
	holidays := []struct {
		date string
		name string
		kind string
	}{
		{"2026-01-01", "元旦", "public"},
		{"2026-02-17", "春节", "public"},
		{"2026-02-18", "春节", "public"},
		{"2026-02-19", "春节", "public"},
		{"2026-04-05", "清明节", "public"},
		{"2026-05-01", "劳动节", "public"},
		{"2026-05-04", "调休补班", "makeup_workday"},
		{"2026-06-19", "端午节", "public"},
		{"2026-09-25", "中秋节", "public"},
		{"2026-10-01", "国庆节", "public"},
		{"2026-10-02", "国庆节", "public"},
		{"2026-10-03", "国庆节", "public"},
	}
	hCnt := 0
	for _, h := range holidays {
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO holidays (holiday_date, name, kind, org_id)
			 VALUES ($1::date, $2, $3, $4::uuid)
			 ON CONFLICT (org_id, holiday_date) DO UPDATE SET
			   name = EXCLUDED.name, kind = EXCLUDED.kind`,
			h.date, h.name, h.kind, orgID); err != nil {
			return hCnt, 0, err
		}
		hCnt++
	}

	// 典型曲线：4 季 × 3 日类 = 12 条
	curves := []struct {
		name, season, dayType string
		base                  float64
		peakAmp               float64
	}{
		{"春季工作日", "spring", "workday", 1200, 600},
		{"春季周末", "spring", "weekend", 900, 400},
		{"春季节假日", "spring", "holiday", 700, 350},
		{"夏季工作日", "summer", "workday", 1800, 900},
		{"夏季周末", "summer", "weekend", 1400, 700},
		{"夏季节假日", "summer", "holiday", 1100, 600},
		{"秋季工作日", "autumn", "workday", 1100, 550},
		{"秋季周末", "autumn", "weekend", 850, 400},
		{"秋季节假日", "autumn", "holiday", 700, 350},
		{"冬季工作日", "winter", "workday", 1500, 800},
		{"冬季周末", "winter", "weekend", 1200, 600},
		{"冬季节假日", "winter", "holiday", 900, 450},
	}
	cCnt := 0
	for _, c := range curves {
		curve := make([]float64, 96)
		for p := 0; p < 96; p++ {
			hour := float64(p) / 4
			// 双峰曲线：上午峰 9-12，下午峰 18-21
			factor := 0.5 + 0.5*sineWave(hour, 9)*0.6 + 0.5*sineWave(hour, 20)*0.7
			curve[p] = c.base + c.peakAmp*factor + (rand.Float64()-0.5)*50
		}
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO typical_curves (name, season, day_type, curve_96, enabled, org_id)
			 VALUES ($1, $2, $3, $4, true, $5::uuid)
			 ON CONFLICT (org_id, name) DO UPDATE SET
			   season = EXCLUDED.season, day_type = EXCLUDED.day_type,
			   curve_96 = EXCLUDED.curve_96`,
			c.name, c.season, c.dayType, curve, orgID); err != nil {
			return hCnt, cCnt, err
		}
		cCnt++
	}
	return hCnt, cCnt, nil
}

func sineWave(hour, peakHour float64) float64 {
	// 在 peakHour ± 4 小时之间得到大值
	d := hour - peakHour
	if d > 12 {
		d -= 24
	}
	if d < -12 {
		d += 24
	}
	if d > 4 || d < -4 {
		return 0
	}
	return 1 - (d/4)*(d/4) // 倒抛物线峰
}

// joinWhere 用 AND 连接 where 子句
func joinWhere(parts []string) string {
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += " AND "
		}
		s += p
	}
	return s
}
