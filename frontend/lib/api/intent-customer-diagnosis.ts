import { apiClient } from './client';

export interface IntentCustomerSummary {
  id: string;
  customer_name: string;
  created_at: string;
  updated_at: string;
  last_imported_at?: string | null;
  last_aggregated_at?: string | null;
  coverage_start?: string | null;
  coverage_end?: string | null;
  coverage_days: number;
  missing_days: number;
  completeness: number;
  avg_daily_load: number;
  max_daily_load: number;
  min_daily_load: number;
  missing_meter_days: number;
  interpolated_days: number;
  dirty_days: number;
  meter_count: number;
}

export interface IntentPreviewFileItem {
  filename: string;
  meter_id: string;
  account_id: string;
  extracted_customer_name?: string | null;
  start_date: string;
  end_date: string;
  record_count: number;
  default_multiplier: number;
  parse_errors: string[];
}

export interface IntentPreviewResponse {
  suggested_customer_name?: string | null;
  files: IntentPreviewFileItem[];
  validation: {
    can_import: boolean;
    errors: string[];
    warnings: string[];
  };
}

export interface IntentLoadDataResponse {
  customer: IntentCustomerSummary;
  month_data: Array<{ date: string; label: string; totalLoad: number; isMissing: boolean }>;
  intraday_data: Array<{ time: string; load: number }>;
  selected_day_total: number;
}

export interface IntentImportConfig {
  filename: string;
  meter_id: string;
  account_id: string;
  multiplier: number;
}

export interface IntentWholesaleSummaryRow {
  settlement_month: string;
  total_energy_mwh: number;
  daily_cost_total: number;
  daily_cost_unit_price: number;
  surplus_unit_price: number;
  surplus_cost: number;
  total_cost: number;
  unit_cost_yuan_per_mwh: number;
  unit_cost_yuan_per_kwh: number;
  status: string;
  message: string;
}

export interface IntentWholesalePeriodDetail {
  period: number;
  time_label: string;
  load_mwh: number;
  daily_cost_total: number;
  surplus_cost: number;
  total_cost: number;
  period_type?: string;
  daily_cost_unit_price: number;
  final_unit_price: number;
}

export interface IntentWholesaleDailyDetail {
  date: string;
  total_energy_mwh: number;
  daily_cost_total: number;
  surplus_cost: number;
  total_cost: number;
  unit_cost_yuan_per_mwh: number;
}

export interface IntentWholesaleMonthDetail {
  settlement_month: string;
  summary: IntentWholesaleSummaryRow;
  period_details: IntentWholesalePeriodDetail[];
  daily_details: IntentWholesaleDailyDetail[];
}

export interface IntentWholesaleSimulationResponse {
  customer: IntentCustomerSummary;
  summary_rows: IntentWholesaleSummaryRow[];
  month_details: IntentWholesaleMonthDetail[];
}

export interface IntentRetailPackageOption {
  package_id: string;
  package_name: string;
  package_type?: string | null;
  model_code?: string | null;
  is_green_power: boolean;
  status?: string | null;
}

export interface IntentRetailCalculatedPackageItem {
  package_id: string;
  package_name: string;
  model_code?: string | null;
  updated_at?: string | null;
}

export interface IntentDeleteResultResponse {
  deleted_count: number;
  message: string;
}

export interface IntentRetailMonthResultRow {
  settlement_month: string;
  total_energy_mwh: number;
  wholesale_unit_price: number;
  wholesale_amount: number;
  retail_unit_price: number;
  retail_amount: number;
  monthly_gross_profit: number;
  price_spread_per_mwh: number;
  is_capped: boolean;
}

export interface IntentRetailMonthResultsResponse {
  customer: IntentCustomerSummary;
  package_id: string;
  package_name: string;
  rows: IntentRetailMonthResultRow[];
}

