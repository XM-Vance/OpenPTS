// 组织（省份）仓储：组织 CRUD + 成员查询。多对多成员关系在 user_repository（user_orgs）。
package db

import (
	"context"
	"time"
)

type Org struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type OrgMember struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	IsHQ        bool    `json:"is_hq"`
}

type OrgRepository struct{ pool *Pool }

func NewOrgRepository(pool *Pool) *OrgRepository { return &OrgRepository{pool: pool} }

const orgColumns = "id::text, code, name, is_active, created_at"

// List 返回全部组织（按 code 排序，default 排最后）。
func (r *OrgRepository) List(ctx context.Context) ([]*Org, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+orgColumns+` FROM organizations ORDER BY (code='default'), code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Org, 0)
	for rows.Next() {
		var o Org
		if err := rows.Scan(&o.ID, &o.Code, &o.Name, &o.IsActive, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &o)
	}
	return list, rows.Err()
}

// Create 新建组织（加省份时用）。
func (r *OrgRepository) Create(ctx context.Context, code, name string) (*Org, error) {
	var o Org
	err := r.pool.QueryRow(ctx,
		`INSERT INTO organizations (code, name) VALUES ($1,$2) RETURNING `+orgColumns,
		code, name).Scan(&o.ID, &o.Code, &o.Name, &o.IsActive, &o.CreatedAt)
	return &o, err
}

// Update 改名 / 启停。
func (r *OrgRepository) Update(ctx context.Context, id, name string, isActive bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE organizations SET name=$2, is_active=$3 WHERE id=$1::uuid`,
		id, name, isActive)
	return err
}

// ListMembers 返回某省的成员（user_orgs）。
func (r *OrgRepository) ListMembers(ctx context.Context, orgID string) ([]*OrgMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id::text, u.username, u.display_name, u.is_hq
		   FROM user_orgs uo JOIN users u ON u.id = uo.user_id
		  WHERE uo.org_id = $1::uuid ORDER BY u.username`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*OrgMember, 0)
	for rows.Next() {
		var m OrgMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.DisplayName, &m.IsHQ); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}
