'use client';

// 客户历史电量档案：客户逐月电量（来源：文档解析「确认入库」→ 客户电量档案）。
import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ChartContainer } from '@/components/charts/chart-container';
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { listCustomers } from '@/lib/api/customers';
import { listCustomerEnergy } from '@/lib/api/customer-energy';
import { Zap, Loader2 } from 'lucide-react';

const SELECT_CLASS = 'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

function fmtNum(n: number | null): string {
  if (n === null || n === undefined) return '-';
  return n.toLocaleString('zh-CN', { maximumFractionDigits: 2 });
}

export default function CustomerEnergyPage() {
  const [customerId, setCustomerId] = useState('');

  const { data: customers } = useQuery({
    queryKey: ['customers', 'energy-picker'],
    queryFn: () => listCustomers({ limit: 200 }),
  });

  const { data: rows, isLoading } = useQuery({
    queryKey: ['customer-energy', customerId],
    queryFn: () => listCustomerEnergy({ customer_id: customerId || undefined }),
  });

  // 选定单个客户时，按月份升序展示电量趋势
  const chartData = useMemo(() => {
    if (!customerId || !rows) return [];
    return [...rows]
      .sort((a, b) => a.month.localeCompare(b.month))
      .map((r) => ({ month: r.month, energy: r.monthly_energy }));
  }, [rows, customerId]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">客户历史电量档案</h1>
        <p className="text-sm text-muted-foreground">
          客户逐月电量台账 —— 数据来自文档解析「确认入库 → 客户电量档案」（市场化账单/月度电量）
        </p>
      </div>

      {/* 客户筛选 */}
      <Card>
        <CardContent className="flex flex-wrap items-center gap-3 pt-6">
          <label className="text-sm text-muted-foreground">客户</label>
          <select
            value={customerId}
            onChange={(e) => setCustomerId(e.target.value)}
            className={SELECT_CLASS}
          >
            <option value="">全部客户</option>
            {(customers?.items ?? []).map((c) => (
              <option key={c.id} value={c.id}>
                {c.user_name}
              </option>
            ))}
          </select>
        </CardContent>
      </Card>

      {/* 单客户电量趋势 */}
      {customerId && chartData.length > 0 && (
        <ChartContainer title="月度电量趋势">
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis dataKey="month" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} />
              <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number) => fmtNum(v)} />
              <Bar dataKey="energy" name="月度电量" fill="#6366f1" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartContainer>
      )}

      {/* 电量明细 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Zap className="h-4 w-4" />
            电量明细
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>客户</TableHead>
                <TableHead>月份</TableHead>
                <TableHead className="text-right">月度电量</TableHead>
                <TableHead className="text-right">日均电量</TableHead>
                <TableHead>更新时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground">
                    <Loader2 className="mx-auto h-4 w-4 animate-spin" />
                  </TableCell>
                </TableRow>
              ) : (rows ?? []).length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground">
                    暂无电量数据，请在文档解析中把市场化账单/月度电量「确认入库 → 客户电量档案」
                  </TableCell>
                </TableRow>
              ) : (
                (rows ?? []).map((r) => (
                  <TableRow key={r.id}>
                    <TableCell>{r.customer_name}</TableCell>
                    <TableCell>{r.month}</TableCell>
                    <TableCell className="text-right">{fmtNum(r.monthly_energy)}</TableCell>
                    <TableCell className="text-right">{fmtNum(r.avg_daily_energy)}</TableCell>
                    <TableCell>{r.updated_at?.slice(0, 10)}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
