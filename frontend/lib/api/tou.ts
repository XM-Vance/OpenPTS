import { apiClient } from './client';

// ─── 以下自 v1clone-e.ts 迁入（2026-06，TOU 时段规则） ───
// ─── E4 TOU ───
export interface TOURule {
  id: string;
  rule_name: string;
  effective_from: string;
  effective_to?: string | null;
  periods: { tags?: string[] };
  created_at: string;
}
export async function listTOURules(): Promise<{ items: TOURule[] }> {
  const { data } = await apiClient.get('/api/v1/price/tou-rules');
  return data;
}
export async function genTOUDemo(): Promise<{ rules: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/price/tou-rules/demo-data');
  return data;
}
