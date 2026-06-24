'use client';

import React, { useMemo } from 'react';

/**
 * Time-of-use price heatmap
 * - X-axis: 24 hours, Y-axis: last 7 dates
 * - Color intensity = price level
 * - Pure div grid, no extra library
 */

interface HeatmapCell {
  date: string;
  hour: number;
  value: number | null;
}

export interface PriceHeatmapProps {
  /** Nested map: date string → hour(0-23) → price value */
  data: HeatmapCell[];
  /** Min price for color scale */
  minVal?: number;
  /** Max price for color scale */
  maxVal?: number;
  className?: string;
}

function lerpColor(ratio: number): string {
  // blue → green → yellow → red gradient
  const clamped = Math.max(0, Math.min(1, ratio));
  if (clamped < 0.33) {
    const t = clamped / 0.33;
    const r = Math.round(30 + t * 40);
    const g = Math.round(100 + t * 130);
    const b = Math.round(200 - t * 130);
    return `rgb(${r},${g},${b})`;
  }
  if (clamped < 0.66) {
    const t = (clamped - 0.33) / 0.33;
    const r = Math.round(70 + t * 150);
    const g = Math.round(230 - t * 30);
    const b = Math.round(70 - t * 50);
    return `rgb(${r},${g},${b})`;
  }
  const t = (clamped - 0.66) / 0.34;
  const r = Math.round(220 + t * 35);
  const g = Math.round(200 - t * 140);
  const b = Math.round(20 - t * 20);
  return `rgb(${r},${g},${b})`;
}

export function PriceHeatmap({ data, minVal, maxVal, className }: PriceHeatmapProps) {
  // Group data by date
  const dates = useMemo(() => {
    const set = new Set(data.map((d) => d.date));
    return Array.from(set).slice(-7); // Last 7 days
  }, [data]);

  const lo = minVal ?? Math.min(...data.map((d) => d.value ?? Infinity));
  const hi = maxVal ?? Math.max(...data.map((d) => d.value ?? -Infinity));
  const range = hi - lo || 1;

  // Build lookup map
  const cellMap = useMemo(() => {
    const m = new Map<string, number | null>();
    data.forEach((c) => m.set(`${c.date}_${c.hour}`, c.value));
    return m;
  }, [data]);

  if (dates.length === 0) return null;

  return (
    <div className={className}>
      {/* Header: hour labels */}
      <div className="flex">
        <div className="w-14 shrink-0" /> {/* date label space */}
        <div className="grid flex-1" style={{ gridTemplateColumns: 'repeat(24, 1fr)' }}>
          {[0, 6, 12, 18].map((h) => (
            <div
              key={h}
              className="text-center text-[10px] text-zinc-400"
              style={{ gridColumn: h + 1 }}
            >
              {String(h).padStart(2, '0')}
            </div>
          ))}
        </div>
      </div>
      {/* Rows */}
      {dates.map((date) => (
        <div key={date} className="flex items-center">
          <div className="w-14 shrink-0 truncate pr-1 text-right text-[10px] text-zinc-400">
            {date.slice(5)}
          </div>
          <div
            className="grid flex-1 gap-px"
            style={{ gridTemplateColumns: 'repeat(24, 1fr)' }}
          >
            {Array.from({ length: 24 }, (_, hour) => {
              const val = cellMap.get(`${date}_${hour}`);
              const ratio = val != null ? (val - lo) / range : -1;
              const bg = val != null ? lerpColor(ratio) : 'rgba(50,50,50,0.3)';
              return (
                <div
                  key={hour}
                  className="aspect-square rounded-[2px] transition-colors"
                  style={{ backgroundColor: bg }}
                  title={`${date} ${String(hour).padStart(2, '0')}:00 — ${val != null ? val.toFixed(2) : 'N/A'} ¥/MWh`}
                />
              );
            })}
          </div>
        </div>
      ))}
      {/* Legend */}
      <div className="mt-2 flex items-center justify-end gap-1">
        <span className="text-[10px] text-zinc-400">低</span>
        <div className="flex h-2.5 w-32 rounded">
          {Array.from({ length: 16 }, (_, i) => (
            <div
              key={i}
              className="flex-1"
              style={{ backgroundColor: lerpColor(i / 15) }}
            />
          ))}
        </div>
        <span className="text-[10px] text-zinc-400">高</span>
      </div>
    </div>
  );
}
