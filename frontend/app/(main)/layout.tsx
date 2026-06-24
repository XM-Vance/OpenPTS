import type { ReactNode } from 'react';
import { AppShell } from '@/components/layout/app-shell';

// (main) 路由组布局：所有业务页面共用侧边栏 + 顶栏外壳。
// 注意 (main) 是路由组，不影响 URL（/dashboard 仍是 /dashboard）。
export default function MainLayout({ children }: { children: ReactNode }) {
  return <AppShell>{children}</AppShell>;
}
