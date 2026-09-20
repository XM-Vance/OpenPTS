import { apiClient } from './client';

export interface PriceForecast {
  forecast: number[];
  lower: number[];
  upper: number[];
  total: number;
  peak: number;
  valley: number;
  method: string;
  sample_days: number;
}

export interface PriceForecastResult {
  target_date: string;
  history_days: number;
  forecast: PriceForecast;
}

// LEAR 预测结果（含 forecast_id 与外生变量天数）
export interface LearForecastResult {
  target_date: string;
  history_days: number;
  exogenous_days: number;
  forecast_id: string;
  forecast: PriceForecast;
}

// 价格预测历史记录（含准确率，若已回填）
export interface PriceForecastRecord {
  forecast_id: string;
  forecast_date: string;
  target_date: string;
  price_da_forecast: number[];
  price_rt_forecast?: number[];
  forecast_method: string;
  accuracy_metrics?: {
    wmape: number;
    mae: number;
    rmse: number;
  };
  operator?: string;
  created_at: string;
}

export async function forecastPrice(targetDate: string): Promise<PriceForecastResult> {
  const { data } = await apiClient.post('/api/v1/price/forecast', { target_date: targetDate });
  return data;
}

/**
 * LEAR 日前价格预测（LASSO 自回归 + 外生变量）。
 * 算法原理：按 Lago et al. (Applied Energy 2021, 293:116983) 自实现。
 * 默认取系统负荷作外生变量；无数据时自动退化为纯自回归。
 */
export async function forecastPriceLear(
  targetDate: string,
  useExogenous = true,
): Promise<LearForecastResult> {
  const { data } = await apiClient.post<LearForecastResult>(
    '/api/v1/price/forecast/lear',
    { target_date: targetDate, use_exogenous: useExogenous },
  );
  return data;
}

/**
 * 查询某目标日的所有预测版本（含准确率，目标日已过会懒回填）。
 */
export async function listPriceForecastResults(
  targetDate: string,
): Promise<{ target_date: string; items: PriceForecastRecord[] }> {
  const { data } = await apiClient.get('/api/v1/price/forecast/results', {
    params: { target_date: targetDate },
  });
  return data;
}
