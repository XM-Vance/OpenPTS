// 市场分析。
// 2026-06 自 p1_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ─────────────── V5 市场分析 ───────────────

type MarketAnalysis struct {
	ID              string    `json:"id"`
	TradeDate       time.Time `json:"trade_date"`
	HighPrice       float64   `json:"high_price"`
	LowPrice        float64   `json:"low_price"`
	AvgPrice        float64   `json:"avg_price"`
	Volatility      float64   `json:"volatility"`
	TotalVolumeMWh  float64   `json:"total_volume_mwh"`
	PeakValleyGap   float64   `json:"peak_valley_gap"`
	MarketSentiment *string   `json:"market_sentiment,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type MarketAnalysisRepository struct{ pool *Pool }

func NewMarketAnalysisRepository(pool *Pool) *MarketAnalysisRepository {
	return &MarketAnalysisRepository{pool: pool}
}

func (r *MarketAnalysisRepository) List(ctx context.Context, days int) ([]*MarketAnalysis, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `
		SELECT id, trade_date, high_price, low_price, avg_price, volatility,
		       total_volume_mwh, peak_valley_gap, market_sentiment, created_at
		FROM market_analysis_daily WHERE trade_date >= $1`
	n := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", n)
		n++
	}
	q += " ORDER BY trade_date DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MarketAnalysis, 0)
	for rows.Next() {
		var m MarketAnalysis
		if err := rows.Scan(&m.ID, &m.TradeDate, &m.HighPrice, &m.LowPrice,
			&m.AvgPrice, &m.Volatility, &m.TotalVolumeMWh, &m.PeakValleyGap,
			&m.MarketSentiment, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *MarketAnalysisRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	sentiments := []string{"bullish", "neutral", "bearish"}
	cnt := 0
	for i := 0; i < 30; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		avg := 380 + rand.Float64()*120
		high := avg * (1.1 + rand.Float64()*0.3)
		low := avg * (0.6 + rand.Float64()*0.2)
		vol := (high - low) / avg
		volume := 350000 + rand.Float64()*100000
		gap := high - low
		sent := sentiments[rand.Intn(len(sentiments))]
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO market_analysis_daily
			(trade_date, high_price, low_price, avg_price, volatility, total_volume_mwh,
			 peak_valley_gap, market_sentiment, org_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid)
			ON CONFLICT (org_id, trade_date) DO UPDATE SET
			  high_price = EXCLUDED.high_price, low_price = EXCLUDED.low_price,
			  avg_price = EXCLUDED.avg_price, volatility = EXCLUDED.volatility,
			  total_volume_mwh = EXCLUDED.total_volume_mwh,
			  peak_valley_gap = EXCLUDED.peak_valley_gap,
			  market_sentiment = EXCLUDED.market_sentiment`,
			d, high, low, avg, vol, volume, gap, sent, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}

// 占位避免未用 imports（部分场景）
var _ = strings.TrimSpace
