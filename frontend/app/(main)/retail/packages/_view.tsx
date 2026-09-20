'use client';

import { useState, useMemo, type ReactNode } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts';
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
import { Dropdown } from '@/components/ui/dropdown';
import { MobileCardList, type MobileCardField, type MobileRow } from '@/components/data-display/mobile-card-list';
import { useIsMobile } from '@/hooks/use-mobile';
import { RetailTabs } from '@/components/retail/retail-tabs';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createPackage,
  deletePackage,
  listPackages,
  listPricingModels,
  settlePreview,
  settleAndSave,
  updatePackage,
  type PricingModel,
  type RetailPackage,
} from '@/lib/api/retail';
import { listCustomers } from '@/lib/api/customers';
import { PricingEditor, type PricingConfig } from '@/components/retail/pricing-editor';
import { ChartLoading, EmptyState } from '@/components/feedback';
import { Skeleton } from '@/components/ui/skeleton';

const STATUS_LABEL: Record<string, string> = {
  active: '启用',
  archived: '已归档',
  draft: '草稿',
};

const money = (v: number) => v.toLocaleString('zh-CN', { maximumFractionDigits: 2 });

// 分时电价模拟数据（基于套餐类型）——仅用于「套餐分时电价对比」图的粗略可视化。
function getMockTOUPrices(type: string): { period: string; price: number }[] {
  const base: Record<string, number[]> = {
    '分时': [0.35, 0.65, 1.05],
    '月度': [0.55, 0.55, 0.55],
    '月内': [0.50, 0.60, 0.80],
    '年度': [0.48, 0.48, 0.48],
  };
  const prices = base[type] ?? [0.50, 0.60, 0.70];
  return [
    { period: '谷时', price: prices[0] },
    { period: '平时', price: prices[1] },
    { period: '峰时', price: prices[2] },
  ];
}

