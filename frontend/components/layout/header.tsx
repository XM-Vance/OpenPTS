'use client';

import { useState, useCallback } from 'react';
import { LogOut, Users, Search, Menu } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { useAuth } from '@/lib/auth/context';
import { ThemeToggle } from '@/components/theme/theme-toggle';
import { LocaleToggle } from '@/components/i18n/locale-toggle';
import { useI18n } from '@/lib/i18n/context';
import { apiClient } from '@/lib/api/client';

interface HeaderProps {
  /** 移动端汉堡按钮点击：打开导航抽屉 */
  onMenuClick?: () => void;
  /** 是否显示汉堡按钮（移动端） */
  showMenuButton?: boolean;
}

export function Header({ onMenuClick, showMenuButton }: HeaderProps) {
  const { user, logout, activeOrg, setActiveOrg, accessibleOrgs, isHQ } = useAuth();
  const { t } = useI18n();
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchValue, setSearchValue] = useState('');
  const { data: online } = useQuery({
    queryKey: ['online-users'],
    queryFn: async () => {
      const { data } = await apiClient.get<{ count: number; users: string[]; connections: number }>(
        '/api/v1/online',
      );
      return data;
    },
    refetchInterval: 15_000,
  });

  const handleSearch = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    alert('搜索功能开发中');
  }, []);

  return (
    <header className="flex h-14 shrink-0 items-center justify-between gap-2 border-b bg-card px-4">
      {/* 左侧：移动端汉堡 + 全局搜索 */}
      <div className="flex min-w-0 items-center gap-2">
        {showMenuButton && (
          <Button variant="ghost" size="icon" onClick={onMenuClick} aria-label="打开菜单">
            <Menu className="h-5 w-5" />
          </Button>
        )}
        {searchOpen ? (
          <form onSubmit={handleSearch} className="flex items-center gap-2">
            <Input
              autoFocus
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              placeholder="搜索..."
              className="h-8 w-40 text-sm sm:w-56"
              onBlur={() => {
                if (!searchValue) setSearchOpen(false);
              }}
            />
          </form>
        ) : (
          <Button variant="ghost" size="icon" onClick={() => setSearchOpen(true)} aria-label="打开搜索">
            <Search className="h-4 w-4" />
          </Button>
        )}
      </div>

      {/* 右侧：控件簇。移动端折叠次要项（在线徽章/locale/theme/用户名），保留省份切换 + 退出。 */}
      <div className="flex items-center gap-2">
        {online && online.count > 0 && (
          <Badge
            variant="success"
            className="hidden gap-1 sm:inline-flex"
            title={`在线用户: ${online.users.join(', ')}`}
          >
            <Users className="h-3 w-3" />
            {online.count} 在线
          </Badge>
        )}
        {(accessibleOrgs.length > 1 || isHQ) && (
          <select
            value={activeOrg}
            onChange={(e) => setActiveOrg(e.target.value)}
            title="切换省份"
            className="h-8 max-w-[7rem] rounded-md border bg-background px-2 text-sm sm:max-w-none"
          >
            {isHQ && <option value="*">全部省</option>}
            {accessibleOrgs.map((o) => (
              <option key={o.id} value={o.id}>
                {o.name}
              </option>
            ))}
          </select>
        )}
        <span className="hidden text-sm text-muted-foreground md:inline">
          {user?.display_name || user?.username}
        </span>
        <div className="hidden md:flex md:items-center md:gap-2">
          <LocaleToggle />
          <ThemeToggle />
        </div>
        <Button variant="ghost" size="sm" onClick={logout}>
          <LogOut className="h-4 w-4" />
          <span className="hidden sm:inline">{t('common.logout')}</span>
        </Button>
      </div>
    </header>
  );
}
