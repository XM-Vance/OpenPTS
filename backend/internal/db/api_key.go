// 用户 API Key 仓储：生成记录、按前缀反查（exchange）、列表（脱敏）、吊销。
// 明文 Key 只在 handler 创建时返回一次；本仓储只存 prefix + bcrypt 哈希。
package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrAPIKeyNotFound 按 prefix 未查到有效 Key。
var ErrAPIKeyNotFound = errors.New("api key not found")

// UserAPIKey 列表/展示用（脱敏，无 hash）。
type UserAPIKey struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// UserAPIKeyWithHash exchange 反查用（含 user_id + hash，内部使用，不外泄）。
type UserAPIKeyWithHash struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	KeyHash   string
	ExpiresAt *time.Time
	RevokedAt *time.Time
}

type APIKeyRepository struct{ pool *Pool }

func NewAPIKeyRepository(pool *Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

// Create 存一条 Key 记录（prefix + hash）。返回记录 id。
func (r *APIKeyRepository) Create(ctx context.Context, userID uuid.UUID, name, prefix, hash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx,
		`INSERT INTO user_api_keys (user_id, name, key_prefix, key_hash)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, name, prefix, hash).Scan(&id)
	return id, err
}

// ListByUser 列出某用户的所有 Key（按创建时间倒序，脱敏无 hash）。
func (r *APIKeyRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*UserAPIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, key_prefix, last_used_at, expires_at, revoked_at, created_at
		 FROM user_api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*UserAPIKey{}
	for rows.Next() {
		var k UserAPIKey
		if err := rows.Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.LastUsedAt,
			&k.ExpiresAt, &k.RevokedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &k)
	}
	return list, rows.Err()
}

// GetByPrefix 按前缀反查（exchange 用）。只返回未吊销的。
// 注意：可能有哈希碰撞的前缀冲突——调用方需对每条候选 bcrypt 校验，取匹配的。
func (r *APIKeyRepository) GetByPrefix(ctx context.Context, prefix string) ([]*UserAPIKeyWithHash, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, key_hash, expires_at, revoked_at
		 FROM user_api_keys WHERE key_prefix = $1 AND revoked_at IS NULL`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []*UserAPIKeyWithHash{}
	for rows.Next() {
		var k UserAPIKeyWithHash
		if err := rows.Scan(&k.ID, &k.UserID, &k.KeyHash, &k.ExpiresAt, &k.RevokedAt); err != nil {
			return nil, err
		}
		list = append(list, &k)
	}
	return list, rows.Err()
}

// UpdateLastUsed 更新最后使用时间（exchange 成功后调）。
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE user_api_keys SET last_used_at = now() WHERE id = $1", id)
	return err
}

// Revoke 吊销（只允许 Key 归属用户自己吊销自己的）。
func (r *APIKeyRepository) Revoke(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE user_api_keys SET revoked_at = now()
		 WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}
