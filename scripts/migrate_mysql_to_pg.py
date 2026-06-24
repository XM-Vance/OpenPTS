#!/usr/bin/env python3
"""
OpenPTS V1 数据迁移脚本：MySQL (atr_ele/forecast_data/fujian_biz) → PostgreSQL (ptis)

依赖安装：
    pip install -r scripts/requirements.txt

环境变量（全部必填）：
    MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASSWORD
    DATABASE_URL          — postgresql://user:password@host:port/dbname

用法：
    python scripts/migrate_mysql_to_pg.py                  # 全量迁移
    python scripts/migrate_mysql_to_pg.py --dry-run        # 只统计不写入
    python scripts/migrate_mysql_to_pg.py --only 日前价格  # 只跑指定步骤
    python scripts/migrate_mysql_to_pg.py --strict         # 任何关键步骤失败则非零退出
"""

import json, sys, os, argparse, traceback
from datetime import datetime, timedelta
from collections import defaultdict
from urllib.parse import urlparse, unquote

import pymysql
import psycopg2
import psycopg2.extras

psycopg2.extras.register_uuid()

# ═══════════════════════════════════════════════════════════════
# B1: 配置外置 — 全部从环境变量读取
# ═══════════════════════════════════════════════════════════════
def _load_mysql_cfg() -> dict:
    missing = [k for k in ("MYSQL_HOST", "MYSQL_PORT", "MYSQL_USER", "MYSQL_PASSWORD") if not os.environ.get(k)]
    if missing:
        print(f"❌ 缺少环境变量: {', '.join(missing)}", file=sys.stderr)
        sys.exit(1)
    return dict(
        host=os.environ["MYSQL_HOST"],
        port=int(os.environ["MYSQL_PORT"]),
        user=os.environ["MYSQL_USER"],
        password=os.environ["MYSQL_PASSWORD"],
        charset="utf8mb4",
    )


def _load_pg_cfg() -> dict:
    database_url = os.environ.get("DATABASE_URL")
    if not database_url:
        print("❌ 缺少环境变量: DATABASE_URL", file=sys.stderr)
        sys.exit(1)
    # DATABASE_URL 格式：postgresql://user:password@host:port/dbname
    u = urlparse(database_url)
    return dict(
        host=u.hostname,
        port=u.port or 5432,
        user=u.username,
        password=unquote(u.password or ""),
        dbname=u.path.lstrip("/"),
    )


MYSQL_CFG = _load_mysql_cfg()
PG_CFG = _load_pg_cfg()

# ── 全局参数 ──
DRY_RUN = False
STRICT = False

# ═══════════════════════════════════════════════════════════════
# 连接工具
# ═══════════════════════════════════════════════════════════════
def mysql_cur(db="atr_ele"):
    conn = pymysql.connect(**{**MYSQL_CFG, "database": db, "cursorclass": pymysql.cursors.DictCursor})
    return conn, conn.cursor()


def pg_conn():
    return psycopg2.connect(**PG_CFG)


def busi_date_to_period(s: str):
    """'2025-03-01 00:15:00' or '2025-03-01 24:00:00' → ('2025-03-01', period_int 0-95)"""
    if not s:
        return None, None
    s = s.strip()
    if s.endswith("24:00:00"):
        dt = datetime.strptime(s.replace("24:00:00", "00:00:00"), "%Y-%m-%d %H:%M:%S")
        dt += timedelta(days=1)
        p = 0
        return dt.strftime("%Y-%m-%d"), p
    dt = datetime.strptime(s, "%Y-%m-%d %H:%M:%S")
    p = (dt.hour * 60 + dt.minute) // 15
    return dt.strftime("%Y-%m-%d"), p


# ═══════════════════════════════════════════════════════════════
# B5: 游标分批读取工具
# ═══════════════════════════════════════════════════════════════
BATCH_SIZE = 2000


def mysql_count(cur, table: str, where: str = "") -> int:
    cur.execute(f"SELECT COUNT(*) as cnt FROM {table} {where}")
    return cur.fetchone()["cnt"]


def mysql_fetch_batches(cur, table: str, cols: str = "*", where: str = "", order_by: str = ""):
    """游标分批从 MySQL 读取，yield 每批 rows"""
    offset = 0
    while True:
        sql = f"SELECT {cols} FROM {table} {where} {order_by} LIMIT {BATCH_SIZE} OFFSET {offset}"
        cur.execute(sql)
        rows = cur.fetchall()
        if not rows:
            break
        yield rows
        offset += BATCH_SIZE


