'use client';

import React from 'react';
import { cn } from '@/lib/utils';

export interface PageLoadingProps {
  /** 提示文字 */
  text?: string;
  /** 容器 className */
  className?: string;
}

/**
 * 页面加载中状态
 * - 全屏居中 spinner + 文字
 */
export function PageLoading({ text = '加载中...', className }: PageLoadingProps) {
  return (
    <div className={cn('flex h-full min-h-[300px] flex-col items-center justify-center gap-3', className)}>
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      <p className="text-sm text-muted-foreground">{text}</p>
    </div>
  );
}
