import { apiClient } from './client';
import type { SchemaScheduledJob } from './types.gen';

// ScheduledJob 类型由 OpenAPI 规范生成（与 audit.ts / customers.ts 同模式）。
export type ScheduledJob = SchemaScheduledJob;

export interface JobRun {
  id: string;
  job_id: string;
  job_name: string;
  started_at: string;
  finished_at?: string | null;
  status: string;
  error?: string | null;
  duration_ms?: number | null;
  trigger: string;
}

export async function listJobs(): Promise<{ items: ScheduledJob[] }> {
  const { data } = await apiClient.get('/api/v1/scheduler/jobs');
  return data;
}

export async function listRuns(limit = 50): Promise<{ items: JobRun[] }> {
  const { data } = await apiClient.get('/api/v1/scheduler/runs', {
    params: { limit },
  });
  return data;
}

export async function triggerJob(id: string): Promise<void> {
  await apiClient.post(`/api/v1/scheduler/jobs/${id}/trigger`);
}

export async function setJobEnabled(id: string, enabled: boolean): Promise<void> {
  await apiClient.put(`/api/v1/scheduler/jobs/${id}/enabled`, { enabled });
}
