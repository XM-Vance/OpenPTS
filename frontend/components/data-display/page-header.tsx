'use client';

import React from 'react';
import { cn } from '@/lib/utils';

export interface PageHeaderProps {
  /** 页面标题 */
  title: string;
  /** 描述文字 */
  description?: string;
  /** 面包屑 */
  breadcrumb?: React.ReactNode;
  /** 右侧操作按钮区 */
  actions?: React.ReactNode;
  /** 容器 className */
  className?: string;
}

/**
 * 页面标题栏
 * - 标题 + 描述 + 面包屑 + 操作按钮
 */
export function PageHeader({
  title,
  description,
  breadcrumb,
  actions,
  className,
}: PageHeaderProps) {
  return (
    <div className={cn('mb-6', className)}>
      {breadcrumb && <div className="mb-2">{breadcrumb}</div>}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
          {description && (
            <p className="mt-1 text-sm text-muted-foreground">{description}</p>
          )}
        </div>
        {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
      </div>
    </div>
  );
}
