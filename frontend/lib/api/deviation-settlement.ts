import { apiClient } from './client';

// ─── 数据类型 ───

export interface DeviationSettlement {
  id: string;
  operating_date: string;
  declared_energy_mwh: number;
  actual_energy_mwh: number;
  deviation_energy_mwh: number;
  deviation_rate: number;
  deviation_cost: number;
  penalty_cost: number;
  total_settlement: number;
  category: string;
  created_at: string;
}

export interface DeviationSummaryItem {
  category: string;
  total_deviation_energy_mwh: number;
  total_cost: number;
  avg_deviation_rate: number;
  count: number;
}

// ─── API 调用 ───

export async function listDeviations(
  days = 30,
  category = '',
): Promise<{ items: DeviationSettlement[] }> {
  const params: Record<string, string | number> = { days };
  if (category) params.category = category;
  const { data } = await apiClient.get('/api/v1/settlement/deviation', { params });
  return data;
}

export async function getDeviationSummary(
  days = 30,
): Promise<{ items: DeviationSummaryItem[] }> {
  const { data } = await apiClient.get('/api/v1/settlement/deviation/summary', {
    params: { days },
  });
  return data;
}

export async function generateDeviationDemoData(): Promise<{
  rows: number;
  message: string;
}> {
  const { data } = await apiClient.post('/api/v1/settlement/deviation/demo-data');
  return data;
}

// ─── 偏差考核试算（按省分发：闽价差回收 / 赣标准值法 / 皖燃煤基准，只算不落库）───

export interface DeviationCalcInput {
  province?: string; // 传入你的市场区域标识
  contract_qty_mwh: number; // 合同电量 MWh
  actual_qty_mwh: number; // 实际电量 MWh
  benchmark_price?: number; // FJ：参考主体合同结算均价 元/MWh
  spot_price?: number; // FJ：现货实时加权均价 元/MWh
  short_threshold?: number; // FJ/JX 缺额阈值、AH 区间下限，默认 0.8
  long_threshold?: number; // FJ/JX 超额阈值、AH 区间上限，默认 1.2
  contract_price?: number; // JX：中长期合同价 元/MWh
  ref_price?: number; // JX：参考点价（实时统一结算点）元/MWh
  mkt_contract_avg?: number; // JX：市场化主体中长期均价 元/MWh（缺省=合同价）
  coal_benchmark?: number; // AH：燃煤基准电价 元/MWh（缺省读安徽规则表 384.4）
  assess_rate?: number; // AH：偏差考核比例（缺省读安徽规则表 0.30）
}

// 闽：价差回收结构。
export interface FJDeviationResult {
  contract_ratio: number;
  direction: 'short' | 'long' | 'none';
  recovery_qty_mwh: number;
  recovery_price: number;
  recovery_fee: number;
  triggered: boolean;
}

// 赣：标准值法结构（结算细则 72-78 条）。
export interface JXDeviationResult {
  contract_ratio: number;
  direction: 'short' | 'long' | 'none';
  standard_fee: number; // 标准值电费 元（按 80%/120% 调整合约后重算 CfD）
  recovery_fee: number; // 回收电费 = 标准 − 实际（为负置 0）
}

// 皖：燃煤基准 30% 考核结构（53 号方案）。
export interface AHDeviationResult {
  contract_ratio: number;
  direction: 'short' | 'long' | 'none';
  deviation_qty_mwh: number; // 偏差电量（区间外）MWh
  assess_price: number; // 考核单价 = 燃煤基准 × 30% 元/MWh
  assess_fee: number; // 考核电费 元
}

export type DeviationCalcResult = FJDeviationResult & Partial<JXDeviationResult & AHDeviationResult>;

export async function calculateDeviation(
  input: DeviationCalcInput,
): Promise<DeviationCalcResult> {
  const { data } = await apiClient.post<{ result: DeviationCalcResult }>(
    '/api/v1/settlement/deviation/calculate',
    input,
  );
  return data.result;
}
