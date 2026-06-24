'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState, type ReactNode } from 'react';
import { AuthProvider } from './auth/context';
import { ThemeProvider } from '@/components/theme/theme-provider';
import { I18nProvider } from './i18n/context';

// 缓存策略：按 queryKey 第一段分类，提供差异化 staleTime（react-query v5）。
// - LONG：字典/参考数据，30 分钟（很少变）
// - MID：列表/详情，30 秒（默认）
// - LIVE：实时面板，5 秒（仪表盘、调度、安全大屏）
const LONG_CACHE_PREFIXES = new Set([
  'pricing-models',
  'tou-rules',
  'grid-agency',
  'approval-templates',
]);
const LIVE_CACHE_PREFIXES = new Set([
  'dashboard-summary',
  'dashboard-settlement-series',
  'dashboard-freq-series',
  'scheduler-jobs',
  'scheduler-runs',
  'rpa-runs',
  'rpa-jobs',
  'security-overview',
]);

export function Providers({ children }: { children: ReactNode }) {
  const [client] = useState(() => {
    const qc = new QueryClient({
      defaultOptions: {
        queries: {
          staleTime: 30_000, // 默认 30 秒（列表/详情，即 MID 档）
          gcTime: 5 * 60_000, // 5 分钟未引用回收
          retry: 1,
          refetchOnWindowFocus: false,
        },
      },
    });
    // 按 queryKey 首段分档覆盖 staleTime——用官方 setQueryDefaults(按 key 前缀匹配),
    // 取代旧版「订阅 cache added 事件 + (query as any).setOptions」的脆弱 hack
    // (依赖内部 API,react-query 升级易碎)。
    for (const k of LIVE_CACHE_PREFIXES) qc.setQueryDefaults([k], { staleTime: 5_000 });
    for (const k of LONG_CACHE_PREFIXES) qc.setQueryDefaults([k], { staleTime: 30 * 60_000 });
    return qc;
  });

  return (
    <ThemeProvider>
      <I18nProvider>
        <QueryClientProvider client={client}>
          <AuthProvider>{children}</AuthProvider>
        </QueryClientProvider>
      </I18nProvider>
    </ThemeProvider>
  );
}
