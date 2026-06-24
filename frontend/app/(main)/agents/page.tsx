'use client';

import { useState, useCallback, useMemo } from 'react';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createAgent,
  deleteAgent,
  listAgents,
  updateAgent,
  type Agent,
} from '@/lib/api/agents';

const STATUS_MAP: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' | 'destructive' }> = {
  active: { label: '启用', variant: 'default' },
  inactive: { label: '停用', variant: 'secondary' },
};

// 区域配色
const REGION_COLORS = ['#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#84cc16'];

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const AgentTop10Bar = dynamic(
  () => import('./_charts').then((m) => ({ default: m.AgentTop10Bar })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);
const AgentRegionPie = dynamic(
  () => import('./_charts').then((m) => ({ default: m.AgentRegionPie })),
  { ssr: false, loading: () => <div className="h-[240px] w-[60%]" /> },
);

export default function AgentsPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('customer_management:write');
  const canDelete = has('customer_management:delete');

  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = Number(searchParams.get('page') ?? '1') || 1;
  const pageSize = Number(searchParams.get('pageSize') ?? '20') || 20;
  const search = searchParams.get('search') ?? '';
  const statusFilter = searchParams.get('status') ?? '';

  const [editing, setEditing] = useState<Agent | 'new' | null>(null);
  const [keyword, setKeyword] = useState(search);

  const setParams = useCallback(
    (updates: Record<string, string | number>) => {
      const sp = new URLSearchParams(searchParams.toString());
      for (const [k, v] of Object.entries(updates)) {
        if (v === '' || v === 0) {
          sp.delete(k);
        } else {
          sp.set(k, String(v));
        }
      }
      router.push(`${pathname}?${sp.toString()}`);
    },
    [searchParams, router, pathname],
  );

  const { data, isLoading } = useQuery({
    queryKey: ['agents', search, statusFilter, page, pageSize],
    queryFn: () =>
      listAgents({
        keyword: search || undefined,
        status: statusFilter || undefined,
        page,
        page_size: pageSize,
      }),
  });

  // 全量数据用于统计（取前 200 条）
  const { data: allAgentsData } = useQuery({
    queryKey: ['agents-all'],
    queryFn: () => listAgents({ limit: 200 }),
  });

  const onDelete = async (id: string, name: string) => {
    if (!window.confirm(`确认删除代理商「${name}」？`)) return;
    try {
      await deleteAgent(id);
      qc.invalidateQueries({ queryKey: ['agents'] });
      qc.invalidateQueries({ queryKey: ['agents-all'] });
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  // ── TOP10 业绩排行（模拟：按佣金比例排序）──
  const top10Data = useMemo(() => {
    const allAgents = allAgentsData?.items ?? items;
    return [...allAgents]
      .sort((a, b) => b.commission_rate - a.commission_rate)
      .slice(0, 10)
      .map((a) => ({
        name: a.agent_name.length > 5 ? a.agent_name.slice(0, 5) + '…' : a.agent_name,
        rate: +(a.commission_rate * 100).toFixed(1),
        region: a.region || '未知',
      }));
  }, [allAgentsData, items]);

  // ── 佣金统计卡片 ──
  const commissionStats = useMemo(() => {
    const allAgents = allAgentsData?.items ?? items;
    const rates = allAgents.map((a) => a.commission_rate * 100);
    const avg = rates.length > 0 ? rates.reduce((s, r) => s + r, 0) / rates.length : 0;
    const max = rates.length > 0 ? Math.max(...rates) : 0;
    const min = rates.length > 0 ? Math.min(...rates) : 0;
    const activeCount = allAgents.filter((a) => a.status === 'active').length;
    return { avg, max, min, activeCount, total: allAgents.length };
  }, [allAgentsData, items]);

  // ── 区域分布饼图 ──
  const regionData = useMemo(() => {
    const allAgents = allAgentsData?.items ?? items;
    const map: Record<string, number> = {};
    for (const a of allAgents) {
      const r = a.region || '未知';
      map[r] = (map[r] || 0) + 1;
    }
    return Object.entries(map)
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value);
  }, [allAgentsData, items]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">代理商管理</h1>
        {canWrite && <Button onClick={() => setEditing('new')}>新建代理商</Button>}
      </div>

      {/* ── 佣金统计卡片 ── */}
      <div className="grid gap-4 md:grid-cols-5">
        <div className="rounded-lg border bg-card p-4">
          <p className="text-xs text-muted-foreground">代理商总数</p>
          <p className="mt-1 text-2xl font-bold">{commissionStats.total}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-xs text-muted-foreground">启用中</p>
          <p className="mt-1 text-2xl font-bold text-emerald-600">{commissionStats.activeCount}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-xs text-muted-foreground">平均佣金比例</p>
          <p className="mt-1 text-2xl font-bold text-blue-600">{commissionStats.avg.toFixed(1)}%</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-xs text-muted-foreground">最高佣金</p>
          <p className="mt-1 text-2xl font-bold text-amber-600">{commissionStats.max.toFixed(1)}%</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="text-xs text-muted-foreground">最低佣金</p>
          <p className="mt-1 text-2xl font-bold">{commissionStats.min.toFixed(1)}%</p>
        </div>
      </div>

      {/* ── 图表区 ── */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* TOP10 代理商业绩排行 */}
        <ChartContainer title="代理商业绩排行 TOP10（佣金比例）" minHeight={300}>
          {top10Data.length > 0 ? (
            <AgentTop10Bar data={top10Data} colors={REGION_COLORS} />
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无数据
            </div>
          )}
        </ChartContainer>

        {/* 客户区域分布 */}
        <ChartContainer title="代理商区域分布" minHeight={300}>
          {regionData.length > 0 ? (
            <div className="flex items-center justify-center gap-6">
              <AgentRegionPie data={regionData} colors={REGION_COLORS} />
              <div className="space-y-1.5 text-xs">
                {regionData.map((r, i) => (
                  <div key={r.name} className="flex items-center gap-2">
                    <span
                      className="inline-block h-3 w-3 rounded-sm"
                      style={{ backgroundColor: REGION_COLORS[i % REGION_COLORS.length] }}
                    />
                    <span className="text-zinc-600">{r.name}</span>
                    <span className="font-medium">{r.value}</span>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              暂无数据
            </div>
          )}
        </ChartContainer>
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="搜索代理商名称 / 联系人"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') setParams({ search: keyword, page: 1 });
          }}
          className="max-w-xs"
        />
        <Button
          variant="outline"
          onClick={() => setParams({ search: keyword, page: 1 })}
        >
          搜索
        </Button>
        <select
          value={statusFilter}
          onChange={(e) => setParams({ status: e.target.value === 'all' ? '' : e.target.value, page: 1 })}
          className="h-9 rounded-md border border-input bg-background px-3 text-sm"
        >
          <option value="all">全部状态</option>
          <option value="active">启用</option>
          <option value="inactive">停用</option>
        </select>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>代理商名称</TableHead>
              <TableHead>联系人</TableHead>
              <TableHead>联系电话</TableHead>
              <TableHead>区域</TableHead>
              <TableHead>佣金比例</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((agent) => {
              const st = STATUS_MAP[agent.status] ?? { label: agent.status, variant: 'outline' };
              return (
                <TableRow key={agent.id}>
                  <TableCell className="font-medium">{agent.agent_name}</TableCell>
                  <TableCell>{agent.contact_person || '-'}</TableCell>
                  <TableCell>{agent.phone || '-'}</TableCell>
                  <TableCell>{agent.region || '-'}</TableCell>
                  <TableCell>{(agent.commission_rate * 100).toFixed(1)}%</TableCell>
                  <TableCell>
                    <Badge variant={st.variant}>{st.label}</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      {canWrite && (
                        <Button size="sm" variant="ghost" onClick={() => setEditing(agent)}>
                          编辑
                        </Button>
                      )}
                      {canDelete && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => onDelete(agent.id, agent.agent_name)}
                        >
                          删除
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
            {total === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">
            共 {total} 条，第 {page}/{totalPages} 页
          </span>
          <div className="flex gap-1">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setParams({ page: page - 1 })}
            >
              上一页
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => setParams({ page: page + 1 })}
            >
              下一页
            </Button>
          </div>
        </div>
      )}

      {editing && (
        <AgentFormDialog
          agent={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            qc.invalidateQueries({ queryKey: ['agents'] });
            qc.invalidateQueries({ queryKey: ['agents-all'] });
          }}
        />
      )}
    </div>
  );
}

function AgentFormDialog({
  agent,
  onClose,
  onSaved,
}: {
  agent: Agent | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !agent;
  const [agentName, setAgentName] = useState(agent?.agent_name ?? '');
  const [contactPerson, setContactPerson] = useState(agent?.contact_person ?? '');
  const [phone, setPhone] = useState(agent?.phone ?? '');
  const [email, setEmail] = useState(agent?.email ?? '');
  const [region, setRegion] = useState(agent?.region ?? '');
  const [commissionRate, setCommissionRate] = useState(
    agent ? String(agent.commission_rate * 100) : '0',
  );
  const [status, setStatus] = useState(agent?.status ?? 'active');
  const [description, setDescription] = useState(agent?.description ?? '');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      const input = {
        agent_name: agentName,
        contact_person: contactPerson || undefined,
        phone: phone || undefined,
        email: email || undefined,
        region: region || undefined,
        commission_rate: parseFloat(commissionRate) / 100 || 0,
        status,
        description: description || undefined,
      };
      if (isNew) {
        await createAgent(input);
      } else if (agent) {
        await updateAgent(agent.id, input);
      }
      onSaved();
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      <DialogHeader>
        <DialogTitle>{isNew ? '新建代理商' : '编辑代理商'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>代理商名称</Label>
          <Input value={agentName} onChange={(e) => setAgentName(e.target.value)} />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>联系人</Label>
            <Input value={contactPerson} onChange={(e) => setContactPerson(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>联系电话</Label>
            <Input value={phone} onChange={(e) => setPhone(e.target.value)} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>电子邮箱</Label>
            <Input value={email} onChange={(e) => setEmail(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>服务区域</Label>
            <Input value={region} onChange={(e) => setRegion(e.target.value)} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>佣金比例（%）</Label>
            <Input
              type="number"
              min="0"
              max="100"
              step="0.1"
              value={commissionRate}
              onChange={(e) => setCommissionRate(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label>状态</Label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
            >
              <option value="active">启用</option>
              <option value="inactive">停用</option>
            </select>
          </div>
        </div>
        <div className="space-y-2">
          <Label>备注</Label>
          <Input value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        {error && <p className="text-xs text-destructive">{error}</p>}
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          取消
        </Button>
        <Button onClick={submit} disabled={submitting}>
          {submitting ? '保存中...' : '保存'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
