# OpenPTS 数据采集模块 v3

每日自动采集三类市场参考数据，**直写 PostgreSQL** `md_*` 表，无需 MySQL 中间层。

## 脚本清单

| 脚本 | 数据源 | 采集内容 | 写入 PG 表 | 耗时 |
|------|--------|----------|-----------|------|
| `fetch_carbon_data.py` | tanpaifang/eastmoney/web | CCER/CEA 碳价 | `md_carbon_eua` | ~10s |
| `fetch_market_data.py` | AKShare | 宏观/期货/汇率/利率 28类 | `md_macro_*` `md_fuel_*` `md_futures_*` `md_fx_*` `md_index_*` `md_rate_*` `md_bond_*` | ~200s |
| `fetch_weather_data.py` | Open-Meteo | 5省~200区县逐时气象+逐日水文 | `md_weather_hourly` `md_weather_hydrology_daily` | ~180s |

## 数据流向

```
采集脚本                      PostgreSQL (ptis.md_*)
──────                        ──────────────────────
fetch_carbon_data.py   ──→   md_carbon_eua
fetch_market_data.py   ──→   md_macro_* / md_fuel_* / md_futures_* / md_fx_* / md_index_* / md_rate_* / md_bond_*
fetch_weather_data.py  ──→   md_weather_hourly / md_weather_hydrology_daily
```

## 依赖安装

```bash
pip install -r scripts/data-collection/requirements.txt
```

## 环境变量（可选，有默认值）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PG_HOST` | `localhost` | PostgreSQL 地址 |
| `PG_PORT` | `5432` | PostgreSQL 端口 |
| `PG_USER` | `ptis` | 用户名 |
| `PG_PASSWORD` | `ptis` | 密码（也可用 `POSTGRES_PASSWORD`） |
| `PG_DB` | `ptis` | 数据库名 |

## 用法

```bash
# 碳市场（仅交易日采集）
python scripts/data-collection/fetch_carbon_data.py

# 宏观/期货/汇率/利率
python scripts/data-collection/fetch_market_data.py              # 全部
python scripts/data-collection/fetch_market_data.py macro        # 仅宏观经济
python scripts/data-collection/fetch_market_data.py fuel         # 仅燃料
python scripts/data-collection/fetch_market_data.py futures      # 仅金融期货
python scripts/data-collection/fetch_market_data.py fx           # 仅汇率/指数
python scripts/data-collection/fetch_market_data.py rate         # 仅利率

# 天气
python scripts/data-collection/fetch_weather_data.py             # 全部省份
python scripts/data-collection/fetch_weather_data.py fujian      # 仅福建
python scripts/data-collection/fetch_weather_data.py anhui       # 仅安徽
```

## 定时调度

推荐使用 Windows 任务计划程序或 Linux cron 每日运行：

```bash
# 示例：每日 7:00 采集
0 7 * * * cd /path/to/project && python scripts/data-collection/fetch_carbon_data.py && python scripts/data-collection/fetch_market_data.py all && python scripts/data-collection/fetch_weather_data.py
```
