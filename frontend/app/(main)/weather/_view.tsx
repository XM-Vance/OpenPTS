'use client';

// 气象数据 —— 负荷预测的气象输入面板。
// 数据分层：主数据来自 weather_actuals（PR-A 后由 Open-Meteo 自动采集的真实观测）；
// 负荷系数（load_factor）来自 weather_data 演示表，标注 DemoBadge「估算」。
import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Area,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
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
import {
  genWeatherDemo,
  listWeather,
  getWeatherLocations,
  getWeatherActualsRange,
  type WeatherLocation,
  type DailyWeatherSummary,
} from '@/lib/api/weather';
import { Thermometer, Droplets, Wind, CloudRain, AlertTriangle } from 'lucide-react';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';

type Metric = 'temp' | 'humidity' | 'wind';

export default function WeatherPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('load_management:write');

  // 站点选择：空 = 全部地点（actuals range 需指定站点，默认取首个动态站点）。
  const [location, setLocation] = useState('');
  const [metric, setMetric] = useState<Metric>('temp');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: locations } = useQuery({
    queryKey: ['weather-locations'],
    queryFn: () => getWeatherLocations(),
  });

  // 实际生效站点：用户未选时取首个动态站点。
  const activeLocation = location || locations?.[0]?.name || '';
  const { data: rangeData, isLoading } = useQuery({
    queryKey: ['weather-actuals-range', activeLocation],
    queryFn: () => getWeatherActualsRange(activeLocation, 14),
    enabled: !!activeLocation,
  });

  // 负荷系数补充（来自 weather_data 演示表，仅用于负荷系数 tab）。
  const { data: wfData } = useQuery({
    queryKey: ['weather', activeLocation],
    queryFn: () => listWeather({ days: 14, location: activeLocation }),
    enabled: !!activeLocation,
  });

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genWeatherDemo();
      qc.invalidateQueries({ queryKey: ['weather'] });
    } catch (e) {
      setError(extractErrorMessage(e));
    } finally {
      setBusy(false);
    }
  };

  // 真实观测（actuals），按日期升序。
  const items = useMemo(
    () => (rangeData?.items ?? []).slice().sort((a, b) => a.date.localeCompare(b.date)),
    [rangeData],
  );

  // 负荷系数 map（date → load_factor），来自演示表，按日期补到走势里。
  const loadFactorMap = useMemo(() => {
    const m = new Map<string, number>();
    for (const w of wfData?.items ?? []) {
      if (w.load_factor != null) m.set(w.obs_date.slice(0, 10), w.load_factor);
    }
    return m;
  }, [wfData]);

  // KPI 聚合（真实 actuals）。
  const stats = useMemo(() => {
    if (items.length === 0) {
      return { avgTemp: 0, maxTemp: 0, minTemp: 0, humidity: 0, precip: 0, wind: 0 };
    }
    const sum = (f: (i: DailyWeatherSummary) => number) =>
      items.reduce((s, i) => s + f(i), 0);
    return {
      avgTemp: sum((i) => i.avg_temp) / items.length,
      maxTemp: Math.max(...items.map((i) => i.max_temp)),
      minTemp: Math.min(...items.map((i) => i.min_temp)),
      humidity: sum((i) => i.humidity) / items.length,
      precip: sum((i) => i.avg_precipitation),
      wind: sum((i) => i.wind_speed) / items.length,
    };
  }, [items]);

  // 走势图数据：actuals + 负荷系数补充。
  const trend = useMemo(
    () =>
      items.map((i) => ({
        date: i.date.slice(5).replace('-', '/'),
        avg: i.avg_temp,
        high: i.max_temp,
        low: i.min_temp,
        humidity: i.humidity,
        wind: i.wind_speed,
        precip: i.avg_precipitation,
        loadFactor: (loadFactorMap.get(i.date) ?? 0) * 100,
      })),
    [items, loadFactorMap],
  );

  // 极端天气预警（actuals 阈值判断）。
  const alerts = useMemo(() => {
    const r: { type: string; level: string; message: string; date: string }[] = [];
    for (const i of items) {
      if (i.max_temp >= 38) {
        r.push({ type: '高温预警', level: 'red', message: `${i.date} 最高温 ${i.max_temp.toFixed(1)}℃，超 38℃ 阈值`, date: i.date });
      }
      if (i.min_temp <= 2) {
        r.push({ type: '低温预警', level: 'blue', message: `${i.date} 最低温 ${i.min_temp.toFixed(1)}℃，低于 2℃`, date: i.date });
      }
      if (i.avg_precipitation >= 50) {
        r.push({ type: '暴雨预警', level: 'orange', message: `${i.date} 降水 ${i.avg_precipitation.toFixed(1)}mm，超 50mm`, date: i.date });
      }
      if (i.wind_speed >= 60) {
        r.push({ type: '大风预警', level: 'yellow', message: `${i.date} 风速 ${i.wind_speed.toFixed(1)}km/h，超 60km/h`, date: i.date });
      }
    }
    return r;
  }, [items]);

  const levelColor: Record<string, string> = {
    red: 'border-red-500 bg-red-50 text-red-700',
    orange: 'border-orange-500 bg-orange-50 text-orange-700',
    yellow: 'border-yellow-500 bg-yellow-50 text-yellow-700',
    blue: 'border-blue-500 bg-blue-50 text-blue-700',
  };

  const noStation = !activeLocation;

  return (
    <div className="space-y-4">
      {/* 页头 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">气象数据</h1>
          <p className="text-sm text-muted-foreground">
            气温 / 湿度 / 降水 / 风速 —— 负荷预测的关键输入（Open-Meteo 自动采集）
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            className={SELECT_CLASS}
          >
            <option value="">{locations?.[0]?.name ? `默认（${locations[0].name}）` : '全部地点'}</option>
            {locations?.slice(1).map((loc: WeatherLocation) => (
              <option key={loc.location_id} value={loc.name}>
                {loc.name}
              </option>
            ))}
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

      {noStation ? (
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">
              暂无气象站点。数据由系统每日 17:00 自动采集；若无站点，说明采集任务尚未运行或未配置。
            </p>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* KPI（真实 actuals 聚合） */}
          <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
            <StatCard title="平均气温" value={`${stats.avgTemp.toFixed(1)} ℃`} icon={<Thermometer className="h-4 w-4" />} />
            <StatCard title="最高温" value={`${stats.maxTemp.toFixed(1)} ℃`} icon={<Thermometer className="h-4 w-4" />} />
            <StatCard title="最低温" value={`${stats.minTemp.toFixed(1)} ℃`} icon={<Thermometer className="h-4 w-4" />} />
            <StatCard title="平均湿度" value={`${stats.humidity.toFixed(0)}%`} icon={<Droplets className="h-4 w-4" />} />
            <StatCard title="累计降水" value={`${stats.precip.toFixed(1)} mm`} icon={<CloudRain className="h-4 w-4" />} />
            <StatCard title="平均风速" value={`${stats.wind.toFixed(1)} km/h`} icon={<Wind className="h-4 w-4" />} />
          </div>

          {/* 极端天气预警 */}
          {alerts.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <AlertTriangle className="h-4 w-4 text-amber-500" />
                  极端天气预警
                  <Badge variant="destructive">{alerts.length}</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-2 md:grid-cols-2">
                  {alerts.map((a, i) => (
                    <div key={i} className={`rounded-md border p-3 ${levelColor[a.level] ?? 'border-gray-300 bg-gray-50'}`}>
                      <div className="flex items-center gap-2 text-sm font-medium">
                        <AlertTriangle className="h-3.5 w-3.5" />
                        {a.type}
                        <span className="ml-auto text-xs opacity-70">{a.date}</span>
                      </div>
                      <p className="mt-1 text-xs">{a.message}</p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* 14 日气象走势（温/湿/风 切换；负荷系数 tab 标 DemoBadge） */}
          <ChartContainer
            title="最近 14 日气象走势"
            actions={
              <div className="flex items-center gap-1">
                {(['temp', 'humidity', 'wind'] as Metric[]).map((m) => (
                  <Button
                    key={m}
                    variant={metric === m ? 'default' : 'outline'}
                    size="sm"
                    className="h-7 text-xs"
                    onClick={() => setMetric(m)}
                  >
                    {m === 'temp' ? '温度' : m === 'humidity' ? '湿度' : '风速'}
                  </Button>
                ))}
                {metric === 'wind' && (
                  <DemoBadge className="ml-1" tooltip="负荷系数来自演示数据（估算），非真实观测" />
                )}
              </div>
            }
          >
            {trend.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {isLoading ? '加载中...' : '暂无观测数据，等待每日 17:00 自动采集'}
              </p>
            ) : metric === 'temp' ? (
              <ResponsiveContainer width="100%" height={280}>
                <ComposedChart data={trend} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
                  <YAxis yAxisId="temp" tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="℃" />
                  <YAxis yAxisId="precip" orientation="right" tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="mm" />
                  <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number, n: string) => (n === '降水' ? `${v.toFixed(1)} mm` : `${v.toFixed(1)} ℃`)} />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Area yAxisId="temp" type="monotone" dataKey="high" name="最高" stroke="#f97316" fill="#fed7aa" strokeWidth={2} isAnimationActive={false} />
                  <Area yAxisId="temp" type="monotone" dataKey="low" name="最低" stroke="#3b82f6" fill="#dbeafe" strokeWidth={2} isAnimationActive={false} />
                  <Line yAxisId="temp" type="monotone" dataKey="avg" name="均温" stroke="#8b5cf6" strokeWidth={2} dot={false} strokeDasharray="5 5" isAnimationActive={false} />
                </ComposedChart>
              </ResponsiveContainer>
            ) : metric === 'humidity' ? (
              <ResponsiveContainer width="100%" height={280}>
                <ComposedChart data={trend} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
                  <YAxis yAxisId="humidity" tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="%" />
                  <YAxis yAxisId="precip" orientation="right" tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="mm" />
                  <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number, n: string) => (n === '降水' ? `${v.toFixed(1)} mm` : `${v.toFixed(0)}%`)} />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Area yAxisId="humidity" type="monotone" dataKey="humidity" name="湿度" stroke="#06b6d4" fill="#cffafe" strokeWidth={2} isAnimationActive={false} />
                  <Line yAxisId="precip" type="monotone" dataKey="precip" name="降水" stroke="#3b82f6" strokeWidth={2} dot={false} isAnimationActive={false} />
                </ComposedChart>
              </ResponsiveContainer>
            ) : (
              <ResponsiveContainer width="100%" height={280}>
                <ComposedChart data={trend} margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
                  <YAxis yAxisId="wind" tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="km/h" />
                  <YAxis yAxisId="lf" orientation="right" tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="%" />
                  <Tooltip contentStyle={{ fontSize: 12 }} formatter={(v: number, n: string) => (n === '负荷系数' ? `${v.toFixed(1)}%` : `${v.toFixed(1)} km/h`)} />
                  <Legend wrapperStyle={{ fontSize: 12 }} />
                  <Area yAxisId="wind" type="monotone" dataKey="wind" name="风速" stroke="#f97316" fill="#fed7aa" strokeWidth={2} isAnimationActive={false} />
                  <Line yAxisId="lf" type="monotone" dataKey="loadFactor" name="负荷系数(估算)" stroke="#8b5cf6" strokeWidth={2} dot={false} strokeDasharray="5 5" isAnimationActive={false} />
                </ComposedChart>
              </ResponsiveContainer>
            )}
          </ChartContainer>

          {/* 气象数据明细表 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">气象数据明细（近 14 日）</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>日期</TableHead>
                      <TableHead className="text-right">均温 ℃</TableHead>
                      <TableHead className="text-right">最高 ℃</TableHead>
                      <TableHead className="text-right">最低 ℃</TableHead>
                      <TableHead className="text-right">湿度 %</TableHead>
                      <TableHead className="text-right">降水 mm</TableHead>
                      <TableHead className="text-right">风速 km/h</TableHead>
                      <TableHead>天气</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {isLoading && (
                      <TableRow>
                        <TableCell colSpan={8} className="text-center text-muted-foreground">加载中...</TableCell>
                      </TableRow>
                    )}
                    {items.slice().reverse().map((i) => (
                      <TableRow key={i.date}>
                        <TableCell className="font-medium">{i.date}</TableCell>
                        <TableCell className="text-right text-purple-600">{i.avg_temp.toFixed(1)}</TableCell>
                        <TableCell className="text-right text-orange-600">{i.max_temp.toFixed(1)}</TableCell>
                        <TableCell className="text-right text-blue-600">{i.min_temp.toFixed(1)}</TableCell>
                        <TableCell className="text-right">{i.humidity.toFixed(0)}</TableCell>
                        <TableCell className="text-right">{i.avg_precipitation.toFixed(1)}</TableCell>
                        <TableCell className="text-right">{i.wind_speed.toFixed(1)}</TableCell>
                        <TableCell>
                          <span className="text-base">{i.weather_icon}</span>
                          <span className="ml-1 text-xs text-muted-foreground">{i.weather_type}</span>
                        </TableCell>
                      </TableRow>
                    ))}
                    {items.length === 0 && !isLoading && (
                      <TableRow>
                        <TableCell colSpan={8} className="text-center text-muted-foreground">暂无观测数据</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
