'use client';

import { Languages } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { LOCALES } from '@/lib/i18n/messages';
import { useI18n } from '@/lib/i18n/context';

// 两态循环：zh-CN ↔ en-US
export function LocaleToggle() {
  const { locale, setLocale } = useI18n();
  const cycle = () => setLocale(locale === 'zh-CN' ? 'en-US' : 'zh-CN');
  const label = LOCALES.find((l) => l.code === locale)?.label ?? locale;
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={cycle}
      title={`当前：${label}（点击切换）`}
      aria-label={`切换语言（当前：${label}）`}
      className="h-9 gap-1 px-2"
    >
      <Languages className="h-4 w-4" />
      <span className="text-xs uppercase">{locale === 'zh-CN' ? '中' : 'EN'}</span>
    </Button>
  );
}
