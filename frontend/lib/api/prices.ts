import { apiClient } from './client';

export const getGridAgencyPricePdf = (id: string, params?: Record<string, unknown>) =>
  apiClient.get(`/api/v1/prices/sgcc/${id}/pdf`, { params, responseType: 'blob' });

export const importGridAgencyPrice = (formData: FormData) =>
  apiClient.post('/api/v1/prices/sgcc/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });

export const getRetailSettlementPrice = (month: string) =>
  apiClient.get(`/api/v1/prices/retail-settlement/${month}`);

export const importRetailSettlementPrice = (importDateType: string, formData: FormData) =>
  apiClient.post(`/api/v1/prices/retail-settlement/import?price_date_type=${importDateType}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });

export const deleteRetailSettlementPrice = (month: string) =>
  apiClient.delete(`/api/v1/prices/retail-settlement/${month}`);

export const getSgccPriceList = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/prices/sgcc', { params });

export const getSgccPricePdf = (id: string) =>
  apiClient.get(`/api/v1/prices/sgcc/${id}/pdf`, { responseType: 'blob' });

export const importSgccPrices = (data?: Record<string, unknown>) =>
  apiClient.post('/api/v1/prices/sgcc/import', data);

export const getRetailSettlementPriceByMonth = (month: string) =>
  apiClient.get(`/api/v1/prices/retail-settlement/${month}`);

export const deleteRetailSettlementPriceByMonth = (month: string) =>
  apiClient.delete(`/api/v1/prices/retail-settlement/${month}`);
