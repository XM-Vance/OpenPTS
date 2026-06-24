import { apiClient } from './client';

export interface PricingModel {
  code: string;
  display_name: string;
  package_type: string;
  pricing_mode: string;
  enabled: boolean;
  sort_order: number;
}

export interface RetailPackage {
  id: string;
  package_name: string;
  package_type: string;
  model_code?: string | null;
  is_green_power: boolean;
  status: string;
  description?: string | null;
  pricing_config?: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
}

export interface PackageInput {
  package_name: string;
  package_type: string;
  model_code?: string;
  is_green_power: boolean;
  status?: string;
  description?: string;
  pricing_config?: Record<string, unknown>;
}

export interface RetailContract {
  id: string;
  customer_id: string;
  customer_name: string;
  package_id: string;
  package_name_snapshot: string;
  purchasing_energy_mwh: number;
  green_power_ratio?: number | null;
  purchase_start_month: string;
  purchase_end_month: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ContractInput {
  customer_id: string;
  package_id: string;
  purchasing_energy_mwh: number;
  green_power_ratio?: number;
  purchase_start_month: string;
  purchase_end_month: string;
  status?: string;
}

// ─── 定价模型（只读）───
export async function listPricingModels(): Promise<PricingModel[]> {
  const { data } = await apiClient.get('/api/v1/retail/pricing-models');
  return data.items;
}

// ─── 零售套餐 ───
export async function listPackages(
  params: { keyword?: string; status?: string } = {},
): Promise<RetailPackage[]> {
  const { data } = await apiClient.get('/api/v1/retail/packages', { params });
  return data.items;
}

export async function createPackage(input: PackageInput): Promise<RetailPackage> {
  const { data } = await apiClient.post('/api/v1/retail/packages', input);
  return data;
}

export async function updatePackage(id: string, input: PackageInput): Promise<RetailPackage> {
  const { data } = await apiClient.put(`/api/v1/retail/packages/${id}`, input);
  return data;
}

export async function deletePackage(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/retail/packages/${id}`);
}

// ─── 零售合同 ───
export async function listContracts(
  params: { keyword?: string; status?: string } = {},
): Promise<RetailContract[]> {
  const { data } = await apiClient.get('/api/v1/retail/contracts', { params });
  return data.items;
}

export async function createContract(input: ContractInput): Promise<RetailContract> {
  const { data } = await apiClient.post('/api/v1/retail/contracts', input);
  return data;
}

export async function updateContract(id: string, input: ContractInput): Promise<RetailContract> {
  const { data } = await apiClient.put(`/api/v1/retail/contracts/${id}`, input);
  return data;
}

export async function deleteContract(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/retail/contracts/${id}`);
}

// X1 生成合同 PDF。无 MinIO 时后端直接返回 PDF 字节流，有 MinIO 时返回 JSON 含附件元信息。
export async function generateContractPDF(id: string): Promise<{ ok: boolean; mode: 'download' | 'minio' }> {
  const resp = await apiClient.post(`/api/v1/retail/contracts/${id}/pdf`, null, {
    responseType: 'blob',
  });
  const ct = (resp.headers['content-type'] as string) || '';
  if (ct.includes('application/pdf')) {
    const blob = new Blob([resp.data as Blob], { type: 'application/pdf' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `contract_${id}.pdf`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    return { ok: true, mode: 'download' };
  }
  return { ok: true, mode: 'minio' };
}
