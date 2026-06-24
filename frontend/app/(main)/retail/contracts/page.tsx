'use client';

import Link from 'next/link';
import { useState, useMemo, useEffect } from 'react';
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
import { RetailTabs } from '@/components/retail/retail-tabs';
import { AttachmentPanel } from '@/components/attachments/attachment-panel';
import { ChartContainer } from '@/components/charts/chart-container';

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const StatusPie = dynamic(
  () => import('./_status-pie').then((m) => ({ default: m.StatusPie })),
  { ssr: false, loading: () => <div className="h-[180px]" /> },
);
import { listApprovalTemplates, submitApproval } from '@/lib/api/approval';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createContract,
  deleteContract,
  generateContractPDF,
  listContracts,
  listPackages,
  updateContract,
  type RetailContract,
  type RetailPackage,
} from '@/lib/api/retail';
import { listCustomers, searchCustomersAllOrg, type Customer } from '@/lib/api/customers';
import { takeContractPrefill, type ContractPrefill } from '@/lib/contract-prefill';

const STATUS_LABEL: Record<string, string> = {
  active: '生效中',
  expired: '已到期',
  terminated: '已终止',
};

const SELECT_CLASS =
  'flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm';

// 合同到期颜色：7天内红色，30天内黄色，其他正常
function getExpiryClass(endMonth: string, status: string): string {
  if (status !== 'active') return '';
  const end = new Date(endMonth + '-28');
  const now = new Date();
  const diffDays = (end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24);
  if (diffDays <= 7) return 'bg-red-50';
  if (diffDays <= 30) return 'bg-yellow-50';
  return '';
}

function getExpiryBadge(endMonth: string, status: string): { text: string; variant: 'destructive' | 'secondary' } | null {
  if (status !== 'active') return null;
  const end = new Date(endMonth + '-28');
  const now = new Date();
  const diffDays = (end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24);
  if (diffDays <= 7) return { text: '即将到期', variant: 'destructive' };
  if (diffDays <= 30) return { text: '30天内到期', variant: 'secondary' };
  return null;
}

