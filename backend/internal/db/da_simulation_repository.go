// 日前模拟仓储：场景 CRUD + 时段结果。
package db

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ─────────────── 日前模拟场景 ───────────────

type DASimulationScenario struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	SimDate        time.Time `json:"sim_date"`
	TotalVolumeMWh float64   `json:"total_volume_mwh"`
	AvgPrice       float64   `json:"avg_price"`
	TotalCost      float64   `json:"total_cost"`
	Profit         float64   `json:"profit"`
	Status         string    `json:"status"`
	CreatedBy      *string   `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DASimulationPeriodResult 日前模拟时段结果
type DASimulationPeriodResult struct {
	ID                string    `json:"id"`
	ScenarioID        string    `json:"scenario_id"`
	Period            int       `json:"period"`
	DeclaredVolumeMWh float64   `json:"declared_volume_mwh"`
	SimulatedPrice    float64   `json:"simulated_price"`
	SimulatedCost     float64   `json:"simulated_cost"`
	SpotActualPrice   *float64  `json:"spot_actual_price,omitempty"`
	SettlementAmount  float64   `json:"settlement_amount"`
	CreatedAt         time.Time `json:"created_at"`
}

type DASimulationRepository struct{ pool *Pool }

func NewDASimulationRepository(pool *Pool) *DASimulationRepository {
	return &DASimulationRepository{pool: pool}
}

// CreateScenario 创建模拟场景（草稿状态）
func (r *DASimulationRepository) CreateScenario(ctx context.Context, name, description string, simDate time.Time, userID string) (string, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return "", ErrOrgRequired
	}
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO da_simulation_scenarios (name, description, sim_date, status, created_by, org_id)
		 VALUES ($1, NULLIF($2,''), $3, 'draft', NULLIF($4,'')::uuid, $5::uuid)
		 RETURNING id`,
		name, description, simDate, userID, org).Scan(&id)
	return id, err
}

// ListScenarios 列出模拟场景
func (r *DASimulationRepository) ListScenarios(ctx context.Context, status string, limit int) ([]*DASimulationScenario, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	args := []any{}
	q := `SELECT id, name, COALESCE(description,''), sim_date, total_volume_mwh,
	             avg_price, total_cost, profit, status,
	             COALESCE(created_by::text,''), created_at, updated_at
	      FROM da_simulation_scenarios WHERE 1=1`
	idx := 1
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", idx)
		idx++
	}
	if status != "" {
		args = append(args, status)
		q += " AND status = $" + itoaNew(idx)
		idx++
	}
	q += " ORDER BY created_at DESC LIMIT $" + itoaNew(idx)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DASimulationScenario, 0)
	for rows.Next() {
		var s DASimulationScenario
		var createdBy string
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.SimDate,
			&s.TotalVolumeMWh, &s.AvgPrice, &s.TotalCost, &s.Profit,
			&s.Status, &createdBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if createdBy != "" {
			s.CreatedBy = &createdBy
		}
		list = append(list, &s)
	}
	return list, rows.Err()
}

