// 登录态自 P1-8 起走 httpOnly Cookie(后端 /auth/login 写入、/auth/logout 清除),
// 前端不再读写 JWT——XSS 无法窃取凭证。本文件仅保留"当前活跃省"的本地存取。

// 当前活跃省（组织 ID）存取。请求拦截器读取它注入 X-Org-Id 头。
const ORG_KEY = 'ptis.active_org';

export function getActiveOrg(): string | null {
  if (typeof window === 'undefined') return null;
  return window.localStorage.getItem(ORG_KEY);
}

export function setActiveOrg(orgId: string): void {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(ORG_KEY, orgId);
}

export function clearActiveOrg(): void {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(ORG_KEY);
}

// 历史遗留:清除旧版本存在 localStorage 的 JWT(升级后一次性清理,可在数月后移除)。
const LEGACY_TOKEN_KEY = 'ptis.token';
export function purgeLegacyToken(): void {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(LEGACY_TOKEN_KEY);
}
