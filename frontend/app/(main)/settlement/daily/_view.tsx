'use client';

import { useState, useMemo } from 'react';
import { format, subDays } from 'date-fns';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
  Legend,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { usePermission } from '@/lib/auth/use-permission';
import { extractErrorMessage } from '@/lib/api/client';
import {
  generateSettlementDemoData,
  getSettlement,
  listSettlements,
} from '@/lib/api/settlement';
import { StatCard } from '@/components/data-display/stat-card';
import { ChartContainer } from '@/components/charts/chart-container';
import { CustomTooltip } from '@/components/charts/custom-tooltip';
import { DataTable, type DataTableColumn, type DataRow } from '@/components/data-display/data-table';
import type { MobileCardField } from '@/components/data-display/mobile-card-list';
import { PageHeader } from '@/components/data-display/page-header';
import { Label } from '@/components/ui/label';
import {
  DollarSign,
  BarChart3,
  AlertCircle,
  RefreshCw,
  TrendingUp,
  TrendingDown,
  Wallet,
  PieChart,
  CheckCircle2,
  Circle,
  ArrowRightLeft,
  ShieldAlert,
} from 'lucide-react';
import { EmptyState } from '@/components/feedback';

// ─── 常量 ───

/** 偏差费用预警阈值 (¥) */
const DEVIATION_FEE_THRESHOLD = 50000;

// ─── 辅助函数 ───

function pointTime(i: number): string {
  const m = i * 30;
  return `${String(Math.floor(m / 60)).padStart(2, '0')}:${String(m % 60).padStart(2, '0')}`;
}

