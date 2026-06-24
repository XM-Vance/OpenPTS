import { C } from './constants';

export function ChartTooltip({
  active,
  payload,
  label,
  formatter,
}: {
  active?: boolean;
  payload?: Array<{ value: number; name?: string; color?: string }>;
  label?: string;
  formatter?: (v: number) => string;
}) {
  if (!active || !payload || payload.length === 0) return null;
  return (
    <div
      style={{
        background: 'rgba(10, 22, 40, 0.95)',
        border: `1px solid ${C.blue}60`,
        borderRadius: 6,
        padding: '8px 12px',
        fontSize: 12,
      }}
    >
      <div style={{ color: C.textDim, marginBottom: 4 }}>{label}</div>
      {payload.map((p, i) => (
        <div key={i} style={{ color: p.color || C.text, fontWeight: 600 }}>
          {p.name ? `${p.name}: ` : ''}
          {formatter ? formatter(p.value) : p.value.toLocaleString()}
        </div>
      ))}
    </div>
  );
}
