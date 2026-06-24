import { apiClient } from './client';

/* ── 大屏概览数据 ── */
export interface DisplayOverview {
  total_customers: number;
  active_contracts: number;
  today_energy_mwh: number;
  today_revenue: number;
  month_energy_mwh: number;
  month_revenue: number;
  avg_price: number;
  deviation_rate: number;
  alert_count: number;
  green_ratio: number;
}

/* ── 大屏趋势数据 ── */
export interface DisplayTrendItem {
  date: string;
  energy_mwh: number;
  revenue: number;
  avg_price: number;
}

export async function getDisplayOverview(): Promise<DisplayOverview> {
  const { data } = await apiClient.get('/api/v1/display/overview');
  return data;
}

export async function getDisplayTrend(
  days = 14,
): Promise<{ items: DisplayTrendItem[] }> {
  const { data } = await apiClient.get('/api/v1/display/trend', {
    params: { days },
  });
  return data;
}

export async function postDisplayDemoData(): Promise<{
  rows: number;
  message: string;
}> {
  const { data } = await apiClient.post('/api/v1/display/demo-data');
  return data;
}
