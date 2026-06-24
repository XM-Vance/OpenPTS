'use client';

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react';
import { usePathname, useRouter } from 'next/navigation';
import {
  login as loginApi,
  logout as logoutApi,
  me as meApi,
  myPermissions as myPermissionsApi,
  type LoginParams,
  type MeResponse,
  type Org,
} from '@/lib/api/auth';
import {
  getActiveOrg,
  setActiveOrg as setActiveOrgStorage,
  clearActiveOrg,
  purgeLegacyToken,
} from './token';

interface AuthState {
  user: MeResponse | null;
  permissions: string[];
  loading: boolean;
  login: (params: LoginParams) => Promise<void>;
  logout: () => void;
  // 多租户
  activeOrg: string;
  setActiveOrg: (orgId: string) => void;
  accessibleOrgs: Org[];
  isHQ: boolean;
}

// 选定初始活跃省：优先用本地存储（且仍可访问），否则用主省/第一个可访问省。
function pickActiveOrg(u: MeResponse): string {
  const stored = getActiveOrg();
  const ids = (u.orgs ?? []).map((o) => o.id);
  if (stored && (u.is_hq || ids.includes(stored))) return stored;
  return u.org_active || ids[0] || '';
}

const AuthContext = createContext<AuthState | null>(null);

const PUBLIC_PATHS = new Set(['/login']);

export function AuthProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<MeResponse | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeOrg, setActiveOrgState] = useState('');

  // 启动时尝试恢复登录态 + 权限。
  // 登录态在 httpOnly Cookie 里(JS 不可读),无法预判"是否已登录",
  // 直接探测 /auth/me:成功即恢复,401 即去登录页。
  useEffect(() => {
    purgeLegacyToken(); // 清理旧版本遗留在 localStorage 的 JWT
    Promise.all([meApi(), myPermissionsApi()])
      .then(([u, perms]) => {
        setUser(u);
        setPermissions(perms);
        const a = pickActiveOrg(u);
        setActiveOrgStorage(a);
        setActiveOrgState(a);
      })
      .catch(() => {
        if (!PUBLIC_PATHS.has(pathname)) {
          router.replace('/login');
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const login = useCallback(
    async (params: LoginParams) => {
      // 登录态由后端经 Set-Cookie 下发(httpOnly),前端不接触 token
      await loginApi(params);
      const [u, perms] = await Promise.all([meApi(), myPermissionsApi()]);
      setUser(u);
      setPermissions(perms);
      const a = pickActiveOrg(u);
      setActiveOrgStorage(a);
      setActiveOrgState(a);
      router.replace('/dashboard');
    },
    [router],
  );

  const logout = useCallback(() => {
    // 通知后端清 Cookie;失败(如已过期)不阻断本地状态清理
    logoutApi().catch(() => {});
    clearActiveOrg();
    setUser(null);
    setPermissions([]);
    setActiveOrgState('');
    router.replace('/login');
  }, [router]);

  const setActiveOrg = useCallback((orgId: string) => {
    setActiveOrgStorage(orgId);
    setActiveOrgState(orgId);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        permissions,
        loading,
        login,
        logout,
        activeOrg,
        setActiveOrg,
        accessibleOrgs: user?.orgs ?? [],
        isHQ: user?.is_hq ?? false,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth 必须在 AuthProvider 内使用');
  return ctx;
}
