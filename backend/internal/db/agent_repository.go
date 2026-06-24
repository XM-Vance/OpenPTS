// 代理商管理仓储：CRUD + 关联客户查询。
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

// Agent 代理商结构体。
type Agent struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          *uuid.UUID `json:"org_id,omitempty"`
	AgentName      string     `json:"agent_name"`
	ContactPerson  string     `json:"contact_person"`
	Phone          string     `json:"phone"`
	Email          string     `json:"email"`
	Region         string     `json:"region"`
	CommissionRate float64    `json:"commission_rate"`
	Status         string     `json:"status"`
	Description    string     `json:"description"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

var ErrAgentNotFound = errors.New("代理商不存在")

// AgentInput 创建/更新入参。
type AgentInput struct {
	AgentName      string
	ContactPerson  string
	Phone          string
	Email          string
	Region         string
	CommissionRate float64
	Status         string
	Description    string
}

// AgentListFilter 列表过滤参数。
type AgentListFilter struct {
	Keyword string
	Status  string
	Limit   int
	Offset  int
}

type AgentRepository struct {
	pool *Pool
}

func NewAgentRepository(pool *Pool) *AgentRepository {
	return &AgentRepository{pool: pool}
}

const agentColumns = "id, org_id, agent_name, contact_person, phone, email, region, commission_rate, status, description, created_by, created_at, updated_at"

func (r *AgentRepository) scan(row pgx.Row) (*Agent, error) {
	var a Agent
	err := row.Scan(
		&a.ID, &a.OrgID, &a.AgentName, &a.ContactPerson, &a.Phone, &a.Email,
		&a.Region, &a.CommissionRate, &a.Status, &a.Description,
		&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *AgentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	q := `SELECT ` + agentColumns + ` FROM agents WHERE id = $1`
	args := []any{id}
	if org, scoped := OrgFilter(ctx); scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *AgentRepository) List(ctx context.Context, f AgentListFilter) ([]*Agent, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		where = append(where, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}
	if f.Keyword != "" {
		args = append(args, "%"+f.Keyword+"%")
		where = append(where, fmt.Sprintf("(agent_name ILIKE $%d OR contact_person ILIKE $%d)", len(args), len(args)))
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
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM agents"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := "SELECT " + agentColumns + " FROM agents" + whereClause +
		" ORDER BY created_at DESC LIMIT " + strconv.Itoa(f.Limit) + " OFFSET " + strconv.Itoa(f.Offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*Agent, 0, f.Limit)
	for rows.Next() {
		a, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}

func (r *AgentRepository) Create(ctx context.Context, in AgentInput, createdBy *uuid.UUID) (*Agent, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	q := `INSERT INTO agents
		(org_id, agent_name, contact_person, phone, email, region, commission_rate, status, description, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING ` + agentColumns
	return r.scan(r.pool.QueryRow(ctx, q,
		org, in.AgentName, in.ContactPerson, in.Phone, in.Email, in.Region,
		in.CommissionRate, in.Status, in.Description, createdBy,
	))
}

func (r *AgentRepository) Update(ctx context.Context, id uuid.UUID, in AgentInput) (*Agent, error) {
	q := `UPDATE agents SET
		agent_name = $2, contact_person = $3, phone = $4, email = $5,
		region = $6, commission_rate = $7, status = $8, description = $9
		WHERE id = $1`
	args := []any{id,
		in.AgentName, in.ContactPerson, in.Phone, in.Email, in.Region,
		in.CommissionRate, in.Status, in.Description,
	}
	if org, scoped := OrgFilter(ctx); scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	q += ` RETURNING ` + agentColumns
	return r.scan(r.pool.QueryRow(ctx, q, args...))
}

func (r *AgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM agents WHERE id = $1`
	args := []any{id}
	if org, scoped := OrgFilter(ctx); scoped {
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args)+1)
		args = append(args, org)
	}
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAgentNotFound
	}
	return nil
}

// GetCustomers 查询代理商关联的客户列表（通过客户 extra 字段或单独关联）。
// 此处返回通过 source 字段关联代理商名的客户。
func (r *AgentRepository) GetCustomers(ctx context.Context, agentID uuid.UUID) ([]*Customer, error) {
	// 先获取代理商名
	agent, err := r.GetByID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	q := `SELECT ` + customerColumns + ` FROM customers WHERE source = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, agent.AgentName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Customer, 0)
	for rows.Next() {
		c, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func scanCustomer(row pgx.Row) (*Customer, error) {
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
