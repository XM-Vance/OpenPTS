import { apiClient } from './client';

export interface WeatherLocation {
  location_id: string;
  name: string;
  latitude: number;
  longitude: number;
  enabled: boolean;
}

export interface WeatherHourlyData {
  timestamp: string;
  apparent_temperature: number;
  shortwave_radiation: number;
  wind_speed_10m: number;
  wind_speed_100m: number;
  relative_humidity_2m: number;
  precipitation: number;
  cloud_cover: number;
}

export interface DailyWeatherSummary {
  date: string;
  weather_type: string;
  weather_icon: string;
  min_temp: number;
  max_temp: number;
  avg_temp: number;
  humidity: number;
  wind_speed: number;
  avg_precipitation: number;
  avg_cloud_cover: number;
}

export const getWeatherLocations = async (): Promise<WeatherLocation[]> => {
  const { data } = await apiClient.get('/api/v1/weather/locations');
  return data;
};

// 某站近 N 天实况日摘要列表（一次取回多日完整字段，供走势图/KPI 聚合）。
export const getWeatherActualsRange = async (
  locationId: string,
  days = 14,
): Promise<{ items: DailyWeatherSummary[] }> => {
  const { data } = await apiClient.get('/api/v1/weather/actuals/range', {
    params: { location_id: locationId, days },
  });
  return data;
};

// ─── 以下自 v1clone.ts 迁入（2026-06，气象（演示列表）） ───
// ─── D4 气象 ───
export interface WeatherRecord {
  id: string;
  obs_date: string;
  location: string;
  temp_high?: number | null;
  temp_low?: number | null;
  humidity?: number | null;
  precip_mm?: number | null;
  wind_kmh?: number | null;
  load_factor?: number | null;
  description?: string | null;
  created_at: string;
}
export async function listWeather(
  params: { days?: number; location?: string } = {},
): Promise<{ items: WeatherRecord[] }> {
  const { data } = await apiClient.get('/api/v1/weather', { params });
  return data;
}
export async function genWeatherDemo(): Promise<{ records: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/weather/demo-data');
  return data;
}
