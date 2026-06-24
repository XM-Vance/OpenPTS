// 电网代理价 API。
// 2026-06 自 v1clone-e.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── E5 电网代理价 ───
export interface GridAgencyPrice {
  id: string;
  operating_month: string;
  voltage_level: string;
  avg_price: number;
  peak_price: number;
  flat_price: number;
  valley_price: number;
  created_at: string;
}
export async function listGridAgency(
  params: { voltage?: string; months?: number } = {},
): Promise<{ items: GridAgencyPrice[] }> {
  const { data } = await apiClient.get('/api/v1/price/grid-agency', { params });
  return data;
}
export async function genGridAgencyDemo(): Promise<{ rows: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/price/grid-agency/demo-data');
  return data;
}
