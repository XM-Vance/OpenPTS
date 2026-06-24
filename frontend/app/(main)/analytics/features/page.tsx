'use client';

import Link from 'next/link';
import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DemoBadge } from '@/components/feedback';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  ackAlert,
  generateAnalyticsDemoData,
  getAlertStats,
  listAlerts,
  listCharacteristics,
  type CustomerAlert,
} from '@/lib/api/analytics';

const ALERT_TYPE_LABEL: Record<string, string> = {
  load_drop: '负荷下降',
  shape_change: '形态变化',
  quality_drop: '质量下降',
  spike: '尖峰',
};

const SEVERITY_LABEL: Record<string, string> = {
  info: '提示',
  warn: '警告',
  critical: '严重',
};

function severityVariant(s: string): 'default' | 'secondary' | 'destructive' | 'success' {
  if (s === 'critical') return 'destructive';
  if (s === 'warn') return 'default';
  if (s === 'info') return 'secondary';
  return 'secondary';
}

const PIE_COLORS = ['#3b82f6', '#10b981', '#f97316', '#8b5cf6', '#06b6d4', '#94a3b8', '#ec4899', '#14b8a6'];

// recharts 懒加载：从首屏 JS 剥离，挂载后异步加载（渲染内容不变）。
const FeatureRadar = dynamic(
  () => import('./_charts').then((m) => ({ default: m.FeatureRadar })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);
const ValueScatter = dynamic(
  () => import('./_charts').then((m) => ({ default: m.ValueScatter })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);
const IndustryPie = dynamic(
  () => import('./_charts').then((m) => ({ default: m.IndustryPie })),
  { ssr: false, loading: () => <div className="h-full w-full" /> },
);

export default function AnalyticsFeaturesPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('analytics:write');

  const [includeAcked, setIncludeAcked] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [ackingId, setAckingId] = useState<string | null>(null);

  const { data: stats } = useQuery({
    queryKey: ['analytics-stats'],
    queryFn: getAlertStats,
  });

  const { data: alerts, isLoading: alertsLoading } = useQuery({
    queryKey: ['analytics-alerts', includeAcked],
    queryFn: () => listAlerts({ limit: 100, include_acked: includeAcked }),
  });

  const { data: chars } = useQuery({
    queryKey: ['analytics-characteristics'],
    queryFn: () => listCharacteristics(20),
  });

  const onGenerate = async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateAnalyticsDemoData();
      setNotice(
        `已生成：${r.customers} 客户 / ${r.characteristics} 条特征 / ${r.alerts} 条告警`,
      );
      qc.invalidateQueries({ queryKey: ['analytics-stats'] });
      qc.invalidateQueries({ queryKey: ['analytics-alerts'] });
      qc.invalidateQueries({ queryKey: ['analytics-characteristics'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  };

  const onAck = async (a: CustomerAlert) => {
    setAckingId(a.id);
    try {
      await ackAlert(a.id);
      qc.invalidateQueries({ queryKey: ['analytics-stats'] });
      qc.invalidateQueries({ queryKey: ['analytics-alerts'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setAckingId(null);
    }
  };

  // ── Chart 1: 客户特征雷达图 ──
  const radarData = useMemo(() => {
    if (!chars?.items || chars.items.length === 0) return { data: [], customers: [] as string[] };
    const dims = ['用电量', '稳定性', '收益性', '风险度', '成长性'];
    const customers = chars.items.slice(0, 4).map((c) => c.customer_name);
    const data = dims.map((dim) => {
      const row: Record<string, string | number> = { dimension: dim };
      chars.items.slice(0, 4).forEach((c) => {
        const score = c.regularity_score != null
          ? Math.round(c.regularity_score * 100)
          : Math.round(40 + Math.random() * 50);
        row[c.customer_name] = score;
      });
      return row;
    });
    return { data, customers };
  }, [chars]);

  const RADAR_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444'];

  // ── Chart 2: 客户价值矩阵 ScatterChart ──
  const scatterData = useMemo(() => {
    if (!chars?.items) return [];
    return chars.items.map((c) => ({
      customer: c.customer_name,
      收益: Math.round((c.regularity_score ?? 0.5) * 80 + Math.random() * 20),
      风险: Math.round(Math.random() * 100),
      用电量: Math.round(50 + Math.random() * 200),
    }));
  }, [chars]);

  // ── Chart 3: 行业分布饼图 ──
  const industryPie = useMemo(() => {
    if (!chars?.items) return [];
    const industries = ['制造业', '商业', '居民', '工业', '服务业', '教育'];
    const map = new Map<string, number>();
    chars.items.forEach((_, i) => {
      const ind = industries[i % industries.length];
      map.set(ind, (map.get(ind) || 0) + 1);
    });
    return Array.from(map.entries()).map(([name, value]) => ({ name, value }));
  }, [chars]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">客户分析</h1>
          <p className="text-sm text-muted-foreground">客户异动告警 + 特征画像</p>
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

      <div className="grid gap-4 md:grid-cols-4">
        <StatCard label="告警总数" value={String(stats?.total ?? 0)} />
        <StatCard
          label="待处理"
          value={String(stats?.pending ?? 0)}
          tone="warn"
        />
        <StatCard
          label="已确认"
          value={String(stats?.acknowledged ?? 0)}
          tone="ok"
        />
        <StatCard
          label="严重告警"
          value={String(stats?.critical ?? 0)}
          tone="critical"
        />
      </div>

      {/* 异动告警 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">异动告警</CardTitle>
            <label className="flex items-center gap-2 text-sm text-muted-foreground">
              <input
                type="checkbox"
                checked={includeAcked}
                onChange={(e) => setIncludeAcked(e.target.checked)}
              />
              显示已确认
            </label>
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>日期</TableHead>
                  <TableHead>客户</TableHead>
                  <TableHead>类型</TableHead>
                  <TableHead>严重度</TableHead>
                  <TableHead className="text-right">置信度</TableHead>
                  <TableHead>原因</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {alertsLoading && (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center text-muted-foreground">
                      加载中...
                    </TableCell>
                  </TableRow>
                )}
                {alerts?.items.map((a) => (
                  <TableRow key={a.id}>
                    <TableCell className="font-medium">
                      {a.alert_date.slice(0, 10)}
                    </TableCell>
                    <TableCell>
                      <Link
                        href={`/customers?highlight=${a.customer_id}`}
                        className="text-blue-600 hover:underline"
                      >
                        {a.customer_name}
                      </Link>
                    </TableCell>
                    <TableCell>{ALERT_TYPE_LABEL[a.alert_type] ?? a.alert_type}</TableCell>
                    <TableCell>
                      <Badge variant={severityVariant(a.severity)}>
                        {SEVERITY_LABEL[a.severity] ?? a.severity}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      {a.confidence != null ? `${a.confidence.toFixed(1)}%` : '-'}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {a.reason ?? '-'}
                    </TableCell>
                    <TableCell>
                      {a.acknowledged ? (
                        <Badge variant="success">已确认</Badge>
                      ) : (
                        <Badge variant="outline">待处理</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      {!a.acknowledged && canWrite && (
                        <Button
                          size="sm"
                          variant="ghost"
                          disabled={ackingId === a.id}
                          onClick={() => onAck(a)}
                        >
                          {ackingId === a.id ? '提交...' : '确认'}
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                {alerts?.items.length === 0 && !alertsLoading && (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center text-muted-foreground">
                      {includeAcked
                        ? '暂无告警数据'
                        : '没有待处理告警 🎉'}
                      {canWrite && !alerts?.items.length && '，可点右上「生成演示数据」'}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-3">
        {/* Chart 1: 客户特征雷达图 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">客户特征雷达图
              <DemoBadge className="ml-1" tooltip="无评分客户随机生成特征分数" />
            </CardTitle>
          </CardHeader>
          <CardContent>
            {radarData.data.length > 0 ? (
              <div className="h-64 w-full [&_.recharts-surface:focus]:outline-none">
                <FeatureRadar data={radarData.data} customers={radarData.customers} colors={RADAR_COLORS} />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                暂无特征数据{canWrite ? '，请点右上「生成演示数据」' : ''}
              </p>
            )}
          </CardContent>
        </Card>

        {/* Chart 2: 客户价值矩阵 ScatterChart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">客户价值矩阵（收益 × 风险 × 用电量）
              <DemoBadge className="ml-1" tooltip="收益/风险/用电量均由 Math.random 生成" />
            </CardTitle>
          </CardHeader>
          <CardContent>
            {scatterData.length > 0 ? (
              <div className="h-64 w-full [&_.recharts-surface:focus]:outline-none">
                <ValueScatter data={scatterData} />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">暂无数据</p>
            )}
            <p className="mt-1 text-xs text-muted-foreground">气泡大小 = 用电量 | X = 收益 | Y = 风险</p>
          </CardContent>
        </Card>

        {/* Chart 3: 行业分布饼图 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">行业分布</CardTitle>
          </CardHeader>
          <CardContent>
            {industryPie.length > 0 ? (
              <div className="h-64 w-full [&_.recharts-surface:focus]:outline-none">
                <IndustryPie data={industryPie} colors={PIE_COLORS} />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">暂无数据</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 客户特征画像（每客户最新） */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">客户特征画像（每客户最新）</CardTitle>
        </CardHeader>
        <CardContent>
          {chars?.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              暂无特征数据{canWrite ? '，请点右上「生成演示数据」' : ''}
            </p>
          ) : (
            <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
              {chars?.items.map((ch) => (
                <div key={ch.id} className="rounded-md border bg-muted/30 p-4">
                  <Link
                    href={`/customers?highlight=${ch.customer_id}`}
                    className="font-medium text-blue-600 hover:underline"
                  >
                    {ch.customer_name}
                  </Link>
                  <p className="text-xs text-muted-foreground">
                    截至 {ch.data_date.slice(0, 10)}
                  </p>
                  <div className="mt-3 flex flex-wrap gap-1">
                    {ch.tags?.map((t) => (
                      <Badge key={t} variant="outline" className="text-xs">
                        {t}
                      </Badge>
                    ))}
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-2 text-sm">
                    <div>
                      <p className="text-xs text-muted-foreground">规律性</p>
                      <p className="font-medium">
                        {ch.regularity_score != null
                          ? ch.regularity_score.toFixed(2)
                          : '-'}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">数据质量</p>
                      <p className="font-medium">{ch.quality_rating ?? '-'}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone?: 'warn' | 'ok' | 'critical';
}) {
  const color =
    tone === 'critical'
      ? 'text-destructive'
      : tone === 'warn'
        ? 'text-amber-600'
        : tone === 'ok'
          ? 'text-emerald-600'
          : 'text-foreground';
  return (
    <Card>
      <CardContent className="pt-6">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className={`mt-1 text-2xl font-bold ${color}`}>{value}</p>
      </CardContent>
    </Card>
  );
}
