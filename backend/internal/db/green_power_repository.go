// 绿电交易。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// ─────────────── 绿电交易 ───────────────

type GreenPowerTrade struct {
	ID             string          `json:"id"`
	OrgID          string          `json:"org_id,omitempty"`
	TradeDate      time.Time       `json:"trade_date"`
	ProductName    string          `json:"product_name"`
	EnergyMWh      float64         `json:"energy_mwh"`
	Price          decimal.Decimal `json:"price"`  // P4: numeric(18,4)
	Amount         decimal.Decimal `json:"amount"` // P4: numeric(18,4)
	GreenCertCount int             `json:"green_cert_count"`
	Status         string          `json:"status"`
	Counterparty   string          `json:"counterparty"`
	CreatedAt      time.Time       `json:"created_at"`
}

var ErrGreenPowerTradeNotFound = errors.New("绿电交易记录不存在")

// GreenPowerTradeInput 创建入参。Price/Amount 为 numeric(18,4)。
type GreenPowerTradeInput struct {
	TradeDate      time.Time
	ProductName    string
	EnergyMWh      float64
	Price          decimal.Decimal
	Amount         decimal.Decimal
	GreenCertCount int
	Status         string
	Counterparty   string
}

type GreenPowerRepository struct{ pool *Pool }

func NewGreenPowerRepository(pool *Pool) *GreenPowerRepository {
	return &GreenPowerRepository{pool: pool}
}

const greenPowerColumns = "id, org_id, trade_date, product_name, energy_mwh, price, amount, green_cert_count, status, counterparty, created_at"

func (r *GreenPowerRepository) scan(row pgx.Row) (*GreenPowerTrade, error) {
	var g GreenPowerTrade
	err := row.Scan(&g.ID, &g.OrgID, &g.TradeDate, &g.ProductName, &g.EnergyMWh,
		&g.Price, &g.Amount, &g.GreenCertCount, &g.Status, &g.Counterparty, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGreenPowerTradeNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *GreenPowerRepository) List(ctx context.Context, status string, days int) ([]*GreenPowerTrade, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT ` + greenPowerColumns + `
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
		g, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

func (r *GreenPowerRepository) Create(ctx context.Context, in GreenPowerTradeInput) (*GreenPowerTrade, error) {
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	status := in.Status
	if status == "" {
		status = "pending"
	}
	q := `INSERT INTO green_power_trades
		(org_id, trade_date, product_name, energy_mwh, price, amount,
		 green_cert_count, status, counterparty)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING ` + greenPowerColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		org, in.TradeDate, in.ProductName, in.EnergyMWh, in.Price, in.Amount,
		in.GreenCertCount, status, in.Counterparty))
}

func (r *GreenPowerRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM green_power_trades WHERE id = $1::uuid`
	args := []any{id}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrGreenPowerTradeNotFound
	}
	return nil
}

func (r *GreenPowerRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃组织，否则用默认组织
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
			// P4: 金额 decimal（amount = energy*price）；电量保持 float。
			price := decimal.NewFromFloat(50 + rand.Float64()*30).Round(4)
			amount := decimal.NewFromFloat(energy).Mul(price).Round(4)
			certs := int(energy / 1000)
			if certs < 1 {
				certs = 1
			}
			status := statuses[rand.IntN(len(statuses))]
			product := products[rand.IntN(len(products))]
			party := parties[rand.IntN(len(parties))]
			if _, err := r.pool.Exec(ctx,
				`INSERT INTO green_power_trades
				   (trade_date, product_name, energy_mwh, price, amount,
				    green_cert_count, status, counterparty, org_id, is_demo)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid,TRUE)`,
				d, product, energy, price, amount, certs, status, party, orgID); err != nil {
				return cnt, err
			}
			cnt++
		}
	}
	return cnt, nil
}
