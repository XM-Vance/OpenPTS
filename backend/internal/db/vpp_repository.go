// 虚拟电厂。
// 2026-06 自 new_modules_repository.go 按域拆分迁移（纯移动，无逻辑变更）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// ─────────────── 虚拟电厂 ───────────────

type VPPResource struct {
	ID           string          `json:"id"`
	ResourceName string          `json:"resource_name"`
	ResourceType string          `json:"resource_type"`
	CapacityMW   float64         `json:"capacity_mw"`
	Status       string          `json:"status"`
	Location     string          `json:"location"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type VPPDispatch struct {
	ID              string    `json:"id"`
	DispatchDate    time.Time `json:"dispatch_date"`
	ResourceID      string    `json:"resource_id"`
	ResourceName    string    `json:"resource_name,omitempty"`
	DispatchType    string    `json:"dispatch_type"`
	DispatchedMW    float64   `json:"dispatched_mw"`
	DurationMin     int       `json:"duration_min"`
	ResponseTimeSec int       `json:"response_time_sec"`
	Revenue         float64   `json:"revenue"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type VPPRepository struct{ pool *Pool }

func NewVPPRepository(pool *Pool) *VPPRepository {
	return &VPPRepository{pool: pool}
}

func (r *VPPRepository) ListResources(ctx context.Context) ([]*VPPResource, error) {
	q := `SELECT id, resource_name, resource_type, capacity_mw, status, location, metadata, created_at
		FROM vpp_resources`
	args := []any{}
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" WHERE org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY resource_type ASC, resource_name ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*VPPResource, 0)
	for rows.Next() {
		var v VPPResource
		if err := rows.Scan(&v.ID, &v.ResourceName, &v.ResourceType, &v.CapacityMW,
			&v.Status, &v.Location, &v.Metadata, &v.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

func (r *VPPRepository) ListDispatches(ctx context.Context, days int) ([]*VPPDispatch, error) {
	if days <= 0 || days > 60 {
		days = 14
	}
	since := time.Now().AddDate(0, 0, -days)
	args := []any{since}
	q := `SELECT d.id, d.dispatch_date, d.resource_id::text, r.resource_name,
	       d.dispatch_type, d.dispatched_mw, d.duration_min, d.response_time_sec,
	       d.revenue, d.status, d.created_at
	FROM vpp_dispatches d
	JOIN vpp_resources r ON r.id = d.resource_id
	WHERE d.dispatch_date >= $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND d.org_id = $%d::uuid", len(args))
	}
	q += " ORDER BY d.dispatch_date DESC LIMIT 200"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*VPPDispatch, 0)
	for rows.Next() {
		var d VPPDispatch
		if err := rows.Scan(&d.ID, &d.DispatchDate, &d.ResourceID, &d.ResourceName,
			&d.DispatchType, &d.DispatchedMW, &d.DurationMin, &d.ResponseTimeSec,
			&d.Revenue, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *VPPRepository) GenerateDemo(ctx context.Context) (int, error) {
	// 确定 org_id：scoped 用活跃省，否则用 FJ
	org, scoped := OrgFilter(ctx)
	orgID := org
	if !scoped {
		if err := r.pool.QueryRow(ctx,
			"SELECT id FROM organizations WHERE code='FJ'").Scan(&orgID); err != nil {
			return 0, fmt.Errorf("resolve FJ org: %w", err)
		}
	}
	resources := []struct {
		name, rtype, location string
		cap                   float64
	}{
		{"工业园储能A", "storage", "佛山", 10},
		{"商业楼宇群", "demand_response", "广州", 5},
		{"光伏聚合1号", "solar", "深圳", 8},
		{"充电桩聚合", "ev_charger", "东莞", 3},
		{"分布式储能B", "storage", "中山", 6},
	}
	resIDs := make([]string, 0, len(resources))
	for _, res := range resources {
		var id string
		if err := r.pool.QueryRow(ctx,
			`INSERT INTO vpp_resources (resource_name, resource_type, capacity_mw, status, location, metadata, org_id)
			 VALUES ($1,$2,$3,'active',$4,'{}'::jsonb,$5::uuid)
			 ON CONFLICT (org_id, resource_name) DO UPDATE SET capacity_mw = EXCLUDED.capacity_mw, status = 'active'
			 RETURNING id`, res.name, res.rtype, res.cap, res.location, orgID).Scan(&id); err != nil {
			return 0, err
		}
		resIDs = append(resIDs, id)
	}
	dispatchTypes := []string{"peak_shaving", "frequency_reg", "energy_arbitrage", "demand_response"}
	statuses := []string{"completed", "completed", "completed", "scheduled", "failed"}
	cnt := len(resIDs)
	for i := 0; i < 14; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		for _, rid := range resIDs {
			for j := 0; j < 3; j++ {
				dtype := dispatchTypes[rand.Intn(len(dispatchTypes))]
				dispMW := 2 + rand.Float64()*8
				dur := 15 + rand.Intn(180)
				resp := 1 + rand.Intn(300)
				rev := dispMW * float64(dur) / 60 * (300 + rand.Float64()*200)
				status := statuses[rand.Intn(len(statuses))]
				if _, err := r.pool.Exec(ctx,
					`INSERT INTO vpp_dispatches
					   (dispatch_date, resource_id, dispatch_type, dispatched_mw, duration_min,
					    response_time_sec, revenue, status, org_id)
					 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid)`,
					d, rid, dtype, dispMW, dur, resp, rev, status, orgID); err != nil {
					return cnt, err
				}
				cnt++
			}
		}
	}
	return cnt, nil
}
