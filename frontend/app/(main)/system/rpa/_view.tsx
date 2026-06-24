'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { genRPADemo, listRPAJobs, listRPARuns } from '@/lib/api/rpa';

function fmtTime(s?: string | null): string {
  if (!s) return '-';
  const d = new Date(s);
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

function fmtBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(2)} MB`;
}

function statusVariant(s?: string | null): 'default' | 'secondary' | 'destructive' | 'success' {
  if (s === 'success') return 'success';
  if (s === 'failed') return 'destructive';
  if (s === 'running') return 'default';
  return 'secondary';
}

export default function SystemRPAPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('task_scheduler:write');

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: jobs, isLoading: jobsLoading } = useQuery({
    queryKey: ['rpa-jobs'],
    queryFn: listRPAJobs,
  });
  const { data: runs } = useQuery({
    queryKey: ['rpa-runs'],
    queryFn: () => listRPARuns(50),
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genRPADemo();
      qc.invalidateQueries({ queryKey: ['rpa-jobs'] });
      qc.invalidateQueries({ queryKey: ['rpa-runs'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  const stats = (runs?.items ?? []).reduce(
    (acc, r) => {
      acc.total += 1;
      if (r.status === 'success') acc.success += 1;
      else if (r.status === 'failed') acc.failed += 1;
      else if (r.status === 'running') acc.running += 1;
      return acc;
    },
    { total: 0, success: 0, failed: 0, running: 0 },
  );

  // ── Status Pie data ──
  const statusPie = useMemo(() => [
    { name: '运行中', value: stats.running || jobs?.items.filter((j) => j.enabled && j.last_status === 'running').length || 0, fill: '#3b82f6' },
    { name: '空闲', value: (jobs?.items.length ?? 0) - (stats.running || 0) - (stats.failed || 0), fill: '#10b981' },
    { name: '异常', value: stats.failed, fill: '#ef4444' },
  ].filter((d) => d.value > 0), [stats, jobs]);

  // ── Result Bar data: success rate + duration distribution by job ──
  const resultBar = useMemo(() => {
    const runItems = runs?.items ?? [];
    const map = new Map<string, { job: string; success: number; failed: number; avgDuration: number; count: number }>();
    for (const r of runItems) {
      const entry = map.get(r.job_name) ?? { job: r.job_name, success: 0, failed: 0, avgDuration: 0, count: 0 };
      if (r.status === 'success') entry.success += 1;
      else entry.failed += 1;
      entry.avgDuration += r.duration_sec ?? 0;
      entry.count += 1;
      map.set(r.job_name, entry);
    }
    return Array.from(map.values()).map((e) => ({
      job: e.job,
      成功: e.success,
      失败: e.failed,
      平均耗时: e.count > 0 ? Math.round(e.avgDuration / e.count) : 0,
    }));
  }, [runs]);

  // ── Work log timeline (last 20 runs) ──
  const timeline = useMemo(() => {
    return (runs?.items ?? []).slice(0, 20).map((r) => ({
      id: r.id,
      job: r.job_name,
      time: fmtTime(r.started_at),
      status: r.status,
      duration: r.duration_sec,
      error: r.error,
    }));
  }, [runs]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">RPA 监控</h1>
          <p className="text-sm text-muted-foreground">
            外部 RPA 流程的运行健康度（区别于进程内调度任务）
          </p>
        </div>
        {canWrite && (
          <Button variant="outline" onClick={onGen} disabled={busy}>
            {busy ? '生成中...' : '生成演示数据'}
          </Button>
        )}
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">总运行次数</p>
            <p className="mt-1 text-2xl font-bold">{stats.total}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">成功</p>
            <p className="mt-1 text-2xl font-bold text-emerald-600">{stats.success}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">失败</p>
            <p className="mt-1 text-2xl font-bold text-destructive">{stats.failed}</p>
          </CardContent>
        </Card>
      </div>

      {/* ═══════════ Status Pie + Result Bar ═══════════ */}
      <div className="grid gap-4 lg:grid-cols-2">
        <ChartContainer title="RPA 运行状态实时面板">
          {statusPie.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无数据</p>
          ) : (
            <ResponsiveContainer width="100%" height={280}>
              <PieChart>
                <Pie
                  data={statusPie}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={90}
                  dataKey="value"
                  nameKey="name"
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  isAnimationActive={false}
                >
                  {statusPie.map((entry, i) => (
                    <Cell key={i} fill={entry.fill} />
                  ))}
                </Pie>
                <Tooltip contentStyle={{ fontSize: 12 }} />
                <Legend wrapperStyle={{ fontSize: 12 }} />
              </PieChart>
            </ResponsiveContainer>
          )}
        </ChartContainer>

        <ChartContainer title="执行结果统计（成功/失败 & 平均耗时）">
          {resultBar.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无数据</p>
          ) : (
            <ResponsiveContainer width="100%" height={280}>
              <BarChart data={resultBar} margin={{ top: 4, right: 12, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis dataKey="job" tick={{ fontSize: 10, fill: '#6b7280' }} />
                <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} />
                <Tooltip contentStyle={{ fontSize: 12 }} />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Bar dataKey="成功" stackId="a" fill="#10b981" isAnimationActive={false} />
                <Bar dataKey="失败" stackId="a" fill="#ef4444" radius={[4, 4, 0, 0]} isAnimationActive={false} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </ChartContainer>
      </div>

      {/* ═══════════ Work Log Timeline ═══════════ */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">工作日志时间线（最近 20 条）</CardTitle>
        </CardHeader>
        <CardContent>
          {timeline.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无运行记录</p>
          ) : (
            <div className="relative pl-6 space-y-0">
              {/* vertical line */}
              <div className="absolute left-2.5 top-1 bottom-1 w-0.5 bg-border" />
              {timeline.map((t) => (
                <div key={t.id} className="relative pb-4">
                  {/* dot */}
                  <div
                    className={`absolute -left-[13px] top-1.5 h-3 w-3 rounded-full border-2 border-background ${
                      t.status === 'success'
                        ? 'bg-emerald-500'
                        : t.status === 'failed'
                          ? 'bg-red-500'
                          : 'bg-blue-500'
                    }`}
                  />
                  <div className="rounded-md border p-2">
                    <div className="flex items-center gap-2 text-sm">
                      <span className="font-medium">{t.job}</span>
                      <Badge variant={statusVariant(t.status)} className="text-xs">
                        {t.status}
                      </Badge>
                      <span className="ml-auto text-xs text-muted-foreground">{t.time}</span>
                    </div>
                    <div className="mt-0.5 text-xs text-muted-foreground">
                      耗时: {t.duration ? `${t.duration}s` : '-'}
                      {t.error && <span className="ml-2 text-destructive">错误: {t.error}</span>}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">RPA 任务清单</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>任务名</TableHead>
                  <TableHead>说明</TableHead>
                  <TableHead>计划</TableHead>
                  <TableHead>上次执行</TableHead>
                  <TableHead>上次结果</TableHead>
                  <TableHead>状态</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {jobsLoading && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-muted-foreground">
                      加载中...
                    </TableCell>
                  </TableRow>
                )}
                {jobs?.items.map((j) => (
                  <TableRow key={j.id}>
                    <TableCell className="font-medium">{j.name}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {j.description ?? '-'}
                    </TableCell>
                    <TableCell>
                      <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
                        {j.schedule ?? '-'}
                      </code>
                    </TableCell>
                    <TableCell>{fmtTime(j.last_run_at)}</TableCell>
                    <TableCell>
                      {j.last_status ? (
                        <Badge variant={statusVariant(j.last_status)}>{j.last_status}</Badge>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    <TableCell>
                      {j.enabled ? (
                        <Badge variant="success">启用</Badge>
                      ) : (
                        <Badge variant="secondary">已停用</Badge>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                {jobs?.items.length === 0 && !jobsLoading && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-muted-foreground">
                      暂无 RPA 任务{canWrite && '，可点右上「生成演示数据」'}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">最近 50 次运行记录</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>任务</TableHead>
                  <TableHead>开始</TableHead>
                  <TableHead>结束</TableHead>
                  <TableHead className="text-right">耗时（秒）</TableHead>
                  <TableHead className="text-right">产出文件</TableHead>
                  <TableHead className="text-right">数据量</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>错误</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs?.items.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell className="font-medium">{r.job_name}</TableCell>
                    <TableCell>{fmtTime(r.started_at)}</TableCell>
                    <TableCell>{fmtTime(r.finished_at)}</TableCell>
                    <TableCell className="text-right">{r.duration_sec ?? '-'}</TableCell>
                    <TableCell className="text-right">{r.output_files}</TableCell>
                    <TableCell className="text-right">
                      {r.output_bytes > 0 ? fmtBytes(r.output_bytes) : '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={statusVariant(r.status)}>{r.status}</Badge>
                    </TableCell>
                    <TableCell className="text-xs text-destructive">{r.error ?? ''}</TableCell>
                  </TableRow>
                ))}
                {runs?.items.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center text-muted-foreground">
                      暂无运行记录
                    </TableCell>
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
