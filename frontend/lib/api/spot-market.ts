import { apiClient } from './client';

export const getSpotMarketStatistics = (params: Record<string, unknown>) =>
  // 路由已实现（routes.golden SpotMarketHandler.Statistics），旧 TODO 注释过时已删。
  apiClient.get('/api/v1/price/spot-market/statistics', { params });

export const getSpotMarketPriceCurve = (params: Record<string, unknown>) =>
  // 路由已实现（routes.golden SpotMarketHandler.PriceCurve），旧 TODO 注释过时已删。
  apiClient.get('/api/v1/price/spot-market/price-curve', { params });

export const getSpotMarketList = (params: Record<string, unknown>) =>
  apiClient.get('/api/v1/price/spot-market', { params });

// 生成演示数据（填充 spot_market_daily 表）
export const generateSpotMarketDemoData = () =>
  apiClient.post('/api/v1/price/spot-market/demo-data');
