// 权限检查服务：查询用户拥有的权限码集合，带 5 分钟缓存。
// 用户角色发生变更时必须调用 Invalidate。
package auth

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PermissionLister 是 PermissionService 依赖的数据查询接口。
// 由 db.PermissionRepository 实现。这里用接口解耦，避免 auth 包反向依赖 db 包。
type PermissionLister interface {
	ListUserPermissionCodes(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type PermissionService struct {
	repo  PermissionLister
	cache sync.Map // userID(uuid.UUID) -> cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	codes    map[string]struct{}
	cachedAt time.Time
}

func NewPermissionService(repo PermissionLister) *PermissionService {
	return &PermissionService{
		repo: repo,
		ttl:  5 * time.Minute,
	}
}

// GetUserPermissions 返回用户拥有的所有权限码集合（按集合 set 形式返回便于 O(1) 检查）。
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID uuid.UUID) (map[string]struct{}, error) {
	if v, ok := s.cache.Load(userID); ok {
		ce := v.(cacheEntry)
		if time.Since(ce.cachedAt) < s.ttl {
			return ce.codes, nil
		}
	}

	codes, err := s.repo.ListUserPermissionCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		set[c] = struct{}{}
	}

	s.cache.Store(userID, cacheEntry{codes: set, cachedAt: time.Now()})
	return set, nil
}

// Has 检查用户是否拥有某个权限。
func (s *PermissionService) Has(ctx context.Context, userID uuid.UUID, permCode string) (bool, error) {
	codes, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	_, ok := codes[permCode]
	return ok, nil
}

// Invalidate 清除某用户的缓存（用户角色变更后必须调用）。
func (s *PermissionService) Invalidate(userID uuid.UUID) {
	s.cache.Delete(userID)
}
