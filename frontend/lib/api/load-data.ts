import { apiClient } from './client';

export const getLoadDataCustomers = (params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /load-data/customers
  apiClient.get('/api/v1/load/total', { params });

export const diagnoseLoadData = (data?: Record<string, unknown>) =>
  apiClient.post('/api/v1/load/diagnosis', data);

export const exportLoadDataMpMissing = (params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /load-data/export/mp-missing
  apiClient.get('/api/v1/load/total', { params });

export const getLoadDataSignedCustomers = (params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /load-data/signed-customers
  apiClient.get('/api/v1/load/total', { params });

export const getLoadDataCustomerDetail = (customerId: string) =>
  // TODO(P1): 后端无对应路由 GET /load-data/customers/:id
  apiClient.get(`/api/v1/load/total`, { params: { customer_id: customerId } });

export const getLoadDataCustomerCalendar = (customerId: string, params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /load-data/customers/:id/calendar
  apiClient.get(`/api/v1/load/total`, { params: { customer_id: customerId, ...params } });

export const getLoadDataCustomerCurves = (customerId: string, params?: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /load-data/customers/:id/curves
  apiClient.get(`/api/v1/load/total`, { params: { customer_id: customerId, ...params } });

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

export const importMeterData = (file: File, overwrite = false) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('overwrite', overwrite.toString());
  return apiClient.post<ImportResult>('/api/v1/import/load', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};

export const importMpData = (file: File, overwrite = false) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('overwrite', overwrite.toString());
  return apiClient.post<ImportResult>('/api/v1/import/load', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};

// 重新聚合
export interface ReaggregateResponse {
  success: boolean;
  message: string;
  processed_count?: number;
  details?: Record<string, unknown>;
}

export const reaggregateLoadData = (
  dataType: 'all' | 'mp' | 'meter' = 'all',
  params?: {
    customer_id?: string;
    start_date?: string;
    end_date?: string;
    mode?: 'incremental' | 'full';
  },
) =>
  // TODO(P1): 后端无对应路由 POST /load-data/reaggregate
  apiClient.post<ReaggregateResponse>('/api/v1/load/demo-data', null, {
    params: { data_type: dataType, ...params },
  });

// 校准
export const previewCalibration = (customerId: string, startDate: string, endDate: string) =>
  // TODO(P1): 后端无对应路由 POST /load-data/calibration/preview
  apiClient.post('/api/v1/load/demo-data', null, {
    params: { customer_id: customerId, start_date: startDate, end_date: endDate },
  });

export const calculateCalibration = (customerId: string, startDate: string, endDate: string, accountNo?: string) =>
  // TODO(P1): 后端无对应路由 POST /load-data/calibration/calculate
  apiClient.post('/api/v1/load/demo-data', null, {
    params: { customer_id: customerId, start_date: startDate, end_date: endDate, account_no: accountNo },
  });

export const applyCalibration = (data: {
  customer_id: string;
  coefficients: { meter_id: string; value: number }[];
  update_history: boolean;
  history_range?: [string, string];
}) =>
  // TODO(P1): 后端无对应路由 POST /load-data/calibration/apply
  apiClient.post('/api/v1/load/demo-data', data);

export const getCalibrationDetails = (customerId: string, startDate: string, endDate: string) =>
  // TODO(P1): 后端无对应路由 POST /load-data/calibration/details
  apiClient.post('/api/v1/load/demo-data', null, {
    params: { customer_id: customerId, start_date: startDate, end_date: endDate },
  });
