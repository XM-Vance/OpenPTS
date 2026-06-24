// 滚动撮合报价仓储。
// 2026-06 自 v1clone_f_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"math/rand"
	"time"
)

// ─────────────── F3 滚动撮合报价 ───────────────

type MatchQuote struct {
	ID            string    `json:"id"`
	MatchDate     time.Time `json:"match_date"`
	MatchSession  int       `json:"match_session"`
	Side          string    `json:"side"`
	DeclaredMW    float64   `json:"declared_mw"`
	ClearedMW     float64   `json:"cleared_mw"`
	DeclaredPrice float64   `json:"declared_price"`
	ClearedPrice  float64   `json:"cleared_price"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type MatchQuoteRepository struct{ pool *Pool }

func NewMatchQuoteRepository(pool *Pool) *MatchQuoteRepository {
	return &MatchQuoteRepository{pool: pool}
}

func (r *MatchQuoteRepository) List(ctx context.Context, days, limit int) ([]*MatchQuote, error) {
	if days <= 0 || days > 30 {
		days = 7
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	since := time.Now().AddDate(0, 0, -days)
	rows, err := r.pool.Query(ctx,
		`SELECT id, match_date, match_session, side, declared_mw, cleared_mw,
			declared_price, cleared_price, status, created_at
		 FROM rolling_match_quotes WHERE match_date >= $1
		 ORDER BY match_date DESC, match_session ASC LIMIT $2`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*MatchQuote, 0, limit)
	for rows.Next() {
		var q MatchQuote
		if err := rows.Scan(&q.ID, &q.MatchDate, &q.MatchSession, &q.Side,
			&q.DeclaredMW, &q.ClearedMW, &q.DeclaredPrice, &q.ClearedPrice,
			&q.Status, &q.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &q)
	}
	return list, rows.Err()
}

func (r *MatchQuoteRepository) GenerateDemo(ctx context.Context) (int, error) {
	statuses := []string{"cleared", "cleared", "cleared", "partial", "failed"}
	cnt := 0
	for d := 0; d < 7; d++ {
		date := time.Now().AddDate(0, 0, -d).Truncate(24 * time.Hour)
		for sess := 1; sess <= 12; sess++ {
			for _, side := range []string{"buy", "sell"} {
				decl := 50 + rand.Float64()*200
				declPrice := 380 + rand.Float64()*80
				status := statuses[rand.Intn(len(statuses))]
				var cleared, clearedPrice float64
				switch status {
				case "cleared":
					cleared = decl
					clearedPrice = declPrice * (0.95 + rand.Float64()*0.1)
				case "partial":
					cleared = decl * (0.3 + rand.Float64()*0.5)
					clearedPrice = declPrice * (0.92 + rand.Float64()*0.12)
				default:
					cleared = 0
					clearedPrice = 0
				}
				if _, err := r.pool.Exec(ctx,
					`INSERT INTO rolling_match_quotes
					   (match_date, match_session, side, declared_mw, cleared_mw,
					    declared_price, cleared_price, status)
					 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
					date, sess, side, decl, cleared, declPrice, clearedPrice, status); err != nil {
					return cnt, err
				}
				cnt++
			}
		}
	}
	return cnt, nil
}
