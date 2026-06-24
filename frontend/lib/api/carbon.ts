import { apiClient } from './client';

// 碳交易行情（CEA 全国碳配额 / CCER 国家核证自愿减排 / EUA 欧盟碳配额）。
// 碳价为全国统一行情，不分省。

export interface CarbonQuote {
  id: number;
  product: string;
  trade_date: string;
  open_price: number | null;
  high_price: number | null;
  low_price: number | null;
  close_price: number | null;
  volume: number | null;
  turnover: number | null;
  created_at: string;
}

export interface CarbonProductSummary {
  product: string;
  name: string;
  unit: string;
  latest_date: string | null;
  close: number | null;
  prev_close: number | null;
  change: number | null;
  change_pct: number | null;
  volume: number | null;
  high_52w: number | null;
  low_52w: number | null;
}

export const getCarbonSummary = () =>
  apiClient.get<{ items: CarbonProductSummary[] }>('/api/v1/carbon/summary');

export const getCarbonQuotes = (params?: { product?: string; days?: number }) =>
  apiClient.get<{ items: CarbonQuote[] }>('/api/v1/carbon/quotes', { params });

export const genCarbonDemo = () =>
  apiClient.post<{ rows: number; message: string }>('/api/v1/carbon/demo-data');
