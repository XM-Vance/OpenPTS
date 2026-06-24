import { apiClient } from './client';

export interface AuthPermission {
  code: string;
  name: string;
  module?: string;
  module_code?: string | null;
  action?: string;
  permission_type?: string;
  is_exception?: boolean;
  is_system?: boolean;
  is_active?: boolean;
  description?: string;
}

export interface AuthRole {
  code: string;
  name: string;
  description?: string;
  permissions: string[];
  is_system?: boolean;
  is_active?: boolean;
}

export interface AuthModule {
  menu_group: string;
  module_name: string;
  module_code: string;
  route_paths: string[];
  sort_order: number;
  is_active?: boolean;
}

export interface AuthUser {
  _id?: string;
  username: string;
  display_name?: string;
  email?: string;
  email_verified?: boolean;
  email_mfa_enabled?: boolean;
  roles: string[];
  is_active?: boolean;
  must_change_password?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface AuthAuditLog {
  event: string;
  operator: string;
  target?: string;
  detail?: Record<string, unknown>;
  created_at: string;
}

export interface AuthSession {
  sid: string;
  username: string;
  status: string;
  login_at?: string;
  logout_at?: string;
  duration_seconds?: number;
  login_ip?: string;
  login_city?: string;
  logout_reason?: string;
}

export interface PagedResult<T> {
  total: number;
  items: T[];
}

export async function listPermissions(): Promise<PagedResult<AuthPermission>> {
  const { data } = await apiClient.get<{ total: number; permissions: AuthPermission[] }>(
    '/api/v1/auth/management/permissions',
  );
  return { total: data.total || 0, items: data.permissions || [] };
}

export async function listRoles(): Promise<PagedResult<AuthRole>> {
  const { data } = await apiClient.get<{ total: number; roles: AuthRole[] }>(
    '/api/v1/auth/management/roles',
  );
  return { total: data.total || 0, items: data.roles || [] };
}

export async function listModules(): Promise<PagedResult<AuthModule>> {
  const { data } = await apiClient.get<{ total: number; modules: AuthModule[] }>(
    '/api/v1/auth/management/modules',
  );
  return { total: data.total || 0, items: data.modules || [] };
}

export async function createRole(payload: {
  code: string;
  name: string;
  description?: string;
  permissions?: string[];
}): Promise<void> {
  await apiClient.post('/api/v1/auth/management/roles', payload);
}

export async function updateRole(
  roleCode: string,
  payload: { name?: string; description?: string },
): Promise<void> {
  await apiClient.put(`/api/v1/auth/management/roles/${roleCode}`, payload);
}

export async function updateRolePermissions(
  roleCode: string,
  permissions: string[],
): Promise<void> {
  await apiClient.put(`/api/v1/auth/management/roles/${roleCode}/permissions`, { permissions });
}

export async function deleteRole(roleCode: string): Promise<void> {
  await apiClient.delete(`/api/v1/auth/management/roles/${roleCode}`);
}

export async function listUsers(page: number, pageSize: number): Promise<PagedResult<AuthUser>> {
  const { data } = await apiClient.get<{ total: number; users: AuthUser[] }>(
    '/api/v1/auth/management/users',
    { params: { page, page_size: pageSize } },
  );
  return { total: data.total || 0, items: data.users || [] };
}

export async function createUser(payload: {
  username: string;
  password?: string;
  display_name?: string;
  email?: string;
  require_email_verification?: boolean;
  email_mfa_enabled?: boolean;
  roles: string[];
}): Promise<void> {
  await apiClient.post('/api/v1/auth/management/users', payload);
}

export async function updateUserRoles(username: string, roles: string[]): Promise<void> {
  await apiClient.put(`/api/v1/auth/management/users/${username}/roles`, { roles });
}

export async function updateUserStatus(username: string, isActive: boolean): Promise<void> {
  await apiClient.put(`/api/v1/auth/management/users/${username}/status`, { is_active: isActive });
}

export async function updateUserEmailMfa(
  username: string,
  emailMfaEnabled: boolean,
): Promise<void> {
  await apiClient.put(`/api/v1/auth/management/users/${username}/email-mfa-toggle`, {
    email_mfa_enabled: emailMfaEnabled,
  });
}

export async function deleteUser(username: string): Promise<void> {
  await apiClient.delete(`/api/v1/auth/management/users/${username}`);
}

export async function resetUserPassword(
  username: string,
  newPassword?: string,
): Promise<void> {
  await apiClient.put(`/api/v1/auth/management/users/${username}/password/reset`, {
    new_password: newPassword,
  });
}

export async function listAuditLogs(params: {
  operator?: string;
  event?: string;
  dateFrom?: string;
  dateTo?: string;
  page: number;
  pageSize: number;
}): Promise<PagedResult<AuthAuditLog>> {
  const { data } = await apiClient.get<{ total: number; logs: AuthAuditLog[] }>(
    '/api/v1/auth/management/audit-logs',
    {
      params: {
        operator: params.operator || undefined,
        event: params.event || undefined,
        date_from: params.dateFrom || undefined,
        date_to: params.dateTo || undefined,
        page: params.page,
        page_size: params.pageSize,
      },
    },
  );
  return { total: data.total || 0, items: data.logs || [] };
}

export async function listAuthSessions(params: {
  username?: string;
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  page: number;
  pageSize: number;
}): Promise<PagedResult<AuthSession>> {
  const { data } = await apiClient.get<{ total: number; sessions: AuthSession[] }>(
    '/api/v1/auth/management/sessions',
    {
      params: {
        username: params.username || undefined,
        status: params.status || undefined,
        date_from: params.dateFrom || undefined,
        date_to: params.dateTo || undefined,
        page: params.page,
        page_size: params.pageSize,
      },
    },
  );
  return { total: data.total || 0, items: data.sessions || [] };
}

export async function updateMyProfile(payload: {
  display_name?: string;
  email?: string;
}): Promise<void> {
  await apiClient.put('/api/v1/auth/management/me/profile', payload);
}

export async function changeMyPassword(payload: {
  old_password: string;
  new_password: string;
}): Promise<void> {
  await apiClient.put('/api/v1/auth/management/me/password', payload);
}