# ═══════════════════════════════════════════════════════════════
# Step 1: 日前价格
# ═══════════════════════════════════════════════════════════════
def step_price_da() -> dict:
    print("\n=== 1 日前现货价格 → day_ahead_spot_price ===")
    conn, cur = mysql_cur()
    total = mysql_count(cur, "t_bus_day_ahead_marginal_price")
    print(f"  MySQL: {total} 条")
    conn.close()

    if DRY_RUN:
        return {"mysql源": total, "pg写入": 0}

    written = 0
    pg = pg_conn()
    c = pg.cursor()

    conn, cur = mysql_cur()
    for batch_rows in mysql_fetch_batches(cur, "t_bus_day_ahead_marginal_price", order_by="ORDER BY busiDate"):
        batch = []
        for r in batch_rows:
            d, p = busi_date_to_period(r["busiDate"])
            if d is None:
                continue
            v = float(r["value"]) if r["value"] is not None else None
            if v is None:  # B6: 仅 NULL 跳过，0 和负值保留
                continue
            batch.append((d, p, v))
        if batch:
            psycopg2.extras.execute_values(c,
                """INSERT INTO day_ahead_spot_price (date, period, price_da) VALUES %s
                   ON CONFLICT (date, period) DO UPDATE SET price_da = EXCLUDED.price_da""",
                batch, template="(%s::date, %s::int, %s::numeric)")
            written += len(batch)
        print(f"  进度: {written} 条写入")
    conn.close()

    pg.commit()
    print(f"  PG: {written} 条写入")
    c.close(); pg.close()
    return {"mysql源": total, "pg写入": written}


# ═══════════════════════════════════════════════════════════════
# Step 2: 实时价格
# ═══════════════════════════════════════════════════════════════
def step_price_rt() -> dict:
    print("\n=== 2 实时价格 → real_time_spot_price ===")
    conn, cur = mysql_cur()
    total = mysql_count(cur, "t_bus_marginal_price")
    print(f"  MySQL: {total} 条")
    conn.close()

    if DRY_RUN:
        return {"mysql源": total, "pg写入": 0}

    written = 0
    pg = pg_conn()
    c = pg.cursor()

    conn, cur = mysql_cur()
    for batch_rows in mysql_fetch_batches(cur, "t_bus_marginal_price", order_by="ORDER BY busiDate"):
        batch = []
        for r in batch_rows:
            d, p = busi_date_to_period(r["busiDate"])
            if d is None:
                continue
            v = float(r["value"]) if r["value"] is not None else None
            if v is None:  # B6: 仅 NULL 跳过
                continue
            batch.append((d, p, v))
        if batch:
            psycopg2.extras.execute_values(c,
                """INSERT INTO real_time_spot_price (date, period, price_rt) VALUES %s
                   ON CONFLICT (date, period) DO UPDATE SET price_rt = EXCLUDED.price_rt""",
                batch, template="(%s::date, %s::int, %s::numeric)")
            written += len(batch)
        print(f"  进度: {written} 条写入")
    conn.close()

    pg.commit()
    print(f"  PG: {written} 条写入")
    c.close(); pg.close()
    return {"mysql源": total, "pg写入": written}


# ═══════════════════════════════════════════════════════════════
# Step 3: 负荷数据 → total_load_daily (curve_96)
# ═══════════════════════════════════════════════════════════════
def step_load() -> dict:
    print("\n=== 3 负荷数据 → total_load_daily ===")
    conn, cur = mysql_cur()
    total = mysql_count(cur, "t_bus_load_forecast_date")
    print(f"  负荷预测 MySQL: {total} 条")
    conn.close()

    if DRY_RUN:
        return {"mysql源": total, "pg写入": 0}

    # 仍需全量聚合，按批读入 daily dict
    daily = defaultdict(lambda: [0.0] * 96)
    conn, cur = mysql_cur()
    for batch_rows in mysql_fetch_batches(cur, "t_bus_load_forecast_date", order_by="ORDER BY busiDate"):
        for r in batch_rows:
            d, p = busi_date_to_period(r["busiDate"])
            if d is None:
                continue
            v = float(r["value"]) if r["value"] is not None else 0
            daily[d][p] = v
    conn.close()

    batch = []
    for date_str in sorted(daily):
        curve = daily[date_str]
        # 全为零 → 无效数据日，跳过
        if all(x == 0.0 for x in curve):
            continue
        peak = max(curve); valley = min(curve); avg = sum(curve) / len(curve); total_val = sum(curve)
        batch.append((date_str, "system", curve, peak, valley, avg, total_val))

    pg = pg_conn()
    c = pg.cursor()
    psycopg2.extras.execute_values(c,
        """INSERT INTO total_load_daily (data_date, region, curve_96, peak_load, valley_load, avg_load, total_mwh)
           VALUES %s
           ON CONFLICT (data_date) DO UPDATE SET
             curve_96 = EXCLUDED.curve_96, peak_load = EXCLUDED.peak_load,
             valley_load = EXCLUDED.valley_load, avg_load = EXCLUDED.avg_load,
             total_mwh = EXCLUDED.total_mwh""",
        batch,
        template="(%s::date, %s::text, %s::double precision[], %s::double precision, %s::double precision, %s::double precision, %s::double precision)")
    pg.commit()
    print(f"  PG: {len(batch)} 天写入")
    c.close(); pg.close()
    return {"mysql源": total, "pg写入": len(batch)}


