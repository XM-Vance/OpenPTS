import { apiClient } from './client';
import type { SchemaDashboardSummary, SchemaDailySeriesPoint } from './types.gen';

// 实体类型由 OpenAPI 规范生成（与 customers.ts 同模式）。
export type DashboardSummary = SchemaDashboardSummary;
export type DailySeriesPoint = SchemaDailySeriesPoint;

export async function getDashboardSummary(): Promise<DashboardSummary> {
  const { data } = await apiClient.get('/api/v1/dashboard/summary');
  return data;
}

export async function getSettlementSeries(
  days = 14,
): Promise<{ items: DailySeriesPoint[] }> {
  const { data } = await apiClient.get('/api/v1/dashboard/series/settlement', {
    params: { days },
  });
  return data;
}

export async function getFreqSeries(
  days = 14,
): Promise<{ items: DailySeriesPoint[] }> {
  const { data } = await apiClient.get('/api/v1/dashboard/series/freq', {
    params: { days },
  });
  return data;
}

/* ── Settlement KPI ── */
export interface SettlementKpi {
  yearly_gross_profit: number;
  monthly_gross_profit: number;
  price_spread: number;
  retail_avg_price: number;
}

export interface SettlementChartPoint {
  label: string;
  monthly_gross_profit?: number;
  yearly_gross_profit?: number;
  total_purchase?: number;
  total_retail?: number;
  total_wholesale?: number;
}

export interface SettlementSummaryResponse {
  kpi: SettlementKpi | null;
  monthly_chart: SettlementChartPoint[];
  yearly_chart: SettlementChartPoint[];
  customer_overview: {
    total: number;
    by_type: Record<string, number>;
    by_status: Record<string, number>;
  } | null;
  alerts: Array<{ id: string; level: string; message: string; created_at: string }>;
}

export async function getSettlementSummary(): Promise<SettlementSummaryResponse> {
  const { data } = await apiClient.get('/api/v1/dashboard/settlement-summary');
  return data;
}