export default function RetailPackagesPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('retail_management:write');
  const canDelete = has('retail_management:delete');
  const canSettle = has('settlement_management:write'); // 落库走结算模块写权限

  const [editing, setEditing] = useState<RetailPackage | 'new' | null>(null);
  const isMobile = useIsMobile();
  // 移动端套餐卡片字段 + actions（编辑主按钮，删除收「⋯」）
  const pkgMobileFields: MobileCardField<MobileRow>[] = [
    {
      primary: true,
      render: (p) => (
        <span>
          <span className="font-medium">{(p as any).package_name}</span>
          <Badge variant={(p as any).status === 'active' ? 'default' : 'secondary'} className="ml-2">{STATUS_LABEL[(p as any).status] || (p as any).status}</Badge>
        </span>
      ),
    },
    { label: '类型', render: (p) => (p as any).package_type },
    { label: '定价模型', render: (p) => (p as any).model_code || '-' },
    { label: '绿电', render: (p) => ((p as any).is_green_power ? <Badge variant="success">绿电</Badge> : '-') },
  ];
  // 收益模拟器状态
  const [simKwh, setSimKwh] = useState('10000');
  const [simPkg, setSimPkg] = useState('');
  const [simMarketPrice, setSimMarketPrice] = useState(''); // 市场参考价（market_spread 模式用）
  const [simWholesale, setSimWholesale] = useState(''); // 批发均价（填了才算收益）
  const [simBenchmark, setSimBenchmark] = useState(''); // P基准 元/kWh（价差分享套餐 + 封顶价用）
  const [simPurchase, setSimPurchase] = useState(''); // P购电均价 元/kWh（价差分享套餐用）
  const [simCustomer, setSimCustomer] = useState(''); // 落库为测算时归属的客户
  const [simMonth, setSimMonth] = useState(() => new Date().toISOString().slice(0, 7)); // YYYY-MM
  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState<string | null>(null);

  const { data: packages, isLoading } = useQuery({
    queryKey: ['retail-packages', ''],
    queryFn: () => listPackages(),
  });
  const { data: models } = useQuery({
    queryKey: ['pricing-models'],
    queryFn: listPricingModels,
  });

  const onDelete = async (id: string, name: string) => {
    if (!window.confirm(`确认删除套餐「${name}」？`)) return;
    try {
      await deletePackage(id);
      qc.invalidateQueries({ queryKey: ['retail-packages'] });
    } catch (e) {
      window.alert(extractErrorMessage(e));
    }
  };

  const pkgList = useMemo(() => packages ?? [], [packages]);

  // ── 套餐对比柱状图 ──
  const comparisonData = useMemo(() => {
    const activePkgs = pkgList.filter((p) => p.status === 'active').slice(0, 6);
    if (activePkgs.length === 0) return [];

    const periods = ['谷时', '平时', '峰时'];
    return periods.map((period, pi) => {
      const row: Record<string, string | number> = { period };
      for (const pkg of activePkgs) {
        const prices = getMockTOUPrices(pkg.package_type);
        row[pkg.package_name.length > 6 ? pkg.package_name.slice(0, 6) + '…' : pkg.package_name] = prices[pi].price;
      }
      return row;
    });
  }, [pkgList]);

  const activePkgNames = useMemo(
    () => pkgList.filter((p) => p.status === 'active').slice(0, 6).map((p) =>
      p.package_name.length > 6 ? p.package_name.slice(0, 6) + '…' : p.package_name,
    ),
    [pkgList],
  );

  // ── 收益试算（调 P0 结算引擎 /settlement/preview，按套餐真实定价计算）──
  const simKwhNum = Number(simKwh);
  const { data: simPreview, isFetching: simLoading, error: simError } = useQuery({
    queryKey: ['settle-preview', simPkg, simKwh, simMarketPrice, simWholesale, simBenchmark, simPurchase],
    queryFn: () =>
      settlePreview({
        package_id: simPkg,
        energy_kwh: simKwhNum,
        market_price: simMarketPrice ? Number(simMarketPrice) : undefined,
        benchmark_price: simBenchmark ? Number(simBenchmark) : undefined,
        purchase_avg_price: simPurchase ? Number(simPurchase) : undefined,
        wholesale_avg_price: simWholesale ? Number(simWholesale) : undefined,
      }),
    enabled: !!simPkg && simKwhNum > 0,
  });

  const { data: custData } = useQuery({
    queryKey: ['customers', 'sim-picker'],
    queryFn: () => listCustomers({ limit: 200 }),
  });
  const custList = useMemo(() => custData?.items ?? [], [custData]);

  // 把当前试算结果存为该客户该月的「测算」（is_estimate=true）。
  const canSaveEstimate = !!simPkg && !!simCustomer && !!simMonth && !!simWholesale && !!simPreview?.profit;
  const onSaveEstimate = async () => {
    if (!canSaveEstimate) return;
    setSaving(true);
    setSaveMsg(null);
    try {
      const res = await settleAndSave({
        package_id: simPkg,
        customer_id: simCustomer,
        operating_month: simMonth,
        energy_kwh: simKwhNum,
        market_price: simMarketPrice ? Number(simMarketPrice) : undefined,
        benchmark_price: simBenchmark ? Number(simBenchmark) : undefined,
        purchase_avg_price: simPurchase ? Number(simPurchase) : undefined,
        wholesale_avg_price: simWholesale ? Number(simWholesale) : undefined,
        is_estimate: true,
      });
      setSaveMsg(res.message || '已保存为测算');
      qc.invalidateQueries({ queryKey: ['customer-profit'] });
    } catch (e) {
      setSaveMsg('保存失败：' + extractErrorMessage(e));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">零售管理</h1>
        <p className="text-sm text-muted-foreground">零售合同、套餐与定价</p>
      </div>
      <RetailTabs />

      <div className="flex justify-end pt-2">
        {canWrite && <Button onClick={() => setEditing('new')}>新建套餐</Button>}
      </div>

      {/* ── 套餐对比柱状图 ── */}
      <ChartContainer title="套餐分时电价对比（元/kWh）" minHeight={300}>
        {comparisonData.length > 0 ? (
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={comparisonData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="period" tick={{ fontSize: 12, fill: '#374151' }} />
              <YAxis tick={{ fontSize: 12 }} width={50} unit="元" />
              <RechartsTooltip
                formatter={(v: number) => [`${v.toFixed(4)} 元/kWh`]}
                contentStyle={{ fontSize: 12 }}
              />
              <Legend wrapperStyle={{ fontSize: 11 }} />
              {activePkgNames.map((name, idx) => {
                const colors = ['#2563eb', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];
                return (
                  <Bar key={name} dataKey={name} fill={colors[idx % colors.length]} isAnimationActive={false} />
                );
              })}
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <EmptyState compact className="h-full" title="暂无启用的套餐数据" />
        )}
      </ChartContainer>

      {/* ── 套餐收益模拟器 ── */}
      <ChartContainer title="套餐收益模拟器" minHeight={200}>
        <div className="space-y-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="space-y-1">
              <Label className="text-sm">月用电量 (kWh)</Label>
              <Input type="number" value={simKwh} onChange={(e) => setSimKwh(e.target.value)} className="w-36" placeholder="如 10000" />
            </div>
            <div className="space-y-1">
              <Label className="text-sm">选择套餐</Label>
              <select
                value={simPkg}
                onChange={(e) => setSimPkg(e.target.value)}
                className="h-9 rounded-md border border-input bg-background px-3 text-sm"
              >
                <option value="">— 选择套餐试算 —</option>
                {pkgList.filter((p) => p.status === 'active').map((p) => (
                  <option key={p.id} value={p.id}>{p.package_name}</option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <Label className="text-sm">市场参考价 (元/kWh)</Label>
              <Input type="number" value={simMarketPrice} onChange={(e) => setSimMarketPrice(e.target.value)} className="w-36" placeholder="价差模式用" />
            </div>
            <div className="space-y-1">
              <Label className="text-sm">批发均价 (元/kWh)</Label>
              <Input type="number" value={simWholesale} onChange={(e) => setSimWholesale(e.target.value)} className="w-36" placeholder="填了算收益" />
            </div>
            <div className="space-y-1">
              <Label className="text-sm">P基准 (元/kWh)</Label>
              <Input type="number" step="0.001" value={simBenchmark} onChange={(e) => setSimBenchmark(e.target.value)} className="w-36" placeholder="封顶/价差分享" />
            </div>
            <div className="space-y-1">
              <Label className="text-sm">P购电均价 (元/kWh)</Label>
              <Input type="number" step="0.001" value={simPurchase} onChange={(e) => setSimPurchase(e.target.value)} className="w-36" placeholder="价差分享用" />
            </div>
          </div>

          {!simPkg || simKwhNum <= 0 ? (
            <p className="text-sm text-muted-foreground">选择具体套餐并填月用电量，按真实定价规则试算零售 / 批发 / 收益。</p>
          ) : simLoading ? (
            <p className="text-sm text-muted-foreground">试算中…</p>
          ) : simError ? (
            <p className="text-sm text-destructive">试算失败：{extractErrorMessage(simError)}</p>
          ) : simPreview ? (
            <div className="space-y-3">
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                <SimStat
                  label="综合单价 (元/kWh)"
                  value={simPreview.retail.avg_unit_price.toFixed(4)}
                  suffix={simPreview.retail.is_capped ? (
                    <Badge variant="warning" className="ml-1 text-xs">已封顶 {simPreview.retail.capped_unit_price?.toFixed(4)}</Badge>
                  ) : null}
                />
                <SimStat label="电费 (元)" value={money(simPreview.retail.energy_amount)} />
                <SimStat label="服务费 (元)" value={money(simPreview.retail.service_amount)} />
                <SimStat label="零售总额 (元)" value={money(simPreview.retail.total_amount)} accent />
              </div>
              {simPreview.profit && simPreview.wholesale && (
                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                  <SimStat label="批发成本 (元)" value={money(simPreview.wholesale.total_cost)} />
                  <SimStat label="购电均价 (元/kWh)" value={simPreview.wholesale.avg_price.toFixed(4)} />
                  <SimStat label="收益 / 毛利 (元)" value={money(simPreview.profit.gross_profit)} accent />
                  <SimStat label="毛利率" value={`${simPreview.profit.gross_margin.toFixed(2)}%`} />
                </div>
              )}
              {simPreview.retail.breakdown?.length > 0 && (
                <div className="rounded-lg border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>分项</TableHead>
                        <TableHead className="text-right">电量 (kWh)</TableHead>
                        <TableHead className="text-right">单价 (元/kWh)</TableHead>
                        <TableHead className="text-right">金额 (元)</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {simPreview.retail.breakdown.map((b, i) => (
                        <TableRow key={i}>
                          <TableCell>{b.label}</TableCell>
                          <TableCell className="text-right">{money(b.energy_kwh)}</TableCell>
                          <TableCell className="text-right">{b.unit_price.toFixed(4)}</TableCell>
                          <TableCell className="text-right">{money(b.amount)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
              <p className="text-[10px] text-muted-foreground">
                由 P0 结算引擎按套餐 pricing_config 实算（decimal 精确）。分时套餐需按峰平谷电量试算，此处仅按总量。
              </p>

              {/* 落库为某客户该月的「测算」（需填批发均价才有收益可存） */}
              {canSettle && (
                <div className="flex flex-wrap items-end gap-3 border-t pt-3">
                  <div className="space-y-1">
                    <Label className="text-sm">归属客户</Label>
                    <select
                      value={simCustomer}
                      onChange={(e) => setSimCustomer(e.target.value)}
                      className="h-9 rounded-md border border-input bg-background px-3 text-sm"
                    >
                      <option value="">— 选择客户 —</option>
                      {custList.map((cst) => (
                        <option key={cst.id} value={cst.id}>{cst.user_name}</option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-1">
                    <Label className="text-sm">结算月份</Label>
                    <Input type="month" value={simMonth} onChange={(e) => setSimMonth(e.target.value)} className="w-40" />
                  </div>
                  <Button onClick={onSaveEstimate} disabled={!canSaveEstimate || saving}>
                    {saving ? '保存中…' : '保存为测算'}
                  </Button>
                  {!simPreview.profit && (
                    <span className="text-xs text-muted-foreground">填「批发均价」后可存为测算（需含收益）</span>
                  )}
                  {saveMsg && <span className="text-xs text-emerald-700">{saveMsg}</span>}
                </div>
              )}
            </div>
          ) : null}
        </div>
      </ChartContainer>

      {isMobile ? (
        isLoading ? (
          <ChartLoading className="py-8" />
        ) : (packages?.length ?? 0) === 0 ? (
          <EmptyState compact className="py-8" title="暂无数据" />
        ) : (
          <MobileCardList
            items={(packages ?? []) as unknown as MobileRow[]}
            itemKey={(p) => String((p as any).id)}
            fields={pkgMobileFields}
            actions={(p) => (
              <>
                {canWrite && (
                  <Button size="sm" variant="ghost" onClick={() => setEditing(p as any)}>编辑</Button>
                )}
                {canDelete && (
                  <Dropdown
                    items={[{ label: '删除', danger: true, onClick: () => onDelete((p as any).id, (p as any).package_name) }]}
                  />
                )}
              </>
            )}
          />
        )
      ) : (
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>套餐名称</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>定价模型</TableHead>
              <TableHead>绿电</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={6}><Skeleton className="h-5 w-full" /></TableCell>
              </TableRow>
            )}
            {packages?.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-medium">{p.package_name}</TableCell>
                <TableCell>{p.package_type}</TableCell>
                <TableCell>{p.model_code || '-'}</TableCell>
                <TableCell>
                  {p.is_green_power ? <Badge variant="success">绿电</Badge> : '-'}
                </TableCell>
                <TableCell>
                  <Badge variant={p.status === 'active' ? 'default' : 'secondary'}>
                    {STATUS_LABEL[p.status] || p.status}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-1">
                    {canWrite && (
                      <Button size="sm" variant="ghost" onClick={() => setEditing(p)}>
                        编辑
                      </Button>
                    )}
                    {canDelete && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => onDelete(p.id, p.package_name)}
                      >
                        删除
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
            {packages?.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={6}><EmptyState compact title="暂无数据" /></TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      )}

      {editing && (
        <PackageFormDialog
          pkg={editing === 'new' ? null : editing}
          models={models ?? []}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            qc.invalidateQueries({ queryKey: ['retail-packages'] });
          }}
        />
      )}
    </div>
  );
}

function PackageFormDialog({
  pkg,
  models,
  onClose,
  onSaved,
}: {
  pkg: RetailPackage | null;
  models: PricingModel[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = !pkg;
  const [name, setName] = useState(pkg?.package_name ?? '');
  const [type, setType] = useState(pkg?.package_type ?? '');
  const [modelCode, setModelCode] = useState(pkg?.model_code ?? '');
  const [isGreen, setIsGreen] = useState(pkg?.is_green_power ?? false);
  const [status, setStatus] = useState(pkg?.status ?? 'active');
  const [description, setDescription] = useState(pkg?.description ?? '');
  const [pricingConfig, setPricingConfig] = useState<PricingConfig>(
    (pkg?.pricing_config as unknown as PricingConfig) ?? { mode: 'tou' as const, service_fee: 0.01 }
  );
  const [showPricing, setShowPricing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const submit = async () => {
    setError(null);
    if (!name.trim()) {
      setError('请填写套餐名称');
      return;
    }
    if (!type.trim()) {
      setError('请填写套餐类型');
      return;
    }
    setSubmitting(true);
    try {
      const input = {
        package_name: name,
        package_type: type,
        model_code: modelCode || undefined,
        is_green_power: isGreen,
        status,
        description: description || undefined,
        pricing_config: showPricing ? (pricingConfig as unknown as Record<string, unknown>) : undefined,
      };
      if (isNew) await createPackage(input);
      else if (pkg) await updatePackage(pkg.id, input);
      onSaved();
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open onClose={onClose} fullScreenOnMobile>
      <DialogHeader>
        <DialogTitle>{isNew ? '新建套餐' : '编辑套餐'}</DialogTitle>
      </DialogHeader>
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>套餐名称</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>套餐类型</Label>
          <Input
            value={type}
            onChange={(e) => setType(e.target.value)}
            placeholder="如 月度 / 月内 / 分时"
          />
        </div>
        <div className="space-y-2">
          <Label>定价模型（可选）</Label>
          <select
            value={modelCode}
            onChange={(e) => setModelCode(e.target.value)}
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm"
          >
            <option value="">（无）</option>
            {models.map((m) => (
              <option key={m.code} value={m.code}>
                {m.display_name}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label>状态</Label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm"
          >
            <option value="active">启用</option>
            <option value="draft">草稿</option>
            <option value="archived">已归档</option>
          </select>
        </div>
        <div className="space-y-2">
          <Label>描述</Label>
          <Input value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>

        {/* ── Pricing config editor ── */}
        <div className="space-y-2">
          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={showPricing}
              onChange={(e) => setShowPricing(e.target.checked)}
            />
            <span className="font-medium">配置定价参数</span>
            {pricingConfig.mode && showPricing && (
              <Badge variant="info" className="ml-1 text-xs">
                {pricingConfig.mode === 'tou' ? '分时电价' :
                 pricingConfig.mode === 'tiered' ? '阶梯电价' :
                 pricingConfig.mode === 'fixed' ? '固定单价' : '市场价差'}
              </Badge>
            )}
          </label>
          {showPricing && (
            <PricingEditor
              value={pricingConfig}
              onChange={setPricingConfig}
              pricingMode={models.find((m) => m.code === modelCode)?.pricing_mode}
            />
          )}
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={isGreen}
            onChange={(e) => setIsGreen(e.target.checked)}
          />
          绿电套餐
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

// SimStat 试算结果小卡片。accent 用于突出总额/收益；suffix 附在值后（如封顶标记）。
function SimStat({
  label,
  value,
  accent,
  suffix,
}: {
  label: string;
  value: string;
  accent?: boolean;
  suffix?: ReactNode;
}) {
  return (
    <div className="rounded-lg border p-3 space-y-1">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={accent ? 'text-lg font-bold text-blue-600' : 'text-base font-semibold'}>
        {value}
        {suffix}
      </p>
    </div>
  );
}
