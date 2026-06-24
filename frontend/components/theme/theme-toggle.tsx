'use client';

import { Monitor, Moon, Sun } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useTheme } from '@/components/theme/theme-provider';

// 三态循环按钮：system → light → dark → system
export function ThemeToggle() {
  const { theme, resolved, setTheme } = useTheme();

  const cycle = () => {
    setTheme(theme === 'system' ? 'light' : theme === 'light' ? 'dark' : 'system');
  };

  const Icon = theme === 'system' ? Monitor : resolved === 'dark' ? Moon : Sun;
  const label =
    theme === 'system' ? '跟随系统' : resolved === 'dark' ? '深色' : '浅色';

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={cycle}
      title={`主题：${label}（点击切换）`}
      aria-label={`切换主题（当前：${label}）`}
      className="h-9 px-2"
    >
      <Icon className="h-4 w-4" />
    </Button>
  );
}
