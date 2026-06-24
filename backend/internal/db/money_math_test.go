// 结算金额数学真库测试（P1-9）。
// 连接 DATABASE_URL 指向的已迁移 Postgres（未设置时整体跳过——CI 在 migrate up 后执行）。
// 所有数据写入临时组织并在测试后清理，不污染既有数据；同时覆盖多租户隔离语义。
package db

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

func testPool(t *testing.T) *Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL 未设置，跳过真库测试")
	}
	pool, err := New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	t.Cleanup(pool.Close)
	// schema 未迁移时直接跳过（避免在裸库上误报失败）
	var ok bool
	if err := pool.QueryRow(context.Background(),
		"SELECT to_regclass('deviation_settlement') IS NOT NULL").Scan(&ok); err != nil || !ok {
		t.Skip("deviation_settlement 表不存在（schema 未迁移），跳过")
	}
	return pool
}

// createTempOrg 建一个一次性组织并注册清理（先删业务行再删组织，LIFO 顺序安全）。
func createTempOrg(t *testing.T, pool *Pool, tag string) string {
	t.Helper()
	code := fmt.Sprintf("TST-%s-%d", tag, time.Now().UnixNano())
	var id string
	if err := pool.QueryRow(context.Background(),
		"INSERT INTO organizations (code, name) VALUES ($1, $2) RETURNING id::text",
		code, "金额测试临时组织 "+tag).Scan(&id); err != nil {
		t.Fatalf("创建临时组织失败: %v", err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		for _, tbl := range []string{"deviation_settlement", "pre_settlement_daily"} {
			if _, err := pool.Exec(ctx, "DELETE FROM "+tbl+" WHERE org_id = $1::uuid", id); err != nil {
				t.Errorf("清理 %s 失败: %v", tbl, err)
			}
		}
		if _, err := pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1::uuid", id); err != nil {
			t.Errorf("清理临时组织失败: %v", err)
		}
	})
	return id
}

