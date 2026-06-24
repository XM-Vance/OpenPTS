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
