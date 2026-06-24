import { C } from './constants';

export function KpiCard({
  title,
  value,
  unit,
  color = C.cyan,
  subtitle,
}: {
  title: string;
  value: string;
  unit?: string;
  color?: string;
  subtitle?: string;
}) {
  return (
    <div
      style={{
        background: C.cardBg,
        border: `1px solid ${C.cardBorder}`,
        borderRadius: 8,
        padding: '14px 18px',
        display: 'flex',
        flexDirection: 'column',
        gap: 4,
        backdropFilter: 'blur(8px)',
      }}
    >
      <div style={{ fontSize: 12, color: C.textDim, fontWeight: 500, letterSpacing: 0.5 }}>
        {title}
      </div>
      <div
        style={{
          fontSize: 24,
          fontWeight: 700,
          color,
          fontFamily: '"DIN Alternate", "Roboto Mono", monospace',
          textShadow: `0 0 12px ${color}40`,
        }}
      >
        {value}
      </div>
      {(unit || subtitle) && (
        <div style={{ fontSize: 11, color: C.textDim }}>
          {subtitle || unit}
        </div>
      )}
    </div>
  );
}
