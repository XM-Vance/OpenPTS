'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ChartContainer } from '@/components/charts/chart-container';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  getCarbonSummary,
  getCarbonQuotes,
  genCarbonDemo,
  type CarbonQuote,
  type CarbonProductSummary,
} from '@/lib/api/carbon';
import { Leaf, TrendingUp, TrendingDown } from 'lucide-react';
import type { CarbonTrendPoint } from './_charts';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const fmt = (v: number | null | undefined) =>
  v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: 2 });
const fmtInt = (v: number | null | undefined) =>
  v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: 0 });

// 产品配色（与走势图一致）
const PRODUCT_COLOR: Record<string, string> = {
  CEA: 'text-emerald-600',
  CCER: 'text-blue-600',
  EUA: 'text-amber-600',
};

// recharts 懒加载：剥离出首屏。
const CarbonTrendLine = dynamic(
  () => import('./_charts').then((m) => ({ default: m.CarbonTrendLine })),
  { ssr: false, loading: () => <div className="h-[320px] w-full" /> },
);

export default function CarbonPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('price_management:write');

  const [days, setDays] = useState(180);
  const [product, setProduct] = useState(''); // 表格产品筛选
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: summaryData } = useQuery({
    queryKey: ['carbon-summary'],
    queryFn: () => getCarbonSummary().then((r) => r.data),
  });
  const { data: quotesData, isLoading } = useQuery({
    queryKey: ['carbon-quotes', days],
    queryFn: () => getCarbonQuotes({ days }).then((r) => r.data),
  });

  const summary: CarbonProductSummary[] = summaryData?.items ?? [];
  const quotes = useMemo<CarbonQuote[]>(() => quotesData?.items ?? [], [quotesData]);

  // 走势图数据：按日期归并三个产品的收盘价。
  const trend: CarbonTrendPoint[] = useMemo(() => {
    const byDate = new Map<string, CarbonTrendPoint>();
    for (const q of quotes) {
      const d = q.trade_date.slice(0, 10);
      let pt = byDate.get(d);
      if (!pt) {
        pt = { date: d, CEA: null, CCER: null, EUA: null };
        byDate.set(d, pt);
      }
      if (q.product === 'CEA' || q.product === 'CCER' || q.product === 'EUA') {
        pt[q.product] = q.close_price;
      }
    }
    return Array.from(byDate.values()).sort((a, b) => a.date.localeCompare(b.date));
  }, [quotes]);

  // 表格行（按产品筛选 + 日期倒序，接口已倒序）。
  const tableRows = useMemo(
    () => (product ? quotes.filter((q) => q.product === product) : quotes),
    [quotes, product],
  );

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genCarbonDemo();
      qc.invalidateQueries({ queryKey: ['carbon-summary'] });
      qc.invalidateQueries({ queryKey: ['carbon-quotes'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">碳交易</h1>
          <p className="text-sm text-muted-foreground">
            全国碳排放配额(CEA)、国家核证自愿减排量(CCER)与欧盟配额(EUA)行情 · 全国统一报价
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={days}
            onChange={(e) => setDays(Number(e.target.value))}
            className={SELECT_CLASS}
          >
            <option value={30}>近 30 天</option>
            <option value={90}>近 90 天</option>
            <option value={180}>近 180 天</option>
            <option value={365}>近 1 年</option>
          </select>
          {canWrite && (
            <Button variant="outline" onClick={onGen} disabled={busy}>
              {busy ? '生成中...' : '生成演示数据'}
            </Button>
          )}
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* ── 三个碳产品最新行情卡片 ── */}
      <div className="grid gap-4 md:grid-cols-3">
        {summary.map((s) => {
          const up = (s.change ?? 0) >= 0;
          return (
            <Card key={s.product}>
              <CardContent className="pt-6">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <Leaf className={`h-4 w-4 ${PRODUCT_COLOR[s.product] ?? ''}`} />
                      <span className="text-sm font-semibold">{s.product}</span>
                      <span className="text-xs text-muted-foreground">{s.name}</span>
                    </div>
                    <p className={`mt-2 text-3xl font-bold ${PRODUCT_COLOR[s.product] ?? ''}`}>
                      {fmt(s.close)}
                    </p>
                    <p className="text-xs text-muted-foreground">{s.unit}</p>
                  </div>
                  {s.close != null && s.change != null && (
                    <div
                      className={`flex items-center gap-1 text-sm font-medium ${
                        up ? 'text-red-600' : 'text-green-600'
                      }`}
                    >
                      {up ? (
                        <TrendingUp className="h-4 w-4" />
                      ) : (
                        <TrendingDown className="h-4 w-4" />
                      )}
                      <span>
                        {up ? '+' : ''}
                        {fmt(s.change)}
                        {s.change_pct != null && (
                          <span className="ml-1">
                            ({up ? '+' : ''}
                            {s.change_pct.toFixed(2)}%)
                          </span>
                        )}
                      </span>
                    </div>
                  )}
                </div>
                <div className="mt-3 flex justify-between text-xs text-muted-foreground">
                  <span>
                    年内区间 {fmt(s.low_52w)} ~ {fmt(s.high_52w)}
                  </span>
                  <span>成交量 {fmtInt(s.volume)}</span>
                </div>
                {s.latest_date && (
                  <p className="mt-1 text-[11px] text-muted-foreground">
                    截至 {s.latest_date.slice(0, 10)}
                  </p>
                )}
              </CardContent>
            </Card>
          );
        })}
        {summary.length === 0 && (
          <Card className="md:col-span-3">
            <CardContent className="py-8 text-center text-sm text-muted-foreground">
              暂无碳行情数据{canWrite && '，可点右上「生成演示数据」'}
            </CardContent>
          </Card>
        )}
      </div>

      {/* ── 碳价走势 ── */}
      <ChartContainer title="碳价走势（收盘价）">
        <CarbonTrendLine data={trend} />
      </ChartContainer>
      <p className="-mt-2 text-xs text-muted-foreground">
        注：CEA / CCER 以人民币元计价，EUA 以欧元计价，三者叠加仅作走势对比。
      </p>

      {/* ── 行情明细 ── */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted-foreground">产品筛选</span>
        <select
          value={product}
          onChange={(e) => setProduct(e.target.value)}
          className={SELECT_CLASS}
        >
          <option value="">全部</option>
          <option value="CEA">CEA 全国碳配额</option>
          <option value="CCER">CCER 自愿减排</option>
          <option value="EUA">EUA 欧盟配额</option>
        </select>
      </div>
      <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>交易日期</TableHead>
              <TableHead>产品</TableHead>
              <TableHead className="text-right">开盘</TableHead>
              <TableHead className="text-right">最高</TableHead>
              <TableHead className="text-right">最低</TableHead>
              <TableHead className="text-right">收盘</TableHead>
              <TableHead className="text-right">成交量</TableHead>
              <TableHead className="text-right">成交额</TableHead>
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
            {tableRows.slice(0, 500).map((q) => (
              <TableRow key={q.id}>
                <TableCell className="font-medium">{q.trade_date.slice(0, 10)}</TableCell>
                <TableCell>
                  <span className={PRODUCT_COLOR[q.product] ?? ''}>{q.product}</span>
                </TableCell>
                <TableCell className="text-right">{fmt(q.open_price)}</TableCell>
                <TableCell className="text-right">{fmt(q.high_price)}</TableCell>
                <TableCell className="text-right">{fmt(q.low_price)}</TableCell>
                <TableCell className="text-right font-medium">{fmt(q.close_price)}</TableCell>
                <TableCell className="text-right">{fmtInt(q.volume)}</TableCell>
                <TableCell className="text-right">{fmtInt(q.turnover)}</TableCell>
              </TableRow>
            ))}
            {tableRows.length === 0 && !isLoading && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">
                  暂无数据{canWrite && '，可点右上「生成演示数据」'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
