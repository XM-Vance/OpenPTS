# 性能压测

基于 [k6](https://k6.io/) 的脚本集。本地安装：

```bash
brew install k6              # macOS
# 或 docker：docker run --rm -i grafana/k6 run - < smoke.js
```

## 脚本

| 文件 | 场景 | 用法 |
|---|---|---|
| `smoke.js` | 5 个核心接口冒烟（1 VU × 5 iterations） | 部署后 / CI 流水线 |
| `load.js` | 阶梯压测：20→50 VU，4 分钟，混合业务流量 | 性能基线回归 |

## 运行示例

```bash
# 冒烟（默认 localhost:8080）
k6 run smoke.js

# 指定环境
k6 run smoke.js --env BASE=http://api.ptis.example.com --env USERNAME=admin --env PASSWORD=xxx

# 阶梯压测 + 导出 JSON
k6 run --out json=result.json load.js

# 通过 nginx 压测
k6 run load.js --env BASE=http://localhost  # 默认走 nginx :80
```

## 阈值（load.js）

| 指标 | 阈值 |
|---|---|
| 整体错误率 | < 2% |
| 整体 P95 | < 800 ms |
| 整体 P99 | < 1500 ms |
| `dashboard_summary` P95 | < 400 ms |
| `settlement_series` P95 | < 600 ms |

不满足阈值 k6 退出码非 0，可作为 CD 卡口。
