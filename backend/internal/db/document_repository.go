// 文档解析管线仓储：documents / document_extractions / document_applies。
// 文档按省份(org)隔离：上传必须有具体活跃省；读取按活跃省过滤（总部「全部省」看全部）。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Document 文档主记录。
type Document struct {
	ID                string          `json:"id"`
	OrgID             *string         `json:"org_id"`
	Filename          string          `json:"filename"`
	ContentType       string          `json:"content_type"`
	Size              int64           `json:"size"`
	Sha256            *string         `json:"sha256"`
	SourceKind        string          `json:"source_kind"`
	OriginalObjectKey *string         `json:"original_object_key"`
	ParsedObjectKey   *string         `json:"parsed_object_key"`
	DocType           *string         `json:"doc_type"`
	Status            string          `json:"status"`
	PageCount         int             `json:"page_count"`
	TextContent       *string         `json:"text_content,omitempty"`
	Tables            json.RawMessage `json:"tables,omitempty"`
	Entities          json.RawMessage `json:"entities,omitempty"`
	Summary           *string         `json:"summary"`
	Error             *string         `json:"error"`
	UploadedBy        *string         `json:"uploaded_by"`
	AutoApplied       bool            `json:"auto_applied"`
	CustomerID        *string         `json:"customer_id,omitempty"`
	ContractID        *string         `json:"contract_id,omitempty"`
	IntentCustomerID  *string         `json:"intent_customer_id,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// DocumentExtraction 结构化提取字段。
type DocumentExtraction struct {
	ID         int64      `json:"id"`
	DocumentID string     `json:"document_id"`
	GroupNo    int        `json:"group_no"`
	FieldKey   string     `json:"field_key"`
	FieldLabel string     `json:"field_label"`
	ValueText  *string    `json:"value_text"`
	ValueNum   *float64   `json:"value_num"`
	ValueDate  *time.Time `json:"value_date"`
	Unit       *string    `json:"unit"`
	Confidence *float64   `json:"confidence"`
	Source     string     `json:"source"`
}

// ExtractionInput 写入用的提取字段。
type ExtractionInput struct {
	GroupNo    int
	FieldKey   string
	FieldLabel string
	ValueText  string
	Unit       string
	Confidence float64
	Source     string
}

// DocumentApply 入库审计记录。
type DocumentApply struct {
	ID          int64           `json:"id"`
	DocumentID  string          `json:"document_id"`
	Target      string          `json:"target"`
	AppliedRows int             `json:"applied_rows"`
	Detail      json.RawMessage `json:"detail,omitempty"`
	AppliedBy   *string         `json:"applied_by"`
	AppliedAt   time.Time       `json:"applied_at"`
}

type DocumentRepository struct{ pool *Pool }

func NewDocumentRepository(pool *Pool) *DocumentRepository { return &DocumentRepository{pool: pool} }

const documentColumns = `id::text, org_id::text, filename, content_type, size, sha256, source_kind,
	original_object_key, parsed_object_key, doc_type, status, page_count,
	summary, error, uploaded_by, auto_applied, customer_id::text, contract_id::text, intent_customer_id::text,
	created_at, updated_at`

func scanDocument(row interface{ Scan(...any) error }, withBody bool) (*Document, error) {
	var d Document
	dest := []any{&d.ID, &d.OrgID, &d.Filename, &d.ContentType, &d.Size, &d.Sha256, &d.SourceKind,
		&d.OriginalObjectKey, &d.ParsedObjectKey, &d.DocType, &d.Status, &d.PageCount,
		&d.Summary, &d.Error, &d.UploadedBy, &d.AutoApplied, &d.CustomerID, &d.ContractID, &d.IntentCustomerID,
		&d.CreatedAt, &d.UpdatedAt}
	if withBody {
		dest = append(dest, &d.TextContent, &d.Tables, &d.Entities)
	}
	if err := row.Scan(dest...); err != nil {
		return nil, err
	}
	return &d, nil
}

// Create 新建文档（上传后）。写操作要求具体活跃省（总部「全部省」返回 ErrOrgRequired）。
func (r *DocumentRepository) Create(ctx context.Context, filename, contentType string, size int64,
	sha, sourceKind, originalKey string, uploadedBy *uuid.UUID) (*Document, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, ErrOrgRequired
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO documents (org_id, filename, content_type, size, sha256, source_kind,
		                       original_object_key, status, uploaded_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, 'uploaded', $8)
		RETURNING `+documentColumns,
		org, filename, contentType, size, sha, sourceKind, originalKey, uploadedBy)
	return scanDocument(row, false)
}

