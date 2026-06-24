package db

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// CarbonQuote 碳交易行情单日记录（CEA / CCER / EUA）。
type CarbonQuote struct {
	ID         int64     `json:"id"`
	Product    string    `json:"product"`
	TradeDate  time.Time `json:"trade_date"`
	OpenPrice  *float64  `json:"open_price"`
	HighPrice  *float64  `json:"high_price"`
	LowPrice   *float64  `json:"low_price"`
	ClosePrice *float64  `json:"close_price"`
	Volume     *float64  `json:"volume"`
	Turnover   *float64  `json:"turnover"`
	CreatedAt  time.Time `json:"created_at"`
}

// CarbonProductSummary 单个碳产品的最新行情概览（用于页面顶部卡片）。
type CarbonProductSummary struct {
	Product    string     `json:"product"`
	Name       string     `json:"name"`
	Unit       string     `json:"unit"`
	LatestDate *time.Time `json:"latest_date"`
	Close      *float64   `json:"close"`
	PrevClose  *float64   `json:"prev_close"`
	Change     *float64   `json:"change"`      // 较前一交易日涨跌额
	ChangePct  *float64   `json:"change_pct"`  // 涨跌幅（%）
	Volume     *float64   `json:"volume"`      // 最新交易日成交量
	High52w    *float64   `json:"high_52w"`    // 近一年最高
	Low52w     *float64   `json:"low_52w"`     // 近一年最低
}

// CarbonRepository 碳交易行情仓储。
//
// 碳价为全国统一行情（CEA/CCER 为国内碳市场，EUA 为欧盟碳市场），属共享参考数据，
// 因此本仓储不做省份(org_id)隔离 —— 各省与总部看到同一套碳价，「全部省」视图下亦可读写。
type CarbonRepository struct{ pool *Pool }

func NewCarbonRepository(pool *Pool) *CarbonRepository { return &CarbonRepository{pool: pool} }

// carbonProducts 受支持的碳产品白名单 + 展示元信息（顺序即页面展示顺序）。
var carbonProducts = []struct{ Code, Name, Unit string }{
	{"CEA", "全国碳排放配额", "元/吨"},
	{"CCER", "国家核证自愿减排量", "元/吨"},
	{"EUA", "欧盟碳排放配额", "欧元/吨"},
}

// carbonProductValid 校验产品代码是否在白名单内。
func carbonProductValid(code string) bool {
	for _, p := range carbonProducts {
		if p.Code == code {
			return true
		}
	}
	return false
}

