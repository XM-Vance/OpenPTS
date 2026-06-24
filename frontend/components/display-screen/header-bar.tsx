import { C } from './constants';
import { NumberFlip } from './number-flip';

interface HeaderBarProps {
  timeStr: string;
  dateStr: string;
  /** 关键指标概览 */
  overview?: {
    today_energy_mwh: number;
    today_revenue: number;
    deviation_rate: number;
    alert_count: number;
  };
}

export function HeaderBar({ timeStr, dateStr, overview }: HeaderBarProps) {
  return (
    <div
      style={{
        height: 70,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        borderBottom: `2px solid ${C.blue}40`,
      }}
    >
      {/* 左侧装饰线 */}
      <div
        style={{
          position: 'absolute',
          left: 40,
          top: '50%',
          width: 280,
          height: 1,
          background: `linear-gradient(90deg, transparent, ${C.blue})`,
          transform: 'translateY(-50%)',
        }}
      />
      <div
        style={{
          position: 'absolute',
          left: 320,
          top: '50%',
          width: 5,
          height: 5,
          borderRadius: '50%',
          background: C.blue,
          transform: 'translateY(-50%)',
        }}
      />

      {/* 左侧翻牌器指标 */}
      {overview && (
        <div
          style={{
            position: 'absolute',
            left: 50,
            top: '50%',
            transform: 'translateY(-50%)',
            display: 'flex',
            gap: 32,
          }}
        >
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: C.textDim, marginBottom: 2 }}>
              今日用电量
            </div>
            <NumberFlip
              value={overview.today_energy_mwh}
              decimals={1}
              fontSize={16}
              color={C.blue}
              suffix=" MWh"
            />
          </div>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: C.textDim, marginBottom: 2 }}>
              今日营收
            </div>
            <NumberFlip
              value={overview.today_revenue}
              decimals={0}
              fontSize={16}
              color={C.gold}
              prefix="¥"
              useGrouping
            />
          </div>
        </div>
      )}

      {/* 标题 */}
      <div
        style={{
          fontSize: 30,
          fontWeight: 700,
          color: C.text,
          letterSpacing: 6,
          textShadow: `0 0 20px ${C.blue}40`,
        }}
      >
        电力交易运营大屏
      </div>

      {/* 右侧翻牌器指标 */}
      {overview && (
        <div
          style={{
            position: 'absolute',
            right: 200,
            top: '50%',
            transform: 'translateY(-50%)',
            display: 'flex',
            gap: 32,
          }}
        >
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: C.textDim, marginBottom: 2 }}>
              偏差率
            </div>
            <NumberFlip
              value={overview.deviation_rate}
              decimals={2}
              fontSize={16}
              color={overview.deviation_rate > 5 ? C.alertRed : C.green}
              suffix="%"
            />
          </div>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: C.textDim, marginBottom: 2 }}>
              告警数
            </div>
            <NumberFlip
              value={overview.alert_count}
              decimals={0}
              fontSize={16}
              color={overview.alert_count > 0 ? C.alertRed : C.green}
            />
          </div>
        </div>
      )}

      {/* 右侧装饰线 */}
      <div
        style={{
          position: 'absolute',
          right: 320,
          top: '50%',
          width: 5,
          height: 5,
          borderRadius: '50%',
          background: C.blue,
          transform: 'translateY(-50%)',
        }}
      />
      <div
        style={{
          position: 'absolute',
          right: 40,
          top: '50%',
          width: 280,
          height: 1,
          background: `linear-gradient(270deg, transparent, ${C.blue})`,
          transform: 'translateY(-50%)',
        }}
      />
      {/* 时间 */}
      <div
        style={{
          position: 'absolute',
          right: 24,
          top: 18,
          fontSize: 12,
          color: C.cyanDim,
          fontFamily: '"Roboto Mono", monospace',
          letterSpacing: 1,
          textAlign: 'right',
        }}
      >
        <div>{dateStr}</div>
        <div style={{ fontSize: 16, color: C.cyan, fontWeight: 600 }}>
          {timeStr}
        </div>
      </div>
    </div>
  );
}
