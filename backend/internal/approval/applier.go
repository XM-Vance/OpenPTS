// 审批通过后的自动落库（applier）：按 resource 注册一个函数，approve 时调用。
// 解耦 approval handler 与各业务 repo —— 业务方在 main.go 注册自己的 applier。
package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor 抽象落库所需的查询能力。*db.Pool 与 pgx.Tx 都满足此接口，
// 让 Approve 能把 Transition 与 Apply 包在同一事务内（P0-B2 原子化）。
// 方法集与 db.Executor 一致，使两接口可互换（applier 收到的 tx 可直接喂给 repo 的 *Tx 方法）。
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// ApplyFunc 把审批 payload 应用到目标资源；返回 nil 表示落库成功。
// ex 是事务执行器（Approve 时为 pgx.Tx，非事务场景传 *db.Pool）。
type ApplyFunc func(ctx context.Context, ex Executor, resourceID string, payload json.RawMessage) error

type Registry struct {
	mu       sync.RWMutex
	appliers map[string]ApplyFunc
}

var ErrApplierNotRegistered = errors.New("未注册 applier")

func NewRegistry() *Registry {
	return &Registry{appliers: map[string]ApplyFunc{}}
}

func (r *Registry) Register(resource string, fn ApplyFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.appliers[resource] = fn
}

// Apply 根据 resource 名调用对应 applier；未注册的资源直接返回 nil（允许「只走流程不落库」）。
// ex 透传给 applier，使其在同一事务内落库。
func (r *Registry) Apply(ctx context.Context, ex Executor, resource, resourceID string, payload json.RawMessage) error {
	r.mu.RLock()
	fn, ok := r.appliers[resource]
	r.mu.RUnlock()
	if !ok {
		return nil // 静默放过：业务方未声明落库逻辑就只走审批流程
	}
	if err := fn(ctx, ex, resourceID, payload); err != nil {
		return fmt.Errorf("apply %s/%s: %w", resource, resourceID, err)
	}
	return nil
}
