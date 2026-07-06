// 多租户隔离回归：analytics(告警/特征) + customer_load + load_diagnosis 的读按活跃省过滤，
// A 省读不到 B 省数据。覆盖本轮隔离收口。
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMultiTenantIsolation(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	// schema 守卫：user_load_data 在则跑（真库已迁移）
	var ok bool
	if err := pool.QueryRow(ctx, "SELECT to_regclass('user_load_data') IS NOT NULL").Scan(&ok); err != nil || !ok {
		t.Skip("user_load_data 不存在，跳过")
	}

	mkOrg := func(tag string) string {
		var id string
		code := fmt.Sprintf("TST-ISO-%s-%d", tag, time.Now().UnixNano())
		if err := pool.QueryRow(ctx,
			"INSERT INTO organizations (code,name) VALUES ($1,$2) RETURNING id::text", code, "隔离测试"+tag).Scan(&id); err != nil {
			t.Fatalf("建组织失败: %v", err)
		}
		return id
	}
	orgA, orgB := mkOrg("A"), mkOrg("B")
	t.Cleanup(func() {
		for _, o := range []string{orgA, orgB} {
			for _, tbl := range []string{"customer_anomaly_alerts", "customer_characteristics", "user_load_data", "customer_diagnosis", "customers"} {
				_, _ = pool.Exec(ctx, "DELETE FROM "+tbl+" WHERE org_id=$1::uuid", o)
			}
			_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id=$1::uuid", o)
		}
	})

	mkCust := func(org, name string) uuid.UUID {
		var id string
		if err := pool.QueryRow(ctx,
			`INSERT INTO customers (user_name, org_id, lifecycle_stage) VALUES ($1,$2::uuid,'service') RETURNING id::text`,
			name, org).Scan(&id); err != nil {
			t.Fatalf("建客户失败: %v", err)
		}
		return uuid.MustParse(id)
	}
	custA, custB := mkCust(orgA, "A省客户"), mkCust(orgB, "B省客户")

	an := NewAnalyticsRepository(pool)
	lr := NewLoadRepository(pool)
	cl := NewCustomerLoadRepository(pool)
	ld := NewLoadDiagnosisRepository(pool)

	// 曲线含 25 个零点 → 触发 load_diagnosis「零点过多」异常，使其进入列表。
	curve := make([]float64, 96)
	for i := range curve {
		if i < 25 {
			curve[i] = 0
		} else {
			curve[i] = 100
		}
	}
	today := time.Now().Truncate(24 * time.Hour)

	seed := func(org string, cust uuid.UUID) {
		sc := WithOrg(ctx, org)
		if err := an.InsertAlert(sc, cust, today, "spike", "warn", "rule_spike", "测试", 80); err != nil {
			t.Fatalf("InsertAlert: %v", err)
		}
		if err := an.UpsertCharacteristic(sc, cust, today,
			json.RawMessage(`{}`), json.RawMessage(`{}`), []string{"x"}, 0.8, "A"); err != nil {
			t.Fatalf("UpsertCharacteristic: %v", err)
		}
		if err := lr.UpsertCurve(sc, cust, today.AddDate(0, 0, -1), curve, 100); err != nil {
			t.Fatalf("UpsertCurve: %v", err)
		}
	}
	seed(orgA, custA)
	seed(orgB, custB)

	// A 省读：必须见 custA、不得见 custB。
	hasOnly := func(label string, ids []string, want, notWant uuid.UUID) {
		t.Helper()
		seenWant, seenNot := false, false
		for _, id := range ids {
			if id == want.String() {
				seenWant = true
			}
			if id == notWant.String() {
				seenNot = true
			}
		}
		if !seenWant {
			t.Errorf("%s: 未见本省客户", label)
		}
		if seenNot {
			t.Errorf("%s: 串省泄漏——读到他省客户", label)
		}
	}
	scA := WithOrg(ctx, orgA)

	alerts, err := an.ListAlerts(scA, 100, true)
	if err != nil {
		t.Fatal(err)
	}
	ids := func() []string {
		out := []string{}
		for _, a := range alerts {
			out = append(out, a.CustomerID.String())
		}
		return out
	}()
	hasOnly("ListAlerts", ids, custA, custB)

	chars, err := an.ListLatestCharacteristics(scA, 100)
	if err != nil {
		t.Fatal(err)
	}
	cids := []string{}
	for _, x := range chars {
		cids = append(cids, x.CustomerID.String())
	}
	hasOnly("ListLatestCharacteristics", cids, custA, custB)

	sum, err := cl.Summary(scA, 14)
	if err != nil {
		t.Fatal(err)
	}
	sids := []string{}
	for _, s := range sum {
		sids = append(sids, s.CustomerID)
	}
	hasOnly("CustomerLoad.Summary", sids, custA, custB)

	diag, err := ld.List(scA, 14)
	if err != nil {
		t.Fatal(err)
	}
	dids := []string{}
	for _, d := range diag {
		dids = append(dids, d.CustomerID)
	}
	hasOnly("LoadDiagnosis.List", dids, custA, custB)

	// GetAlertStats 按省计数：A 省应只数到 1 条（custA 的告警）。
	stats, err := an.GetAlertStats(scA)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 1 {
		t.Errorf("GetAlertStats(A) 总数应为 1（仅本省），实际 %d", stats.Total)
	}
}