// GetScenario 获取单个场景
func (r *DASimulationRepository) GetScenario(ctx context.Context, id string) (*DASimulationScenario, error) {
	var s DASimulationScenario
	var createdBy string
	args := []any{id}
	q := `SELECT id, name, COALESCE(description,''), sim_date, total_volume_mwh,
		        avg_price, total_cost, profit, status,
		        COALESCE(created_by::text,''), created_at, updated_at
		 FROM da_simulation_scenarios WHERE id = $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	err := r.pool.QueryRow(ctx, q, args...).Scan(
		&s.ID, &s.Name, &s.Description, &s.SimDate,
		&s.TotalVolumeMWh, &s.AvgPrice, &s.TotalCost, &s.Profit,
		&s.Status, &createdBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if createdBy != "" {
		s.CreatedBy = &createdBy
	}
	return &s, nil
}

// GetPeriodResults 获取场景时段结果
// period_results 通过 scenario_id JOIN scenarios 做 org 过滤，无需自身 org_id。
func (r *DASimulationRepository) GetPeriodResults(ctx context.Context, scenarioID string) ([]*DASimulationPeriodResult, error) {
	args := []any{scenarioID}
	q := `SELECT pr.id, pr.scenario_id::text, pr.period, pr.declared_volume_mwh,
	        pr.simulated_price, pr.simulated_cost, pr.spot_actual_price,
	        pr.settlement_amount, pr.created_at
		 FROM da_simulation_period_results pr
		 JOIN da_simulation_scenarios sc ON sc.id = pr.scenario_id
		 WHERE pr.scenario_id = $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND sc.org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " ORDER BY pr.period ASC"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DASimulationPeriodResult, 0)
	for rows.Next() {
		var p DASimulationPeriodResult
		if err := rows.Scan(&p.ID, &p.ScenarioID, &p.Period,
			&p.DeclaredVolumeMWh, &p.SimulatedPrice, &p.SimulatedCost,
			&p.SpotActualPrice, &p.SettlementAmount, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// DeleteScenario 删除场景
func (r *DASimulationRepository) DeleteScenario(ctx context.Context, id string) error {
	args := []any{id}
	q := `DELETE FROM da_simulation_scenarios WHERE id = $1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

// SaveSimulationResults 保存模拟计算结果并更新场景汇总
func (r *DASimulationRepository) SaveSimulationResults(ctx context.Context, scenarioID string, results []*DASimulationPeriodResult, avgPrice, totalCost, profit float64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 删除旧结果（通过 scenario_id 关联，不需要单独 org 过滤）
	if _, err := tx.Exec(ctx, `DELETE FROM da_simulation_period_results WHERE scenario_id = $1`, scenarioID); err != nil {
		return err
	}

	// 单条多行 INSERT 批量写入时段结果（替代逐条 Exec；最多 96 个时段 → 1 次往返）。
	var totalVol float64
	if len(results) > 0 {
		const cols = 7
		placeholders := make([]string, 0, len(results))
		args := make([]any, 0, len(results)*cols)
		for i, p := range results {
			b := i * cols
			placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				b+1, b+2, b+3, b+4, b+5, b+6, b+7))
			args = append(args, scenarioID, p.Period, p.DeclaredVolumeMWh, p.SimulatedPrice,
				p.SimulatedCost, p.SpotActualPrice, p.SettlementAmount)
			totalVol += p.DeclaredVolumeMWh
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO da_simulation_period_results
			   (scenario_id, period, declared_volume_mwh, simulated_price,
			    simulated_cost, spot_actual_price, settlement_amount)
			 VALUES `+strings.Join(placeholders, ","),
			args...); err != nil {
			return err
		}
	}

	// 更新场景汇总 + 状态（org 过滤通过 scenario_id 归属保证）
	if _, err := tx.Exec(ctx,
		`UPDATE da_simulation_scenarios
		    SET total_volume_mwh = $2, avg_price = $3, total_cost = $4,
		        profit = $5, status = 'submitted', updated_at = now()
		  WHERE id = $1`,
		scenarioID, totalVol, avgPrice, totalCost, profit); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GenerateDemo 生成演示数据
func (r *DASimulationRepository) GenerateDemo(ctx context.Context) (int, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return 0, ErrOrgRequired
	}
	statuses := []string{"draft", "submitted", "submitted", "settled", "settled"}
	cnt := 0
	for i := 0; i < 10; i++ {
		d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		status := statuses[rand.Intn(len(statuses))]

		var id string
		err := r.pool.QueryRow(ctx,
			`INSERT INTO da_simulation_scenarios (name, description, sim_date, status, org_id)
			 VALUES ($1,$2,$3,$4,$5::uuid) RETURNING id`,
			"模拟场景 "+d.Format("01-02"), "自动生成的模拟场景", d, status, org).Scan(&id)
		if err != nil {
			return cnt, err
		}

		// 批量构造 96 个时段，单条多行 INSERT 写入（替代逐条 Exec）。
		var totalVol, totalCost, totalProfit float64
		ph := make([]string, 0, 96)
		args := make([]any, 0, 96*7)
		for period := 1; period <= 96; period++ {
			vol := 15 + rand.Float64()*35
			simPrice := 250 + rand.Float64()*200
			simCost := vol * simPrice
			var spotPrice *float64
			var settlement float64
			if status == "settled" || status == "submitted" {
				sp := simPrice * (0.85 + rand.Float64()*0.3)
				spotPrice = &sp
				settlement = vol * (sp - simPrice)
			}
			b := len(args)
			ph = append(ph, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				b+1, b+2, b+3, b+4, b+5, b+6, b+7))
			args = append(args, id, period, vol, simPrice, simCost, spotPrice, settlement)
			totalVol += vol
			totalCost += simCost
			if settlement != 0 {
				totalProfit += settlement
			}
		}
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO da_simulation_period_results
			   (scenario_id, period, declared_volume_mwh, simulated_price,
			    simulated_cost, spot_actual_price, settlement_amount)
			 VALUES `+strings.Join(ph, ","),
			args...); err != nil {
			return cnt, err
		}

		avgPrice := 0.0
		if totalVol > 0 {
			avgPrice = totalCost / totalVol
		}
		if _, err := r.pool.Exec(ctx,
			`UPDATE da_simulation_scenarios
			    SET total_volume_mwh = $2, avg_price = $3, total_cost = $4,
			        profit = $5, updated_at = now()
			  WHERE id = $1`,
			id, totalVol, avgPrice, totalCost, totalProfit); err != nil {
			return cnt, err
		}
		cnt++
	}
	return cnt, nil
}
