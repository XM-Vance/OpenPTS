// 系统配置 API。
// 2026-06 自 p0p1.ts 按域拆分迁移（纯移动，无逻辑变更）。
import { apiClient } from './client';

// ─── W2 系统配置 ───
export interface SystemSetting {
  key: string;
  value: string;
  value_type: string;
  category: string;
  description?: string | null;
  is_editable: boolean;
  is_sensitive: boolean;
  updated_by?: string | null;
  updated_at: string;
}
export async function listSettings(category?: string): Promise<{ items: SystemSetting[] }> {
  const { data } = await apiClient.get('/api/v1/system/settings', { params: { category } });
  return data;
}
export async function updateSetting(key: string, value: string): Promise<void> {
  await apiClient.put(`/api/v1/system/settings/${encodeURIComponent(key)}`, { value });
}
