// 现货价(day_ahead_spot_price)多租户隔离回归：A 省读不到 B 省的现货曲线。
package db

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSpotPriceIsolation(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	var ok bool
	if err := pool.QueryRow(ctx, "SELECT to_regclass('day_ahead_spot_price') IS NOT NULL").Scan(&ok); err != nil || !ok {
		t.Skip("day_ahead_spot_price 不存在，跳过")
	}

	mkOrg := func(tag string) string {
		var id string
		code := fmt.Sprintf("TST-SPOT-%s-%d", tag, time.Now().UnixNano())
		if err := pool.QueryRow(ctx,
			"INSERT INTO organizations (code,name) VALUES ($1,$2) RETURNING id::text", code, "现货隔离"+tag).Scan(&id); err != nil {
			t.Fatalf("建组织失败: %v", err)
		}
		return id
	}
	orgA, orgB := mkOrg("A"), mkOrg("B")
	t.Cleanup(func() {
		for _, o := range []string{orgA, orgB} {
			_, _ = pool.Exec(ctx, "DELETE FROM day_ahead_spot_price WHERE org_id=$1::uuid", o)
			_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id=$1::uuid", o)
		}
	})

	pr := NewPriceRepository(pool)
	day := time.Date(2099, 1, 2, 0, 0, 0, 0, time.UTC) // 未来日期，避开既有 demo
	seed := func(org string, price float64) {
		sc := WithOrg(ctx, org)
		for p := 1; p <= 48; p++ {
			if err := pr.UpsertDayAheadPrice(sc, day, p, price); err != nil {
				t.Fatalf("UpsertDayAheadPrice: %v", err)
			}
		}
	}
	seed(orgA, 100)
	seed(orgB, 200)

	// A 省读该日曲线：应为本省 48 点、全 100；若未隔离会混入 B 省 → 96 点或出现 200。
	curves, err := pr.GetRecentDayAheadCurves(WithOrg(ctx, orgA), day.AddDate(0, 0, 1), 10)
	if err != nil {
		t.Fatal(err)
	}
	var found *DailyPriceCurve
	for _, c := range curves {
		if c.Date.Format("2006-01-02") == day.Format("2006-01-02") {
			found = c
		}
	}
	if found == nil {
		t.Fatal("A 省未读到该日现货曲线")
	}
	if len(found.Curve48) != 48 {
		t.Errorf("曲线应为 48 点（仅本省），实际 %d（疑似混入他省）", len(found.Curve48))
	}
	for _, v := range found.Curve48 {
		if v != 100 {
			t.Errorf("A 省曲线含非本省价 %v（B 省=200）→ 串省泄漏", v)
			break
		}
	}

	// 写入也按省：B 省读应全 200、不见 100。
	cb, err := pr.GetRecentDayAheadCurves(WithOrg(ctx, orgB), day.AddDate(0, 0, 1), 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cb {
		if c.Date.Format("2006-01-02") == day.Format("2006-01-02") {
			for _, v := range c.Curve48 {
				if v != 200 {
					t.Errorf("B 省曲线含非本省价 %v", v)
					break
				}
			}
		}
	}
}
