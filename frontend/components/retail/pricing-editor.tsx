'use client';

import { useMemo } from 'react';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Bar,
  BarChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  CartesianGrid,
} from 'recharts';
import { Plus, Trash2 } from 'lucide-react';

// 定价配置类型
export interface TimeOfUsePeriod {
  name: string;   // 峰时 / 平时 / 谷时
  price: number;  // 元/kWh
  hours?: string; // 时段描述，如 "08:00-11:00"
}

export interface TieredPrice {
  min_mwh: number;
  max_mwh: number | null; // null = 无上限
  price: number;          // 元/kWh
}

export interface PricingConfig {
  mode: 'tou' | 'tiered' | 'fixed' | 'market_spread';
  fixed_price?: number;
  service_fee?: number;       // 元/kWh 服务费
  periods?: TimeOfUsePeriod[];
  tiers?: TieredPrice[];
  market_spread?: number;     // 市场价 +/- 价差
  description?: string;
}

interface PricingEditorProps {
  value: PricingConfig;
  onChange: (v: PricingConfig) => void;
  pricingMode?: string; // 从 PricingModel.pricing_mode 传入
}

export function PricingEditor({ value, onChange, pricingMode }: PricingEditorProps) {
  const mode = value.mode || 'tou';

  const update = (patch: Partial<PricingConfig>) => {
    onChange({ ...value, ...patch });
  };

  // 模式优先用传入的 pricingMode，否则用 value.mode
  const effectiveMode = pricingMode || mode;

  const chartData = useMemo(() => {
    if (effectiveMode === 'tou' && value.periods) {
      return value.periods.map((p) => ({ name: p.name, 价格: p.price }));
    }
    if (effectiveMode === 'tiered' && value.tiers) {
      return value.tiers.map((t, i) => ({
        name: t.max_mwh ? `${t.min_mwh}-${t.max_mwh}` : `${t.min_mwh}+`,
        价格: t.price,
      }));
    }
    if (effectiveMode === 'fixed' && value.fixed_price != null) {
      return [{ name: '固定单价', 价格: value.fixed_price }];
    }
    return [];
  }, [effectiveMode, value]);

  return (
    <div className="space-y-4 rounded-lg border p-4 bg-muted/20">
      {/* Mode selector */}
      <div className="flex items-center gap-2">
        <Label className="text-sm whitespace-nowrap">定价模式</Label>
        <select
          value={effectiveMode}
          onChange={(e) => {
            const m = e.target.value as PricingConfig['mode'];
            const defaults: PricingConfig = {
              mode: m,
              service_fee: value.service_fee ?? 0.01,
              ...(m === 'tou' && {
                periods: value.periods ?? [
                  { name: '峰时', price: 1.05, hours: '08:00-11:00, 18:00-21:00' },
                  { name: '平时', price: 0.65, hours: '其余时段' },
                  { name: '谷时', price: 0.35, hours: '23:00-07:00' },
                ],
              }),
              ...(m === 'tiered' && {
                tiers: value.tiers ?? [
                  { min_mwh: 0, max_mwh: 100, price: 0.65 },
                  { min_mwh: 100, max_mwh: 500, price: 0.60 },
                  { min_mwh: 500, max_mwh: null, price: 0.55 },
                ],
              }),
              ...(m === 'fixed' && { fixed_price: value.fixed_price ?? 0.60 }),
              ...(m === 'market_spread' && { market_spread: value.market_spread ?? -0.02 }),
            };
            onChange(defaults);
          }}
          className="h-8 rounded-md border border-input bg-background px-2 text-sm"
        >
          <option value="tou">分时电价 (峰/平/谷)</option>
          <option value="tiered">阶梯电价</option>
          <option value="fixed">固定单价</option>
          <option value="market_spread">市场价差</option>
        </select>
      </div>

      {/* Fixed price */}
      {effectiveMode === 'fixed' && (
        <div className="flex items-center gap-2">
          <Label className="text-sm whitespace-nowrap">单价 (元/kWh)</Label>
          <Input
            type="number"
            step="0.01"
            value={value.fixed_price ?? ''}
            onChange={(e) => update({ fixed_price: parseFloat(e.target.value) || 0 })}
            className="w-32"
          />
        </div>
      )}

      {/* Market spread */}
      {effectiveMode === 'market_spread' && (
        <div className="flex items-center gap-2">
          <Label className="text-sm whitespace-nowrap">价差 (元/kWh)</Label>
          <Input
            type="number"
            step="0.005"
            value={value.market_spread ?? ''}
            onChange={(e) => update({ market_spread: parseFloat(e.target.value) || 0 })}
            className="w-32"
          />
          <span className="text-xs text-muted-foreground">
            正数=加价，负数=优惠
          </span>
        </div>
      )}

      {/* Time-of-use periods */}
      {effectiveMode === 'tou' && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label className="text-sm">分时段配置</Label>
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-7 px-2 text-xs"
              onClick={() => update({
                periods: [...(value.periods ?? []), { name: '新时段', price: 0.5, hours: '' }],
              })}
            >
              <Plus className="h-3 w-3 mr-1" /> 添加时段
            </Button>
          </div>
          {(value.periods ?? []).map((p, i) => (
            <div key={i} className="flex items-center gap-2">
              <Input
                value={p.name}
                onChange={(e) => {
                  const periods = [...(value.periods ?? [])];
                  periods[i] = { ...p, name: e.target.value };
                  update({ periods });
                }}
                className="w-24"
                placeholder="时段名"
              />
              <Input
                value={p.hours ?? ''}
                onChange={(e) => {
                  const periods = [...(value.periods ?? [])];
                  periods[i] = { ...p, hours: e.target.value };
                  update({ periods });
                }}
                className="flex-1"
                placeholder="时段范围，如 08:00-11:00"
              />
              <Input
                type="number"
                step="0.01"
                value={p.price}
                onChange={(e) => {
                  const periods = [...(value.periods ?? [])];
                  periods[i] = { ...p, price: parseFloat(e.target.value) || 0 };
                  update({ periods });
                }}
                className="w-24"
                placeholder="元/kWh"
              />
              <Button
                type="button"
                size="sm"
                variant="ghost"
                className="h-7 px-2"
                onClick={() => update({ periods: (value.periods ?? []).filter((_, idx) => idx !== i) })}
              >
                <Trash2 className="h-3.5 w-3.5 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      )}

      {/* Tiered pricing */}
      {effectiveMode === 'tiered' && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label className="text-sm">阶梯区间</Label>
            <Button
              type="button"
              size="sm"
              variant="outline"
              className="h-7 px-2 text-xs"
              onClick={() => {
                const tiers = value.tiers ?? [];
                const last = tiers[tiers.length - 1];
                update({
                  tiers: [...tiers, { min_mwh: last?.max_mwh ?? 0, max_mwh: null, price: 0.5 }],
                });
              }}
            >
              <Plus className="h-3 w-3 mr-1" /> 添加阶梯
            </Button>
          </div>
          {(value.tiers ?? []).map((t, i) => (
            <div key={i} className="flex items-center gap-2">
              <Badge variant="secondary" className="text-xs">阶梯 {i + 1}</Badge>
              <Input
                type="number"
                value={t.min_mwh}
                onChange={(e) => {
                  const tiers = [...(value.tiers ?? [])];
                  tiers[i] = { ...t, min_mwh: parseFloat(e.target.value) || 0 };
                  update({ tiers });
                }}
                className="w-28"
                placeholder="最小 MWh"
              />
              <span className="text-xs text-muted-foreground">~</span>
              <Input
                type="number"
                value={t.max_mwh ?? ''}
                onChange={(e) => {
                  const tiers = [...(value.tiers ?? [])];
                  tiers[i] = { ...t, max_mwh: e.target.value ? parseFloat(e.target.value) : null };
                  update({ tiers });
                }}
                className="w-28"
                placeholder="最大 (空=无限)"
              />
              <Input
                type="number"
                step="0.01"
                value={t.price}
                onChange={(e) => {
                  const tiers = [...(value.tiers ?? [])];
                  tiers[i] = { ...t, price: parseFloat(e.target.value) || 0 };
                  update({ tiers });
                }}
                className="w-24"
                placeholder="元/kWh"
              />
              <Button
                type="button"
                size="sm"
                variant="ghost"
                className="h-7 px-2"
                onClick={() => update({ tiers: (value.tiers ?? []).filter((_, idx) => idx !== i) })}
              >
                <Trash2 className="h-3.5 w-3.5 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      )}

      {/* Service fee (common to all modes) */}
      <div className="flex items-center gap-2 border-t pt-2">
        <Label className="text-sm whitespace-nowrap">服务费 (元/kWh)</Label>
        <Input
          type="number"
          step="0.001"
          value={value.service_fee ?? ''}
          onChange={(e) => update({ service_fee: parseFloat(e.target.value) || 0 })}
          className="w-28"
        />
      </div>

      {/* Chart preview */}
      {chartData.length > 0 && (
        <div className="border-t pt-2">
          <p className="mb-1 text-xs font-medium text-muted-foreground">价格预览</p>
          <ResponsiveContainer width="100%" height={140}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="name" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} domain={[0, 'auto']} />
              <Tooltip
                formatter={(v: number) => [`${v.toFixed(3)} 元/kWh`, '价格']}
                contentStyle={{ fontSize: 12 }}
              />
              <Bar dataKey="价格" fill="#3B82F6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
