import { apiClient } from './client';

export const getMechanismEnergy = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/settlement/mechanism-energy', { params });

export const importMechanismEnergy = (data?: Record<string, unknown>) =>
  apiClient.post('/api/v1/settlement/mechanism-energy/import', data);

export const getMechanismEnergyByMonth = (month: string) =>
  apiClient.get(`/api/v1/mechanism-energy/${month}`);

export const importMechanismEnergyFormData = (formData: FormData) =>
  apiClient.post('/api/v1/settlement/mechanism-energy/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