// FindBySha 按指纹查重（同省内）。未找到返回 nil。
func (r *DocumentRepository) FindBySha(ctx context.Context, sha string) (*Document, error) {
	org, scoped := OrgFilter(ctx)
	if !scoped {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx,
		`SELECT `+documentColumns+` FROM documents WHERE sha256=$1 AND org_id=$2::uuid LIMIT 1`,
		sha, org)
	d, err := scanDocument(row, false)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

// List 文档列表（按活跃省过滤；status/docType 可选；非管理员仅看自己上传的）。
func (r *DocumentRepository) List(ctx context.Context, status, docType string, limit int, uploadedBy *uuid.UUID) ([]*Document, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{}
	conds := []string{}
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		conds = append(conds, fmt.Sprintf("org_id = $%d::uuid", len(args)))
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, fmt.Sprintf("status = $%d", len(args)))
	}
	if docType != "" {
		args = append(args, docType)
		conds = append(conds, fmt.Sprintf("doc_type = $%d", len(args)))
	}
	if uploadedBy != nil {
		args = append(args, *uploadedBy)
		conds = append(conds, fmt.Sprintf("uploaded_by = $%d::uuid", len(args)))
	}
	q := `SELECT ` + documentColumns + ` FROM documents`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY created_at DESC LIMIT " + strconv.Itoa(limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*Document, 0)
	for rows.Next() {
		d, err := scanDocument(rows, false)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

// Get 文档详情（含全文/表格/实体），按活跃省过滤；不存在或非本省返回 nil。
func (r *DocumentRepository) Get(ctx context.Context, id uuid.UUID) (*Document, error) {
	args := []any{id}
	q := `SELECT ` + documentColumns + `, text_content, tables, entities FROM documents WHERE id=$1`
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	d, err := scanDocument(r.pool.QueryRow(ctx, q, args...), true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

// GetByIDInternal 后台 worker 用：按 id 直取，不做 org 过滤（worker 无请求上下文）。
func (r *DocumentRepository) GetByIDInternal(ctx context.Context, id uuid.UUID) (*Document, error) {
	q := `SELECT ` + documentColumns + `, text_content, tables, entities FROM documents WHERE id=$1`
	d, err := scanDocument(r.pool.QueryRow(ctx, q, id), true)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

// SetOriginalKey 回填原件归档 key（上传先建记录拿 id、再按 id 归档）。
func (r *DocumentRepository) SetOriginalKey(ctx context.Context, id uuid.UUID, key string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET original_object_key=$2, updated_at=now() WHERE id=$1`, id, key)
	return err
}

// SetStatus 状态流转（parsing / failed 等）。
func (r *DocumentRepository) SetStatus(ctx context.Context, id uuid.UUID, status, errMsg string) error {
	var errVal any
	if errMsg != "" {
		errVal = errMsg
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET status=$2, error=$3, updated_at=now() WHERE id=$1`,
		id, status, errVal)
	return err
}

// SetParsed 解析成功：写回类型/页数/全文/表格/实体/摘要/解析件 key。
func (r *DocumentRepository) SetParsed(ctx context.Context, id uuid.UUID, docType string, pageCount int,
	textContent string, tables, entities json.RawMessage, summary, parsedKey string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE documents SET status='parsed', doc_type=$2, page_count=$3, text_content=$4,
		       tables=$5, entities=$6, summary=$7, parsed_object_key=NULLIF($8,''),
		       error=NULL, updated_at=now()
		WHERE id=$1`,
		id, docType, pageCount, textContent, tables, entities, summary, parsedKey)
	return err
}

// ReplaceExtractions 重写文档的全部提取字段（解析/重解析时）。
func (r *DocumentRepository) ReplaceExtractions(ctx context.Context, docID uuid.UUID, list []ExtractionInput) error {
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM document_extractions WHERE document_id=$1`, docID); err != nil {
		return err
	}
	for _, in := range list {
		num, date := parseValueTyped(in.ValueText)
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO document_extractions
			  (document_id, group_no, field_key, field_label, value_text, value_num, value_date,
			   unit, confidence, source)
			VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,NULLIF($8,''),$9,$10)`,
			docID, in.GroupNo, in.FieldKey, in.FieldLabel, in.ValueText, num, date,
			in.Unit, in.Confidence, in.Source); err != nil {
			return err
		}
	}
	return nil
}

// ListExtractions 文档的提取字段（按行组、字段序）。
func (r *DocumentRepository) ListExtractions(ctx context.Context, docID uuid.UUID) ([]*DocumentExtraction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, document_id::text, group_no, field_key, field_label,
		       value_text, value_num, value_date, unit, confidence, source
		FROM document_extractions WHERE document_id=$1 ORDER BY group_no, id`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DocumentExtraction, 0)
	for rows.Next() {
		var e DocumentExtraction
		if err := rows.Scan(&e.ID, &e.DocumentID, &e.GroupNo, &e.FieldKey, &e.FieldLabel,
			&e.ValueText, &e.ValueNum, &e.ValueDate, &e.Unit, &e.Confidence, &e.Source); err != nil {
			return nil, err
		}
		list = append(list, &e)
	}
	return list, rows.Err()
}

// AddExtraction 人工补充一个提取字段（source=manual），返回新记录。
func (r *DocumentRepository) AddExtraction(ctx context.Context, docID uuid.UUID,
	groupNo int, fieldKey, fieldLabel, valueText, unit string) (*DocumentExtraction, error) {
	num, date := parseValueTyped(valueText)
	var e DocumentExtraction
	err := r.pool.QueryRow(ctx, `
		INSERT INTO document_extractions
		  (document_id, group_no, field_key, field_label, value_text, value_num, value_date,
		   unit, confidence, source)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,NULLIF($8,''),1,'manual')
		RETURNING id, document_id::text, group_no, field_key, field_label,
		          value_text, value_num, value_date, unit, confidence, source`,
		docID, groupNo, fieldKey, fieldLabel, valueText, num, date, unit).
		Scan(&e.ID, &e.DocumentID, &e.GroupNo, &e.FieldKey, &e.FieldLabel,
			&e.ValueText, &e.ValueNum, &e.ValueDate, &e.Unit, &e.Confidence, &e.Source)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// DeleteExtraction 删除某个提取字段。
func (r *DocumentRepository) DeleteExtraction(ctx context.Context, docID uuid.UUID, extID int64) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM document_extractions WHERE id=$2 AND document_id=$1`, docID, extID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("提取字段不存在")
	}
	return nil
}

// UpdateExtraction 人工修正某个提取字段的值（重算 num/date 冗余列）。
func (r *DocumentRepository) UpdateExtraction(ctx context.Context, docID uuid.UUID, extID int64, valueText string) error {
	num, date := parseValueTyped(valueText)
	tag, err := r.pool.Exec(ctx, `
		UPDATE document_extractions SET value_text=NULLIF($3,''), value_num=$4, value_date=$5
		WHERE id=$2 AND document_id=$1`,
		docID, extID, valueText, num, date)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("提取字段不存在")
	}
	return nil
}

// InsertApply 记录一次「确认入库」。
func (r *DocumentRepository) InsertApply(ctx context.Context, docID uuid.UUID, target string,
	rowsApplied int, detail json.RawMessage, by *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO document_applies (document_id, target, applied_rows, detail, applied_by)
		VALUES ($1,$2,$3,$4,$5)`,
		docID, target, rowsApplied, detail, by)
	return err
}

