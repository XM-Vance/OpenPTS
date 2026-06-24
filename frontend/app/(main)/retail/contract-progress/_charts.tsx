'use client';

// 合同进度趋势图（计划 vs 实际 + 完成率）。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import {
  Bar,
  CartesianGrid,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { CustomTooltip } from '@/components/charts/custom-tooltip';

export function PlanActualComposed({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <ComposedChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis dataKey="month" tick={{ fontSize: 12 }} />
        <YAxis tick={{ fontSize: 12 }} width={72} />
        <RechartsTooltip content={<CustomTooltip />} />
        <Bar dataKey="计划电量" fill="#94a3b8" isAnimationActive={false} name="计划电量 (MWh)" />
        <Bar dataKey="实际电量" fill="#2563eb" isAnimationActive={false} name="实际电量 (MWh)" />
      </ComposedChart>
    </ResponsiveContainer>
  );
}

export function CompletionComposed({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <ComposedChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis dataKey="month" tick={{ fontSize: 12 }} />
        <YAxis tick={{ fontSize: 12 }} width={48} domain={[0, 'auto']} />
        <RechartsTooltip content={<CustomTooltip unit="%" />} />
        <Bar dataKey="完成率" fill="#10b981" isAnimationActive={false} name="完成率 (%)" />
        <Line
          type="monotone"
          dataKey="完成率"
          stroke="#f59e0b"
          strokeWidth={2}
          dot={false}
          isAnimationActive={false}
          name="完成率趋势 (%)"
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
