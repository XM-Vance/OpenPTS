package db

import (
	"context"
)

type orgCtxKey struct{}

// WithOrg 把活跃组织（组织 ID）注入 context，供 repo 层通过 OrgFromCtx 读取。
func WithOrg(ctx context.Context, org string) context.Context {
	return context.WithValue(ctx, orgCtxKey{}, org)
}

// OrgFromCtx 从 context 取出活跃组织。
func OrgFromCtx(ctx context.Context) string {
	s, _ := ctx.Value(orgCtxKey{}).(string)
	return s
}

// OrgFilter 返回 (org, scoped)。
// scoped=false 表示总部「全部省」（"*" 或空）：查询不加 org 过滤。
// scoped=true 表示具体省：查询需加 org_id 过滤。
func OrgFilter(ctx context.Context) (org string, scoped bool) {
	o := OrgFromCtx(ctx)
	if o == "" || o == "*" {
		return "", false
	}
	return o, true
}

// MustScoped 是写操作专用的 OrgFilter：要求已选定具体省，否则返回 ErrOrgRequired。
// 写方法不允许在「全部省」下落库（会跨租户污染），故统一用此 helper 收口，
// 替代散落在各 repo 写方法开头的三行样板：
//
//	org, scoped := OrgFilter(ctx)
//	if !scoped { return ErrOrgRequired }
//
// 改为一行（返回签名带 error 的位置直接透传 err）：
//
//	org, err := MustScoped(ctx)
//	if err != nil { return err }
func MustScoped(ctx context.Context) (org string, err error) {
	o, scoped := OrgFilter(ctx)
	if !scoped {
		return "", ErrOrgRequired
	}
	return o, nil
}

// OrgOrFJ 返回当前活跃组织 org_id；若未选定（总部「全部省」），解析 FJ 省的 org_id。
// 用于 GenerateDemo / 回填场景：演示数据必须有归属省，总部调用时回退 FJ。
// 注意：会查库解析 FJ 的 UUID，故仅用于写演示/回填，不要在热路径查询里调用。
// 这是 P2-1 收口：替代散落在 11+ repo GenerateDemo 里的「OrgFilter + 查 FJ」三行样板。
func OrgOrFJ(ctx context.Context, pool *Pool) (string, error) {
	org, scoped := OrgFilter(ctx)
	if scoped {
		return org, nil
	}
	var fjID string
	if err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE code='default'").Scan(&fjID); err != nil {
		return "", err
	}
	return fjID, nil
}
