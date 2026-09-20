'use client';

import React from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Inbox } from 'lucide-react';

export interface EmptyStateProps {
  /** 图标 */
  icon?: React.ReactNode;
  /** 标题 */
  title?: React.ReactNode;
  /** 描述 */
  description?: string;
  /** 操作按钮 */
  action?: { label: string; onClick: () => void };
  /** 容器 className */
  className?: string;
  /** 紧凑模式：行内小图标 + 单行文字，用于图表卡/列表内的局部空态 */
  compact?: boolean;
}

/**
 * 空数据状态
 * - 图标 + 标题 + 描述 + 操作按钮（页面级）
 * - compact：行内小图标 + 单行文字（图表卡/列表局部）
 */
export function EmptyState({
  icon,
  title = '暂无数据',
  description,
  action,
  className,
  compact,
}: EmptyStateProps) {
  if (compact) {
    return (
      <div
        className={cn(
          'flex w-full items-center justify-center gap-2 text-sm text-muted-foreground',
          className,
        )}
      >
        {icon ?? <Inbox className="h-4 w-4 shrink-0 text-muted-foreground/50" aria-hidden />}
        <span>
          {title}
          {description ? `，${description}` : ''}
        </span>
      </div>
    );
  }

  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-16 text-center',
        className,
      )}
    >
      <div className="mb-4 text-muted-foreground/50">
        {icon ?? <Inbox className="h-12 w-12" aria-hidden />}
      </div>
      <h3 className="text-lg font-medium">{title}</h3>
      {description && (
        <p className="mt-1 max-w-md text-sm text-muted-foreground">{description}</p>
      )}
      {action && (
        <Button className="mt-4" onClick={action.onClick}>
          {action.label}
        </Button>
      )}
    </div>
  );
}
