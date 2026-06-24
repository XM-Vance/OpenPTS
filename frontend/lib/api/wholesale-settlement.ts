import { apiClient } from './client';

export const getWholesaleSettlementYears = () =>
  apiClient.get('/api/v1/wholesale-monthly-settlement/years');

export const getWholesaleSettlementByYear = (year: string) =>
  apiClient.get(`/api/v1/wholesale-monthly-settlement/year/${year}`);

export const getWholesaleSettlementByMonth = (month: string) =>
  apiClient.get(`/api/v1/wholesale-monthly-settlement/${month}`);

export const getWholesaleSettlementReconciliation = (month: string) =>
  apiClient.get(`/api/v1/wholesale-monthly-settlement/${month}/reconciliation`);

export const postWholesaleSettlementImport = (formData: FormData, params?: Record<string, unknown>) =>
  apiClient.post('/api/v1/wholesale-monthly-settlement/import', formData, {
    params,
    headers: { 'Content-Type': 'multipart/form-data' },
  });

export const importWholesaleMonthlySettlement = (data?: Record<string, unknown>) =>
  apiClient.post('/api/v1/wholesale-monthly-settlement/import', data);

export const getWholesaleMonthlySettlementYears = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/wholesale-monthly-settlement/years', { params });

export const getWholesaleMonthlySettlementByYear = (year: string) =>
  apiClient.get(`/api/v1/wholesale-monthly-settlement/year/${year}`);

export const getWholesaleReconciliation = (month: string) =>
  apiClient.get(`/api/v1/wholesale-monthly-settlement/${month}/reconciliation`);
