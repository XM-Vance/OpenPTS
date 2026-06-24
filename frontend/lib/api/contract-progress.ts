import { apiClient } from './client';

// ─── 数据类型 ───

export interface ContractProgress {
  id: string;
  contract_id: string;
  customer_name?: string;
  operating_month: string;
  planned_energy_mwh: number;
  actual_energy_mwh: number;
  completion_rate: number;
  status: string;
  milestones?: unknown;
  note?: string;
  created_at: string;
  updated_at: string;
}

// ─── API 调用 ───

export async function listContractProgress(
  month = '',
  status = '',
  limit = 100,
): Promise<{ items: ContractProgress[] }> {
  const params: Record<string, string | number> = { limit };
  if (month) params.month = month;
  if (status) params.status = status;
  const { data } = await apiClient.get('/api/v1/retail/contract-progress', { params });
  return data;
}

export async function generateContractProgressDemoData(): Promise<{
  rows: number;
  message: string;
}> {
  const { data } = await apiClient.post('/api/v1/retail/contract-progress/demo-data');
  return data;
}
