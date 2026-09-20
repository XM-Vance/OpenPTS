'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  PieChart,
  Pie,
  Cell,
  Area,
  AreaChart,
  BarChart,
  Bar,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { getSecurityOverview } from '@/lib/api/security';
import { EmptyState } from '@/components/feedback';
import { Skeleton } from '@/components/ui/skeleton';

const HOURS_OPTIONS = [
  { value: 1, label: '近 1 时' },
  { value: 24, label: '近 24 时' },
  { value: 72, label: '近 3 日' },
  { value: 168, label: '近 7 日' },
];

function fmtTime(s: string): string {
  const d = new Date(s);
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

/* ---------- 半圆仪表盘组件 ---------- */
function GaugeChart({ value, label }: { value: number; label: string }) {
  // value 0-100, 模拟半圆 gauge
  const clamped = Math.min(100, Math.max(0, value));
  const angle = (clamped / 100) * 180;

  // 用 PieChart 模拟半圆：两段（已用 + 剩余）
  const gaugeData = [
    { name: '健康度', value: clamped },
    { name: '剩余', value: 100 - clamped },
  ];

  let color = '#10b981'; // 绿色
  if (clamped < 60) color = '#ef4444'; // 红色
  else if (clamped < 80) color = '#f59e0b'; // 黄色

  const COLORS = [color, '#e5e7eb'];

  return (
    <div className="flex flex-col items-center">
      <div style={{ width: 220, height: 130, overflow: 'hidden' }}>
        <ResponsiveContainer width="100%" height={260}>
          <PieChart>
            <Pie
              data={gaugeData}
              cx="50%"
              cy="100%"
              startAngle={180}
              endAngle={0}
              innerRadius={70}
              outerRadius={110}
              dataKey="value"
              stroke="none"
              isAnimationActive={false}
            >
              {gaugeData.map((_, idx) => (
                <Cell key={`gauge-${idx}`} fill={COLORS[idx]} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
      </div>
      <div className="-mt-8 text-center">
        <span className="text-3xl font-bold" style={{ color }}>
          {clamped}
        </span>
        <span className="text-sm text-muted-foreground"> / 100</span>
        <p className="text-sm text-muted-foreground mt-1">{label}</p>
      </div>
    </div>
  );
}

/* ---------- 异常登录检测：后端暂无数据源，不展示伪造记录 ---------- */

export default function SecurityPage() {
  const [hours, setHours] = useState(24);

  const { data, isLoading } = useQuery({
    queryKey: ['security-overview', hours],
    queryFn: () => getSecurityOverview(hours),
    refetchInterval: 30_000,
  });

  const chartData = (data?.error_hourly ?? []).map((p) => ({
    hour: fmtTime(p.bucket),
    count: p.count,
  }));

  /* 计算系统健康度（无数据时显示加载态，不编造分数） */
  const healthScore = (() => {
    if (!data) return null;
    const total = data.total || 1;
    const errorRate = ((data.errors_4xx + data.errors_5xx) / total) * 100;
    const deleteRate = (data.delete_ops / total) * 100;
    let score = 100 - errorRate * 2 - deleteRate * 3;
    if (data.failed_sched_jobs > 0) score -= data.failed_sched_jobs * 5;
    return Math.round(Math.min(100, Math.max(0, score)));
  })();

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">安全大屏</h1>
          <p className="text-sm text-muted-foreground">
            异常聚合监控：4xx/5xx 错误、敏感操作、用户/IP 排行
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {HOURS_OPTIONS.map((o) => (
            <Button
              key={o.value}
              size="sm"
              variant={hours === o.value ? 'default' : 'outline'}
              onClick={() => setHours(o.value)}
            >
              {o.label}
            </Button>
          ))}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4 lg:grid-cols-7">
        {[
          { label: '请求总数', value: data?.total ?? 0, tone: 'text-foreground' },
          { label: '4xx 错误', value: data?.errors_4xx ?? 0, tone: 'text-amber-700' },
          { label: '5xx 错误', value: data?.errors_5xx ?? 0, tone: 'text-destructive' },
          { label: 'DELETE', value: data?.delete_ops ?? 0, tone: 'text-orange-600' },
          { label: '活跃用户', value: data?.unique_users ?? 0, tone: 'text-blue-600' },
          { label: '活跃 IP', value: data?.unique_ips ?? 0, tone: 'text-blue-600' },
          { label: '失败任务', value: data?.failed_sched_jobs ?? 0, tone: 'text-destructive' },
        ].map((k) => (
          <Card key={k.label}>
            <CardContent className="pt-6">
              <p className="text-xs text-muted-foreground">{k.label}</p>
              <p className={`mt-1 text-2xl font-bold ${k.tone}`}>{k.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* ========== 系统健康度 + 异常登录 ========== */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* 系统健康度仪表盘 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">系统健康度</CardTitle>
          </CardHeader>
          <CardContent className="flex justify-center">
            {healthScore == null ? (
              <Skeleton className="h-5 w-full" />
            ) : (
              <GaugeChart value={healthScore} label="综合健康评分" />
            )}
          </CardContent>
        </Card>

        {/* 异常登录检测面板：后端暂无数据源，展示空态而非伪造记录 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">异常登录检测</CardTitle>
          </CardHeader>
          <CardContent>
            <EmptyState compact title="暂无异常登录检测数据（该功能依赖登录审计数据源，建设中）。" />
          </CardContent>
        </Card>
      </div>

      {/* ========== 每小时错误数趋势（真实数据） ========== */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">每小时错误数趋势</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-5 w-full" />
          ) : chartData.length === 0 ? (
            <EmptyState compact title="暂无错误（🎉）" />
          ) : (
            <div
              className="[&_.recharts-surface:focus]:outline-none"
              style={{ width: '100%', height: 200 }}
            >
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="hour" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 12 }} width={40} />
                  <Tooltip contentStyle={{ fontSize: 12 }} />
                  <Area
                    type="monotone"
                    dataKey="count"
                    stroke="#ef4444"
                    fill="#fecaca"
                    strokeWidth={2}
                    isAnimationActive={false}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          )}
        </CardContent>
      </Card>

      {/* ========== 失败请求 & 活跃 IP 排行（保留） ========== */}
      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">失败请求 Top 5 用户</CardTitle>
          </CardHeader>
          <CardContent>
            {(data?.top_failed_users ?? []).length === 0 ? (
              <p className="text-sm text-muted-foreground">无失败请求 🎉</p>
            ) : (
              <ul className="space-y-2">
                {(data?.top_failed_users ?? []).map((u) => (
                  <li key={u.user} className="flex items-center justify-between text-sm">
                    <span>{u.user}</span>
                    <Badge variant="destructive">{u.count}</Badge>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">活跃 IP Top 5</CardTitle>
          </CardHeader>
          <CardContent>
            {(data?.top_active_ips ?? []).length === 0 ? (
              <EmptyState compact title="暂无" />
            ) : (
              <ul className="space-y-2">
                {(data?.top_active_ips ?? []).map((i) => (
                  <li key={i.ip} className="flex items-center justify-between text-sm">
                    <code className="text-xs">{i.ip}</code>
                    <Badge variant="outline">{i.count}</Badge>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>

      {/* ========== 最近 DELETE 操作（保留） ========== */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">最近 DELETE 操作（敏感）</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead>用户</TableHead>
                  <TableHead>路径</TableHead>
                  <TableHead>资源</TableHead>
                  <TableHead>状态</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(data?.recent_deletes ?? []).map((r, i) => (
                  <TableRow key={i}>
                    <TableCell className="whitespace-nowrap">{fmtTime(r.created_at)}</TableCell>
                    <TableCell>{r.username ?? '(匿名)'}</TableCell>
                    <TableCell className="font-mono text-xs">{r.path}</TableCell>
                    <TableCell className="text-muted-foreground">{r.resource ?? '-'}</TableCell>
                    <TableCell>
                      <Badge
                        variant={r.status_code < 300 ? 'success' : r.status_code < 500 ? 'default' : 'destructive'}
                      >
                        {r.status_code}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
                {(data?.recent_deletes ?? []).length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5}><EmptyState compact title="暂无 DELETE 操作" /></TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
