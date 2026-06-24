// 意向客户 API。
// 2026-06 自 v1clone-e.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── E1 意向客户 ───
export interface IntentCustomer {
  id: string;
  customer_name: string;
  meters: unknown;
  coverage_start?: string | null;
  coverage_end?: string | null;
  coverage_days?: number | null;
  completeness?: number | null;
  avg_daily_load?: number | null;
  status: string;
  extra: unknown;
  created_at: string;
}
export interface IntentDiagnosis extends IntentCustomer {
  data_score: number;
  load_score: number;
  coverage_score: number;
  overall_score: number;
  matched_package?: string | null;
  recommendation: string;
}
export async function listIntentCustomers(): Promise<{ items: IntentCustomer[] }> {
  const { data } = await apiClient.get('/api/v1/intent-customers');
  return data;
}
export async function diagnoseIntent(): Promise<{ items: IntentDiagnosis[] }> {
  const { data } = await apiClient.get('/api/v1/intent-customers/diagnose');
  return data;
}
export async function genIntentDemo(): Promise<{ customers: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/intent-customers/demo-data');
  return data;
}
