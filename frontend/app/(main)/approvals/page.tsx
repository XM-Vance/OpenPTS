'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { usePermission } from '@/lib/auth/use-permission';
import { useAuth } from '@/lib/auth/context';
import { extractErrorMessage } from '@/lib/api/client';
import {
  approvalsByResource,
  approveRequest,
  listApprovals,
  rejectRequest,
  withdrawRequest,
  type ApprovalRequest,
} from '@/lib/api/approval';
import { Clock, CheckCircle2, XCircle, ArrowRight } from 'lucide-react';

const STATUS_LABEL: Record<string, string> = {
  draft: '草稿',
  pending: '待审批',
  approved: '已通过',
  rejected: '已驳回',
  withdrawn: '已撤回',
};

function statusVariant(s: string): 'default' | 'secondary' | 'destructive' | 'success' {
  if (s === 'approved') return 'success';
  if (s === 'rejected' || s === 'withdrawn') return 'destructive';
  if (s === 'pending') return 'default';
  return 'secondary';
}

function fmtTime(s?: string | null): string {
  if (!s) return '-';
  const d = new Date(s);
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

function durationLabel(created: string, reviewed?: string | null): string {
  if (!reviewed) return '-';
  const ms = new Date(reviewed).getTime() - new Date(created).getTime();
  const hours = Math.floor(ms / 3600000);
  const mins = Math.floor((ms % 3600000) / 60000);
  if (hours > 0) return `${hours}h${mins}m`;
  return `${mins}m`;
}

type Tab = 'pending' | 'mine' | 'all';

export default function ApprovalsPage() {
  const qc = useQueryClient();
  const { user } = useAuth();
  const { has } = usePermission();
  const canReview = has('system:write');

  const [tab, setTabState] = useState<Tab>('pending');
  const [selected, setSelected] = useState<ApprovalRequest | null>(null);
  const [error, setError] = useState<string | null>(null);
  const PAGE_SIZE = 50;
  const [page, setPage] = useState(1);

  // 切换分页视图时回到第 1 页(否则停留在旧页码会越界)
  const setTab = (t: Tab) => {
    setTabState(t);
    setPage(1);
  };

  const queryArgs =
    tab === 'mine'
      ? { mine: true }
      : tab === 'pending'
      ? { status: 'pending' }
      : { status: 'pending,approved,rejected,withdrawn' };

  const { data, isLoading } = useQuery({
    queryKey: ['approvals', tab, user?.username, page],
    queryFn: () =>
      listApprovals({ ...queryArgs, limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE }),
  });

  const items = useMemo(() => data?.items ?? [], [data]);
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // ── Approval efficiency stats ──
  const effStats = useMemo(() => {
    const reviewed = items.filter((i) => i.reviewed_at && i.created_at);
    const avgMs =
      reviewed.length > 0
        ? reviewed.reduce((s, i) => s + (new Date(i.reviewed_at!).getTime() - new Date(i.created_at).getTime()), 0) / reviewed.length
        : 0;
    const avgHours = Math.round(avgMs / 3600000 * 10) / 10;

    // by reviewer
    const reviewerMap = new Map<string, { count: number; totalMs: number }>();
    for (const i of reviewed) {
      if (!i.reviewed_by) continue;
      const entry = reviewerMap.get(i.reviewed_by) ?? { count: 0, totalMs: 0 };
      entry.count += 1;
      entry.totalMs += new Date(i.reviewed_at!).getTime() - new Date(i.created_at).getTime();
      reviewerMap.set(i.reviewed_by, entry);
    }
    const reviewers = Array.from(reviewerMap.entries()).map(([name, { count, totalMs }]) => ({
      name,
      count,
      avgHours: Math.round(totalMs / count / 3600000 * 10) / 10,
    }));

    return { avgHours, reviewers };
  }, [items]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">审批中心</h1>
        <p className="text-sm text-muted-foreground">
          通用审批：合同变更 / 客户档案修改 / 价格调整等任意业务资源
        </p>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex gap-2">
        <Button
          variant={tab === 'pending' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setTab('pending')}
        >
          待审批
        </Button>
        <Button
          variant={tab === 'mine' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setTab('mine')}
        >
          我提交的
        </Button>
        <Button
          variant={tab === 'all' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setTab('all')}
        >
          全部
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">条目总数</p>
            <p className="mt-1 text-2xl font-bold">{total}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">待审批（本页）</p>
            <p className="mt-1 text-2xl font-bold text-amber-600">
              {items.filter((i) => i.status === 'pending').length}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-xs text-muted-foreground">已通过（本页）</p>
            <p className="mt-1 text-2xl font-bold text-emerald-600">
              {items.filter((i) => i.status === 'approved').length}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* ═══════════ Approval Flow Step Visualization ═══════════ */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">审批流程可视化（本页）</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-0 overflow-x-auto pb-2">
            {(['draft', 'pending', 'approved'] as const).map((step, idx) => {
              const count = items.filter((i) => i.status === step).length;
              const isActive = tab === 'pending' && step === 'pending';
              return (
                <div key={step} className="flex items-center">
                  <div className={`flex flex-col items-center rounded-lg border px-6 py-3 min-w-[100px] ${isActive ? 'border-primary bg-primary/5' : 'bg-muted/30'}`}>
                    <div className={`flex h-8 w-8 items-center justify-center rounded-full ${step === 'approved' ? 'bg-emerald-100 text-emerald-600' : step === 'pending' ? 'bg-amber-100 text-amber-600' : 'bg-gray-100 text-gray-500'}`}>
                      {step === 'approved' ? <CheckCircle2 className="h-4 w-4" /> : step === 'pending' ? <Clock className="h-4 w-4" /> : <div className="h-2 w-2 rounded-full bg-current" />}
                    </div>
                    <span className="mt-1 text-sm font-medium">{STATUS_LABEL[step]}</span>
                    <Badge variant={isActive ? 'default' : 'secondary'} className="mt-1 text-xs">{count}</Badge>
                  </div>
                  {idx < 2 && (
                    <ArrowRight className="mx-2 h-4 w-4 text-muted-foreground shrink-0" />
                  )}
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* ═══════════ Approval Efficiency Stats ═══════════ */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">审批效率统计（本页）</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between rounded-md border p-3">
                <span className="text-sm text-muted-foreground">平均审批时长</span>
                <span className="text-lg font-bold">{effStats.avgHours}h</span>
              </div>
              <div className="flex items-center justify-between rounded-md border p-3">
                <span className="text-sm text-muted-foreground">已审批总数</span>
                <span className="text-lg font-bold">{items.filter((i) => i.reviewed_at).length}</span>
              </div>
              <div className="flex items-center justify-between rounded-md border p-3">
                <span className="text-sm text-muted-foreground">通过率</span>
                <span className="text-lg font-bold text-emerald-600">
                  {items.length > 0 ? Math.round(items.filter((i) => i.status === 'approved').length / items.length * 100) : 0}%
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">各审批人响应速度（本页）</CardTitle>
          </CardHeader>
          <CardContent>
            {effStats.reviewers.length === 0 ? (
              <p className="text-sm text-muted-foreground">暂无审批人数据</p>
            ) : (
              <div className="space-y-2">
                {effStats.reviewers.map((r) => (
                  <div key={r.name} className="flex items-center justify-between rounded-md border p-2">
                    <span className="text-sm font-medium">{r.name}</span>
                    <div className="flex items-center gap-3">
                      <Badge variant="outline" className="text-xs">{r.count} 次</Badge>
                      <span className="text-sm">平均 {r.avgHours}h</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>提交时间</TableHead>
              <TableHead>资源</TableHead>
              <TableHead>标题</TableHead>
              <TableHead>提交人</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>审批人</TableHead>
              <TableHead>审批时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
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
            {items.map((a) => (
              <TableRow key={a.id}>
                <TableCell className="whitespace-nowrap">{fmtTime(a.created_at)}</TableCell>
                <TableCell>
                  <Badge variant="outline">{a.resource}</Badge>
                </TableCell>
                <TableCell className="font-medium">{a.title}</TableCell>
                <TableCell>{a.submitted_by}</TableCell>
                <TableCell>
                  <Badge variant={statusVariant(a.status)}>{STATUS_LABEL[a.status] ?? a.status}</Badge>
                </TableCell>
                <TableCell>{a.reviewed_by ?? '-'}</TableCell>
                <TableCell className="whitespace-nowrap">{fmtTime(a.reviewed_at)}</TableCell>
                <TableCell className="text-right">
                  <Button size="sm" variant="ghost" onClick={() => setSelected(a)}>
                    查看
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {items.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  暂无审批条目
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* 分页:total 来自后端同条件 COUNT,翻页保留当前标签页 */}
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

      {selected && (
        <ApprovalDetailDialog
          approval={selected}
          canReview={canReview}
          currentUser={user?.username ?? ''}
          onClose={() => setSelected(null)}
          onReviewed={() => {
            setSelected(null);
            qc.invalidateQueries({ queryKey: ['approvals'] });
          }}
          onError={(e) => setError(e)}
        />
      )}
    </div>
  );
}

function ApprovalDetailDialog({
  approval,
  canReview,
  currentUser,
  onClose,
  onReviewed,
  onError,
}: {
  approval: ApprovalRequest;
  canReview: boolean;
  currentUser: string;
  onClose: () => void;
  onReviewed: () => void;
  onError: (msg: string) => void;
}) {
  const [note, setNote] = useState('');
  const [busy, setBusy] = useState(false);

  const { data: history } = useQuery({
    queryKey: ['approval-history', approval.resource, approval.resource_id],
    queryFn: () => approvalsByResource(approval.resource, approval.resource_id),
  });

  const isPending = approval.status === 'pending';
  const canWithdraw = isPending && approval.submitted_by === currentUser;

  const act = async (fn: () => Promise<ApprovalRequest>) => {
    setBusy(true);
    try {
      await fn();
      onReviewed();
    } catch (e) {
      onError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      <DialogHeader>
        <DialogTitle>审批详情</DialogTitle>
      </DialogHeader>

      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <p className="text-xs text-muted-foreground">资源</p>
            <Badge variant="outline">{approval.resource}</Badge>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">资源 ID</p>
            <p className="font-mono text-xs">{approval.resource_id}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">提交人</p>
            <p>{approval.submitted_by}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">状态</p>
            <Badge variant={statusVariant(approval.status)}>
              {STATUS_LABEL[approval.status] ?? approval.status}
            </Badge>
          </div>
        </div>

        <div>
          <p className="text-xs text-muted-foreground">标题</p>
          <p className="text-sm font-medium">{approval.title}</p>
        </div>

        {/* ═══ Flow Steps ═══ */}
        <div>
          <p className="mb-2 text-xs text-muted-foreground">审批流程</p>
          <div className="flex items-center gap-1">
            {(['draft', 'pending', 'approved'] as const).map((step, idx) => {
              const isCurrent = approval.status === step;
              const isDone = ['approved'].includes(step) && approval.status === 'approved';
              const isRejected = approval.status === 'rejected';
              return (
                <div key={step} className="flex items-center">
                  <div className={`flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium ${isCurrent ? 'bg-primary text-primary-foreground' : isDone ? 'bg-emerald-100 text-emerald-700' : isRejected && step === 'pending' ? 'bg-red-100 text-red-700' : 'bg-muted text-muted-foreground'}`}>
                    {isDone && <CheckCircle2 className="h-3 w-3" />}
                    {isRejected && step === 'pending' && <XCircle className="h-3 w-3" />}
                    {STATUS_LABEL[step]}
                  </div>
                  {idx < 2 && <ArrowRight className="mx-1 h-3 w-3 text-muted-foreground" />}
                </div>
              );
            })}
          </div>
        </div>

        <div>
          <p className="mb-1 text-xs text-muted-foreground">变更详情（payload）</p>
          <pre className="max-h-60 overflow-auto rounded-md bg-muted p-3 text-xs">
            {JSON.stringify(approval.payload, null, 2)}
          </pre>
        </div>

        {approval.reviewed_by && (
          <div className="rounded-md border bg-muted/30 p-3 text-sm">
            <p className="text-xs text-muted-foreground">
              {approval.status === 'approved' ? '通过' : approval.status === 'rejected' ? '驳回' : '处理'}
              于 {fmtTime(approval.reviewed_at)} · 由 {approval.reviewed_by}
            </p>
            {approval.review_note && (
              <p className="mt-1">备注：{approval.review_note}</p>
            )}
          </div>
        )}

        {isPending && canReview && (
          <div>
            <p className="mb-1 text-xs text-muted-foreground">审批备注（可选）</p>
            <Input
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="给提交人留个说明..."
            />
          </div>
        )}

        {history && history.items.length > 1 && (
          <div>
            <p className="mb-1 text-xs text-muted-foreground">
              该资源的历史变更（{history.items.length} 条）
            </p>
            <div className="max-h-48 overflow-auto rounded-md border bg-muted/30 p-2">
              <ul className="space-y-2">
                {history.items.map((h) => (
                  <li
                    key={h.id}
                    className={`flex items-start gap-2 text-xs ${h.id === approval.id ? 'font-bold' : ''}`}
                  >
                    <Badge variant={statusVariant(h.status)} className="shrink-0">
                      {STATUS_LABEL[h.status] ?? h.status}
                    </Badge>
                    <div className="min-w-0 flex-1">
                      <p className="truncate">{h.title}</p>
                      <p className="text-[10px] opacity-60">
                        {fmtTime(h.created_at)} · {h.submitted_by}
                        {h.reviewed_by && ` → ${h.reviewed_by}`}
                      </p>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}
      </div>

      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          关闭
        </Button>
        {canWithdraw && (
          <Button
            variant="outline"
            disabled={busy}
            onClick={() => act(() => withdrawRequest(approval.id))}
          >
            撤回
          </Button>
        )}
        {isPending && canReview && (
          <>
            <Button
              variant="outline"
              disabled={busy}
              onClick={() => act(() => rejectRequest(approval.id, note))}
            >
              驳回
            </Button>
            <Button
              disabled={busy}
              onClick={() => act(() => approveRequest(approval.id, note))}
            >
              {busy ? '提交中...' : '通过'}
            </Button>
          </>
        )}
      </DialogFooter>
    </Dialog>
  );
}
