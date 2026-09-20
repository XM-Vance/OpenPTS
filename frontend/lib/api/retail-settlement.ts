import { apiClient } from './client';

// 零售月度结算 API。
// 注：原先有 9 个带 TODO(P1) 的函数（monthly-customers/chart-data/progress/calculate/
// monthly-calc/daily/monthly-customer-detail 等），后端从未实现且无任何页面调用（死代码），
// 已删除（契约闸 KNOWN_PENDING 对应条目同步移除）。仅保留被页面实际使用的函数。

export const getRetailSettlementMonthlySummaries = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/retail/monthly-settlement', { params });

// ─── 以下自 p0p1.ts 迁入（2026-06，零售月度结算） ───
// ─── U1 零售月结 ───
export interface RetailMonthlySettle {
  id: string;
  contract_id: string;
  customer_name: string;
  operating_month: string;
  contract_energy_mwh: number;
  actual_energy_mwh: number;
  weighted_avg_price: number;
  receivable_amount: number;
  actual_amount: number;
  deviation_energy_mwh: number;
  penalty_amount: number;
  note?: string | null;
  created_at: string;
}
