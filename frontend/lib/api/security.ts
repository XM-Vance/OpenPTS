import { apiClient } from './client';

export interface SecurityOverview {
  window_hours: number;
  total: number;
  errors_4xx: number;
  errors_5xx: number;
  delete_ops: number;
  unique_users: number;
  unique_ips: number;
  top_failed_users: { user: string; count: number }[];
  top_active_ips: { ip: string; count: number }[];
  recent_deletes: {
    username?: string | null;
    path: string;
    resource?: string | null;
    status_code: number;
    created_at: string;
  }[];
  failed_sched_jobs: number;
  error_hourly: { bucket: string; count: number }[];
}

export async function getSecurityOverview(hours = 24): Promise<SecurityOverview> {
  const { data } = await apiClient.get('/api/v1/system/security/overview', {
    params: { hours },
  });
  return data;
}
