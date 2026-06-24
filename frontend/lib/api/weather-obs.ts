import { apiClient } from './client';

// 外部气象观测（风电场风速 / 水库水文）—— 原市场行情，现并入气象数据模块。

export interface WeatherStation {
  code: string;
  name: string;
}

export interface WindHourlyRow {
  location_name: string;
  obs_time: string;
  wind_speed_100m: number | null;
  wind_dir_100m: number | null;
  temperature_2m: number | null;
  humidity_2m: number | null;
}

export interface HydroDailyRow {
  location_name: string;
  obs_date: string;
  temp_mean: number | null;
  humidity_mean: number | null;
  precipitation_sum: number | null;
  rain_sum: number | null;
  et0_evapotranspiration: number | null;
  wind_speed_10m_mean: number | null;
}

export const getWindFarm = (params?: { station?: string; hours?: number }) =>
  apiClient.get<{ stations: WeatherStation[]; station: string; items: WindHourlyRow[] }>(
    '/api/v1/weather/wind-farm',
    { params },
  );

export const getHydrology = (params?: { station?: string; days?: number }) =>
  apiClient.get<{ stations: WeatherStation[]; station: string; items: HydroDailyRow[] }>(
    '/api/v1/weather/hydrology',
    { params },
  );

export const genWeatherObsDemo = () =>
  apiClient.post<{ rows: number; message: string }>('/api/v1/weather/obs-demo-data');
