'use client';

import type { ReactNode } from 'react';
import { useAuth } from '@/lib/auth/context';
import { Sidebar } from './sidebar';
import { Header } from './header';
import { OrgReadonlyNotice } from './org-readonly-notice';
import { AlertStream } from '@/components/realtime/alert-stream';

// 业务页面统一外壳：侧边栏 + 顶栏 + 内容区，并做登录态守卫。
export function AppShell({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">加载中...</p>
      </div>
    );
  }
  // 未登录时 AuthContext 已触发跳转 /login，这里直接不渲染
  if (!user) return null;

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <OrgReadonlyNotice />
        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
      <AlertStream />
    </div>
  );
}
