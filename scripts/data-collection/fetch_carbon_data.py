#!/usr/bin/env python3
"""
碳市场每日数据采集脚本（多策略 fallback）— 直写 PostgreSQL
- CCER: ccer.com.cn（curl → web搜索 → tanpaifang）
- CEA:  tanpaifang → 东方财富 → web搜索 → 碳中和网
- 存入 PostgreSQL ptis.md_carbon_eua

环境变量（可选覆盖默认值）：
  PG_HOST (default: localhost)
  PG_PORT (default: 5432)
  PG_USER (default: ptis)
  PG_PASSWORD (default: ptis)
  PG_DB (default: ptis)
"""
import os, re, sys, subprocess, json
from datetime import date, datetime
import psycopg2
from psycopg2.extras import execute_values

# ── PG 连接 ──
PG = {
    "host": os.environ.get("PG_HOST", "localhost"),
    "port": int(os.environ.get("PG_PORT", "5432")),
    "user": os.environ.get("PG_USER", "ptis"),
    "password": os.environ.get("PG_PASSWORD", os.environ.get("POSTGRES_PASSWORD", "ptis")),
    "dbname": os.environ.get("PG_DB", "ptis"),
}


def get_conn():
    return psycopg2.connect(**PG)


def upsert_carbon(rows):
    """批量 upsert 到 md_carbon_eua"""
    if not rows:
        return 0
    conn = get_conn()
    cur = conn.cursor()
    cols = ["trade_date", "open_price", "high_price", "low_price", "close_price", "volume"]
    sql = f"""
        INSERT INTO md_carbon_eua ({', '.join(cols)})
        VALUES %s
        ON CONFLICT (trade_date) DO UPDATE SET
            open_price  = EXCLUDED.open_price,
            high_price  = EXCLUDED.high_price,
            low_price   = EXCLUDED.low_price,
            close_price = EXCLUDED.close_price,
            volume      = EXCLUDED.volume
    """
    vals = [(r.get("trade_date"), r.get("open_price"), r.get("high_price"),
             r.get("low_price"), r.get("close_price"), r.get("volume")) for r in rows]
    try:
        execute_values(cur, sql, vals)
        conn.commit()
        return cur.rowcount
    finally:
        cur.close()
        conn.close()


def curl(url, timeout=15):
    r = subprocess.run(
        ["curl", "-sL", "--max-time", str(timeout),
         "-H", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
         "-H", "Accept: text/html,application/xhtml+xml",
         url],
        capture_output=True, text=True, timeout=timeout+5
    )
    return r.stdout


def today_or_arg():
    if len(sys.argv) >= 3 and sys.argv[1] == "--date":
        return date.fromisoformat(sys.argv[2])
    return date.today()


def is_trading_day(d):
    if d.weekday() >= 5:
        return False
    holidays_2026 = [
        date(2026,1,1),
        date(2026,1,26), date(2026,1,27), date(2026,1,28),
        date(2026,1,29), date(2026,1,30), date(2026,1,31), date(2026,2,1),
        date(2026,4,4), date(2026,4,5), date(2026,4,6),
        date(2026,5,1), date(2026,5,2), date(2026,5,3), date(2026,5,4), date(2026,5,5),
        date(2026,5,31), date(2026,6,1), date(2026,6,2),
        date(2026,9,25), date(2026,9,26), date(2026,9,27),
        date(2026,10,1), date(2026,10,2), date(2026,10,3),
        date(2026,10,4), date(2026,10,5), date(2026,10,6), date(2026,10,7),
    ]
    return d not in holidays_2026


# ── 辅助：web 搜索策略 ──────────────────────────────────

def web_search_ccer(target_date):
    """策略3: 通过 web 搜索 CCER 行情"""
    d = target_date
    date_cn = f"{d.year}年{d.month}月{d.day}日"
    query = f'"{date_cn}" "核证自愿减排量" "成交量" site:ccer.com.cn OR site:tanpaifang.com'
    url = f"https://www.bing.com/search?q={subprocess.list2cmdline([query])}"
    html = curl(url, timeout=10)
    if not html:
        return None
    text = re.sub(r'<[^>]+>', ' ', html)
    text = re.sub(r'\s+', ' ', text)
    vol_m = re.search(r'成交量\s*([\d,]+)\s*吨', text)
    amt_m = re.search(r'成交额\s*([\d,.]+)\s*元', text)
    pri_m = re.search(r'成交均价\s*([\d,.]+)\s*元/吨', text)
    if vol_m and pri_m:
        price = float(pri_m.group(1).replace(",", ""))
        return {
            "trade_date": d.isoformat(),
            "open_price": price,
            "high_price": price,
            "low_price": price,
            "close_price": price,
            "volume": int(vol_m.group(1).replace(",", "")),
        }
    return None