export default function RetailContractsPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('retail_management:write');
  const canDelete = has('retail_management:delete');

  const [keyword, setKeyword] = useState('');
  const [search, setSearch] = useState('');
  const [editing, setEditing] = useState<RetailContract | 'new' | null>(null);
  const [prefill, setPrefill] = useState<ContractPrefill | null>(null);
  const [filesFor, setFilesFor] = useState<RetailContract | null>(null);
  const [approvalFor, setApprovalFor] = useState<RetailContract | null>(null);

  // 从合同文档跳转而来时，自动打开新建对话框并预填
  useEffect(() => {
    const p = takeContractPrefill();
    if (p) {
      setPrefill(p);
      setEditing('new');
    }
  }, []);

  const { data: contracts, isLoading } = useQuery({
    queryKey: ['retail-contracts', search],
    queryFn: () => listContracts({ keyword: search }),
  });
  const { data: customers } = useQuery({
    queryKey: ['customers', 'retail-form'],
    queryFn: () => listCustomers({ limit: 200 }),
  });

  // 当从文档预填跳来时，用候选名搜索 API 匹配客户（解决 limit:200 不覆盖全部客户的问题）
  const [searchedCustomers, setSearchedCustomers] = useState<Customer[]>([]);
  useEffect(() => {
    const candidates = prefill?.customerCandidates;
    if (!candidates || candidates.length === 0) return;
    let cancelled = false;
    (async () => {
      const found: Customer[] = [];
      const seen = new Set<string>();
      for (const name of candidates) {
        if (!name?.trim()) continue;
        try {
          // 跨省搜索：文档中的客户可能在其他省份
          const items = await searchCustomersAllOrg(name.trim(), 5);
          for (const c of items) {
            if (!seen.has(c.id)) {
              seen.add(c.id);
              found.push(c);
            }
          }
        } catch { /* ignore */ }
      }
      if (!cancelled && found.length > 0) setSearchedCustomers(found);
    })();
    return () => { cancelled = true; };
  }, [prefill]);

  // 合并已有客户列表 + 搜索到的客户（去重）
  const allCustomers = useMemo(() => {
    const merged = [...(customers?.items ?? [])];
    const seen = new Set(merged.map((c) => c.id));
    for (const c of searchedCustomers) {
      if (!seen.has(c.id)) {
        merged.push(c);
        seen.add(c.id);
      }
    }
    return merged;
  }, [customers, searchedCustomers]);
  const { data: packages } = useQuery({
    queryKey: ['retail-packages', ''],
    queryFn: () => listPackages(),
  });

  const onDelete = async (id: string, name: string) => {
    if (!window.confirm(`确认删除「${name}」的合同？`)) return;
    try {
      await deleteContract(id);
      qc.invalidateQueries({ queryKey: ['retail-contracts'] });
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  const onGenPDF = async (ct: RetailContract) => {
    try {
      const r = await generateContractPDF(ct.id);
      if (r.mode === 'minio') {
        qc.invalidateQueries({ queryKey: ['attachments', 'retail_contracts', ct.id] });
        window.alert('已生成 PDF 并加为附件，请到附件面板查看');
      }
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  const contractList = useMemo(() => contracts ?? [], [contracts]);

  // ── 合同甘特图数据 ──
  const ganttData = useMemo(() => {
    if (!contractList.length) return [];
    // 找出全局时间范围
    const allMonths = contractList.flatMap((c) => [c.purchase_start_month, c.purchase_end_month]);
    const minMonth = allMonths.sort()[0];
    const maxMonth = allMonths.sort().at(-1) ?? minMonth;
    const minDate = new Date(minMonth + '-01');
    const maxDate = new Date(maxMonth + '-28');
    const totalDays = Math.max(1, (maxDate.getTime() - minDate.getTime()) / (1000 * 60 * 60 * 24));

    return contractList.slice(0, 10).map((c) => {
      const start = new Date(c.purchase_start_month + '-01');
      const end = new Date(c.purchase_end_month + '-28');
      const leftPct = Math.max(0, ((start.getTime() - minDate.getTime()) / (1000 * 60 * 60 * 24)) / totalDays * 100);
      const widthPct = Math.max(2, ((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) / totalDays * 100);
      const statusColor = c.status === 'active' ? '#2563eb' : c.status === 'expired' ? '#94a3b8' : '#ef4444';
      return {
        name: c.customer_name.length > 8 ? c.customer_name.slice(0, 8) + '…' : c.customer_name,
        pkg: c.package_name_snapshot,
        leftPct,
        widthPct,
        statusColor,
        startMonth: c.purchase_start_month,
        endMonth: c.purchase_end_month,
        status: c.status,
      };
    });
  }, [contractList]);

  // ── 合同执行进度环形图 ──
  const pieData = useMemo(() => {
    const active = contractList.filter((c) => c.status === 'active').length;
    const expired = contractList.filter((c) => c.status === 'expired').length;
    const terminated = contractList.filter((c) => c.status === 'terminated').length;
    return [
      { name: '生效中', value: active, color: '#2563eb' },
      { name: '已到期', value: expired, color: '#94a3b8' },
      { name: '已终止', value: terminated, color: '#ef4444' },
    ].filter((d) => d.value > 0);
  }, [contractList]);

  // ── 到期提醒统计 ──
  const expiryStats = useMemo(() => {
    let within7 = 0;
    let within30 = 0;
    for (const c of contractList) {
      if (c.status !== 'active') continue;
      const end = new Date(c.purchase_end_month + '-28');
      const diff = (end.getTime() - Date.now()) / (1000 * 60 * 60 * 24);
      if (diff <= 7) within7++;
      else if (diff <= 30) within30++;
    }
    return { within7, within30 };
  }, [contractList]);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">零售管理</h1>
        <p className="text-sm text-muted-foreground">零售合同、套餐与定价</p>
      </div>
      <RetailTabs />

      {/* ── 到期提醒卡片 ── */}
      {(expiryStats.within7 > 0 || expiryStats.within30 > 0) && (
        <div className="grid gap-4 md:grid-cols-2">
          {expiryStats.within7 > 0 && (
            <div className="rounded-lg border border-red-300 bg-red-50 p-4">
              <p className="text-sm font-semibold text-red-700">⚠️ 7天内到期合同</p>
              <p className="mt-1 text-2xl font-bold text-red-600">{expiryStats.within7} 个</p>
              <p className="text-xs text-red-500 mt-1">请尽快联系客户续签</p>
            </div>
          )}
          {expiryStats.within30 > 0 && (
            <div className="rounded-lg border border-yellow-300 bg-yellow-50 p-4">
              <p className="text-sm font-semibold text-yellow-700">⏰ 30天内到期合同</p>
              <p className="mt-1 text-2xl font-bold text-yellow-600">{expiryStats.within30} 个</p>
              <p className="text-xs text-yellow-500 mt-1">提前准备续签或变更方案</p>
            </div>
          )}
        </div>
      )}

      {/* ── 图表区 ── */}
      <div className="grid gap-4 md:grid-cols-3">
        {/* 合同甘特图 */}
        <div className="md:col-span-2">
          <ChartContainer title="合同执行周期时间轴" minHeight={Math.max(200, ganttData.length * 36 + 60)}>
            {ganttData.length > 0 ? (
              <div className="space-y-1.5">
                <div className="flex items-center gap-3 text-xs text-zinc-400 pb-1">
                  <span className="flex items-center gap-1"><span className="inline-block h-3 w-6 rounded-sm" style={{ backgroundColor: '#2563eb' }} /> 生效中</span>
                  <span className="flex items-center gap-1"><span className="inline-block h-3 w-6 rounded-sm" style={{ backgroundColor: '#94a3b8' }} /> 已到期</span>
                  <span className="flex items-center gap-1"><span className="inline-block h-3 w-6 rounded-sm" style={{ backgroundColor: '#ef4444' }} /> 已终止</span>
                </div>
                {ganttData.map((row, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    <div className="w-24 shrink-0 truncate text-right text-xs text-zinc-500">{row.name}</div>
                    <div className="relative h-6 flex-1 rounded bg-zinc-50">
                      <div
                        className="absolute top-0 h-full rounded-sm"
                        style={{
                          left: `${row.leftPct}%`,
                          width: `${row.widthPct}%`,
                          backgroundColor: row.statusColor,
                        }}
                        title={`${row.name}: ${row.startMonth} ~ ${row.endMonth} (${STATUS_LABEL[row.status] ?? row.status})`}
                      />
                    </div>
                    <span className="w-20 shrink-0 text-xs text-zinc-400">{row.startMonth}~{row.endMonth.slice(5)}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
                暂无合同数据
              </div>
            )}
          </ChartContainer>
        </div>

        {/* 合同执行进度环形图 */}
        <ChartContainer title="合同状态分布" minHeight={240}>
          {pieData.length > 0 ? (
            <div className="flex flex-col items-center">
              <StatusPie data={pieData} />
              <div className="flex items-center gap-4 text-xs mt-2">
                {pieData.map((d) => (
                  <span key={d.name} className="flex items-center gap-1">
                    <span className="inline-block h-3 w-3 rounded-sm" style={{ backgroundColor: d.color }} />
                    {d.name} ({d.value})
                  </span>
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

      <div className="flex items-center justify-between pt-2">
        <div className="flex gap-2">
          <Input
            placeholder="搜索客户 / 套餐名"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && setSearch(keyword)}
            className="max-w-xs"
          />
          <Button variant="outline" onClick={() => setSearch(keyword)}>
            搜索
          </Button>
        </div>
        {canWrite && <Button onClick={() => setEditing('new')}>新建合同</Button>}
      </div>

      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>客户</TableHead>
              <TableHead>套餐</TableHead>
              <TableHead>购电量 (MWh)</TableHead>
              <TableHead>绿电占比</TableHead>
              <TableHead>购电区间</TableHead>
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
            {contractList.map((ct) => {
              const expiryBadge = getExpiryBadge(ct.purchase_end_month, ct.status);
              const rowClass = getExpiryClass(ct.purchase_end_month, ct.status);
              return (
                <TableRow key={ct.id} className={rowClass}>
                  <TableCell className="font-medium">
                    <Link
                      href={`/customers?highlight=${ct.customer_id}`}
                      className="text-blue-600 hover:underline"
                    >
                      {ct.customer_name}
                    </Link>
                    {expiryBadge && (
                      <Badge variant={expiryBadge.variant} className="ml-2 text-[10px] px-1.5 py-0">
                        {expiryBadge.text}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>{ct.package_name_snapshot}</TableCell>
                  <TableCell>{ct.purchasing_energy_mwh}</TableCell>
                  <TableCell>
                    {ct.green_power_ratio != null ? `${ct.green_power_ratio}%` : '-'}
                  </TableCell>
                  <TableCell>
                    {ct.purchase_start_month} ~ {ct.purchase_end_month}
                  </TableCell>
                  <TableCell>
                    <Badge variant={ct.status === 'active' ? 'success' : 'secondary'}>
                      {STATUS_LABEL[ct.status] || ct.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      <Button size="sm" variant="ghost" onClick={() => setFilesFor(ct)}>
                        附件
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => setApprovalFor(ct)}>
                        申请变更
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => onGenPDF(ct)}>
                        PDF
                      </Button>
                      {canWrite && (
                        <Button size="sm" variant="ghost" onClick={() => setEditing(ct)}>
                          编辑
                        </Button>
                      )}
                      {canDelete && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => onDelete(ct.id, ct.customer_name)}
                        >
                          删除
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
            {contractList.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {editing && (
        <ContractFormDialog
          contract={editing === 'new' ? null : editing}
          prefill={editing === 'new' ? prefill : null}
          customers={allCustomers}
          packages={packages ?? []}
          onClose={() => {
            setEditing(null);
            setPrefill(null);
          }}
          onSaved={() => {
            setEditing(null);
            setPrefill(null);
            qc.invalidateQueries({ queryKey: ['retail-contracts'] });
          }}
        />
      )}

      {filesFor && (
        <Dialog open onClose={() => setFilesFor(null)}>
          <DialogHeader>
            <DialogTitle>
              合同附件 · {filesFor.customer_name} / {filesFor.package_name_snapshot}
            </DialogTitle>
          </DialogHeader>
          <AttachmentPanel
            resource="retail_contracts"
            resourceId={filesFor.id}
            canWrite={canWrite}
            canDelete={canDelete}
            title="合同相关文档"
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setFilesFor(null)}>
              关闭
            </Button>
          </DialogFooter>
        </Dialog>
      )}

      {approvalFor && (
        <ContractApprovalDialog
          contract={approvalFor}
          onClose={() => setApprovalFor(null)}
        />
      )}
    </div>
  );
}

function ContractApprovalDialog({
  contract,
  onClose,
}: {
  contract: RetailContract;
  onClose: () => void;
}) {
  const { data: tpls } = useQuery({
    queryKey: ['approval-templates', 'retail_contracts'],
    queryFn: () => listApprovalTemplates('retail_contracts'),
  });
  const [field, setField] = useState<'purchasing_energy_mwh' | 'green_power_ratio' | 'purchase_end_month' | 'status'>('purchasing_energy_mwh');
  const oldValue =
    field === 'purchasing_energy_mwh'
      ? String(contract.purchasing_energy_mwh)
      : field === 'green_power_ratio'
        ? String(contract.green_power_ratio ?? '')
        : field === 'purchase_end_month'
          ? contract.purchase_end_month
          : contract.status;
  const [newValue, setNewValue] = useState('');
  const [note, setNote] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const [ok, setOk] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setErr(null);
    if (!newValue) {
      setErr('请填写新值');
      return;
    }
    setBusy(true);
    try {
      await submitApproval({
        resource: 'retail_contracts',
        resource_id: contract.id,
        title: `合同变更：${contract.customer_name} · ${field} = ${newValue}`,
        payload: {
          field,
          old: oldValue,
          new: newValue,
          note,
        },
      });
      setOk(true);
      setTimeout(onClose, 1500);
    } catch (e) {
      setErr(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open onClose={onClose}>
      <DialogHeader>
        <DialogTitle>申请合同变更 · {contract.customer_name}</DialogTitle>
      </DialogHeader>
      <div className="space-y-3">
        <p className="text-xs text-muted-foreground">
          提交后将进入审批中心；通过后由审批人手工同步到合同（系统支持自动落库）。
        </p>
        {tpls && tpls.items.length > 0 && (
          <div className="space-y-1">
            <Label>选择模板</Label>
            <select
              defaultValue=""
              onChange={(e) => {
                const t = tpls.items.find((x) => x.id === e.target.value);
                if (t) setField(t.field as typeof field);
              }}
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm"
            >
              <option value="">（自定义）</option>
              {tpls.items.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          </div>
        )}
        <div className="space-y-1">
          <Label>变更字段</Label>
          <select
            value={field}
            onChange={(e) => {
              setField(e.target.value as typeof field);
              setNewValue('');
            }}
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm"
          >
            <option value="purchasing_energy_mwh">购电量 (MWh)</option>
            <option value="green_power_ratio">绿电占比 (%)</option>
            <option value="purchase_end_month">购电结束月</option>
            <option value="status">合同状态</option>
          </select>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <Label>当前值</Label>
            <Input value={oldValue} readOnly className="bg-muted" />
          </div>
          <div className="space-y-1">
            <Label>申请新值</Label>
            <Input
              value={newValue}
              onChange={(e) => setNewValue(e.target.value)}
              placeholder={field === 'purchase_end_month' ? 'YYYY-MM' : ''}
            />
          </div>
        </div>
        <div className="space-y-1">
          <Label>变更原因</Label>
          <Input value={note} onChange={(e) => setNote(e.target.value)} placeholder="可选" />
        </div>
        {err && <p className="text-xs text-destructive">{err}</p>}
        {ok && (
          <p className="text-xs text-emerald-600">已提交审批，可在「审批中心」查看进度</p>
        )}
      </div>
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>
          取消
        </Button>
        <Button onClick={submit} disabled={busy || ok}>
          {busy ? '提交中...' : '提交审批'}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}

function ContractFormDialog({
  contract,
  prefill,
  customers,
  packages,
  onClose,
  onSaved,
}: {
  contract: RetailContract | null;
  prefill?: ContractPrefill | null;
  customers: Customer[];
  packages: RetailPackage[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !contract;
  const [customerId, setCustomerId] = useState(contract?.customer_id ?? '');
  const [packageId, setPackageId] = useState(contract?.package_id ?? '');
  const [energy, setEnergy] = useState(
    contract ? String(contract.purchasing_energy_mwh) : (prefill?.energyMwh ?? ''),
  );
  const [greenRatio, setGreenRatio] = useState(
    contract?.green_power_ratio != null ? String(contract.green_power_ratio) : (prefill?.greenRatio ?? ''),
  );
  const [startMonth, setStartMonth] = useState(contract?.purchase_start_month ?? prefill?.startMonth ?? '');
  const [endMonth, setEndMonth] = useState(contract?.purchase_end_month ?? prefill?.endMonth ?? '');
  const [status, setStatus] = useState(contract?.status ?? 'active');

  // ── 扩展字段（从文档预填，可修改）──
  const [contractNo, setContractNo] = useState(prefill?.contractNo ?? '');
  const [price, setPrice] = useState(prefill?.price ?? '');
  const [totalAmount, setTotalAmount] = useState(prefill?.totalAmount ?? '');
  const [signDate, setSignDate] = useState(prefill?.signDate ?? '');
  const [voltageLevel, setVoltageLevel] = useState(prefill?.voltageLevel ?? '');
  const [settlementMethod, setSettlementMethod] = useState(prefill?.settlementMethod ?? '');
  const [showAllFields, setShowAllFields] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // 预填：客户列表就绪后，按甲方/乙方候选名自动匹配客户（仅在未选时）
  useEffect(() => {
    if (!prefill || customerId || customers.length === 0) return;
    const cands = (prefill.customerCandidates ?? []).map((s) => s.trim()).filter(Boolean);
    for (const cand of cands) {
      const hit = customers.find(
        (c) => c.user_name && (c.user_name.includes(cand) || cand.includes(c.user_name)),
      );
      if (hit) {
        setCustomerId(hit.id);
        break;
      }
    }
  }, [prefill, customers, customerId]);

  // 预填：套餐列表就绪后，按套餐名称提示自动匹配
  useEffect(() => {
    if (!prefill?.packageNameHint || packageId || packages.length === 0) return;
    const hint = prefill.packageNameHint.trim();
    const hit = packages.find(
      (p) => p.package_name.includes(hint) || hint.includes(p.package_name),
    );
    if (hit) setPackageId(hit.id);
  }, [prefill, packages, packageId]);

  const submit = async () => {
    setError(null);
    if (!customerId) {
      setError('请选择客户');
      return;
    }
    if (!packageId) {
      setError('请选择套餐');
      return;
    }
    if (!startMonth || !endMonth) {
      setError('请选择购电起止月份');
      return;
    }
    setSubmitting(true);
    try {
      const input = {
        customer_id: customerId,
        package_id: packageId,
        purchasing_energy_mwh: energy !== '' ? Number(energy) : 0,
        green_power_ratio: greenRatio !== '' ? Number(greenRatio) : undefined,
        purchase_start_month: startMonth,
        purchase_end_month: endMonth,
        status,
      };
      if (isNew) await createContract(input);
      else if (contract) await updateContract(contract.id, input);
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
        <DialogTitle>{isNew ? '新建合同' : '编辑合同'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4 max-h-[70vh] overflow-y-auto pr-1">
        {prefill && (
          <div className="rounded-md border border-amber-300/60 bg-amber-50 px-3 py-2 text-xs text-amber-800 dark:border-amber-700/50 dark:bg-amber-950/40 dark:text-amber-300">
            已从文档{prefill.fromDoc ? `「${prefill.fromDoc}」` : ''}自动识别并预填以下字段，请核对后修改保存。
            {!customerId && (prefill.customerCandidates?.length ?? 0) > 0 && (
              <div className="mt-1">文档中的单位：{prefill.customerCandidates!.join(' / ')}（未匹配到已有客户，请手动选择）</div>
            )}
          </div>
        )}
        {/* ── 核心字段 ── */}
        <div className="space-y-2">
          <Label>客户 <span className="text-destructive">*</span></Label>
          <select
            value={customerId}
            onChange={(e) => setCustomerId(e.target.value)}
            className={SELECT_CLASS}
          >
            <option value="">请选择客户</option>
            {customers.map((c) => (
              <option key={c.id} value={c.id}>
                {c.user_name}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label>套餐 <span className="text-destructive">*</span></Label>
          <select
            value={packageId}
            onChange={(e) => setPackageId(e.target.value)}
            className={SELECT_CLASS}
          >
            <option value="">请选择套餐</option>
            {packages.map((p) => (
              <option key={p.id} value={p.id}>
                {p.package_name}
              </option>
            ))}
          </select>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>购电起始月 <span className="text-destructive">*</span></Label>
            <Input
              type="month"
              value={startMonth}
              onChange={(e) => setStartMonth(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label>购电结束月 <span className="text-destructive">*</span></Label>
            <Input
              type="month"
              value={endMonth}
              onChange={(e) => setEndMonth(e.target.value)}
            />
          </div>
        </div>

        {/* ── 扩展字段（从文档预填，可修改）── */}
        <div className="rounded-lg border bg-muted/30 p-3 space-y-3">
          <p className="text-xs font-medium text-muted-foreground">📋 文档识别信息（可修改）</p>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs">购电量 (MWh)</Label>
              <Input
                type="number"
                value={energy}
                onChange={(e) => setEnergy(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">绿电占比 (%)</Label>
              <Input
                type="number"
                value={greenRatio}
                onChange={(e) => setGreenRatio(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">合同编号</Label>
              <Input
                value={contractNo}
                onChange={(e) => setContractNo(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">签订日期</Label>
              <Input
                value={signDate}
                onChange={(e) => setSignDate(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">电价 (元/MWh)</Label>
              <Input
                type="number"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">合同金额 (万元)</Label>
              <Input
                type="number"
                value={totalAmount}
                onChange={(e) => setTotalAmount(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">电压等级</Label>
              <Input
                value={voltageLevel}
                onChange={(e) => setVoltageLevel(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">结算方式</Label>
              <Input
                value={settlementMethod}
                onChange={(e) => setSettlementMethod(e.target.value)}
                placeholder="自动识别"
                className="h-8 text-sm"
              />
            </div>
          </div>
          {prefill?.partyA && (
            <div className="flex gap-4 text-xs text-muted-foreground">
              <span>甲方：{prefill.partyA}</span>
              {prefill.partyB && <span>乙方：{prefill.partyB}</span>}
            </div>
          )}
          {/* 提取字段参考（全部） */}
          {prefill?.allFields && prefill.allFields.length > 0 && (
            <div>
              <button
                className="text-xs text-primary hover:underline"
                onClick={() => setShowAllFields((v) => !v)}
              >
                {showAllFields ? '收起' : `查看全部 ${prefill.allFields!.length} 个识别字段`}
              </button>
              {showAllFields && (
                <div className="mt-2 max-h-40 overflow-auto rounded border bg-background p-2">
                  <table className="w-full text-xs">
                    <tbody>
                      {prefill.allFields.map((f, i) => (
                        <tr key={i} className="border-b border-border/30">
                          <td className="py-1 pr-2 text-muted-foreground whitespace-nowrap">
                            {f.label || f.key}
                          </td>
                          <td className="py-1 break-all">{f.value}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="space-y-2">
          <Label>状态</Label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={SELECT_CLASS}
          >
            <option value="active">生效中</option>
            <option value="expired">已到期</option>
            <option value="terminated">已终止</option>
          </select>
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
