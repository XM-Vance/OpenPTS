import { apiClient } from './client';

export interface ApprovalRequest {
  id: string;
  resource: string;
  resource_id: string;
  title: string;
  payload: Record<string, unknown>;
  status: 'draft' | 'pending' | 'approved' | 'rejected' | 'withdrawn';
  submitted_by: string;
  reviewed_by?: string | null;
  review_note?: string | null;
  reviewed_at?: string | null;
  created_at: string;
  updated_at: string;
}

export async function listApprovals(
  params: { status?: string; resource?: string; mine?: boolean; limit?: number; offset?: number } = {},
): Promise<{ items: ApprovalRequest[]; total: number }> {
  const { data } = await apiClient.get('/api/v1/approvals', {
    params: { ...params, mine: params.mine ? 'true' : undefined },
  });
  return data;
}

export async function getApproval(id: string): Promise<ApprovalRequest> {
  const { data } = await apiClient.get(`/api/v1/approvals/${id}`);
  return data;
}

export interface ApprovalTemplate {
  id: string;
  name: string;
  resource: string;
  title_tpl: string;
  field: string;
  description?: string | null;
}

export async function listApprovalTemplates(resource?: string): Promise<{ items: ApprovalTemplate[] }> {
  const { data } = await apiClient.get('/api/v1/approvals/templates', {
    params: resource ? { resource } : {},
  });
  return data;
}

export async function approvalsByResource(
  resource: string,
  resourceId: string,
): Promise<{ items: ApprovalRequest[] }> {
  const { data } = await apiClient.get('/api/v1/approvals/by-resource', {
    params: { resource, resource_id: resourceId },
  });
  return data;
}

export async function submitApproval(input: {
  resource: string;
  resource_id: string;
  title: string;
  payload?: Record<string, unknown>;
}): Promise<ApprovalRequest> {
  const { data } = await apiClient.post('/api/v1/approvals', input);
  return data;
}

export async function approveRequest(id: string, note?: string): Promise<ApprovalRequest> {
  const { data } = await apiClient.post(`/api/v1/approvals/${id}/approve`, { note });
  return data;
}

export async function rejectRequest(id: string, note?: string): Promise<ApprovalRequest> {
  const { data } = await apiClient.post(`/api/v1/approvals/${id}/reject`, { note });
  return data;
}

export async function withdrawRequest(id: string): Promise<ApprovalRequest> {
  const { data } = await apiClient.post(`/api/v1/approvals/${id}/withdraw`);
  return data;
}
