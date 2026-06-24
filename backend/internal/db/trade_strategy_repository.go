// 交易策略。
// 2026-06 自 new_modules2_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 交易策略 ───────────────

type TradeStrategy struct {
	ID             string          `json:"id"`
	StrategyName   string          `json:"strategy_name"`
	StrategyType   string          `json:"strategy_type"`
	TargetMarket   string          `json:"target_market"`
	Parameters     json.RawMessage `json:"parameters"`
	Status         string          `json:"status"`
	BacktestReturn *float64        `json:"backtest_return,omitempty"`
	BacktestSharpe *float64        `json:"backtest_sharpe,omitempty"`
	WinRate        *float64        `json:"win_rate,omitempty"`
	TotalTrades    int             `json:"total_trades"`
	Note           *string         `json:"note,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type TradeStrategyRepository struct{ pool *Pool }

func NewTradeStrategyRepository(pool *Pool) *TradeStrategyRepository {
	return &TradeStrategyRepository{pool: pool}
}

func (r *TradeStrategyRepository) List(ctx context.Context) ([]*TradeStrategy, error) {
	org, scoped := OrgFilter(ctx)
	args := []any{}
	q := `SELECT id, strategy_name, strategy_type, target_market, parameters,
	       status, backtest_return, backtest_sharpe, win_rate, total_trades,
	       note, created_at, updated_at
	      FROM trade_strategies`
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY updated_at DESC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*TradeStrategy, 0)
	for rows.Next() {
		var s TradeStrategy
		if err := rows.Scan(&s.ID, &s.StrategyName, &s.StrategyType, &s.TargetMarket,
			&s.Parameters, &s.Status, &s.BacktestReturn, &s.BacktestSharpe,
			&s.WinRate, &s.TotalTrades, &s.Note, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

type TradeStrategyInput struct {
	StrategyName string
	StrategyType string
	TargetMarket string
	Parameters   json.RawMessage
	Status       string
	Note         string
}

func (r *TradeStrategyRepository) Create(ctx context.Context, in TradeStrategyInput) (string, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return "", ErrOrgRequired
	}
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO trade_strategies
		   (strategy_name, strategy_type, target_market, parameters, status, note, org_id)
		 VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7::uuid)
		 RETURNING id`,
		in.StrategyName, in.StrategyType, in.TargetMarket,
		in.Parameters, in.Status, in.Note, org).Scan(&id)
	return id, err
}

func (r *TradeStrategyRepository) GenerateDemo(ctx context.Context) (int, error) {
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	strategies := []struct {
		name, stype, market string
	}{
		{"日间价差套利", "arbitrage", "day_ahead"},
		{"峰谷套利", "arbitrage", "spot"},
		{"保守申报策略", "conservative", "day_ahead"},
		{"激进竞价策略", "aggressive", "rolling"},
		{"绿电配比优化", "optimization", "green"},
		{"需求响应调度", "dispatch", "ancillary"},
	}
	cnt := 0
	for _, s := range strategies {
		ret := rand.Float64() * 15
		sharpe := 0.5 + rand.Float64()*2
		win := 55 + rand.Float64()*30
		trades := 50 + rand.Intn(200)
		params, _ := json.Marshal(map[string]any{"version": "v1.0"})
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO trade_strategies
			   (strategy_name, strategy_type, target_market, parameters, status,
			    backtest_return, backtest_sharpe, win_rate, total_trades, org_id)
			 VALUES ($1,$2,$3,$4,'active',$5,$6,$7,$8,$9::uuid)
			 ON CONFLICT (org_id, strategy_name) DO UPDATE SET
			   strategy_type = EXCLUDED.strategy_type, target_market = EXCLUDED.target_market,
			   parameters = EXCLUDED.parameters, status = EXCLUDED.status,
			   backtest_return = EXCLUDED.backtest_return, backtest_sharpe = EXCLUDED.backtest_sharpe,
			   win_rate = EXCLUDED.win_rate, total_trades = EXCLUDED.total_trades`,
			s.name, s.stype, s.market, params, ret, sharpe, win, trades, orgID); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
