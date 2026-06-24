#!/usr/bin/env python3
"""
AKShare 市场数据采集脚本 v3 — 直写 PostgreSQL
用法:
  python fetch_market_data.py              # 采集所有
  python fetch_market_data.py macro        # 仅宏观经济
  python fetch_market_data.py fuel         # 仅燃料
  python fetch_market_data.py futures      # 仅金融期货
  python fetch_market_data.py fx           # 仅汇率/指数
  python fetch_market_data.py rate         # 仅利率

数据源: AKShare (免费)
存储: PostgreSQL ptis.md_* 表

环境变量（可选覆盖默认值）:
  PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DB
"""
import os, sys, traceback
from datetime import datetime, timedelta
import akshare as ak
import psycopg2
from psycopg2.extras import execute_values

PG = {
    "host": os.environ.get("PG_HOST", "localhost"),
    "port": int(os.environ.get("PG_PORT", "5432")),
    "user": os.environ.get("PG_USER", "ptis"),
    "password": os.environ.get("PG_PASSWORD", os.environ.get("POSTGRES_PASSWORD", "ptis")),
    "dbname": os.environ.get("PG_DB", "ptis"),
}

INCREMENTAL_START = (datetime.now() - timedelta(days=60)).strftime("%Y-%m-%d")


def get_conn():
    return psycopg2.connect(**PG)


def upsert_rows(table, rows, conflict_col="trade_date"):
    """通用 upsert：rows 是 list[dict]，key 是列名"""
    if not rows:
        return 0
    conn = get_conn()
    cur = conn.cursor()
    cols = list(rows[0].keys())
    col_str = ", ".join(cols)
    update_cols = [c for c in cols if c != conflict_col]
    update_str = ", ".join(f"{c}=EXCLUDED.{c}" for c in update_cols)
    sql = (
        f"INSERT INTO {table} ({col_str}) VALUES %s "
        f"ON CONFLICT ({conflict_col}) DO UPDATE SET {update_str}"
    )
    vals = [tuple(r.get(c) for c in cols) for r in rows]
    try:
        execute_values(cur, sql, vals)
        conn.commit()
        return cur.rowcount
    except Exception as e:
        conn.rollback()
        print(f"  [!] upsert {table} 失败: {e}")
        return 0
    finally:
        cur.close()
        conn.close()


def safe_float(v):
    if v is None or str(v).strip() in ("", "-", "None", "nan"):
        return None
    try:
        return float(v)
    except (ValueError, TypeError):
        return None


import re as _re

def parse_cn_date(s):
    """中文日期 → YYYY-MM-DD（PG date 类型兼容）"""
    if not s:
        return None
    s = str(s).strip()
    # 2026年第1季度 → 2026-03-31
    m = _re.match(r'(\d{4})年第(\d)季度', s)
    if m:
        q = int(m.group(2))
        month = q * 3
        day = 30 if month in (6, 9) else 31
        return f"{m.group(1)}-{month:02d}-{day}"
    # 2026年05月份 → 2026-05-01
    m = _re.match(r'(\d{4})年(\d{1,2})月份?', s)
    if m:
        return f"{m.group(1)}-{int(m.group(2)):02d}-01"
    # 2026.05 → 2026-05-01
    m = _re.match(r'(\d{4})\.(\d{1,2})$', s)
    if m:
        return f"{m.group(1)}-{int(m.group(2)):02d}-01"
    # 已经是标准格式，截取前10位
    return s[:10]


# ── 宏观经济 ────────────────────────────────────────

def fetch_gdp():
    print("[GDP] 采集中...")
    df = ak.macro_china_gdp()
    rows = []
    for _, r in df.iterrows():
        d = str(r.get("季度", "")).strip()
        sd = parse_cn_date(d)
        if not sd or not _re.match(r'\d{4}-\d{2}-\d{2}', sd):
            continue  # 跳过无法解析的（如"2025年第1-4季"汇总行）
        rows.append({"stat_date": sd, "gdp_yoy": safe_float(r.get("GDP同比增长"))})
    n = upsert_rows("md_macro_gdp", rows, "stat_date")
    print(f"  ✓ {n} rows → md_macro_gdp")


