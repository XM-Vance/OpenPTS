'use client';

// 外部气象观测板块：风电场风速（逐时）+ 水库水文（逐日）。
// 数据原属「市场行情」的 md_weather_wind_hourly / md_weather_hydrology_daily，现并入「气象数据」。
import { useState, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Area,
  Bar,
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
  getWindFarm,
  getHydrology,
  genWeatherObsDemo,
  type WeatherStation,
} from '@/lib/api/weather-obs';
import { Wind, Droplets } from 'lucide-react';

const SELECT_CLASS =
  'flex h-9 rounded-md border border-input bg-transparent px-3 text-sm';
const num = (v: number | null | undefined, d = 1) =>
  v == null ? '-' : v.toLocaleString('zh-CN', { maximumFractionDigits: d });
const fmtHour = (iso: string) => iso.slice(5, 16).replace('T', ' '); // MM-DD HH:mm
const fmtDay = (iso: string) => iso.slice(0, 10);

function StationSelect({
  stations,
  value,
  onChange,
}: {
  stations: WeatherStation[];
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <select value={value} onChange={(e) => onChange(e.target.value)} className={SELECT_CLASS}>
      {stations.length === 0 && <option value="">（暂无站点）</option>}
      {stations.map((s) => (
        <option key={s.code} value={s.code}>
          {s.name}
        </option>
      ))}
    </select>
  );
}

