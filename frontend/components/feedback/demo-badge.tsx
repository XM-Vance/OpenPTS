'use client';

import React from 'react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { Sparkles } from 'lucide-react';

export interface DemoBadgeProps {
  /** 角标文案，默认「演示数据」 */
  label?: string;
  /** 悬浮提示，说明该数据为何是演示态 */
  tooltip?: string;
  /** 额外 className */
  className?: string;
}

/**
 * 演示数据角标。
 *
 * 标记由 Math.random / 本地 mock 生成的非实时数据，避免用户误以为是真业务结果。
 * 复用项目既有的 demo 机制（后端 DemoGate + GenerateDemoData 端点）。
 *
 * 用法：
 *   <Card>
 *     <CardHeader>
 *       <DemoBadge tooltip="后端尚未实现实时监控端点" />
 *       <CardTitle>实时出力</CardTitle>
 *     </CardHeader>
 *     ...
 *   </Card>
 */
export function DemoBadge({
  label = '演示数据',
  tooltip = '此数据为演示用占位，非实时业务数据',
  className,
}: DemoBadgeProps) {
  return (
    <Badge
      variant="warning"
      title={tooltip}
      className={cn('gap-1 bg-amber-100 text-amber-700 hover:bg-amber-100', className)}
    >
      <Sparkles className="h-3 w-3" />
      {label}
    </Badge>
  );
}
