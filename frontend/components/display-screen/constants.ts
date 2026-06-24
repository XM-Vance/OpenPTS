import type { DisplayOverview } from '@/lib/api/display-screen';

/* ========== 布局常量 ========== */
export const BASE_WIDTH = 1920;
export const BASE_HEIGHT = 1080;
export const POLL_INTERVAL = 30_000; // 30 秒轮询

/* ========== 深色配色（Sprint 6 更新） ========== */
export const C = {
  bg: '#0D1B2A',
  bgGradient: 'radial-gradient(ellipse at 50% 0%, #112D4E 0%, #0D1B2A 70%)',
  cardBg: 'rgba(13, 27, 42, 0.85)',
  cardBorder: 'rgba(24, 144, 255, 0.3)',
  blue: '#1890FF',           // 科技蓝（主数据色）
  cyan: '#00e5ff',
  cyanDim: 'rgba(0, 229, 255, 0.5)',
  gold: '#ffd54f',
  goldGlow: 'rgba(255, 213, 79, 0.3)',
  green: '#69f0ae',
  pink: '#ff80ab',
  purple: '#b388ff',
  orange: '#ffab40',
  alertRed: '#FF4D4F',      // 预警红
  text: '#ffffff',
  textDim: '#8ba4c7',
  gridLine: 'rgba(24, 144, 255, 0.08)',
  divider: 'rgba(24, 144, 255, 0.2)',
  chartGrid: 'rgba(24, 144, 255, 0.12)',
};

// 饼图配色
export const PIE_COLORS = [C.cyan, C.gold, C.green, C.pink, C.purple, C.orange];

export const FALLBACK_OVERVIEW: DisplayOverview = {
  total_customers: 156,
  active_contracts: 89,
  today_energy_mwh: 4852.6,
  today_revenue: 2356800,
  month_energy_mwh: 138750,
  month_revenue: 62340000,
  avg_price: 415.38,
  deviation_rate: 3.25,
  alert_count: 3,
  green_ratio: 28.5,
};
