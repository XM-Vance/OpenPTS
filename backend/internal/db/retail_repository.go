// 零售业务仓储：定价模型（只读）、零售套餐（CRUD）、零售合同（CRUD）。
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrPackageNotFound  = errors.New("零售套餐不存在")
	ErrContractNotFound = errors.New("零售合同不存在")
)

// ═══════ 定价模型 ═══════

type PricingModel struct {
	Code        string `json:"code"`
	DisplayName string `json:"display_name"`
	PackageType string `json:"package_type"`
	PricingMode string `json:"pricing_mode"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int    `json:"sort_order"`
}

// ═══════ 零售套餐 ═══════

type RetailPackage struct {
	ID           uuid.UUID `json:"id"`
	PackageName  string    `json:"package_name"`
	PackageType  string    `json:"package_type"`
	ModelCode    *string   `json:"model_code,omitempty"`
	IsGreenPower bool      `json:"is_green_power"`
	Status       string    `json:"status"`
	Description  *string   `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PackageInput struct {
	PackageName  string
	PackageType  string
	ModelCode    string
	IsGreenPower bool
	Status       string
	Description  string
}

// ═══════ 零售合同 ═══════

type RetailContract struct {
	ID                  uuid.UUID `json:"id"`
	CustomerID          uuid.UUID `json:"customer_id"`
	CustomerName        string    `json:"customer_name"`
	PackageID           uuid.UUID `json:"package_id"`
	PackageNameSnapshot string    `json:"package_name_snapshot"`
	PurchasingEnergyMWH float64   `json:"purchasing_energy_mwh"`
	GreenPowerRatio     *float64  `json:"green_power_ratio,omitempty"`
	PurchaseStartMonth  string    `json:"purchase_start_month"`
	PurchaseEndMonth    string    `json:"purchase_end_month"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ContractInput struct {
	CustomerID          uuid.UUID
	PackageID           uuid.UUID
	PurchasingEnergyMWH float64
	GreenPowerRatio     *float64
	PurchaseStartMonth  string
	PurchaseEndMonth    string
	Status              string
}

type RetailRepository struct {
	pool *Pool
}

func NewRetailRepository(pool *Pool) *RetailRepository {
	return &RetailRepository{pool: pool}
}

// ─── 定价模型 ───

func (r *RetailRepository) ListPricingModels(ctx context.Context) ([]*PricingModel, error) {
	const q = `SELECT code, display_name, package_type, pricing_mode, enabled, sort_order
		FROM pricing_models ORDER BY sort_order, code`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*PricingModel, 0, 8)
	for rows.Next() {
		var m PricingModel
		if err := rows.Scan(&m.Code, &m.DisplayName, &m.PackageType, &m.PricingMode,
			&m.Enabled, &m.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

// ─── 零售套餐 ───

const packageColumns = "id, package_name, package_type, model_code, is_green_power, status, description, created_at, updated_at"

func (r *RetailRepository) scanPackage(row pgx.Row) (*RetailPackage, error) {
	var p RetailPackage
	err := row.Scan(&p.ID, &p.PackageName, &p.PackageType, &p.ModelCode,
		&p.IsGreenPower, &p.Status, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPackageNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *RetailRepository) ListPackages(ctx context.Context, keyword, status string) ([]*RetailPackage, error) {
	where := make([]string, 0, 3)
	args := make([]any, 0, 3)
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf("package_name ILIKE $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	q := "SELECT " + packageColumns + " FROM retail_packages"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*RetailPackage, 0, 16)
	for rows.Next() {
		p, err := r.scanPackage(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *RetailRepository) CreatePackage(ctx context.Context, in PackageInput, createdBy *uuid.UUID) (*RetailPackage, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	q := `INSERT INTO retail_packages
		(package_name, package_type, model_code, is_green_power, status, description, created_by, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::uuid) RETURNING ` + packageColumns
	return r.scanPackage(r.pool.QueryRow(ctx, q,
		in.PackageName, in.PackageType, nullStr(in.ModelCode), in.IsGreenPower,
		status, nullStr(in.Description), createdBy, org))
}

func (r *RetailRepository) UpdatePackage(ctx context.Context, id uuid.UUID, in PackageInput) (*RetailPackage, error) {
	q := `UPDATE retail_packages SET
		package_name=$2, package_type=$3, model_code=$4, is_green_power=$5, status=$6, description=$7
		WHERE id=$1`
	args := []any{id}
	status := in.Status
	if status == "" {
		status = "active"
	}
	args = append(args,
		in.PackageName, in.PackageType, nullStr(in.ModelCode), in.IsGreenPower,
		status, nullStr(in.Description))
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " RETURNING " + packageColumns
	return r.scanPackage(r.pool.QueryRow(ctx, q, args...))
}

func (r *RetailRepository) DeletePackage(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM retail_packages WHERE id = $1`
	args := []any{id}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPackageNotFound
	}
	return nil
}

