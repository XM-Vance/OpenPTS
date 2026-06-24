// 滚动撮合交易。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ─────────────── 滚动撮合交易 ───────────────

type RollingTrade struct {
	ID            string    `json:"id"`
	TradeDate     time.Time `json:"trade_date"`
	TradeSession  int       `json:"trade_session"`
	Side          string    `json:"side"`
	EnergyMW      float64   `json:"energy_mw"`
	DeclaredPrice float64   `json:"declared_price"`
	ClearedPrice  float64   `json:"cleared_price"`
	ClearedEnergy float64   `json:"cleared_energy_mw"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type RollingTradeRepository struct{ pool *Pool }

func NewRollingTradeRepository(pool *Pool) *RollingTradeRepository {
	return &RollingTradeRepository{pool: pool}
}

func (r *RollingTradeRepository) List(ctx context.Context, days, limit int) ([]*RollingTrade, error) {
	if days <= 0 || days > 30 {
		days = 7
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since, limit}
	q := `SELECT id, trade_date, trade_session, side, energy_mw, declared_price,
		        cleared_price, cleared_energy_mw, status, created_at
		 FROM rolling_trades WHERE trade_date >= $1`
	idx := 3
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", idx)
		idx++
	}
	q += fmt.Sprintf(" ORDER BY trade_date DESC, trade_session ASC LIMIT $2")
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*RollingTrade, 0, limit)
	for rows.Next() {
		var t RollingTrade
		if err := rows.Scan(&t.ID, &t.TradeDate, &t.TradeSession, &t.Side,
			&t.EnergyMW, &t.DeclaredPrice, &t.ClearedPrice, &t.ClearedEnergy,
			&t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *RollingTradeRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	statuses := []string{"cleared", "cleared", "cleared", "partial", "uncleared"}
	ph := make([]string, 0, 7*12*2)
	args := make([]any, 0, 7*12*2*9)
	for d := 0; d < 7; d++ {
		date := time.Now().AddDate(0, 0, -d).Truncate(24 * time.Hour)
		for sess := 1; sess <= 12; sess++ {
			for _, side := range []string{"buy", "sell"} {
				energy := 50 + rand.Float64()*200
				declPrice := 380 + rand.Float64()*80
				status := statuses[rand.Intn(len(statuses))]
				var clearedE, clearedP float64
				switch status {
				case "cleared":
					clearedE = energy
					clearedP = declPrice * (0.95 + rand.Float64()*0.1)
				case "partial":
					clearedE = energy * (0.3 + rand.Float64()*0.5)
					clearedP = declPrice * (0.92 + rand.Float64()*0.12)
				default:
					clearedE = 0
					clearedP = 0
				}
				b := len(args)
				ph = append(ph, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
					b+1, b+2, b+3, b+4, b+5, b+6, b+7, b+8, b+9))
				args = append(args, date, sess, side, energy, declPrice, clearedP, clearedE, status, orgID)
			}
		}
	}
	if len(ph) == 0 {
		return 0, nil
	}
	// 单条多行 INSERT 批量写入（替代 7×12×2 次逐条 Exec）。
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO rolling_trades
		   (trade_date, trade_session, side, energy_mw, declared_price,
		    cleared_price, cleared_energy_mw, status, org_id)
		 VALUES `+strings.Join(ph, ","),
		args...); err != nil {
		return 0, err
	}
	return len(ph), nil
}
