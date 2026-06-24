import { apiClient } from './client';

export interface Contract {
  id: string;
  _id?: string;
  contract_name: string;
  package_name: string;
  package_status?: string;
  package_id: string;
  customer_name: string;
  customer_id: string;
  purchasing_electricity_quantity: number;
  green_power_ratio: number;
  purchase_start_month: string;
  purchase_end_month: string;
  status: 'pending' | 'active' | 'expired';
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface ContractListResponse {
  items: Contract[];
  total: number;
  page: number;
  page_size: number;
}

export interface ContractCreate {
  contract_name?: string;
  package_name: string;
  package_id: string;
  customer_name: string;
  customer_id: string;
  purchasing_electricity_quantity: number;
  green_power_ratio?: number;
  purchase_start_month: string;
  purchase_end_month: string;
}

export interface ContractUpdate {
  contract_name?: string;
  package_name?: string;
  package_id?: string;
  customer_name?: string;
  customer_id?: string;
  purchasing_electricity_quantity?: number;
  green_power_ratio?: number;
  purchase_start_month?: string;
  purchase_end_month?: string;
}

export interface ContractListParams {
  contract_name?: string;
  package_name?: string;
  customer_name?: string;
  status?: 'pending' | 'active' | 'expired' | 'all';
  purchase_start_month?: string;
  purchase_end_month?: string;
  year?: number;
  sort_field?: string;
  sort_order?: 'asc' | 'desc';
  page?: number;
  page_size?: number;
}

export interface ImportError {
  row: number;
  field: string;
  value: string | number | boolean | null;
  message: string;
  suggestion?: string;
}

export interface ImportResult {
  total: number;
  success: number;
  failed: number;
  errors: ImportError[];
}

export interface ExportParams {
  package_name?: string;
  customer_name?: string;
  status?: 'pending' | 'active' | 'expired' | 'all';
  start_month?: string;
  end_month?: string;
}

// PDF 相关
export interface PdfUploadMatchedItem {
  filename: string;
  contract_id: string;
  contract_name: string;
  customer_name: string;
}

export interface PdfCandidateContract {
  _id: string;
  contract_name: string;
  customer_name: string;
  purchase_start_month: string;
  purchase_end_month: string;
  has_pdf: boolean;
  contract_year?: number;
}

export interface PdfUploadPendingItem {
  filename: string;
  reason: string;
  candidates: PdfCandidateContract[];
  target_contract: PdfCandidateContract | null;
}

export interface PdfUploadErrorItem {
  filename: string;
  error: string;
}

export interface PdfUploadResult {
  matched: PdfUploadMatchedItem[];
  pending: PdfUploadPendingItem[];
  errors: PdfUploadErrorItem[];
  summary: {
    total: number;
    matched_count: number;
    pending_count: number;
    error_count: number;
  };
}

// 导入创建相关
export interface MeterPointData {
  meter_id: string;
  measuring_point: string;
  voltage_level: string;
}

export interface ParsePdfResponse {
  customer_name?: string;
  customer_short_name?: string;
  period?: string;
  package_name?: string;
  total_electricity?: number;
  attachment2?: MeterPointData[];
  location?: string;
  is_customer_new: boolean;
  is_package_new: boolean;
  is_contract_duplicate: boolean;
  duplicate_contract_id?: string;
}

export interface ImportCreateRequest {
  customer_name: string;
  customer_short_name: string;
  location?: string;
  period: string;
  package_name: string;
  total_electricity: number;
  attachment2: MeterPointData[];
}

export interface ImportCreateResponse {
  success: boolean;
  contract_id: string;
  customer_id: string;
  package_id: string;
}

let contractYearsCache: { data: number[]; expiresAt: number } | null = null;

export const getContracts = (params?: ContractListParams) =>
  apiClient.get<ContractListResponse>('/api/v1/retail/contracts', { params });

export const getContractYears = () => {
  if (contractYearsCache && contractYearsCache.expiresAt > Date.now()) {
    return Promise.resolve({ data: contractYearsCache.data } as { data: number[] });
  }
  return apiClient.get<number[]>('/api/v1/retail/contracts/years').then((response) => {
    contractYearsCache = { data: response.data, expiresAt: Date.now() + 10 * 60 * 1000 };
    return response;
  });
};

export const getContract = (contractId: string) =>
  apiClient.get<Contract>(`/api/v1/retail/contracts/${contractId}`);

export const createContract = (contractData: ContractCreate) =>
  apiClient.post<Contract>('/api/v1/retail/contracts', contractData);

export const updateContract = (contractId: string, contractData: ContractUpdate) =>
  apiClient.put<Contract>(`/api/v1/retail/contracts/${contractId}`, contractData);

export const deleteContract = (contractId: string) =>
  apiClient.delete(`/api/v1/retail/contracts/${contractId}`);

export const importContracts = (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  return apiClient.post<ImportResult>('/api/v1/retail/contracts/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};

export const exportContracts = (params?: ExportParams) =>
  apiClient.get('/api/v1/retail/contracts/export', { params, responseType: 'blob' });

// PDF 管理
export const uploadContractPdfs = (files: File[]) => {
  const formData = new FormData();
  files.forEach((file) => formData.append('files', file));
  return apiClient.post<PdfUploadResult>('/api/v1/retail/contracts/upload-pdfs', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};

// TODO(P1): 后端无对应 GET /retail/contracts/:id/pdf 路由
export const getContractPdf = (contractId: string) =>
  apiClient.get(`/api/v1/retail/contracts/${contractId}/pdf`, { responseType: 'blob' });

// TODO(P1): 后端无对应 POST /retail/contracts/:id/upload-pdf 路由
export const uploadContractPdf = (contractId: string, file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  return apiClient.post(`/api/v1/retail/contracts/${contractId}/upload-pdf`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};

// TODO(P1): 后端无对应 GET /retail/contracts/:id/has-pdf 路由
export const checkContractHasPdf = (contractId: string) =>
  apiClient.get<{ has_pdf: boolean }>(`/api/v1/retail/contracts/${contractId}/has-pdf`);

// 导入创建
export const parseContractPdf = (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  return apiClient.post<ParsePdfResponse>('/api/v1/retail/contracts/parse-pdf', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};

export const importAndCreateContract = (data: ImportCreateRequest) =>
  apiClient.post<ImportCreateResponse>('/api/v1/retail/contracts/import-create', data);

export interface ContractFormData {
  contract_name: string;
  package_name: string;
  package_id: string;
  customer_name: string;
  customer_id: string;
  purchasing_electricity_quantity: number;
  green_power_ratio: number;
  purchase_start_month: Date | null;
  purchase_end_month: Date | null;
}

export const getRetailContracts = (params?: Record<string, unknown>) =>
  apiClient.get('/api/v1/retail/contracts', { params });

export const getRetailContractById = (id: string) =>
  apiClient.get(`/api/v1/retail/contracts/${id}`);
