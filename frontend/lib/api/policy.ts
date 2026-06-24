import { apiClient } from './client';

// 政策文件库：把政策文件归纳为结构化条目（来源：文档解析「确认入库 → 政策文件」或手动新增）。

export interface PolicyDocument {
  id: string;
  document_id: string | null;
  document_name: string | null;
  title: string;
  doc_no: string | null;
  category: string | null;
  effective_date: string | null; // YYYY-MM-DD
  summary: string | null;
  source: string; // manual / document
  created_at: string;
  updated_at: string;
}

export interface PolicyInput {
  title: string;
  doc_no?: string;
  category?: string;
  effective_date?: string;
  summary?: string;
  document_id?: string;
}

export async function listPolicies(params: { category?: string; limit?: number } = {}): Promise<PolicyDocument[]> {
  const { data } = await apiClient.get<{ items: PolicyDocument[] }>('/api/v1/policies', { params });
  return data.items ?? [];
}

export async function createPolicy(input: PolicyInput): Promise<{ id: string; message: string }> {
  const { data } = await apiClient.post('/api/v1/policies', input);
  return data;
}

export async function deletePolicy(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/policies/${id}`);
}
