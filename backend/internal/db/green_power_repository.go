// 绿电交易。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 绿电交易 ───────────────

type GreenPowerTrade struct {
	ID             string    `json:"id"`
	TradeDate      time.Time `json:"trade_date"`
	ProductName    string    `json:"product_name"`
	EnergyMWh      float64   `json:"energy_mwh"`
	Price          float64   `json:"price"`
	Amount         float64   `json:"amount"`
	GreenCertCount int       `json:"green_cert_count"`
	Status         string    `json:"status"`
	Counterparty   string    `json:"counterparty"`
	CreatedAt      time.Time `json:"created_at"`
}

type GreenPowerRepository struct{ pool *Pool }

func NewGreenPowerRepository(pool *Pool) *GreenPowerRepository {
	return &GreenPowerRepository{pool: pool}
}

func (r *GreenPowerRepository) List(ctx context.Context, status string, days int) ([]*GreenPowerTrade, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT id, trade_date, product_name, energy_mwh, price, amount,
			  green_cert_count, status, counterparty, created_at
		  FROM green_power_trades WHERE trade_date >= $1`
	idx := 2
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", idx)
		idx++
	}
	if status != "" {
		args = append(args, status)
		q += fmt.Sprintf(" AND status = $%d", idx)
		idx++
	}
	q += " ORDER BY trade_date DESC LIMIT 200"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*GreenPowerTrade, 0)
	for rows.Next() {
		var g GreenPowerTrade
		if err := rows.Scan(&g.ID, &g.TradeDate, &g.ProductName, &g.EnergyMWh,
			&g.Price, &g.Amount, &g.GreenCertCount, &g.Status, &g.Counterparty,
			&g.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &g)
	}
	return list, rows.Err()
}

func (r *GreenPowerRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用默认组织
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='default'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve default org: %w", err)
		}
	}
	products := []string{"风电-GEC", "光伏-GEC", "水电-GEC", "风电-省内"}
	parties := []string{"广东绿能", "南网综合能源", "华能新能源", "大唐清洁能源"}
	statuses := []string{"completed", "completed", "completed", "pending", "settling"}
	cnt := 0
	for i := 0; i < 30; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		for j := 0; j < 2; j++ {
			energy := 500 + rand.Float64()*3000
			price := 50 + rand.Float64()*30
			amount := energy * price
			certs := int(energy / 1000)
			if certs < 1 {
				certs = 1
			}
			status := statuses[rand.Intn(len(statuses))]
			product := products[rand.Intn(len(products))]
			party := parties[rand.Intn(len(parties))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO green_power_trades
				   (trade_date, product_name, energy_mwh, price, amount,
				    green_cert_count, status, counterparty, org_id)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid)`,
				d, product, energy, price, amount, certs, status, party, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
