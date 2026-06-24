#!/usr/bin/env python3
"""
天气数据采集脚本（Open-Meteo 免费API）— 直写 PostgreSQL
采集：逐小时气象 + 逐日水文
存储：PostgreSQL md_weather_hourly + md_weather_hydrology_daily

监测点坐标在下方 REGIONS 字典定义（默认仅含示例区域 example）。
二次开发时按你的业务区域补充监测点坐标即可。

用法：
  python fetch_weather_data.py              # 采集全部区域
  python fetch_weather_data.py example      # 仅采集指定区域

环境变量（可选覆盖默认值）:
  PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DB
"""
import os, sys, json, time
from datetime import datetime, date, timedelta
import psycopg2
from psycopg2.extras import execute_values
import urllib.request

PG = {
    "host": os.environ.get("PG_HOST", "localhost"),
    "port": int(os.environ.get("PG_PORT", "5432")),
    "user": os.environ.get("PG_USER", "ptis"),
    "password": os.environ.get("PG_PASSWORD", os.environ.get("POSTGRES_PASSWORD", "ptis")),
    "dbname": os.environ.get("PG_DB", "ptis"),
}


# ── 区域 → 监测点坐标 ────────────────────────────────────
# 二次开发：按你自己的业务区域补充监测点（区/县）坐标。
# 坐标可从 Open-Meteo 或公开地理数据获取。每个监测点需 code/name/city/lat/lon。
# key 为区域名（命令行参数），value 为该区域下的监测点列表。
REGIONS = {
    "example": [
        {'code': 'EX_D1', 'name': '示例区', 'city': '示例市', 'lat': 31.23, 'lon': 121.47},
        {'code': 'EX_D2', 'name': '示例县', 'city': '示例市', 'lat': 31.10, 'lon': 121.38},
    ],
}


def get_conn():
    return psycopg2.connect(**PG)


def safe_float(v):
    if v is None:
        return None
    try:
        return float(v)
    except (ValueError, TypeError):
        return None


def fetch_open_meteo(lat, lon, days=3):
    """调用 Open-Meteo API 获取逐小时+逐日数据"""
    end_date = date.today()
    start_date = end_date - timedelta(days=days)

    url = (
        f"https://api.open-meteo.com/v1/forecast"
        f"?latitude={lat}&longitude={lon}"
        f"&hourly=temperature_2m,apparent_temperature,relative_humidity_2m,dew_point_2m,"
        f"precipitation,rain,snowfall,snow_depth,wind_speed_10m,wind_speed_100m,"
        f"wind_direction_10m,wind_direction_100m,wind_gusts_10m,pressure_msl,"
        f"surface_pressure,cloud_cover,cloud_cover_low,cloud_cover_mid,cloud_cover_high,"
        f"shortwave_radiation,direct_radiation,diffuse_radiation,uv_index,visibility,"
        f"weather_code,is_day,et0_fao_evapotranspiration,soil_temperature_0_to_7cm,"
        f"soil_moisture_0_to_7cm"
        f"&daily=temperature_2m_max,temperature_2m_min,temperature_2m_mean,"
        f"relative_humidity_2m_mean,precipitation_sum,rain_sum,snowfall_sum,"
        f"et0_fao_evapotranspiration,wind_speed_10m_max"
        f"&start_date={start_date.isoformat()}&end_date={end_date.isoformat()}"
        f"&timezone=Asia%2FShanghai"
    )
    req = urllib.request.Request(url)
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read().decode())


def upsert_hourly(loc, data):
    """写入 md_weather_hourly"""
    hourly = data.get("hourly", {})
    times = hourly.get("time", [])
    if not times:
        return 0

    # 先删除该地点旧的小时数据（避免重复）
    conn = get_conn()
    cur = conn.cursor()
    adcode = loc["code"]
    cur.execute(
        "DELETE FROM md_weather_hourly WHERE adcode = %s AND obs_time >= NOW() - INTERVAL '7 days'",
        (adcode,),
    )
    conn.commit()
    cur.close()
    conn.close()

    rows = []
    for i, t in enumerate(times):
        rows.append({
            "adcode": adcode,
            "district_name": loc["name"],
            "city_name": loc.get("city", ""),
            "province_name": loc.get("province", ""),
            "lat": loc["lat"],
            "lon": loc["lon"],
            "obs_time": t,
            "temperature_2m": safe_float(hourly.get("temperature_2m", [None])[i]) if i < len(hourly.get("temperature_2m", [])) else None,
            "apparent_temperature": safe_float(hourly.get("apparent_temperature", [None])[i]) if i < len(hourly.get("apparent_temperature", [])) else None,
            "relative_humidity_2m": safe_float(hourly.get("relative_humidity_2m", [None])[i]) if i < len(hourly.get("relative_humidity_2m", [])) else None,
            "wind_speed_10m": safe_float(hourly.get("wind_speed_10m", [None])[i]) if i < len(hourly.get("wind_speed_10m", [])) else None,
            "wind_speed_100m": safe_float(hourly.get("wind_speed_100m", [None])[i]) if i < len(hourly.get("wind_speed_100m", [])) else None,
            "wind_direction_10m": int(hourly["wind_direction_10m"][i]) if i < len(hourly.get("wind_direction_10m", [])) and hourly["wind_direction_10m"][i] is not None else None,
            "wind_direction_100m": int(hourly["wind_direction_100m"][i]) if i < len(hourly.get("wind_direction_100m", [])) and hourly["wind_direction_100m"][i] is not None else None,
            "pressure_msl": safe_float(hourly.get("pressure_msl", [None])[i]) if i < len(hourly.get("pressure_msl", [])) else None,
            "precipitation": safe_float(hourly.get("precipitation", [None])[i]) if i < len(hourly.get("precipitation", [])) else None,
        })

    if not rows:
        return 0

    conn = get_conn()
    cur = conn.cursor()
    cols = list(rows[0].keys())
    col_str = ", ".join(cols)
    sql = f"INSERT INTO md_weather_hourly ({col_str}) VALUES %s"
    vals = [tuple(r[c] for c in cols) for r in rows]
    try:
        execute_values(cur, sql, vals)
        conn.commit()
        return len(rows)
    except Exception as e:
        conn.rollback()
        print(f"    [!] hourly upsert 失败: {e}")
        return 0
    finally:
        cur.close()
        conn.close()


