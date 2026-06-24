// 系统总负荷 API。
// 2026-06 自 p0p1.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── V1 系统总负荷 ───
export interface TotalLoadDaily {
  id: string;
  data_date: string;
  region: string;
  curve_96: number[];
  peak_load: number;
  valley_load: number;
  avg_load: number;
  total_mwh: number;
}
export async function listTotalLoad(days = 14): Promise<{ items: TotalLoadDaily[] }> {
  const { data } = await apiClient.get('/api/v1/load/total', { params: { days } });
  return data;
}
export async function genTotalLoadDemo() {
  const { data } = await apiClient.post('/api/v1/load/total/demo-data');
  return data;
}
