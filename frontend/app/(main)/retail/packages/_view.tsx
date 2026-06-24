'use client';

import { useState, useMemo } from 'react';
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
import { RetailTabs } from '@/components/retail/retail-tabs';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  createPackage,
  deletePackage,
  listPackages,
  listPricingModels,
  updatePackage,
  type PricingModel,
  type RetailPackage,
} from '@/lib/api/retail';
import { PricingEditor, type PricingConfig } from '@/components/retail/pricing-editor';

const STATUS_LABEL: Record<string, string> = {
  active: '启用',
  archived: '已归档',
  draft: '草稿',
};

// 分时电价模拟数据（基于套餐类型）
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

  const [editing, setEditing] = useState<RetailPackage | 'new' | null>(null);

  // 收益模拟器状态
  const [simKwh, setSimKwh] = useState('10000');
  const [simPkg, setSimPkg] = useState('');

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

  // ── 收益模拟计算 ──
  const simResult = useMemo(() => {
    const kwh = Number(simKwh);
    if (!kwh || kwh <= 0) return null;

    const selectedPkg = simPkg ? pkgList.find((p) => p.id === simPkg) : null;
    const pkgsToCalc = selectedPkg ? [selectedPkg] : pkgList.filter((p) => p.status === 'active').slice(0, 4);

    return pkgsToCalc.map((p) => {
      const tou = getMockTOUPrices(p.package_type);
      // 简化计算：谷30%，平40%，峰30%
      const avgPrice = tou[0].price * 0.3 + tou[1].price * 0.4 + tou[2].price * 0.3;
      const totalCost = avgPrice * kwh;
      return {
        name: p.package_name,
        avgPrice: avgPrice.toFixed(4),
        totalCost: totalCost.toFixed(2),
        type: p.package_type,
      };
    });
  }, [simKwh, simPkg, pkgList]);

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
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="period" tick={{ fontSize: 12, fill: '#374151' }} />
              <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="元" />
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
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            暂无启用的套餐数据
          </div>
        )}
      </ChartContainer>

      {/* ── 套餐收益模拟器 ── */}
      <ChartContainer title="套餐收益模拟器" minHeight={200}>
        <div className="space-y-4">
          <div className="flex items-end gap-4">
            <div className="space-y-1">
              <Label className="text-sm">月用电量 (kWh)</Label>
              <Input
                type="number"
                value={simKwh}
                onChange={(e) => setSimKwh(e.target.value)}
                className="w-40"
                placeholder="如 10000"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-sm">选择套餐（可选）</Label>
              <select
                value={simPkg}
                onChange={(e) => setSimPkg(e.target.value)}
                className="h-9 rounded-md border border-input bg-background px-3 text-sm"
              >
                <option value="">全部活跃套餐</option>
                {pkgList.filter((p) => p.status === 'active').map((p) => (
                  <option key={p.id} value={p.id}>{p.package_name}</option>
                ))}
              </select>
            </div>
          </div>

          {simResult && simResult.length > 0 ? (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {simResult.map((r) => (
                <div key={r.name} className="rounded-lg border p-3 space-y-1">
                  <p className="text-sm font-medium truncate">{r.name}</p>
                  <p className="text-xs text-muted-foreground">类型：{r.type}</p>
                  <p className="text-xs text-muted-foreground">加权均价：{r.avgPrice} 元/kWh</p>
                  <p className="text-lg font-bold text-blue-600">¥ {Number(r.totalCost).toLocaleString()}</p>
                  <p className="text-[10px] text-muted-foreground">预估月电费（谷30% 平40% 峰30%）</p>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">请输入用电量查看模拟结果</p>
          )}
        </div>
      </ChartContainer>

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
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  加载中...
                </TableCell>
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
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

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
    <Dialog open onClose={onClose}>
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
