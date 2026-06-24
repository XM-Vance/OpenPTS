// 审计日志仓储：写入由中间件触发，查询带筛选 + 分页。
package db

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type AuditLog struct {
	ID           string    `json:"id"`
	UserID       *string   `json:"user_id,omitempty"`
	Username     *string   `json:"username,omitempty"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Resource     *string   `json:"resource,omitempty"`
	ResourceID   *string   `json:"resource_id,omitempty"`
	StatusCode   int       `json:"status_code"`
	IP           *string   `json:"ip,omitempty"`
	UserAgent    *string   `json:"user_agent,omitempty"`
	DurationMs   int       `json:"duration_ms"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type AuditCreateInput struct {
	UserID       *string
	Username     *string
	Method       string
	Path         string
	Resource     *string
	ResourceID   *string
	StatusCode   int
	IP           *string
	UserAgent    *string
	DurationMs   int
	ErrorMessage *string
}

type AuditRepository struct {
	pool *Pool
}

func NewAuditRepository(pool *Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

const auditInsertSQL = `INSERT INTO audit_logs (user_id, username, method, path, resource, resource_id,
		status_code, ip, user_agent, duration_ms, error_message)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

func (r *AuditRepository) Create(ctx context.Context, in AuditCreateInput) error {
	_, err := r.pool.Exec(ctx, auditInsertSQL,
		in.UserID, in.Username, in.Method, in.Path, in.Resource, in.ResourceID,
		in.StatusCode, in.IP, in.UserAgent, in.DurationMs, in.ErrorMessage)
	return err
}

// CreateBatch 把多条审计记录在单次往返内管线化写入(pgx Batch),供异步批量写入器使用。
func (r *AuditRepository) CreateBatch(ctx context.Context, items []AuditCreateInput) error {
	if len(items) == 0 {
		return nil
	}
	b := &pgx.Batch{}
	for _, in := range items {
		b.Queue(auditInsertSQL,
			in.UserID, in.Username, in.Method, in.Path, in.Resource, in.ResourceID,
			in.StatusCode, in.IP, in.UserAgent, in.DurationMs, in.ErrorMessage)
	}
	br := r.pool.SendBatch(ctx, b)
	defer br.Close()
	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

type AuditFilter struct {
	Username string
	Method   string
	Resource string
	Days     int
	Limit    int
	Offset   int
}

// List 返回当前页与同条件下的总行数(total),供前端分页;
// 此前 LIMIT 硬截断会在日志增长后静默丢数据(P1-7)。
func (r *AuditRepository) List(ctx context.Context, f AuditFilter) ([]*AuditLog, int, error) {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	if f.Days <= 0 || f.Days > 90 {
		f.Days = 7
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	args := []any{time.Now().AddDate(0, 0, -f.Days)}
	conds := []string{"created_at >= $1"}
	if f.Username != "" {
		args = append(args, "%"+f.Username+"%")
		conds = append(conds, "username ILIKE $"+itoa(len(args)))
	}
	if f.Method != "" {
		args = append(args, f.Method)
		conds = append(conds, "method = $"+itoa(len(args)))
	}
	if f.Resource != "" {
		args = append(args, f.Resource)
		conds = append(conds, "resource = $"+itoa(len(args)))
	}
	where := strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM audit_logs WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := "SELECT id, user_id::text, username, method, path, resource, resource_id, " +
		"status_code, ip::text, user_agent, duration_ms, error_message, created_at " +
		"FROM audit_logs WHERE " + where +
		" ORDER BY created_at DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]*AuditLog, 0, f.Limit)
	for rows.Next() {
		var a AuditLog
		if err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Method, &a.Path,
			&a.Resource, &a.ResourceID, &a.StatusCode, &a.IP, &a.UserAgent,
			&a.DurationMs, &a.ErrorMessage, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &a)
	}
	return list, total, rows.Err()
}

// 简易 itoa（避免引 strconv）。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
