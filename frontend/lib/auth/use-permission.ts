'use client';

import { useAuth } from './context';

// 权限检查 Hook，基于 AuthContext 中加载的当前用户权限码集合。
export function usePermission() {
  const { permissions } = useAuth();
  return {
    permissions,
    has: (code: string) => permissions.includes(code),
    hasAny: (...codes: string[]) => codes.some((c) => permissions.includes(c)),
    hasAll: (...codes: string[]) => codes.every((c) => permissions.includes(c)),
  };
}
