'use client';

// 绿电交易图表（绿证/碳价折线 + 绿电占比 Composed）。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
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

export interface GreenCertPricePoint {
  month: string;
  greenCert: number;
  carbonPrice: number;
}

export interface GreenRatioPoint {
  month: string;
  greenRatio: number;
  totalEnergy: number;
  greenEnergy: number;
}

export function GreenCertPriceLine({ data }: { data: GreenCertPricePoint[] }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis dataKey="month" tick={{ fontSize: 11 }} />
        <YAxis tick={{ fontSize: 11 }} width={56} unit="元" />
        <Tooltip
          contentStyle={{ fontSize: 12 }}
          formatter={(v: number, name: string) => [`${v} 元`, name]}
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Line
          type="monotone"
          dataKey="greenCert"
          stroke="#10b981"
          strokeWidth={2}
          dot={{ r: 3 }}
          name="绿证价格"
          isAnimationActive={false}
        />
        <Line
          type="monotone"
          dataKey="carbonPrice"
          stroke="#3b82f6"
          strokeWidth={2}
          dot={{ r: 3 }}
          name="碳配额价格"
          isAnimationActive={false}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}

export function GreenRatioComposed({ data }: { data: GreenRatioPoint[] }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <ComposedChart data={data} margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis dataKey="month" tick={{ fontSize: 11 }} />
        <YAxis
          yAxisId="ratio"
          tick={{ fontSize: 11 }}
          width={56}
          unit="%"
          domain={[0, 100]}
        />
        <YAxis
          yAxisId="energy"
          orientation="right"
          tick={{ fontSize: 11 }}
          width={56}
        />
        <Tooltip
          contentStyle={{ fontSize: 12 }}
          formatter={(v: number, name: string) =>
            name === '绿电占比' ? `${v}%` : `${v} MWh`
          }
        />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        <Area
          yAxisId="ratio"
          type="monotone"
          dataKey="greenRatio"
          stroke="#10b981"
          fill="#10b981"
          fillOpacity={0.2}
          strokeWidth={2}
          name="绿电占比"
          isAnimationActive={false}
        />
        <Line
          yAxisId="energy"
          type="monotone"
          dataKey="totalEnergy"
          stroke="#94a3b8"
          strokeWidth={1}
          strokeDasharray="4 4"
          dot={false}
          name="总用电量"
          isAnimationActive={false}
        />
        <Line
          yAxisId="energy"
          type="monotone"
          dataKey="greenEnergy"
          stroke="#059669"
          strokeWidth={1}
          dot={false}
          name="绿电量"
          isAnimationActive={false}
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
