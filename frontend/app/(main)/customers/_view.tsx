'use client';

import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DemoBadge } from '@/components/feedback';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Dialog, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { AttachmentPanel } from '@/components/attachments/attachment-panel';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createCustomer,
  deleteCustomer,
  listCustomers,
  updateCustomer,
  type Customer,
} from '@/lib/api/customers';
import { listCustomFields, type CustomField } from '@/lib/api/custom-fields';
import { listTags as fetchTags, type Tag } from '@/lib/api/tags';
import {
  LayoutGrid,
  List,
  Star,
  Building2,
  Zap,
  TrendingUp,
} from 'lucide-react';
import dynamic from 'next/dynamic';

// 图表组件懒加载（recharts 从首屏 JS 剥离，与项目既有 _charts 约定一致）。
const CustomerBubbleChart = dynamic(() => import('./_charts').then((m) => m.CustomerBubbleChart), { ssr: false });
const LifecycleFunnel = dynamic(() => import('./_charts').then((m) => m.LifecycleFunnel), { ssr: false });

/* ── View mode toggle ── */
type ViewMode = 'table' | 'cards';

// 客户 360° 视图相关类型与常量已拆至 _shared.ts（供 _charts.tsx 共用）。
import {
  STATUS_NAMES,
  STATUS_LABELS,
  STATUS_COLORS,
  type Customer360,
} from './_shared';

function enrichCustomers(items: Customer[]): Customer360[] {
  const industries = ['工业', '商业', '数据中心', '新能源', '化工', '冶金', '建材', '市政'];

  return items.map((c, i) => {
    const industry = industries[i % industries.length];
    const electricity = 50 + Math.floor(Math.random() * 950);
    const revenue = Math.round((electricity * (0.3 + Math.random() * 0.4)) / 10);
    const risk = 1 + Math.floor(Math.random() * 5);
    const rating = 1 + Math.floor(Math.random() * 5);
    const status = STATUS_NAMES[i % STATUS_NAMES.length];
    return {
      id: c.id,
      name: c.user_name,
      industry,
      electricity,
      revenue,
      risk,
      rating,
      status,
      cooperation: STATUS_LABELS[status],
    };
  });
}

/* ── Stars renderer ── */
function Stars({ count, color = '#f59e0b' }: { count: number; color?: string }) {
  return (
    <span className="inline-flex gap-0.5">
      {Array.from({ length: 5 }).map((_, i) => (
        <Star
          key={i}
          className="h-3.5 w-3.5"
          style={{ color: i < count ? color : '#e2e8f0' }}
          fill={i < count ? color : 'none'}
        />
      ))}
    </span>
  );
}