def fetch_cpi():
    print("[CPI] 采集中...")
    df = ak.macro_china_cpi()
    rows = []
    for _, r in df.iterrows():
        d = str(r.get("月份", "")).strip()
        rows.append({
            "stat_date": parse_cn_date(d),
            "cpi_yoy": safe_float(r.get("同比增长")),
            "cpi_mom": safe_float(r.get("环比增长")),
        })
    n = upsert_rows("md_macro_cpi", rows, "stat_date")
    print(f"  ✓ {n} rows → md_macro_cpi")


def fetch_ppi():
    print("[PPI] 采集中...")
    df = ak.macro_china_ppi()
    rows = []
    for _, r in df.iterrows():
        d = str(r.get("月份", "")).strip()
        rows.append({
            "stat_date": parse_cn_date(d),
            "ppi_yoy": safe_float(r.get("同比增长")),
            "ppi_mom": safe_float(r.get("环比增长")),
        })
    n = upsert_rows("md_macro_ppi", rows, "stat_date")
    print(f"  ✓ {n} rows → md_macro_ppi")


def fetch_pmi():
    print("[PMI] 采集中...")
    df = ak.macro_china_pmi()
    rows = []
    for _, r in df.iterrows():
        d = str(r.get("月份", "")).strip()
        rows.append({"stat_date": parse_cn_date(d), "pmi_value": safe_float(r.get("制造业-LF"))})
    n = upsert_rows("md_macro_pmi", rows, "stat_date")
    print(f"  ✓ {n} rows → md_macro_pmi")


def fetch_m2():
    print("[M2] 采集中...")
    df = ak.macro_china_money_supply()
    rows = []
    for _, r in df.iterrows():
        d = str(r.get("月份", "")).strip()
        rows.append({
            "stat_date": parse_cn_date(d),
            "m2_yoy": safe_float(r.get("M2-同比增长")),
            "m2_balance": safe_float(r.get("M2-数量")),
        })
    n = upsert_rows("md_macro_m2", rows, "stat_date")
    print(f"  ✓ {n} rows → md_macro_m2")



# ── 燃料 + 期货（通用 OHLC 采集） ────────────────────

def fetch_futures_ohlc(pg_table, ak_symbol, name):
    """通用期货历史数据采集（新浪接口）"""
    print(f"[{name}] symbol={ak_symbol} 采集中...")
    try:
        df = ak.futures_zh_daily_sina(symbol=ak_symbol)
        if df is None or len(df) == 0:
            print(f"  [!] {name} 无数据")
            return
        df = df[df["date"] >= INCREMENTAL_START]
        rows = []
        for _, r in df.iterrows():
            d = str(r.get("date", "")).strip()
            rows.append({
                "trade_date": d[:10],
                "open_price": safe_float(r.get("open")),
                "high_price": safe_float(r.get("high")),
                "low_price": safe_float(r.get("low")),
                "close_price": safe_float(r.get("close")),
                "volume": safe_float(r.get("volume")),
            })
        n = upsert_rows(pg_table, rows)
        print(f"  ✓ {n} rows → {pg_table}")
    except Exception as e:
        print(f"  [!] {name} 失败: {e}")


def fetch_cn_oil_price():
    print("[CN_OIL] 国内油价采集中...")
    try:
        df = ak.energy_oil_hist()
        rows = []
        for _, r in df.iterrows():
            d = str(r.iloc[0]).strip()
            rows.append({
                "adjust_date": d[:10],
                "gasoline_price": safe_float(r.iloc[1]) if len(r) > 1 else None,
                "diesel_price": safe_float(r.iloc[2]) if len(r) > 2 else None,
            })
        n = upsert_rows("md_fuel_cn_oil_price", rows, "adjust_date")
        print(f"  ✓ {n} rows → md_fuel_cn_oil_price")
    except Exception as e:
        print(f"  [!] CN_OIL 失败: {e}")


# ── 金融期货列表 ────────────────────────────────────

FUTURES = [
    ("md_futures_rb", "RB0", "螺纹钢"),
    ("md_futures_hc", "HC0", "热轧卷板"),
    ("md_futures_i",  "I0",  "铁矿石"),
    ("md_futures_zc", "ZC0", "动力煤"),
    ("md_futures_cu", "CU0", "沪铜"),
    ("md_futures_al", "AL0", "沪铝"),
    ("md_futures_zn", "ZN0", "沪锌"),
    ("md_futures_au", "AU0", "黄金"),
    ("md_futures_fg", "FG0", "玻璃"),
    ("md_futures_sa", "SA0", "纯碱"),
]

