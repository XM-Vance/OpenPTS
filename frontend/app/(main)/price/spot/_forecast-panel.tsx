'use client';

/**
 * LEAR 日前电价预测面板（懒加载，含 recharts）。
 *
 * 集成到 spot 页底部。提供：
 * - 算法切换：相似日法（GET /price/forecast）/ LEAR（POST /price/forecast/lear）
 * - 48 点预测曲线 + 置信带（AreaChart）
 * - 历史预测准确率（目标日已过时从 /price/forecast/results 拉取 wmape/mae/rmse）
 *
 * 算法原理：Lago et al. "Forecasting day-ahead electricity prices" (Applied Energy 2021)。
 * 后端 algo-service/app/lear_forecast.py 自实现，未拷贝 epftoolbox（规避 AGPL）。
 */
import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  CartesianGrid,
  ComposedChart,
  Area,
  Line,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ChartContainer } from '@/components/charts/chart-container';
import { CustomTooltip } from '@/components/charts/custom-tooltip';
import { Brain } from 'lucide-react';
import { format } from 'date-fns';
import {
  forecastPrice,
  forecastPriceLear,
  listPriceForecastResults,
  type PriceForecast,
} from '@/lib/api/price';
import { EmptyState } from '@/components/feedback';

function pointTime(i: number, points = 48): string {
  const mins = (i * 24 * 60) / points;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
}

type Method = 'lear' | 'similar';

export function ForecastPanel() {
  const [method, setMethod] = useState<Method>('lear');
  const targetDate = format(new Date(Date.now() + 86400000), 'yyyy-MM-dd'); // 默认明天

  // 查历史预测准确率（目标日用昨天，确保有 actual）
  const yesterday = format(new Date(Date.now() - 86400000), 'yyyy-MM-dd');
  const { data: historyResults } = useQuery({
    queryKey: ['price-forecast-results', yesterday],
    queryFn: () => listPriceForecastResults(yesterday),
    staleTime: 60_000,
  });

  const forecastQuery = useQuery({
    queryKey: ['price-forecast', method, targetDate],
    queryFn: () =>
      method === 'lear'
        ? forecastPriceLear(targetDate).then((r) => ({ ...r.forecast, forecast_id: r.forecast_id, exogenous_days: r.exogenous_days }))
        : forecastPrice(targetDate).then((r) => r.forecast as PriceForecast),
    enabled: !!targetDate,
  });

  const fc: PriceForecast | undefined = forecastQuery.data as any;
  const loading = forecastQuery.isLoading;

  const chartData = useMemo(() => {
    if (!fc?.forecast) return [];
    return fc.forecast.map((v: number, i: number) => ({
      time: pointTime(i, fc.forecast.length),
      预测价格: Math.round(v * 100) / 100,
      下界: fc.lower ? Math.round(fc.lower[i] * 100) / 100 : null,
      上界: fc.upper ? Math.round(fc.upper[i] * 100) / 100 : null,
    }));
  }, [fc]);

  // 准确率（取昨日 LEAR 记录，若已回填）
  const accuracy = historyResults?.items?.find((r) => r.forecast_method === 'LEAR')?.accuracy_metrics;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2 flex-wrap">
          <Brain className="h-4 w-4 text-primary" />
          日前电价预测
          <span className="text-xs font-normal text-muted-foreground ml-2">
            {fc?.method ?? '加载中…'}
          </span>
          {/* 方法切换 */}
          <div className="ml-auto flex gap-1">
            <Button
              size="sm"
              variant={method === 'lear' ? 'default' : 'outline'}
              onClick={() => setMethod('lear')}
              className="h-7 text-xs"
            >
              LEAR
            </Button>
            <Button
              size="sm"
              variant={method === 'similar' ? 'default' : 'outline'}
              onClick={() => setMethod('similar')}
              className="h-7 text-xs"
            >
              相似日法
            </Button>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="h-[240px] flex items-center justify-center text-sm text-muted-foreground">
            预测计算中（LEAR 训练 24 个 LASSO 模型，约 2-5 秒）…
          </div>
        ) : !fc ? (
          <EmptyState compact className="h-[240px]" title="暂无预测数据，请先生成演示价格数据" />
        ) : (
          <>
            {/* 摘要指标 */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4 text-sm">
              <div>
                <div className="text-xs text-muted-foreground">样本天数</div>
                <div className="font-medium">{fc.sample_days}</div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">峰值预测</div>
                <div className="font-medium">{fc.peak.toFixed(2)} ¥/MWh</div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">谷值预测</div>
                <div className="font-medium">{fc.valley.toFixed(2)} ¥/MWh</div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">昨日准确率</div>
                <div className="font-medium">
                  {accuracy ? `${((1 - accuracy.wmape) * 100).toFixed(1)}%` : '—'}
                </div>
              </div>
            </div>

            <ChartContainer title="48 点日前价预测曲线" minHeight={240}>
              <ResponsiveContainer width="100%" height="100%">
                <ComposedChart data={chartData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} interval={3} />
                  <YAxis tick={{ fontSize: 11 }} width={64} />
                  <RechartsTooltip content={<CustomTooltip unit="¥/MWh" />} />
                  {/* 置信带：上下界围成区域 */}
                  <Area
                    type="monotone"
                    dataKey="上界"
                    stroke="none"
                    fill="#3b82f6"
                    fillOpacity={0.08}
                    isAnimationActive={false}
                  />
                  <Area
                    type="monotone"
                    dataKey="下界"
                    stroke="none"
                    fill="#ffffff"
                    fillOpacity={1}
                    isAnimationActive={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="预测价格"
                    stroke="#2563eb"
                    strokeWidth={2}
                    dot={false}
                    isAnimationActive={false}
                    name="预测价格"
                  />
                </ComposedChart>
              </ResponsiveContainer>
            </ChartContainer>

            {/* 准确率详情 */}
            {accuracy && (
              <div className="mt-3 flex gap-4 text-xs text-muted-foreground">
                <span>WMAPE: {(accuracy.wmape * 100).toFixed(2)}%</span>
                <span>MAE: {accuracy.mae.toFixed(2)}</span>
                <span>RMSE: {accuracy.rmse.toFixed(2)}</span>
                <span className="ml-auto">目标日: {yesterday}（已回填实际值）</span>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
