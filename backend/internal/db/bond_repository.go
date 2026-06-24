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
)

// Bond 保函结构体。
type Bond struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	BondType    string     `json:"bond_type"`
	Amount      float64    `json:"amount"`
	Issuer      string     `json:"issuer"`
	Beneficiary string     `json:"beneficiary"`
	IssueDate   *time.Time `json:"issue_date,omitempty"`
	ExpireDate  *time.Time `json:"expire_date,omitempty"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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

const bondColumns = "id, name, bond_type, amount, issuer, beneficiary, issue_date, expire_date, status, description, created_by, created_at, updated_at"

func (r *BondRepository) scan(row pgx.Row) (*Bond, error) {
	var b Bond
	err := row.Scan(
		&b.ID, &b.Name, &b.BondType, &b.Amount, &b.Issuer, &b.Beneficiary,
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
	return r.scan(r.pool.QueryRow(ctx, q, id))
}

func (r *BondRepository) List(ctx context.Context, f BondListFilter) ([]*Bond, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 2)
	args := make([]any, 0, 2)
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
	q := `INSERT INTO bonds
		(name, bond_type, amount, issuer, beneficiary, issue_date, expire_date, status, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING ` + bondColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		in.Name, in.BondType, in.Amount, in.Issuer, in.Beneficiary,
		in.IssueDate, in.ExpireDate, in.Status, in.Description, createdBy,
	))
}

func (r *BondRepository) Update(ctx context.Context, id uuid.UUID, in BondInput) (*Bond, error) {
	q := `UPDATE bonds SET
		name = $2, bond_type = $3, amount = $4, issuer = $5, beneficiary = $6,
		issue_date = $7, expire_date = $8, status = $9, description = $10
		WHERE id = $1
		RETURNING ` + bondColumns
	return r.scan(r.pool.QueryRow(ctx, q, id,
		in.Name, in.BondType, in.Amount, in.Issuer, in.Beneficiary,
		in.IssueDate, in.ExpireDate, in.Status, in.Description,
	))
}

func (r *BondRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM bonds WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrBondNotFound
	}
	return nil
}