def fetch_all_futures():
    for pg_table, symbol, name in FUTURES:
        fetch_futures_ohlc(pg_table, symbol, name)


def fetch_all_fuel():
    # 国际原油/天然气用 futures_foreign_hist（AKShare 外盘接口）
    try:
        print("[WTI原油] 采集中...")
        df = ak.futures_foreign_hist(symbol="CL")
        if df is not None and len(df) > 0:
            df = df.tail(60)
            rows = []
            for _, r in df.iterrows():
                d = str(r.get("日期", r.iloc[0])).strip()
                rows.append({
                    "trade_date": d[:10],
                    "open_price": safe_float(r.get("开盘")),
                    "high_price": safe_float(r.get("最高")),
                    "low_price": safe_float(r.get("最低")),
                    "close_price": safe_float(r.get("收盘")),
                    "volume": safe_float(r.get("成交量")),
                })
            n = upsert_rows("md_fuel_wti", rows)
            print(f"  ✓ {n} rows → md_fuel_wti")
    except Exception as e:
        print(f"  [!] WTI原油 失败: {e}")

    try:
        print("[天然气HH] 采集中...")
        df = ak.futures_foreign_hist(symbol="NG")
        if df is not None and len(df) > 0:
            df = df.tail(60)
            rows = []
            for _, r in df.iterrows():
                d = str(r.get("日期", r.iloc[0])).strip()
                rows.append({
                    "trade_date": d[:10],
                    "open_price": safe_float(r.get("开盘")),
                    "high_price": safe_float(r.get("最高")),
                    "low_price": safe_float(r.get("最低")),
                    "close_price": safe_float(r.get("收盘")),
                    "volume": safe_float(r.get("成交量")),
                })
            n = upsert_rows("md_fuel_natgas_hh", rows)
            print(f"  ✓ {n} rows → md_fuel_natgas_hh")
    except Exception as e:
        print(f"  [!] 天然气HH 失败: {e}")

    fetch_futures_ohlc("md_fuel_ine_crude", "sc0", "INE原油")
    fetch_cn_oil_price()


# ── 汇率/指数 ────────────────────────────────────────

def fetch_fx_usdcny():
    print("[FX_USDCNY] 采集中...")
    try:
        df = ak.currency_boc_sina(symbol="美元")
        rows = []
        for _, r in df.iterrows():
            d = str(r.iloc[0]).strip()
            close = safe_float(r.iloc[1])
            rows.append({
                "trade_date": d[:10],
                "open_price": close,
                "high_price": close,
                "low_price": close,
                "close_price": close,
            })
        n = upsert_rows("md_fx_usdcny", rows)
        print(f"  ✓ {n} rows → md_fx_usdcny")
    except Exception as e:
        print(f"  [!] FX_USDCNY 失败: {e}")


def fetch_dxy():
    print("[DXY] 美元指数采集中...")
    try:
        df = ak.currency_boc_sina(symbol="美元指数")
        rows = []
        for _, r in df.iterrows():
            d = str(r.iloc[0]).strip()
            close = safe_float(r.iloc[1])
            rows.append({
                "trade_date": d[:10],
                "open_price": close,
                "high_price": close,
                "low_price": close,
                "close_price": close,
            })
        n = upsert_rows("md_index_dxy", rows)
        print(f"  ✓ {n} rows → md_index_dxy")
    except Exception as e:
        print(f"  [!] DXY 失败: {e}")


def fetch_bdi():
    print("[BDI] 波罗的海干散货指数采集中...")
    try:
        df = ak.index_global_bdi()
        rows = []
        for _, r in df.iterrows():
            d = str(r.iloc[0]).strip()
            rows.append({"trade_date": d[:10], "bdi_value": safe_float(r.iloc[1])})
        n = upsert_rows("md_index_bdi", rows)
        print(f"  ✓ {n} rows → md_index_bdi")
    except Exception as e:
        print(f"  [!] BDI 失败: {e}")


# ── 利率 ────────────────────────────────────────

