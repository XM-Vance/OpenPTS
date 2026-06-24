import { apiClient } from './client';

export interface Role {
  code: string;
  name: string;
  description?: string | null;
  is_system: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  permissions?: string[];
}

export interface Permission {
  code: string;
  name: string;
  module_code: string;
  action: string;
  permission_type: string;
  is_active: boolean;
}

export interface Module {
  code: string;
  name: string;
  menu_group?: string | null;
  route_paths: string[];
  sort_order: number;
  is_active: boolean;
}

export async function listRoles(): Promise<Role[]> {
  const { data } = await apiClient.get('/api/v1/roles');
  return data.items;
}

export async function getRole(code: string): Promise<Role> {
  const { data } = await apiClient.get(`/api/v1/roles/${code}`);
  return data;
}

export async function createRole(params: {
  code: string;
  name: string;
  description?: string;
  permissions?: string[];
}): Promise<Role> {
  const { data } = await apiClient.post('/api/v1/roles', params);
  return data;
}

export async function updateRole(
  code: string,
  params: { name: string; description?: string; is_active: boolean },
): Promise<Role> {
  const { data } = await apiClient.put(`/api/v1/roles/${code}`, params);
  return data;
}

export async function deleteRole(code: string): Promise<void> {
  await apiClient.delete(`/api/v1/roles/${code}`);
}

export async function setRolePermissions(code: string, permissions: string[]): Promise<void> {
  await apiClient.put(`/api/v1/roles/${code}/permissions`, { permissions });
}

export async function listPermissions(): Promise<Permission[]> {
  const { data } = await apiClient.get('/api/v1/permissions');
  return data.items;
}

export async function listModules(): Promise<Module[]> {
  const { data } = await apiClient.get('/api/v1/modules');
  return data.items;
}
