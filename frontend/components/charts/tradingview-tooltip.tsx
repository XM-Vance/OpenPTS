'use client';

import React from 'react';

/**
 * TradingView-style crosshair tooltip
 * - Vertical line via CSS (handled by Recharts cursor)
 * - Floating data panel with price details
 */
export interface TradingViewTooltipProps {
  active?: boolean;
  payload?: Record<string, any>[];
  label?: string;
  unit?: string;
  /** Extra fields to display */
  extraFields?: Record<string, string>;
}

const LINE_COLORS: Record<string, string> = {
  日前价格: '#3b82f6',
  实时价格: '#f59e0b',
  预测价格: '#10b981',
  MA5: '#a855f7',
  MA20: '#ec4899',
};

export function TradingViewTooltip({
  active,
  payload,
  label,
  unit = '¥/MWh',
  extraFields,
}: TradingViewTooltipProps) {
  if (!active || !payload || payload.length === 0) return null;

  return (
    <div
      className="pointer-events-none rounded border border-zinc-700 bg-zinc-900/95 px-3 py-2 shadow-xl"
      style={{ zIndex: 100 }}
    >
      <p className="mb-1.5 border-b border-zinc-700 pb-1 text-xs font-semibold text-zinc-300">
        {label}
      </p>
      {payload.map((pld, idx) => {
        const dataKey = pld.dataKey as string;
        const val = pld.value;
        const isValidNum = val !== null && val !== undefined && !isNaN(Number(val));
        const displayValue = isValidNum ? Number(val).toFixed(2) : 'N/A';
        const color = LINE_COLORS[dataKey] || ((pld.stroke || pld.fill || '#ccc') as string);

        return (
          <div key={idx} className="flex items-center justify-between gap-4 py-0.5 text-xs">
            <span className="flex items-center gap-1.5">
              <span
                className="inline-block h-2 w-2 shrink-0 rounded-full"
                style={{ backgroundColor: color }}
              />
              <span className="text-zinc-400">{pld.name || dataKey}</span>
            </span>
            <span className="font-mono font-semibold" style={{ color }}>
              {displayValue} {isValidNum ? unit : ''}
            </span>
          </div>
        );
      })}
      {extraFields &&
        Object.entries(extraFields).map(([k, v]) => (
          <div key={k} className="flex items-center justify-between gap-4 py-0.5 text-xs">
            <span className="text-zinc-400">{k}</span>
            <span className="font-mono text-zinc-300">{v}</span>
          </div>
        ))}
    </div>
  );
}
