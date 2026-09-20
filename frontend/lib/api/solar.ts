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

// pvlib 光伏预测结果（PVWatts 物理模型，由算法服务 heavy 端点计算）
export interface PVForecastPoint {
  datetime: string;
  ac_power_kw: number;
  dc_power_kw?: number;
  cell_temperature?: number | null;
  ghi?: number;
  temp_air?: number;
}
export interface PVForecastResult {
  data: PVForecastPoint[];
  summary: {
    total_energy_kwh?: number;
    peak_power_kw?: number;
    system_capacity_kw?: number;
    capacity_factor?: number;
    hours?: number;
    [k: string]: unknown;
  };
}

/**
 * 调用算法服务 pvlib 对指定站点做真实光伏出力预测（替代高斯钟形伪造）。
 * 算法服务不可达时抛错，调用方应回退到 listSolarForecast（DB 历史预测）。
 */
export async function pvForecastByStation(stationId: string): Promise<PVForecastResult> {
  const { data } = await apiClient.get('/api/v1/solar/forecast/pv', { params: { station_id: stationId } });
  return data;
}