export interface IntentRetailSimulationDetail {
  customer_id: string;
  customer_name: string;
  settlement_month: string;
  package_id: string;
  package_name: string;
  model_code?: string | null;
  price_model: Record<string, unknown>;
  pre_stage: Record<string, unknown>;
  sttl_stage: Record<string, unknown>;
  refund_context: Record<string, unknown>;
  final_stage: Record<string, unknown>;
  period_details: Array<Record<string, unknown>>;
  daily_details: Array<Record<string, unknown>>;
  final_energy_mwh: number;
  final_retail_fee: number;
  final_retail_unit_price: number;
  final_wholesale_fee: number;
  final_wholesale_unit_price: number;
  final_gross_profit: number;
  final_price_spread_per_mwh: number;
  final_excess_refund_fee: number;
  sttl_balancing_energy_mwh: number;
  sttl_balancing_retail_fee: number;
  sttl_balancing_wholesale_fee: number;
  pre_energy_mwh?: number;
  pre_retail_fee?: number;
  pre_retail_unit_price?: number;
  pre_wholesale_fee?: number;
  pre_wholesale_unit_price?: number;
  pre_gross_profit?: number;
  pre_price_spread_per_mwh?: number;
  sttl_energy_mwh?: number;
  sttl_retail_fee?: number;
  sttl_retail_unit_price?: number;
  sttl_wholesale_fee?: number;
  sttl_wholesale_unit_price?: number;
  sttl_gross_profit?: number;
  sttl_price_spread_per_mwh?: number;
}

export const listIntentCustomers = async () => {
  const { data } = await apiClient.get<{ items: IntentCustomerSummary[] }>('/api/v1/intent-customer-diagnosis/customers');
  return data.items;
};

export const previewIntentCustomerFiles = async (files: File[]) => {
  const formData = new FormData();
  files.forEach((file) => formData.append('files', file));
  const { data } = await apiClient.post<IntentPreviewResponse>('/api/v1/intent-customer-diagnosis/preview', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
};

export const importIntentCustomerFiles = async (configs: IntentImportConfig[], files: File[]) => {
  const formData = new FormData();
  formData.append('meter_configs_json', JSON.stringify(configs));
  files.forEach((file) => formData.append('files', file));
  const { data } = await apiClient.post('/api/v1/intent-customer-diagnosis/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
};

export const getIntentCustomerLoadData = async (customerId: string, month: string, date: string) => {
  const { data } = await apiClient.get<IntentLoadDataResponse>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/load-data`,
    { params: { month, date } },
  );
  return data;
};

export const deleteIntentCustomer = async (customerId: string, password: string) => {
  await apiClient.delete(`/api/v1/intent-customer-diagnosis/customers/${customerId}`, {
    data: { password },
  });
};

export const calculateIntentCustomerWholesaleSimulation = async (customerId: string) => {
  const { data } = await apiClient.post<IntentWholesaleSimulationResponse>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/wholesale-simulation`,
  );
  return data;
};

export const getIntentCustomerWholesaleSimulation = async (customerId: string) => {
  const { data } = await apiClient.get<IntentWholesaleSimulationResponse>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/wholesale-simulation`,
  );
  return data;
};

export const listIntentRetailActivePackages = async () => {
  const { data } = await apiClient.get<{ items: IntentRetailPackageOption[] }>(
    '/api/v1/intent-customer-diagnosis/retail-simulation/packages/active',
  );
  return data.items;
};

export const listIntentRetailCalculatedPackages = async (customerId: string) => {
  const { data } = await apiClient.get<{ items: IntentRetailCalculatedPackageItem[] }>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/retail-simulation/packages`,
  );
  return data.items;
};

export const listIntentRetailMonthResults = async (customerId: string, packageId: string) => {
  const { data } = await apiClient.get<IntentRetailMonthResultsResponse>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/retail-simulation/months`,
    { params: { package_id: packageId } },
  );
  return data;
};

export const calculateIntentRetailSimulation = async (customerId: string, packageId: string) => {
  const { data } = await apiClient.post<IntentRetailSimulationDetail>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/retail-simulation`,
    { package_id: packageId },
  );
  return data;
};

export const getIntentRetailSimulationDetail = async (customerId: string, packageId: string, settlementMonth: string) => {
  const { data } = await apiClient.get<IntentRetailSimulationDetail>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/retail-simulation/detail`,
    { params: { package_id: packageId, settlement_month: settlementMonth } },
  );
  return data;
};

export const deleteIntentRetailSimulationPackage = async (customerId: string, packageId: string) => {
  const { data } = await apiClient.delete<IntentDeleteResultResponse>(
    `/api/v1/intent-customer-diagnosis/customers/${customerId}/retail-simulation/packages/${packageId}`,
  );
  return data;
};
