'use client';

/**
 * 客户分群图表（从 _view.tsx 拆出，降低主文件体积）。
 * 由 _view.tsx 经 next/dynamic 懒加载，recharts 从首屏 JS 剥离。
 */
import { useMemo } from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  ResponsiveContainer,
  Scatter,
  ScatterChart,
  Tooltip,
  XAxis,
  YAxis,
  ZAxis,
} from 'recharts';
import { Users, GitMerge } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DemoBadge } from '@/components/feedback';
import {
  STATUS_NAMES,
  STATUS_LABELS,
  STATUS_COLORS,
  SCATTER_COLORS,
  type Customer360,
} from './_shared';

/** 客户分群气泡图（用电量 × 收益 × 风险） */
export function CustomerBubbleChart({ data }: { data: Customer360[] }) {
  // Group by status for separate scatter series
  const grouped = useMemo(() => {
    const map: Record<string, Customer360[]> = {};
    for (const s of STATUS_NAMES) map[s] = [];
    for (const c of data) {
      if (map[c.status]) map[c.status].push(c);
    }
    return map;
  }, [data]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5 text-sm font-medium">
          <Users className="h-4 w-4 text-indigo-500" />
          客户分群气泡图
          <DemoBadge className="ml-auto" tooltip="用电量/收益/风险由 enrichCustomers 随机生成，非真实数据" />
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-80">
          <ResponsiveContainer width="100%" height="100%">
            <ScatterChart margin={{ top: 10, right: 20, bottom: 10, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                type="number"
                dataKey="electricity"
                name="用电量"
                unit=" MWh"
                tick={{ fontSize: 11 }}
                label={{ value: '用电量 (MWh/月)', position: 'insideBottom', offset: -5, style: { fontSize: 11 } }}
              />
              <YAxis
                type="number"
                dataKey="revenue"
                name="收益"
                unit=" 万"
                tick={{ fontSize: 11 }}
                label={{ value: '收益 (万/月)', angle: -90, position: 'insideLeft', style: { fontSize: 11 } }}
              />
              <ZAxis type="number" dataKey="risk" range={[60, 400]} name="风险等级" />
              <Tooltip
                cursor={{ strokeDasharray: '3 3' }}
                contentStyle={{ fontSize: 12 }}
                formatter={(value: number, name: string) => {
                  if (name === '用电量') return [`${value} MWh`, name];
                  if (name === '收益') return [`¥${value} 万`, name];
                  if (name === '风险等级') return [`${value}`, name];
                  return [value, name];
                }}
              />
              <Legend />
              {STATUS_NAMES.map((s, i) => (
                <Scatter
                  key={s}
                  name={STATUS_LABELS[s]}
                  data={grouped[s]}
                  fill={SCATTER_COLORS[i]}
                />
              ))}
            </ScatterChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}

/** 客户生命周期漏斗（横向条形） */
export function LifecycleFunnel({ data }: { data: Customer360[] }) {
  const funnelData = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const s of STATUS_NAMES) counts[s] = 0;
    for (const c of data) counts[c.status] = (counts[c.status] ?? 0) + 1;
    return STATUS_NAMES.map((s) => ({
      name: STATUS_LABELS[s],
      count: counts[s] ?? 0,
      fill: STATUS_COLORS[s],
    }));
  }, [data]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-1.5 text-sm font-medium">
          <GitMerge className="h-4 w-4 text-emerald-500" />
          客户生命周期漏斗
          <DemoBadge className="ml-auto" tooltip="漏斗数据基于随机生成的生命周期状态" />
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={funnelData}
              layout="vertical"
              margin={{ top: 5, right: 30, left: 50, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" horizontal={false} />
              <XAxis type="number" tick={{ fontSize: 11 }} />
              <YAxis
                type="category"
                dataKey="name"
                tick={{ fontSize: 12 }}
                width={45}
              />
              <Tooltip contentStyle={{ fontSize: 12 }} />
              <Bar dataKey="count" name="客户数" radius={[0, 6, 6, 0]} barSize={32}>
                {funnelData.map((entry, index) => (
                  <Cell key={index} fill={entry.fill} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
