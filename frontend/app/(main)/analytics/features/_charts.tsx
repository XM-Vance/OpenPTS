'use client';

// 客户特征分析图表（雷达 + 价值矩阵散点 + 行业分布饼）。
// 单独成文件以便 next/dynamic 懒加载，剥离 recharts 出首屏。
import {
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  PolarAngleAxis,
  PolarGrid,
  PolarRadiusAxis,
  Radar,
  RadarChart,
  ResponsiveContainer,
  Scatter,
  ScatterChart,
  Tooltip,
  XAxis,
  YAxis,
  ZAxis,
} from 'recharts';

export function FeatureRadar({ data, customers, colors }: { data: any[]; customers: string[]; colors: string[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <RadarChart cx="50%" cy="50%" outerRadius="70%" data={data}>
        <PolarGrid />
        <PolarAngleAxis dataKey="dimension" tick={{ fontSize: 11 }} />
        <PolarRadiusAxis angle={30} domain={[0, 100]} tick={{ fontSize: 10 }} />
        {customers.map((name, idx) => (
          <Radar
            key={name}
            name={name}
            dataKey={name}
            stroke={colors[idx % colors.length]}
            fill={colors[idx % colors.length]}
            fillOpacity={0.15}
            isAnimationActive={false}
          />
        ))}
        <Legend wrapperStyle={{ fontSize: 10 }} />
        <Tooltip />
      </RadarChart>
    </ResponsiveContainer>
  );
}

export function ValueScatter({ data }: { data: any[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <ScatterChart margin={{ top: 8, right: 16, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis
          type="number"
          dataKey="收益"
          name="收益评分"
          tick={{ fontSize: 11 }}
        />
        <YAxis
          type="number"
          dataKey="风险"
          name="风险评分"
          tick={{ fontSize: 11 }}
        />
        <ZAxis type="number" dataKey="用电量" range={[40, 200]} />
        <Tooltip
          cursor={{ strokeDasharray: '3 3' }}
          formatter={(value: number, name: string) => [
            `${value}`,
            name,
          ]}
        />
        <Scatter data={data} fill="#6366f1" isAnimationActive={false} />
      </ScatterChart>
    </ResponsiveContainer>
  );
}

export function IndustryPie({ data, colors }: { data: any[]; colors: string[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          outerRadius={90}
          dataKey="value"
          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
          isAnimationActive={false}
        >
          {data.map((_, idx) => (
            <Cell key={idx} fill={colors[idx % colors.length]} />
          ))}
        </Pie>
        <Tooltip />
      </PieChart>
    </ResponsiveContainer>
  );
}
