import { apiClient } from './client';

export interface AlertStats {
  total: number;
  pending: number;
  acknowledged: number;
  critical: number;
}

export interface CustomerAlert {
  id: string;
  customer_id: string;
  customer_name: string;
  alert_date: string;
  alert_type: string;
  severity: string;
  confidence?: number | null;
  reason?: string | null;
  rule_id?: string | null;
  acknowledged: boolean;
  acknowledged_by?: string | null;
  acknowledged_at?: string | null;
  note?: string | null;
  created_at: string;
}

export interface CustomerCharacteristic {
  id: string;
  customer_id: string;
  customer_name: string;
  data_date: string;
  long_term: unknown;
  short_term: unknown;
  tags: string[];
  regularity_score?: number | null;
  quality_rating?: string | null;
}

export async function getAlertStats(): Promise<AlertStats> {
  const { data } = await apiClient.get('/api/v1/analytics/alerts/stats');
  return data;
}

export async function listAlerts(
  params: { limit?: number; include_acked?: boolean } = {},
): Promise<{ items: CustomerAlert[] }> {
  const { data } = await apiClient.get('/api/v1/analytics/alerts', { params });
  return data;
}

export async function ackAlert(id: string): Promise<void> {
  await apiClient.post(`/api/v1/analytics/alerts/${id}/ack`);
}

export async function listCharacteristics(
  limit = 20,
): Promise<{ items: CustomerCharacteristic[] }> {
  const { data } = await apiClient.get('/api/v1/analytics/characteristics', {
    params: { limit },
  });
  return data;
}

export async function generateAnalyticsDemoData(): Promise<{
  customers: number;
  characteristics: number;
  alerts: number;
  message: string;
}> {
  const { data } = await apiClient.post('/api/v1/analytics/demo-data');
  return data;
}
