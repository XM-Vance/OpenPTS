'use client';

import React from 'react';
import { format, parseISO, getDay, isValid } from 'date-fns';

/**
 * Recharts 自定义 Tooltip 接收的属性类型
 */
export interface CustomTooltipProps {
  active?: boolean;
  payload?: Record<string, any>[];
  label?: string;
  /** 统一单位 */
  unit?: string;
  /** 各数据键对应单位映射 */
  unitMap?: Record<string, string>;
  /** 额外渲染行 */
  extraRows?: (payload: Record<string, any>[]) => React.ReactNode;
}

/** 从 YYYY-MM-DD 格式的 label 推导中文星期 */
const getWeekday = (label: string): string => {
  try {
    if (!label || label.length !== 10) return '';
    const date = parseISO(label);
    if (isValid(date)) {
      const days = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
      return ` (${days[getDay(date)]})`;
    }
  } catch {
    // ignore
  }
  return '';
};

/**
 * Recharts 自定义悬浮提示
 * - 适配 LineChart / BarChart / AreaChart 等
 * - 支持单位映射、日期星期推导
 */
export function CustomTooltip({
  active,
  payload,
  label,
  unit = '',
  unitMap = {},
  extraRows,
}: CustomTooltipProps) {
  if (!active || !payload || payload.length === 0) return null;

  const periodType = payload[0]?.payload?.period_type as string | undefined;
  const weekdayStr = getWeekday(label ?? '');

  return (
    <div className="rounded-md border bg-white px-3 py-2 shadow-lg dark:bg-zinc-900">
      <p className="mb-1.5 text-xs font-semibold text-zinc-700 dark:text-zinc-300">
        {label}
        {weekdayStr}
        {periodType ? ` (${periodType})` : ''}
      </p>

      {payload.map((pld, idx) => {
        const dataKey = pld.dataKey as string;
        // 优先取 payload 中的 value，兼容 LineChart
        const val = pld.value !== undefined
          ? pld.value
          : (pld.payload as Record<string, unknown>)?.[dataKey];
        const isValidNum = val !== null && val !== undefined && !isNaN(Number(val));
        const displayValue = isValidNum ? Number(val).toFixed(2) : 'N/A';
        const displayUnit = isValidNum && unit ? (unitMap[dataKey] || unit) : '';
        const color = (pld.stroke || pld.fill || pld.color || '#333') as string;

        return (
          <div key={idx} className="flex items-center gap-1.5 py-0.5 text-xs">
            <span
              className="inline-block h-2 w-2 shrink-0 rounded-sm"
              style={{ backgroundColor: color }}
            />
            <span style={{ color }}>
              {pld.name}: <span className="font-semibold">{displayValue} {displayUnit}</span>
            </span>
          </div>
        );
      })}

      {extraRows && extraRows(payload)}
    </div>
  );
}
