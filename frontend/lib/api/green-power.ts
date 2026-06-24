import { apiClient } from './client';

// ─── 类型 ───

export interface GreenPowerTrade {
  id: string;
  trade_date: string;
  product_name: string;
  energy_mwh: number;
  price: number;
  amount: number;
  green_cert_count: number;
  status: string;
  counterparty: string;
  created_at: string;
}

// ─── API ───

export async function listGreenPowerTrades(
  params: { status?: string; days?: number } = {},
): Promise<{ items: GreenPowerTrade[] }> {
  const { data } = await apiClient.get('/api/v1/trade/green-power', { params });
  return data;
}

export async function genGreenPowerDemo(): Promise<{
  trades: number;
  message: string;
}> {
  const { data } = await apiClient.post('/api/v1/trade/green-power/demo-data');
  return data;
}
