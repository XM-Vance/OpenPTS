// 通用审批流仓储：approval_requests 表的 CRUD + 状态机转换。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrApprovalNotFound = errors.New("审批请求不存在")
var ErrInvalidApprovalTransition = errors.New("非法状态流转")

type Approval struct {
	ID               string          `json:"id"`
	Resource         string          `json:"resource"`
	ResourceID       string          `json:"resource_id"`
	Title            string          `json:"title"`
	Payload          json.RawMessage `json:"payload"`
	Status           string          `json:"status"`
	SubmittedBy      string          `json:"submitted_by"`
	SubmittedByName  *string         `json:"submitted_by_name,omitempty"`
	ReviewedBy       *string         `json:"reviewed_by,omitempty"`
	ReviewedByName   *string         `json:"reviewed_by_name,omitempty"`
	ReviewNote       *string         `json:"review_note,omitempty"`
	ReviewedAt       *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type ApprovalInput struct {
	Resource    string
	ResourceID  string
	Title       string
	Payload     json.RawMessage
	SubmittedBy string
}

type ApprovalFilter struct {
	Status    string // 多个用逗号分隔
	Resource  string
	Submitter string
	Limit     int
	Offset    int
}

type ApprovalRepository struct{ pool *Pool }

func NewApprovalRepository(pool *Pool) *ApprovalRepository {
	return &ApprovalRepository{pool: pool}
}

// 合法的状态流转表（from → toSet）。
var validTransitions = map[string]map[string]bool{
	"draft":    {"pending": true, "withdrawn": true},
	"pending":  {"approved": true, "rejected": true, "withdrawn": true},
	"approved": {},
	"rejected": {},
	"withdrawn": {},
}

func (r *ApprovalRepository) Create(ctx context.Context, in ApprovalInput) (*Approval, error) {
	if in.Payload == nil {
		in.Payload = json.RawMessage("{}")
	}
	var a Approval
	err := r.pool.QueryRow(ctx, `
		INSERT INTO approval_requests
		  (resource, resource_id, title, payload, status, submitted_by)
		VALUES ($1, $2, $3, $4, 'pending', $5::uuid)
		RETURNING id, resource, resource_id, title, payload, status,
		          submitted_by::text,
		          (SELECT username FROM users WHERE id = approval_requests.submitted_by),
		          reviewed_by::text,
		          (SELECT username FROM users WHERE id = approval_requests.reviewed_by),
		          review_note, reviewed_at, created_at, updated_at`,
		in.Resource, in.ResourceID, in.Title, in.Payload, in.SubmittedBy).
		Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Title, &a.Payload, &a.Status,
			&a.SubmittedBy, &a.SubmittedByName,
			&a.ReviewedBy, &a.ReviewedByName,
			&a.ReviewNote, &a.ReviewedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ApprovalRepository) Get(ctx context.Context, id string) (*Approval, error) {
	var a Approval
	err := r.pool.QueryRow(ctx, `
		SELECT ar.id, ar.resource, ar.resource_id, ar.title, ar.payload, ar.status,
		       ar.submitted_by::text, u_sub.username,
		       ar.reviewed_by::text, u_rev.username,
		       ar.review_note, ar.reviewed_at, ar.created_at, ar.updated_at
		FROM approval_requests ar
		LEFT JOIN users u_sub ON u_sub.id = ar.submitted_by
		LEFT JOIN users u_rev ON u_rev.id = ar.reviewed_by
		WHERE ar.id = $1`, id).
		Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Title, &a.Payload, &a.Status,
			&a.SubmittedBy, &a.SubmittedByName,
			&a.ReviewedBy, &a.ReviewedByName,
			&a.ReviewNote, &a.ReviewedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, ErrApprovalNotFound
	}
	return &a, nil
}

