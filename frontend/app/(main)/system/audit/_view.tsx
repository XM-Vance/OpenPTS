'use client';

import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { listAuditLogs, type AuditLog } from '@/lib/api/audit';

const SELECT_CLASS =
  'flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm';

const METHOD_OPTIONS = [
  { value: '', label: '全部方法' },
  { value: 'POST', label: 'POST' },
  { value: 'PUT', label: 'PUT' },
  { value: 'DELETE', label: 'DELETE' },
];

const RESOURCE_OPTIONS = [
  { value: '', label: '全部资源' },
  { value: 'auth', label: '登录认证' },
  { value: 'users', label: '用户' },
  { value: 'roles', label: '角色' },
  { value: 'customers', label: '客户档案' },
  { value: 'retail_contracts', label: '零售合同' },
  { value: 'retail_packages', label: '零售套餐' },
  { value: 'load', label: '负荷' },
  { value: 'price', label: '价格' },
  { value: 'settlement', label: '结算' },
  { value: 'freq', label: '调频' },
  { value: 'storage', label: '储能' },
  { value: 'analytics', label: '客户分析' },
  { value: 'scheduler', label: '任务调度' },
];

const PIE_COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#84cc16'];

function fmtTime(s: string): string {
  const d = new Date(s);
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(
    d.getHours(),
  )}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

function methodVariant(m: string): 'default' | 'secondary' | 'destructive' | 'success' {
  if (m === 'POST') return 'success';
  if (m === 'PUT') return 'default';
  if (m === 'DELETE') return 'destructive';
  return 'secondary';
}

function statusVariant(code: number): 'default' | 'secondary' | 'destructive' | 'success' {
  if (code >= 200 && code < 300) return 'success';
  if (code >= 400 && code < 500) return 'default';
  if (code >= 500) return 'destructive';
  return 'secondary';
}

/** 判断是否为高风险操作 */
function isHighRisk(a: AuditLog): boolean {
  if (a.method === 'DELETE') return true;
  if (a.status_code >= 500) return true;
  if (a.resource === 'auth' && a.status_code >= 400) return true;
  return false;
}

export default function SystemAuditPage() {
  const [username, setUsername] = useState('');
  const [method, setMethod] = useState('');
  const [resource, setResource] = useState('');
  const [days, setDays] = useState(7);
  const [applied, setApplied] = useState({ username: '', method: '', resource: '', days: 7 });
  const PAGE_SIZE = 50;
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', applied, page],
    queryFn: () =>
      listAuditLogs({
        username: applied.username || undefined,
        method: applied.method || undefined,
        resource: applied.resource || undefined,
        days: applied.days,
        limit: PAGE_SIZE,
        offset: (page - 1) * PAGE_SIZE,
      }),
  });

  const items = useMemo(() => data?.items ?? [], [data]);
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  /* ---------- 饼图数据：按操作类型 ---------- */
  const pieData = useMemo(() => {
    const map = new Map<string, number>();
    for (const a of items) {
      map.set(a.method, (map.get(a.method) ?? 0) + 1);
    }
    return Array.from(map.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value);
  }, [items]);

  /* ---------- 柱状图数据：按用户 TOP10 ---------- */
  const barData = useMemo(() => {
    const map = new Map<string, number>();
    for (const a of items) {
      const u = a.username ?? '(匿名)';
      map.set(u, (map.get(u) ?? 0) + 1);
    }
    return Array.from(map.entries())
      .sort((a, b) => b[1] - a[1])
      .slice(0, 10)
      .map(([name, count]) => ({ name, count }));
  }, [items]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">操作审计</h1>
        <p className="text-sm text-muted-foreground">
          所有写操作（POST/PUT/DELETE）由中间件自动落库，可按用户、方法、资源、时间筛选。
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">筛选</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-5">
            <div className="space-y-1">
              <Label>用户名</Label>
              <Input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="模糊匹配"
              />
            </div>
            <div className="space-y-1">
              <Label>方法</Label>
              <select
                value={method}
                onChange={(e) => setMethod(e.target.value)}
                className={SELECT_CLASS}
              >
                {METHOD_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <Label>资源</Label>
              <select
                value={resource}
                onChange={(e) => setResource(e.target.value)}
                className={SELECT_CLASS}
              >
                {RESOURCE_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <Label>近 N 日</Label>
              <Input
                type="number"
                value={days}
                onChange={(e) => setDays(Number(e.target.value) || 7)}
                min={1}
                max={90}
              />
            </div>
            <div className="flex items-end gap-2">
              <Button
                onClick={() => {
                  setPage(1);
                  setApplied({ username, method, resource, days });
                }}
              >
                查询
              </Button>
              <Button
                variant="outline"
                onClick={() => {
                  setUsername('');
                  setMethod('');
                  setResource('');
                  setDays(7);
                  setPage(1);
                  setApplied({ username: '', method: '', resource: '', days: 7 });
                }}
              >
                重置
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* ========== 可视化图表 ========== */}
      {items.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2">
          {/* 饼图：按操作类型 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">操作类型分布(当前页)</CardTitle>
            </CardHeader>
            <CardContent>
              <div style={{ width: '100%', height: 260 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      outerRadius={90}
                      dataKey="value"
                      label={({ name, percent }) =>
                        `${name} ${(percent * 100).toFixed(0)}%`
                      }
                      isAnimationActive={false}
                    >
                      {pieData.map((_, idx) => (
                        <Cell
                          key={`cell-${idx}`}
                          fill={PIE_COLORS[idx % PIE_COLORS.length]}
                        />
                      ))}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* 柱状图：按用户 TOP10 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">操作用户 TOP10(当前页)</CardTitle>
            </CardHeader>
            <CardContent>
              <div style={{ width: '100%', height: 260 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={barData} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis dataKey="name" tick={{ fontSize: 11, fill: '#6b7280' }} />
                    <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={40} />
                    <Tooltip contentStyle={{ fontSize: 12 }} />
                    <Bar dataKey="count" fill="#6366f1" radius={[4, 4, 0, 0]} isAnimationActive={false} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* ========== 审计表格 ========== */}
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>时间</TableHead>
              <TableHead>用户</TableHead>
              <TableHead>方法</TableHead>
              <TableHead>路径</TableHead>
              <TableHead>资源</TableHead>
              <TableHead>状态码</TableHead>
              <TableHead className="text-right">耗时 (ms)</TableHead>
              <TableHead>IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((a) => {
              const highRisk = isHighRisk(a);
              return (
                <TableRow
                  key={a.id}
                  className={highRisk ? 'bg-red-50 dark:bg-red-950/30' : ''}
                >
                  <TableCell className="whitespace-nowrap">{fmtTime(a.created_at)}</TableCell>
                  <TableCell>{a.username ?? '(匿名)'}</TableCell>
                  <TableCell>
                    <Badge variant={methodVariant(a.method)}>{a.method}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{a.path}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {a.resource ?? '-'}
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusVariant(a.status_code)}>{a.status_code}</Badge>
                  </TableCell>
                  <TableCell className="text-right">{a.duration_ms}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {a.ip ?? '-'}
                  </TableCell>
                </TableRow>
              );
            })}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  暂无审计记录
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* 分页:total 来自后端同条件 COUNT,翻页保留筛选 */}
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>
          共 {total} 条 · 第 {page} / {totalPages} 页
        </span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1 || isLoading}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            上一页
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages || isLoading}
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          >
            下一页
          </Button>
        </div>
      </div>
    </div>
  );
}
