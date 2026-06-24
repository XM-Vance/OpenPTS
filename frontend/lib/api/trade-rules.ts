import { apiClient, extractErrorMessage } from './client';

export interface TradeRule {
  id: string;
  category: string;
  rule_key: string;
  rule_value: string;
  effective_date: string;
  expiry_date: string | null;
  description?: string | null;
  created_at: string;
  updated_at: string;
}

export interface TradeRuleInput {
  category: string;
  rule_key: string;
  rule_value: string;
  effective_date: string;
  expiry_date?: string | null;
  description?: string;
}

export async function listTradeRules(category?: string): Promise<{ items: TradeRule[] }> {
  try {
    const res = await apiClient.get<{ items: TradeRule[] }>('/trade-rules', {
      params: category ? { category } : undefined,
    });
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to fetch trade rules'));
  }
}

export async function createTradeRule(data: TradeRuleInput): Promise<TradeRule> {
  try {
    const res = await apiClient.post<TradeRule>('/trade-rules', data);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to create trade rule'));
  }
}

export async function updateTradeRule(id: string, data: Partial<TradeRuleInput>): Promise<TradeRule> {
  try {
    const res = await apiClient.put<TradeRule>(`/trade-rules/${id}`, data);
    return res.data;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to update trade rule'));
  }
}

export async function deleteTradeRule(id: string): Promise<void> {
  try {
    await apiClient.delete(`/trade-rules/${id}`);
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to delete trade rule'));
  }
}

export async function exportTradeRules(): Promise<Blob> {
  try {
    const res = await apiClient.get('/trade-rules/export', { responseType: 'blob' });
    return res.data as Blob;
  } catch (e) {
    throw new Error(extractErrorMessage(e, 'Failed to export trade rules'));
  }
}