def web_search_cea(target_date):
    """策略4: 通过 web 搜索 CEA 行情"""
    d = target_date
    date_short = f"{d.year}{d.month:02d}{d.day:02d}"
    date_cn = f"{d.year}年{d.month}月{d.day}日"

    query = f'全国碳市场 {date_cn} 收盘价 成交量 site:tanpaifang.com'
    url = f"https://www.bing.com/search?q={subprocess.list2cmdline([query])}"
    html = curl(url, timeout=10)
    if not html:
        return None

    text = re.sub(r'<[^>]+>', ' ', html)
    text = re.sub(r'\s+', ' ', text)
    pri_m = re.search(r'收盘价[为:]?\s*([\d.]+)\s*元', text)
    if not pri_m:
        pri_m = re.search(r'碳价[为:]?\s*([\d.]+)\s*元', text)
    vol_m = re.search(r'成交量\s*([\d,]+)\s*吨', text)

    if pri_m:
        price = float(pri_m.group(1))
        vol = int(vol_m.group(1).replace(",", "")) if vol_m else 0
        return {
            "trade_date": d.isoformat(),
            "open_price": price,
            "high_price": price,
            "low_price": price,
            "close_price": price,
            "volume": vol,
        }
    return None


# ── 策略 1-2: tanpaifang ──────────────────────────────────

def fetch_tanpaifang_api(target_date):
    """策略1: 碳排网 tanpaifang API"""
    d = target_date
    api_url = "https://www.tanpaifang.com/api/carbonmarket/daykline"
    params = f"?market=all&date={d.isoformat()}"
    html = curl(api_url + params, timeout=10)
    if not html:
        return None
    try:
        data = json.loads(html)
        if isinstance(data, dict) and data.get("data"):
            item = data["data"][0] if isinstance(data["data"], list) else data["data"]
            price = safe_float(item.get("close") or item.get("price"))
            if price and price > 0:
                return {
                    "trade_date": d.isoformat(),
                    "open_price": safe_float(item.get("open")),
                    "high_price": safe_float(item.get("high")),
                    "low_price": safe_float(item.get("low")),
                    "close_price": price,
                    "volume": safe_float(item.get("volume")) or 0,
                }
    except (json.JSONDecodeError, KeyError, IndexError):
        pass
    return None


def fetch_eastmoney_api(target_date):
    """策略2: 东方财富碳市场 API"""
    d = target_date
    # 东方财富碳市场行情
    url = f"https://push2.eastmoney.com/api/qt/stock/get?secid=0.301611&fields=f43,f44,f45,f46,f47,f57,f58"
    html = curl(url, timeout=10)
    if not html:
        return None
    try:
        data = json.loads(html)
        if data.get("data"):
            d2 = data["data"]
            price = safe_float(d2.get("f43"))
            if price and price > 0:
                # f43 是最新价(需除以100)
                price = price / 100
                return {
                    "trade_date": d.isoformat(),
                    "open_price": safe_float(d2.get("f46", 0)) / 100 or price,
                    "high_price": safe_float(d2.get("f44", 0)) / 100 or price,
                    "low_price": safe_float(d2.get("f45", 0)) / 100 or price,
                    "close_price": price,
                    "volume": safe_float(d2.get("f47", 0)),
                }
    except (json.JSONDecodeError, KeyError):
        pass
    return None


def safe_float(v):
    if v is None or str(v).strip() in ("", "-", "None", "nan"):
        return None
    try:
        return float(v)
    except (ValueError, TypeError):
        return None


# ── 主流程 ──────────────────────────────────

def main():
    d = today_or_arg()
    if not is_trading_day(d):
        print(f"[CARBON] {d.isoformat()} 非交易日，跳过")
        return

    strategies = [
        ("tanpaifang", fetch_tanpaifang_api),
        ("eastmoney", fetch_eastmoney_api),
        ("web-search", web_search_cea),
    ]

    for name, fn in strategies:
        try:
            row = fn(d)
            if row:
                print(f"[CARBON] ✓ {name}: 价格={row['close_price']} 成交量={row['volume']}")
                n = upsert_carbon([row])
                print(f"[CARBON] 已写入 PG md_carbon_eua ({n} rows)")
                return
        except Exception as e:
            print(f"[CARBON] {name} 失败: {e}")

    print(f"[CARBON] {d.isoformat()} 所有策略均失败，未写入")


if __name__ == "__main__":
    main()
