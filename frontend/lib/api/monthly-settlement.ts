// 月度结算 API。
// 2026-06 自 v1clone.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── D1 月度结算 ───
export interface MonthlySettlement {
  id: string;
  operating_month: string;
  settled_energy_mwh: number;
  energy_fee: number;
  capacity_fee: number;
  ancillary_fee: number;
  policy_subsidy: number;
  total_fee: number;
  version: string;
  created_at: string;
}

export async function listMonthlySettlement(limit = 12): Promise<{ items: MonthlySettlement[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/monthly', { params: { limit } });
  return data;
}
export async function genMonthlySettlementDemo(): Promise<{ months: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/settlement/monthly/demo-data');
  return data;
}