// List 返回当前页与同条件下总行数(total),供前端分页;
// 此前 LIMIT 硬截断会在审批记录增长后静默丢数据(P1-7)。
func (r *ApprovalRepository) List(ctx context.Context, f ApprovalFilter) ([]*Approval, int, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	args := []any{}
	conds := []string{}
	if f.Status != "" {
		statuses := strings.Split(f.Status, ",")
		ph := []string{}
		for _, s := range statuses {
			args = append(args, strings.TrimSpace(s))
			ph = append(ph, "$"+itoaApproval(len(args)))
		}
		conds = append(conds, "ar.status IN ("+strings.Join(ph, ",")+")")
	}
	if f.Resource != "" {
		args = append(args, f.Resource)
		conds = append(conds, "ar.resource = $"+itoaApproval(len(args)))
	}
	if f.Submitter != "" {
		args = append(args, f.Submitter)
		conds = append(conds, "ar.submitted_by = $"+itoaApproval(len(args))+"::uuid")
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM approval_requests ar"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := `SELECT ar.id, ar.resource, ar.resource_id, ar.title, ar.payload, ar.status,
		         ar.submitted_by::text, u_sub.username,
		         ar.reviewed_by::text, u_rev.username,
		         ar.review_note, ar.reviewed_at, ar.created_at, ar.updated_at
		  FROM approval_requests ar
		  LEFT JOIN users u_sub ON u_sub.id = ar.submitted_by
		  LEFT JOIN users u_rev ON u_rev.id = ar.reviewed_by` + where +
		" ORDER BY ar.created_at DESC LIMIT $" + itoaApproval(len(args)-1) + " OFFSET $" + itoaApproval(len(args))
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*Approval, 0)
	for rows.Next() {
		var a Approval
		if err := rows.Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Title, &a.Payload,
			&a.Status, &a.SubmittedBy, &a.SubmittedByName,
			&a.ReviewedBy, &a.ReviewedByName,
			&a.ReviewNote, &a.ReviewedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &a)
	}
	return list, total, rows.Err()
}

type ApprovalTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	TitleTpl    string    `json:"title_tpl"`
	Field       string    `json:"field"`
	Description *string   `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

func (r *ApprovalRepository) ListTemplates(ctx context.Context, resource string) ([]*ApprovalTemplate, error) {
	args := []any{}
	q := `SELECT id, name, resource, title_tpl, field, description, enabled, created_at
		  FROM approval_templates WHERE enabled = true`
	if resource != "" {
		args = append(args, resource)
		q += " AND resource = $1"
	}
	q += " ORDER BY name"
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*ApprovalTemplate, 0)
	for rows.Next() {
		var t ApprovalTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Resource, &t.TitleTpl, &t.Field,
			&t.Description, &t.Enabled, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

// ByResource 列出同一资源（按 resource+resource_id）的所有审批历史。
func (r *ApprovalRepository) ByResource(ctx context.Context, resource, resourceID string) ([]*Approval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ar.id, ar.resource, ar.resource_id, ar.title, ar.payload, ar.status,
		       ar.submitted_by::text, u_sub.username,
		       ar.reviewed_by::text, u_rev.username,
		       ar.review_note, ar.reviewed_at, ar.created_at, ar.updated_at
		FROM approval_requests ar
		LEFT JOIN users u_sub ON u_sub.id = ar.submitted_by
		LEFT JOIN users u_rev ON u_rev.id = ar.reviewed_by
		WHERE ar.resource = $1 AND ar.resource_id = $2
		ORDER BY ar.created_at DESC
		LIMIT 100`, resource, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Approval, 0)
	for rows.Next() {
		var a Approval
		if err := rows.Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Title, &a.Payload,
			&a.Status, &a.SubmittedBy, &a.SubmittedByName,
			&a.ReviewedBy, &a.ReviewedByName,
			&a.ReviewNote, &a.ReviewedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

// Transition 严格状态机：从当前状态流转到目标态。
func (r *ApprovalRepository) Transition(ctx context.Context, id, target, reviewer, note string) (*Approval, error) {
	cur, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	allowed, ok := validTransitions[cur.Status]
	if !ok || !allowed[target] {
		return nil, ErrInvalidApprovalTransition
	}
	// approved/rejected 需要 reviewer
	var (
		reviewedAtSQL = "NULL"
		reviewerArg   any
		noteArg       any
	)
	if target == "approved" || target == "rejected" {
		reviewedAtSQL = "now()"
		if reviewer != "" {
			reviewerArg = reviewer
		}
		if note != "" {
			noteArg = note
		}
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE approval_requests
		SET status = $1,
		    reviewed_by = COALESCE($2::uuid, reviewed_by),
		    review_note = COALESCE($3, review_note),
		    reviewed_at = `+reviewedAtSQL+`,
		    updated_at = now()
		WHERE id = $4`,
		target, reviewerArg, noteArg, id)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func itoaApproval(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
