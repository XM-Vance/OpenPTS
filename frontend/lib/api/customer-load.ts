// 客户负荷分析 API。
// 2026-06 自 v1clone-e.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── E2 客户负荷分析 ───
export interface CustomerLoadSummary {
  customer_id: string;
  customer_name: string;
  days: number;
  avg_daily: number;
  peak_load: number;
  valley_load: number;
  peak_valley_ratio: number;
  cv: number;
}
export interface CustomerLoadCurve {
  customer_id: string;
  customer_name: string;
  date: string;
  curve_96: number[];
}
export async function getCustomerLoadSummary(days = 14): Promise<{ items: CustomerLoadSummary[] }> {
  const { data } = await apiClient.get('/api/v1/analytics/customer-load/summary', {
    params: { days },
  });
  return data;
}
export async function getCustomerLoadCurve(id: string): Promise<CustomerLoadCurve> {
  const { data } = await apiClient.get(`/api/v1/analytics/customer-load/${id}/curve`);
  return data;
}
