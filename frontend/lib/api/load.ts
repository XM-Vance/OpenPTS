import { apiClient } from './client';

export interface LoadForecast {
  forecast: number[];
  lower: number[];
  upper: number[];
  total: number;
  peak: number;
  valley: number;
  method: string;
  sample_days: number;
  target_weekday: number;
}

export interface ForecastResult {
  target_date: string;
  history_days: number;
  forecast: LoadForecast;
}

// 为客户生成演示负荷数据（合成 96 点曲线）。
export async function generateDemoData(
  customerId: string,
  days = 45,
): Promise<{ days: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/load/demo-data', {
    customer_id: customerId,
    days,
  });
  return data;
}

// 短期负荷预测。
export async function forecastLoad(
  customerId: string,
  targetDate: string,
): Promise<ForecastResult> {
  const { data } = await apiClient.post('/api/v1/load/forecast', {
    customer_id: customerId,
    target_date: targetDate,
  });
  return data;
}
