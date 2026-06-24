import { apiClient } from './client';
import type { SchemaCustomer } from './types.gen';

// Customer 类型由 OpenAPI 规范生成（与 audit.ts 同模式）。
// 后端改字段 → `npm run gen-api-types` 重生成 → 此处与页面同步报错，杜绝运行时才发现的漂移。
export type Customer = SchemaCustomer;

export interface CustomerListParams {
  keyword?: string;
  tag?: string;
  manager?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
  limit?: number;
  offset?: number;
}

export interface CustomerInput {
  user_name: string;
  short_name?: string;
  location?: string;
  source?: string;
  manager?: string;
  tags?: string[];
  is_demo?: boolean;
  extra?: Record<string, unknown>;
}

export async function listCustomers(
  params: CustomerListParams = {},
): Promise<{ items: Customer[]; total: number }> {
  const { data } = await apiClient.get('/api/v1/customers', { params });
  return data;
}

/** 跨省搜索客户（文档预填匹配用，不受 X-Org-Id 限制） */
export async function searchCustomersAllOrg(
  keyword: string,
  limit = 5,
): Promise<Customer[]> {
  const { data } = await apiClient.get('/api/v1/customers', {
    params: { keyword, limit },
    // 覆盖请求拦截器设的 X-Org-Id，HQ 模式下后端会返回全部省份
    headers: { 'X-Org-Id': '' },
  });
  return data.items ?? [];
}

export async function getCustomer(id: string): Promise<Customer> {
  const { data } = await apiClient.get(`/api/v1/customers/${id}`);
  return data;
}

export async function createCustomer(input: CustomerInput): Promise<Customer> {
  const { data } = await apiClient.post('/api/v1/customers', input);
  return data;
}

export async function updateCustomer(id: string, input: CustomerInput): Promise<Customer> {
  const { data } = await apiClient.put(`/api/v1/customers/${id}`, input);
  return data;
}

export async function deleteCustomer(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/customers/${id}`);
}