# ═══════════════════════════════════════════════════════════════
# Step 4: 新能源（仅统计）
# ═══════════════════════════════════════════════════════════════
def step_new_energy() -> dict:
    print("\n=== 4 新能源出力（统计信息，暂不迁移完整数据）===")
    conn, cur = mysql_cur()
    cur.execute("SELECT busiType, COUNT(*) as cnt FROM t_bus_new_energy_output_forecast_date GROUP BY busiType")
    total = 0
    for r in cur.fetchall():
        print(f"  类型={r['busiType']}, 条数={r['cnt']}")
        total += r["cnt"]
    conn.close()
    print("  （注：PG spot_market_daily 表结构不同，新能源数据后续按需映射）")
    return {"mysql源": total, "pg写入": 0}


# ═══════════════════════════════════════════════════════════════
# Step 5: 气象+节假日
# ═══════════════════════════════════════════════════════════════
def step_weather() -> dict:
    print("\n=== 5 气象数据 ===")
    fc_conn, fc_cur = mysql_cur("forecast_data")
    pg = pg_conn()
    c = pg.cursor()

    result = {}

    # ── 5a 站点 → weather_locations ──
    fc_cur.execute("SELECT * FROM sites")
    sites = fc_cur.fetchall()
    print(f"  站点: {len(sites)} 个")

    site_batch = []
    for s in sites:
        site_batch.append((
            s["site_id"],
            s.get("name", s["site_id"]),
            float(s["lat"]) if s.get("lat") else None,
            float(s["lon"]) if s.get("lon") else None,
        ))

    if not DRY_RUN and site_batch:
        psycopg2.extras.execute_values(c,
            "INSERT INTO weather_locations (name, city, latitude, longitude) VALUES %s ON CONFLICT DO NOTHING",
            site_batch, template="(%s::varchar, %s::varchar, %s::numeric, %s::numeric)")
        pg.commit()
        print(f"  weather_locations: {len(site_batch)} 条")

    result["weather_locations"] = {"mysql源": len(sites), "pg写入": 0 if DRY_RUN else len(site_batch)}

    # ── 5b weather_actuals (日聚合 weather_hourly) ──
    fc_cur.execute("SELECT COUNT(*) as cnt FROM weather_hourly")
    wh_total = fc_cur.fetchone()["cnt"]
    print(f"  weather_hourly 原始: {wh_total} 条")

    wa_written = 0
    if not DRY_RUN:
        batch_all = []
        fc_cur.execute("SELECT DISTINCT site_id FROM weather_hourly ORDER BY site_id")
        site_ids = [r["site_id"] for r in fc_cur.fetchall()]

        for sid in site_ids:
            fc_cur.execute("""
                SELECT site_id, DATE(ts) as obs_date,
                       AVG(temp_2m) as avg_temp,
                       MAX(temp_2m) as max_temp,
                       MIN(temp_2m) as min_temp,
                       AVG(rh_2m) as humidity,
                       AVG(wind_speed_10m) as wind_speed,
                       SUM(precip_mm) as precipitation
                FROM weather_hourly
                WHERE site_id = %s
                GROUP BY site_id, DATE(ts)
                ORDER BY obs_date
            """, (sid,))
            rows = fc_cur.fetchall()
            for r in rows:
                d = r["obs_date"]
                if isinstance(d, datetime):
                    d = d.strftime("%Y-%m-%d")
                batch_all.append((
                    r["site_id"], d,
                    float(r["avg_temp"]) if r["avg_temp"] else None,
                    float(r["max_temp"]) if r["max_temp"] else None,
                    float(r["min_temp"]) if r["min_temp"] else None,
                    float(r["humidity"]) if r["humidity"] else None,
                    float(r["wind_speed"]) if r["wind_speed"] else None,
                    float(r["precipitation"]) if r["precipitation"] else None,
                ))

        if batch_all:
            psycopg2.extras.execute_values(c,
                "INSERT INTO weather_actuals (location_name, date, avg_temp, max_temp, min_temp, humidity, wind_speed, precipitation) VALUES %s ON CONFLICT DO NOTHING",
                batch_all, template="(%s::varchar, %s::date, %s::numeric, %s::numeric, %s::numeric, %s::numeric, %s::numeric, %s::numeric)", page_size=500)
            pg.commit()
            print(f"  weather_actuals: {len(batch_all)} 条")
            wa_written = len(batch_all)

    result["weather_actuals"] = {"mysql源": wh_total, "pg写入": wa_written}

    # ── 5c weather_forecasts ──
    fc_cur.execute("SELECT COUNT(*) as cnt FROM weather_forecast")
    wf_total = fc_cur.fetchone()["cnt"]
    print(f"  weather_forecast 原始: {wf_total} 条")

    wf_written = 0
    if not DRY_RUN:
        fc_cur.execute("SELECT DISTINCT site_id FROM weather_forecast ORDER BY site_id")
        fc_site_ids = [r["site_id"] for r in fc_cur.fetchall()]

        fc_batch_all = []
        for sid in fc_site_ids:
            fc_cur.execute("""
                SELECT site_id, DATE(forecast_run_ts) as fc_date, DATE(target_ts) as tgt_date,
                       AVG(temp_2m) as temp_fc, AVG(rh_2m) as hum_fc, AVG(wind_speed_10m) as wind_fc
                FROM weather_forecast
                WHERE site_id = %s
                GROUP BY site_id, DATE(forecast_run_ts), DATE(target_ts)
                ORDER BY fc_date, tgt_date
            """, (sid,))
            fc_rows = fc_cur.fetchall()
            for r in fc_rows:
                fd = r["fc_date"]; td = r["tgt_date"]
                if isinstance(fd, datetime):
                    fd = fd.strftime("%Y-%m-%d")
                if isinstance(td, datetime):
                    td = td.strftime("%Y-%m-%d")
                fc_batch_all.append((r["site_id"], fd, td,
                    float(r["temp_fc"]) if r["temp_fc"] else None,
                    float(r["hum_fc"]) if r["hum_fc"] else None,
                    float(r["wind_fc"]) if r["wind_fc"] else None))

        if fc_batch_all:
            psycopg2.extras.execute_values(c,
                "INSERT INTO weather_forecasts (location_name, forecast_date, target_date, temp_forecast, humidity_forecast, wind_forecast) VALUES %s ON CONFLICT DO NOTHING",
                fc_batch_all, template="(%s::varchar, %s::date, %s::date, %s::numeric, %s::numeric, %s::numeric)", page_size=500)
            pg.commit()
            print(f"  weather_forecasts: {len(fc_batch_all)} 条")
            wf_written = len(fc_batch_all)

    result["weather_forecasts"] = {"mysql源": wf_total, "pg写入": wf_written}

    # ── 5d holidays ──
    fc_cur.execute("SELECT COUNT(*) as cnt FROM holidays_cn")
    hol_total = fc_cur.fetchone()["cnt"]
    fc_cur.execute("SELECT * FROM holidays_cn ORDER BY date")
    hols = fc_cur.fetchall()

    hol_written = 0
    if not DRY_RUN:
        hol_batch = []
        for h in hols:
            d = h["date"]
            if isinstance(d, datetime):
                d = d.strftime("%Y-%m-%d")
            kind = "public" if h["is_off_day"] == 1 else "observance"
            hol_batch.append((d, h["name"], kind))

        if hol_batch:
            psycopg2.extras.execute_values(c,
                "INSERT INTO holidays (holiday_date, name, kind) VALUES %s ON CONFLICT DO NOTHING",
                hol_batch, template="(%s::date, %s::text, %s::text)")
            pg.commit()
            print(f"  holidays: {len(hol_batch)} 条")
            hol_written = len(hol_batch)

    result["holidays"] = {"mysql源": hol_total, "pg写入": hol_written}

    c.close(); pg.close()
    fc_conn.close()

    # 合并子步骤汇总
    merged = {}
    for sub_name, sub_stat in result.items():
        for k, v in sub_stat.items():
            merged[f"{sub_name}.{k}"] = v
    return merged


