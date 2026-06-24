// k6 阶梯压测：模拟真实业务流量混合（仪表盘 60% + 列表查询 30% + 趋势 10%）。
// 用法：k6 run load.js --env BASE=http://localhost:8080
//
// 阶段：30s 升到 20 VU → 持续 1m → 30s 升到 50 VU → 持续 2m → 30s 降到 0。
// 阈值：P95 < 800ms，错误率 < 2%。

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Counter } from 'k6/metrics';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m',  target: 20 },
    { duration: '30s', target: 50 },
    { duration: '2m',  target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<800', 'p(99)<1500'],
    'http_req_duration{name:dashboard_summary}': ['p(95)<400'],
    'http_req_duration{name:settlement_series}': ['p(95)<600'],
  },
};

const BASE = __ENV.BASE || 'http://localhost:8080';
const USERNAME = __ENV.USERNAME || 'admin';
// 密码由种子程序随机生成，运行 k6 前用 PASSWORD 环境变量传入
if (!__ENV.PASSWORD) { console.error('请用 PASSWORD 环境变量传入 admin 密码（种子程序启动时随机生成）'); }
const PASSWORD = __ENV.PASSWORD || '';

const dashboardTrend = new Trend('dashboard_p95', true);
const businessErrors = new Counter('business_errors');

export function setup() {
  const res = http.post(`${BASE}/api/v1/auth/login`,
    JSON.stringify({ username: USERNAME, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } });
  if (res.status !== 200) {
    throw new Error(`登录失败: ${res.status} ${res.body}`);
  }
  return { token: res.json('token') };
}

function authHeaders(token) {
  return { headers: { Authorization: `Bearer ${token}` } };
}

export default function (data) {
  const headers = authHeaders(data.token);

  // 业务流量混合：随机 60/30/10 选场景
  const dice = Math.random();

  if (dice < 0.6) {
    group('dashboard', () => {
      const r1 = http.get(`${BASE}/api/v1/dashboard/summary`,
        Object.assign({}, headers, { tags: { name: 'dashboard_summary' } }));
      dashboardTrend.add(r1.timings.duration);
      if (!check(r1, { 'summary 200': (r) => r.status === 200 })) businessErrors.add(1);

      const r2 = http.get(`${BASE}/api/v1/dashboard/series/settlement?days=14`,
        Object.assign({}, headers, { tags: { name: 'settlement_series' } }));
      if (!check(r2, { 'settle 200': (r) => r.status === 200 })) businessErrors.add(1);
    });
  } else if (dice < 0.9) {
    group('lists', () => {
      const r1 = http.get(`${BASE}/api/v1/customers?limit=50`,
        Object.assign({}, headers, { tags: { name: 'customers_list' } }));
      if (!check(r1, { 'customers 200': (r) => r.status === 200 })) businessErrors.add(1);

      const r2 = http.get(`${BASE}/api/v1/retail/contracts`,
        Object.assign({}, headers, { tags: { name: 'contracts_list' } }));
      if (!check(r2, { 'contracts 200': (r) => r.status === 200 })) businessErrors.add(1);
    });
  } else {
    group('analytics', () => {
      const r1 = http.get(`${BASE}/api/v1/analytics/alerts?limit=50`,
        Object.assign({}, headers, { tags: { name: 'alerts_list' } }));
      if (!check(r1, { 'alerts 200': (r) => r.status === 200 })) businessErrors.add(1);

      const r2 = http.get(`${BASE}/api/v1/analytics/customer-load/summary?days=14`,
        Object.assign({}, headers, { tags: { name: 'cust_load_summary' } }));
      if (!check(r2, { 'cust load 200': (r) => r.status === 200 })) businessErrors.add(1);
    });
  }

  sleep(Math.random() * 2 + 0.5);
}