function fmtMoney(v?: number | null): string {
  if (v == null) return '-';
  return v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function fmt(v: number | null | undefined, digits = 2): string {
  if (v == null) return '-';
  return v.toLocaleString('zh-CN', { maximumFractionDigits: digits });
}

// ─── 页面组件 ──

export default function SettlementDailyPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('settlement_management:write');

  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [version, setVersion] = useState('PRELIMINARY');
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  // ── Sprint 3：结算核对状态（本地状态，简单实现） ──
  const [verifiedSet, setVerifiedSet] = useState<Set<string>>(() => new Set());

  const toggleVerified = (id: string) => {
    setVerifiedSet((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  // 结算列表
  const { data: listData, isLoading: listLoading } = useQuery({
    queryKey: ['settlements'],
    queryFn: () => listSettlements(30),
  });

  // 选中日期的分时明细
  const { data: detail, isLoading: detailLoading } = useQuery({
    queryKey: ['settlement-detail', selectedDate, version],
    queryFn: () => getSettlement(selectedDate!, version),
    enabled: !!selectedDate,
  });

  // 生成演示数据
  const onGenerate = async () => {
    setError(null);
    setNotice(null);
    setGenerating(true);
    try {
      const r = await generateSettlementDemoData(30);
      setNotice(`已生成 ${r.days} 天演示结算数据`);
      qc.invalidateQueries({ queryKey: ['settlements'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setGenerating(false);
    }
  };

  // 汇总统计
  const summary = useMemo(() => {
    const items = listData?.items;
    if (!items || items.length === 0) return null;
    const totalEnergyFee = items.reduce((s: number, i: any) => s + (i.total_energy_fee ?? 0), 0);
    const validPrices = items.filter((i: any) => i.energy_avg_price != null);
    const avgPrice = validPrices.length > 0
      ? validPrices.reduce((s: number, i: any) => s + (i.energy_avg_price ?? 0), 0) / validPrices.length
      : 0;
    const totalContractFee = items.reduce((s: number, i: any) => s + (i.contract_fee ?? 0), 0);
    const totalDAFee = items.reduce((s: number, i: any) => s + (i.day_ahead_fee ?? 0), 0);
    const totalRTFee = items.reduce((s: number, i: any) => s + (i.real_time_fee ?? 0), 0);
    const totalDeviationFee = items.reduce((s: number, i: any) => s + (i.deviation_recovery_fee ?? 0), 0);
    return {
      totalEnergyFee,
      avgPrice,
      totalContractFee,
      totalDAFee,
      totalRTFee,
      totalDeviationFee,
      count: items.length,
    };
  }, [listData]);

  // ── Sprint 3：瀑布图数据 ──
  // 业务审查第二轮修正：本页是批发侧日结算（合同/日前/实时均为购电成本组成），
  // 旧版把"批发合同费用"当售电收入、把 |日前|+|实时| 当偏差考核费——成本组成被重复
  // 计入考核费（线上曾显示考核费 7779 万 > 批发总费用 3861 万）。改为批发侧成本构成瀑布。
  const waterfallData = useMemo(() => {
    if (!summary) return [];
    const contractFee = summary.totalContractFee || 0;
    const daFee = summary.totalDAFee || 0;
    const rtFee = summary.totalRTFee || 0;
    const devFee = summary.totalDeviationFee || 0;
    const total = contractFee + daFee + rtFee;

    return [
      { name: '合同电费', base: 0, amount: contractFee, fill: '#2563eb' },
      { name: '日前偏差电费', base: contractFee, amount: daFee, fill: '#f59e0b' },
      { name: '实时偏差电费', base: contractFee + daFee, amount: rtFee, fill: '#10b981' },
      { name: '偏差考核费', base: total, amount: devFee, fill: '#ef4444' },
      {
        name: '批发侧总支出',
        base: 0,
        amount: total + devFee,
        fill: total + devFee >= 0 ? '#3b82f6' : '#dc2626',
      },
    ];
  }, [summary]);

  // ── Sprint 3：全景看板数据（批发侧口径）──
  // 旧版"零售侧预估=合同费用×1.08、净收益/利润率"无业务依据（合同费用是批发成本项，
  // 不代表零售收入），已移除；改为批发侧可从本页数据诚实计算的指标。
  const panoramaData = useMemo(() => {
    if (!summary) return null;
    const wholesale = summary.totalEnergyFee || 0;
    const contract = summary.totalContractFee || 0;
    const spotDeviation = (summary.totalDAFee || 0) + (summary.totalRTFee || 0);
    const assessFee = summary.totalDeviationFee || 0;
    const contractRatio = wholesale > 0 ? (contract / wholesale) * 100 : 0;
    return {
      wholesale,
      contract,
      spotDeviation,
      assessFee,
      contractRatio,
    };
  }, [summary]);

  // 费用组成
  const feeCompositionData = useMemo(() => {
    if (!summary) return [];
    return [
      { name: '合同费用', value: summary.totalContractFee, color: '#2563eb' },
      { name: '日前费用', value: summary.totalDAFee, color: '#f59e0b' },
      { name: '实时费用', value: summary.totalRTFee, color: '#10b981' },
    ].filter((d) => d.value > 0);
  }, [summary]);

  // 近 30 日趋势图
  const trendData = useMemo(() => {
    if (!listData?.items?.length) return [];
    return [...listData.items]
      .reverse()
      .map((s: any) => ({
        date: s.operating_date?.slice(5, 10) ?? '',
        总电费: s.total_energy_fee ?? 0,
        均价: s.energy_avg_price ?? 0,
        合同费用: s.contract_fee ?? 0,
        日前费用: s.day_ahead_fee ?? 0,
        实时费用: s.real_time_fee ?? 0,
      }));
  }, [listData]);

  // 分时明细柱状图
  const periodChartData = useMemo(() => {
    if (!detail?.period_details?.length) return [];
    return detail.period_details.map((p: any) => ({
      t: pointTime(p.period - 1),
      电量: p.volume_mwh,
      价格: p.price,
      费用: p.fee,
    }));
  }, [detail]);

  // ── Sprint 3：超标预警统计 ──
  const deviationWarningCount = useMemo(() => {
    const items = listData?.items;
    if (!items) return 0;
    return items.filter((i: any) => Math.abs(i.deviation_recovery_fee ?? 0) > DEVIATION_FEE_THRESHOLD).length;
  }, [listData]);

  // 结算列表列
  const columns: DataTableColumn[] = [
    {
      key: 'operating_date',
      header: '日期',
      sortable: true,
      render: (row: DataRow) => {
        const date = (row.operating_date as string)?.slice(0, 10);
        return (
          <Button
            variant="link"
            className="h-auto p-0 font-medium"
            onClick={() => setSelectedDate(date)}
          >
            {date}
          </Button>
        );
      },
    },
    {
      key: 'version',
      header: '版本',
      render: (row: DataRow) => (
        <Badge variant="outline">{row.version as string}</Badge>
      ),
    },
    {
      key: 'contract_fee',
      header: '合同费用 (¥)',
      align: 'right',
      render: (row: DataRow) => fmtMoney(row.contract_fee as number | null),
    },
    {
      key: 'day_ahead_fee',
      header: '日前费用 (¥)',
      align: 'right',
      render: (row: DataRow) => fmtMoney(row.day_ahead_fee as number | null),
    },
    {
      key: 'real_time_fee',
      header: '实时费用 (¥)',
      align: 'right',
      render: (row: DataRow) => fmtMoney(row.real_time_fee as number | null),
    },
    {
      key: 'total_energy_fee',
      header: '总电费 (¥)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => (
        <span className="font-semibold">{fmtMoney(row.total_energy_fee as number | null)}</span>
      ),
    },
    {
      key: 'energy_avg_price',
      header: '均价 (¥/MWh)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => {
        const v = row.energy_avg_price as number | null;
        return v != null ? fmt(v, 2) : '-';
      },
    },
    // ── Sprint 3：偏差费用条件格式 ──
    {
      key: 'deviation_recovery_fee',
      header: '偏差费用 (¥)',
      align: 'right',
      sortable: true,
      render: (row: DataRow) => {
        const v = row.deviation_recovery_fee as number | null;
        if (v == null) return <span className="text-muted-foreground">-</span>;
        const isWarning = Math.abs(v) > DEVIATION_FEE_THRESHOLD;
        return (
          <span className={isWarning ? 'font-bold text-red-700 dark:text-red-400 bg-red-50 dark:bg-red-500/15 px-1.5 py-0.5 rounded' : ''}>
            {fmtMoney(v)}
            {isWarning && ' ⚠'}
          </span>
        );
      },
    },
    // ── Sprint 3：结算核对标记 ──
    {
      key: '_verified',
      header: '核对状态',
      align: 'center',
      render: (row: DataRow) => {
        const id = row.id as string;
        const verified = verifiedSet.has(id);
        return (
          <Button
            variant="ghost"
            size="sm"
            className={`h-auto p-1 gap-1 text-xs ${verified ? 'text-emerald-700' : 'text-muted-foreground'}`}
            onClick={() => toggleVerified(id)}
          >
            {verified ? (
              <>
                <CheckCircle2 className="h-3.5 w-3.5" />
                已核对
              </>
            ) : (
              <>
                <Circle className="h-3.5 w-3.5" />
                未核对
              </>
            )}
          </Button>
        );
      },
    },
  ];

  // 移动端卡片字段（窄屏替代 9 列宽表；保留日期/核对交互）
  const mobileCardFields: MobileCardField<DataRow>[] = [
    {
      primary: true,
      render: (row) => {
        const date = (row.operating_date as string)?.slice(0, 10);
        return (
          <Button variant="link" className="h-auto p-0 font-medium" onClick={() => setSelectedDate(date)}>
            {date}
          </Button>
        );
      },
    },
    { label: '版本', render: (row) => <Badge variant="outline">{row.version as string}</Badge> },
    { label: '总电费', emphasize: true, render: (row) => <span className="font-semibold">{fmtMoney(row.total_energy_fee as number | null)}</span> },
    { label: '合同', render: (row) => fmtMoney(row.contract_fee as number | null) },
    { label: '日前', render: (row) => fmtMoney(row.day_ahead_fee as number | null) },
    { label: '实时', render: (row) => fmtMoney(row.real_time_fee as number | null) },
    { label: '均价', render: (row) => { const v = row.energy_avg_price as number | null; return v != null ? fmt(v, 2) : '-'; } },
    {
      label: '核对',
      render: (row) => {
        const id = row.id as string;
        const verified = verifiedSet.has(id);
        return (
          <Button variant="ghost" size="sm" className={`h-auto p-1 gap-1 text-xs ${verified ? 'text-emerald-700' : 'text-muted-foreground'}`} onClick={() => toggleVerified(id)}>
            {verified ? '✓ 已核对' : '未核对'}
          </Button>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="日结算"
        description="批发日结算 · 48 时段分时明细 · 费用汇总分析"
        actions={
          canWrite ? (
            <Button variant="outline" onClick={onGenerate} disabled={generating}>
              {generating ? (
                <RefreshCw className="mr-1 h-4 w-4 animate-spin" />
              ) : (
                <BarChart3 className="mr-1 h-4 w-4" />
              )}
              {generating ? '生成中...' : '生成演示数据'}
            </Button>
          ) : undefined
        }
      />

      {notice && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{notice}</AlertDescription>
        </Alert>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* ── Sprint 3：结算全景看板（批发侧口径；零售收入属零售月结批次，不在本页造数）── */}
      {panoramaData && (
        <Card className="border-blue-200 bg-gradient-to-r from-blue-50/50 to-indigo-50/50 dark:border-blue-500/30 dark:from-blue-500/10 dark:to-indigo-500/10">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm font-medium">
              <ArrowRightLeft className="h-4 w-4 text-blue-600 dark:text-blue-400" />
              结算全景看板 · 批发侧费用结构
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-5">
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">批发侧总费用</p>
                <p className="text-lg font-bold text-red-700 dark:text-red-400">¥ {fmtMoney(panoramaData.wholesale)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">合同电费（占比 {panoramaData.contractRatio.toFixed(1)}%）</p>
                <p className="text-lg font-bold text-blue-600 dark:text-blue-400">¥ {fmtMoney(panoramaData.contract)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">现货偏差电费（日前+实时）</p>
                <p className="text-lg font-bold text-amber-700 dark:text-amber-400">¥ {fmtMoney(panoramaData.spotDeviation)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">偏差考核费</p>
                <p className="text-lg font-bold text-orange-700 dark:text-orange-400">¥ {fmtMoney(panoramaData.assessFee)}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground">综合购电均价 (¥/MWh)</p>
                <p className="text-lg font-bold text-indigo-600 dark:text-indigo-400">{summary ? fmt(summary.avgPrice, 2) : '-'}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 偏差费用超标预警条 */}
      {deviationWarningCount > 0 && (
        <Alert className="border-orange-300 dark:border-orange-500/30 bg-orange-50 dark:bg-orange-500/15">
          <ShieldAlert className="h-4 w-4 text-orange-600" />
          <AlertDescription className="text-orange-800">
            <span className="font-semibold">{deviationWarningCount}</span> 天偏差费用超过 ¥{fmtMoney(DEVIATION_FEE_THRESHOLD)} 预警阈值
          </AlertDescription>
        </Alert>
      )}

      {/* 汇总卡片 */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-5">
          <StatCard title="总电费 (¥)" value={fmtMoney(summary.totalEnergyFee)} icon={<DollarSign className="h-4 w-4" />} />
          <StatCard title="均价 (¥/MWh)" value={fmt(summary.avgPrice, 2)} icon={<Wallet className="h-4 w-4" />} />
          <StatCard title="合同费用 (¥)" value={fmtMoney(summary.totalContractFee)} icon={<TrendingDown className="h-4 w-4" />} />
          <StatCard title="日前费用 (¥)" value={fmtMoney(summary.totalDAFee)} icon={<TrendingUp className="h-4 w-4" />} />
          <StatCard title="实时费用 (¥)" value={fmtMoney(summary.totalRTFee)} icon={<PieChart className="h-4 w-4" />} />
        </div>
      )}

      {/* ── Sprint 3：电费账单瀑布图 + 费用组成 ── */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* 瀑布图 */}
        <ChartContainer title="电费账单瀑布图" minHeight={280}>
          {waterfallData.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={waterfallData} margin={{ top: 8, right: 40, bottom: 8, left: 8 }}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" horizontal={false} />
                <XAxis type="category" dataKey="name" tick={{ fontSize: 12 }} />
                <YAxis tick={{ fontSize: 11 }} width={72} />
                <RechartsTooltip
                  content={({ active, payload }) => {
                    if (!active || !payload?.length) return null;
                    const d = payload[0]?.payload as (typeof waterfallData)[number];
                    if (!d) return null;
                    return (
                      <div className="rounded-md border bg-popover text-popover-foreground px-3 py-2 text-xs shadow-lg">
                        <p className="font-semibold">{d.name}</p>
                        <p>金额: ¥ {fmtMoney(Math.abs(d.amount))}</p>
                        <p className="text-muted-foreground">
                          {d.name === '净收益'
                            ? (d.amount >= 0 ? '盈利' : '亏损')
                            : d.amount >= 0 ? '增加项' : '减少项'}
                        </p>
                      </div>
                    );
                  }}
                />
                <Legend />
                {/* 不可见的底部偏移 */}
                <Bar dataKey="base" stackId="wf" fill="transparent" isAnimationActive={false} name="" />
                {/* 可见的金额 */}
                <Bar dataKey="amount" stackId="wf" isAnimationActive={false} name="金额 (¥)">
                  {waterfallData.map((entry, idx) => (
                    <Cell key={idx} fill={entry.fill} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <EmptyState compact className="h-full" title="暂无数据" />
          )}
        </ChartContainer>

        {/* 费用组成 */}
        <ChartContainer title="费用组成" minHeight={260}>
          {feeCompositionData.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={feeCompositionData} layout="vertical" margin={{ top: 8, right: 40, bottom: 8, left: 8 }}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" horizontal={false} />
                <XAxis type="number" tick={{ fontSize: 12 }} />
                <YAxis type="category" dataKey="name" tick={{ fontSize: 12 }} width={80} />
                <RechartsTooltip
                  content={({ active, payload }) => {
                    if (!active || !payload?.length) return null;
                    const d = payload[0].payload as any;
                    return (
                      <div className="rounded-md border bg-popover text-popover-foreground px-3 py-2 text-xs shadow-lg">
                        <p className="font-semibold">{d.name}</p>
                        <p>¥ {fmtMoney(d.value)}</p>
                      </div>
                    );
                  }}
                />
                <Bar dataKey="value" name="费用" isAnimationActive={false}>
                  {feeCompositionData.map((_, idx) => (
                    <Cell key={idx} fill={feeCompositionData[idx].color} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <EmptyState compact className="h-full" title="暂无数据" />
          )}
        </ChartContainer>
      </div>

      {/* 近 30 日趋势 */}
      <ChartContainer title="近 30 日总电费趋势" minHeight={260}>
        {trendData.length > 0 ? (
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart data={trendData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="date" interval={6} tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} width={64} />
              <RechartsTooltip content={<CustomTooltip unit="¥" />} />
              <Legend />
              <Bar dataKey="总电费" fill="#f59e0b" isAnimationActive={false} name="总电费" />
              <Line type="monotone" dataKey="均价" stroke="#2563eb" strokeWidth={2} dot={false} isAnimationActive={false} name="均价" />
            </ComposedChart>
          </ResponsiveContainer>
        ) : (
          <EmptyState compact className="h-full" title="暂无趋势数据" />
        )}
      </ChartContainer>

      {/* 结算明细表格 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <BarChart3 className="h-4 w-4 text-primary" />
            结算明细
            <span className="text-xs font-normal text-muted-foreground ml-2">
              {listData?.items?.length ?? 0} 条记录
            </span>
            {verifiedSet.size > 0 && (
              <Badge variant="outline" className="ml-2 text-[10px] text-emerald-700 border-emerald-300">
                已核对 {verifiedSet.size} 条
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable
            columns={columns}
            data={(listData?.items as DataRow[]) ?? []}
            rowKey="id"
            pageSize={10}
            showPagination
            loading={listLoading}
            mobileCardFields={mobileCardFields}
          />
        </CardContent>
      </Card>

      {/* 分时明细 */}
      {detail && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <PieChart className="h-4 w-4 text-primary" />
              {detail.operating_date?.slice(0, 10)} 48 时段分时明细
              <Badge variant="outline" className="ml-2">{version === 'PRELIMINARY' ? '初步结算' : '正式结算'}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3 mb-4">
              <StatCard title="总电量 (MWh)" value={fmt((detail.period_details ?? []).reduce((s: number, p: any) => s + p.volume_mwh, 0), 2)} />
              <StatCard title="总电费 (¥)" value={fmtMoney(detail.total_energy_fee)} />
              <StatCard title="均价 (¥/MWh)" value={detail.energy_avg_price != null ? fmt(detail.energy_avg_price, 2) : '-'} />
            </div>

            {periodChartData.length > 0 && (
              <div className="h-72 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <ComposedChart data={periodChartData} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis dataKey="t" interval={11} tick={{ fontSize: 11 }} />
                    <YAxis tick={{ fontSize: 11 }} width={64} />
                    <RechartsTooltip content={<CustomTooltip unit="MWh" />} />
                    <Bar dataKey="电量" fill="#2563eb" isAnimationActive={false} name="电量" />
                    <Line type="monotone" dataKey="价格" stroke="#f59e0b" strokeWidth={2} dot={false} isAnimationActive={false} name="价格" />
                  </ComposedChart>
                </ResponsiveContainer>
              </div>
            )}

            {/* 时段明细表格 */}
            <div className="mt-4">
              <DataTable
                columns={[
                  { key: 'period', header: '时段', render: (row: DataRow) => `${row.period} (${pointTime((row.period as number) - 1)})` },
                  { key: 'volume_mwh', header: '电量 (MWh)', align: 'right', sortable: true, render: (row: DataRow) => fmt(row.volume_mwh as number, 2) },
                  { key: 'price', header: '价格 (¥/MWh)', align: 'right', sortable: true, render: (row: DataRow) => fmt(row.price as number, 2) },
                  { key: 'fee', header: '费用 (¥)', align: 'right', sortable: true, render: (row: DataRow) => fmtMoney(row.fee as number) },
                ]}
                data={(detail.period_details ?? []) as DataRow[]}
                rowKey="period"
                pageSize={48}
                showPagination
              />
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
