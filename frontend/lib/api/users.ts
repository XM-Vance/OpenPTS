import { apiClient } from './client';

export interface User {
  id: string;
  username: string;
  display_name?: string | null;
  email?: string | null;
  phone?: string | null;
  is_active: boolean;
  last_login_at?: string | null;
  created_at: string;
  updated_at: string;
  roles?: string[];
}

export interface UserListParams {
  keyword?: string;
  is_active?: boolean;
  limit?: number;
  offset?: number;
}

export interface CreateUserParams {
  username: string;
  password: string;
  display_name?: string;
  email?: string;
  phone?: string;
  roles?: string[];
}

export interface UpdateUserParams {
  display_name?: string;
  email?: string;
  phone?: string;
  is_active?: boolean;
}

export async function listUsers(
  params: UserListParams = {},
): Promise<{ items: User[]; total: number }> {
  const { data } = await apiClient.get('/api/v1/users', { params });
  return data;
}

export async function getUser(id: string): Promise<User> {
  const { data } = await apiClient.get(`/api/v1/users/${id}`);
  return data;
}

export async function createUser(params: CreateUserParams): Promise<User> {
  const { data } = await apiClient.post('/api/v1/users', params);
  return data;
}

export async function updateUser(id: string, params: UpdateUserParams): Promise<User> {
  const { data } = await apiClient.put(`/api/v1/users/${id}`, params);
  return data;
}

export async function resetUserPassword(id: string, newPassword: string): Promise<void> {
  await apiClient.post(`/api/v1/users/${id}/password`, { new_password: newPassword });
}

export async function setUserRoles(id: string, roles: string[]): Promise<void> {
  await apiClient.put(`/api/v1/users/${id}/roles`, { roles });
}