// SetAutoApplied 标记文档为「系统自动入库」或取消标记。
func (r *DocumentRepository) SetAutoApplied(ctx context.Context, docID uuid.UUID, val bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE documents SET auto_applied=$2, updated_at=now() WHERE id=$1`, docID, val)
	return err
}

// ListApplies 文档的入库历史。
func (r *DocumentRepository) ListApplies(ctx context.Context, docID uuid.UUID) ([]*DocumentApply, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, document_id::text, target, applied_rows, detail, applied_by, applied_at
		FROM document_applies WHERE document_id=$1 ORDER BY applied_at DESC`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]*DocumentApply, 0)
	for rows.Next() {
		var a DocumentApply
		if err := rows.Scan(&a.ID, &a.DocumentID, &a.Target, &a.AppliedRows,
			&a.Detail, &a.AppliedBy, &a.AppliedAt); err != nil {
			return nil, err
		}
		list = append(list, &a)
	}
	return list, rows.Err()
}

// Delete 删除文档（级联删提取/审计），返回原件与解析件的对象 key 供调用方清理。
func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) (originalKey, parsedKey string, err error) {
	args := []any{id}
	q := `DELETE FROM documents WHERE id=$1`
	if org, scoped := OrgFilter(ctx); scoped {
		args = append(args, org)
		q += fmt.Sprintf(" AND org_id = $%d::uuid", len(args))
	}
	q += ` RETURNING COALESCE(original_object_key,''), COALESCE(parsed_object_key,'')`
	if err = r.pool.QueryRow(ctx, q, args...).Scan(&originalKey, &parsedKey); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return "", "", fmt.Errorf("文档不存在")
		}
		return "", "", err
	}
	return originalKey, parsedKey, nil
}

