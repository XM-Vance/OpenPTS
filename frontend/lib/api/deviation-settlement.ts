import { apiClient } from './client';

// ─── 数据类型 ───

export interface DeviationSettlement {
  id: string;
  operating_date: string;
  declared_energy_mwh: number;
  actual_energy_mwh: number;
  deviation_energy_mwh: number;
  deviation_rate: number;
  deviation_cost: number;
  penalty_cost: number;
  total_settlement: number;
  category: string;
  created_at: string;
}

export interface DeviationSummaryItem {
  category: string;
  total_deviation_energy_mwh: number;
  total_cost: number;
  avg_deviation_rate: number;
  count: number;
}

// ─── API 调用 ───

export async function listDeviations(
  days = 30,
  category = '',
): Promise<{ items: DeviationSettlement[] }> {
  const params: Record<string, string | number> = { days };
  if (category) params.category = category;
  const { data } = await apiClient.get('/api/v1/settlement/deviation', { params });
  return data;
}

export async function getDeviationSummary(
  days = 30,
): Promise<{ items: DeviationSummaryItem[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/deviation/summary', {
    params: { days },
  });
  return data;
}

export async function generateDeviationDemoData(): Promise<{
  rows: number;
  message: string;
}> {
  const { data } = await apiClient.post('/api/v1/settlement/deviation/demo-data');
  return data;
}
