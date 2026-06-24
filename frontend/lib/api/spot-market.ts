import { apiClient } from './client';

export const getSpotMarketStatistics = (params: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /price/spot-market/statistics
  apiClient.get('/api/v1/price/spot-market', { params });

export const getSpotMarketPriceCurve = (params: Record<string, unknown>) =>
  // TODO(P1): 后端无对应路由 GET /price/spot-market/price-curve
  apiClient.get('/api/v1/price/spot-market', { params });

export const getSpotMarketList = (params: Record<string, unknown>) =>
  apiClient.get('/api/v1/price/spot-market', { params });
