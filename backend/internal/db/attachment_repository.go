// 文件附件仓储：file_attachments CRUD。
package db

import (
	"context"
	"time"
)

type FileAttachment struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id"`
	Filename    string    `json:"filename"`
	ObjectKey   string    `json:"object_key"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	UploadedBy  *string   `json:"uploaded_by,omitempty"`
	Note        *string   `json:"note,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type AttachmentRepository struct{ pool *Pool }

func NewAttachmentRepository(pool *Pool) *AttachmentRepository {
	return &AttachmentRepository{pool: pool}
}

type AttachmentInput struct {
	Resource    string
	ResourceID  string
	Filename    string
	ObjectKey   string
	ContentType string
	Size        int64
	UploadedBy  string
	Note        string
}

func (r *AttachmentRepository) Create(ctx context.Context, in AttachmentInput) (*FileAttachment, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO file_attachments
			(resource, resource_id, filename, object_key, content_type, size, uploaded_by, note)
		 VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''))
		 RETURNING id, resource, resource_id, filename, object_key, content_type, size, uploaded_by, note, created_at`,
		in.Resource, in.ResourceID, in.Filename, in.ObjectKey, in.ContentType, in.Size,
		in.UploadedBy, in.Note)
	var a FileAttachment
	if err := row.Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Filename, &a.ObjectKey,
		&a.ContentType, &a.Size, &a.UploadedBy, &a.Note, &a.CreatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttachmentRepository) ListByResource(ctx context.Context, resource, resourceID string) ([]*FileAttachment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, resource, resource_id, filename, object_key, content_type, size, uploaded_by, note, created_at
		 FROM file_attachments WHERE resource = $1 AND resource_id = $2
		 ORDER BY created_at DESC`, resource, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*FileAttachment, 0)
	for rows.Next() {
		var a FileAttachment
		if err := rows.Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Filename, &a.ObjectKey,
			&a.ContentType, &a.Size, &a.UploadedBy, &a.Note, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

func (r *AttachmentRepository) GetByID(ctx context.Context, id string) (*FileAttachment, error) {
	var a FileAttachment
	err := r.pool.QueryRow(ctx,
		`SELECT id, resource, resource_id, filename, object_key, content_type, size, uploaded_by, note, created_at
		 FROM file_attachments WHERE id = $1`, id).
		Scan(&a.ID, &a.Resource, &a.ResourceID, &a.Filename, &a.ObjectKey,
			&a.ContentType, &a.Size, &a.UploadedBy, &a.Note, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AttachmentRepository) Delete(ctx context.Context, id string) (string, error) {
	var objectKey string
	err := r.pool.QueryRow(ctx,
		`DELETE FROM file_attachments WHERE id = $1 RETURNING object_key`, id).Scan(&objectKey)
	return objectKey, err
}
