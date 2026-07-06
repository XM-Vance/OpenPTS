// 竞价管理。
// 2026-06 自 new_modules2_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

// ─────────────── 竞价管理 ───────────────

type BiddingRecord struct {
	ID             string          `json:"id"`
	TradeDate      time.Time       `json:"trade_date"`
	BiddingSession string          `json:"bidding_session"`
	DeclaredMW     float64         `json:"declared_mw"`
	DeclaredPrice  float64         `json:"declared_price"`
	ClearedMW      float64         `json:"cleared_mw"`
	ClearedPrice   float64         `json:"cleared_price"`
	Status         string          `json:"status"`
	Strategy       string          `json:"strategy"`
	BidCurve       json.RawMessage `json:"bid_curve,omitempty"`
	Note           *string         `json:"note,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type BiddingRepository struct{ pool *Pool }

func NewBiddingRepository(pool *Pool) *BiddingRepository {
	return &BiddingRepository{pool: pool}
}

// BiddingStatistics 对齐前端契约（bidding/_view.tsx:104-106）：
// strategyCount/avgAccuracy/simulatedProfit。
// 注：bidding_records 是申报流水，前端期望的是策略回测视图；
// strategyCount 用不同策略数，avgAccuracy 用中标率近似，simulatedProfit 无数据源暂为 0。
type BiddingStatistics struct {
	StrategyCount   int            `json:"strategyCount"`   // 不同策略数
	AvgAccuracy     float64        `json:"avgAccuracy"`     // 中标率（近似策略准确度，0-100）
	SimulatedProfit float64        `json:"simulatedProfit"` // 模拟收益（无数据源，0）
	TotalBids       int            `json:"total_bids"`      // 总申报数
	ClearedCount    int            `json:"cleared_count"`   // 中标数
	ByStrategy      map[string]int `json:"by_strategy"`
}

// Statistics 聚合竞价统计（对齐前端 strategyCount/avgAccuracy/simulatedProfit）。
func (r *BiddingRepository) Statistics(ctx context.Context, days int) (*BiddingStatistics, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	// B4：两个查询都补 OrgFilter（与同文件 List 一致），避免跨租户统计泄漏
	org, scoped := OrgFilter(ctx)
	args := []any{since}
	q := `SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status='cleared'),
			COUNT(DISTINCT strategy)
		  FROM bidding_records WHERE trade_date >= $1`
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	var s BiddingStatistics
	if err := r.pool.QueryRow(ctx, q, args...).Scan(
		&s.TotalBids, &s.ClearedCount, &s.StrategyCount); err != nil {
		return nil, err
	}
	if s.TotalBids > 0 {
		s.AvgAccuracy = float64(s.ClearedCount) / float64(s.TotalBids) * 100
	}
	// 按策略分组（同样补 OrgFilter，复用上面的 org/scoped）
	q2 := `SELECT strategy, COUNT(*) FROM bidding_records WHERE trade_date >= $1`
	var rows pgx.Rows
	var err error
	if scoped {
		q2 += " AND org_id = $2::uuid GROUP BY strategy"
		rows, err = r.pool.Query(ctx, q2, since, org)
	} else {
		q2 += " GROUP BY strategy"
		rows, err = r.pool.Query(ctx, q2, since)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	s.ByStrategy = make(map[string]int)
	for rows.Next() {
		var strat string
		var n int
		if err := rows.Scan(&strat, &n); err != nil {
			return nil, err
		}
		s.ByStrategy[strat] = n
	}
	return &s, rows.Err()
}

func (r *BiddingRepository) List(ctx context.Context, days int) ([]*BiddingRecord, error) {
	if days <= 0 || days > 60 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, trade_date, bidding_session, declared_mw, declared_price,
	       cleared_mw, cleared_price, status, strategy, bid_curve, note, created_at
	FROM bidding_records WHERE trade_date >= $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY trade_date DESC, bidding_session ASC LIMIT 200"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*BiddingRecord, 0)
	for rows.Next() {
		var b BiddingRecord
		if err := rows.Scan(&b.ID, &b.TradeDate, &b.BiddingSession, &b.DeclaredMW,
			&b.DeclaredPrice, &b.ClearedMW, &b.ClearedPrice, &b.Status, &b.Strategy,
			&b.BidCurve, &b.Note, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &b)
	}
	return list, rows.Err()
}

type BiddingInput struct {
	TradeDate      time.Time
	BiddingSession string
	DeclaredMW     float64
	DeclaredPrice  float64
	Strategy       string
	Note           string
}

func (r *BiddingRepository) Create(ctx context.Context, in BiddingInput) (string, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return "", err
	}
	var id string
	err = r.pool.QueryRow(ctx,
		`INSERT INTO bidding_records
		   (trade_date, bidding_session, declared_mw, declared_price, strategy, status, note, org_id)
		 VALUES ($1,$2,$3,$4,$5,'pending',NULLIF($6,''),$7::uuid)
		 RETURNING id`,
		in.TradeDate, in.BiddingSession, in.DeclaredMW, in.DeclaredPrice,
		in.Strategy, in.Note, org).Scan(&id)
	return id, err
}

func (r *BiddingRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	sessions := []string{"morning", "afternoon", "evening"}
	strategies := []string{"aggressive", "conservative", "balanced", "marginal"}
	statuses := []string{"cleared", "cleared", "cleared", "partial", "uncleared"}
	cnt := 0
	for i := 0; i < 14; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		for _, sess := range sessions {
			declMW := 100 + rand.Float64()*300
			declPrice := 380 + rand.Float64()*100
			status := statuses[rand.IntN(len(statuses))]
			var clrMW, clrPrice float64
			switch status {
			case "cleared":
				clrMW = declMW
				clrPrice = declPrice * (0.95 + rand.Float64()*0.1)
			case "partial":
				clrMW = declMW * (0.3 + rand.Float64()*0.5)
				clrPrice = declPrice * (0.92 + rand.Float64()*0.12)
			default:
				clrMW = 0
				clrPrice = 0
			}
			strategy := strategies[rand.IntN(len(strategies))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO bidding_records
				   (trade_date, bidding_session, declared_mw, declared_price,
				    cleared_mw, cleared_price, status, strategy, org_id, is_demo)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid,TRUE)
				 ON CONFLICT (org_id, trade_date, bidding_session) DO UPDATE SET
				   declared_mw = EXCLUDED.declared_mw, declared_price = EXCLUDED.declared_price,
				   cleared_mw = EXCLUDED.cleared_mw, cleared_price = EXCLUDED.cleared_price,
				   status = EXCLUDED.status, strategy = EXCLUDED.strategy, is_demo = TRUE`,
				d, sess, declMW, declPrice, clrMW, clrPrice, status, strategy, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
