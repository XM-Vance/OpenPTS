/**
 * 图表调色板（全局唯一来源）
 *
 * - 系列色取 Tailwind 500–600 明度档，明暗两套主题背景（纯白 / slate-900）下均可读；
 *   不随主题切换，避免 Recharts SVG 属性中引用 CSS 变量的兼容性问题。
 * - 语义色遵循国内金融惯例：红涨绿跌（正向红、负向绿）。
 * - 网格线/轴文字颜色由 globals.css 的 .recharts-* 规则按主题统一接管，勿在此定义。
 */
export const CHART_SERIES = [
  '#2563eb', // 蓝——主系列（购电 / 日前价格）
  '#f59e0b', // 琥珀——与品牌强调色呼应（实时 / 均价）
  '#10b981', // 绿
  '#8b5cf6', // 紫
  '#f43f5e', // 玫红
  '#06b6d4', // 青
  '#ec4899', // 粉
  '#14b8a6', // 蓝绿
] as const;

/** 语义色：红涨绿跌 */
export const RISE_COLOR = '#ef4444';
export const FALL_COLOR = '#10b981';

/** 按正负取色：>=0 红（涨/盈），<0 绿（跌/亏） */
export function riseFallColor(value: number): string {
  return value >= 0 ? RISE_COLOR : FALL_COLOR;
}
