import { apiClient } from './client';

// 类型定义
export interface CurvePoint {
  period: number;
  time_str: string;
  price: number;
  quantity?: number;
}

export interface ContractTypeSummary {
  contract_type: string;
  contract_period: string;
  daily_total_quantity: number;
  daily_avg_price: number;
  max_price: number | null;
  min_price: number | null;
  peak_valley_spread: number | null;
}

export interface DailySummaryKPIs {
  total_quantity: number;
  overall_avg_price: number;
  price_range_min: number;
  price_range_max: number;
  yearly_ratio: number;
  monthly_ratio: number;
  within_month_ratio: number;
  yearly_avg_price: number | null;
  monthly_avg_price: number | null;
  within_month_avg_price: number | null;
}

export interface DailySummaryResponse {
  date: string;
  kpis: DailySummaryKPIs;
  contract_curves: CurvePoint[];
  spot_curves: CurvePoint[];
  type_summary: ContractTypeSummary[];
  curves_by_type: { [key: string]: CurvePoint[] };
  curves_by_period: { [key: string]: CurvePoint[] };
}

// API 方法
export const contractPriceApi = {
  fetchDailySummary: (date: string, entity: string = '全市场', spotType: string = 'day_ahead') => {
    // TODO(P1): 后端无对应路由 GET /retail/price-daily/daily-summary，后端仅有 GET /retail/price-daily
    return apiClient.get<DailySummaryResponse>('/api/v1/retail/price-daily', {
      params: { date, entity, spot_type: spotType },
    });
  },
};

// ─── 以下自 v1clone.ts 迁入（2026-06，合同电价日维度） ───
// ─── D6 合同电价日维度 ───
export interface ContractPriceDaily {
  id: string;
  contract_id: string;
  price_date: string;
  unit_price: number;
  daily_energy: number;
  daily_amount: number;
  cumulative_energy: number;
  cumulative_amount: number;
  created_at: string;
}
export async function listContractPriceDaily(
  params: { contract_id?: string; days?: number } = {},
): Promise<{ items: ContractPriceDaily[] }> {
  const { data } = await apiClient.get('/api/v1/retail/price-daily', { params });
  return data;
}
export async function genContractPriceDemo(): Promise<{ rows: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/retail/price-daily/demo-data');
  return data;
}
