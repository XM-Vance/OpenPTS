import { apiClient } from './client';

export const diagnoseLoadData = (data?: Record<string, unknown>) =>
  apiClient.post('/api/v1/load/diagnosis', data);

// 导入相关
export interface ImportResult {
  success: boolean;
  message: string;
  total_records?: number;
  inserted: number;
  updated: number;
  skipped: number;
  parse_errors?: string[];
  errors?: string[];
}
