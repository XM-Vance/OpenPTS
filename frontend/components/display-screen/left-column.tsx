import type { DisplayOverview } from '@/lib/api/display-screen';
import { C } from './constants';
import { fmtNum, fmtMoney, fmtEnergy } from './helpers';
import { KpiCard } from './kpi-card';

export function LeftColumn({ overview }: { overview: DisplayOverview }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      {/* 实时指标 */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: 8,
        }}
      >
        <KpiCard
          title="今日用电量"
          value={overview.today_energy_mwh.toFixed(1)}
          unit="MWh"
          color={C.cyan}
        />
        <KpiCard
          title="今日营收"
          value={fmtMoney(overview.today_revenue)}
          unit="元"
          color={C.gold}
        />
        <KpiCard
          title="客户总数"
          value={fmtNum(overview.total_customers)}
          unit="户"
          color={C.green}
        />
        <KpiCard
          title="活跃合同"
          value={fmtNum(overview.active_contracts)}
          unit="份"
          color={C.purple}
        />
      </div>

      {/* 月度聚合 */}
      <div
        style={{
          background: C.cardBg,
          border: `1px solid ${C.cardBorder}`,
          borderRadius: 8,
          padding: '12px 14px',
          backdropFilter: 'blur(8px)',
          flex: 1,
        }}
      >
        <div
          style={{
            fontSize: 12,
            color: C.blue,
            fontWeight: 600,
            letterSpacing: 1,
            marginBottom: 8,
          }}
        >
          月度汇总
        </div>
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 8,
          }}
        >
          <SummaryRow label="月累计用电" value={fmtEnergy(overview.month_energy_mwh)} color={C.cyan} />
          <SummaryRow label="月累计营收" value={`${fmtMoney(overview.month_revenue)} 元`} color={C.gold} />
          <SummaryRow label="平均电价" value={`${overview.avg_price.toFixed(2)} 元/MWh`} color={C.orange} />
          <SummaryRow
            label="偏差率"
            value={`${overview.deviation_rate.toFixed(2)}%`}
            color={overview.deviation_rate > 5 ? C.pink : C.green}
          />
          <SummaryRow
            label="绿电占比"
            value={`${overview.green_ratio.toFixed(1)}%`}
            color={C.green}
            noBorder
          />
        </div>
      </div>
    </div>
  );
}

function SummaryRow({
  label,
  value,
  color,
  noBorder = false,
}: {
  label: string;
  value: string;
  color: string;
  noBorder?: boolean;
}) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '6px 0',
        borderBottom: noBorder ? undefined : `1px solid ${C.divider}`,
      }}
    >
      <span style={{ fontSize: 12, color: C.textDim }}>{label}</span>
      <span
        style={{
          fontSize: 15,
          color,
          fontWeight: 600,
          fontFamily: '"DIN Alternate", monospace',
        }}
      >
        {value}
      </span>
    </div>
  );
}
