import { apiClient } from './client';

// ─── 合同电价日维度（自 v1clone.ts 迁入，2026-06） ───
// 注：原 contractPriceApi 聚合对象（fetchDailySummary）及相关类型（DailySummaryResponse
// 等）无任何 UI 调用，是死代码，已删除。页面用下面的 listContractPriceDaily。
export interface ContractPriceDaily {
  id: string;
  contract_id: string;
  price_date: string;
  unit_price: number;
  daily_energy: number;
  daily_amount: number;
  cumulative_energy: number;
  cumulative_amount: number;
  created_at: string;
}
export async function listContractPriceDaily(
  params: { contract_id?: string; days?: number } = {},
): Promise<{ items: ContractPriceDaily[] }> {
  const { data } = await apiClient.get('/api/v1/retail/price-daily', { params });
  return data;
}
export async function genContractPriceDemo(): Promise<{ rows: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/retail/price-daily/demo-data');
  return data;
}