/* ── 风电场风速（逐时） ── */
function WindFarmSection() {
  const [station, setStation] = useState('');
  const [hours, setHours] = useState(72);
  const { data, isLoading } = useQuery({
    queryKey: ['weather-wind-farm', station, hours],
    queryFn: () => getWindFarm({ station: station || undefined, hours }).then((r) => r.data),
  });
  const stations = data?.stations ?? [];
  const items = useMemo(() => data?.items ?? [], [data]);
  // 接口按时间倒序；图表用升序。
  const chart = useMemo(
    () =>
      [...items].reverse().map((w) => ({
        t: fmtHour(w.obs_time),
        ws: w.wind_speed_100m,
        temp: w.temperature_2m,
      })),
    [items],
  );

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <Wind className="h-4 w-4 text-sky-600" />
            风电场风速（100m 逐时）
          </CardTitle>
          <div className="flex items-center gap-2">
            <StationSelect
              stations={stations}
              value={station || stations[0]?.code || ''}
              onChange={setStation}
            />
            <select
              value={hours}
              onChange={(e) => setHours(Number(e.target.value))}
              className={SELECT_CLASS}
            >
              <option value={24}>近 24 小时</option>
              <option value={72}>近 3 天</option>
              <option value={168}>近 7 天</option>
            </select>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={260}>
          <LineChart data={chart} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis dataKey="t" tick={{ fontSize: 11 }} minTickGap={40} />
            <YAxis yAxisId="ws" tick={{ fontSize: 11 }} width={48} unit=" m/s" />
            <YAxis yAxisId="temp" orientation="right" tick={{ fontSize: 11 }} width={44} unit="℃" />
            <Tooltip contentStyle={{ fontSize: 12 }} />
            <Legend wrapperStyle={{ fontSize: 12 }} />
            <Line
              yAxisId="ws"
              type="monotone"
              dataKey="ws"
              stroke="#0ea5e9"
              strokeWidth={2}
              dot={false}
              name="100m 风速(m/s)"
              connectNulls
              isAnimationActive={false}
            />
            <Line
              yAxisId="temp"
              type="monotone"
              dataKey="temp"
              stroke="#f97316"
              strokeWidth={1}
              strokeDasharray="4 4"
              dot={false}
              name="气温(℃)"
              connectNulls
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
        <div className="mt-3 max-h-64 overflow-auto rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>时间</TableHead>
                <TableHead>站点</TableHead>
                <TableHead className="text-right">100m 风速(m/s)</TableHead>
                <TableHead className="text-right">风向(°)</TableHead>
                <TableHead className="text-right">气温(℃)</TableHead>
                <TableHead className="text-right">湿度(%)</TableHead>
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
              {items.slice(0, 200).map((w, i) => (
                <TableRow key={i}>
                  <TableCell className="font-medium">{fmtHour(w.obs_time)}</TableCell>
                  <TableCell>{w.location_name}</TableCell>
                  <TableCell className="text-right">{num(w.wind_speed_100m)}</TableCell>
                  <TableCell className="text-right">{num(w.wind_dir_100m, 0)}</TableCell>
                  <TableCell className="text-right">{num(w.temperature_2m)}</TableCell>
                  <TableCell className="text-right">{num(w.humidity_2m, 0)}</TableCell>
                </TableRow>
              ))}
              {items.length === 0 && !isLoading && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    暂无风电场风速数据
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}

/* ── 水库水文（逐日） ── */
function HydrologySection() {
  const [station, setStation] = useState('');
  const [days, setDays] = useState(30);
  const { data, isLoading } = useQuery({
    queryKey: ['weather-hydrology', station, days],
    queryFn: () => getHydrology({ station: station || undefined, days }).then((r) => r.data),
  });
  const stations = data?.stations ?? [];
  const items = useMemo(() => data?.items ?? [], [data]);
  const chart = useMemo(
    () =>
      [...items].reverse().map((h) => ({
        d: fmtDay(h.obs_date),
        precip: h.precipitation_sum,
        temp: h.temp_mean,
      })),
    [items],
  );

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <Droplets className="h-4 w-4 text-blue-600" />
            水库水文（逐日）
          </CardTitle>
          <div className="flex items-center gap-2">
            <StationSelect
              stations={stations}
              value={station || stations[0]?.code || ''}
              onChange={setStation}
            />
            <select
              value={days}
              onChange={(e) => setDays(Number(e.target.value))}
              className={SELECT_CLASS}
            >
              <option value={7}>近 7 天</option>
              <option value={30}>近 30 天</option>
              <option value={90}>近 90 天</option>
            </select>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={260}>
          <ComposedChart data={chart} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis dataKey="d" tick={{ fontSize: 11 }} minTickGap={32} />
            <YAxis yAxisId="p" tick={{ fontSize: 11 }} width={48} unit="mm" />
            <YAxis yAxisId="t" orientation="right" tick={{ fontSize: 11 }} width={44} unit="℃" />
            <Tooltip contentStyle={{ fontSize: 12 }} />
            <Legend wrapperStyle={{ fontSize: 12 }} />
            <Bar yAxisId="p" dataKey="precip" fill="#3b82f6" fillOpacity={0.5} name="降水(mm)" isAnimationActive={false} />
            <Area
              yAxisId="t"
              type="monotone"
              dataKey="temp"
              stroke="#f97316"
              fill="#f97316"
              fillOpacity={0.1}
              strokeWidth={2}
              name="平均气温(℃)"
              connectNulls
              isAnimationActive={false}
            />
          </ComposedChart>
        </ResponsiveContainer>
        <div className="mt-3 max-h-64 overflow-auto rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>日期</TableHead>
                <TableHead>站点</TableHead>
                <TableHead className="text-right">均温(℃)</TableHead>
                <TableHead className="text-right">湿度(%)</TableHead>
                <TableHead className="text-right">降水(mm)</TableHead>
                <TableHead className="text-right">蒸发 ET₀(mm)</TableHead>
                <TableHead className="text-right">10m 风速(m/s)</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground">
                    加载中...
                  </TableCell>
                </TableRow>
              )}
              {items.slice(0, 200).map((h, i) => (
                <TableRow key={i}>
                  <TableCell className="font-medium">{fmtDay(h.obs_date)}</TableCell>
                  <TableCell>{h.location_name}</TableCell>
                  <TableCell className="text-right">{num(h.temp_mean)}</TableCell>
                  <TableCell className="text-right">{num(h.humidity_mean, 0)}</TableCell>
                  <TableCell className="text-right">{num(h.precipitation_sum)}</TableCell>
                  <TableCell className="text-right">{num(h.et0_evapotranspiration)}</TableCell>
                  <TableCell className="text-right">{num(h.wind_speed_10m_mean)}</TableCell>
                </TableRow>
              ))}
              {items.length === 0 && !isLoading && (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground">
                    暂无水库水文数据
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}

/* ── 外部气象观测：风电场风速 + 水库水文 ── */
export function WeatherObservation() {
  const qc = useQueryClient();
  const { has } = usePermission();
  const canWrite = has('load_management:write');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onGen = async () => {
    setBusy(true);
    setError(null);
    try {
      await genWeatherObsDemo();
      qc.invalidateQueries({ queryKey: ['weather-wind-farm'] });
      qc.invalidateQueries({ queryKey: ['weather-hydrology'] });
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
          <h2 className="text-lg font-semibold">外部气象观测</h2>
          <p className="text-sm text-muted-foreground">
            风电场逐时风速、水库水文逐日观测（原「市场行情」气象水文，现并入气象数据）
          </p>
        </div>
        {canWrite && (
          <Button variant="outline" onClick={onGen} disabled={busy}>
            {busy ? '生成中...' : '生成演示数据'}
          </Button>
        )}
      </div>
      {error && <p className="text-sm text-destructive">{error}</p>}
      <div className="grid gap-4 xl:grid-cols-2">
        <WindFarmSection />
        <HydrologySection />
      </div>
    </div>
  );
}