def fetch_lpr():
    print("[LPR] 贷款市场报价利率采集中...")
    try:
        df = ak.macro_china_lpr()
        rows = []
        for _, r in df.iterrows():
            d = str(r.get("TRADE_DATE", r.iloc[0])).strip()
            rows.append({
                "stat_date": d[:10],
                "lpr_1y": safe_float(r.get("LPR1Y", r.iloc[1] if len(r) > 1 else None)),
                "lpr_5y": safe_float(r.get("LPR5Y", r.iloc[2] if len(r) > 2 else None)),
            })
        n = upsert_rows("md_rate_lpr", rows, "stat_date")
        print(f"  ✓ {n} rows → md_rate_lpr")
    except Exception as e:
        print(f"  [!] LPR 失败: {e}")


def fetch_shibor():
    print("[SHIBOR] 上海银行间同业拆放利率采集中...")
    try:
        df = ak.macro_china_shibor_all()
        rows = []
        for _, r in df.iterrows():
            d = str(r.get("日期", r.iloc[0])).strip()
            rows.append({
                "trade_date": d[:10],
                "overnight": safe_float(r.get("O/N_HR", r.iloc[1] if len(r) > 1 else None)),
                "week_1": safe_float(r.get("1W", r.iloc[2] if len(r) > 2 else None)),
                "month_1": safe_float(r.get("1M", r.iloc[3] if len(r) > 3 else None)),
                "month_3": safe_float(r.get("3M", r.iloc[4] if len(r) > 4 else None)),
                "month_6": safe_float(r.get("6M", r.iloc[5] if len(r) > 5 else None)),
                "year_1": safe_float(r.get("1Y", r.iloc[6] if len(r) > 6 else None)),
            })
        n = upsert_rows("md_rate_shibor", rows)
        print(f"  ✓ {n} rows → md_rate_shibor")
    except Exception as e:
        print(f"  [!] SHIBOR 失败: {e}")


def fetch_bond_yield():
    print("[BOND] 中美国债收益率采集中...")
    try:
        df = ak.bond_zh_us_rate()
        rows = []
        for _, r in df.iterrows():
            d = str(r.get("日期", r.iloc[0])).strip()
            rows.append({
                "trade_date": d[:10],
                "cn_2y": safe_float(r.get("中国国债收益率2年", r.get("CN_2Y"))),
                "cn_5y": safe_float(r.get("中国国债收益率5年", r.get("CN_5Y"))),
                "cn_10y": safe_float(r.get("中国国债收益率10年", r.get("CN_10Y"))),
                "cn_30y": safe_float(r.get("中国国债收益率30年", r.get("CN_30Y"))),
                "us_2y": safe_float(r.get("美国国债收益率2年", r.get("US_2Y"))),
                "us_5y": safe_float(r.get("美国国债收益率5年", r.get("US_5Y"))),
                "us_10y": safe_float(r.get("美国国债收益率10年", r.get("US_10Y"))),
                "us_30y": safe_float(r.get("美国国债收益率30年", r.get("US_30Y"))),
            })
        n = upsert_rows("md_bond_zh_us_yield", rows)
        print(f"  ✓ {n} rows → md_bond_zh_us_yield")
    except Exception as e:
        print(f"  [!] BOND 失败: {e}")


# ── 主流程 ────────────────────────────────────────

TASKS = {
    "macro": [fetch_gdp, fetch_cpi, fetch_ppi, fetch_pmi, fetch_m2],
    "fuel": [fetch_all_fuel],
    "futures": [fetch_all_futures],
    "fx": [fetch_fx_usdcny, fetch_dxy],
    "rate": [fetch_lpr, fetch_shibor, fetch_bond_yield],
}


def main():
    cat = sys.argv[1] if len(sys.argv) > 1 else "all"
    if cat == "all":
        for c, fns in TASKS.items():
            for fn in fns:
                try:
                    fn()
                except Exception as e:
                    print(f"  [!] {fn.__name__} 异常: {e}")
    else:
        for fn in TASKS.get(cat, []):
            try:
                fn()
            except Exception as e:
                print(f"  [!] {fn.__name__} 异常: {e}")
    print("\n[MARKET] 采集完成")


if __name__ == "__main__":
    main()
