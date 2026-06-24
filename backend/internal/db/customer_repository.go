// 客户档案仓储：CRUD + 标签/经理过滤 + 按省隔离。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Customer struct {
	ID        uuid.UUID       `json:"id"`
	UserName  string          `json:"user_name"`
	ShortName *string         `json:"short_name,omitempty"`
	Location  *string         `json:"location,omitempty"`
	Source    *string         `json:"source,omitempty"`
	Manager   *string         `json:"manager,omitempty"`
	Tags      []string        `json:"tags"`
	Accounts  json.RawMessage `json:"accounts"`
	IsDemo    bool            `json:"is_demo"`
	Extra     json.RawMessage `json:"extra"`
	CreatedBy *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	OrgID     *string         `json:"org_id,omitempty"`
}

var ErrCustomerNotFound = errors.New("客户不存在")

// CustomerInput 创建/更新入参。
type CustomerInput struct {
	UserName  string
	ShortName string
	Location  string
	Source    string
	Manager   string
	Tags      []string
	Accounts  json.RawMessage
	IsDemo    bool
}

// CustomerListFilter 列表过滤参数。
type CustomerListFilter struct {
	Keyword string
	Tag     string
	Manager string
	Limit   int
	Offset  int
}

type CustomerRepository struct {
	pool *Pool
}

func NewCustomerRepository(pool *Pool) *CustomerRepository {
	return &CustomerRepository{pool: pool}
}

// Pool 暴露底层连接池（用于 360 视图等聚合查询）。
func (r *CustomerRepository) Pool() *Pool { return r.pool }

const customerColumns = "id, user_name, short_name, location, source, manager, tags, accounts, is_demo, extra, created_by, created_at, updated_at"

func (r *CustomerRepository) scan(row pgx.Row) (*Customer, error) {
	var c Customer
	var accounts, extra []byte
	err := row.Scan(
		&c.ID, &c.UserName, &c.ShortName, &c.Location, &c.Source, &c.Manager,
		&c.Tags, &accounts, &c.IsDemo, &extra, &c.CreatedBy,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	c.Accounts = json.RawMessage(accounts)
	c.Extra = json.RawMessage(extra)
	return &c, nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	org, scoped := OrgFilter(ctx)
	q := `SELECT ` + customerColumns + ` FROM customers WHERE id = $1`
	args := []any{id}
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *CustomerRepository) List(ctx context.Context, f CustomerListFilter) ([]*Customer, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 4)
	args := make([]any, 0, 4)

	// org 过滤
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}

	if f.Keyword != "" {
		args = append(args, "%"+f.Keyword+"%")
		where = append(where, fmt.Sprintf("(user_name ILIKE $%d OR short_name ILIKE $%d)", len(args), len(args)))
	}
	if f.Manager != "" {
		args = append(args, f.Manager)
		where = append(where, fmt.Sprintf("manager = $%d", len(args)))
	}
	if f.Tag != "" {
		args = append(args, f.Tag)
		where = append(where, fmt.Sprintf("$%d = ANY(tags)", len(args)))
	}
	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM customers"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := "SELECT " + customerColumns + " FROM customers" + whereClause +
		" ORDER BY created_at DESC LIMIT " + strconv.Itoa(f.Limit) + " OFFSET " + strconv.Itoa(f.Offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*Customer, 0, f.Limit)
	for rows.Next() {
		c, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, c)
	}
	return list, total, rows.Err()
}

func (r *CustomerRepository) Create(ctx context.Context, in CustomerInput, createdBy *uuid.UUID) (*Customer, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	q := `INSERT INTO customers
		(user_name, short_name, location, source, manager, tags, accounts, is_demo, created_by, org_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::uuid)
		RETURNING ` + customerColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		in.UserName, nullStr(in.ShortName), nullStr(in.Location), nullStr(in.Source), nullStr(in.Manager),
		tagsOrEmpty(in.Tags), accountsOrEmpty(in.Accounts), in.IsDemo, createdBy, org,
	))
}

func (r *CustomerRepository) Update(ctx context.Context, id uuid.UUID, in CustomerInput) (*Customer, error) {
	org, scoped := OrgFilter(ctx)
	q := `UPDATE customers SET
		user_name = $2, short_name = $3, location = $4, source = $5, manager = $6,
		tags = $7, accounts = $8, is_demo = $9
		WHERE id = $1`
	args := []any{id,
		in.UserName, nullStr(in.ShortName), nullStr(in.Location), nullStr(in.Source), nullStr(in.Manager),
		tagsOrEmpty(in.Tags), accountsOrEmpty(in.Accounts), in.IsDemo,
	}
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += ` RETURNING ` + customerColumns
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	org, scoped := OrgFilter(ctx)
	q := `DELETE FROM customers WHERE id = $1`
	args := []any{id}
	if scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

// ─── helpers ───────────────────────────────────────

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func tagsOrEmpty(t []string) []string {
	if t == nil {
		return []string{}
	}
	return t
}

func accountsOrEmpty(a json.RawMessage) json.RawMessage {
	if len(a) == 0 {
		return json.RawMessage("[]")
	}
	return a
}
