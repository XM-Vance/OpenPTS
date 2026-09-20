// 计量数据底座真库测试（WP0.3）：导入幂等 + 多表计聚合 + unified_load_curve 镜像。
// 连接 DATABASE_URL 指向的已迁移 Postgres（未设置时整体跳过——CI 在 migrate up 后执行）。
// 数据写入临时组织/临时客户并在测试后清理。
package db

import (
	"context"
	"testing"
	"time"
)

func TestMeterImportAndAggregate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	org := createTempOrg(t, pool, "meter")

	var custID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO customers (user_name, org_id, lifecycle_stage)
		 VALUES ('表计聚合测试客户', $1::uuid, 'service') RETURNING id::text`,
		org).Scan(&custID); err != nil {
		t.Fatalf("建临时客户失败: %v", err)
	}
	t.Cleanup(func() {
		// LIFO：先删引用 customers 的行，再删客户
		pool.Exec(ctx, "DELETE FROM user_load_data WHERE customer_id = $1::uuid", custID)
		pool.Exec(ctx, "DELETE FROM unified_load_curve WHERE customer_id = $1::uuid", custID)
		pool.Exec(ctx, "DELETE FROM raw_meter_data WHERE meter_id LIKE 'TSTMTR-%'")
		pool.Exec(ctx, "DELETE FROM customers WHERE id = $1::uuid", custID)
	})

	sctx := WithOrg(ctx, org)
	repo := NewMeterRepository(pool)
	d := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	curve := func(v float64) []float64 {
		c := make([]float64, 96)
		for i := range c {
			c[i] = v
		}
		return c
	}

	// 两块表同一客户同一天：M-1 示值 100×倍率2 = 200；M-2 示值 50×倍率1 = 50 → 合计 250kW/点
	rows := []*RawMeterRow{
		{MeterID: "TSTMTR-1", CustomerID: custID, Date: d, Multiplier: 2, Curve96: curve(100)},
		{MeterID: "TSTMTR-2", CustomerID: custID, Date: d, Multiplier: 1, Curve96: curve(50)},
	}
	if n, err := repo.BulkUpsertRaw(sctx, rows); err != nil || n != 2 {
		t.Fatalf("首次导入失败: n=%d err=%v", n, err)
	}
	// 幂等：重复导入同 (表号,日期) 不堆行
	if _, err := repo.BulkUpsertRaw(sctx, rows[:1]); err != nil {
		t.Fatalf("重复导入失败: %v", err)
	}
	var cnt int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM raw_meter_data WHERE meter_id LIKE 'TSTMTR-%'").Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	if cnt != 2 {
		t.Fatalf("幂等失效：应 2 行，got %d", cnt)
	}

	written, skipped, err := repo.AggregateMeters(sctx, d, d)
	if err != nil || skipped != 0 {
		t.Fatalf("聚合失败: written=%d skipped=%d err=%v", written, skipped, err)
	}
	if written != 1 {
		t.Fatalf("应聚合 1 个客户日，got %d", written)
	}

	var total float64
	var curveLen int
	if err := pool.QueryRow(ctx,
		`SELECT total_load, cardinality(curve_96) FROM user_load_data
		 WHERE customer_id = $1::uuid AND date = $2`, custID, d).Scan(&total, &curveLen); err != nil {
		t.Fatalf("user_load_data 未写入: %v", err)
	}
	if curveLen != 96 || total < 5999.9 || total > 6000.1 { // 250kW × Σ96点/4 = 6000 kWh
		t.Errorf("聚合电量应为 6000 kWh（250kW×24h），got %v（len=%d）", total, curveLen)
	}
	var method string
	var meters int
	if err := pool.QueryRow(ctx,
		`SELECT aggregation_method, (extra->>'meters')::int FROM unified_load_curve
		 WHERE customer_id = $1::uuid AND date = $2`, custID, d).Scan(&method, &meters); err != nil {
		t.Fatalf("unified_load_curve 未写入: %v", err)
	}
	if method != "meter_sum" || meters != 2 {
		t.Errorf("unified 标记应为 meter_sum/2 表计，got %s/%d", method, meters)
	}
}
