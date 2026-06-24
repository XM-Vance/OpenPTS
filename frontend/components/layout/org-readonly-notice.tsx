'use client';

import { Info } from 'lucide-react';
import { useAuth } from '@/lib/auth/context';

// 总部「全部省」(activeOrg === '*') 是跨省汇总的只读视图：
// 省隔离的写入（新增/生成演示数据等）会被后端拒绝（400「请先选择具体省份」）。
// 这里给出醒目提示，避免用户在该视图下尝试写操作而困惑。
// 仅对处于「全部省」活跃态的总部用户显示；其余情况（已选具体省）不渲染。
export function OrgReadonlyNotice() {
  const { activeOrg } = useAuth();
  if (activeOrg !== '*') return null;
  return (
    <div
      role="status"
      className="flex shrink-0 items-center gap-2 border-b border-amber-300/60 bg-amber-50 px-4 py-2 text-sm text-amber-800 dark:border-amber-700/50 dark:bg-amber-950/40 dark:text-amber-300"
    >
      <Info className="h-4 w-4 shrink-0" />
      <span>
        当前为总部「全部省」汇总视图，仅供查看。如需新增 / 编辑 / 生成数据，请先在右上角切换到具体省份。
      </span>
    </div>
  );
}
