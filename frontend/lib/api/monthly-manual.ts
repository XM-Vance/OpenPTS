// 月度手工数据 API。
// 2026-06 自 v1clone-f.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── F4 ───
export interface ManualItem {
  id: string;
  operating_month: string;
  category: string;
  item_name: string;
  value: number;
  unit: string;
  source?: string | null;
  note?: string | null;
  created_by?: string | null;
  created_at: string;
  updated_at: string;
}
export async function listManualItems(
  params: { month?: string; category?: string } = {},
): Promise<{ items: ManualItem[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/manual-data', { params });
  return data;
}
export async function createManualItem(input: {
  operating_month: string;
  category: string;
  item_name: string;
  value: number;
  unit?: string;
  source?: string;
  note?: string;
}): Promise<{ id: string }> {
  const { data } = await apiClient.post('/api/v1/settlement/manual-data', input);
  return data;
}
export async function genManualDemo(): Promise<{ rows: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/settlement/manual-data/demo-data');
  return data;
}