# ═══════════════════════════════════════════════════════════════
# Step 6: 企业信息 → customers
# B3: 提取统一社会信用代码，ON CONFLICT (user_name, source) DO UPDATE
# ═══════════════════════════════════════════════════════════════
def step_companies() -> dict:
    print("\n=== 6 福建企业 → customers ===")
    conn, cur = mysql_cur("fujian_biz")
    total = mysql_count(cur, "companies")
    print(f"  总计: {total} 条")

    if DRY_RUN:
        conn.close()
        return {"mysql源": total, "pg写入": 0}

    pg = pg_conn()
    c = pg.cursor()
    written = 0

    for batch_rows in mysql_fetch_batches(cur, "companies", order_by="ORDER BY id"):
        batch = []
        for r in batch_rows:
            name = r.get("企业名称") or ""
            loc = r.get("所属城市") or ""
            extra = {}
            credit_code = None
            for f in ["法定代表人", "注册资本", "实缴资本", "成立日期", "统一社会信用代码",
                       "企业地址", "电话", "邮箱", "参保人数", "所属省份", "所属区县",
                       "营业期限", "行业门类", "登记状态", "企业规模", "评分", "信用等级", "是否小微企业"]:
                if r.get(f):
                    extra[f] = r[f]
                    # B3: 提取统一社会信用代码到单独字段
                    if f == "统一社会信用代码":
                        credit_code = str(r[f]).strip()
            batch.append((name, loc, credit_code, json.dumps(extra, ensure_ascii=False)))

        if batch:
            # B3: 用 ON CONFLICT (user_name, source) DO UPDATE 实现幂等
            psycopg2.extras.execute_values(c,
                """INSERT INTO customers (user_name, location, source, credit_code, extra)
                   VALUES %s
                   ON CONFLICT (user_name, source) DO UPDATE SET
                     location = EXCLUDED.location,
                     credit_code = COALESCE(EXCLUDED.credit_code, customers.credit_code),
                     extra = EXCLUDED.extra""",
                batch,
                template="(%s::varchar, %s::varchar, 'fujian_biz'::varchar, %s::varchar, %s::jsonb)",
                page_size=500)
            pg.commit()
            written += len(batch)

        print(f"  进度: {min(written, total)}/{total} ({written} 写入)")

    conn.close(); c.close(); pg.close()
    print(f"  customers: {written} 条")
    return {"mysql源": total, "pg写入": written}


