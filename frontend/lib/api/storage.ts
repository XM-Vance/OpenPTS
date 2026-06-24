import { apiClient } from './client';

export interface StorageStation {
  id: string;
  name: string;
  capacity_mwh: number;
  max_power_mw: number;
  location?: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface StorageOperation {
  id: string;
  station_id: string;
  operation_date: string;
  charge_mwh: number;
  discharge_mwh: number;
  revenue?: number | null;
  avg_soc?: number | null;
  cycles?: number | null;
}

export async function listStorageStations(): Promise<{ items: StorageStation[] }> {
  const { data } = await apiClient.get('/api/v1/storage/stations');
  return data;
}

export async function listStorageOperations(
  stationId: string,
  limit = 30,
): Promise<{ items: StorageOperation[] }> {
  const { data } = await apiClient.get(`/api/v1/storage/stations/${stationId}/operations`, {
    params: { limit },
  });
  return data;
}

export async function generateStorageDemoData(
  days = 30,
): Promise<{ days: number; stations: number; message: string }> {
  const { data } = await apiClient.post('/api/v1/storage/demo-data', { days });
  return data;
}
