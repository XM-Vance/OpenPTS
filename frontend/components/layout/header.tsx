'use client';

import { useState, useCallback } from 'react';
import { LogOut, Users, Search } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { useAuth } from '@/lib/auth/context';
import { ThemeToggle } from '@/components/theme/theme-toggle';
import { LocaleToggle } from '@/components/i18n/locale-toggle';
import { useI18n } from '@/lib/i18n/context';
import { apiClient } from '@/lib/api/client';

export function Header() {
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
    <header className="flex h-14 shrink-0 items-center justify-between border-b bg-card px-4">
      {/* 左侧：全局搜索 */}
      <div className="flex items-center gap-2">
        {searchOpen ? (
          <form onSubmit={handleSearch} className="flex items-center gap-2">
            <Input
              autoFocus
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              placeholder="搜索..."
              className="h-8 w-56 text-sm"
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

      {/* 右侧：原有控件 */}
      <div className="flex items-center gap-2">
        {online && online.count > 0 && (
          <Badge
            variant="success"
            className="gap-1"
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
            className="h-8 rounded-md border bg-background px-2 text-sm"
          >
            {isHQ && <option value="*">全部省</option>}
            {accessibleOrgs.map((o) => (
              <option key={o.id} value={o.id}>
                {o.name}
              </option>
            ))}
          </select>
        )}
        <LocaleToggle />
        <ThemeToggle />
        <span className="text-sm text-muted-foreground">
          {user?.display_name || user?.username}
        </span>
        <Button variant="ghost" size="sm" onClick={logout}>
          <LogOut className="h-4 w-4" />
          {t('common.logout')}
        </Button>
      </div>
    </header>
  );
}
