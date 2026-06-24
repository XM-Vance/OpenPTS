import { apiClient } from './client';
import type { Org } from './auth';

export type { Org };

export interface OrgMember {
  user_id: string;
  username: string;
  display_name?: string | null;
  is_hq: boolean;
}

export async function listOrgs(): Promise<Org[]> {
  const { data } = await apiClient.get<{ orgs: Org[] }>('/api/v1/orgs');
  return data.orgs;
}

export async function createOrg(code: string, name: string): Promise<Org> {
  const { data } = await apiClient.post<Org>('/api/v1/orgs', { code, name });
  return data;
}

export async function updateOrg(id: string, name: string, isActive: boolean): Promise<void> {
  await apiClient.patch(`/api/v1/orgs/${id}`, { name, is_active: isActive });
}

export async function listOrgMembers(id: string): Promise<OrgMember[]> {
  const { data } = await apiClient.get<{ members: OrgMember[] }>(`/api/v1/orgs/${id}/members`);
  return data.members;
}

// 设置某用户可访问的省 + 总部标记 + 主省。
export async function setUserOrgs(
  userId: string,
  orgIds: string[],
  isHQ: boolean,
  primaryOrgId: string,
): Promise<void> {
  await apiClient.put(`/api/v1/users/${userId}/orgs`, {
    org_ids: orgIds,
    is_hq: isHQ,
    primary_org_id: primaryOrgId,
  });
}
