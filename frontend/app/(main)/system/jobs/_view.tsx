'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
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
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  listJobs,
  listRuns,
  setJobEnabled,
  triggerJob,
  type ScheduledJob,
} from '@/lib/api/scheduler';
import { AlertTriangle } from 'lucide-react';

const STATUS_LABEL: Record<string, string> = {
  success: '成功',
  failed: '失败',
  running: '运行中',
};

function fmtTime(s?: string | null): string {
  if (!s) return '-';
  const d = new Date(s);
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours(),
  )}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

function pad(n: number): string {
  return n < 10 ? `0${n}` : String(n);
}

function statusVariant(s?: string | null): 'default' | 'secondary' | 'destructive' | 'success' {
  if (s === 'success') return 'success';
  if (s === 'failed') return 'destructive';
  if (s === 'running') return 'default';
  return 'secondary';
}

export default function SystemJobsPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('task_scheduler:write');

  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const { data: jobs, isLoading: jobsLoading } = useQuery({
    queryKey: ['scheduler-jobs'],
    queryFn: listJobs,
    refetchInterval: 10000,
  });
  const { data: runs } = useQuery({
    queryKey: ['scheduler-runs'],
    queryFn: () => listRuns(30),
    refetchInterval: 5000,
  });

  const onTrigger = async (j: ScheduledJob) => {
    setError(null);
    setNotice(null);
    setBusyId(j.id);
    try {
      await triggerJob(j.id);
      setNotice(`已触发「${j.name}」，请稍后查看执行记录`);
      qc.invalidateQueries({ queryKey: ['scheduler-runs'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusyId(null);
    }
  };

  const onToggle = async (j: ScheduledJob) => {
    setError(null);
    setBusyId(j.id);
    try {
      await setJobEnabled(j.id, !j.enabled);
      qc.invalidateQueries({ queryKey: ['scheduler-jobs'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusyId(null);
    }
  };

  // ── Gantt chart data: horizontal bars for each job's execution timeline ──
  const ganttData = useMemo(() => {
    const runItems = runs?.items ?? [];
    // group runs by job_name, take last run
    const map = new Map<string, { job_name: string; start: number; duration: number; status: string }>();
    for (const r of runItems) {
      const existing = map.get(r.job_name);
      const start = new Date(r.started_at).getTime();
      const end = r.finished_at ? new Date(r.finished_at).getTime() : Date.now();
      const duration = Math.max(0, end - start);
      if (!existing || start > existing.start) {
        map.set(r.job_name, { job_name: r.job_name, start, duration, status: r.status });
      }
    }
    // also add jobs that have no recent runs
    for (const j of jobs?.items ?? []) {
      if (!map.has(j.name)) {
        const start = j.last_run_at ? new Date(j.last_run_at).getTime() : Date.now();
        map.set(j.name, { job_name: j.name, start, duration: 0, status: j.last_status ?? 'idle' });
      }
    }
    return Array.from(map.values()).sort((a, b) => a.start - b.start);
  }, [runs, jobs]);

  const minGanttStart = ganttData.length > 0 ? Math.min(...ganttData.map((d) => d.start)) : 0;

  // ── Success rate trend (by date) ──
  const successTrend = useMemo(() => {
    const runItems = runs?.items ?? [];
    const byDate = new Map<string, { total: number; success: number }>();
    for (const r of runItems) {
      const date = r.started_at.slice(0, 10);
      const entry = byDate.get(date) ?? { total: 0, success: 0 };
      entry.total += 1;
      if (r.status === 'success') entry.success += 1;
      byDate.set(date, entry);
    }
    return Array.from(byDate.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([date, { total, success }]) => ({
        date: date.slice(5),
        rate: total > 0 ? Math.round((success / total) * 100) : 0,
        total,
      }));
  }, [runs]);

  // ── Failed tasks list ──
  const failedRuns = useMemo(
    () => (runs?.items ?? []).filter((r) => r.status === 'failed'),
    [runs],
  );

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">调度任务</h1>
        <p className="text-sm text-muted-foreground">
          进程内 cron 调度器，按表达式自动触发；支持手工触发与启停。
        </p>
      </div>

      {notice && (
        <Alert>
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* ═══════════ Gantt Chart ═══════════ */}
      <ChartContainer title="任务执行甘特图（最近执行时间线）">
        {ganttData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无执行数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={Math.max(200, ganttData.length * 36 + 40)}>
            <BarChart data={ganttData} layout="vertical" margin={{ top: 4, right: 12, left: 10, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis
                type="number"
                tick={{ fontSize: 11, fill: '#6b7280' }}
                tickFormatter={(v: number) => {
                  const d = new Date(v);
                  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
                }}
              />
              <YAxis dataKey="job_name" type="category" tick={{ fontSize: 11, fill: '#6b7280' }} width={120} />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) => {
                  if (name === '偏移') return '';
                  return `${(v / 1000).toFixed(1)}s`;
                }}
                labelFormatter={(l: string) => l}
              />
              <Bar dataKey="start" name="偏移" fill="transparent" stackId="a" isAnimationActive={false} />
              <Bar dataKey="duration" name="耗时" stackId="a" radius={[0, 4, 4, 0]} isAnimationActive={false}>
                {ganttData.map((d, i) => (
                  <Cell
                    key={i}
                    fill={d.status === 'success' ? '#10b981' : d.status === 'failed' ? '#ef4444' : '#3b82f6'}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* ═══════════ Success Rate Trend ═══════════ */}
      <ChartContainer title="执行成功率趋势">
        {successTrend.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={240}>
            <LineChart data={successTrend} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="%" domain={[0, 100]} />
              <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number, name: string) => name === '成功率' ? `${v}%` : v} />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line type="monotone" dataKey="rate" name="成功率" stroke="#10b981" strokeWidth={2} dot={{ r: 3 }} isAnimationActive={false} />
            </LineChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* ═══════════ Failed Tasks Alert List ═══════════ */}
      {failedRuns.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <AlertTriangle className="h-4 w-4 text-destructive" />
              失败任务告警
              <Badge variant="destructive">{failedRuns.length}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>任务</TableHead>
                    <TableHead>开始时间</TableHead>
                    <TableHead>耗时(ms)</TableHead>
                    <TableHead>错误</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {failedRuns.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-medium">{r.job_name}</TableCell>
                      <TableCell>{fmtTime(r.started_at)}</TableCell>
                      <TableCell>{r.duration_ms ?? '-'}</TableCell>
                      <TableCell className="text-xs text-destructive max-w-xs truncate">
                        {r.error ?? '未知错误'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">任务列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>任务名</TableHead>
                  <TableHead>说明</TableHead>
                  <TableHead>Cron</TableHead>
                  <TableHead>上次执行</TableHead>
                  <TableHead>上次结果</TableHead>
                  <TableHead>下次执行</TableHead>
                  <TableHead>状态</TableHead>
                  {canWrite && <TableHead className="text-right">操作</TableHead>}
                </TableRow>
              </TableHeader>
              <TableBody>
                {jobsLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={canWrite ? 8 : 7}
                      className="text-center text-muted-foreground"
                    >
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
                        {j.cron_expr}
                      </code>
                    </TableCell>
                    <TableCell>{fmtTime(j.last_run_at)}</TableCell>
                    <TableCell>
                      {j.last_status ? (
                        <Badge variant={statusVariant(j.last_status)}>
                          {STATUS_LABEL[j.last_status] ?? j.last_status}
                        </Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell>{fmtTime(j.next_run_at)}</TableCell>
                    <TableCell>
                      {j.enabled ? (
                        <Badge variant="success">启用</Badge>
                      ) : (
                        <Badge variant="secondary">已停用</Badge>
                      )}
                    </TableCell>
                    {canWrite && (
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button
                            size="sm"
                            variant="ghost"
                            disabled={busyId === j.id || !j.enabled}
                            onClick={() => onTrigger(j)}
                          >
                            立即触发
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            disabled={busyId === j.id}
                            onClick={() => onToggle(j)}
                          >
                            {j.enabled ? '停用' : '启用'}
                          </Button>
                        </div>
                      </TableCell>
                    )}
                  </TableRow>
                ))}
                {jobs?.items.length === 0 && !jobsLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={canWrite ? 8 : 7}
                      className="text-center text-muted-foreground"
                    >
                      暂无调度任务
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
          <CardTitle className="text-base">最近执行记录（30 条）</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>任务</TableHead>
                  <TableHead>开始时间</TableHead>
                  <TableHead>结束时间</TableHead>
                  <TableHead className="text-right">耗时 (ms)</TableHead>
                  <TableHead>触发方式</TableHead>
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
                    <TableCell className="text-right">
                      {r.duration_ms ?? '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={r.trigger === 'manual' ? 'default' : 'outline'}>
                        {r.trigger === 'manual' ? '手工' : 'Cron'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={statusVariant(r.status)}>
                        {STATUS_LABEL[r.status] ?? r.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-xs text-destructive">
                      {r.error ?? ''}
                    </TableCell>
                  </TableRow>
                ))}
                {runs?.items.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground">
                      暂无执行记录
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
