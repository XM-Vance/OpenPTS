// 审批通过后的自动落库（applier）：按 resource 注册一个函数，approve 时调用。
// 解耦 approval handler 与各业务 repo —— 业务方在 main.go 注册自己的 applier。
package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// ApplyFunc 把审批 payload 应用到目标资源；返回 nil 表示落库成功。
type ApplyFunc func(ctx context.Context, resourceID string, payload json.RawMessage) error

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
func (r *Registry) Apply(ctx context.Context, resource, resourceID string, payload json.RawMessage) error {
	r.mu.RLock()
	fn, ok := r.appliers[resource]
	r.mu.RUnlock()
	if !ok {
		return nil // 静默放过：业务方未声明落库逻辑就只走审批流程
	}
	if err := fn(ctx, resourceID, payload); err != nil {
		return fmt.Errorf("apply %s/%s: %w", resource, resourceID, err)
	}
	return nil
}
