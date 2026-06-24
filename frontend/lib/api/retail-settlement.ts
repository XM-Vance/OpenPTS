import { apiClient } from './client';

export const getRetailSettlementMonthlyCustomers = (params: { month: string }) =>
  // TODO(P1): 后端无对应路由 /retail/monthly-settlement/monthly-customers
  apiClient.get('/api/v1/retail/monthly-settlement', { params });

export const getRetailSettlementMonthlyChartData = (params: { month: string }) =>
  // TODO(P1): 后端无对应路由 /retail/monthly-settlement/monthly-chart-data
  apiClient.get('/api/v1/retail/monthly-settlement', { params });

export const getRetailSettlementMonthlySummaries = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/retail/monthly-settlement', { params });

export const getRetailSettlementMonthlyProgress = (jobId: string) =>
  // TODO(P1): 后端无对应路由 /retail/monthly-settlement/monthly-progress/:jobId
  apiClient.get(`/api/v1/retail/monthly-settlement/${jobId}`);

export const postRetailSettlementCalculate = (data: {
  date?: string;
  month?: string;
  force?: boolean;
  wholesale_version?: string;
}) =>
  // TODO(P1): 后端无对应路由 POST /retail/monthly-settlement/calculate
  apiClient.post('/api/v1/retail/monthly-settlement', data);

export const postRetailSettlementMonthlyCalc = (data: { month: string; force: boolean }) =>
  // TODO(P1): 后端无对应路由 POST /retail/monthly-settlement/monthly-calc
  apiClient.post('/api/v1/retail/monthly-settlement', data);

export const getRetailSettlementDaily = (params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 /retail/monthly-settlement/daily
  apiClient.get('/api/v1/retail/monthly-settlement', { params });

export const calcRetailSettlementMonthly = (data?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 POST /retail/monthly-settlement/monthly-calc
  apiClient.post('/api/v1/retail/monthly-settlement', data);

export const getRetailSettlementMonthlyCustomerDetail = (params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 /retail/monthly-settlement/monthly-customer-detail
  apiClient.get('/api/v1/retail/monthly-settlement', { params });

export const calcRetailSettlementMonthlyForce = (month: string) =>
  // TODO(P1): 后端无对应路由 POST /retail/monthly-settlement/monthly-calc (force)
  apiClient.post('/api/v1/retail/monthly-settlement', { month, force: true });

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
export async function listRetailMonthly(contractId?: string): Promise<{ items: RetailMonthlySettle[] }> {
  const { data } = await apiClient.get('/api/v1/retail/monthly-settlement', { params: { contract_id: contractId } });
  return data;
}
export async function genRetailMonthlyDemo() {
  const { data } = await apiClient.post('/api/v1/retail/monthly-settlement/demo-data');
  return data;
}
