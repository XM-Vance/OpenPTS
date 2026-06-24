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

export async function generatePriceDemoData(
  days = 45,
): Promise<{ days: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/price/demo-data', { days });
  return data;
}

export async function forecastPrice(targetDate: string): Promise<PriceForecastResult> {
  const { data } = await apiClient.post('/api/v1/price/forecast', { target_date: targetDate });
  return data;
}
