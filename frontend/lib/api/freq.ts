import { apiClient } from './client';

export interface FreqDailySummary {
  date: string;
  agc_volume?: number | null;
  agc_price?: number | null;
  agc_revenue?: number | null;
  avc_volume?: number | null;
  avc_price?: number | null;
  avc_revenue?: number | null;
  demand_volume?: number | null;
  demand_price?: number | null;
  comp_fee?: number | null;
}

export async function listFreqSummary(limit = 30): Promise<{ items: FreqDailySummary[] }> {
  const { data } = await apiClient.get('/api/v1/freq/clearing', { params: { limit } });
  return data;
}

export async function generateFreqDemoData(
  days = 30,
): Promise<{ days: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/freq/demo-data', { days });
  return data;
}

// ─── WP5.2：考核指标（雷达图数据源） ───

export interface FreqPerformanceItem {
  id: string;
  settlement_date: string;
  regulation_type: 'AGC' | 'AVC';
  response_score: number;
  precision_score: number;
  duration_score: number;
  delay_score: number;
  capacity_score: number;
  mileage_mw: number;
  is_demo: boolean;
}

export const listFreqPerformance = (limit = 30): Promise<{ items: FreqPerformanceItem[] }> =>
  apiClient.get('/api/v1/freq/performance', { params: { limit } }).then((r) => r.data);
