import { apiClient } from './client';

// ── 类型 ──

export interface MarketDataTableInfo {
  table_name: string;
  label: string;
  category: string;
  scope: string; // "national" | "provincial"
  row_count: number;
  date_range: string;
}

export interface MarketDataOverview {
  total_tables: number;
  categories: Record<
    string,
    {
      count: number;
      tables: number;
      table_list: string[];
    }
  >;
}

export interface MarketDataQueryResult {
  table: string;
  count: number;
  data: Record<string, unknown>[];
  scope: string;
  scope_label: string;
}

// ── API ──

export const getMarketDataOverview = () =>
  apiClient.get<MarketDataOverview>('/api/v1/market-data/overview');

export const getMarketDataTables = () =>
  apiClient.get<{ tables: MarketDataTableInfo[] }>('/api/v1/market-data/tables');

export const queryMarketData = (
  table: string,
  params?: { days?: number; location_code?: string },
) =>
  apiClient.get<MarketDataQueryResult>(`/api/v1/market-data/${table}`, { params });
