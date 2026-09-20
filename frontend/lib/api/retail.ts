import { apiClient } from './client';
import type {
  SchemaPricingModel,
  SchemaRetailPackage,
  SchemaRetailContract,
} from './types.gen';

// 实体类型由 OpenAPI 规范生成（与 customers.ts 同模式）。
// 后端改字段 → `npm run gen-api-types` 重生成 → 此处与页面同步报错。
export type PricingModel = SchemaPricingModel;
export type RetailPackage = SchemaRetailPackage;
export type RetailContract = SchemaRetailContract;

// 请求输入类（非 DB 实体）保留手写。
export interface PackageInput {
  package_name: string;
  package_type: string;
  model_code?: string;
  is_green_power: boolean;
  status?: string;
  description?: string;
  pricing_config?: Record<string, unknown>;
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

// ─── P0 结算试算（settlement.Preview，金额经 decimal 以 JSON number 返回）───
export interface SettleLineItem {
  label: string;
  energy_kwh: number;
  unit_price: number;
  amount: number;
}
export interface RetailSettleResult {
  mode: string;
  energy_kwh: number;
  energy_amount: number;
  service_amount: number;
  total_amount: number;
  avg_unit_price: number;
  breakdown: SettleLineItem[];
  capped_unit_price?: number; // 封顶价 元/kWh（仅触发封顶时返回）
  is_capped: boolean; // 是否因套餐单价 > P封顶 而被压到封顶价结算
}
export interface WholesaleSettleResult {
  total_energy_kwh: number;
  total_cost: number;
  avg_price: number;
}
export interface SettleProfitResult {
  revenue: number;
  cost: number;
  gross_profit: number;
  gross_margin: number;
  unit_spread: number;
}
export interface SettlePreviewResult {
  retail: RetailSettleResult;
  wholesale?: WholesaleSettleResult;
  profit?: SettleProfitResult;
}
export interface SettlePreviewInput {
  package_id: string;
  energy_kwh: number;
  market_price?: number;
  period_energy?: Record<string, number>;
  benchmark_price?: number; // P基准 元/kWh（价差分享套餐 + 封顶价用）
  purchase_avg_price?: number; // P购电均价 元/kWh（价差分享套餐用）
  wholesale_avg_price?: number;
}
export async function settlePreview(input: SettlePreviewInput): Promise<SettlePreviewResult> {
  const { data } = await apiClient.post('/api/v1/settlement/preview', input);
  return data;
}

// 落库：算并写一条 customer_profit（is_estimate 区分测算/实际）。需 customer_id/月份 + 批发输入。
export interface SettleSaveInput extends SettlePreviewInput {
  customer_id: string;
  operating_month: string;
  is_estimate: boolean;
}
export async function settleAndSave(
  input: SettleSaveInput,
): Promise<{ persisted: boolean; is_estimate: boolean; message: string; result: SettlePreviewResult }> {
  const { data } = await apiClient.post('/api/v1/settlement/settle', input);
  return data;
}
