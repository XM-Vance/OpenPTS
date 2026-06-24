import { apiClient } from './client';

// ── 站点 ──

export interface SolarStation {
  id: string;
  station_name: string;
  location: string;
  capacity_kw: number;
  status: string;
  installed_date?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  created_at: string;
}

// ── 发电预测 ──

export interface SolarForecast {
  id: string;
  station_id: string;
  forecast_date: string;
  period: number;
  forecast_power_kw: number;
  actual_power_kw?: number | null;
  deviation_rate?: number | null;
  created_at: string;
}

// ── 收益结算 ──

export interface SolarRevenue {
  id: string;
  station_id: string;
  settlement_month: string;
  energy_kwh: number;
  revenue: number;
  avg_price: number;
  subsidy: number;
  net_income: number;
  created_at: string;
}

// ── API ──

export async function listSolarStations(): Promise<{ items: SolarStation[] }> {
  const { data } = await apiClient.get('/api/v1/solar/stations');
  return data;
}

export async function getSolarStation(id: string): Promise<SolarStation> {
  const { data } = await apiClient.get(`/api/v1/solar/stations/${id}`);
  return data;
}

export async function createSolarStation(payload: {
  station_name: string;
  location?: string;
  capacity_kw: number;
  status?: string;
  installed_date?: string;
  latitude?: number;
  longitude?: number;
}): Promise<SolarStation> {
  const { data } = await apiClient.post('/api/v1/solar/stations', payload);
  return data;
}

export async function updateSolarStation(
  id: string,
  payload: {
    station_name: string;
    location?: string;
    capacity_kw: number;
    status?: string;
    installed_date?: string;
    latitude?: number;
    longitude?: number;
  },
): Promise<SolarStation> {
  const { data } = await apiClient.put(`/api/v1/solar/stations/${id}`, payload);
  return data;
}

export async function deleteSolarStation(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/solar/stations/${id}`);
}

export async function listSolarForecast(params?: {
  station_id?: string;
  limit?: number;
}): Promise<{ items: SolarForecast[] }> {
  const { data } = await apiClient.get('/api/v1/solar/forecast', { params });
  return data;
}

export async function listSolarRevenue(params?: {
  station_id?: string;
  limit?: number;
}): Promise<{ items: SolarRevenue[] }> {
  const { data } = await apiClient.get('/api/v1/solar/revenue', { params });
  return data;
}

export async function generateSolarDemoData(
  days = 30,
): Promise<{ days: number; stations: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/solar/demo-data', { days });
  return data;
}
