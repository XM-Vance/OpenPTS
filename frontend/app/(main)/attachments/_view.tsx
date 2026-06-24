'use client';

// 通用附件浏览：先选业务资源（客户/合同），再列出该资源的附件并支持上传。
import { useState, useMemo, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Pie, PieChart, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { ChartContainer } from '@/components/charts/chart-container';
import { AttachmentPanel } from '@/components/attachments/attachment-panel';
import { Input } from '@/components/ui/input';
import { listCustomers } from '@/lib/api/customers';
import { listContracts } from '@/lib/api/retail';
import { cn } from '@/lib/utils';
import { Trash2, Search, Check, ChevronsUpDown } from 'lucide-react';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

type ResourceKind = 'customers' | 'retail_contracts';

/* ── Mock storage stats by business type ── */
const STORAGE_STATS = [
  { name: '客户资质', value: 2340, fill: '#6366f1' },
  { name: '合同 PDF', value: 1850, fill: '#3b82f6' },
  { name: '报表归档', value: 1230, fill: '#10b981' },
  { name: '结算凭证', value: 780, fill: '#f97316' },
  { name: '其他', value: 420, fill: '#8b5cf6' },
];

function fmtMB(kb: number): string {
  if (kb < 1024) return `${kb} KB`;
  return `${(kb / 1024).toFixed(1)} MB`;
}

export default function AttachmentsPage() {
  const [kind, setKind] = useState<ResourceKind>('customers');
  const [resourceId, setResourceId] = useState('');

  const { data: customers } = useQuery({
    queryKey: ['customers', 'attachments-picker'],
    queryFn: () => listCustomers({ limit: 200 }),
    enabled: kind === 'customers',
  });

  const { data: contracts } = useQuery({
    queryKey: ['retail-contracts', 'attachments-picker'],
    queryFn: () => listContracts({ keyword: '' }),
    enabled: kind === 'retail_contracts',
  });

  const options =
    kind === 'customers'
      ? (customers?.items ?? []).map((c) => ({ id: c.id, label: c.user_name }))
      : (contracts ?? []).map((c) => ({
          id: c.id,
          label: `${c.customer_name} · ${c.package_name_snapshot}`,
        }));

  // 自动选第一项：放 useEffect 里（渲染期内 setTimeout setState 是反模式）
  useEffect(() => {
    if (!resourceId && options.length > 0) {
      setResourceId(options[0].id);
    }
  }, [resourceId, options]);

  const totalKB = STORAGE_STATS.reduce((s, d) => s + d.value, 0);

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">附件管理</h1>
        <p className="text-sm text-muted-foreground">
          统一上传/查看业务文档：合同 PDF / 客户资质扫描件 / 报表归档等
        </p>
      </div>

      {/* ═══════════ Storage Pie Chart ═══════════ */}
      <div className="grid gap-4 lg:grid-cols-2">
        <ChartContainer title="存储用量统计（按业务类型）">
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie
                data={STORAGE_STATS}
                cx="50%"
                cy="50%"
                innerRadius={55}
                outerRadius={95}
                dataKey="value"
                nameKey="name"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                isAnimationActive={false}
              >
                {STORAGE_STATS.map((entry, i) => (
                  <Cell key={i} fill={entry.fill} />
                ))}
              </Pie>
              <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number) => fmtMB(v)} />
              <Legend wrapperStyle={{ fontSize: 12 }} />
            </PieChart>
          </ResponsiveContainer>
        </ChartContainer>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">存储概览</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between rounded-md border p-3">
              <span className="text-sm text-muted-foreground">总存储量</span>
              <span className="text-lg font-bold">{fmtMB(totalKB)}</span>
            </div>
            {STORAGE_STATS.map((s) => (
              <div key={s.name} className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-sm" style={{ backgroundColor: s.fill }} />
                  <span>{s.name}</span>
                </div>
                <span className="text-muted-foreground">{fmtMB(s.value)}</span>
              </div>
            ))}
            <div className="pt-2">
              <Button variant="outline" className="w-full" onClick={() => alert('过期附件清理功能（占位）')}>
                <Trash2 className="mr-2 h-4 w-4" />
                清理过期附件
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">选择资源</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-2">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">资源类型</label>
            <select
              value={kind}
              onChange={(e) => {
                setKind(e.target.value as ResourceKind);
                setResourceId('');
              }}
              className={SELECT_CLASS}
            >
              <option value="customers">客户档案</option>
              <option value="retail_contracts">零售合同</option>
            </select>
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">资源</label>
            <ResourceCombobox
              options={options}
              value={resourceId}
              onChange={setResourceId}
            />
          </div>
        </CardContent>
      </Card>

      {resourceId && (
        <AttachmentPanel
          resource={kind}
          resourceId={resourceId}
          title={`附件（${kind === 'customers' ? '客户' : '合同'} ${resourceId.slice(0, 8)}...）`}
        />
      )}
    </div>
  );
}

/* ── 可搜索的资源选择器：点开后顶部搜索框，按名称实时筛选 ── */
function ResourceCombobox({
  options,
  value,
  onChange,
  placeholder = '请选择...',
}: {
  options: { id: string; label: string }[];
  value: string;
  onChange: (id: string) => void;
  placeholder?: string;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((o) => o.id === value);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((o) => o.label.toLowerCase().includes(q));
  }, [options, query]);

  const close = () => {
    setOpen(false);
    setQuery('');
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={cn(SELECT_CLASS, 'w-full cursor-pointer items-center justify-between')}
      >
        <span className={cn('truncate', !selected && 'text-muted-foreground')}>
          {selected ? selected.label : placeholder}
        </span>
        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 text-muted-foreground" />
      </button>

      {open && (
        <>
          {/* 点击外部关闭 */}
          <div className="fixed inset-0 z-40" onClick={close} />
          <div className="absolute z-50 mt-1 w-full rounded-lg border bg-popover p-0 shadow-lg">
            <div className="border-b p-2">
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                <Input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="搜索名称..."
                  className="h-8 pl-7 text-sm"
                  autoFocus
                />
              </div>
            </div>
            <div className="max-h-[250px] overflow-y-auto p-1">
              {filtered.length === 0 ? (
                <p className="py-3 text-center text-sm text-muted-foreground">未找到匹配的资源</p>
              ) : (
                filtered.map((o) => (
                  <button
                    key={o.id}
                    type="button"
                    onClick={() => {
                      onChange(o.id);
                      close();
                    }}
                    className={cn(
                      'flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-sm hover:bg-accent',
                      o.id === value && 'bg-accent',
                    )}
                  >
                    <span className="truncate">{o.label}</span>
                    {o.id === value && <Check className="ml-2 h-3.5 w-3.5 shrink-0" />}
                  </button>
                ))
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
