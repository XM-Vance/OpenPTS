'use client';

import { useState, useCallback } from 'react';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
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
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createBond,
  deleteBond,
  listBonds,
  updateBond,
  type Bond,
} from '@/lib/api/bonds';

const STATUS_MAP: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' | 'destructive' }> = {
  active: { label: '有效', variant: 'default' },
  expired: { label: '已到期', variant: 'secondary' },
  returned: { label: '已退还', variant: 'outline' },
};

export default function BondsPage() {
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

  const [editing, setEditing] = useState<Bond | 'new' | null>(null);
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
    queryKey: ['bonds', search, statusFilter, page, pageSize],
    queryFn: () =>
      listBonds({
        keyword: search || undefined,
        status: statusFilter || undefined,
        page,
        page_size: pageSize,
      }),
  });

  const onDelete = async (id: string, name: string) => {
    if (!window.confirm(`确认删除保函「${name}」？`)) return;
    try {
      await deleteBond(id);
      qc.invalidateQueries({ queryKey: ['bonds'] });
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  const fmtDate = (d: string | null | undefined) => (d ? d.slice(0, 10) : '-');

  const items = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">保函管理</h1>
        {canWrite && <Button onClick={() => setEditing('new')}>新建保函</Button>}
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="搜索保函名称 / 开立机构"
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
          <option value="active">有效</option>
          <option value="expired">已到期</option>
          <option value="returned">已退还</option>
        </select>
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>保函名称</TableHead>
              <TableHead>保函类型</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>开立机构</TableHead>
              <TableHead>受益人</TableHead>
              <TableHead>开立日期</TableHead>
              <TableHead>到期日期</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={9} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            )}
            {items.map((bond) => {
              const st = STATUS_MAP[bond.status] ?? { label: bond.status, variant: 'outline' };
              return (
                <TableRow key={bond.id}>
                  <TableCell className="font-medium">{bond.name}</TableCell>
                  <TableCell>{bond.bond_type || '-'}</TableCell>
                  <TableCell>
                    {bond.amount.toLocaleString('zh-CN', {
                      style: 'currency',
                      currency: 'CNY',
                      minimumFractionDigits: 2,
                    })}
                  </TableCell>
                  <TableCell>{bond.issuer || '-'}</TableCell>
                  <TableCell>{bond.beneficiary || '-'}</TableCell>
                  <TableCell>{fmtDate(bond.issue_date)}</TableCell>
                  <TableCell>{fmtDate(bond.expire_date)}</TableCell>
                  <TableCell>
                    <Badge variant={st.variant}>{st.label}</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      {canWrite && (
                        <Button size="sm" variant="ghost" onClick={() => setEditing(bond)}>
                          编辑
                        </Button>
                      )}
                      {canDelete && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => onDelete(bond.id, bond.name)}
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
                <TableCell colSpan={9} className="text-center text-muted-foreground">
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
        <BondFormDialog
          bond={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            qc.invalidateQueries({ queryKey: ['bonds'] });
          }}
        />
      )}
    </div>
  );
}

function BondFormDialog({
  bond,
  onClose,
  onSaved,
}: {
  bond: Bond | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !bond;
  const [name, setName] = useState(bond?.name ?? '');
  const [bondType, setBondType] = useState(bond?.bond_type ?? '');
  const [amount, setAmount] = useState(bond ? String(bond.amount) : '0');
  const [issuer, setIssuer] = useState(bond?.issuer ?? '');
  const [beneficiary, setBeneficiary] = useState(bond?.beneficiary ?? '');
  const [issueDate, setIssueDate] = useState(bond?.issue_date ? bond.issue_date.slice(0, 10) : '');
  const [expireDate, setExpireDate] = useState(bond?.expire_date ? bond.expire_date.slice(0, 10) : '');
  const [status, setStatus] = useState(bond?.status ?? 'active');
  const [description, setDescription] = useState(bond?.description ?? '');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      const input = {
        name,
        bond_type: bondType || undefined,
        amount: parseFloat(amount) || 0,
        issuer: issuer || undefined,
        beneficiary: beneficiary || undefined,
        issue_date: issueDate || undefined,
        expire_date: expireDate || undefined,
        status,
        description: description || undefined,
      };
      if (isNew) {
        await createBond(input);
      } else if (bond) {
        await updateBond(bond.id, input);
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
        <DialogTitle>{isNew ? '新建保函' : '编辑保函'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>保函名称</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>保函类型</Label>
            <Input value={bondType} onChange={(e) => setBondType(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>金额（元）</Label>
            <Input
              type="number"
              min="0"
              step="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
            />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>开立机构</Label>
            <Input value={issuer} onChange={(e) => setIssuer(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>受益人</Label>
            <Input value={beneficiary} onChange={(e) => setBeneficiary(e.target.value)} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>开立日期</Label>
            <Input type="date" value={issueDate} onChange={(e) => setIssueDate(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>到期日期</Label>
            <Input type="date" value={expireDate} onChange={(e) => setExpireDate(e.target.value)} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>状态</Label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
            >
              <option value="active">有效</option>
              <option value="expired">已到期</option>
              <option value="returned">已退还</option>
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
