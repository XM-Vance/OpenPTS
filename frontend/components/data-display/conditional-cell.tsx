'use client';

import { cn } from '@/lib/utils';

/* ------------------------------------------------------------------ */
/*  ConditionalCell — 根据值自动着色                                    */
/* ------------------------------------------------------------------ */

export interface ConditionalCellProps {
  /** 数值 */
  value: number;
  /** 显示格式，默认原始值 */
  format?: (v: number) => string;
  /** 正负色模式：正值绿色、负值红色（默认开启） */
  posNeg?: boolean;
  /** 超标阈值：超过则红色高亮 */
  threshold?: number;
  /** 警告阈值：超过则黄色警告 */
  warnThreshold?: number;
  /** 额外 class */
  className?: string;
}

export function ConditionalCell({
  value,
  format,
  posNeg = true,
  threshold,
  warnThreshold,
  className,
}: ConditionalCellProps) {
  const display = format ? format(value) : String(value);

  let colorClass = 'text-foreground';

  if (threshold !== undefined && Math.abs(value) > threshold) {
    colorClass = 'text-red-600 font-semibold';
  } else if (warnThreshold !== undefined && Math.abs(value) > warnThreshold) {
    colorClass = 'text-amber-600';
  } else if (posNeg) {
    if (value > 0) colorClass = 'text-emerald-600';
    else if (value < 0) colorClass = 'text-red-600';
  }

  return <span className={cn(colorClass, className)}>{display}</span>;
}

/* ------------------------------------------------------------------ */
/*  DeltaBadge — 环比变化小标签                                         */
/* ------------------------------------------------------------------ */

export interface DeltaBadgeProps {
  /** 变化值，如 2.3 表示 +2.3% */
  value: number;
  /** 额外 class */
  className?: string;
}

export function DeltaBadge({ value, className }: DeltaBadgeProps) {
  const isUp = value > 0;
  const isZero = value === 0;

  const arrow = isZero ? '—' : isUp ? '↑' : '↓';
  const sign = isUp ? '+' : '';
  const colorClass = isZero
    ? 'text-muted-foreground bg-muted'
    : isUp
      ? 'text-emerald-700 bg-emerald-50 dark:text-emerald-400 dark:bg-emerald-950'
      : 'text-red-700 bg-red-50 dark:text-red-400 dark:bg-red-950';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-xs font-medium',
        colorClass,
        className,
      )}
    >
      {arrow}
      {sign}
      {Math.abs(value).toFixed(1)}%
    </span>
  );
}