// ─── 零售合同 ───

// 合同查询统一 join customers 以带出客户名称。
const contractSelect = `SELECT c.id, c.customer_id, cust.user_name, c.package_id, c.package_name_snapshot,
	c.purchasing_energy_mwh, c.green_power_ratio, c.purchase_start_month, c.purchase_end_month,
	c.status, c.created_at, c.updated_at
	FROM retail_contracts c JOIN customers cust ON cust.id = c.customer_id`

func (r *RetailRepository) scanContract(row pgx.Row) (*RetailContract, error) {
	var c RetailContract
	err := row.Scan(&c.ID, &c.CustomerID, &c.CustomerName, &c.PackageID, &c.PackageNameSnapshot,
		&c.PurchasingEnergyMWH, &c.GreenPowerRatio, &c.PurchaseStartMonth, &c.PurchaseEndMonth,
		&c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *RetailRepository) ListContracts(ctx context.Context, keyword, status string) ([]*RetailContract, error) {
	where := make([]string, 0, 3)
	args := make([]any, 0, 3)
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("c.org_id = $%d::uuid", len(args)))
	}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where = append(where, fmt.Sprintf(
			"(cust.user_name ILIKE $%d OR c.package_name_snapshot ILIKE $%d)", len(args), len(args)))
	}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("c.status = $%d", len(args)))
	}
	q := contractSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY c.created_at DESC"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*RetailContract, 0, 16)
	for rows.Next() {
		c, err := r.scanContract(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *RetailRepository) GetContract(ctx context.Context, id uuid.UUID) (*RetailContract, error) {
	q := contractSelect + " WHERE c.id = $1"
	args := []any{id}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND c.org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	return r.scanContract(r.pool.QueryRow(ctx, q, args...))
}

// CreateContract 通过 INSERT ... SELECT 自动快照套餐名；套餐不存在则返回 ErrPackageNotFound。
func (r *RetailRepository) CreateContract(ctx context.Context, in ContractInput, createdBy *uuid.UUID) (*RetailContract, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	const q = `
		WITH ins AS (
			INSERT INTO retail_contracts
				(customer_id, package_id, package_name_snapshot, purchasing_energy_mwh,
				 green_power_ratio, purchase_start_month, purchase_end_month, status, created_by, org_id)
			SELECT $1, $2, rp.package_name, $3, $4, $5, $6, $7, $8, $9::uuid
			FROM retail_packages rp WHERE rp.id = $2
			RETURNING *
		)
		SELECT ins.id, ins.customer_id, cust.user_name, ins.package_id, ins.package_name_snapshot,
			ins.purchasing_energy_mwh, ins.green_power_ratio, ins.purchase_start_month,
			ins.purchase_end_month, ins.status, ins.created_at, ins.updated_at
		FROM ins JOIN customers cust ON cust.id = ins.customer_id`
	c, err := r.scanContract(r.pool.QueryRow(ctx, q,
		in.CustomerID, in.PackageID, in.PurchasingEnergyMWH, in.GreenPowerRatio,
		in.PurchaseStartMonth, in.PurchaseEndMonth, status, createdBy, org))
	if errors.Is(err, ErrContractNotFound) {
		// INSERT...SELECT 未插入任何行 = 套餐 ID 不存在
		return nil, ErrPackageNotFound
	}
	return c, err
}

func (r *RetailRepository) UpdateContract(ctx context.Context, id uuid.UUID, in ContractInput) (*RetailContract, error) {
	org, scoped := OrgFilter(ctx)
	q := `
		WITH upd AS (
			UPDATE retail_contracts SET
				customer_id=$2, package_id=$3,
				package_name_snapshot=(SELECT package_name FROM retail_packages WHERE id=$3),
				purchasing_energy_mwh=$4, green_power_ratio=$5,
				purchase_start_month=$6, purchase_end_month=$7, status=$8
			WHERE id=$1
			RETURNING *
		)
		SELECT upd.id, upd.customer_id, cust.user_name, upd.package_id, upd.package_name_snapshot,
			upd.purchasing_energy_mwh, upd.green_power_ratio, upd.purchase_start_month,
			upd.purchase_end_month, upd.status, upd.created_at, upd.updated_at
		FROM upd JOIN customers cust ON cust.id = upd.customer_id`
	args := []any{id,
		in.CustomerID, in.PackageID, in.PurchasingEnergyMWH, in.GreenPowerRatio,
		in.PurchaseStartMonth, in.PurchaseEndMonth, in.Status}
	if scoped {
		q = strings.Replace(q, "WHERE id=$1", fmt.Sprintf("WHERE id=$1 AND org_id = $%d::uuid", len(args)+1), 1)
		args = append(args, org)
	}
	return r.scanContract(r.pool.QueryRow(ctx, q, args...))
}

func (r *RetailRepository) DeleteContract(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM retail_contracts WHERE id = $1`
	args := []any{id}
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrContractNotFound
	}
	return nil
}
