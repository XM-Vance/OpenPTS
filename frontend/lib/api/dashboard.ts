import { apiClient } from './client';

export interface DashboardSummary {
  customer_count: number;
  active_contracts: number;
  active_packages: number;
  pending_alerts: number;
  critical_alerts: number;
  active_stations: number;
  storage_30d_revenue: number;
  freq_7d_revenue: number;
  latest_settlement_fee?: number | null;
}

export interface DailySeriesPoint {
  date: string;
  value: number;
}

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
