/**
 * 客户域共享类型与常量（_view.tsx 与 _charts.tsx 共用）。
 * 从 _view.tsx 拆出，避免循环依赖。
 */

export interface Customer360 {
  id: string;
  name: string;
  industry: string;
  electricity: number; // MWh/month
  revenue: number; // 万元/month
  risk: number; // 1-5, bubble size
  rating: number; // 1-5 stars
  status: 'potential' | 'interested' | 'contracted' | 'renewed' | 'churned';
  cooperation: string;
}

// 客户生命周期状态字典（模块级常量，_view 与 _charts 共用，且不进入 Hook 依赖）
export const STATUS_NAMES: Customer360['status'][] = [
  'potential',
  'interested',
  'contracted',
  'renewed',
  'churned',
];

export const STATUS_LABELS: Record<string, string> = {
  potential: '潜在',
  interested: '意向',
  contracted: '签约',
  renewed: '续约',
  churned: '流失',
};

export const STATUS_COLORS: Record<string, string> = {
  potential: '#94a3b8',
  interested: '#60a5fa',
  contracted: '#34d399',
  renewed: '#6366f1',
  churned: '#f87171',
};

export const SCATTER_COLORS = ['#6366f1', '#3b82f6', '#10b981', '#f59e0b', '#ef4444'];
