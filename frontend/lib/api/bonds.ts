import { apiClient } from './client';

export interface Bond {
  id: string;
  name: string;
  bond_type: string;
  amount: number;
  issuer: string;
  beneficiary: string;
  issue_date?: string | null;
  expire_date?: string | null;
  status: string;
  description: string;
  created_by?: string | null;
  created_at: string;
  updated_at: string;
}

export interface BondListParams {
  keyword?: string;
  status?: string;
  page?: number;
  page_size?: number;
  limit?: number;
  offset?: number;
}

export interface BondInput {
  name: string;
  bond_type?: string;
  amount?: number;
  issuer?: string;
  beneficiary?: string;
  issue_date?: string;
  expire_date?: string;
  status?: string;
  description?: string;
}

export async function listBonds(
  params: BondListParams = {},
): Promise<{ items: Bond[]; total: number }> {
  const { data } = await apiClient.get('/api/v1/bonds', { params });
  return data;
}

export async function getBond(id: string): Promise<Bond> {
  const { data } = await apiClient.get(`/api/v1/bonds/${id}`);
  return data;
}

export async function createBond(input: BondInput): Promise<Bond> {
  const { data } = await apiClient.post('/api/v1/bonds', input);
  return data;
}

export async function updateBond(id: string, input: BondInput): Promise<Bond> {
  const { data } = await apiClient.put(`/api/v1/bonds/${id}`, input);
  return data;
}

export async function deleteBond(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/bonds/${id}`);
}