// LinkEntities 关联文档到业务实体（customer/contract/intent_customer）。
func (r *DocumentRepository) LinkEntities(ctx context.Context, id uuid.UUID, customerID, contractID, intentCustomerID string) error {
	args := []any{id}
	sets := []string{}
	if customerID != "" {
		args = append(args, customerID)
		sets = append(sets, fmt.Sprintf("customer_id = $%d::uuid", len(args)))
	}
	if contractID != "" {
		args = append(args, contractID)
		sets = append(sets, fmt.Sprintf("contract_id = $%d::uuid", len(args)))
	}
	if intentCustomerID != "" {
		args = append(args, intentCustomerID)
		sets = append(sets, fmt.Sprintf("intent_customer_id = $%d::uuid", len(args)))
	}
	if len(sets) == 0 {
		return nil
	}
	q := "UPDATE documents SET " + strings.Join(sets, ", ") + ", updated_at=now() WHERE id=$1"
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

// SetContractID 设置文档关联的合同ID（从文档创建合同后回写）。
func (r *DocumentRepository) SetContractID(ctx context.Context, docID, contractID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET contract_id=$2, updated_at=now() WHERE id=$1`, docID, contractID)
	return err
}

// parseValueTyped 尝试把文本值解析成数值/日期冗余列（带千分位与常见单位的容忍）。
func parseValueTyped(s string) (num any, date any) {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", "2006-1-2", "2006年01月02日", "2006年1月2日"} {
		if d, err := time.Parse(layout, t); err == nil {
			return nil, d
		}
	}
	clean := strings.NewReplacer(",", "", "，", "", " ", "").Replace(t)
	if f, err := strconv.ParseFloat(clean, 64); err == nil {
		return f, nil
	}
	return nil, nil
}
