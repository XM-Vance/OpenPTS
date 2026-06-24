import { apiClient } from './client';

// ─── 以下自 v1clone.ts 迁入（2026-06，RPA 监控（任务/运行列表）） ───
// ─── D5 RPA 监控 ───
export interface RPAJob {
  id: string;
  name: string;
  description?: string | null;
  schedule?: string | null;
  enabled: boolean;
  last_run_at?: string | null;
  last_status?: string | null;
  created_at: string;
}
export interface RPARun {
  id: string;
  rpa_job_id: string;
  job_name: string;
  started_at: string;
  finished_at?: string | null;
  status: string;
  duration_sec?: number | null;
  output_files: number;
  output_bytes: number;
  error?: string | null;
}
export async function listRPAJobs(): Promise<{ items: RPAJob[] }> {
  const { data } = await apiClient.get('/api/v1/rpa/jobs');
  return data;
}
export async function listRPARuns(limit = 50): Promise<{ items: RPARun[] }> {
  const { data } = await apiClient.get('/api/v1/rpa/runs', { params: { limit } });
  return data;
}
export async function genRPADemo(): Promise<{ jobs: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/rpa/demo-data');
  return data;
}
