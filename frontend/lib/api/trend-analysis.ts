import { apiClient } from './client';

export interface PriceTrendParams {
  start_date: string;
  end_date: string;
}

export const trendAnalysisApi = {
  fetchPriceTrend: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/price-trend', { params });
  },
  fetchWeekdayPattern: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/weekday-pattern', { params });
  },
  fetchVolatility: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/volatility', { params });
  },
  fetchArbitrage: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/market-analysis', { params });
  },
  fetchAnomaly: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/market-analysis', { params });
  },
  fetchTimeSlotStats: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/timeslot-stats', { params });
  },
  fetchDaFactorTrend: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/da-factor-trend', { params });
  },
  fetchRtFactorTrend: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/rt-factor-trend', { params });
  },
  fetchTimeslotAvgPrice: (params: PriceTrendParams) => {
    return apiClient.get('/api/v1/price/trend/timeslot-avg-price', { params });
  },
};
