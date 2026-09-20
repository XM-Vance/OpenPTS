import { apiClient } from './client';
import type { SchemaLoginRequest, SchemaLoginResponse } from './types.gen';

// 登录入参/响应由 OpenAPI 规范生成，与 audit.ts / customers.ts 同模式。
export type LoginParams = SchemaLoginRequest;
export type LoginResponse = SchemaLoginResponse;

export interface Org {
  id: string;
  code: string;
  name: string;
  is_active: boolean;
}

export interface MeResponse {
  user_id: string;
  username: string;
  display_name?: string | null;
  is_hq?: boolean;
  org_active?: string;
  orgs?: Org[];
}

// 登录成功后后端以 httpOnly Cookie 下发登录态(前端 JS 不可读、不落 localStorage);
// body 中的 token 字段仅供脚本/工具使用,浏览器端无需也不应保存。
export async function login(params: LoginParams): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>('/api/v1/auth/login', params);
  return data;
}

// 退出登录:后端清除登录 Cookie。失败也不阻断前端状态清理(幂等)。
export async function logout(): Promise<void> {
  await apiClient.post('/api/v1/auth/logout');
}

export async function me(): Promise<MeResponse> {
  const { data } = await apiClient.get<MeResponse>('/api/v1/auth/me');
  return data;
}

export async function myPermissions(): Promise<string[]> {
  const { data } = await apiClient.get<{ permissions: string[] }>(
    '/api/v1/auth/me/permissions',
  );
  return data.permissions;
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
  await apiClient.post('/api/v1/auth/change-password', {
    old_password: oldPassword,
    new_password: newPassword,
  });
}

// ─── 用户 API Key（MCP/外部脚本以本人身份连 ptis）───

export interface ApiKey {
  id: string;
  name: string;
  key_prefix: string;
  last_used_at?: string | null;
  expires_at?: string | null;
  revoked_at?: string | null;
  created_at: string;
}

export interface CreatedApiKey extends ApiKey {
  key: string; // 明文，仅创建时返回一次
}

export async function listApiKeys(): Promise<ApiKey[]> {
  const { data } = await apiClient.get<{ items: ApiKey[] }>('/api/v1/auth/api-keys');
  return data.items ?? [];
}

export async function createApiKey(name: string): Promise<CreatedApiKey> {
  const { data } = await apiClient.post<CreatedApiKey>('/api/v1/auth/api-keys', { name });
  return data;
}

export async function revokeApiKey(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/auth/api-keys/${id}`);
}