func almostEq(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// ─── 偏差结算:Summary 聚合数学 + 多租户隔离 ───

func TestDeviationSummaryMathAndOrgIsolation(t *testing.T) {
	pool := testPool(t)
	repo := NewDeviationRepository(pool)
	orgA := createTempOrg(t, pool, "DEVA")
	orgB := createTempOrg(t, pool, "DEVB")

	ins := func(org string, daysAgo int, cat string, declared, deviation, devCost, penalty float64) {
		t.Helper()
		d := time.Now().AddDate(0, 0, -daysAgo).Truncate(24 * time.Hour)
		actual := declared + deviation
		rate := deviation / declared * 100
		if _, err := pool.Exec(context.Background(), `
			INSERT INTO deviation_settlement
			  (operating_date, declared_energy_mwh, actual_energy_mwh, deviation_energy_mwh,
			   deviation_rate, deviation_cost, penalty_cost, total_settlement, category, org_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid)`,
			d, declared, actual, deviation, rate, devCost, penalty, devCost+penalty, cat, org); err != nil {
			t.Fatalf("插入夹具失败: %v", err)
		}
	}
	// org A:day_ahead 两行（偏差 +100/-40），real_time 一行；org B 一行干扰数据
	ins(orgA, 1, "day_ahead", 2000, 100, 35000, 5000) // rate=5%, total=40000
	ins(orgA, 2, "day_ahead", 1000, -40, -14000, 0)   // rate=-4%, total=-14000
	ins(orgA, 3, "real_time", 500, 25, 9000, 1250)    // rate=5%, total=10250
	ins(orgB, 1, "day_ahead", 8888, 888, 88888, 8888)

	ctxA := WithOrg(context.Background(), orgA)
	sums, err := repo.Summary(ctxA, 30)
	if err != nil {
		t.Fatalf("Summary 失败: %v", err)
	}
	if len(sums) != 2 {
		t.Fatalf("org A 应有 2 个分类，得到 %d", len(sums))
	}
	byCat := map[string]*DeviationSummary{}
	for _, s := range sums {
		byCat[s.Category] = s
	}
	da := byCat["day_ahead"]
	if da == nil || da.Count != 2 {
		t.Fatalf("day_ahead 应 2 行，得到 %+v", da)
	}
	if !almostEq(da.TotalDeviationEnergy, 60) { // 100 + (-40)
		t.Errorf("day_ahead 偏差电量合计应 60，得到 %v", da.TotalDeviationEnergy)
	}
	if !almostEq(da.TotalCost, 26000) { // 40000 + (-14000)
		t.Errorf("day_ahead 结算合计应 26000，得到 %v", da.TotalCost)
	}
	if !almostEq(da.AvgDeviationRate, 0.5) { // (5% + -4%)/2
		t.Errorf("day_ahead 平均偏差率应 0.5，得到 %v", da.AvgDeviationRate)
	}
	rt := byCat["real_time"]
	if rt == nil || rt.Count != 1 || !almostEq(rt.TotalCost, 10250) {
		t.Errorf("real_time 汇总不符: %+v", rt)
	}

	// 多租户隔离:org B 视角只看到自己 1 行；org A 的 List 看不到 8888
	listB, err := repo.List(WithOrg(context.Background(), orgB), "", 30)
	if err != nil || len(listB) != 1 || !almostEq(listB[0].DeclaredEnergy, 8888) {
		t.Errorf("org B 应只见自己 1 行，得到 %d 行 err=%v", len(listB), err)
	}
	listA, err := repo.List(ctxA, "day_ahead", 30)
	if err != nil || len(listA) != 2 {
		t.Errorf("org A day_ahead 应 2 行，得到 %d err=%v", len(listA), err)
	}
}

// ─── 偏差结算:演示数据生成器的金额自洽性 ───

func TestDeviationGenerateDemoInvariants(t *testing.T) {
	pool := testPool(t)
	repo := NewDeviationRepository(pool)
	org := createTempOrg(t, pool, "DEVG")
	ctx := WithOrg(context.Background(), org)

	n, err := repo.GenerateDemo(ctx)
	if err != nil {
		t.Fatalf("GenerateDemo 失败: %v", err)
	}
	if n != 90 { // 30 天 × 3 分类
		t.Fatalf("应生成 90 行，得到 %d", n)
	}
	rows, err := repo.List(ctx, "", 31)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(rows) < 80 { // LIMIT 200 内,30 天窗口应接近 90(当天截断容差)
		t.Fatalf("回读行数过少: %d", len(rows))
	}
	for _, r := range rows {
		if !almostEq(r.ActualEnergy-r.DeclaredEnergy, r.DeviationEnergy) {
			t.Errorf("%s/%s: 实际-申报 != 偏差 (%v - %v != %v)",
				r.OperatingDate.Format("01-02"), r.Category, r.ActualEnergy, r.DeclaredEnergy, r.DeviationEnergy)
		}
		if !almostEq(r.DeviationEnergy/r.DeclaredEnergy*100, r.DeviationRate) {
			t.Errorf("偏差率不自洽: %v vs %v", r.DeviationEnergy/r.DeclaredEnergy*100, r.DeviationRate)
		}
		if !almostEq(r.DeviationCost+r.PenaltyCost, r.TotalSettlement) {
			t.Errorf("偏差费+考核费 != 总结算 (%v + %v != %v)", r.DeviationCost, r.PenaltyCost, r.TotalSettlement)
		}
		// 考核费恒非负(按偏差绝对值计),且超 ±5% 时为 |偏差|×50、否则为 0。
		if r.PenaltyCost < 0 {
			t.Errorf("考核费不应为负(倒贴奖励): %v (rate=%v)", r.PenaltyCost, r.DeviationRate)
		}
		if r.DeviationRate > 5 || r.DeviationRate < -5 {
			if !almostEq(r.PenaltyCost, math.Abs(r.DeviationEnergy)*50) {
				t.Errorf("超限考核费应为 |偏差|×50: %v vs %v", r.PenaltyCost, math.Abs(r.DeviationEnergy)*50)
			}
		} else if !almostEq(r.PenaltyCost, 0) {
			t.Errorf("±5%% 内不应有考核费,得到 %v (rate=%v)", r.PenaltyCost, r.DeviationRate)
		}
	}
}

// ─── 预结算:演示数据生成器的金额自洽性(逐 96 点重算) ───

func TestPreSettleGenerateDemoInvariants(t *testing.T) {
	pool := testPool(t)
	repo := NewPreSettleRepository(pool)
	org := createTempOrg(t, pool, "PRE")
	ctx := WithOrg(context.Background(), org)

	n, err := repo.GenerateDemo(ctx)
	if err != nil {
		t.Fatalf("GenerateDemo 失败: %v", err)
	}
	if n != 14 {
		t.Fatalf("应生成 14 行，得到 %d", n)
	}
	rows, err := repo.List(ctx, 15)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(rows) < 13 {
		t.Fatalf("回读行数过少: %d", len(rows))
	}
	for _, p := range rows {
		if len(p.DeclaredCurve96) != 96 || len(p.ClearedCurve96) != 96 || len(p.SpotPrice96) != 96 {
			t.Fatalf("%s: 曲线应为 96 点", p.OperatingDate.Format("01-02"))
		}
		var dec, clr, rev float64
		for i := 0; i < 96; i++ {
			dec += p.DeclaredCurve96[i] * 0.25
			clr += p.ClearedCurve96[i] * 0.25
			rev += p.ClearedCurve96[i] * 0.25 * p.SpotPrice96[i]
		}
		if !almostEq(dec, p.TotalDeclared) || !almostEq(clr, p.TotalCleared) {
			t.Errorf("电量合计与曲线积分不符: dec %v/%v clr %v/%v", dec, p.TotalDeclared, clr, p.TotalCleared)
		}
		if math.Abs(rev-p.EnergyRevenue) > 1e-3 { // 96 点乘加,放宽到毫元级
			t.Errorf("电能收入与逐点重算不符: %v vs %v", rev, p.EnergyRevenue)
		}
		dev := p.TotalCleared - p.TotalDeclared
		if !almostEq(dev, p.TotalDeviation) {
			t.Errorf("偏差电量不自洽: %v vs %v", dev, p.TotalDeviation)
		}
		if !almostEq(p.TotalDeviation/p.TotalDeclared, p.DeviationRatio) {
			t.Errorf("偏差率不自洽")
		}
		wantPenalty := 0.0
		if math.Abs(p.DeviationRatio) > 0.05 {
			wantPenalty = math.Abs(p.TotalDeviation) * 100
		}
		if math.Abs(wantPenalty-p.DeviationPenalty) > 1e-3 {
			t.Errorf("偏差考核费不符: 应 %v 得 %v (ratio=%v)", wantPenalty, p.DeviationPenalty, p.DeviationRatio)
		}
		if math.Abs((p.EnergyRevenue-p.DeviationPenalty)-p.FinalAmount) > 1e-3 {
			t.Errorf("最终金额应 = 收入 - 考核: %v - %v != %v", p.EnergyRevenue, p.DeviationPenalty, p.FinalAmount)
		}
	}
}
