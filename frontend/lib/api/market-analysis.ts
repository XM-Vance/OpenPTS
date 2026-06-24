import { apiClient } from './client';

export const getMarketRealTime = (date: string) =>
  apiClient.get(`/api/v1/price/market-analysis/real-time?date=${date}`);

export const getMarketDayAhead = (date: string) =>
  apiClient.get(`/api/v1/price/market-analysis/day-ahead?date=${date}`);

export const getMarketSpreadAttribution = (params: Record<string, unknown>) =>
  apiClient.get('/api/v1/price/market-analysis/spread-attribution', { params });

export const getMarketDashboard = (params?: Record<string, unknown>, config?: Record<string, unknown>) =>
  apiClient.get('/api/v1/price/market-analysis/dashboard', { params, ...config });

export const getMarketAnalysisSpreadAttribution = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/price/market-analysis/spread-attribution', { params });

export const getMarketAnalysisDayAhead = (date: string) =>
  apiClient.get(`/api/v1/price/market-analysis/day-ahead?date=${date}`);

export const getMarketAnalysisRealTime = (date: string) =>
  apiClient.get(`/api/v1/price/market-analysis/real-time?date=${date}`);

export const getMarketAnalysisSpreadAttributionByDate = (date: string) =>
  apiClient.get(`/api/v1/price/market-analysis/spread-attribution?date=${date}`);
