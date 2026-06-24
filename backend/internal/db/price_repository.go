// 价格仓储：日前现货价历史查询 + 写入（演示数据用）。
package db

import (
	"context"
	"time"
)

type DailyPriceCurve struct {
	Date    time.Time `json:"date"`
	Curve48 []float64 `json:"curve_48"`
}

type PriceRepository struct {
	pool *Pool
}

func NewPriceRepository(pool *Pool) *PriceRepository {
	return &PriceRepository{pool: pool}
}

// GetRecentDayAheadCurves 返回 before 之前最近 limit 天的日前价 48 点完整曲线（按日期升序）。
// 用 array_agg(::float8) 将 NUMERIC 列在 SQL 层转为 float8[]，pgx 可直接读到 []float64。
// HAVING 过滤掉点位不全的天。
func (r *PriceRepository) GetRecentDayAheadCurves(
	ctx context.Context, before time.Time, limit int,
) ([]*DailyPriceCurve, error) {
	const q = `
		SELECT date, array_agg(price_da::float8 ORDER BY period) AS curve
		FROM day_ahead_spot_price
		WHERE date < $1 AND price_da IS NOT NULL
		GROUP BY date
		HAVING count(*) >= 48
		ORDER BY date DESC
		LIMIT $2`
	rows, err := r.pool.Query(ctx, q, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*DailyPriceCurve, 0, limit)
	for rows.Next() {
		var d DailyPriceCurve
		if err := rows.Scan(&d.Date, &d.Curve48); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 反转为升序（最旧在前）
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

// UpsertDayAheadPrice 写入单点日前价。
func (r *PriceRepository) UpsertDayAheadPrice(
	ctx context.Context, d time.Time, period int, priceDA float64,
) error {
	const q = `
		INSERT INTO day_ahead_spot_price (date, period, price_da)
		VALUES ($1, $2, $3)
		ON CONFLICT (date, period)
		DO UPDATE SET price_da = EXCLUDED.price_da`
	_, err := r.pool.Exec(ctx, q, d, period, priceDA)
	return err
}
