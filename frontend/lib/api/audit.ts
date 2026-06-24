import { apiClient } from './client';
import type { SchemaAuditLog } from './types.gen';

// AuditLog 类型由 OpenAPI 规范生成(P1-6 类型契约接通的首个域)。
// 后端改字段 → `npm run gen-api-types` 重生成 → 此处与页面同步报错,杜绝运行时才发现的漂移。
export type AuditLog = SchemaAuditLog;

export interface AuditQuery {
  username?: string;
  method?: string;
  resource?: string;
  days?: number;
  limit?: number;
  offset?: number;
}

export async function listAuditLogs(
  q: AuditQuery = {},
): Promise<{ items: AuditLog[]; total: number }> {
  const { data } = await apiClient.get('/api/v1/audit/logs', { params: q });
  return data;
}
