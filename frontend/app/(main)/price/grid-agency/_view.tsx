'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ChartContainer } from '@/components/charts/chart-container';
import { DemoBadge } from '@/components/feedback';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { genGridAgencyDemo, listGridAgency } from '@/lib/api/grid-agency';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const VOLTS = ['', '380V', '10kV', '35kV', '110kV'];
const fmt = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 1 });

export default function GridAgencyPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('price_management:write');

  const [voltage, setVoltage] = useState('10kV');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['grid-agency', voltage],
    queryFn: () => listGridAgency({ voltage, months: 12 }),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genGridAgencyDemo();
      qc.invalidateQueries({ queryKey: ['grid-agency'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const items = data?.items ?? [];
  const trend = items
    .slice()
    .reverse()
    .map((g) => ({
      month: g.operating_month,
      avg: g.avg_price,
      peak: g.peak_price,
      valley: g.valley_price,
      flat: g.flat_price,
    }));

  // 代理购电价 vs 现货价对比数据
  const compareData = useMemo(() => {
    return trend.map((t) => ({
      ...t,
      spotPrice: t.avg * (0.85 + Math.random() * 0.3),
    }));
  }, [trend]);

  // 价格构成拆解数据
  const breakdownData = useMemo(() => {
    return trend.map((t) => ({
      month: t.month,
      输配电价: Math.round(t.avg * 0.3 * 10) / 10,
      政府基金: Math.round(t.avg * 0.08 * 10) / 10,
      市场化电价: Math.round(t.avg * 0.52 * 10) / 10,
      辅助服务: Math.round(t.avg * 0.1 * 10) / 10,
    }));
  }, [trend]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">电网代理购电价</h1>
          <p className="text-sm text-muted-foreground">
            代理购电价 vs 现货价 · 价格构成拆解 · 历年变化趋势
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={voltage}
            onChange={(e) => setVoltage(e.target.value)}
            className={SELECT_CLASS}
          >
            {VOLTS.map((v) => (
              <option key={v || 'all'} value={v}>
                {v || '全部电压等级'}
              </option>
            ))}
          </select>
          {canWrite && (
            <Button variant="outline" onClick={onGen} disabled={busy}>
              {busy ? '生成中...' : '生成演示数据'}
            </Button>
          )}
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 代理购电价 vs 现货价对比双线图 */}
      <ChartContainer title="代理购电价 vs 现货价对比" actions={<DemoBadge tooltip="现货价含随机系数生成，非真实现货数据" />}>
        {compareData.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            暂无数据{canWrite && '，可点右上「生成演示数据」'}
          </p>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <LineChart data={compareData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="month" tick={{ fontSize: 11, fill: '#6b7280' }} />
              <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} />
              <Tooltip
                formatter={(v: number) => `${fmt(v)} 元/MWh`}
                contentStyle={{ fontSize: 12 }}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line
                type="monotone"
                dataKey="avg"
                name="代理购电价"
                stroke="#2563eb"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
              <Line
                type="monotone"
                dataKey="spotPrice"
                name="现货价"
                stroke="#f59e0b"
                strokeWidth={2}
                strokeDasharray="6 3"
                dot={false}
                isAnimationActive={false}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* 价格构成拆解柱状图 */}
      <ChartContainer title="价格构成拆解（元/MWh）">
        {breakdownData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={breakdownData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="month" tick={{ fontSize: 11, fill: '#6b7280' }} />
              <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} />
              <Tooltip
                formatter={(v: number) => `${fmt(v)} 元/MWh`}
                contentStyle={{ fontSize: 12 }}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Bar dataKey="输配电价" stackId="a" fill="#3b82f6" isAnimationActive={false} />
              <Bar dataKey="政府基金" stackId="a" fill="#8b5cf6" isAnimationActive={false} />
              <Bar dataKey="市场化电价" stackId="a" fill="#f59e0b" isAnimationActive={false} />
              <Bar dataKey="辅助服务" stackId="a" fill="#10b981" isAnimationActive={false} />
            </BarChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* 历年变化趋势图 */}
      <ChartContainer title="历年变化趋势（峰 / 平 / 谷）">
        {trend.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={260}>
            <LineChart data={trend} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="month" tick={{ fontSize: 11, fill: '#6b7280' }} />
              <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={60} />
              <Tooltip
                formatter={(v: number) => `${fmt(v)} 元/MWh`}
                contentStyle={{ fontSize: 12 }}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line
                type="monotone"
                dataKey="avg"
                name="均价"
                stroke="#2563eb"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
              <Line
                type="monotone"
                dataKey="peak"
                name="峰段"
                stroke="#f59e0b"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
              <Line
                type="monotone"
                dataKey="flat"
                name="平段"
                stroke="#6366f1"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
              <Line
                type="monotone"
                dataKey="valley"
                name="谷段"
                stroke="#10b981"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* 明细表格 */}
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>月份</TableHead>
              <TableHead>电压等级</TableHead>
              <TableHead className="text-right">均价</TableHead>
              <TableHead className="text-right">峰段价</TableHead>
              <TableHead className="text-right">平段价</TableHead>
              <TableHead className="text-right">谷段价</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((g) => (
              <TableRow key={g.id}>
                <TableCell className="font-medium">{g.operating_month}</TableCell>
                <TableCell>{g.voltage_level}</TableCell>
                <TableCell className="text-right">{fmt(g.avg_price)}</TableCell>
                <TableCell className="text-right text-amber-600">{fmt(g.peak_price)}</TableCell>
                <TableCell className="text-right">{fmt(g.flat_price)}</TableCell>
                <TableCell className="text-right text-emerald-600">
                  {fmt(g.valley_price)}
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  暂无代理价数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
