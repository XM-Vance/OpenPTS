// 通用审批流仓储：approval_requests 表的 CRUD + 状态机转换。
package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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

// BeginTx 在审批主连接池上开启事务，供 Approve 单事务原子化使用（P0-B2）。
func (r *ApprovalRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
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
	org, err := MustScoped(ctx)
	if err != nil {
		return nil, err
	}
	if in.Payload == nil {
		in.Payload = json.RawMessage("{}")
	}
	var a Approval
	err = r.pool.QueryRow(ctx, `
		INSERT INTO approval_requests
		  (resource, resource_id, title, payload, status, submitted_by, org_id)
		VALUES ($1, $2, $3, $4, 'pending', $5::uuid, $6)
		RETURNING id, resource, resource_id, title, payload, status,
		          submitted_by::text,
		          (SELECT username FROM users WHERE id = approval_requests.submitted_by),
		          reviewed_by::text,
		          (SELECT username FROM users WHERE id = approval_requests.reviewed_by),
		          review_note, reviewed_at, created_at, updated_at`,
		in.Resource, in.ResourceID, in.Title, in.Payload, in.SubmittedBy, org).
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
	return r.get(ctx, r.pool, id)
}

// get 内部复用：可在事务（pgx.Tx）或连接池（*Pool）上执行，支持 org 过滤。
func (r *ApprovalRepository) get(ctx context.Context, q Executor, id string) (*Approval, error) {
	args := []any{id}
	query := `
		SELECT ar.id, ar.resource, ar.resource_id, ar.title, ar.payload, ar.status,
		       ar.submitted_by::text, u_sub.username,
		       ar.reviewed_by::text, u_rev.username,
		       ar.review_note, ar.reviewed_at, ar.created_at, ar.updated_at
		FROM approval_requests ar
		LEFT JOIN users u_sub ON u_sub.id = ar.submitted_by
		LEFT JOIN users u_rev ON u_rev.id = ar.reviewed_by
		WHERE ar.id = $1`
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		query += fmt.Sprintf(" AND ar.org_id = $%d::uuid", len(args))
	}
	var a Approval
	err := q.QueryRow(ctx, query, args...).
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
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		conds = append(conds, "ar.org_id = $"+itoaApproval(len(args))+"::uuid")
	}
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
	args := []any{resource, resourceID}
	query := `
		SELECT ar.id, ar.resource, ar.resource_id, ar.title, ar.payload, ar.status,
		       ar.submitted_by::text, u_sub.username,
		       ar.reviewed_by::text, u_rev.username,
		       ar.review_note, ar.reviewed_at, ar.created_at, ar.updated_at
		FROM approval_requests ar
		LEFT JOIN users u_sub ON u_sub.id = ar.submitted_by
		LEFT JOIN users u_rev ON u_rev.id = ar.reviewed_by
		WHERE ar.resource = $1 AND ar.resource_id = $2`
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		query += fmt.Sprintf(" AND ar.org_id = $%d::uuid", len(args))
	}
	query += " ORDER BY ar.created_at DESC LIMIT 100"
	rows, err := r.pool.Query(ctx, query, args...)
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
// 注意：此方法仍为 read-then-write，仅用于 Reject/Withdraw（无副作用落库）。
// Approve 必须用 TransitionTx（事务 + 乐观锁），见 handler/approval.go。
func (r *ApprovalRepository) Transition(ctx context.Context, id, target, reviewer, note string) (*Approval, error) {
	return r.transitionOn(ctx, r.pool, id, target, reviewer, note)
}

// GetTx / TransitionTx 在指定事务上执行，供 Approve 单事务原子化使用（P0-B2）。
func (r *ApprovalRepository) GetTx(ctx context.Context, ex Executor, id string) (*Approval, error) {
	return r.get(ctx, ex, id)
}
func (r *ApprovalRepository) TransitionTx(ctx context.Context, ex Executor, id, target, reviewer, note string) (*Approval, error) {
	return r.transitionOn(ctx, ex, id, target, reviewer, note)
}

// transitionOn 在指定执行器（连接池或事务）上做状态机流转。
// UPDATE 带 AND status = $原态 乐观锁：并发审批只有一方能改成功，另一方返回
// ErrInvalidApprovalTransition，消除 P0-B8 的 TOCTOU 竞态。
func (r *ApprovalRepository) transitionOn(ctx context.Context, ex Executor, id, target, reviewer, note string) (*Approval, error) {
	cur, err := r.get(ctx, ex, id)
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
	tag, err := ex.Exec(ctx, `
		UPDATE approval_requests
		SET status = $1,
		    reviewed_by = COALESCE($2::uuid, reviewed_by),
		    review_note = COALESCE($3, review_note),
		    reviewed_at = `+reviewedAtSQL+`,
		    updated_at = now()
		WHERE id = $4 AND status = $5`,
		target, reviewerArg, noteArg, id, cur.Status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		// 并发竞争：状态已被别人改掉，当前流转不再合法
		return nil, ErrInvalidApprovalTransition
	}
	return r.get(ctx, ex, id)
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
