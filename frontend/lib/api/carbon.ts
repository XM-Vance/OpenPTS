import { apiClient } from './client';
import type { SchemaCarbonQuote, SchemaCarbonProductSummary } from './types.gen';

// 碳交易行情（CEA 全国碳配额 / CCER 国家核证自愿减排 / EUA 欧盟碳配额）。
// 碳价为全国统一行情，不分省。
// 实体类型由 OpenAPI 规范生成（与 customers.ts 同模式）。
export type CarbonQuote = SchemaCarbonQuote;
export type CarbonProductSummary = SchemaCarbonProductSummary;

export const getCarbonSummary = () =>
  apiClient.get<{ items: CarbonProductSummary[] }>('/api/v1/carbon/summary');

export const getCarbonQuotes = (params?: { product?: string; days?: number }) =>
  apiClient.get<{ items: CarbonQuote[] }>('/api/v1/carbon/quotes', { params });

export const genCarbonDemo = () =>
  apiClient.post<{ rows: number; message: string }>('/api/v1/carbon/demo-data');
