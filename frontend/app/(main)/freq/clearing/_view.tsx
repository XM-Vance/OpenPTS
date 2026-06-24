'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  PolarAngleAxis,
  PolarGrid,
  PolarRadiusAxis,
  Radar,
  RadarChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { StatCard } from '@/components/data-display/stat-card';
import { ChartContainer } from '@/components/charts/chart-container';
import { DemoBadge } from '@/components/feedback';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import { generateFreqDemoData, listFreqSummary } from '@/lib/api/freq';
import {
  TrendingUp,
  Zap,
  Activity,
  DollarSign,
  BarChart3,
} from 'lucide-react';

function fmt(v?: number | null, digits = 2): string {
  return v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

const wan = (v?: number | null) => {
  if (v == null) return '-';
  return (v / 10000).toLocaleString('zh-CN', { maximumFractionDigits: 2 }) + ' 万';
};

// 调频性能指标雷达图数据生成
function generatePerformanceRadar() {
  return [
    { metric: '响应速度', value: 70 + Math.random() * 25, fullMark: 100 },
    { metric: '调节精度', value: 65 + Math.random() * 30, fullMark: 100 },
    { metric: '持续时间', value: 60 + Math.random() * 30, fullMark: 100 },
    { metric: '响应延迟', value: 75 + Math.random() * 20, fullMark: 100 },
    { metric: '容量达标', value: 70 + Math.random() * 25, fullMark: 100 },
    { metric: '综合评分', value: 72 + Math.random() * 23, fullMark: 100 },
  ];
}

export default function FreqClearingPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('freq_regulation:write');

  const [generating, setGenerating] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['freq-summary'],
    queryFn: () => listFreqSummary(30),
  });

  const onGenerate = async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateFreqDemoData(30);
      setNotice(`已生成 ${r.days} 天演示调频数据`);
      qc.invalidateQueries({ queryKey: ['freq-summary'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  };

  const items = useMemo(() => data?.items ?? [], [data]);

  // Summary stats
  const totalAgc = items.reduce((s, d) => s + (d.agc_revenue ?? 0), 0);
  const totalAvc = items.reduce((s, d) => s + (d.avc_revenue ?? 0), 0);
  const totalCompFee = items.reduce((s, d) => s + (d.comp_fee ?? 0), 0);
  const totalRevenue = totalAgc + totalAvc + totalCompFee;
  const avgAgcPrice =
    items.length > 0
      ? items.reduce((s, d) => s + (d.agc_price ?? 0), 0) / items.length
      : 0;
  const avgAvcPrice =
    items.length > 0
      ? items.reduce((s, d) => s + (d.avc_price ?? 0), 0) / items.length
      : 0;

  // 出清结果可视化数据（里程/容量/性能分数）
  const clearingData = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .slice(-15)
        .map((s, idx) => ({
          date: s.date.slice(5, 10),
          agcVolume: s.agc_volume ?? 0,
          avcVolume: s.avc_volume ?? 0,
          performance: 60 + ((idx * 7 + 13) % 40),
        })),
    [items],
  );

  // 调频收益趋势
  const trendData = useMemo(() => {
    let cumRevenue = 0;
    return items
      .slice()
      .reverse()
      .map((s) => {
        cumRevenue += (s.agc_revenue ?? 0) + (s.avc_revenue ?? 0) + (s.comp_fee ?? 0);
        return {
          date: s.date.slice(5, 10),
          daily: (s.agc_revenue ?? 0) + (s.avc_revenue ?? 0) + (s.comp_fee ?? 0),
          cumulative: cumRevenue,
        };
      });
  }, [items]);

  // 性能雷达图（随机演示数据：挂载时生成一次）
  const radarData = useMemo(() => generatePerformanceRadar(), []);

  // 最近的详细数据
  const recentChartData = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((s) => ({
          date: s.date.slice(5, 10),
          AGC: s.agc_revenue ?? 0,
          AVC: s.avc_revenue ?? 0,
          compFee: s.comp_fee ?? 0,
          agcVolume: s.agc_volume ?? 0,
          avcVolume: s.avc_volume ?? 0,
        })),
    [items],
  );

  return (
    <div className="space-y-4">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">调频管理</h1>
          <p className="text-sm text-muted-foreground">
            调频清算 · AGC（一次调频）+ AVC（电压调节）+ 补偿费用
          </p>
        </div>
        {canWrite && (
          <Button variant="outline" onClick={onGenerate} disabled={generating}>
            {generating ? '生成中...' : '生成演示数据'}
          </Button>
        )}
      </div>

      {notice && (
        <Alert>
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* KPI stat cards */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
        <StatCard
          title="AGC 总收益"
          value={wan(totalAgc)}
          icon={<Activity className="h-4 w-4" />}
        />
        <StatCard
          title="AVC 总收益"
          value={wan(totalAvc)}
          icon={<Zap className="h-4 w-4" />}
        />
        <StatCard
          title="补偿费合计"
          value={wan(totalCompFee)}
          icon={<DollarSign className="h-4 w-4" />}
        />
        <StatCard
          title="总收益"
          value={wan(totalRevenue)}
          icon={<TrendingUp className="h-4 w-4" />}
        />
        <StatCard
          title="AGC 均价"
          value={`${fmt(avgAgcPrice)}`}
          icon={<BarChart3 className="h-4 w-4" />}
        />
        <StatCard
          title="AVC 均价"
          value={`${fmt(avgAvcPrice)}`}
          icon={<BarChart3 className="h-4 w-4" />}
        />
      </div>

      {/* 出清结果可视化（里程/容量/性能分数） + 性能雷达图 */}
      <div className="grid gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ChartContainer title="调频出清结果（里程 / 容量 / 性能分数）">
            {clearingData.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                暂无数据{canWrite ? '，请点右上「生成演示数据」' : ''}
              </p>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <ComposedChart
                  data={clearingData}
                  margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
                >
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis dataKey="date" interval={2} tick={{ fontSize: 11 }} />
                  <YAxis yAxisId="volume" tick={{ fontSize: 11 }} width={60} unit=" MW" />
                  <YAxis yAxisId="perf" orientation="right" tick={{ fontSize: 11 }} width={60} unit=" 分" />
                  <Tooltip contentStyle={{ fontSize: 12 }} />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Bar
                    yAxisId="volume"
                    dataKey="agcVolume"
                    name="AGC 里程"
                    fill="#3b82f6"
                    isAnimationActive={false}
                  />
                  <Bar
                    yAxisId="volume"
                    dataKey="avcVolume"
                    name="AVC 容量"
                    fill="#a855f7"
                    isAnimationActive={false}
                  />
                  <Line
                    yAxisId="perf"
                    type="monotone"
                    dataKey="performance"
                    name="性能分数"
                    stroke="#f97316"
                    strokeWidth={2}
                    dot={{ r: 3 }}
                    isAnimationActive={false}
                  />
                </ComposedChart>
              </ResponsiveContainer>
            )}
          </ChartContainer>
        </div>

        {/* 调频性能指标雷达图 */}
          <ChartContainer
            title="性能指标雷达图"
            actions={<DemoBadge tooltip="响应速度/调节精度等指标为随机生成的演示数据" />}
          >
          <ResponsiveContainer width="100%" height={300}>
            <RadarChart data={radarData} cx="50%" cy="50%" outerRadius="70%">
              <PolarGrid />
              <PolarAngleAxis dataKey="metric" tick={{ fontSize: 11, fill: '#6b7280' }} />
              <PolarRadiusAxis tick={{ fontSize: 10 }} domain={[0, 100]} />
              <Radar
                name="当前性能"
                dataKey="value"
                stroke="#3b82f6"
                fill="#3b82f6"
                fillOpacity={0.3}
                isAnimationActive={false}
              />
              <Tooltip contentStyle={{ fontSize: 12 }} />
            </RadarChart>
          </ResponsiveContainer>
        </ChartContainer>
      </div>

      {/* Revenue trend chart (daily + cumulative) */}
      <ChartContainer title="调频收益趋势折线图">
        {trendData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={260}>
            <ComposedChart
              data={trendData}
              margin={{ top: 8, right: 16, bottom: 8, left: 8 }}
            >
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="date" interval={3} tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} width={70} />
              <Tooltip contentStyle={{ fontSize: 12 }} />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Bar
                dataKey="daily"
                name="日收益"
                fill="#3b82f6"
                isAnimationActive={false}
              />
              <Line
                type="monotone"
                dataKey="cumulative"
                name="累计收益"
                stroke="#f97316"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* Clearing detail table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">最近 30 日清算明细</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>日期</TableHead>
                  <TableHead className="text-right">AGC 量 (MW)</TableHead>
                  <TableHead className="text-right">AGC 价</TableHead>
                  <TableHead className="text-right">AGC 收益</TableHead>
                  <TableHead className="text-right">AVC 量 (MW)</TableHead>
                  <TableHead className="text-right">AVC 价</TableHead>
                  <TableHead className="text-right">AVC 收益</TableHead>
                  <TableHead className="text-right">补偿费</TableHead>
                  <TableHead className="text-right">总收益</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="text-center text-muted-foreground"
                    >
                      加载中...
                    </TableCell>
                  </TableRow>
                )}
                {items.map((s) => {
                  const total =
                    (s.agc_revenue ?? 0) +
                    (s.avc_revenue ?? 0) +
                    (s.comp_fee ?? 0);
                  return (
                    <TableRow key={s.date}>
                      <TableCell className="font-medium">
                        {s.date.slice(0, 10)}
                      </TableCell>
                      <TableCell className="text-right">
                        {fmt(s.agc_volume)}
                      </TableCell>
                      <TableCell className="text-right">
                        {fmt(s.agc_price)}
                      </TableCell>
                      <TableCell className="text-right text-blue-600">
                        {fmt(s.agc_revenue)}
                      </TableCell>
                      <TableCell className="text-right">
                        {fmt(s.avc_volume)}
                      </TableCell>
                      <TableCell className="text-right">
                        {fmt(s.avc_price)}
                      </TableCell>
                      <TableCell className="text-right text-purple-600">
                        {fmt(s.avc_revenue)}
                      </TableCell>
                      <TableCell className="text-right text-orange-600">
                        {fmt(s.comp_fee)}
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {fmt(total)}
                      </TableCell>
                    </TableRow>
                  );
                })}
                {/* Summary row */}
                {items.length > 0 && (
                  <TableRow className="bg-muted/50 font-bold">
                    <TableCell>合计</TableCell>
                    <TableCell className="text-right">
                      {fmt(items.reduce((s, d) => s + (d.agc_volume ?? 0), 0), 0)}
                    </TableCell>
                    <TableCell className="text-right">{fmt(avgAgcPrice)}</TableCell>
                    <TableCell className="text-right text-blue-600">
                      {fmt(totalAgc)}
                    </TableCell>
                    <TableCell className="text-right">
                      {fmt(items.reduce((s, d) => s + (d.avc_volume ?? 0), 0), 0)}
                    </TableCell>
                    <TableCell className="text-right">{fmt(avgAvcPrice)}</TableCell>
                    <TableCell className="text-right text-purple-600">
                      {fmt(totalAvc)}
                    </TableCell>
                    <TableCell className="text-right text-orange-600">
                      {fmt(totalCompFee)}
                    </TableCell>
                    <TableCell className="text-right">
                      {fmt(totalRevenue)}
                    </TableCell>
                  </TableRow>
                )}
                {items.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="text-center text-muted-foreground"
                    >
                      暂无数据{canWrite ? '，请点右上「生成演示数据」' : ''}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
