import { apiClient } from './client';

// 客户历史电量档案：客户逐月电量（来源：文档解析「确认入库」→ 客户电量档案）。

export interface CustomerMonthlyEnergy {
  id: string;
  customer_id: string;
  customer_name: string;
  month: string; // YYYY-MM
  monthly_energy: number;
  avg_daily_energy: number | null;
  variation_cv: number | null;
  created_at: string;
  updated_at: string;
}

export async function listCustomerEnergy(params: {
  customer_id?: string;
  limit?: number;
} = {}): Promise<CustomerMonthlyEnergy[]> {
  const { data } = await apiClient.get<{ items: CustomerMonthlyEnergy[] }>(
    '/api/v1/customer-energy',
    { params },
  );
  return data.items ?? [];
}
