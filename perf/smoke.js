// k6 冒烟测试：验证 5 条关键接口在低负载下都能 200。
// 用法：k6 run smoke.js --env BASE=http://localhost:8080

import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  vus: 1,
  iterations: 5,
  thresholds: {
    http_req_failed: ['rate<0.01'],     // 错误率 < 1%
    http_req_duration: ['p(95)<500'],   // P95 < 500ms
  },
};

const BASE = __ENV.BASE || 'http://localhost:8080';
const USERNAME = __ENV.USERNAME || 'admin';
// 密码由种子程序随机生成，运行 k6 前用 PASSWORD 环境变量传入
if (!__ENV.PASSWORD) { console.error('请用 PASSWORD 环境变量传入 admin 密码（种子程序启动时随机生成）'); }
const PASSWORD = __ENV.PASSWORD || '';

export function setup() {
  const res = http.post(`${BASE}/api/v1/auth/login`,
    JSON.stringify({ username: USERNAME, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } });
  check(res, { 'login 200': (r) => r.status === 200 });
  return { token: res.json('token') };
}

export default function (data) {
  const auth = { headers: { Authorization: `Bearer ${data.token}` } };

  group('dashboard', () => {
    const r1 = http.get(`${BASE}/api/v1/dashboard/summary`, auth);
    check(r1, { 'summary 200': (r) => r.status === 200 });

    const r2 = http.get(`${BASE}/api/v1/dashboard/series/settlement?days=14`, auth);
    check(r2, { 'settlement series 200': (r) => r.status === 200 });

    const r3 = http.get(`${BASE}/api/v1/dashboard/series/freq?days=14`, auth);
    check(r3, { 'freq series 200': (r) => r.status === 200 });
  });

  group('lists', () => {
    const r1 = http.get(`${BASE}/api/v1/customers?limit=20`, auth);
    check(r1, { 'customers 200': (r) => r.status === 200 });

    const r2 = http.get(`${BASE}/api/v1/retail/contracts`, auth);
    check(r2, { 'contracts 200': (r) => r.status === 200 });
  });

  sleep(1);
}
