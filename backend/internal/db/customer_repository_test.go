// 客户档案真库集成测试（P1-9）。
// 复用 money_math_test.go 的 testPool / createTempOrg 辅助（同 package）。
// 覆盖：多租户隔离（核心铁律）、ErrOrgRequired 写入保护、CRUD 往返。
package db

import (
	"context"
	"errors"
	"testing"
)

// newCustomerInput 构造一个合法的 CustomerInput（调用方覆盖关键字段）。
func newCustomerInput(name string) CustomerInput {
	return CustomerInput{
		UserName:  name,
		ShortName: name,
		Location:  "测试地区",
		Source:    "测试来源",
		Manager:   "测试经理",
		Tags:      []string{"测试"},
	}
}

// ─── 写入保护：「全部省」上下文 Create 必须返回 ErrOrgRequired ───

func TestCustomerCreateRejectsAllOrgScope(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepository(pool)

	// 无 org（HQ「全部省」）上下文，Create 必须被拒。
	_, err := repo.Create(context.Background(), newCustomerInput("被拒客户"), nil)
	if !errors.Is(err, ErrOrgRequired) {
		t.Fatalf("HQ 上下文 Create 应返回 ErrOrgRequired，得到 %v", err)
	}
}

// ─── 多租户隔离 + CRUD 往返 ───

func TestCustomerCRUDAndOrgIsolation(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomerRepository(pool)
	orgA := createTempOrg(t, pool, "CUSA")
	orgB := createTempOrg(t, pool, "CUSB")

	// 在 orgA 创建两个客户，orgB 创建一个干扰客户。
	ctxA := WithOrg(context.Background(), orgA)
	ctxB := WithOrg(context.Background(), orgB)

	// 注册客户行清理（createTempOrg 的 LIFO 清理只删业务表，不含 customers；
	// 客户外键会阻止删组织，故测试结束先删本测试创建的客户）。
	cleanupCustomers := func() {
		for _, oid := range []string{orgA, orgB} {
			if _, err := pool.Exec(context.Background(),
				"DELETE FROM customers WHERE org_id = $1::uuid", oid); err != nil {
				t.Errorf("清理客户失败(org=%s): %v", oid, err)
			}
		}
	}

	cust1, err := repo.Create(ctxA, newCustomerInput("租户A客户1"), nil)
	if err != nil {
		t.Fatalf("orgA Create 失败: %v", err)
	}
	// 注意：customerColumns 当前不含 org_id（scan 不回填），直接查库验证归属。
	var orgID *string
	if err := pool.QueryRow(context.Background(),
		"SELECT org_id::text FROM customers WHERE id = $1", cust1.ID).Scan(&orgID); err != nil {
		t.Fatalf("查客户 org_id 失败: %v", err)
	}
	if orgID == nil || *orgID != orgA {
		t.Fatalf("新建客户 org_id 应为 orgA，得到 %v", orgID)
	}
	cust2, err := repo.Create(ctxA, newCustomerInput("租户A客户2"), nil)
	if err != nil {
		t.Fatalf("orgA Create 第二个失败: %v", err)
	}
	if _, err := repo.Create(ctxB, newCustomerInput("租户B客户"), nil); err != nil {
		t.Fatalf("orgB Create 失败: %v", err)
	}
	// 所有客户创建完毕后注册清理（t.Cleanup 是 LIFO，先于 createTempOrg 的组织清理执行）。
	t.Cleanup(cleanupCustomers)

	// List 多租户隔离：orgA 只看到自己的 2 个。
	listA, totalA, err := repo.List(ctxA, CustomerListFilter{Limit: 100})
	if err != nil {
		t.Fatalf("orgA List 失败: %v", err)
	}
	if totalA < 2 {
		t.Errorf("orgA 至少 2 个客户，得到 %d", totalA)
	}
	// 确认 orgA 列表里没有 orgB 的客户。
	for _, c := range listA {
		if c.OrgID != nil && *c.OrgID == orgB {
			t.Errorf("orgA List 泄漏了 orgB 的客户: %s", c.UserName)
		}
	}

	// GetByID 多租户隔离：orgB 视角取 orgA 的客户应取不到（空结果/ErrCustomerNotFound）。
	_, err = repo.GetByID(ctxB, cust1.ID)
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("orgB 取 orgA 客户应返回 ErrCustomerNotFound，得到 err=%v", err)
	}
	// orgA 视角能取到自己的客户。
	got, err := repo.GetByID(ctxA, cust1.ID)
	if err != nil {
		t.Fatalf("orgA GetByID 失败: %v", err)
	}
	if got.UserName != cust1.UserName {
		t.Errorf("GetByID 往返不一致: %s vs %s", got.UserName, cust1.UserName)
	}

	// Update 往返 + 多租户作用域：orgA 改自己客户成功；orgB 改 orgA 客户应失败。
	upd := newCustomerInput("租户A客户1-已改名")
	upd.Tags = []string{"测试", "已更新"}
	updated, err := repo.Update(ctxA, cust1.ID, upd)
	if err != nil {
		t.Fatalf("orgA Update 失败: %v", err)
	}
	if updated.UserName != "租户A客户1-已改名" {
		t.Errorf("Update 后名称不符: %s", updated.UserName)
	}
	// orgB 尝试更新 orgA 的客户 → 应返回 ErrCustomerNotFound（org 过滤后查无此行）。
	_, err = repo.Update(ctxB, cust1.ID, newCustomerInput("越权改名"))
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("orgB 改 orgA 客户应被拒（ErrCustomerNotFound），得到 err=%v", err)
	}

	// Delete 多租户作用域：orgB 删 orgA 客户应失败。
	if err := repo.Delete(ctxB, cust1.ID); !errors.Is(err, ErrCustomerNotFound) {
		t.Errorf("orgB 删 orgA 客户应被拒，得到 err=%v", err)
	}
	// orgA 删自己客户成功。
	if err := repo.Delete(ctxA, cust2.ID); err != nil {
		t.Errorf("orgA Delete 失败: %v", err)
	}
}
