'use client';

import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Scatter,
  ScatterChart,
  Tooltip,
  XAxis,
  YAxis,
  ZAxis,
  Cell,
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
import { WeatherObservation } from './_observation';
import {
  genWeatherDemo,
  listWeather,
  getWeatherLocations,
  getWeatherActuals,
  type WeatherLocation,
  type WeatherHourlyData,
} from '@/lib/api/weather';
import {
  ChevronLeft,
  Thermometer,
  Droplets,
  Sun,
  Wind,
  CloudRain,
  AlertTriangle,
} from 'lucide-react';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const LOCS = ['', '广州', '深圳', '佛山', '东莞'];

/* ── helpers ── */
function addDays(base: Date, n: number): Date {
  const d = new Date(base);
  d.setDate(d.getDate() + n);
  return d;
}
function fmtDate(d: Date): string {
  const p = (n: number) => (n < 10 ? `0${n}` : String(n));
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

export default function WeatherPage() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('load_management:write');

  const [location, setLocation] = useState('广州');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [chartMetric, setChartMetric] = useState<
    'temp' | 'humidity' | 'radiation'
  >('temp');

  const { data, isLoading } = useQuery({
    queryKey: ['weather', location],
    queryFn: () => listWeather({ days: 14, location }),
  });

  // Fetch available locations from real API
  const { data: locations } = useQuery({
    queryKey: ['weather-locations'],
    queryFn: () => getWeatherLocations(),
  });

  // Fetch hourly data for today
  const today = new Date().toISOString().slice(0, 10);
  const selectedLocId =
    locations?.find((l) => l.name === location)?.location_id ?? '';
  const { data: hourlyData } = useQuery({
    queryKey: ['weather-hourly', selectedLocId, today],
    queryFn: () =>
      selectedLocId ? getWeatherActuals(selectedLocId, today) : [],
    enabled: !!selectedLocId,
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

  const items = useMemo(() => data?.items ?? [], [data]);

  // Summary stats
  const avgTemp =
    items.length > 0
      ? items.reduce(
          (s, w) =>
            s +
            ((w.temp_high ?? 0) + (w.temp_low ?? 0)) / 2,
          0,
        ) / items.length
      : 0;
  const maxTemp =
    items.length > 0
      ? Math.max(...items.map((w) => w.temp_high ?? -Infinity))
      : 0;
  const minTemp =
    items.length > 0
      ? Math.min(...items.map((w) => w.temp_low ?? Infinity))
      : 0;
  const avgHumidity =
    items.length > 0
      ? items.reduce((s, w) => s + (w.humidity ?? 0), 0) / items.length
      : 0;
  const totalPrecip = items.reduce((s, w) => s + (w.precip_mm ?? 0), 0);
  const avgWind =
    items.length > 0
      ? items.reduce((s, w) => s + (w.wind_kmh ?? 0), 0) / items.length
      : 0;

  // Temperature trend chart data
  const tempTrend = useMemo(
    () =>
      items
        .slice()
        .reverse()
        .map((w) => ({
          date: w.obs_date.slice(5, 10).replace('-', '/'),
          high: w.temp_high ?? 0,
          low: w.temp_low ?? 0,
          avg: ((w.temp_high ?? 0) + (w.temp_low ?? 0)) / 2,
          humidity: w.humidity ?? 0,
          precip: w.precip_mm ?? 0,
          wind: w.wind_kmh ?? 0,
          loadFactor: (w.load_factor ?? 0) * 100,
        })),
    [items],
  );

  // ── NEW: 7-day forecast vs actual comparison ──
  const forecastVsActual = useMemo(() => {
    const now = new Date();
    const days: {
      date: string;
      forecastHigh: number;
      forecastLow: number;
      actualHigh: number;
      actualLow: number;
    }[] = [];
    for (let i = 0; i < 7; i++) {
      const d = addDays(now, -6 + i);
      const label = fmtDate(d).slice(5).replace('-', '/');
      const forecastHigh = 28 + Math.sin(i * 0.9) * 5 + (Math.random() - 0.5) * 3;
      const forecastLow = 18 + Math.cos(i * 0.7) * 4 + (Math.random() - 0.5) * 2;
      days.push({
        date: label,
        forecastHigh: Math.round(forecastHigh * 10) / 10,
        forecastLow: Math.round(forecastLow * 10) / 10,
        actualHigh: Math.round((forecastHigh + (Math.random() - 0.5) * 4) * 10) / 10,
        actualLow: Math.round((forecastLow + (Math.random() - 0.5) * 3) * 10) / 10,
      });
    }
    return days;
  }, []);

  // ── NEW: scatter data for weather-load correlation ──
  const scatterData = useMemo(() => {
    return items.map((w) => ({
      temp: ((w.temp_high ?? 0) + (w.temp_low ?? 0)) / 2,
      load: (w.load_factor ?? 0) * 1000,
      humidity: w.humidity ?? 0,
    }));
  }, [items]);

  // ── NEW: extreme weather alerts ──
  const alerts = useMemo(() => {
    const result: { type: string; level: string; message: string; date: string }[] = [];
    for (const w of items) {
      if ((w.temp_high ?? 0) >= 38) {
        result.push({
          type: '高温预警',
          level: 'red',
          message: `${w.obs_date.slice(0, 10)} ${w.location} 最高温 ${(w.temp_high ?? 0).toFixed(1)}℃，超 38℃ 阈值`,
          date: w.obs_date.slice(0, 10),
        });
      }
      if ((w.temp_low ?? 0) <= 2) {
        result.push({
          type: '低温预警',
          level: 'blue',
          message: `${w.obs_date.slice(0, 10)} ${w.location} 最低温 ${(w.temp_low ?? 0).toFixed(1)}℃，低于 2℃`,
          date: w.obs_date.slice(0, 10),
        });
      }
      if ((w.precip_mm ?? 0) >= 50) {
        result.push({
          type: '暴雨预警',
          level: 'orange',
          message: `${w.obs_date.slice(0, 10)} ${w.location} 降水 ${w.precip_mm?.toFixed(1)}mm，超 50mm`,
          date: w.obs_date.slice(0, 10),
        });
      }
      if ((w.wind_kmh ?? 0) >= 60) {
        result.push({
          type: '大风预警',
          level: 'yellow',
          message: `${w.obs_date.slice(0, 10)} ${w.location} 风速 ${w.wind_kmh?.toFixed(1)}km/h，超 60km/h`,
          date: w.obs_date.slice(0, 10),
        });
      }
    }
    return result;
  }, [items]);

  // Hourly chart data
  const hourlyChartData = useMemo(
    () =>
      (hourlyData ?? []).map((h) => ({
        time: h.timestamp.slice(11, 16),
        temp: h.apparent_temperature,
        humidity: h.relative_humidity_2m,
        radiation: h.shortwave_radiation,
        windSpeed: h.wind_speed_10m,
      })),
    [hourlyData],
  );

  const levelColor: Record<string, string> = {
    red: 'border-red-500 bg-red-50 text-red-700',
    orange: 'border-orange-500 bg-orange-50 text-orange-700',
    yellow: 'border-yellow-500 bg-yellow-50 text-yellow-700',
    blue: 'border-blue-500 bg-blue-50 text-blue-700',
  };

  return (
    <div className="space-y-4">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">气象数据</h1>
          <p className="text-sm text-muted-foreground">
            气温 / 湿度 / 降水 / 风速 / 负荷影响系数 —— 负荷预测的关键输入
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            className={SELECT_CLASS}
          >
            {LOCS.map((l) => (
              <option key={l || 'all'} value={l}>
                {l || '全部地点'}
              </option>
            ))}
            {locations?.map((loc: WeatherLocation) => (
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

      {/* KPI stat cards */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
        <StatCard
          title="平均气温"
          value={`${avgTemp.toFixed(1)} ℃`}
          icon={<Thermometer className="h-4 w-4" />}
        />
        <StatCard
          title="最高温"
          value={`${maxTemp.toFixed(1)} ℃`}
          icon={<Thermometer className="h-4 w-4" />}
        />
        <StatCard
          title="最低温"
          value={`${minTemp.toFixed(1)} ℃`}
          icon={<Thermometer className="h-4 w-4" />}
        />
        <StatCard
          title="平均湿度"
          value={`${avgHumidity.toFixed(0)}%`}
          icon={<Droplets className="h-4 w-4" />}
        />
        <StatCard
          title="累计降水"
          value={`${totalPrecip.toFixed(1)} mm`}
          icon={<CloudRain className="h-4 w-4" />}
        />
        <StatCard
          title="平均风速"
          value={`${avgWind.toFixed(1)} km/h`}
          icon={<Wind className="h-4 w-4" />}
        />
      </div>

      {/* ═══════════ NEW: 7-Day Forecast vs Actual ═══════════ */}
      <ChartContainer
        title="7天预报 vs 实际温度对比"
        actions={<DemoBadge tooltip="预报/实况对比数据为演示生成，非真实观测" />}
      >
        <ResponsiveContainer width="100%" height={300}>
          <LineChart
            data={forecastVsActual}
            margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
            <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} width={50} unit="℃" />
            <Tooltip contentStyle={{ fontSize: 12 }} />
            <Legend wrapperStyle={{ fontSize: 12 }} />
            <Line type="monotone" dataKey="forecastHigh" name="预报最高" stroke="#f97316" strokeWidth={2} dot={{ r: 3 }} isAnimationActive={false} />
            <Line type="monotone" dataKey="actualHigh" name="实际最高" stroke="#f97316" strokeWidth={2} strokeDasharray="5 5" dot={{ r: 3 }} isAnimationActive={false} />
            <Line type="monotone" dataKey="forecastLow" name="预报最低" stroke="#3b82f6" strokeWidth={2} dot={{ r: 3 }} isAnimationActive={false} />
            <Line type="monotone" dataKey="actualLow" name="实际最低" stroke="#3b82f6" strokeWidth={2} strokeDasharray="5 5" dot={{ r: 3 }} isAnimationActive={false} />
          </LineChart>
        </ResponsiveContainer>
      </ChartContainer>

      {/* ═══════════ NEW: Weather-Load Scatter ═══════════ */}
      <ChartContainer title="气象与负荷关联散点图（温度 vs 负荷）">
        {scatterData.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无数据</p>
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <ScatterChart margin={{ top: 8, right: 12, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="temp" type="number" name="温度" unit="℃" tick={{ fontSize: 12, fill: '#6b7280' }} label={{ value: '温度 (℃)', position: 'insideBottom', offset: -4, fontSize: 12 }} />
              <YAxis dataKey="load" type="number" name="负荷" unit="MW" tick={{ fontSize: 12, fill: '#6b7280' }} width={60} label={{ value: '负荷 (MW)', angle: -90, position: 'insideLeft', fontSize: 12 }} />
              <ZAxis dataKey="humidity" range={[40, 200]} name="湿度" />
              <Tooltip contentStyle={{ fontSize: 12 }} formatter={(_v: number, name: string) => name === '湿度' ? `${_v.toFixed(0)}%` : `${_v.toFixed(1)}`} cursor={{ strokeDasharray: '3 3' }} />
              <Scatter data={scatterData} fill="#8b5cf6" isAnimationActive={false} />
            </ScatterChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* ═══════════ NEW: Extreme Weather Alerts ═══════════ */}
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
                <div
                  key={i}
                  className={`rounded-md border p-3 ${levelColor[a.level] ?? 'border-gray-300 bg-gray-50'}`}
                >
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

      {/* Temperature / Humidity / Radiation trend chart */}
      <ChartContainer
        title="最近 14 日气象走势"
        actions={
          <div className="flex gap-1">
            <Button
              variant={chartMetric === 'temp' ? 'default' : 'outline'}
              size="sm"
              className="h-7 text-xs"
              onClick={() => setChartMetric('temp')}
            >
              温度
            </Button>
            <Button
              variant={chartMetric === 'humidity' ? 'default' : 'outline'}
              size="sm"
              className="h-7 text-xs"
              onClick={() => setChartMetric('humidity')}
            >
              湿度
            </Button>
            <Button
              variant={chartMetric === 'radiation' ? 'default' : 'outline'}
              size="sm"
              className="h-7 text-xs"
              onClick={() => setChartMetric('radiation')}
            >
              负荷系数
            </Button>
          </div>
        }
      >
        {tempTrend.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            暂无数据{canWrite && '，可点右上「生成演示数据」'}
          </p>
        ) : chartMetric === 'temp' ? (
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart
              data={tempTrend}
              margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                yAxisId="temp"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="℃"
              />
              <YAxis
                yAxisId="precip"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="mm"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) =>
                  name === '降水'
                    ? `${v.toFixed(1)} mm`
                    : `${v.toFixed(1)} ℃`
                }
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Area
                yAxisId="temp"
                type="monotone"
                dataKey="high"
                name="最高"
                stroke="#f97316"
                fill="#fed7aa"
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Area
                yAxisId="temp"
                type="monotone"
                dataKey="low"
                name="最低"
                stroke="#3b82f6"
                fill="#dbeafe"
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Line
                yAxisId="temp"
                type="monotone"
                dataKey="avg"
                name="均温"
                stroke="#8b5cf6"
                strokeWidth={2}
                dot={false}
                strokeDasharray="5 5"
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        ) : chartMetric === 'humidity' ? (
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart
              data={tempTrend}
              margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                yAxisId="humidity"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="%"
              />
              <YAxis
                yAxisId="wind"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="km/h"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) =>
                  name === '风速'
                    ? `${v.toFixed(1)} km/h`
                    : `${v.toFixed(0)}%`
                }
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Area
                yAxisId="humidity"
                type="monotone"
                dataKey="humidity"
                name="湿度"
                stroke="#06b6d4"
                fill="#cffafe"
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Line
                yAxisId="wind"
                type="monotone"
                dataKey="wind"
                name="风速"
                stroke="#f97316"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart
              data={tempTrend}
              margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="date" tick={{ fontSize: 12, fill: '#6b7280' }} />
              <YAxis
                yAxisId="loadFactor"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="%"
              />
              <YAxis
                yAxisId="precip"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="mm"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) =>
                  name === '降水'
                    ? `${v.toFixed(1)} mm`
                    : `${v.toFixed(1)}%`
                }
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Area
                yAxisId="loadFactor"
                type="monotone"
                dataKey="loadFactor"
                name="负荷系数"
                stroke="#8b5cf6"
                fill="#ede9fe"
                strokeWidth={2}
                isAnimationActive={false}
              />
              <Line
                yAxisId="precip"
                type="monotone"
                dataKey="precip"
                name="降水"
                stroke="#3b82f6"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        )}
      </ChartContainer>

      {/* Hourly detail chart (if hourly data available) */}
      {hourlyChartData.length > 0 && (
        <ChartContainer title={`${location} ${today} 逐时气象`}>
          <ResponsiveContainer width="100%" height={280}>
            <ComposedChart
              data={hourlyChartData}
              margin={{ top: 8, right: 12, left: 0, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis dataKey="time" tick={{ fontSize: 11, fill: '#6b7280' }} />
              <YAxis
                yAxisId="temp"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="℃"
              />
              <YAxis
                yAxisId="hum"
                orientation="right"
                tick={{ fontSize: 12, fill: '#6b7280' }}
                width={50}
                unit="%"
              />
              <Tooltip
                contentStyle={{ fontSize: 12 }}
                formatter={(v: number, name: string) =>
                  name === '湿度'
                    ? `${v.toFixed(0)}%`
                    : name === '辐照度'
                      ? `${v.toFixed(1)} W/m²`
                      : `${v.toFixed(1)} ℃`
                }
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Area
                yAxisId="hum"
                type="monotone"
                dataKey="humidity"
                name="湿度"
                stroke="#06b6d4"
                fill="#cffafe"
                strokeWidth={1}
                isAnimationActive={false}
              />
              <Line
                yAxisId="temp"
                type="monotone"
                dataKey="temp"
                name="温度"
                stroke="#f97316"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </ChartContainer>
      )}

      {/* Weather data table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">气象数据明细</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>日期</TableHead>
                  <TableHead>地点</TableHead>
                  <TableHead className="text-right">最高 ℃</TableHead>
                  <TableHead className="text-right">最低 ℃</TableHead>
                  <TableHead className="text-right">湿度 %</TableHead>
                  <TableHead className="text-right">降水 mm</TableHead>
                  <TableHead className="text-right">风速 km/h</TableHead>
                  <TableHead className="text-right">负荷系数</TableHead>
                  <TableHead>天气</TableHead>
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
                {items.map((w) => (
                  <TableRow key={w.id}>
                    <TableCell className="font-medium">
                      {w.obs_date.slice(0, 10)}
                    </TableCell>
                    <TableCell>{w.location}</TableCell>
                    <TableCell className="text-right text-orange-600">
                      {w.temp_high?.toFixed(1) ?? '-'}
                    </TableCell>
                    <TableCell className="text-right text-blue-600">
                      {w.temp_low?.toFixed(1) ?? '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      {w.humidity?.toFixed(0) ?? '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      {w.precip_mm?.toFixed(1) ?? '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      {w.wind_kmh?.toFixed(1) ?? '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      {w.load_factor?.toFixed(2) ?? '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{w.description ?? '-'}</Badge>
                    </TableCell>
                  </TableRow>
                ))}
                {items.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="text-center text-muted-foreground"
                    >
                      暂无气象数据
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* 外部气象观测：风电场风速 + 水库水文（原市场行情，现并入气象数据） */}
      <WeatherObservation />
    </div>
  );
}
