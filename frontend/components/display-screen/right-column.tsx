import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Tooltip,
} from 'recharts';
import type { DisplayOverview } from '@/lib/api/display-screen';
import { C } from './constants';
import { ChartTooltip } from './chart-tooltip';

interface TradeStatItem {
  name: string;
  value: number;
  color: string;
}

interface RightColumnProps {
  overview: DisplayOverview;
  tradeStats: TradeStatItem[];
}

export function RightColumn({ overview, tradeStats }: RightColumnProps) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      {/* 交易结构饼图 */}
      <div
        style={{
          background: C.cardBg,
          border: `1px solid ${C.cardBorder}`,
          borderRadius: 8,
          padding: '10px 14px',
          backdropFilter: 'blur(8px)',
          flex: 0.6,
        }}
      >
        <div
          style={{
            fontSize: 12,
            color: C.blue,
            fontWeight: 600,
            letterSpacing: 1,
            marginBottom: 2,
          }}
        >
          交易结构
        </div>
        <ResponsiveContainer width="100%" height="85%">
          <PieChart>
            <Pie
              data={tradeStats}
              cx="50%"
              cy="50%"
              innerRadius={38}
              outerRadius={60}
              paddingAngle={3}
              dataKey="value"
              stroke="none"
            >
              {tradeStats.map((entry, i) => (
                <Cell key={i} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip
              content={<ChartTooltip formatter={(v) => `${v.toFixed(1)}%`} />}
            />
          </PieChart>
        </ResponsiveContainer>
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: '6px 12px',
            justifyContent: 'center',
            marginTop: 2,
          }}
        >
          {tradeStats.map((s) => (
            <div
              key={s.name}
              style={{ display: 'flex', alignItems: 'center', gap: 4 }}
            >
              <div
                style={{
                  width: 8,
                  height: 8,
                  borderRadius: '50%',
                  background: s.color,
                }}
              />
              <span style={{ fontSize: 10, color: C.textDim }}>
                {s.name} {s.value}%
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* 告警 & 绿电 */}
      <div
        style={{
          background: C.cardBg,
          border: `1px solid ${C.cardBorder}`,
          borderRadius: 8,
          padding: '12px 14px',
          backdropFilter: 'blur(8px)',
          flex: 0.4,
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 8,
          }}
        >
          <span
            style={{
              fontSize: 12,
              color: C.blue,
              fontWeight: 600,
              letterSpacing: 1,
            }}
          >
            运行状态
          </span>
        </div>
        <div
          style={{
            display: 'flex',
            gap: 12,
            alignItems: 'center',
            justifyContent: 'space-around',
          }}
        >
          <StatusItem
            value={String(overview.alert_count)}
            label="待处理告警"
            color={overview.alert_count > 0 ? C.pink : C.green}
          />
          <div style={{ width: 1, height: 30, background: C.divider }} />
          <StatusItem
            value={`${overview.green_ratio.toFixed(1)}%`}
            label="绿电占比"
            color={C.green}
          />
          <div style={{ width: 1, height: 30, background: C.divider }} />
          <StatusItem
            value={overview.avg_price.toFixed(0)}
            label="均价"
            color={C.cyan}
          />
        </div>
      </div>

      {/* 快速导航 */}
      <div
        onClick={() => window.open('/dashboard', '_self')}
        style={{
          background: `linear-gradient(135deg, ${C.blue}20, ${C.blue}08)`,
          border: `1px solid ${C.blue}40`,
          borderRadius: 8,
          padding: '10px 14px',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          transition: 'all 0.2s',
          flex: 0.3,
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.borderColor = C.blue;
          e.currentTarget.style.boxShadow = `0 0 20px ${C.blue}30`;
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.borderColor = `${C.blue}40`;
          e.currentTarget.style.boxShadow = 'none';
        }}
      >
        <span style={{ fontSize: 13, color: C.text, fontWeight: 600 }}>
          进入系统主页
        </span>
        <span style={{ fontSize: 16, color: C.cyan }}>→</span>
      </div>
    </div>
  );
}

function StatusItem({ value, label, color }: { value: string; label: string; color: string }) {
  return (
    <div style={{ textAlign: 'center' }}>
      <div
        style={{
          fontSize: 22,
          fontWeight: 700,
          color,
          fontFamily: '"DIN Alternate", monospace',
        }}
      >
        {value}
      </div>
      <div style={{ fontSize: 10, color: C.textDim }}>{label}</div>
    </div>
  );
}