# ═══════════════════════════════════════════════════════════════
# 主入口
# ═══════════════════════════════════════════════════════════════
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="OpenPTS V1 MySQL → PostgreSQL 数据迁移")
    parser.add_argument("--dry-run", action="store_true", help="只统计不写入")
    parser.add_argument("--only", type=str, default=None, help="只跑指定步骤（步骤名，如 '日前价格'）")
    parser.add_argument("--strict", action="store_true", help="任何关键步骤失败则非零退出")
    args = parser.parse_args()

    DRY_RUN = args.dry_run
    STRICT = args.strict

    steps = [
        ("日前价格", step_price_da),
        ("实时价格", step_price_rt),
        ("负荷数据", step_load),
        ("新能源", step_new_energy),
        ("气象数据", step_weather),
        ("企业信息", step_companies),
    ]

    if args.only:
        steps = [(n, fn) for n, fn in steps if n == args.only]
        if not steps:
            print(f"❌ 未知步骤名: {args.only}", file=sys.stderr)
            print(f"   可选步骤: {', '.join(n for n, _ in [('日前价格', None), ('实时价格', None), ('负荷数据', None), ('新能源', None), ('气象数据', None), ('企业信息', None)])}", file=sys.stderr)
            sys.exit(1)

    print("=" * 60)
    print(f"OpenPTS V1 数据迁移开始: {datetime.now()}")
    if DRY_RUN:
        print("  *** DRY-RUN 模式: 只统计不写入 ***")
    if args.only:
        print(f"  *** 只跑步骤: {args.only} ***")
    if STRICT:
        print("  *** STRICT 模式: 步骤失败则退出 ***")
    print("=" * 60)

    # B4: 对账汇总
    reconciliation = {}
    has_error = False

    for name, fn in steps:
        try:
            stat = fn()
            reconciliation[name] = stat
        except Exception as e:
            print(f"  ❌ {name} 失败: {e}", file=sys.stderr)
            traceback.print_exc()
            reconciliation[name] = {"error": str(e)}
            if STRICT:
                has_error = True

    # B4: 打印对账汇总表
    print("\n" + "=" * 60)
    print("📊 对账汇总")
    print("=" * 60)
    for step_name, stat in reconciliation.items():
        if "error" in stat:
            print(f"  {step_name}: ❌ 错误 — {stat['error']}")
        else:
            parts = []
            for k, v in stat.items():
                parts.append(f"{k}={v}")
            print(f"  {step_name}: {' | '.join(parts)}")
    print("=" * 60)
    print(f"迁移完成! {datetime.now()}")
    print("=" * 60)

    if has_error and STRICT:
        sys.exit(1)
