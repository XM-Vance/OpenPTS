import axios, { AxiosError, type AxiosInstance } from 'axios';
import { getActiveOrg } from '@/lib/auth/token';

// ── 类型契约（P1-6）─────────────────────────────────────────────
// 业务实体的 TS 类型应优先复用 `./types.gen.ts` 里由 OpenAPI 规范生成的 `Schema*`
// 类型，而非手写 interface，避免后端改字段后前端漂移到运行时才暴露。
// 已接通的域：audit / customers / auth / scheduler。
// 新增域的接通步骤：
//   1. 在 backend/internal/handler/openapi.yaml 的 components.schemas 补/改该实体
//      （字段与后端 db 结构对齐，required 标必填）。
//   2. `npm run gen-api-types` 重新生成 types.gen.ts。
//   3. 在对应 lib/api/<域>.ts 里 `export type X = SchemaX;` 替换手写 interface。
//   4. CI 的 OpenAPI 漂移门禁（ci.yml）会校验 types.gen.ts 与 openapi.yaml 一致。
// ────────────────────────────────────────────────────────────────

// 后端 API 基址。
// 客户端始终留空（走 Next.js rewrites 同源代理，浏览器无法直连 Docker 内部网络）。
// 服务端渲染时可通过 API_SERVER_BASE 直连后端（跳过 rewrites）。
const baseURL = typeof window === 'undefined'
  ? (process.env.API_SERVER_BASE || '')
  : '';

export const apiClient: AxiosInstance = axios.create({
  baseURL,
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
});

// 请求拦截:登录态在 httpOnly Cookie 中,同源请求浏览器自动携带,无需注入 Authorization。
apiClient.interceptors.request.use((config) => {
  // 多租户：注入当前活跃省，后端租户中间件据此校验/作用域
  const org = getActiveOrg();
  if (org) {
    config.headers['X-Org-Id'] = org;
  }
  return config;
});

export interface ApiError {
  error: string;
}

export function extractErrorMessage(err: unknown, fallback = '请求失败'): string {
  if (axios.isAxiosError<ApiError>(err)) {
    return err.response?.data?.error || err.message || fallback;
  }
  return fallback;
}
