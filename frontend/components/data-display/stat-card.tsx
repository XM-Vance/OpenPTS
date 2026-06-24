'use client';

import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

export interface StatCardProps {
  /** 统计项标题 */
  title: string;
  /** 数值 */
  value: string | number;
  /** 变化率（百分比），如 12.5 表示 +12.5%，-3.2 表示 -3.2% */
  trend?: number | null;
  /** 趋势说明文字 */
  trendLabel?: string;
  /** 图标 */
  icon?: React.ReactNode;
  /** 卡片 className */
  className?: string;
  /** 点击事件 */
  onClick?: () => void;
}

/**
 * 统计数字卡片
 * - 标题 + 数值 + 趋势箭头 + 图标
 * - 趋势正值绿色向上，负值红色向下
 */
export function StatCard({
  title,
  value,
  trend,
  trendLabel,
  icon,
  className,
  onClick,
}: StatCardProps) {
  const isPositive = trend !== null && trend !== undefined && trend > 0;
  const isNegative = trend !== null && trend !== undefined && trend < 0;
  const isFlat = trend === 0;

  return (
    <Card
      className={cn('transition-shadow hover:shadow-md', onClick && 'cursor-pointer', className)}
      onClick={onClick}
    >
      <CardContent className="flex items-start justify-between p-4">
        <div className="space-y-1">
          <p className="text-sm text-muted-foreground">{title}</p>
          <p className="text-2xl font-bold leading-none">{value}</p>
          {(trend !== null && trend !== undefined) && (
            <div className="flex items-center gap-1 pt-1">
              {isPositive && (
                <TrendingUp className="h-3.5 w-3.5 text-emerald-500" />
              )}
              {isNegative && (
                <TrendingDown className="h-3.5 w-3.5 text-red-500" />
              )}
              {isFlat && (
                <Minus className="h-3.5 w-3.5 text-muted-foreground" />
              )}
              <span
                className={cn(
                  'text-xs font-medium',
                  isPositive && 'text-emerald-500',
                  isNegative && 'text-red-500',
                  isFlat && 'text-muted-foreground',
                )}
              >
                {isPositive && '+'}
                {trend.toFixed(1)}%
              </span>
              {trendLabel && (
                <span className="text-xs text-muted-foreground">{trendLabel}</span>
              )}
            </div>
          )}
        </div>
        {icon && (
          <div className="rounded-md bg-muted p-2 text-muted-foreground">{icon}</div>
        )}
      </CardContent>
    </Card>
  );
}
