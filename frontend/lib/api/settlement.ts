import { apiClient } from './client';

export interface SettlementDaily {
  id: string;
  operating_date: string;
  version: string;
  contract_fee?: number | null;
  day_ahead_fee?: number | null;
  real_time_fee?: number | null;
  total_energy_fee?: number | null;
  energy_avg_price?: number | null;
  deviation_recovery_fee?: number | null;
  created_at: string;
}

export interface PeriodDetail {
  period: number;
  volume_mwh: number;
  price: number;
  fee: number;
}

export interface SettlementDetail extends SettlementDaily {
  period_details: PeriodDetail[];
}

export async function listSettlements(limit = 30): Promise<{ items: SettlementDaily[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/daily', { params: { limit } });
  return data;
}

export async function getSettlement(
  date: string,
  version = 'PRELIMINARY',
): Promise<SettlementDetail> {
  const { data } = await apiClient.get(`/api/v1/settlement/daily/${date}`, {
    params: { version },
  });
  return data;
}

export async function generateSettlementDemoData(
  days = 30,
): Promise<{ days: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/settlement/demo-data', { days });
  return data;
}
