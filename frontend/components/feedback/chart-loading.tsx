'use client';

import { cn } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';

export interface ChartLoadingProps {
  /** 容器 className（可用 h-XX 覆盖默认 h-40） */
  className?: string;
}

/**
 * 图表/面板区加载骨架屏
 * - 替代「加载中...」文字占位，避免加载完成后页面跳动
 * - 高度默认 h-40，通过 className="h-48" 等按容器调整
 */
export function ChartLoading({ className }: ChartLoadingProps) {
  return (
    <div
      className={cn(
        'flex h-40 w-full flex-col justify-center gap-3',
        className,
      )}
    >
      <Skeleton className="h-4 w-1/3" />
      <Skeleton className="h-16 w-full" />
    </div>
  );
}
