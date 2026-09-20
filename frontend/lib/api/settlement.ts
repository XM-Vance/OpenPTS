import { apiClient } from './client';
import type {
  SchemaSettlementDaily,
  SchemaPeriodDetail,
  SchemaSettlementDetail,
} from './types.gen';

// 实体类型由 OpenAPI 规范生成（与 customers.ts 同模式）。
// 后端改字段 → `npm run gen-api-types` 重生成 → 此处与页面同步报错。
export type SettlementDaily = SchemaSettlementDaily;
export type PeriodDetail = SchemaPeriodDetail;
export type SettlementDetail = SchemaSettlementDetail;

export async function listSettlements(limit = 30): Promise<{ items: SettlementDaily[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/daily', { params: { limit } });
  return data;
}

export async function getSettlement(
  date: string,
  version = 'PRELIMINARY',
): Promise<SettlementDetail> {
  const { data } = await apiClient.get(`/api/v1/settlement/daily/${date}`, {
    params: { version },
  });
  return data;
}

export async function generateSettlementDemoData(
  days = 30,
): Promise<{ days: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/settlement/demo-data', { days });
  return data;
}
