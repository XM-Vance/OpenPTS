// 预结算 API。
// 2026-06 自 p0p1.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── U2 预结算 ───
export interface PreSettleDaily {
  id: string;
  operating_date: string;
  declared_curve_96: number[];
  cleared_curve_96: number[];
  spot_price_96: number[];
  total_declared: number;
  total_cleared: number;
  total_deviation: number;
  deviation_ratio: number;
  energy_revenue: number;
  deviation_penalty: number;
  final_amount: number;
}
export async function listPreSettle(days = 14): Promise<{ items: PreSettleDaily[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/pre', { params: { days } });
  return data;
}
export async function getPreSettle(date: string): Promise<PreSettleDaily> {
  const { data } = await apiClient.get(`/api/v1/settlement/pre/${date}`);
  return data;
}
export async function genPreSettleDemo() {
  const { data } = await apiClient.post('/api/v1/settlement/pre/demo-data');
  return data;
}
