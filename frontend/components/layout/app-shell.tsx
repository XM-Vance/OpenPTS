'use client';

import type { ReactNode } from 'react';
import { useMemo, useState } from 'react';
import { usePathname } from 'next/navigation';
import { ShieldX } from 'lucide-react';
import { useAuth } from '@/lib/auth/context';
import { usePermission } from '@/lib/auth/use-permission';
import { MENU } from '@/lib/menu';
import { useIsMobile } from '@/hooks/use-mobile';
import { Sheet } from '@/components/ui/sheet';
import { Sidebar, SidebarNav } from './sidebar';
import { Header } from './header';
import { OrgReadonlyNotice } from './org-readonly-notice';
import { AlertStream } from '@/components/realtime/alert-stream';
import { EmptyState } from '@/components/feedback';

// 预构建 href → 所需权限码映射（菜单加载时一次）。用于页面级权限守卫：
// 菜单只控制「可见性」，知道 URL 仍能直连；此处对命中映射的页面校验 read 权限，
// 无权限则渲染禁止访问态而非业务内容（P0-F1）。
function buildHrefPermissionMap(): Map<string, string> {
  const m = new Map<string, string>();
  for (const g of MENU) {
    for (const it of g.items) {
      if (it.permission) m.set(it.href, it.permission);
    }
  }
  return m;
}

// 业务页面统一外壳：侧边栏 + 顶栏 + 内容区，并做登录态 + 页面级权限守卫。
// 移动端（<768px）：侧边栏收进 Sheet 抽屉，顶栏加汉堡按钮，内容区 padding 收窄。
// 桌面端（≥768px）：与原布局完全一致（hidden md:flex / md:p-6）。
export function AppShell({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const { has } = usePermission();
  const pathname = usePathname();
  const hrefPerm = useMemo(() => buildHrefPermissionMap(), []);
  const isMobile = useIsMobile();
  const [navOpen, setNavOpen] = useState(false);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">加载中...</p>
      </div>
    );
  }
  // 未登录时 AuthContext 已触发跳转 /login，这里直接不渲染
  if (!user) return null;

  // 移动端导航抽屉：点链接后自动关闭。
  const mobileNav = isMobile ? (
    <Sheet open={navOpen} onClose={() => setNavOpen(false)} title={undefined}>
      <SidebarNav onNavigate={() => setNavOpen(false)} />
    </Sheet>
  ) : null;

  const openNav = () => setNavOpen(true);

  // 页面级权限守卫（P0-F1）：当前路径若绑了权限码且用户无该权限，拦截渲染。
  // 用最长前缀匹配处理子路由（如 /retail/contracts/xxx 命中 /retail/contracts）。
  const required = lookupPermission(hrefPerm, pathname);
  if (required && !has(required)) {
    return (
      <div className="flex h-screen overflow-hidden">
        <Sidebar />
        {mobileNav}
        <div className="flex flex-1 flex-col overflow-hidden">
          <Header onMenuClick={openNav} showMenuButton={isMobile} />
          <OrgReadonlyNotice />
          <main className="flex-1 overflow-y-auto p-3 md:p-6">
            <EmptyState
              icon={<ShieldX className="h-12 w-12" />}
              title="无访问权限"
              description="您没有访问该页面的权限，请联系管理员开通。"
            />
          </main>
        </div>
        <AlertStream />
      </div>
    );
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar />
      {mobileNav}
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header onMenuClick={openNav} showMenuButton={isMobile} />
        <OrgReadonlyNotice />
        <main className="flex-1 overflow-y-auto p-3 md:p-6">{children}</main>
      </div>
      <AlertStream />
    </div>
  );
}

// lookupPermission 最长前缀匹配：/retail/contracts/123 → /retail/contracts。
function lookupPermission(map: Map<string, string>, pathname: string): string | undefined {
  // 精确命中
  if (map.has(pathname)) return map.get(pathname);
  // 前缀匹配（按段，取最长）
  let best: string | undefined;
  let bestLen = -1;
  for (const [href, perm] of map) {
    if (pathname === href || pathname.startsWith(href + '/')) {
      if (href.length > bestLen) {
        best = perm;
        bestLen = href.length;
      }
    }
  }
  return best;
}