// List 返回指定产品（product 为空表示全部）近 days 天的行情，按日期倒序。
func (r *CarbonRepository) List(ctx context.Context, product string, days int) ([]*CarbonQuote, error) {
	if days <= 0 || days > 3650 {
		days = 180
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, product, trade_date, open_price, high_price, low_price,
	             close_price, volume, turnover, created_at
	      FROM carbon_quotes WHERE trade_date >= $1`
	if product != "" {
		if !carbonProductValid(product) {
			return nil, fmt.Errorf("未知的碳产品: %s", product)
		}
		args = append(args, product)
		q += fmt.Sprintf(" AND product = $%d", len(args))
	}
	q += " ORDER BY trade_date DESC, product"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*CarbonQuote, 0)
	for rows.Next() {
		var c CarbonQuote
		if err := rows.Scan(&c.ID, &c.Product, &c.TradeDate, &c.OpenPrice, &c.HighPrice,
			&c.LowPrice, &c.ClosePrice, &c.Volume, &c.Turnover, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

// Summary 返回三个碳产品的最新行情概览（最新价/涨跌/近一年高低）。
// 即便某产品暂无数据，也会返回一条占位（价格为空），保证页面三卡片齐全。
func (r *CarbonRepository) Summary(ctx context.Context) ([]*CarbonProductSummary, error) {
	// 一次取回每个产品最近两个交易日的收盘价 + 最新日期/成交量。
	type latest struct {
		date      *time.Time
		close     *float64
		volume    *float64
		prevClose *float64
	}
	latestByProduct := map[string]*latest{}
	rows, err := r.pool.Query(ctx, `
		WITH ranked AS (
		    SELECT product, trade_date, close_price, volume,
		           ROW_NUMBER() OVER (PARTITION BY product ORDER BY trade_date DESC) AS rn
		    FROM carbon_quotes
		)
		SELECT product,
		       MAX(trade_date)  FILTER (WHERE rn = 1) AS latest_date,
		       MAX(close_price) FILTER (WHERE rn = 1) AS close,
		       MAX(volume)      FILTER (WHERE rn = 1) AS volume,
		       MAX(close_price) FILTER (WHERE rn = 2) AS prev_close
		FROM ranked WHERE rn <= 2 GROUP BY product`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var p string
		var l latest
		if err := rows.Scan(&p, &l.date, &l.close, &l.volume, &l.prevClose); err != nil {
			rows.Close()
			return nil, err
		}
		lc := l
		latestByProduct[p] = &lc
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 近一年最高 / 最低。
	type hilo struct{ hi, lo *float64 }
	hiloByProduct := map[string]*hilo{}
	yearAgo := time.Now().AddDate(-1, 0, 0)
	rows2, err := r.pool.Query(ctx, `
		SELECT product, MAX(high_price) AS hi, MIN(low_price) AS lo
		FROM carbon_quotes WHERE trade_date >= $1 GROUP BY product`, yearAgo)
	if err != nil {
		return nil, err
	}
	for rows2.Next() {
		var p string
		var h hilo
		if err := rows2.Scan(&p, &h.hi, &h.lo); err != nil {
			rows2.Close()
			return nil, err
		}
		hc := h
		hiloByProduct[p] = &hc
	}
	rows2.Close()
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	out := make([]*CarbonProductSummary, 0, len(carbonProducts))
	for _, p := range carbonProducts {
		s := &CarbonProductSummary{Product: p.Code, Name: p.Name, Unit: p.Unit}
		if l := latestByProduct[p.Code]; l != nil {
			s.LatestDate = l.date
			s.Close = l.close
			s.Volume = l.volume
			s.PrevClose = l.prevClose
			if l.close != nil && l.prevClose != nil {
				chg := round2(*l.close - *l.prevClose)
				s.Change = &chg
				if *l.prevClose != 0 {
					pct := round2(chg / *l.prevClose * 100)
					s.ChangePct = &pct
				}
			}
		}
		if h := hiloByProduct[p.Code]; h != nil {
			s.High52w = h.hi
			s.Low52w = h.lo
		}
		out = append(out, s)
	}
	return out, nil
}

// GenerateDemo 为 CEA/CCER/EUA 各生成约 180 天的演示行情（随机游走 OHLC）。
// 返回写入的记录条数。ON CONFLICT 幂等，可重复调用。
func (r *CarbonRepository) GenerateDemo(ctx context.Context) (int, error) {
	const days = 180
	// 各产品的起步价与日成交量区间（贴近真实量级）。
	base := map[string]float64{"CEA": 75, "CCER": 62, "EUA": 70}
	volRange := map[string][2]float64{
		"CEA":  {200_000, 800_000},
		"CCER": {50_000, 300_000},
		"EUA":  {1_000_000, 5_000_000},
	}

	cnt := 0
	for _, p := range carbonProducts {
		price := base[p.Code]
		// 由最早的一天走到今天，保证价格序列连续。
		for i := days - 1; i >= 0; i-- {
			d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
			// 日收益率约 ±2.5%，含极轻微正向漂移。
			price *= 1 + (rand.Float64()-0.48)*0.05
			if price < 1 {
				price = 1
			}
			closeP := round2(price)
			openP := round2(closeP * (0.99 + rand.Float64()*0.02))
			highP := round2(math.Max(openP, closeP) * (1.0 + rand.Float64()*0.02))
			lowP := round2(math.Min(openP, closeP) * (1.0 - rand.Float64()*0.02))
			vr := volRange[p.Code]
			volume := round2(vr[0] + rand.Float64()*(vr[1]-vr[0]))
			turnover := round2(volume * closeP)
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO carbon_quotes
				   (product, trade_date, open_price, high_price, low_price, close_price, volume, turnover)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
				 ON CONFLICT (product, trade_date) DO UPDATE SET
				   open_price = EXCLUDED.open_price, high_price = EXCLUDED.high_price,
				   low_price = EXCLUDED.low_price, close_price = EXCLUDED.close_price,
				   volume = EXCLUDED.volume, turnover = EXCLUDED.turnover`,
				p.Code, d, openP, highP, lowP, closeP, volume, turnover); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}

// round2 保留两位小数。
func round2(v float64) float64 { return math.Round(v*100) / 100 }
