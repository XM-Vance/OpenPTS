'use client';

import { useState, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Building2,
  FileText,
  Calculator,
  FolderOpen,
  Activity,
  AlertTriangle,
  TrendingUp,
  Sun,
  ArrowLeft,
} from 'lucide-react';
import { apiClient, extractErrorMessage } from '@/lib/api/client';

async function fetchCustomer360(id: string) {
  try {
    const res = await apiClient.get(`/customers/${id}/360`);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to fetch customer 360 data'));
  }
}

type TabKey = 'contracts' | 'settlements' | 'documents' | 'load' | 'alerts' | 'profit' | 'stations';

const TABS: { key: TabKey; label: string; icon: React.ElementType }[] = [
  { key: 'contracts', label: '合同', icon: FileText },
  { key: 'settlements', label: '结算', icon: Calculator },
  { key: 'documents', label: '文档', icon: FolderOpen },
  { key: 'load', label: '负荷', icon: Activity },
  { key: 'alerts', label: '告警', icon: AlertTriangle },
  { key: 'profit', label: '利润', icon: TrendingUp },
  { key: 'stations', label: '电站', icon: Sun },
];

/* ── Tab Content Renderers ── */
function ContractsTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无合同数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>合同编号</TableHead>
            <TableHead>合同类型</TableHead>
            <TableHead>签约电量</TableHead>
            <TableHead>开始日期</TableHead>
            <TableHead>结束日期</TableHead>
            <TableHead>状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((c: any, i: number) => (
            <TableRow key={c.id ?? i}>
              <TableCell className="font-medium">{c.contract_no ?? c.id ?? '-'}</TableCell>
              <TableCell>{c.contract_type ?? '-'}</TableCell>
              <TableCell>{c.volume ?? '-'} MWh</TableCell>
              <TableCell>{c.start_date ?? '-'}</TableCell>
              <TableCell>{c.end_date ?? '-'}</TableCell>
              <TableCell>
                <Badge variant={c.status === 'active' ? 'default' : 'secondary'}>
                  {c.status === 'active' ? '生效中' : c.status ?? '-'}
                </Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function SettlementsTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无结算数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>结算月份</TableHead>
            <TableHead>结算电量</TableHead>
            <TableHead>结算金额</TableHead>
            <TableHead>偏差电量</TableHead>
            <TableHead>偏差费用</TableHead>
            <TableHead>状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((s: any, i: number) => (
            <TableRow key={s.id ?? i}>
              <TableCell className="font-medium">{s.month ?? '-'}</TableCell>
              <TableCell>{s.electricity ?? '-'} MWh</TableCell>
              <TableCell>¥{s.amount ?? '-'}</TableCell>
              <TableCell>{s.deviation ?? '-'}</TableCell>
              <TableCell>¥{s.deviation_cost ?? '-'}</TableCell>
              <TableCell>
                <Badge variant="secondary">{s.status ?? '-'}</Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function DocumentsTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无文档数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>文件名</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>上传时间</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((d: any, i: number) => (
            <TableRow key={d.id ?? i}>
              <TableCell className="font-medium">{d.filename ?? '-'}</TableCell>
              <TableCell>{d.doc_type ?? '-'}</TableCell>
              <TableCell>
                <Badge variant="secondary">{d.status ?? '-'}</Badge>
              </TableCell>
              <TableCell>{d.created_at ?? '-'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function LoadTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无负荷数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>日期</TableHead>
            <TableHead>最大负荷</TableHead>
            <TableHead>最小负荷</TableHead>
            <TableHead>用电量</TableHead>
            <TableHead>负荷率</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((l: any, i: number) => (
            <TableRow key={i}>
              <TableCell className="font-medium">{l.date ?? '-'}</TableCell>
              <TableCell>{l.max_load ?? '-'} kW</TableCell>
              <TableCell>{l.min_load ?? '-'} kW</TableCell>
              <TableCell>{l.electricity ?? '-'} MWh</TableCell>
              <TableCell>{l.load_rate ?? '-'}%</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function AlertsTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无告警数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>告警类型</TableHead>
            <TableHead>告警内容</TableHead>
            <TableHead>级别</TableHead>
            <TableHead>时间</TableHead>
            <TableHead>状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((a: any, i: number) => (
            <TableRow key={a.id ?? i}>
              <TableCell className="font-medium">{a.alert_type ?? '-'}</TableCell>
              <TableCell>{a.message ?? a.content ?? '-'}</TableCell>
              <TableCell>
                <Badge
                  variant={
                    a.severity === 'critical'
                      ? 'destructive'
                      : a.severity === 'warning'
                        ? 'default'
                        : 'secondary'
                  }
                >
                  {a.severity === 'critical' ? '严重' : a.severity === 'warning' ? '警告' : '提示'}
                </Badge>
              </TableCell>
              <TableCell>{a.created_at ?? a.time ?? '-'}</TableCell>
              <TableCell>
                <Badge variant={a.resolved ? 'secondary' : 'default'}>
                  {a.resolved ? '已处理' : '未处理'}
                </Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function ProfitTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无利润数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>月份</TableHead>
            <TableHead>收入</TableHead>
            <TableHead>成本</TableHead>
            <TableHead>毛利润</TableHead>
            <TableHead>利润率</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((p: any, i: number) => (
            <TableRow key={i}>
              <TableCell className="font-medium">{p.month ?? '-'}</TableCell>
              <TableCell>¥{p.revenue ?? '-'}</TableCell>
              <TableCell>¥{p.cost ?? '-'}</TableCell>
              <TableCell className={Number(p.profit) >= 0 ? 'text-emerald-600' : 'text-red-600'}>
                ¥{p.profit ?? '-'}
              </TableCell>
              <TableCell>{p.margin ?? '-'}%</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function StationsTab({ data }: { data: any[] }) {
  if (!data || data.length === 0) {
    return <div className="py-8 text-center text-sm text-muted-foreground">暂无电站数据</div>;
  }
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>电站名称</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>装机容量</TableHead>
            <TableHead>并网日期</TableHead>
            <TableHead>状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((s: any, i: number) => (
            <TableRow key={s.id ?? i}>
              <TableCell className="font-medium">{s.name ?? '-'}</TableCell>
              <TableCell>{s.type ?? '-'}</TableCell>
              <TableCell>{s.capacity ?? '-'} kW</TableCell>
              <TableCell>{s.grid_date ?? '-'}</TableCell>
              <TableCell>
                <Badge variant={s.status === 'active' ? 'default' : 'secondary'}>
                  {s.status === 'active' ? '运行中' : s.status ?? '-'}
                </Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

/* ── Main Page ── */
export default function Customer360Page() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;
  const [activeTab, setActiveTab] = useState<TabKey>('contracts');

  const { data, isLoading, error } = useQuery({
    queryKey: ['customer-360', id],
    queryFn: () => fetchCustomer360(id),
    enabled: !!id,
  });

  const customer = data?.customer ?? data?.info ?? null;
  const tabData = useMemo(() => {
    if (!data) return {} as Record<TabKey, any[]>;
    return {
      contracts: data.contracts ?? [],
      settlements: data.settlements ?? [],
      documents: data.documents ?? [],
      load: data.load ?? data.loads ?? [],
      alerts: data.alerts ?? [],
      profit: data.profit ?? data.profits ?? [],
      stations: data.stations ?? [],
    } as Record<TabKey, any[]>;
  }, [data]);

  if (isLoading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center text-sm text-muted-foreground">
        加载中…
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <Button variant="outline" onClick={() => router.back()}>
          <ArrowLeft className="mr-1 h-4 w-4" />
          返回
        </Button>
        <div className="rounded-lg border border-red-200 bg-red-50 p-6 text-center text-red-700">
          加载客户数据失败：{error instanceof Error ? error.message : '未知错误'}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Back button */}
      <Button variant="outline" onClick={() => router.push('/customers')}>
        <ArrowLeft className="mr-1 h-4 w-4" />
        返回客户列表
      </Button>

      {/* Customer Info Card */}
      {customer && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5 text-indigo-500" />
              {customer.user_name ?? customer.name ?? '客户详情'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm sm:grid-cols-3 lg:grid-cols-4">
              {customer.short_name && (
                <div>
                  <span className="text-muted-foreground">简称：</span>
                  <span className="font-medium">{customer.short_name}</span>
                </div>
              )}
              {customer.location && (
                <div>
                  <span className="text-muted-foreground">所在地：</span>
                  <span className="font-medium">{customer.location}</span>
                </div>
              )}
              {customer.manager && (
                <div>
                  <span className="text-muted-foreground">客户经理：</span>
                  <span className="font-medium">{customer.manager}</span>
                </div>
              )}
              {customer.source && (
                <div>
                  <span className="text-muted-foreground">来源：</span>
                  <span className="font-medium">{customer.source}</span>
                </div>
              )}
              {customer.industry && (
                <div>
                  <span className="text-muted-foreground">行业：</span>
                  <span className="font-medium">{customer.industry}</span>
                </div>
              )}
              {customer.status && (
                <div>
                  <span className="text-muted-foreground">状态：</span>
                  <Badge variant="secondary">{customer.status}</Badge>
                </div>
              )}
              {customer.tags && customer.tags.length > 0 && (
                <div className="col-span-full">
                  <span className="text-muted-foreground">标签：</span>
                  <span className="ml-1 inline-flex gap-1">
                    {customer.tags.map((tag: any, i: number) => (
                      <Badge key={i} variant="outline" className="text-xs">
                        {typeof tag === 'string' ? tag : tag.name}
                      </Badge>
                    ))}
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <div className="border-b">
        <nav className="-mb-px flex gap-1 overflow-x-auto" aria-label="Tabs">
          {TABS.map((tab) => {
            const Icon = tab.icon;
            const count = tabData[tab.key]?.length ?? 0;
            return (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={`flex items-center gap-1.5 whitespace-nowrap border-b-2 px-4 py-2 text-sm font-medium transition-colors ${
                  activeTab === tab.key
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-muted-foreground hover:border-gray-300 hover:text-foreground'
                }`}
              >
                <Icon className="h-4 w-4" />
                {tab.label}
                {count > 0 && (
                  <span className="ml-1 rounded-full bg-gray-100 px-1.5 py-0.5 text-xs text-muted-foreground">
                    {count}
                  </span>
                )}
              </button>
            );
          })}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'contracts' && <ContractsTab data={tabData.contracts} />}
      {activeTab === 'settlements' && <SettlementsTab data={tabData.settlements} />}
      {activeTab === 'documents' && <DocumentsTab data={tabData.documents} />}
      {activeTab === 'load' && <LoadTab data={tabData.load} />}
      {activeTab === 'alerts' && <AlertsTab data={tabData.alerts} />}
      {activeTab === 'profit' && <ProfitTab data={tabData.profit} />}
      {activeTab === 'stations' && <StationsTab data={tabData.stations} />}
    </div>
  );
}
