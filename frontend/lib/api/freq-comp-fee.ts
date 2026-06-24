import { apiClient } from './client';

export const getFreqCompFee = (month: string) =>
  apiClient.get(`/api/v1/freq-comp-fee/${month}`);

export const deleteFreqCompFee = (month: string) =>
  apiClient.delete(`/api/v1/freq-comp-fee/${month}`);

export const importFreqCompFeeFormData = (formData: FormData) =>
  apiClient.post('/api/v1/freq-comp-fee/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