def upsert_hydrology(loc, data):
    """写入 md_weather_hydrology_daily"""
    daily = data.get("daily", {})
    dates = daily.get("time", [])
    if not dates:
        return 0

    rows = []
    for i, d in enumerate(dates):
        rows.append({
            "location_code": loc["code"],
            "location_name": loc["name"],
            "lat": loc["lat"],
            "lon": loc["lon"],
            "obs_date": d,
            "temp_mean": safe_float(daily.get("temperature_2m_mean", [None])[i]) if i < len(daily.get("temperature_2m_mean", [])) else None,
            "humidity_mean": safe_float(daily.get("relative_humidity_2m_mean", [None])[i]) if i < len(daily.get("relative_humidity_2m_mean", [])) else None,
            "precipitation_sum": safe_float(daily.get("precipitation_sum", [None])[i]) if i < len(daily.get("precipitation_sum", [])) else None,
            "rain_sum": safe_float(daily.get("rain_sum", [None])[i]) if i < len(daily.get("rain_sum", [])) else None,
            "et0_evapotranspiration": safe_float(daily.get("et0_fao_evapotranspiration", [None])[i]) if i < len(daily.get("et0_fao_evapotranspiration", [])) else None,
            "wind_speed_10m_mean": safe_float(daily.get("wind_speed_10m_max", [None])[i]) if i < len(daily.get("wind_speed_10m_max", [])) else None,
        })

    conn = get_conn()
    cur = conn.cursor()
    # delete existing for this location + date range
    cur.execute(
        "DELETE FROM md_weather_hydrology_daily WHERE location_code = %s AND obs_date >= %s",
        (loc["code"], dates[0]),
    )
    conn.commit()

    cols = list(rows[0].keys())
    col_str = ", ".join(cols)
    sql = f"INSERT INTO md_weather_hydrology_daily ({col_str}) VALUES %s"
    vals = [tuple(r[c] for c in cols) for r in rows]
    try:
        execute_values(cur, sql, vals)
        conn.commit()
        return len(rows)
    except Exception as e:
        conn.rollback()
        print(f"    [!] hydrology upsert 失败: {e}")
        return 0
    finally:
        cur.close()
        conn.close()


def fetch_province(prov_name, locations):
    total_h = 0
    total_d = 0
    for i, loc in enumerate(locations):
        loc["province"] = prov_name
        try:
            data = fetch_open_meteo(loc["lat"], loc["lon"])
            h = upsert_hourly(loc, data)
            d = upsert_hydrology(loc, data)
            total_h += h
            total_d += d
            if (i + 1) % 10 == 0:
                print(f"  [{prov_name}] {i+1}/{len(locations)} 已完成")
        except Exception as e:
            print(f"  [!] {loc['name']} 失败: {e}")
        time.sleep(0.1)  # 限速
    print(f"  [{prov_name}] 完成: {total_h} hourly + {total_d} daily rows")
    return total_h, total_d


def main():
    prov_arg = sys.argv[1].lower() if len(sys.argv) > 1 else "all"
    if prov_arg == "all":
        for prov, locs in PROVINCES.items():
            print(f"\n=== {prov} ({len(locs)} 区县) ===")
            fetch_province(prov, locs)
    elif prov_arg in PROVINCES:
        print(f"\n=== {prov_arg} ({len(PROVINCES[prov_arg])} 区县) ===")
        fetch_province(prov_arg, PROVINCES[prov_arg])
    else:
        print(f"未知省份: {prov_arg}，可选: {', '.join(PROVINCES.keys())}")
        sys.exit(1)
    print("\n[WEATHER] 采集完成")


if __name__ == "__main__":
    main()