/* ── Status badge ── */
function StatusBadge({ status }: { status: Customer360['status'] }) {
  const colors: Record<string, string> = {
    potential: 'bg-slate-100 text-slate-600',
    interested: 'bg-blue-100 text-blue-700',
    contracted: 'bg-emerald-100 text-emerald-700',
    renewed: 'bg-indigo-100 text-indigo-700',
    churned: 'bg-red-100 text-red-700',
  };
  const labels: Record<string, string> = {
    potential: '潜在',
    interested: '意向',
    contracted: '签约',
    renewed: '续约',
    churned: '流失',
  };
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colors[status]}`}>
      {labels[status]}
    </span>
  );
}

/* ── Customer Card View ── */
function CustomerCards({ customers360 }: { customers360: Customer360[] }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {customers360.map((c) => (
        <Card key={c.id} className="flex flex-col">
          <CardHeader className="pb-2">
            <div className="flex items-start justify-between">
              <CardTitle className="flex items-center gap-2 text-sm">
                <Building2 className="h-4 w-4 text-indigo-500" />
                <span className="truncate">{c.name}</span>
              </CardTitle>
              <StatusBadge status={c.status} />
            </div>
          </CardHeader>
          <CardContent className="flex flex-1 flex-col gap-2 text-sm">
            <div className="flex items-center gap-1.5">
              <Badge variant="outline" className="text-xs">{c.industry}</Badge>
            </div>
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <Zap className="h-3.5 w-3.5" />
              <span>用电量: {c.electricity} MWh/月</span>
            </div>
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <TrendingUp className="h-3.5 w-3.5" />
              <span>收益: ¥{c.revenue} 万/月</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground">信用评级:</span>
              <Stars count={c.rating} />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

/* ── Main Page ── */
export default function CustomersPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('customer_management:write');
  const canDelete = has('customer_management:delete');
  const canViewSensitive = has('customer_management:view_sensitive');

  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // Derive all pagination / sorting state from URL searchParams
  const page = Number(searchParams.get('page') ?? '1') || 1;
  const pageSize = Number(searchParams.get('pageSize') ?? '20') || 20;
  const sortBy = searchParams.get('sort_by') ?? '';
  const sortOrder = (searchParams.get('sort_order') as 'asc' | 'desc') || 'asc';
  const search = searchParams.get('search') ?? '';
  const tagFilter = searchParams.get('tag') ?? '';

  const highlightId = searchParams.get('highlight');
  const highlightRef = useRef<HTMLTableRowElement | null>(null);
  const [editing, setEditing] = useState<Customer | 'new' | null>(null);
  const [filesFor, setFilesFor] = useState<Customer | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('table');

  // Local keyword input (not committed to URL until search)
  const [keyword, setKeyword] = useState(search);

  // Helper: build a new URL by merging current params with updates
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

  // Fetch tag definitions for filter + display
  const { data: tagDefsData } = useQuery({
    queryKey: ['tags', 'customer'],
    queryFn: () => fetchTags('customer'),
  });
  const tagDefs = useMemo(() => tagDefsData?.items ?? [], [tagDefsData]);
  const tagColorMap = useMemo(() => {
    const m: Record<string, string> = {};
    tagDefs.forEach((t) => { m[t.name] = t.color; });
    return m;
  }, [tagDefs]);

  const { data, isLoading } = useQuery({
    queryKey: ['customers', search, tagFilter, page, pageSize, sortBy, sortOrder],
    queryFn: () =>
      listCustomers({
        keyword: search || undefined,
        tag: tagFilter || undefined,
        page,
        page_size: pageSize,
        sort_by: sortBy || undefined,
        sort_order: sortBy ? sortOrder : undefined,
      }),
  });

  const items = useMemo(() => data?.items ?? [], [data]);
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  // Enrich with 360° data
  const customers360 = useMemo(() => enrichCustomers(items), [items]);

  // Scroll to highlighted row
  useEffect(() => {
    if (highlightId && highlightRef.current) {
      highlightRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, [highlightId, data]);

  const onDelete = async (id: string, name: string) => {
    if (!window.confirm(`确认删除客户「${name}」？`)) return;
    try {
      await deleteCustomer(id);
      qc.invalidateQueries({ queryKey: ['customers'] });
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  const toggleSort = (col: string) => {
    const newOrder = sortBy === col && sortOrder === 'asc' ? 'desc' : 'asc';
    setParams({ sort_by: col, sort_order: newOrder, page: 1 });
  };

  const sortIcon = (col: string) =>
    sortBy === col ? (sortOrder === 'asc' ? ' ↑' : ' ↓') : '';

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">客户档案管理</h1>
        <div className="flex items-center gap-2">
          {/* View toggle */}
          <div className="flex rounded-lg border bg-white p-0.5 shadow-sm">
            <Button
              size="sm"
              variant={viewMode === 'table' ? 'default' : 'ghost'}
              className="h-7 gap-1 px-3 text-xs"
              onClick={() => setViewMode('table')}
            >
              <List className="h-3.5 w-3.5" />
              列表
            </Button>
            <Button
              size="sm"
              variant={viewMode === 'cards' ? 'default' : 'ghost'}
              className="h-7 gap-1 px-3 text-xs"
              onClick={() => setViewMode('cards')}
            >
              <LayoutGrid className="h-3.5 w-3.5" />
              卡片
            </Button>
          </div>
          {canWrite && <Button onClick={() => setEditing('new')}>新建客户</Button>}
        </div>
      </div>

      <div className="flex gap-2">
        <Input
          placeholder="搜索客户名 / 简称"
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
          value={tagFilter}
          onChange={(e) => setParams({ tag: e.target.value, page: 1 })}
          className="h-9 max-w-[160px] rounded-md border border-input bg-background px-3 text-sm"
        >
          <option value="">全部标签</option>
          {tagDefs.map((t) => (
            <option key={t.id} value={t.name}>{t.name}</option>
          ))}
        </select>
      </div>

      {/* 360° View: Charts (always visible) */}
      {viewMode === 'cards' && (
        <>
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <CustomerBubbleChart data={customers360} />
            <LifecycleFunnel data={customers360} />
          </div>
          <CustomerCards customers360={customers360} />
        </>
      )}

      {/* Customer table with server-side sort & pagination */}
      {viewMode === 'table' && (
        <>
          {/* Charts above table in table mode too */}
          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <CustomerBubbleChart data={customers360} />
            <LifecycleFunnel data={customers360} />
          </div>

          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead
                    className="cursor-pointer"
                    onClick={() => toggleSort('user_name')}
                  >
                    客户名称{sortIcon('user_name')}
                  </TableHead>
                  <TableHead
                    className="cursor-pointer"
                    onClick={() => toggleSort('short_name')}
                  >
                    简称{sortIcon('short_name')}
                  </TableHead>
                  <TableHead
                    className="cursor-pointer"
                    onClick={() => toggleSort('location')}
                  >
                    所在地{sortIcon('location')}
                  </TableHead>
                  <TableHead
                    className="cursor-pointer"
                    onClick={() => toggleSort('manager')}
                  >
                    客户经理{sortIcon('manager')}
                  </TableHead>
                  <TableHead>标签</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-muted-foreground">
                      加载中...
                    </TableCell>
                  </TableRow>
                )}
                {items.map((cust) => (
                  <TableRow
                    key={cust.id}
                    ref={highlightId === cust.id ? highlightRef : undefined}
                    className={
                      highlightId === cust.id ? 'bg-amber-50 ring-2 ring-amber-300' : ''
                    }
                  >
                    <TableCell className="font-medium">
                      {cust.user_name}
                      {!canViewSensitive && (
                        <span className="ml-1 text-xs text-muted-foreground" title="无查看敏感信息权限">
                          🔒
                        </span>
                      )}
                      {cust.is_demo && (
                        <Badge variant="secondary" className="ml-2">
                          演示
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>{cust.short_name || '-'}</TableCell>
                    <TableCell>{cust.location || '-'}</TableCell>
                    <TableCell>{cust.manager || '-'}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {cust.tags?.map((t) => {
                          const color = tagColorMap[t];
                          return color ? (
                            <Badge key={t} variant="outline" style={{ borderColor: color, color }}>
                              {t}
                            </Badge>
                          ) : (
                            <Badge key={t} variant="outline">{t}</Badge>
                          );
                        })}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button size="sm" variant="ghost" onClick={() => setFilesFor(cust)}>
                          附件
                        </Button>
                        {canWrite && (
                          <Button size="sm" variant="ghost" onClick={() => setEditing(cust)}>
                            编辑
                          </Button>
                        )}
                        {canDelete && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => onDelete(cust.id, cust.user_name)}
                          >
                            删除
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {total === 0 && !isLoading && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-muted-foreground">
                      暂无数据
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
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
        </>
      )}

      {editing && (
        <CustomerFormDialog
          customer={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            qc.invalidateQueries({ queryKey: ['customers'] });
          }}
        />
      )}

      {filesFor && (
        <Dialog open onClose={() => setFilesFor(null)}>
          <DialogHeader>
            <DialogTitle>客户附件 · {filesFor.user_name}</DialogTitle>
          </DialogHeader>
          <AttachmentPanel
            resource="customers"
            resourceId={filesFor.id}
            canWrite={canWrite}
            canDelete={canDelete}
            title="客户相关文档"
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setFilesFor(null)}>
              关闭
            </Button>
          </DialogFooter>
        </Dialog>
      )}
    </div>
  );
}

function CustomerFormDialog({
  customer,
  onClose,
  onSaved,
}: {
  customer: Customer | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !customer;
  const [userName, setUserName] = useState(customer?.user_name ?? '');
  const [shortName, setShortName] = useState(customer?.short_name ?? '');
  const [location, setLocation] = useState(customer?.location ?? '');
  const [source, setSource] = useState(customer?.source ?? '');
  const [manager, setManager] = useState(customer?.manager ?? '');
  const [isDemo, setIsDemo] = useState(customer?.is_demo ?? false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // ── Tag selector state ──
  const [selectedTags, setSelectedTags] = useState<string[]>(customer?.tags ?? []);
  const [tagInput, setTagInput] = useState('');
  const { data: tagDefsData } = useQuery({
    queryKey: ['tags', 'customer'],
    queryFn: () => fetchTags('customer'),
  });
  const availableTags = tagDefsData?.items ?? [];

  // ── Custom fields state ──
  const { data: customFieldsData } = useQuery({
    queryKey: ['custom-fields', 'customer'],
    queryFn: () => listCustomFields('customer'),
  });
  const customFields: CustomField[] = customFieldsData?.items ?? [];
  const customerExtra = useMemo(() => {
    if (customer?.extra && typeof customer.extra === 'object') {
      return customer.extra as Record<string, unknown>;
    }
    return {};
  }, [customer]);
  const [extra, setExtra] = useState<Record<string, unknown>>(customerExtra);

  const toggleTag = (name: string) => {
    setSelectedTags((prev) =>
      prev.includes(name) ? prev.filter((t) => t !== name) : [...prev, name],
    );
  };

  const addCustomTag = () => {
    const name = tagInput.trim();
    if (name && !selectedTags.includes(name)) {
      setSelectedTags((prev) => [...prev, name]);
    }
    setTagInput('');
  };

  const submit = async () => {
    setError(null);
    setSubmitting(true);
    try {
      const input = {
        user_name: userName,
        short_name: shortName || undefined,
        location: location || undefined,
        source: source || undefined,
        manager: manager || undefined,
        tags: selectedTags,
        is_demo: isDemo,
        extra: Object.keys(extra).length > 0 ? extra : undefined,
      };
      if (isNew) {
        await createCustomer(input);
      } else if (customer) {
        await updateCustomer(customer.id, input);
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
        <DialogTitle>{isNew ? '新建客户' : '编辑客户'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>客户名称</Label>
          <Input value={userName} onChange={(e) => setUserName(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>简称</Label>
          <Input value={shortName} onChange={(e) => setShortName(e.target.value)} />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>所在地</Label>
            <Input value={location} onChange={(e) => setLocation(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label>客户经理</Label>
            <Input value={manager} onChange={(e) => setManager(e.target.value)} />
          </div>
        </div>
        <div className="space-y-2">
          <Label>来源</Label>
          <Input value={source} onChange={(e) => setSource(e.target.value)} />
        </div>

        {/* ── Tag selector ── */}
        <div className="space-y-2">
          <Label>标签</Label>
          {selectedTags.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-1">
              {selectedTags.map((name) => {
                const def = availableTags.find((t) => t.name === name);
                const color = def?.color;
                return (
                  <Badge
                    key={name}
                    variant="outline"
                    style={color ? { borderColor: color, color } : undefined}
                    className="cursor-pointer"
                    onClick={() => toggleTag(name)}
                  >
                    {name} ✕
                  </Badge>
                );
              })}
            </div>
          )}
          <div className="flex gap-1">
            <Input
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addCustomTag(); } }}
              placeholder="输入标签名或点击下方选择"
              className="text-sm"
            />
            {tagInput.trim() && (
              <Button type="button" size="sm" variant="outline" onClick={addCustomTag}>添加</Button>
            )}
          </div>
          {availableTags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {availableTags
                .filter((t) => !selectedTags.includes(t.name))
                .map((t) => (
                  <Badge
                    key={t.id}
                    variant="outline"
                    style={{ borderColor: t.color, color: t.color }}
                    className="cursor-pointer"
                    onClick={() => toggleTag(t.name)}
                  >
                    + {t.name}
                  </Badge>
                ))}
            </div>
          )}
        </div>

        {/* ── Custom fields ── */}
        {customFields.length > 0 && (
          <div className="space-y-3 border-t pt-3">
            <p className="text-sm font-medium text-muted-foreground">自定义字段</p>
            {customFields.map((field) => {
              const val = extra[field.field_key];
              const setVal = (v: unknown) =>
                setExtra((prev) => ({ ...prev, [field.field_key]: v }));
              if (field.field_type === 'boolean') {
                return (
                  <label key={field.id} className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={!!val}
                      onChange={(e) => setVal(e.target.checked)}
                    />
                    {field.field_label}
                  </label>
                );
              }
              if (field.field_type === 'select') {
                return (
                  <div key={field.id} className="space-y-1">
                    <Label className="text-sm">{field.field_label}</Label>
                    <select
                      value={(val as string) ?? ''}
                      onChange={(e) => setVal(e.target.value || undefined)}
                      className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"
                    >
                      <option value="">请选择</option>
                      {(field.options ?? []).map((opt) => (
                        <option key={opt} value={opt}>{opt}</option>
                      ))}
                    </select>
                  </div>
                );
              }
              if (field.field_type === 'multi_select') {
                const selected = Array.isArray(val) ? val as string[] : [];
                return (
                  <div key={field.id} className="space-y-1">
                    <Label className="text-sm">{field.field_label}</Label>
                    <div className="flex flex-wrap gap-1">
                      {(field.options ?? []).map((opt) => (
                        <Badge
                          key={opt}
                          variant={selected.includes(opt) ? 'default' : 'outline'}
                          className="cursor-pointer"
                          onClick={() => {
                            if (selected.includes(opt)) {
                              setVal(selected.filter((v) => v !== opt));
                            } else {
                              setVal([...selected, opt]);
                            }
                          }}
                        >
                          {opt}
                        </Badge>
                      ))}
                    </div>
                  </div>
                );
              }
              return (
                <div key={field.id} className="space-y-1">
                  <Label className="text-sm">{field.field_label}</Label>
                  <Input
                    type={field.field_type === 'number' ? 'number' : field.field_type === 'date' ? 'date' : 'text'}
                    value={(val as string | number) ?? ''}
                    onChange={(e) => setVal(e.target.value || undefined)}
                    placeholder={field.default_value ?? ''}
                  />
                </div>
              );
            })}
          </div>
        )}

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isDemo}
            onChange={(e) => setIsDemo(e.target.checked)}
          />
          演示客户（脱敏展示）
        </label>
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
