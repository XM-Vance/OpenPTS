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
  avg_precipitation: number;
  avg_cloud_cover: number;
}

export const getWeatherLocations = async (): Promise<WeatherLocation[]> => {
  const { data } = await apiClient.get('/api/v1/weather/locations');
  return data;
};

export const createWeatherLocation = async (
  location: Omit<WeatherLocation, 'location_id'> & { location_id: string },
): Promise<WeatherLocation> => {
  const { data } = await apiClient.post('/api/v1/weather/locations', location);
  return data;
};

export const updateWeatherLocation = async (
  locationId: string,
  location: Partial<WeatherLocation>,
): Promise<WeatherLocation> => {
  const { data } = await apiClient.put(`/api/v1/weather/locations/${locationId}`, location);
  return data;
};

export const deleteWeatherLocation = async (locationId: string): Promise<void> => {
  await apiClient.delete(`/api/v1/weather/locations/${locationId}`);
};

export const getWeatherActuals = async (locationId: string, date: string): Promise<WeatherHourlyData[]> => {
  const { data } = await apiClient.get('/api/v1/weather/actuals', {
    params: { location_id: locationId, date },
  });
  return data;
};

export const getWeatherActualsSummary = async (locationId: string, date: string): Promise<DailyWeatherSummary> => {
  const { data } = await apiClient.get('/api/v1/weather/actuals/summary', {
    params: { location_id: locationId, date },
  });
  return data;
};

export const getWeatherForecasts = async (
  locationId: string,
  forecastDate: string,
  targetDate: string,
): Promise<WeatherHourlyData[]> => {
  const { data } = await apiClient.get('/api/v1/weather/forecasts', {
    params: { location_id: locationId, forecast_date: forecastDate, target_date: targetDate },
  });
  return data;
};

export const getWeatherForecastsSummary = async (
  locationId: string,
  forecastDate: string,
): Promise<DailyWeatherSummary[]> => {
  const { data } = await apiClient.get('/api/v1/weather/forecasts/summary', {
    params: { location_id: locationId, forecast_date: forecastDate },
  });
  return data;
};

export const getAvailableForecastDates = async (
  locationId: string,
  targetDate?: string,
): Promise<string[]> => {
  const { data } = await apiClient.get('/api/v1/weather/forecast-dates', {
    params: { location_id: locationId, target_date: targetDate },
  });
  return data;
};

export const getWeatherType = (
  precipitation: number,
  cloudCover: number,
  temperature: number,
): { icon: string; text: string } => {
  if (precipitation > 0) {
    if (temperature < 0) {
      if (precipitation > 5) return { icon: '❄️', text: '大雪' };
      return { icon: '🌨️', text: '小雪' };
    }
    if (temperature <= 2) return { icon: '🌨️', text: '雨夹雪' };
    if (precipitation > 8) return { icon: '🌧️', text: '大雨' };
    if (precipitation > 2.5) return { icon: '🌧️', text: '中雨' };
    return { icon: '🌦️', text: '小雨' };
  }
  if (cloudCover < 20) return { icon: '☀️', text: '晴' };
  if (cloudCover < 50) return { icon: '🌤️', text: '少云' };
  if (cloudCover < 80) return { icon: '⛅', text: '多云' };
  return { icon: '☁️', text: '阴' };
};

export const calculateAccuracy = (actual: number[], predicted: number[]): number => {
  if (actual.length === 0 || actual.length !== predicted.length) return 0;
  const sumAbsDiff = actual.reduce((sum, a, i) => sum + Math.abs(a - predicted[i]), 0);
  const sumActual = actual.reduce((sum, a) => sum + Math.abs(a), 0);
  if (sumActual === 0) return 100;
  const wmape = (sumAbsDiff / sumActual) * 100;
  return Math.max(0, Math.round((100 - wmape) * 10) / 10);
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
