// 政策文件库仓储：把政策文件归纳为结构化条目（policy_documents），可关联来源解析文档。
// 数据来源：文档解析「确认入库 → 政策文件」自动归纳，或页面手动新增。按活跃省隔离。
package db

import (
	"context"
	"fmt"
	"time"
)

type PolicyDocument struct {
	ID            string    `json:"id"`
	DocumentID    *string   `json:"document_id"`
	DocumentName  *string   `json:"document_name"` // 来源文档文件名（LEFT JOIN）
	Title         string    `json:"title"`
	DocNo         *string   `json:"doc_no"`
	Category      *string   `json:"category"`
	EffectiveDate *string   `json:"effective_date"` // YYYY-MM-DD
	Summary       *string   `json:"summary"`
	Source        string    `json:"source"` // manual / document
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PolicyInput struct {
	DocumentID    string // 来源文档 id，可空
	Title         string
	DocNo         string
	Category      string
	EffectiveDate string // YYYY-MM-DD，可空
	Summary       string
	Source        string // manual / document
	CreatedBy     string // 用户 id，可空
}

type PolicyRepository struct{ pool *Pool }

func NewPolicyRepository(pool *Pool) *PolicyRepository {
	return &PolicyRepository{pool: pool}
}

// List 列出政策文件；category 为空时返回全部分类。按生效日期倒序。
func (r *PolicyRepository) List(ctx context.Context, category string, limit int) ([]*PolicyDocument, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []any{}
	q := `SELECT pd.id, pd.document_id, d.filename, pd.title, pd.doc_no, pd.category,
			to_char(pd.effective_date, 'YYYY-MM-DD'), pd.summary, pd.source,
			pd.created_at, pd.updated_at
		 FROM policy_documents pd
		 LEFT JOIN documents d ON d.id = pd.document_id
		 WHERE 1=1`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND pd.org_id = $%d::uuid", len(args))
	}
	if category != "" {
		args = append(args, category)
		q += fmt.Sprintf(" AND pd.category = $%d", len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY pd.effective_date DESC NULLS LAST, pd.created_at DESC LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*PolicyDocument, 0, limit)
	for rows.Next() {
		var p PolicyDocument
		if err := rows.Scan(&p.ID, &p.DocumentID, &p.DocumentName, &p.Title, &p.DocNo,
			&p.Category, &p.EffectiveDate, &p.Summary, &p.Source,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// Create 新增政策条目。写操作要求具体活跃省。返回新建 id。
func (r *PolicyRepository) Create(ctx context.Context, in PolicyInput) (string, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return "", ErrOrgRequired
	}
	if in.Source == "" {
		in.Source = "manual"
	}
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO policy_documents
		   (org_id, document_id, title, doc_no, category, effective_date, summary, source, created_by)
		 VALUES ($1::uuid, NULLIF($2,'')::uuid, $3, NULLIF($4,''), NULLIF($5,''),
		         NULLIF($6,'')::date, NULLIF($7,''), $8, NULLIF($9,'')::uuid)
		 RETURNING id`,
		org, in.DocumentID, in.Title, in.DocNo, in.Category, in.EffectiveDate,
		in.Summary, in.Source, in.CreatedBy).Scan(&id)
	return id, err
}

// Delete 删除政策条目（按活跃省限定）。
func (r *PolicyRepository) Delete(ctx context.Context, id string) error {
	args := []any{id}
	q := `DELETE FROM policy_documents WHERE id = $1::uuid`
	org, scoped := OrgFilter(ctx)
	if scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}
