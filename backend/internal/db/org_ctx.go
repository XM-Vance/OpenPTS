package db

import "context"

type orgCtxKey struct{}

// WithOrg 把活跃省（组织 ID）注入 context，供 repo 层通过 OrgFromCtx 读取。
func WithOrg(ctx context.Context, org string) context.Context {
	return context.WithValue(ctx, orgCtxKey{}, org)
}

// OrgFromCtx 从 context 取出活跃省。
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
