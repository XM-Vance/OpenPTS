import { apiClient } from './client';

export interface Agent {
  id: string;
  agent_name: string;
  contact_person: string;
  phone: string;
  email: string;
  region: string;
  commission_rate: number;
  status: string;
  description: string;
  created_by?: string | null;
  created_at: string;
  updated_at: string;
}

export interface AgentListParams {
  keyword?: string;
  status?: string;
  page?: number;
  page_size?: number;
  limit?: number;
  offset?: number;
}

export interface AgentInput {
  agent_name: string;
  contact_person?: string;
  phone?: string;
  email?: string;
  region?: string;
  commission_rate?: number;
  status?: string;
  description?: string;
}

export async function listAgents(
  params: AgentListParams = {},
): Promise<{ items: Agent[]; total: number }> {
  const { data } = await apiClient.get('/api/v1/agents', { params });
  return data;
}

export async function getAgent(id: string): Promise<Agent> {
  const { data } = await apiClient.get(`/api/v1/agents/${id}`);
  return data;
}

export async function createAgent(input: AgentInput): Promise<Agent> {
  const { data } = await apiClient.post('/api/v1/agents', input);
  return data;
}

export async function updateAgent(id: string, input: AgentInput): Promise<Agent> {
  const { data } = await apiClient.put(`/api/v1/agents/${id}`, input);
  return data;
}

export async function deleteAgent(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/agents/${id}`);
}

export async function listAgentCustomers(id: string): Promise<{ items: unknown[] }> {
  const { data } = await apiClient.get(`/api/v1/agents/${id}/customers`);
  return data;
}
