// 保函管理仓储：CRUD。
package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Bond 保函结构体。
type Bond struct {
	ID          uuid.UUID       `json:"id"`
	OrgID       *uuid.UUID      `json:"org_id,omitempty"` // 多租户：所属省份（0102 新增）
	Name        string          `json:"name"`
	BondType    string          `json:"bond_type"`
	Amount      decimal.Decimal `json:"amount"` // P4: numeric(18,4)
	Issuer      string          `json:"issuer"`
	Beneficiary string          `json:"beneficiary"`
	IssueDate   *time.Time      `json:"issue_date,omitempty"`
	ExpireDate  *time.Time      `json:"expire_date,omitempty"`
	Status      string          `json:"status"`
	Description string          `json:"description"`
	CreatedBy   *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

var ErrBondNotFound = errors.New("保函不存在")

// BondInput 创建/更新入参。
type BondInput struct {
	Name        string
	BondType    string
	Amount      float64
	Issuer      string
	Beneficiary string
	IssueDate   *time.Time
	ExpireDate  *time.Time
	Status      string
	Description string
}

// BondListFilter 列表过滤参数。
type BondListFilter struct {
	Keyword string
	Status  string
	Limit   int
	Offset  int
}

type BondRepository struct {
	pool *Pool
}

func NewBondRepository(pool *Pool) *BondRepository {
	return &BondRepository{pool: pool}
}

const bondColumns = "id, org_id, name, bond_type, amount, issuer, beneficiary, issue_date, expire_date, status, description, created_by, created_at, updated_at"

func (r *BondRepository) scan(row pgx.Row) (*Bond, error) {
	var b Bond
	err := row.Scan(
		&b.ID, &b.OrgID, &b.Name, &b.BondType, &b.Amount, &b.Issuer, &b.Beneficiary,
		&b.IssueDate, &b.ExpireDate, &b.Status, &b.Description,
		&b.CreatedBy, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBondNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *BondRepository) GetByID(ctx context.Context, id uuid.UUID) (*Bond, error) {
	q := `SELECT ` + bondColumns + ` FROM bonds WHERE id = $1`
	args := []any{id}
	// 读隔离：非总部只能看本省保函
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += " AND org_id = $2::uuid"
		args = append(args, org)
	}
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *BondRepository) List(ctx context.Context, f BondListFilter) ([]*Bond, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 3)
	args := make([]any, 0, 3)
	// 读隔离：按活跃组织过滤
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}
	if f.Keyword != "" {
		args = append(args, "%"+f.Keyword+"%")
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR issuer ILIKE $%d)", len(args), len(args)))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bonds"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := "SELECT " + bondColumns + " FROM bonds" + whereClause +
		" ORDER BY created_at DESC LIMIT " + strconv.Itoa(f.Limit) + " OFFSET " + strconv.Itoa(f.Offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*Bond, 0, f.Limit)
	for rows.Next() {
		b, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, b)
	}
	return list, total, rows.Err()
}

func (r *BondRepository) Create(ctx context.Context, in BondInput, createdBy *uuid.UUID) (*Bond, error) {
	// 写操作：要求具体活跃组织（总部「全部省」禁止写）
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	q := `INSERT INTO bonds
		(org_id, name, bond_type, amount, issuer, beneficiary, issue_date, expire_date, status, description, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING ` + bondColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		org, in.Name, in.BondType, decimal.NewFromFloat(in.Amount), in.Issuer, in.Beneficiary, // P4: 金额以 decimal 写入 numeric
		in.IssueDate, in.ExpireDate, in.Status, in.Description, createdBy,
	))
}

func (r *BondRepository) Update(ctx context.Context, id uuid.UUID, in BondInput) (*Bond, error) {
	q := `UPDATE bonds SET
		name = $2, bond_type = $3, amount = $4, issuer = $5, beneficiary = $6,
		issue_date = $7, expire_date = $8, status = $9, description = $10, updated_at = now()
		WHERE id = $1`
	args := []any{id,
		in.Name, in.BondType, decimal.NewFromFloat(in.Amount), in.Issuer, in.Beneficiary, // P4: 金额以 decimal 写入 numeric
		in.IssueDate, in.ExpireDate, in.Status, in.Description,
	}
	// 防跨省改：非总部追加 org 条件
	org, scoped := OrgFilter(ctx)
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += " RETURNING " + bondColumns
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *BondRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM bonds WHERE id = $1`
	args := []any{id}
	// 防跨省删
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
		return ErrBondNotFound
	}
	return nil
}
